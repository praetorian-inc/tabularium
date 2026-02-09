// Package capmodel provides capability model types for external tool writers.
//
// Usage:
//
//	col, err := capmodel.Convert(capmodel.NewIPAsset("1.2.3.4", "example.com"))
//	assets := collection.Get[*model.Asset](col)
//
//	col, err = capmodel.Convert(capmodel.Port{
//	    Asset:    capmodel.Asset{DNS: "example.com", Name: "1.2.3.4"},
//	    Protocol: "tcp",
//	    Port:     443,
//	    Service:  "https",
//	})
//	ports := collection.Get[*model.Port](col)
package capmodel

// Asset is the low-level capability model for asset representation. It can be
// used standalone with Convert to create an Asset, or embedded as a parent
// reference in types like Port and Attribute.
//
// DNS is the parent/grouping domain; Name is the specific asset identifier.
//
//   - Domain:  Asset{DNS: "example.com", Name: "example.com"}
//   - CIDR:    Asset{DNS: "10.0.0.0/8", Name: "10.0.0.0/8"}
//   - IP:      Asset{DNS: "example.com", Name: "1.2.3.4"}  (or use NewIPAsset)
//
// For top-level assets (domains, CIDRs), DNS and Name are the same because
// the asset is its own group. For IPs, DNS is the parent domain.
//
// For more ergonomic alternatives, see [IP], [Domain], and [CIDR] which
// provide intuitive field names and handle the DNS/Name mapping automatically.
type Asset struct {
	DNS  string `json:"dns"`
	Name string `json:"name"`
}

// TargetModel returns the registry name for Asset conversions.
func (Asset) TargetModel() string { return "asset" }

// NewIPAsset creates an Asset for an IP address discovery.
// For standalone IPs (no parent domain), set parentDomain to the IP address itself.
func NewIPAsset(address, parentDomain string) Asset {
	return Asset{DNS: parentDomain, Name: address}
}
