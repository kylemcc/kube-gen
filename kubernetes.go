package kubegen

import (
	kapi "k8s.io/api/core/v1"
	kselector "k8s.io/apimachinery/pkg/fields"
	kclient "k8s.io/client-go/kubernetes"
	krest "k8s.io/client-go/rest"
	kcache "k8s.io/client-go/tools/cache"
	kcmd "k8s.io/client-go/tools/clientcmd"
)

// Initializes a new Kubernetes API Client
func newKubeClient(c Config) (*kclient.Clientset, error) {
	var config *krest.Config
	var err error
	if c.Host == "" && !c.UseInClusterConfig {
		// use the current context in kubeconfig
		config, err = kcmd.BuildConfigFromFlags("", c.Kubeconfig)
		if err != nil {
			return nil, err
		}
	} else if c.UseInClusterConfig {
		config, err = krest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		config = &krest.Config{
			Host:          c.Host,
			ContentConfig: krest.ContentConfig{GroupVersion: &kapi.SchemeGroupVersion},
		}
	}
	return kclient.NewForConfig(config)
}

func podsListWatch(client *kclient.Clientset) *kcache.ListWatch {
	return kcache.NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", kapi.NamespaceAll, kselector.Everything())
}

func svcListWatch(client *kclient.Clientset) *kcache.ListWatch {
	return kcache.NewListWatchFromClient(client.CoreV1().RESTClient(), "services", kapi.NamespaceAll, kselector.Everything())
}

func epListWatch(client *kclient.Clientset) *kcache.ListWatch {
	return kcache.NewListWatchFromClient(client.CoreV1().RESTClient(), "endpoints", kapi.NamespaceAll, kselector.Everything())
}

func watchPods(client *kclient.Clientset, ch chan<- *kapi.Pod, stopCh chan struct{}) kcache.Store {
	store, controller := kcache.NewInformer(
		podsListWatch(client),
		&kapi.Pod{},
		0,
		kcache.ResourceEventHandlerFuncs{
			AddFunc: func(v any) {
				if p, ok := v.(*kapi.Pod); ok {
					ch <- p
				}
			},
			UpdateFunc: func(ov, nv any) {
				if p, ok := nv.(*kapi.Pod); ok {
					ch <- p
				}
			},
			DeleteFunc: func(v any) {
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
			AddFunc: func(v any) {
				if s, ok := v.(*kapi.Service); ok {
					ch <- s
				}
			},
			UpdateFunc: func(ov, nv any) {
				if s, ok := nv.(*kapi.Service); ok {
					ch <- s
				}
			},
			DeleteFunc: func(v any) {
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
			AddFunc: func(v any) {
				if e, ok := v.(*kapi.Endpoints); ok {
					ch <- e
				}
			},
			UpdateFunc: func(ov, nv any) {
				if e, ok := nv.(*kapi.Endpoints); ok {
					ch <- e
				}
			},
			DeleteFunc: func(v any) {
				if e, ok := v.(*kapi.Endpoints); ok {
					ch <- e
				}
			},
		})
	go controller.Run(stopCh)
	return store
}

// IsPodReady returns true if a pod is ready; false otherwise.
func IsPodReady(pod *kapi.Pod) bool {
	return isPodReadyConditionTrue(pod.Status)
}

// IsPodReadyConditionTrue returns true if a pod is ready; false otherwise.
func isPodReadyConditionTrue(status kapi.PodStatus) bool {
	condition := getPodReadyCondition(status)
	return condition != nil && condition.Status == kapi.ConditionTrue
}

// GetPodReadyCondition extracts the pod ready condition from the given status and returns that.
// Returns nil if the condition is not present.
func getPodReadyCondition(status kapi.PodStatus) *kapi.PodCondition {
	_, condition := getPodCondition(&status, kapi.PodReady)
	return condition
}

// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func getPodCondition(status *kapi.PodStatus, conditionType kapi.PodConditionType) (int, *kapi.PodCondition) {
	if status == nil {
		return -1, nil
	}
	return getPodConditionFromList(status.Conditions, conditionType)
}

// GetPodConditionFromList extracts the provided condition from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the condition is not present.
func getPodConditionFromList(conditions []kapi.PodCondition, conditionType kapi.PodConditionType) (int, *kapi.PodCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}
