package callback

import (
	"context"
	"fmt"
	"sync"
)

type Registry struct {
	mu        sync.RWMutex
	callbacks map[string]Func
}

func NewRegistry() *Registry {
	return &Registry{
		callbacks: make(map[string]Func),
	}
}

func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	
	r.MustRegister("log-result", LogResult)
	r.MustRegister("noop", NoOp)
	return r
}

func (r *Registry) Register(name string, fn Func) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.callbacks[name]; exists {
		return fmt.Errorf("callback %q is already registered", name)
	}

	r.callbacks[name] = fn
	return nil
}

func (r *Registry) MustRegister(name string, fn Func) {
	if err := r.Register(name, fn); err != nil {
		panic(err)
	}
}

func (r *Registry) Lookup(name string) (Func, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, ok := r.callbacks[name]
	return fn, ok
}

func (r *Registry) Invoke(ctx context.Context, name string, result Result) error {
	fn, ok := r.Lookup(name)
	if !ok {
		return fmt.Errorf("callback %q not found", name)
	}
	return fn(ctx, result)
}
