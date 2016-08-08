package kubegen

import (
	kapi "k8s.io/kubernetes/pkg/api"
)

type Context struct {
	Pods      []kapi.Pod
	Services  []kapi.Service
	Endpoints []kapi.Endpoints
}
