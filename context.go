package kubegen

import (
	"os"
	"strings"
	"sync"

	kapi "k8s.io/kubernetes/pkg/api"
)

var (
	envOnce sync.Once
	envMap  map[string]string
)

type Context struct {
	Pods      []kapi.Pod
	Services  []kapi.Service
	Endpoints []kapi.Endpoints
}

// TODO: if running in k8s, make annotations on containing pod available
func (c *Context) Env() map[string]string {
	envOnce.Do(loadEnv)
	return envMap
}

func loadEnv() {
	env := os.Environ()
	envMap = make(map[string]string, len(env))
	for _, v := range env {
		vs := strings.SplitN(v, "=", 2)
		envMap[vs[0]] = vs[1]
	}
}
