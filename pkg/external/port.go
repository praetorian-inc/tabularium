package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// Port is a simplified port for external tool writers.
// It contains essential fields needed to identify an open port on an asset.
type Port struct {
	Protocol string `json:"protocol"` // Protocol: "tcp" or "udp"
	Port     int    `json:"port"`     // Port number (1-65535)
	Service  string `json:"service"`  // Service name (e.g., "https", "ssh")
	Parent   Asset  `json:"parent"`   // Parent asset this port belongs to
}

// Group implements Target interface.
func (p Port) Group() string { return p.Parent.DNS }

// Identifier implements Target interface.
func (p Port) Identifier() string {
	return fmt.Sprintf("%s:%d", p.Parent.Name, p.Port)
}

// ToTarget converts to a full Tabularium Port.
func (p Port) ToTarget() (model.Target, error) {
	if p.Port <= 0 || p.Port > 65535 {
		return nil, fmt.Errorf("port number must be between 1 and 65535")
	}
	if p.Protocol == "" {
		return nil, fmt.Errorf("port requires protocol (tcp or udp)")
	}
	if p.Protocol != "tcp" && p.Protocol != "udp" {
		return nil, fmt.Errorf("port protocol must be tcp or udp")
	}

	parentAsset, err := p.Parent.ToModel()
	if err != nil {
		return nil, fmt.Errorf("invalid parent asset: %w", err)
	}

	port := model.NewPort(p.Protocol, p.Port, parentAsset)
	if p.Service != "" {
		port.Service = p.Service
	}

	return &port, nil
}

// ToModel converts to a full Tabularium Port (convenience method).
func (p Port) ToModel() (*model.Port, error) {
	target, err := p.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.Port), nil
}

// PortFromModel converts a Tabularium Port to an external Port.
func PortFromModel(m *model.Port) Port {
	asset := m.Asset()
	return Port{
		Protocol: m.Protocol,
		Port:     m.Port,
		Service:  m.Service,
		Parent:   Asset{DNS: asset.DNS, Name: asset.Name},
	}
}
