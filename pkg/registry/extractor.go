package registry

import "fmt"

// Extractable is an optional interface that models can implement to populate
// derived fields before extraction. For example, Port uses this to reconstruct
// its Parent from the Source key when the Parent isn't present.
type Extractable interface {
	PrepareForExtract()
}

// ExtractorFunc converts a registered Model into its simplified capmodel representation.
type ExtractorFunc func(m Model) (any, error)

// RegisterExtractor registers an extractor function for the given type name.
func (r *TypeRegistry) RegisterExtractor(name string, fn ExtractorFunc) error {
	if _, ok := r.extractors[name]; ok {
		return fmt.Errorf("extractor %s already registered", name)
	}
	r.extractors[name] = fn
	return nil
}

// MustRegisterExtractor registers an extractor function, panicking on failure.
func (r *TypeRegistry) MustRegisterExtractor(name string, fn ExtractorFunc) {
	if err := r.RegisterExtractor(name, fn); err != nil {
		panic(err)
	}
}

// Extract looks up and invokes the extractor for the given type name.
func (r *TypeRegistry) Extract(name string, m Model) (any, error) {
	fn, ok := r.extractors[name]
	if !ok {
		return nil, fmt.Errorf("no extractor registered for %s", name)
	}
	return fn(m)
}
