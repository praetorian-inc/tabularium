// Package slim provides simplified types for external tool writers.
//
// Usage:
//
//	col, err := slim.Convert(slim.NewIPAsset("1.2.3.4", "example.com"))
//	assets := collection.Get[*model.Asset](col)
//
//	col, err = slim.Convert(slim.SlimPort{
//	    Asset:    slim.SlimAsset{DNS: "example.com", Name: "1.2.3.4"},
//	    Protocol: "tcp",
//	    Port:     443,
//	    Service:  "https",
//	})
//	ports := collection.Get[*model.Port](col)
package slim

// SlimAsset is a simplified asset representation. It can be used standalone
// with Convert to create an Asset, or embedded as a parent reference in
// types like SlimPort and SlimAttribute.
//
// DNS is the parent/grouping domain; Name is the specific asset identifier.
//
//   - Domain:  SlimAsset{DNS: "example.com", Name: "example.com"}
//   - CIDR:    SlimAsset{DNS: "10.0.0.0/8", Name: "10.0.0.0/8"}
//   - IP:      SlimAsset{DNS: "example.com", Name: "1.2.3.4"}  (or use NewIPAsset)
//
// For top-level assets (domains, CIDRs), DNS and Name are the same because
// the asset is its own group. For IPs, DNS is the parent domain.
type SlimAsset struct {
	DNS  string `json:"dns"`
	Name string `json:"name"`
}

// TargetModel returns the registry name for SlimAsset conversions.
func (SlimAsset) TargetModel() string { return "asset" }

// NewIPAsset creates a SlimAsset for an IP address discovery.
// For standalone IPs (no parent domain), set parentDomain to the IP address itself.
func NewIPAsset(address, parentDomain string) SlimAsset {
	return SlimAsset{DNS: parentDomain, Name: address}
}
