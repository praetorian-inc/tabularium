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
// The bool return value indicates whether the parent should be injected into
// the child JSON as a GraphModelWrapper-compatible "parent" object. Types like
// SlimPort and SlimAttribute return true because their full models use
// GraphModelWrapper. SlimWebpage returns false because Webpage.Parent is
// *WebApplication, not GraphModelWrapper.
type parentAssetProvider interface {
	GetParentAsset() (SlimAsset, bool)
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
	// optionally inject it into the child's JSON payload so that hooks can
	// reference it. The inject bool controls whether the parent is added as
	// a GraphModelWrapper-compatible "parent" field in the child JSON.
	var parentJSON []byte
	var inject bool
	if p, ok := slim.(parentAssetProvider); ok {
		parentAsset, shouldInject := p.GetParentAsset()
		inject = shouldInject

		parentModel, injectionJSON, err := convertParent(parentAsset)
		if err != nil {
			return nil, err
		}

		col.Add(parentModel)
		parentJSON = injectionJSON
	}

	// Marshal the slim type itself.
	b, err := json.Marshal(slim)
	if err != nil {
		return nil, fmt.Errorf("slim: marshal: %w", err)
	}

	// If we have a parent and the slim type opts into parent injection,
	// inject it into the child JSON as a GraphModelWrapper-compatible
	// "parent" object so that the child hooks can resolve Parent.Model
	// (e.g., for key construction).
	if parentJSON != nil && inject {
		b, err = injectParent(b, parentJSON, "asset")
		if err != nil {
			return nil, fmt.Errorf("slim: inject parent: %w", err)
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

// convertParent converts a SlimAsset into a full model and returns the model
// along with its JSON representation (for optional injection into the child).
func convertParent(parentAsset SlimAsset) (registry.Model, []byte, error) {
	m, ok := registry.Registry.MakeType(parentAsset.TargetModel())
	if !ok {
		return nil, nil, fmt.Errorf("slim: unknown parent model %q", parentAsset.TargetModel())
	}
	pb, err := json.Marshal(parentAsset)
	if err != nil {
		return nil, nil, fmt.Errorf("slim: marshal parent: %w", err)
	}
	if err := registry.UnmarshalModel(pb, m); err != nil {
		return nil, nil, fmt.Errorf("slim: unmarshal parent: %w", err)
	}
	injectionJSON, err := json.Marshal(m)
	if err != nil {
		return nil, nil, fmt.Errorf("slim: re-marshal parent: %w", err)
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
