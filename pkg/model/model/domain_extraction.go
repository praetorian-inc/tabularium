package model

import "github.com/praetorian-inc/tabularium/pkg/registry"

// DomainExtractionMessage triggers domain extraction for integration-discovered assets.
// After an integration (NS1, Okta, Azure AD) discovers assets, this message is sent to
// extract unique primary domains and create seed assets for further discovery.
type DomainExtractionMessage struct {
	Username  string            `json:"username" desc:"Username for AWS context and permissions"`
	JobKey    string            `json:"jobKey" desc:"Job identifier for tracking and logging"`
	Source    string            `json:"source" desc:"Integration source (ns1, okta, azuread-discovery)"`
	TargetDNS string            `json:"targetDNS" desc:"Target domain for asset filtering"`
	Config    map[string]string `json:"config" desc:"Job configuration parameters"`
	CreatedAt string            `json:"createdAt" desc:"Timestamp when message was created"`
}

func init() {
	registry.Registry.MustRegisterModel(&DomainExtractionMessage{})
}

// GetKey returns a unique identifier for the message
func (m *DomainExtractionMessage) GetKey() string {
	return m.JobKey
}

// Identifier returns a human-readable identifier
func (m *DomainExtractionMessage) Identifier() string {
	return m.JobKey + ":" + m.Source
}

// GetDescription returns a description for the DomainExtractionMessage model
func (m *DomainExtractionMessage) GetDescription() string {
	return "Message to trigger automatic domain extraction from integration-discovered assets"
}

// Defaulted is called to set default values on the model
func (m *DomainExtractionMessage) Defaulted() {
	if m.CreatedAt == "" {
		m.CreatedAt = Now()
	}
}

// GetHooks returns a list of hooks for this model
func (m *DomainExtractionMessage) GetHooks() []registry.Hook {
	return []registry.Hook{}
}
