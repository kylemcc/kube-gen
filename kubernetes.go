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
			AddFunc:    func(v interface{}) { ch <- v.(*kapi.Pod) },
			UpdateFunc: func(ov, nv interface{}) { ch <- ov.(*kapi.Pod) },
			DeleteFunc: func(v interface{}) { ch <- v.(*kapi.Pod) },
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
			AddFunc:    func(v interface{}) { ch <- v.(*kapi.Service) },
			UpdateFunc: func(ov, nv interface{}) { ch <- nv.(*kapi.Service) },
			DeleteFunc: func(v interface{}) { ch <- v.(*kapi.Service) },
		})
	go controller.Run(stopCh)
	return store
}

func watchEndpoints(client *kclient.Client, ch chan<- *kapi.Endpoints, stopCh chan struct{}) kcache.Store {
	store, controller := kframework.NewInformer(
		svcListWatch(client),
		&kapi.Endpoints{},
		0,
		kframework.ResourceEventHandlerFuncs{
			AddFunc:    func(v interface{}) { ch <- v.(*kapi.Endpoints) },
			UpdateFunc: func(ov, nv interface{}) { ch <- nv.(*kapi.Endpoints) },
			DeleteFunc: func(v interface{}) { ch <- v.(*kapi.Endpoints) },
		})
	go controller.Run(stopCh)
	return store
}
