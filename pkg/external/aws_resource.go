package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// AWSResource is a simplified AWS resource for external tool writers.
type AWSResource struct {
	ARN               string                  `json:"arn"`               // AWS ARN (Amazon Resource Name)
	AccountRef        string                  `json:"accountRef"`        // AWS account reference
	ResourceType      model.CloudResourceType `json:"resourceType"`      // Type of AWS resource
	Properties        map[string]any          `json:"properties"`        // Resource-specific properties
	OrgPolicyFilename string                  `json:"orgPolicyFilename"` // Organization policy filename (optional)
}

// Group implements Target interface.
func (a AWSResource) Group() string { return a.AccountRef }

// Identifier implements Target interface.
func (a AWSResource) Identifier() string { return a.ARN }

// ToTarget converts to a full Tabularium AWSResource.
func (a AWSResource) ToTarget() (model.Target, error) {
	if a.ARN == "" {
		return nil, fmt.Errorf("aws resource requires arn")
	}
	if a.AccountRef == "" {
		return nil, fmt.Errorf("aws resource requires accountRef")
	}

	resource, err := model.NewAWSResource(a.ARN, a.AccountRef, a.ResourceType, a.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to create aws resource: %w", err)
	}

	if a.OrgPolicyFilename != "" {
		resource.OrgPolicyFilename = a.OrgPolicyFilename
	}

	return &resource, nil
}

// ToModel converts to a full Tabularium AWSResource (convenience method).
func (a AWSResource) ToModel() (*model.AWSResource, error) {
	target, err := a.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.AWSResource), nil
}

// AWSResourceFromModel converts a Tabularium AWSResource to an external AWSResource.
func AWSResourceFromModel(m *model.AWSResource) AWSResource {
	return AWSResource{
		ARN:               m.Name,
		AccountRef:        m.AccountRef,
		ResourceType:      m.ResourceType,
		Properties:        m.Properties,
		OrgPolicyFilename: m.OrgPolicyFilename,
	}
}
