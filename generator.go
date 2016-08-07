package kubegen

import (
	"fmt"
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

	g.Wait()
	return nil
}

func (g *generator) execute() {
}

func (g *generator) watchEvents() {
	if !g.Config.Watch {
		return
	}

	var (
		nWatchers int
		podCh     chan *kapi.Pod
		svcCh     chan *kapi.Service
		epCh      chan *kapi.Endpoints
		ticker    *time.Ticker
		tickerCh  <-chan time.Time
	)

	stopCh := make(chan struct{})
	sigCh := newSigChan()

	if len(g.Config.ResourceTypes) == 0 || containsString(g.Config.ResourceTypes, "pods") {
		nWatchers++
		podCh = make(chan *kapi.Pod)
		watchPods(g.Client, podCh, stopCh)
	}
	if len(g.Config.ResourceTypes) == 0 || containsString(g.Config.ResourceTypes, "services") {
		nWatchers++
		svcCh := make(chan *kapi.Service)
		watchServices(g.Client, svcCh, stopCh)
	}
	if len(g.Config.ResourceTypes) == 0 || containsString(g.Config.ResourceTypes, "endpoints") {
		nWatchers++
		epCh := make(chan *kapi.Endpoints)
		watchEndpoints(g.Client, epCh, stopCh)
	}
	if g.Config.Interval > 0 {
		ticker = time.NewTicker(time.Duration(g.Config.Interval) * time.Second)
		tickerCh = ticker.C
	}

	eventCh := make(chan interface{})
	debounceCh := newDebouncer(eventCh, g.Config.MinWait, g.Config.MaxWait)
	go func() {
		for _ = range debounceCh {
			g.execute()
		}
	}()

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
				minTimer = nil
				maxTimer = nil
				outCh <- latestEvent
			case <-maxTimer:
				minTimer = nil
				maxTimer = nil
				outCh <- latestEvent
			}
		}
	}()

	return outCh
}
