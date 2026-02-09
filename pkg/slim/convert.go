//go:generate go run ../../cmd/slimgen -output .

package slim

import (
	"encoding/json"
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/collection"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// Converter is implemented by slim types to declare their target model name.
type Converter interface {
	TargetModel() string // returns the registry name, e.g., "asset", "port"
}

// parentAssetProvider is implemented by slim types that embed a parent SlimAsset.
// When Convert encounters a type that implements this interface, it converts the
// parent SlimAsset into a full Asset and adds it to the resulting Collection.
type parentAssetProvider interface {
	GetParentAsset() SlimAsset
}

// parentInjectable is a marker interface for slim types whose parent asset
// should be injected into the child JSON as a GraphModelWrapper-compatible
// "parent" object. Types like SlimPort and SlimAttribute implement this because
// their full models use GraphModelWrapper. SlimWebpage does NOT implement it
// because Webpage.Parent is *WebApplication, not GraphModelWrapper.
type parentInjectable interface {
	injectParent()
}

// Convert takes a slim type, marshals it to JSON, and converts it to a full
// Tabularium model via UnmarshalModel (which calls Defaulted + hooks).
// Returns a Collection containing the converted model(s).
//
// Use collection.Get[*model.Asset](col) to retrieve typed models from the
// returned Collection.
//
// When the slim type implements parentAssetProvider (e.g., SlimPort, SlimAttribute,
// SlimWebpage), Convert first converts the embedded SlimAsset into a full Asset,
// adds it to the Collection, then injects the parent into the child model's JSON
// so that the child's hooks can compute its Key and Source correctly.
func Convert(slim Converter) (*collection.Collection, error) {
	col := &collection.Collection{}

	// If the slim type has a parent asset, convert the parent first and
	// inject it into the child's JSON payload so that hooks can reference it.
	var parentJSON []byte
	if p, ok := slim.(parentAssetProvider); ok {
		parentAsset := p.GetParentAsset()

		parentModel, pOK := registry.Registry.MakeType(parentAsset.TargetModel())
		if !pOK {
			return nil, fmt.Errorf("slim: unknown parent model %q", parentAsset.TargetModel())
		}

		pb, err := json.Marshal(parentAsset)
		if err != nil {
			return nil, fmt.Errorf("slim: marshal parent: %w", err)
		}

		if err := registry.UnmarshalModel(pb, parentModel); err != nil {
			return nil, fmt.Errorf("slim: unmarshal parent: %w", err)
		}

		col.Add(parentModel)

		// Marshal the fully-constructed parent so we can inject it into the child JSON.
		parentJSON, err = json.Marshal(parentModel)
		if err != nil {
			return nil, fmt.Errorf("slim: re-marshal parent: %w", err)
		}
	}

	// Marshal the slim type itself.
	b, err := json.Marshal(slim)
	if err != nil {
		return nil, fmt.Errorf("slim: marshal: %w", err)
	}

	// If we have a parent and the slim type opts into parent injection
	// (via the parentInjectable marker), inject it into the child JSON as a
	// GraphModelWrapper-compatible "parent" object so that the child hooks
	// can resolve Parent.Model (e.g., for key construction).
	if parentJSON != nil {
		if _, ok := slim.(parentInjectable); ok {
			b, err = injectParent(b, parentJSON, "asset")
			if err != nil {
				return nil, fmt.Errorf("slim: inject parent: %w", err)
			}
		}
	}

	model, ok := registry.Registry.MakeType(slim.TargetModel())
	if !ok {
		return nil, fmt.Errorf("slim: unknown model %q", slim.TargetModel())
	}

	if err := registry.UnmarshalModel(b, model); err != nil {
		return nil, fmt.Errorf("slim: unmarshal model: %w", err)
	}

	col.Add(model)
	return col, nil
}

// injectParent takes the child JSON (b) and inserts a "parent" field that is a
// GraphModelWrapper-compatible object: {"type": parentType, "model": parentJSON}.
func injectParent(childJSON, parentJSON []byte, parentType string) ([]byte, error) {
	var child map[string]json.RawMessage
	if err := json.Unmarshal(childJSON, &child); err != nil {
		return nil, err
	}

	wrapper := struct {
		Type  string          `json:"type"`
		Model json.RawMessage `json:"model"`
	}{
		Type:  parentType,
		Model: parentJSON,
	}

	wb, err := json.Marshal(wrapper)
	if err != nil {
		return nil, err
	}

	child["parent"] = wb
	return json.Marshal(child)
}
