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
