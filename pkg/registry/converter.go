package registry

import "fmt"

// ConverterFunc converts JSON-encoded capmodel data into a registered Model.
type ConverterFunc func(data []byte) (Model, error)

// RegisterConverter registers a converter function for the given type name.
func (r *TypeRegistry) RegisterConverter(name string, fn ConverterFunc) error {
	if _, ok := r.converters[name]; ok {
		return fmt.Errorf("converter %s already registered", name)
	}
	r.converters[name] = fn
	return nil
}

// MustRegisterConverter registers a converter function, panicking on failure.
func (r *TypeRegistry) MustRegisterConverter(name string, fn ConverterFunc) {
	if err := r.RegisterConverter(name, fn); err != nil {
		panic(err)
	}
}

// Convert looks up and invokes the converter for the given type name.
func (r *TypeRegistry) Convert(name string, data []byte) (Model, error) {
	fn, ok := r.converters[name]
	if !ok {
		return nil, fmt.Errorf("no converter registered for %s", name)
	}
	return fn(data)
}
