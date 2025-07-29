package registry

import (
	"encoding/json"
	"reflect"
)

// Model defines the interface that all registered data models must implement.
type Model interface {
	// GetDescription returns a brief description of the model's purpose.
	GetDescription() string
	// GetHooks returns a list of hooks for this model
	GetHooks() []Hook
	// Defaulted sets the fields of this model to their default values
	Defaulted()
	// GetKey returns the key of the model. the model key should uniquely identify this entity and be contained within the field `Key`
	GetKey() string
}

// BaseModel provides default implementations of optional model methods
// Models are expected to implement any required methods explicitly
type BaseModel struct{}

func (BaseModel) GetKey() string {
	return ""
}

func (BaseModel) GetHooks() []Hook {
	return []Hook{}
}

func (BaseModel) Defaulted() {}

// UnmarshalModel unmarshals a model, by:
//   - setting its default field values
//   - unmarshalling into the model
//   - calling the model's hooks
//   - recursively calling hooks on any submodels
func UnmarshalModel(b []byte, model Model) error {
	defaultModel(model)
	err := json.Unmarshal(b, model)
	if err != nil {
		return err
	}
	return callHooks(model)
}

func defaultModel(model Model) {
	model.Defaulted()
	for _, submodel := range submodels(model) {
		defaultModel(submodel)
	}
}

func callHooks(model Model) error {
	err := CallHooks(model)
	if err != nil {
		return err
	}
	for _, submodel := range submodels(model) {
		err = callHooks(submodel)
		if err != nil {
			return err
		}
	}
	return nil
}

func submodels(model Model) []Model {
	val := reflect.ValueOf(model)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	var result []Model
	modelType := reflect.TypeOf((*Model)(nil)).Elem()
	valType := val.Type()

	for _, structField := range reflect.VisibleFields(valType) {
		if !structField.IsExported() {
			continue
		}

		field := val.FieldByIndex(structField.Index)
		fieldType := field.Type()

		if (field.Kind() == reflect.Pointer || field.Kind() == reflect.Interface) && field.IsZero() {
			continue
		}

		if fieldType.Implements(modelType) {
			result = append(result, field.Interface().(Model))
		} else if reflect.PointerTo(fieldType).Implements(modelType) {
			if field.CanAddr() {
				ptrVal := field.Addr()
				result = append(result, ptrVal.Interface().(Model))
			}
		}
	}

	return result
}
