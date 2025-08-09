package model

import (
	"strings"
	"sync"
)

// Label represents a Neo4j label with preserved casing
type Label string

// String returns the string representation of the Label
func (l Label) String() string {
	return string(l)
}

// LabelRegistry manages the mapping of lowercase keys to properly-cased labels
type LabelRegistry struct {
	mu     sync.RWMutex
	labels map[string]Label
}

// globalRegistry is the singleton instance of LabelRegistry
var globalRegistry = &LabelRegistry{
	labels: make(map[string]Label),
}

// NewLabel creates a new Label and registers it in the global registry
func NewLabel(value string) Label {
	label := Label(value)
	globalRegistry.Register(label)
	return label
}

// Register adds a label to the registry with a lowercase key
func (r *LabelRegistry) Register(label Label) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	key := strings.ToLower(string(label))
	r.labels[key] = label
}

// Get retrieves a label from the registry using a case-insensitive key
func (r *LabelRegistry) Get(key string) *Label {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	lowercaseKey := strings.ToLower(key)
	label, exists := r.labels[lowercaseKey]
	if !exists {
		return nil
	}
	return &label
}

// GetOrCreate retrieves a label from the registry or creates it if it doesn't exist
func (r *LabelRegistry) GetOrCreate(lowercaseKey string, properCaseValue string) Label {
	lowercaseKey = strings.ToLower(lowercaseKey)
	
	// Try to get existing label first (with read lock)
	r.mu.RLock()
	if label, exists := r.labels[lowercaseKey]; exists {
		r.mu.RUnlock()
		return label
	}
	r.mu.RUnlock()
	
	// Need to create new label (with write lock)
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Double-check in case another goroutine created it
	if label, exists := r.labels[lowercaseKey]; exists {
		return label
	}
	
	// Create and register new label
	label := Label(properCaseValue)
	r.labels[lowercaseKey] = label
	return label
}

// List returns all registered labels
func (r *LabelRegistry) List() []Label {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]Label, 0, len(r.labels))
	for _, label := range r.labels {
		result = append(result, label)
	}
	return result
}

// Clear removes all labels from the registry
func (r *LabelRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.labels = make(map[string]Label)
}

// GetLabelRegistry returns the global LabelRegistry instance
func GetLabelRegistry() *LabelRegistry {
	return globalRegistry
}

// LabelsToStrings converts a slice of Labels to a slice of strings
func LabelsToStrings(labels []Label) []string {
	if labels == nil {
		return nil
	}
	
	result := make([]string, len(labels))
	for i, label := range labels {
		result[i] = string(label)
	}
	return result
}

// StringsToLabels converts a slice of strings to a slice of Labels
func StringsToLabels(strings []string) []Label {
	if strings == nil {
		return nil
	}
	
	result := make([]Label, len(strings))
	for i, str := range strings {
		result[i] = NewLabel(str)
	}
	return result
}