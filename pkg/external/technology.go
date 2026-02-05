package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// Technology is a simplified technology for external tool writers.
// It represents a specific technology (software, library, framework) identified on an asset.
type Technology struct {
	CPE  string `json:"cpe"`            // The full CPE string (e.g., "cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*")
	Name string `json:"name,omitempty"` // Optional common name for the technology (e.g., "Apache httpd")
}

// ToModel converts to a full Tabularium Technology.
func (t Technology) ToModel() (*model.Technology, error) {
	if t.CPE == "" {
		return nil, fmt.Errorf("technology requires cpe")
	}

	tech, err := model.NewTechnology(t.CPE)
	if err != nil {
		return nil, fmt.Errorf("invalid cpe: %w", err)
	}

	if t.Name != "" {
		tech.Name = t.Name
	}

	return &tech, nil
}

