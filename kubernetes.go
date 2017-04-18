package kubegen

import (
	kclient "k8s.io/client-go/kubernetes"
	kapi "k8s.io/client-go/pkg/api/v1"
	kselector "k8s.io/client-go/pkg/fields"
	krest "k8s.io/client-go/rest"
	kcache "k8s.io/client-go/tools/cache"
)

// Initializes a new Kubernetes API Client
// TODO: Authentication
// TODO: look inte pkg/client/clientcmd for loading config
// TODO: standard environment variables?
func newKubeClient(c Config) (*kclient.Clientset, error) {
	config := &krest.Config{
		Host:          c.Host,
		ContentConfig: krest.ContentConfig{GroupVersion: &kapi.SchemeGroupVersion},
	}
	return kclient.NewForConfig(config)
}

func podsListWatch(client *kclient.Clientset) *kcache.ListWatch {
	return kcache.NewListWatchFromClient(client.Core().RESTClient(), "pods", kapi.NamespaceAll, kselector.Everything())
}

func svcListWatch(client *kclient.Clientset) *kcache.ListWatch {
	return kcache.NewListWatchFromClient(client.Core().RESTClient(), "services", kapi.NamespaceAll, kselector.Everything())
}

func epListWatch(client *kclient.Clientset) *kcache.ListWatch {
	return kcache.NewListWatchFromClient(client.Core().RESTClient(), "endpoints", kapi.NamespaceAll, kselector.Everything())
}

func watchPods(client *kclient.Clientset, ch chan<- *kapi.Pod, stopCh chan struct{}) kcache.Store {
	store, controller := kcache.NewInformer(
		podsListWatch(client),
		&kapi.Pod{},
		0,
		kcache.ResourceEventHandlerFuncs{
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

func watchServices(client *kclient.Clientset, ch chan<- *kapi.Service, stopCh chan struct{}) kcache.Store {
	store, controller := kcache.NewInformer(
		svcListWatch(client),
		&kapi.Service{},
		0,
		kcache.ResourceEventHandlerFuncs{
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

func watchEndpoints(client *kclient.Clientset, ch chan<- *kapi.Endpoints, stopCh chan struct{}) kcache.Store {
	store, controller := kcache.NewInformer(
		epListWatch(client),
		&kapi.Endpoints{},
		0,
		kcache.ResourceEventHandlerFuncs{
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
