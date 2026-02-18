package registry

import (
	"fmt"
	"sync"
)

type ToolFunc func(args string) (string, error)

type Registry struct {
	tools map[string]ToolFunc
	mu    sync.RWMutex
}

func New() *Registry {
	return &Registry{
		tools: make(map[string]ToolFunc),
	}
}

func (r *Registry) Register(name string, fn ToolFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[name] = fn
}

func (r *Registry) Execute(name string, args string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("tool %s not found", name)
	}
	return fn(args)
}
