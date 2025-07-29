package collection

import (
	"log/slog"
	"reflect"

	"github.com/praetorian-inc/tabularium/pkg/lib/plural"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// Collection is a universal container type for registered types
type Collection struct {
	Items map[string][]registry.Model `json:"items"`
	Count int                         `json:"count"`
}

// init lazily initializes Collection
func (c *Collection) init() {
	if c.Items == nil {
		c.Items = map[string][]registry.Model{}
	}
}

func (c *Collection) Add(model registry.Model) {
	c.init()
	name := plural.Plural(registry.Name(model))
	c.Items[name] = append(c.Items[name], model)
	c.Count++
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
