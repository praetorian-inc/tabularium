package model

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// Flag - praetorian-controlled booleans used to enable/disable features
type Flag struct {
	registry.BaseModel
	baseTableModel
	Username string `dynamodbav:"username" json:"username" desc:"Chariot username associated with the flag." example:"user@example.com"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the flag." example:"#flag#feature-x"`
	Name     string `dynamodbav:"name" json:"name" desc:"Name of the feature flag." example:"feature-x"`
}

func init() {
	registry.Registry.MustRegisterModel(&Flag{})
}

func (f *Flag) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				f.Key = fmt.Sprintf("#flag#%s", f.Name)
				return nil
			},
		},
	}
}

func NewFlag(name string) Flag {
	f := Flag{
		Name: name,
	}
	f.Defaulted()
	registry.CallHooks(&f)
	return f
}

// GetDescription returns a description for the Flag model.
func (f *Flag) GetDescription() string {
	return "Represents a flag or notable marker on an asset or finding, often indicating proof of compromise or significance."
}
