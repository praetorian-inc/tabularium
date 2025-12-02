package model

import (
	"fmt"
	"reflect"
	"strings"
)

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
