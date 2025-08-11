package registry

import (
	"encoding/gob"
	"fmt"
	"reflect"
	"strings"
)

// Registry is a singleton type registry for this process
var Registry *TypeRegistry

// init sets up the singleton registry
func init() {
	Registry = NewTypeRegistry()
}

func GenericName(item any) (string, error) {
	tipe := reflect.TypeOf(item)

	model := new(Model)
	rType := reflect.TypeOf(model).Elem()
	if !tipe.Implements(rType) {
		return "", fmt.Errorf("type %q does not implement Model", tipe.Name())
	}

	return strings.ToLower(tipe.Elem().Name()), nil
}

func Name(model Model) string {
	tipe := reflect.TypeOf(model)
	if tipe.Kind() == reflect.Ptr {
		tipe = tipe.Elem()
	}
	return strings.ToLower(tipe.Name())
}

// TypeRegistry holds information about all registered types
type TypeRegistry struct {
	types  map[string]reflect.Type
	labels map[string]string
}

// NewTypeRegistry creates a new type registry
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types:  make(map[string]reflect.Type),
		labels: make(map[string]string),
	}
}

type labeled interface {
	GetLabels() []string
}

// RegisterModel registers a model type with the registry.
// It returns an error if the type is already registered or if it doesn't
// implement the registry.Model interface.
func (r *TypeRegistry) RegisterModel(model Model) error {
	gob.Register(model)
	tipe := reflect.TypeOf(model)
	name := Name(model)

	if _, ok := r.types[name]; ok {
		return fmt.Errorf("type %s already registered", name)
	}

	if labeled, ok := model.(labeled); ok {
		if err := r.registerLabels(labeled.GetLabels()); err != nil {
			return err
		}
	}

	r.types[name] = tipe
	return nil
}

func (r *TypeRegistry) registerLabels(labels []string) error {
	for _, label := range labels {
		if label == "" {
			continue
		}

		if registered, ok := r.labels[label]; ok && registered != label {
			return fmt.Errorf("label %q already registered as %q", label, registered)
		}

		r.labels[strings.ToLower(label)] = label
	}

	return nil
}

func (r *TypeRegistry) FormatLabel(label string) (string, bool) {
	label = strings.ToLower(label)
	registered, ok := r.labels[label]
	return registered, ok
}

// MustRegisterModel registers a model, and panics on failure. Useful for registering models in init()
func (r *TypeRegistry) MustRegisterModel(model Model) {
	err := r.RegisterModel(model)
	if err != nil {
		panic(err)
	}
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
	return reflect.New(typ.Elem()).Interface().(Model), true
}

// GetAllTypes returns all registered types
func (r *TypeRegistry) GetAllTypes() map[string]reflect.Type {
	return r.types
}

// GetTypes retrieves all type names from a registry that have type T, or implement T
func GetTypes[T Model](r *TypeRegistry) []string {
	out := []string{}
	for name, tipe := range r.GetAllTypes() {
		tt := reflect.TypeOf((*T)(nil)).Elem()
		if tt.AssignableTo(tipe) || (tt.Kind() == reflect.Interface && tipe.Implements(tt)) {
			out = append(out, name)
		}
	}
	return out
}
