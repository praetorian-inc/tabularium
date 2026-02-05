package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// Preseed is a simplified preseed for external tool writers.
// It contains essential fields for creating preseed records.
type Preseed struct {
	Type       string            `json:"type"`                 // Type of preseed data (e.g., "whois", "edgar")
	Title      string            `json:"title"`                // Title or category within type (e.g., "registrant_email")
	Value      string            `json:"value"`                // The actual preseed value (REQUIRED)
	Display    string            `json:"display,omitempty"`    // Display hint (e.g., "text", "image", "base64")
	Metadata   map[string]string `json:"metadata,omitempty"`   // Additional metadata
	Status     string            `json:"status,omitempty"`     // Status code (defaults to "P" for Pending)
	Capability string            `json:"capability,omitempty"` // Associated capability
}

// Group implements Target interface.
func (p Preseed) Group() string { return p.Type }

// Identifier implements Target interface.
func (p Preseed) Identifier() string { return p.Value }

// ToTarget converts to a full Tabularium Preseed.
func (p Preseed) ToTarget() (model.Target, error) {
	if p.Value == "" {
		return nil, fmt.Errorf("preseed requires value")
	}
	if p.Type == "" {
		return nil, fmt.Errorf("preseed requires type")
	}
	if p.Title == "" {
		return nil, fmt.Errorf("preseed requires title")
	}

	preseed := model.NewPreseed(p.Type, p.Title, p.Value)

	// Apply optional fields if provided
	if p.Display != "" {
		preseed.Display = p.Display
	}
	if p.Status != "" {
		preseed.Status = p.Status
	}
	if p.Capability != "" {
		preseed.Capability = p.Capability
	}
	if p.Metadata != nil {
		preseed.Metadata = p.Metadata
	}

	return &preseed, nil
}

// ToModel converts to a full Tabularium Preseed (convenience method).
func (p Preseed) ToModel() (*model.Preseed, error) {
	target, err := p.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.Preseed), nil
}

// PreseedFromModel creates an external Preseed from a model Preseed.
func PreseedFromModel(p *model.Preseed) Preseed {
	return Preseed{
		Type:       p.Type,
		Title:      p.Title,
		Value:      p.Value,
		Display:    p.Display,
		Metadata:   p.Metadata,
		Status:     p.Status,
		Capability: p.Capability,
	}
}
