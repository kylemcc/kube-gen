package kubegen

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	kapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "k8s.io/client-go/kubernetes"
)

var (
	validTypes = map[string]bool{
		"pods":      true,
		"services":  true,
		"endpoints": true,
	}
)

type Config struct {
	Host               string
	Kubeconfig         string
	TemplatePath       string
	TemplateString     string
	Output             string
	Overwrite          bool
	Watch              bool
	PreCmd             string
	PostCmd            string
	LogCmdOutput       bool
	Interval           int
	MinWait            time.Duration
	MaxWait            time.Duration
	ResourceTypes      []string
	UseInClusterConfig bool
}

type Generator interface {
	Generate() error
}

type generator struct {
	sync.WaitGroup
	Config Config
	Client *kclient.Clientset

	loadPods bool
	loadSvcs bool
	loadEps  bool
}

func NewGenerator(c Config) (Generator, error) {
	kclient, err := newKubeClient(c)
	return &generator{
		Config:   c,
		Client:   kclient,
		loadPods: len(c.ResourceTypes) == 0 || containsString(c.ResourceTypes, "pods"),
		loadSvcs: len(c.ResourceTypes) == 0 || containsString(c.ResourceTypes, "services"),
		loadEps:  len(c.ResourceTypes) == 0 || containsString(c.ResourceTypes, "endpoints"),
	}, err
}

func (g *generator) Generate() error {
	if err := g.validateConfig(); err != nil {
		return err
	}

	if g.Config.Watch {
		// watch for updates
		if err := g.watchEvents(); err != nil {
			return err
		}
	} else {
		// initial render
		return g.execute()
	}

	g.Wait()
	return nil
}

func (g *generator) execute() error {
	ctx := &Context{}

	log.Println("refreshing state...")
	start := time.Now()
	if g.loadPods {
		if p, err := g.Client.CoreV1().Pods(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{}); err != nil {
			return fmt.Errorf("error loading pods: %v", err)
		} else {
			ctx.Pods = p.Items
		}
	}
	if g.loadSvcs {
		if p, err := g.Client.CoreV1().Services(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{}); err != nil {
			return fmt.Errorf("error loading services: %v", err)
		} else {
			ctx.Services = p.Items
		}
	}
	if g.loadEps {
		if p, err := g.Client.CoreV1().Endpoints(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{}); err != nil {
			return fmt.Errorf("error loading endpoints: %v", err)
		} else {
			ctx.Endpoints = p.Items
		}
	}
	log.Printf("done. took %v\n", time.Since(start))

	var (
		content []byte
		err     error
	)
	if g.Config.TemplateString != "" {
		content, err = execTemplateString(g.Config.TemplateString, ctx)
	} else {
		content, err = execTemplateFile(g.Config.TemplatePath, ctx)
	}
	if err != nil {
		return err
	}

	if err := g.runCmd(g.Config.PreCmd); err != nil {
		return err
	}
	if err := g.writeFile(content); err != nil {
		return err
	}
	return g.runCmd(g.Config.PostCmd)
}

func (g *generator) watchEvents() error {
	if !g.Config.Watch {
		return nil
	}

	var (
		nWatchers int
		podCh     chan *kapi.Pod
		svcCh     chan *kapi.Service
		epCh      chan *kapi.Endpoints
		ticker    *time.Ticker
		tickerCh  <-chan time.Time
	)

	// channel for signaling shutdown to watchers
	stopCh := make(chan struct{})
	// channel for receiving signals
	sigCh := newSigChan()

	if g.loadPods {
		nWatchers++
		podCh = make(chan *kapi.Pod)
		watchPods(g.Client, podCh, stopCh)
	}
	if g.loadSvcs {
		nWatchers++
		svcCh = make(chan *kapi.Service)
		watchServices(g.Client, svcCh, stopCh)
	}
	if g.loadEps {
		nWatchers++
		epCh = make(chan *kapi.Endpoints)
		watchEndpoints(g.Client, epCh, stopCh)
	}
	if g.Config.Interval > 0 {
		ticker = time.NewTicker(time.Duration(g.Config.Interval) * time.Second)
		tickerCh = ticker.C
	}

	// channel for receiving events from the kubernetes api
	eventCh := make(chan interface{})
	// debounce rapidly occurring events
	debounceCh := newDebouncer(eventCh, g.Config.MinWait, g.Config.MaxWait)
	go func() {
		for range debounceCh {
			if err := g.execute(); err != nil {
				log.Printf("error rendering template: %v\n", err)
			}
		}
	}()

	// watch for various events that trigger template rendering
	g.Add(1)
	go func() {
		defer g.Done()
		for {
			select {
			case p := <-podCh:
				eventCh <- p
			case s := <-svcCh:
				eventCh <- s
			case e := <-epCh:
				eventCh <- e
			case t := <-tickerCh:
				eventCh <- t
			case sig := <-sigCh:
				switch sig {
				case syscall.SIGHUP:
					eventCh <- sig
				case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
					if ticker != nil {
						ticker.Stop()
					}
					for i := 0; i < nWatchers; i++ {
						stopCh <- struct{}{}
					}
					return
				}
			}
		}
	}()
	return nil
}

func (g *generator) writeFile(content []byte) error {
	if g.Config.Output == "" {
		os.Stdout.Write(content)
		return nil
	}

	// write to a temp file first so we can copy it into place with a single atomic operation
	tmp, err := ioutil.TempFile("", fmt.Sprintf("kube-gen-%d", time.Now().UnixNano()))
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}

	if _, err := tmp.Write(content); err != nil {
		return fmt.Errorf("error writing temp file: %v", err)
	}

	var (
		oldContent []byte
		exists     bool
	)
	if fi, err := os.Stat(g.Config.Output); err == nil {
		exists = true
		// set permissions and ownership on new file
		if err := setFileModeAndOwnership(tmp, fi); err != nil {
			return err
		}
		if oldContent, err = ioutil.ReadFile(g.Config.Output); err != nil {
			return fmt.Errorf("error comparing old version: %v", err)
		}
	}

	if !bytes.Equal(oldContent, content) {
		// Always overwrite in watch mode - doesn't make sense
		// to watch and not overwrite
		if exists && !g.Config.Watch && !g.Config.Overwrite {
			return fmt.Errorf("output file already exists")
		}

		if err = moveFile(tmp, g.Config.Output); err != nil {
			return fmt.Errorf("error creating output file: %v", err)
		}
		log.Printf("output file [%s] created\n", g.Config.Output)
	}

	return nil
}

func (g *generator) runCmd(cs string) error {
	if cs == "" {
		return nil
	}

	log.Printf("running command [%v]\n", cs)
	cmd := exec.Command(SHELL_EXE, SHELL_ARG, cs)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running command: %v", err)
	}
	if g.Config.LogCmdOutput {
		log.Printf("%s: %s\n", cs, out)
	}
	return nil
}

func (g *generator) validateConfig() error {
	if err := validateTypes(g.Config.ResourceTypes); err != nil {
		return err
	}
	return nil
}

func validateTypes(types []string) error {
	for _, t := range types {
		if !validTypes[t] {
			return fmt.Errorf("invalid type: %s", t)
		}
	}
	return nil
}

func newSigChan() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	return ch
}

func newDebouncer(inCh chan interface{}, minWait, maxWait time.Duration) chan interface{} {
	if minWait < 1 {
		return inCh
	}

	outCh := make(chan interface{}, 100)
	go func() {
		var (
			latestEvent interface{}
			minTimer    <-chan time.Time
			maxTimer    <-chan time.Time
		)
		for {
			select {
			case obj := <-inCh:
				latestEvent = obj
				minTimer = time.After(minWait)
				if maxTimer == nil && maxWait > 0 {
					maxTimer = time.After(maxWait)
				}
			case <-minTimer:
				log.Printf("min wait time reached")
				minTimer = nil
				maxTimer = nil
				outCh <- latestEvent
			case <-maxTimer:
				log.Printf("max wait time reached")
				minTimer = nil
				maxTimer = nil
				outCh <- latestEvent
			}
		}
	}()

	return outCh
}
