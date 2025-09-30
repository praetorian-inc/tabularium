package model

import (
	"fmt"
	"regexp"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&Integration{})
}

type Integration struct {
	BaseAsset
	Name  string `neo4j:"name" json:"name" desc:"Name of the integration." example:"github"`
	Value string `neo4j:"value" json:"value" desc:"Value of the integration." example:"1234567890"`
}

const (
	IntegrationLabel = "Integration"
)

var (
	integrationKey = regexp.MustCompile(`^#integration(#[^#]+){2,}$`)
)

func (i *Integration) GetLabels() []string {
	return []string{IntegrationLabel, AssetLabel, TTLLabel}
}

func (i *Integration) Valid() bool {
	return integrationKey.MatchString(i.Key)
}

func (i *Integration) Identifier() string {
	return i.Value
}

func (i *Integration) Group() string {
	return i.Name
}

func (i *Integration) WithStatus(status string) Target {
	ret := *i
	ret.Status = status
	return &ret
}

func (i *Integration) Defaulted() {
	i.BaseAsset.Defaulted()
	i.Source = AccountSource
}

func (i *Integration) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				i.Key = fmt.Sprintf("#integration#%s#%s", i.Name, i.Value)
				i.Class = i.Name
				return nil
			},
			Description: "Construct the integration key",
		},
		{
			Call: func() error {
				i.BaseAsset.Identifier = i.Value
				i.BaseAsset.Group = i.Name
				return nil
			},
		},
	}
}

func NewIntegration(name, value string) Integration {
	ia := Integration{
		Name:  name,
		Value: value,
	}

	ia.Defaulted()
	registry.CallHooks(&ia)
	return ia
}

// helper to support legacy model.Asset objects for cloud types. Will remove this once the new cloud types are fully migrated.
func NewCloudOrIntegration(name, value string) Assetlike {
	if isCloudProvider(name) {
		a := NewAsset(name, value)
		a.Source = AccountSource
		a.Class = a.GetClass()
		return &a
	} else {
		a := NewIntegration(name, value)
		return &a
	}
}

// helper to branch out for cloud; should be removed in future once cloud types are unified
func isCloudProvider(provider string) bool {
	switch provider {
	case "amazon", "gcp", "azure":
		return true
	default:
		return false
	}
}
