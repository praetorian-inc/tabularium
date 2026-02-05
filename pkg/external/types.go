package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// Asset is a simplified asset for external tool writers.
// It contains only the fields needed to identify an asset.
type Asset struct {
	DNS  string `json:"dns"`  // Domain or group identifier
	Name string `json:"name"` // Asset name or IP address
}

// Group implements Target interface.
func (a Asset) Group() string { return a.DNS }

// Identifier implements Target interface.
func (a Asset) Identifier() string { return a.Name }

// ToTarget converts to a full Tabularium Asset.
func (a Asset) ToTarget() (model.Target, error) {
	if a.DNS == "" && a.Name == "" {
		return nil, fmt.Errorf("asset requires dns or name")
	}
	asset := model.NewAsset(a.DNS, a.Name)
	return &asset, nil
}

// ToModel converts to a full Tabularium Asset (convenience method).
func (a Asset) ToModel() (*model.Asset, error) {
	target, err := a.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.Asset), nil
}

// Port is a simplified port for external tool writers.
type Port struct {
	Protocol string `json:"protocol"`        // tcp or udp
	Port     int    `json:"port"`            // Port number
	Service  string `json:"service"`         // Service name (e.g., "https", "ssh")
	Parent   Asset  `json:"parent"`          // Parent asset
}

// Group implements Target interface.
func (p Port) Group() string { return p.Parent.DNS }

// Identifier implements Target interface.
func (p Port) Identifier() string {
	return fmt.Sprintf("%s:%d", p.Parent.Name, p.Port)
}

// ToTarget converts to a full Tabularium Port.
func (p Port) ToTarget() (model.Target, error) {
	if p.Protocol == "" {
		return nil, fmt.Errorf("port requires protocol")
	}
	if p.Port <= 0 || p.Port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535")
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

// Risk is a simplified risk/vulnerability for external tool writers.
type Risk struct {
	Name   string `json:"name"`   // Vulnerability name (e.g., "CVE-2023-1234")
	Status string `json:"status"` // Status code (e.g., "TH", "OH", "OC")
	Target Target `json:"target"` // The target this risk is associated with
}

// ToModel converts to a full Tabularium Risk.
func (r Risk) ToModel() (*model.Risk, error) {
	if r.Name == "" {
		return nil, fmt.Errorf("risk requires name")
	}
	if r.Target == nil {
		return nil, fmt.Errorf("risk requires target")
	}

	target, err := r.Target.ToTarget()
	if err != nil {
		return nil, fmt.Errorf("invalid target: %w", err)
	}

	status := r.Status
	if status == "" {
		status = model.TriageHigh // Default to "TH" (Triage High)
	}

	risk := model.NewRisk(target, r.Name, status)
	return &risk, nil
}
