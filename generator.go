package kubegen

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	kapi "k8s.io/kubernetes/pkg/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
)

var (
	validTypes = map[string]bool{
		"pods":      true,
		"services":  true,
		"endpoints": true,
	}
)

type Config struct {
	Host          string
	Template      string
	Output        string
	Overwrite     bool
	Watch         bool
	NotifyCmd     string
	Interval      int
	MinWait       time.Duration
	MaxWait       time.Duration
	ResourceTypes []string
}

type Generator interface {
	Generate() error
}

type generator struct {
	sync.WaitGroup
	Config Config
	Client *kclient.Client

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
		g.execute()
	}

	g.Wait()
	return nil
}

func (g *generator) execute() error {
	ctx := &Context{}

	log.Println("refreshing state...")
	start := time.Now()
	if g.loadPods {
		if p, err := g.Client.Pods(kapi.NamespaceAll).List(kapi.ListOptions{}); err != nil {
			return fmt.Errorf("error loading pods: %v", err)
		} else {
			ctx.Pods = p.Items
		}
	}
	if g.loadSvcs {
		if p, err := g.Client.Services(kapi.NamespaceAll).List(kapi.ListOptions{}); err != nil {
			return fmt.Errorf("error loading services: %v", err)
		} else {
			ctx.Services = p.Items
		}
	}
	if g.loadEps {
		if p, err := g.Client.Endpoints(kapi.NamespaceAll).List(kapi.ListOptions{}); err != nil {
			return fmt.Errorf("error loading endpoints: %v", err)
		} else {
			ctx.Endpoints = p.Items
		}
	}
	log.Printf("done. took %v\n", time.Since(start))
	return nil
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
		if err := watchPods(g.Client, podCh, stopCh); err != nil {
			return err
		}
	}
	if g.loadSvcs {
		nWatchers++
		svcCh := make(chan *kapi.Service)
		if err := watchServices(g.Client, svcCh, stopCh); err != nil {
			return err
		}
	}
	if g.loadEps {
		nWatchers++
		epCh := make(chan *kapi.Endpoints)
		if err := watchEndpoints(g.Client, epCh, stopCh); err != nil {
			return err
		}
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
		for e := range debounceCh {
			log.Printf("event received: %#v\n", e)
			g.execute()
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
				if maxTimer == nil {
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
