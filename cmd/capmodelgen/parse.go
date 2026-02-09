package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

var typeMap = map[string]string{
	"SmartBytes":        "[]byte",
	"CloudResourceType": "string",
	"GobSafeBool":       "bool",
}

type parentKind string

const (
	parentInject    parentKind = "inject"    // GraphModelWrapper: wrap with NewGraphModelWrapper, set before hooks
	parentConcrete  parentKind = "concrete"  // concrete pointer: set directly, before hooks
	parentInterface parentKind = "interface" // interface field: set after hooks
)

type field struct {
	SourceFieldName string
	SourceJSONNames []string
	JSONName        string
	GoType          string
}

type parentField struct {
	SourceFieldName string
	JSONName        string
	SlimType        string
	Kind            parentKind
}

type slimType struct {
	Name           string
	SourceTypeName string
	Fields         []field
	Parent         *parentField
}

func parseSlimTags(reg *registry.TypeRegistry) []slimType {
	builders := map[string]*slimType{}
	// Dedup: embedded fields can be visited multiple times when
	// walking different registered types that share the same embed.
	seen := map[string]map[string]bool{}

	for _, name := range sortedRegistryNames(reg) {
		typ := mustGetType(reg, name)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		goTypeName := typ.Name()

		for _, sf := range reflect.VisibleFields(typ) {
			if !sf.IsExported() {
				continue
			}
			slimTag := sf.Tag.Get("capmodel")
			if slimTag == "" {
				continue
			}

			for _, entry := range strings.Split(slimTag, ",") {
				entry = strings.TrimSpace(entry)
				if entry == "" {
					continue
				}

				typeName, jsonName, embedType := parseEntry(entry)

				if seen[typeName] == nil {
					seen[typeName] = map[string]bool{}
				}
				if seen[typeName][sf.Name] {
					continue
				}
				seen[typeName][sf.Name] = true

				sourceJSONName := jsonTagName(sf)
				if jsonName == "" {
					jsonName = sourceJSONName
				}

				b, ok := builders[typeName]
				if !ok {
					b = &slimType{
						Name:           typeName,
						SourceTypeName: resolveSourceTypeName(reg, typeName, goTypeName),
					}
					builders[typeName] = b
				}

				if embedType != "" {
					b.Parent = &parentField{
						SourceFieldName: sf.Name,
						JSONName:        jsonName,
						SlimType:        embedType,
						Kind:            resolveParentKind(sf.Type),
					}
					continue
				}

				// Merge: multiple source fields can map to the same slim JSON name
				// (e.g., DNS and Name both map to "ip" in the IP slim type).
				if mergeField(b, jsonName, sourceJSONName) {
					continue
				}

				b.Fields = append(b.Fields, field{
					SourceFieldName: sf.Name,
					SourceJSONNames: []string{sourceJSONName},
					JSONName:        jsonName,
					GoType:          resolveGoType(sf.Type),
				})
			}
		}
	}

	result := make([]slimType, 0, len(builders))
	for _, b := range builders {
		result = append(result, *b)
	}
	return result
}

func resolveParentKind(t reflect.Type) parentKind {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch {
	case t.Kind() == reflect.Interface:
		return parentInterface
	case t.Name() == "GraphModelWrapper":
		return parentInject
	default:
		return parentConcrete
	}
}

// resolveSourceTypeName looks up the Go type name for a slim type. Falls back to
// the declaring struct when the slim type isn't a registered model (e.g., IP → Asset).
func resolveSourceTypeName(reg *registry.TypeRegistry, slimTypeName, fallback string) string {
	if typ, ok := reg.GetType(strings.ToLower(slimTypeName)); ok {
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		return typ.Name()
	}
	return fallback
}

func mergeField(b *slimType, jsonName, sourceJSONName string) bool {
	for i, existing := range b.Fields {
		if existing.JSONName == jsonName {
			for _, name := range b.Fields[i].SourceJSONNames {
				if name == sourceJSONName {
					return true
				}
			}
			b.Fields[i].SourceJSONNames = append(b.Fields[i].SourceJSONNames, sourceJSONName)
			return true
		}
	}
	return false
}

// sortedRegistryNames returns canonical (non-alias) names in sorted order.
// A name is canonical if it equals strings.ToLower of the Go type name.
func sortedRegistryNames(reg *registry.TypeRegistry) []string {
	var names []string
	for name, typ := range reg.GetAllTypes() {
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		if name == strings.ToLower(typ.Name()) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func mustGetType(reg *registry.TypeRegistry, name string) reflect.Type {
	typ, ok := reg.GetType(name)
	if !ok {
		panic("registered type not found: " + name)
	}
	return typ
}

// parseEntry parses a slim tag entry: "TypeName[=jsonname[(EmbedType)]]"
func parseEntry(entry string) (typeName, jsonName, embedType string) {
	parts := strings.SplitN(entry, "=", 2)
	typeName = parts[0]
	if len(parts) == 2 {
		rest := parts[1]
		if idx := strings.Index(rest, "("); idx != -1 {
			jsonName = rest[:idx]
			embedType = strings.TrimSuffix(rest[idx+1:], ")")
		} else {
			jsonName = rest
		}
	}
	return
}

func jsonTagName(sf reflect.StructField) string {
	tag := sf.Tag.Get("json")
	if tag == "" || tag == "-" {
		return strings.ToLower(sf.Name)
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}

func resolveGoType(t reflect.Type) string {
	if t.Kind() != reflect.Ptr {
		if mapped, ok := typeMap[t.Name()]; ok {
			return mapped
		}
	}
	switch t.Kind() {
	case reflect.Ptr:
		return "*" + resolveGoType(t.Elem())
	case reflect.Slice:
		return "[]" + resolveGoType(t.Elem())
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", resolveGoType(t.Key()), resolveGoType(t.Elem()))
	case reflect.Interface:
		return "any"
	default:
		return t.Name()
	}
}
