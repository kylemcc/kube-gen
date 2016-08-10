package kubegen

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kcache "k8s.io/kubernetes/pkg/client/cache"
	krest "k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kframework "k8s.io/kubernetes/pkg/controller/framework"
	kselector "k8s.io/kubernetes/pkg/fields"
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

func watchPods(client *kclient.Client, ch chan<- *kapi.Pod, stopCh chan struct{}) kcache.Store {
	store, controller := kframework.NewInformer(
		podsListWatch(client),
		&kapi.Pod{},
		0,
		kframework.ResourceEventHandlerFuncs{
			AddFunc: func(v interface{}) {
				if p, ok := v.(*kapi.Pod); ok {
					ch <- p
				}
			},
			UpdateFunc: func(ov, nv interface{}) {
				if p, ok := nv.(*kapi.Pod); ok {
					ch <- p
				}
			},
			DeleteFunc: func(v interface{}) {
				if p, ok := v.(*kapi.Pod); ok {
					ch <- p
				}
			},
		})
	go controller.Run(stopCh)
	return store
}

func watchServices(client *kclient.Client, ch chan<- *kapi.Service, stopCh chan struct{}) kcache.Store {
	store, controller := kframework.NewInformer(
		svcListWatch(client),
		&kapi.Service{},
		0,
		kframework.ResourceEventHandlerFuncs{
			AddFunc: func(v interface{}) {
				if s, ok := v.(*kapi.Service); ok {
					ch <- s
				}
			},
			UpdateFunc: func(ov, nv interface{}) {
				if s, ok := nv.(*kapi.Service); ok {
					ch <- s
				}
			},
			DeleteFunc: func(v interface{}) {
				if s, ok := v.(*kapi.Service); ok {
					ch <- s
				}
			},
		})
	go controller.Run(stopCh)
	return store
}

func watchEndpoints(client *kclient.Client, ch chan<- *kapi.Endpoints, stopCh chan struct{}) kcache.Store {
	store, controller := kframework.NewInformer(
		epListWatch(client),
		&kapi.Endpoints{},
		0,
		kframework.ResourceEventHandlerFuncs{
			AddFunc: func(v interface{}) {
				if e, ok := v.(*kapi.Endpoints); ok {
					ch <- e
				}
			},
			UpdateFunc: func(ov, nv interface{}) {
				if e, ok := nv.(*kapi.Endpoints); ok {
					ch <- e
				}
			},
			DeleteFunc: func(v interface{}) {
				if e, ok := v.(*kapi.Endpoints); ok {
					ch <- e
				}
			},
		})
	go controller.Run(stopCh)
	return store
}
