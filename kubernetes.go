package kubegen

import (
	"fmt"
	"log"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kcache "k8s.io/kubernetes/pkg/client/cache"
	krest "k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kselector "k8s.io/kubernetes/pkg/fields"
	kwatch "k8s.io/kubernetes/pkg/watch"
)

// Initializes a new Kubernetes API Client
// TODO: Authentication
// TODO: look inte pkg/client/clientcmd for loading config
// TODO: standard environment variables?
func newKubeClient(c Config) (*kclient.Client, error) {
	config := &krest.Config{
		Host:          c.Host,
		ContentConfig: krest.ContentConfig{GroupVersion: &unversioned.GroupVersion{Version: "v1"}},
	}
	return kclient.New(config)
}

func podsListWatch(client *kclient.Client) *kcache.ListWatch {
	return kcache.NewListWatchFromClient(client, "pods", kapi.NamespaceAll, kselector.Everything())
}

func svcListWatch(client *kclient.Client) *kcache.ListWatch {
	return kcache.NewListWatchFromClient(client, "services", kapi.NamespaceAll, kselector.Everything())
}

func epListWatch(client *kclient.Client) *kcache.ListWatch {
	return kcache.NewListWatchFromClient(client, "endpoints", kapi.NamespaceAll, kselector.Everything())
}

func watchPods(client *kclient.Client, ch chan<- *kapi.Pod, stopCh chan struct{}) error {
	w, err := client.Pods(kapi.NamespaceAll).Watch(kapi.ListOptions{Watch: true})
	if err != nil {
		return fmt.Errorf("error starting pod watcher: %v", err)
	}
	go func() {
		for {
			select {
			case e := <-w.ResultChan():
				switch e.Type {
				case kwatch.Added, kwatch.Modified, kwatch.Deleted:
					if p, ok := e.Object.(*kapi.Pod); ok {
						ch <- p
					}
				case kwatch.Error:
					log.Printf("watch error: %v\n", e.Object)
				}
			case <-stopCh:
				w.Stop()
			}
		}
	}()
	return nil
}

func watchServices(client *kclient.Client, ch chan<- *kapi.Service, stopCh chan struct{}) error {
	w, err := client.Services(kapi.NamespaceAll).Watch(kapi.ListOptions{Watch: true})
	if err != nil {
		return fmt.Errorf("error starting service watcher: %v", err)
	}
	go func() {
		for {
			select {
			case e := <-w.ResultChan():
				switch e.Type {
				case kwatch.Added, kwatch.Modified, kwatch.Deleted:
					if p, ok := e.Object.(*kapi.Service); ok {
						ch <- p
					}
				case kwatch.Error:
					log.Printf("watch error: %v\n", e.Object)
				}
			case <-stopCh:
				w.Stop()
			}
		}
	}()
	return nil
}

func watchEndpoints(client *kclient.Client, ch chan<- *kapi.Endpoints, stopCh chan struct{}) error {
	w, err := client.Endpoints(kapi.NamespaceAll).Watch(kapi.ListOptions{Watch: true})
	if err != nil {
		return fmt.Errorf("error starting pod watcher: %v", err)
	}
	go func() {
		for {
			select {
			case e := <-w.ResultChan():
				switch e.Type {
				case kwatch.Added, kwatch.Modified, kwatch.Deleted:
					if p, ok := e.Object.(*kapi.Endpoints); ok {
						ch <- p
					}
				case kwatch.Error:
					log.Printf("watch error: %v\n", e.Object)
				}
			case <-stopCh:
				w.Stop()
			}
		}
	}()
	return nil
}
