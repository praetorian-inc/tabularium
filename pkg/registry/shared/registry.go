package shared

import "github.com/praetorian-inc/tabularium/pkg/registry/model"

// Registry is a singleton type registry for this process
var Registry *model.TypeRegistry

// init sets up the singleton registry
func init() {
	Registry = model.NewTypeRegistry()
}
