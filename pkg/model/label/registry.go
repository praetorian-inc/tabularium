package label

import (
	"fmt"
	"strings"
	"sync"
)

type Registry struct {
	mu     sync.RWMutex
	labels map[string]string
}

var (
	globalRegistry *Registry
	registerOnce   sync.Once
)

func (r *Registry) mustRegister(label string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := strings.ToLower(label)
	if existing, exists := r.labels[key]; exists {
		panic(fmt.Sprintf("label collision: cannot register %q because %q is already registered with key %q",
			label, existing, key))
	}
	r.labels[key] = label
}

func (r *Registry) Get(key string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lowercaseKey := strings.ToLower(key)
	label, exists := r.labels[lowercaseKey]
	return label, exists
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, 0, len(r.labels))
	for _, label := range r.labels {
		result = append(result, label)
	}
	return result
}

func GetRegistry() *Registry {
	registerOnce.Do(func() {
		globalRegistry = &Registry{
			labels: make(map[string]string),
		}
	})

	return globalRegistry
}
