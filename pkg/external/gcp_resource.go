package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// GCPResource is a simplified GCP resource for external tool writers.
type GCPResource struct {
	Name         string                  `json:"name"`         // GCP resource name/ID
	AccountRef   string                  `json:"accountRef"`   // GCP account reference
	ResourceType model.CloudResourceType `json:"resourceType"` // Type of GCP resource
	Properties   map[string]any          `json:"properties"`   // Resource-specific properties
}

// Group implements Target interface.
func (g GCPResource) Group() string { return g.AccountRef }

// Identifier implements Target interface.
func (g GCPResource) Identifier() string { return g.Name }

// ToTarget converts to a full Tabularium GCPResource.
func (g GCPResource) ToTarget() (model.Target, error) {
	if g.Name == "" {
		return nil, fmt.Errorf("gcp resource requires name")
	}
	if g.AccountRef == "" {
		return nil, fmt.Errorf("gcp resource requires accountRef")
	}

	properties := g.Properties
	if properties == nil {
		properties = make(map[string]any)
	}

	resource, err := model.NewGCPResource(g.Name, g.AccountRef, g.ResourceType, properties)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcp resource: %w", err)
	}

	return &resource, nil
}

// ToModel converts to a full Tabularium GCPResource (convenience method).
func (g GCPResource) ToModel() (*model.GCPResource, error) {
	target, err := g.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.GCPResource), nil
}

// GCPResourceFromModel converts a Tabularium GCPResource to an external GCPResource.
func GCPResourceFromModel(m *model.GCPResource) GCPResource {
	return GCPResource{
		Name:         m.Name,
		AccountRef:   m.AccountRef,
		ResourceType: m.ResourceType,
		Properties:   m.Properties,
	}
}
