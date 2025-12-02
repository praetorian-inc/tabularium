package model

import (
	"encoding/gob"
	"fmt"
	"reflect"
	"strings"
)

// TypeRegistry holds information about all registered types
type TypeRegistry struct {
	types   map[string]reflect.Type
	aliases map[string]string
}

// NewTypeRegistry creates a new type registry
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types:   make(map[string]reflect.Type),
		aliases: make(map[string]string),
	}
}

// MustRegisterModel registers a model, and panics on failure. Useful for registering models in init()
func (r *TypeRegistry) MustRegisterModel(model Model, aliases ...string) {
	err := r.RegisterModel(model, aliases...)
	if err != nil {
		panic(err)
	}
}

// RegisterModel registers a model type with the registry.
// It returns an error if the type is already registered or if it doesn't
// implement the registry.Model interface.
func (r *TypeRegistry) RegisterModel(model Model, aliases ...string) error {
	gob.Register(model)
	tipe := reflect.TypeOf(model)
	name := Name(model)

	if _, ok := r.types[name]; ok {
		return fmt.Errorf("type %s already registered", name)
	}

	r.types[name] = tipe
	for _, alias := range aliases {
		r.types[strings.ToLower(alias)] = tipe
		r.aliases[strings.ToLower(alias)] = name
	}

	return nil
}

func (r *TypeRegistry) GetAliases(name string) []string {
	aliases := []string{name}
	for alias, n := range r.aliases {
		if n == name {
			aliases = append(aliases, alias)
		}
	}
	return aliases
}

// GetType returns the registered type for a given name
func (r *TypeRegistry) GetType(name string) (reflect.Type, bool) {
	typ, ok := r.types[name]
	return typ, ok
}

// MakeType returns an instance of the registered type for a given name
func (r *TypeRegistry) MakeType(name string) (Model, bool) {
	name = strings.ToLower(name)
	typ, ok := r.types[name]
	if !ok {
		return nil, false
	}

	model := reflect.New(typ.Elem()).Interface().(Model)
	if alias, ok := model.(Alias); ok {
		alias.SetAlias(name)
	}

	return model, true
}

// GetAllTypes returns all registered types
func (r *TypeRegistry) GetAllTypes() map[string]reflect.Type {
	return r.types
}
