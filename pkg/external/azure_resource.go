package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// AzureResource is a simplified Azure resource for external tool writers.
type AzureResource struct {
	Name          string                  `json:"name"`          // Azure resource name/ID
	AccountRef    string                  `json:"accountRef"`    // Azure account reference
	ResourceType  model.CloudResourceType `json:"resourceType"`  // Type of Azure resource
	Properties    map[string]any          `json:"properties"`    // Resource-specific properties
	ResourceGroup string                  `json:"resourceGroup"` // Azure resource group
}

// Group implements Target interface.
func (a AzureResource) Group() string { return a.AccountRef }

// Identifier implements Target interface.
func (a AzureResource) Identifier() string { return a.Name }

// ToTarget converts to a full Tabularium AzureResource.
func (a AzureResource) ToTarget() (model.Target, error) {
	if a.Name == "" {
		return nil, fmt.Errorf("azure resource requires name")
	}
	if a.AccountRef == "" {
		return nil, fmt.Errorf("azure resource requires accountRef")
	}

	properties := a.Properties
	if properties == nil {
		properties = make(map[string]any)
	}
	if a.ResourceGroup != "" {
		properties["resourceGroup"] = a.ResourceGroup
	}

	resource, err := model.NewAzureResource(a.Name, a.AccountRef, a.ResourceType, properties)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure resource: %w", err)
	}

	return &resource, nil
}

// ToModel converts to a full Tabularium AzureResource (convenience method).
func (a AzureResource) ToModel() (*model.AzureResource, error) {
	target, err := a.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.AzureResource), nil
}

