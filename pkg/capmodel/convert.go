//go:generate go run ../../cmd/capmodelgen -output .

package capmodel

import (
	"encoding/json"
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/collection"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// Converter is implemented by capability model types to declare their target model name.
type Converter interface {
	TargetModel() string // returns the registry name, e.g., "asset", "port"
}

// jsonProvider is optionally implemented by capability model types that need
// custom JSON serialization for Convert. This enables types with ergonomic
// field names (e.g., IP.Address) to map them to the underlying model's JSON
// shape (e.g., {"dns":"...","name":"..."}).
type jsonProvider interface {
	ConvertJSON() ([]byte, error)
}

// parentAssetProvider is implemented by capability model types that embed a parent Asset.
// When Convert encounters a type that implements this interface, it converts the
// parent Asset into a full Asset and adds it to the resulting Collection.
// The bool return value indicates whether the parent should be injected into
// the child JSON as a GraphModelWrapper-compatible "parent" object. Types like
// Port and Attribute return true because their full models use
// GraphModelWrapper. Webpage returns false because Webpage.Parent is
// *WebApplication, not GraphModelWrapper.
type parentAssetProvider interface {
	GetParentAsset() (Asset, bool)
}

// Convert takes a capability model type, marshals it to JSON, and converts it to a full
// Tabularium model via UnmarshalModel (which calls Defaulted + hooks).
// Returns a Collection containing the converted model(s).
//
// Use collection.Get[*model.Asset](col) to retrieve typed models from the
// returned Collection.
//
// When the capability model type implements parentAssetProvider (e.g., Port, Attribute,
// Webpage), Convert first converts the embedded Asset into a full Asset,
// adds it to the Collection, then injects the parent into the child model's JSON
// so that the child's hooks can compute its Key and Source correctly.
func Convert(cm Converter) (*collection.Collection, error) {
	col := &collection.Collection{}

	// If the capability model type has a parent asset, convert the parent first and
	// optionally inject it into the child's JSON payload so that hooks can
	// reference it. The inject bool controls whether the parent is added as
	// a GraphModelWrapper-compatible "parent" field in the child JSON.
	var parentJSON []byte
	var inject bool
	if p, ok := cm.(parentAssetProvider); ok {
		parentAsset, shouldInject := p.GetParentAsset()
		inject = shouldInject

		parentModel, injectionJSON, err := convertParent(parentAsset)
		if err != nil {
			return nil, err
		}

		col.Add(parentModel)
		parentJSON = injectionJSON
	}

	// Marshal the capability model type itself. Types that implement
	// jsonProvider supply their own JSON (mapping ergonomic fields to the
	// underlying model's JSON shape); all others use standard json.Marshal.
	var b []byte
	var err error
	if jp, ok := cm.(jsonProvider); ok {
		b, err = jp.ConvertJSON()
	} else {
		b, err = json.Marshal(cm)
	}
	if err != nil {
		return nil, fmt.Errorf("capmodel: marshal: %w", err)
	}

	// If we have a parent and the capability model type opts into parent injection,
	// inject it into the child JSON as a GraphModelWrapper-compatible
	// "parent" object so that the child hooks can resolve Parent.Model
	// (e.g., for key construction).
	if parentJSON != nil && inject {
		b, err = injectParent(b, parentJSON, "asset")
		if err != nil {
			return nil, fmt.Errorf("capmodel: inject parent: %w", err)
		}
	}

	model, ok := registry.Registry.MakeType(cm.TargetModel())
	if !ok {
		return nil, fmt.Errorf("capmodel: unknown model %q", cm.TargetModel())
	}

	if err := registry.UnmarshalModel(b, model); err != nil {
		return nil, fmt.Errorf("capmodel: unmarshal model: %w", err)
	}

	col.Add(model)
	return col, nil
}

// convertParent converts an Asset into a full model and returns the model
// along with its JSON representation (for optional injection into the child).
func convertParent(parentAsset Asset) (registry.Model, []byte, error) {
	m, ok := registry.Registry.MakeType(parentAsset.TargetModel())
	if !ok {
		return nil, nil, fmt.Errorf("capmodel: unknown parent model %q", parentAsset.TargetModel())
	}
	pb, err := json.Marshal(parentAsset)
	if err != nil {
		return nil, nil, fmt.Errorf("capmodel: marshal parent: %w", err)
	}
	if err := registry.UnmarshalModel(pb, m); err != nil {
		return nil, nil, fmt.Errorf("capmodel: unmarshal parent: %w", err)
	}
	injectionJSON, err := json.Marshal(m)
	if err != nil {
		return nil, nil, fmt.Errorf("capmodel: re-marshal parent: %w", err)
	}
	return m, injectionJSON, nil
}

// injectParent takes the child JSON (a compact JSON object from json.Marshal)
// and appends a "parent" field containing a GraphModelWrapper-compatible
// object: {"type": parentType, "model": parentJSON}.
func injectParent(childJSON, parentJSON []byte, parentType string) ([]byte, error) {
	if len(childJSON) < 2 || childJSON[len(childJSON)-1] != '}' {
		return nil, fmt.Errorf("injectParent: expected JSON object, got %q", childJSON)
	}
	// Build: ,"parent":{"type":"<parentType>","model":<parentJSON>}}
	// childJSON is compact (from json.Marshal), so we trim the trailing '}'.
	var buf []byte
	buf = append(buf, childJSON[:len(childJSON)-1]...)
	buf = append(buf, `,"parent":{"type":"`...)
	buf = append(buf, parentType...)
	buf = append(buf, `","model":`...)
	buf = append(buf, parentJSON...)
	buf = append(buf, "}}"...)
	return buf, nil
}
