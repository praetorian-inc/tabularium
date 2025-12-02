package collection

import (
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/lib/plural"
	modelpkg "github.com/praetorian-inc/tabularium/pkg/model/model"
)

// Collection is a universal container type for registered types
type Collection struct {
	Label string                   `json:"-"`
	Items map[string][]model.Model `json:"items"`
	Count int                      `json:"count"`
}

// init lazily initializes Collection
func (c *Collection) init() {
	if c.Items == nil {
		c.Items = map[string][]model.Model{}
	}
}

func (c *Collection) Add(m model.Model) {
	c.init()

	if ok := addInterface[modelpkg.Seedable](c, m); ok {
		return
	}

	name := plural.Plural(model.Name(m))
	c.Items[name] = append(c.Items[name], m)
	c.Count++
}

func addInterface[T model.Model](c *Collection, model model.Model) bool {
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

func hasLabel(c *Collection, model model.Model) bool {
	graphModel, ok := model.(modelpkg.GraphModel)
	if !ok {
		return false
	}
	return c.Label != "" && slices.Contains(graphModel.GetLabels(), c.Label)
}

// Get retrieves all Items from a collection that have type T, or implement T
func Get[T model.Model](c *Collection) []T {
	c.init()
	out := []T{}
	for _, name := range model.GetTypes[T](shared.Registry) {
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
