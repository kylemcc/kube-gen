package kubegen

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	ResourceTypes []string
}

type Generator interface {
	Generate() error
}

type generator struct {
	sync.WaitGroup
	Config Config
	Client *kclient.Client
}

func NewGenerator(c Config) (Generator, error) {
	kclient, err := newKubeClient(c)
	return &generator{
		Config: c,
		Client: kclient,
	}, err
}

func (g *generator) Generate() error {
	if err := g.validateConfig(); err != nil {
		return err
	}

	g.execute()
	g.watchEvents()
	g.watchSignals()

	g.Wait()
	return nil
}

func (g *generator) execute() {
}

func (g *generator) watchEvents() {
	if !g.Config.Watch {
		return
	}

	stopCh := make(chan struct{})
	podCh := make(chan *kapi.Pod)
	svcCh := make(chan *kapi.Service)
	epCh := make(chan *kapi.Endpoints)
	sigCh := newSigChan()

	var nWatchers int
	watchPods(g.Client, podCh, stopCh)
	watchServices(g.Client, svcCh, stopCh)
	watchEndpoints(g.Client, epCh, stopCh)

	g.Add(1)
	go func() {
		defer g.Done()
		for {
			select {
			case <-podCh:
				g.execute()
			case <-svcCh:
				g.execute()
			case <-epCh:
				g.execute()
			case sig := <-sigCh:
				switch sig {
				case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
					for i := 0; i < nWatchers; i++ {
						stopCh <- struct{}{}
					}
					return
				}
			}
		}
	}()
}

func (g *generator) watchSignals() {
	if !g.Config.Watch {
		return
	}

	g.Add(1)
	go func() {
		defer g.Done()
		sigCh := newSigChan()
		for {
			sig := <-sigCh
			log.Printf("received signal: %s\n", sig)
			switch sig {
			case syscall.SIGHUP:
				g.execute()
			case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
				return
			}
		}
	}()
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
