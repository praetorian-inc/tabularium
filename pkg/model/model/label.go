package model

import (
	"fmt"
	"strings"
	"sync"
)

// LabelRegistry manages the mapping of lowercase keys to properly-cased labels
type LabelRegistry struct {
	mu     sync.RWMutex
	labels map[string]string
}

// globalRegistry is the singleton instance of LabelRegistry
var globalRegistry = &LabelRegistry{
	labels: make(map[string]string),
}

// NewLabel creates a new label and registers it in the global registry
func NewLabel(value string) string {
	globalRegistry.MustRegister(value)
	return value
}

// MustRegister adds a label to the registry with a lowercase key
// Panics if a label with the same lowercase key already exists
func (r *LabelRegistry) MustRegister(label string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	key := strings.ToLower(label)
	if existing, exists := r.labels[key]; exists && existing != label {
		panic(fmt.Sprintf("label collision: cannot register %q because %q is already registered with key %q", 
			label, existing, key))
	}
	r.labels[key] = label
}

// Get retrieves a label from the registry using a case-insensitive key
func (r *LabelRegistry) Get(key string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	lowercaseKey := strings.ToLower(key)
	label, exists := r.labels[lowercaseKey]
	return label, exists
}

// List returns all registered labels
func (r *LabelRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]string, 0, len(r.labels))
	for _, label := range r.labels {
		result = append(result, label)
	}
	return result
}

// GetLabelRegistry returns the global LabelRegistry instance
func GetLabelRegistry() *LabelRegistry {
	return globalRegistry
}