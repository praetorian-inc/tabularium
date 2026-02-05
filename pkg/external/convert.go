package external

import (
	"fmt"
	"net"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// --- Entity Types for Semantic Transforms ---
// These provide convenient ways to create Assets from common entity types.

// IP represents an IP address for conversion to an Asset.
type IP struct {
	Address string `json:"address"` // IPv4 or IPv6 address
	Domain  string `json:"domain"`  // Optional associated domain (defaults to Address)
}

// ToAsset converts an IP to an external Asset.
func (ip IP) ToAsset() (Asset, error) {
	parsed := net.ParseIP(ip.Address)
	if parsed == nil {
		return Asset{}, fmt.Errorf("invalid IP address: %s", ip.Address)
	}

	domain := ip.Domain
	if domain == "" {
		domain = ip.Address
	}

	return Asset{DNS: domain, Name: ip.Address}, nil
}

// ToModel converts an IP directly to a Tabularium Asset.
func (ip IP) ToModel() (*model.Asset, error) {
	asset, err := ip.ToAsset()
	if err != nil {
		return nil, err
	}
	return asset.ToModel()
}

// Domain represents a domain for conversion to an Asset.
type Domain struct {
	Name string `json:"name"` // Domain name (e.g., "example.com")
}

// ToAsset converts a Domain to an external Asset.
func (d Domain) ToAsset() (Asset, error) {
	if d.Name == "" {
		return Asset{}, fmt.Errorf("domain name is required")
	}

	// Normalize: remove protocol prefix if present
	name := strings.TrimPrefix(d.Name, "https://")
	name = strings.TrimPrefix(name, "http://")
	name = strings.TrimSuffix(name, "/")

	return Asset{DNS: name, Name: name}, nil
}

// ToModel converts a Domain directly to a Tabularium Asset.
func (d Domain) ToModel() (*model.Asset, error) {
	asset, err := d.ToAsset()
	if err != nil {
		return nil, err
	}
	return asset.ToModel()
}

// CIDR represents a CIDR block for conversion to an Asset.
type CIDR struct {
	Block string `json:"block"` // CIDR notation (e.g., "10.0.0.0/8")
}

// ToAsset converts a CIDR to an external Asset.
func (c CIDR) ToAsset() (Asset, error) {
	_, _, err := net.ParseCIDR(c.Block)
	if err != nil {
		return Asset{}, fmt.Errorf("invalid CIDR block: %s: %w", c.Block, err)
	}

	// For CIDR assets, DNS must contain the CIDR for class detection
	return Asset{DNS: c.Block, Name: c.Block}, nil
}

// ToModel converts a CIDR directly to a Tabularium Asset.
func (c CIDR) ToModel() (*model.Asset, error) {
	asset, err := c.ToAsset()
	if err != nil {
		return nil, err
	}
	return asset.ToModel()
}

// --- Convenience Functions ---

// IPToAsset converts an IP address string to a Tabularium Asset.
func IPToAsset(address string, domain ...string) (*model.Asset, error) {
	ip := IP{Address: address}
	if len(domain) > 0 {
		ip.Domain = domain[0]
	}
	return ip.ToModel()
}

// DomainToAsset converts a domain string to a Tabularium Asset.
func DomainToAsset(name string) (*model.Asset, error) {
	return Domain{Name: name}.ToModel()
}

// CIDRToAsset converts a CIDR block string to a Tabularium Asset.
func CIDRToAsset(block string) (*model.Asset, error) {
	return CIDR{Block: block}.ToModel()
}

// --- Target Helpers ---

// AssetFromModel creates an external Asset from a Tabularium Asset.
// This is useful for projecting full models to simplified external types.
func AssetFromModel(a *model.Asset) Asset {
	return Asset{
		DNS:  a.DNS,
		Name: a.Name,
	}
}
