package collection

import (
	"log/slog"
	"reflect"
	"slices"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/lib/plural"
	modelpkg "github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// Collection is a universal container type for registered types
type Collection struct {
	Label string                      `json:"-"`
	Items map[string][]registry.Model `json:"items"`
	Count int                         `json:"count"`
}

// NewCollectionFromRelationship builds a collection from a relationship's nodes.
func NewCollectionFromRelationship(rel modelpkg.GraphRelationship) *Collection {
	collection := &Collection{}
	if rel == nil {
		return collection
	}
	source, target := rel.Nodes()
	if source != nil {
		collection.Add(source)
	}
	if target != nil {
		collection.Add(target)
	}
	return collection
}

// init lazily initializes Collection
func (c *Collection) init() {
	if c.Items == nil {
		c.Items = map[string][]registry.Model{}
	}
}

func (c *Collection) Add(model registry.Model) {
	c.init()

	if ok := addInterface[modelpkg.Seedable](c, model); ok {
		return
	}

	name := plural.Plural(registry.Name(model))
	c.Items[name] = append(c.Items[name], model)
	c.Count++
}

func addInterface[T registry.Model](c *Collection, model registry.Model) bool {
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	modelType := reflect.TypeOf(model)

	hasLabel := hasLabel(c, model)
	labelMatchesInterface := hasLabel && strings.HasPrefix(interfaceType.Name(), c.Label)

	if !hasLabel || !labelMatchesInterface {
		return false
	}

	label := plural.Plural(c.Label)
	label = strings.ToLower(label)

	if modelType.Implements(interfaceType) {
		c.Items[label] = append(c.Items[label], model)
		c.Count++
		return true
	}
	return false
}

func hasLabel(c *Collection, model registry.Model) bool {
	graphModel, ok := model.(modelpkg.GraphModel)
	if !ok {
		return false
	}
	return c.Label != "" && slices.Contains(graphModel.GetLabels(), c.Label)
}

// Get retrieves all Items from a collection that have type T, or implement T
func Get[T registry.Model](c *Collection) []T {
	c.init()
	out := []T{}
	for _, name := range registry.GetTypes[T](registry.Registry) {
		for _, item := range c.Items[plural.Plural(name)] {
			i, ok := item.(T)
			if !ok {
				slog.Warn("failed to convert item", "requested", name, "type", reflect.TypeOf(item).Name())
				continue
			}
			out = append(out, i)
		}
	}
	return out
}
