package schema

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	_ "github.com/praetorian-inc/tabularium/pkg/model/model" // Ensure init() functions run
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// GenerateOpenAPISchema creates an OpenAPI v3 document based on registered models.
func GenerateOpenAPISchema() (*openapi3.T, error) {
	types := registry.Registry.GetAllTypes()

	doc := &openapi3.T{
		OpenAPI: "3.1.0",
		Info: &openapi3.Info{
			Title:       "Chariot Unified Data Model",
			Version:     "0.1.1", // Consider making this dynamic if needed
			Description: "Unified Data Model for Chariot",
		},
		Components: &openapi3.Components{
			Schemas: make(openapi3.Schemas),
		},
	}

	// First pass: create all base schema definitions.
	for name := range types {
		openapiSchema := &openapi3.Schema{
			Type:       &openapi3.Types{openapi3.TypeObject},
			Properties: make(openapi3.Schemas),
		}
		doc.Components.Schemas[name] = &openapi3.SchemaRef{
			Value: openapiSchema,
		}
	}

	// Second pass: populate schemas with properties and descriptions.
	for name, typ := range types {
		schemaRef := doc.Components.Schemas[name]
		openapiSchema := schemaRef.Value
		requiredFields := make(map[string]struct{})

		// Need to handle potential pointer types from registry
		if typ.Kind() == reflect.Ptr {
			// This shouldn't happen if registry stores non-pointer types
			typ = typ.Elem()
		}

		if typ.Kind() != reflect.Struct {
			continue
		}

		// Add description from the Model interface (guaranteed by registry)
		// Create a zero-value pointer instance of the type to call the method
		modelPtrInstance := reflect.New(typ).Interface()
		modelWithDesc := modelPtrInstance.(registry.Model) // Type assertion is now safe
		openapiSchema.Description = modelWithDesc.GetDescription()

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			fieldType := field.Type

			if !field.IsExported() {
				continue
			}

			// Handle embedded registered types.
			if field.Anonymous && isRegisteredType(fieldType, types) {
				registeredTypeName := getRegisteredTypeName(fieldType, types)
				embeddedSchemaRef, exists := doc.Components.Schemas[registeredTypeName]
				if !exists || embeddedSchemaRef.Value == nil {
					fmt.Fprintf(os.Stderr, "Warning: Schema for embedded type %s not found or is nil.\n", registeredTypeName)
					continue
				}

				// Copy properties.
				for embeddedPropName, embeddedPropSchemaRef := range embeddedSchemaRef.Value.Properties {
					openapiSchema.Properties[embeddedPropName] = embeddedPropSchemaRef
				}

				// Reflect on the original embedded Go struct for required fields.
				embeddedGoType := fieldType
				if embeddedGoType.Kind() == reflect.Ptr {
					embeddedGoType = embeddedGoType.Elem()
				}
				if embeddedGoType.Kind() == reflect.Struct {
					for j := 0; j < embeddedGoType.NumField(); j++ {
						originalEmbeddedField := embeddedGoType.Field(j)
						if !originalEmbeddedField.IsExported() {
							continue
						}
						jsonFieldName, isOmitempty, include := getJSONInfo(originalEmbeddedField)
						if include && !isOmitempty {
							requiredFields[jsonFieldName] = struct{}{}
						}
					}
				}
				continue
			}

			// Process regular (non-embedded) fields.
			fieldName, isOmitempty, include := getJSONInfo(field)
			if !include {
				continue
			}

			if !isOmitempty {
				requiredFields[fieldName] = struct{}{}
			}

			var fieldSchema *openapi3.Schema

			// Create references for non-embedded registered types.
			if isRegisteredType(fieldType, types) {
				registeredTypeName := getRegisteredTypeName(fieldType, types)
				openapiSchema.Properties[fieldName] = &openapi3.SchemaRef{
					Ref: "#/components/schemas/" + registeredTypeName,
				}
				continue
			}

			// Handle arrays of registered types.
			if (fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array) && isRegisteredType(fieldType.Elem(), types) {
				registeredTypeName := getRegisteredTypeName(fieldType.Elem(), types)
				fieldSchema = &openapi3.Schema{
					Type: &openapi3.Types{openapi3.TypeArray},
					Items: &openapi3.SchemaRef{
						Ref: "#/components/schemas/" + registeredTypeName,
					},
				}
			} else {
				// Handle regular fields and primitive arrays.
				fieldSchema = &openapi3.Schema{
					Type: &openapi3.Types{getOpenAPIType(fieldType)},
				}

				if format := getOpenAPIFormat(fieldType); format != "" {
					fieldSchema.Format = format
				}

				if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
					elemType := fieldType.Elem()
					itemSchema := &openapi3.Schema{
						Type: &openapi3.Types{getOpenAPIType(elemType)},
					}
					if format := getOpenAPIFormat(elemType); format != "" {
						itemSchema.Format = format
					}
					// TODO: Add description/example tags for array items?
					fieldSchema.Items = &openapi3.SchemaRef{
						Value: itemSchema,
					}
				}
			}

			// Add description and example from tags if available.
			if desc := field.Tag.Get("desc"); desc != "" {
				fieldSchema.Description = desc
			}
			if example := field.Tag.Get("example"); example != "" {
				fieldSchema.Example = example
			}

			openapiSchema.Properties[fieldName] = &openapi3.SchemaRef{
				Value: fieldSchema,
			}
		}

		// Add required fields to the schema.
		if len(requiredFields) > 0 {
			finalRequiredList := make([]string, 0, len(requiredFields))
			for reqField := range requiredFields {
				finalRequiredList = append(finalRequiredList, reqField)
			}
			sort.Strings(finalRequiredList) // Ensure consistent order
			openapiSchema.Required = finalRequiredList
		}
	}

	return doc, nil
}

// --- Internal Helper Functions --- //

func getOpenAPIType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map:
		return "object"
	case reflect.Struct:
		return "object"
	case reflect.Ptr:
		return getOpenAPIType(t.Elem())
	case reflect.Interface:
		return "object"
	default:
		// Default to object for unknown types, safer than assuming string.
		return "object"
	}
}

func getOpenAPIFormat(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int32:
		return "int32"
	case reflect.Int64:
		return "int64"
	case reflect.Float32:
		return "float"
	case reflect.Float64:
		return "double"
	case reflect.Struct:
		if t.PkgPath() == "time" && t.Name() == "Time" {
			return "date-time"
		}
	}
	return ""
}

func isRegisteredType(t reflect.Type, types map[string]reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for _, registeredType := range types {
		if t == registeredType {
			return true
		}
	}
	return false
}

func getRegisteredTypeName(t reflect.Type, types map[string]reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for name, registeredType := range types {
		if t == registeredType {
			return name
		}
	}
	return ""
}

func getJSONInfo(field reflect.StructField) (name string, omitempty bool, include bool) {
	jsonTag := field.Tag.Get("json")
	name = field.Name
	include = true

	if jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] == "-" {
			include = false
			return
		}
		if parts[0] != "" {
			name = parts[0]
		}
		for _, part := range parts[1:] {
			if part == "omitempty" {
				omitempty = true
				break
			}
		}
	}
	return
}
