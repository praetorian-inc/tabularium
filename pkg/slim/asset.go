// Package slim provides simplified types for external tool writers.
//
// Usage:
//
//	col, err := slim.Convert(slim.IP{Address: "1.2.3.4", ParentDomain: "example.com"})
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
// DNS is the grouping domain and Name is the specific identifier.
// For domains and CIDRs, set both DNS and Name to the same value.
// For IPs, DNS is the parent domain and Name is the IP address.
// Consider using the IP type for IP address discoveries.
type SlimAsset struct {
	DNS  string `json:"dns"`
	Name string `json:"name"`
}

// IP is a convenience type for IP address discoveries. It is equivalent to
// SlimAsset with more descriptive field names for the IP use case.
// For standalone IPs (no parent domain), set ParentDomain to the IP address itself.
type IP struct {
	Address      string `json:"name"`
	ParentDomain string `json:"dns"`
}

// TargetModel returns the registry name for SlimAsset conversions.
func (SlimAsset) TargetModel() string { return "asset" }

// TargetModel returns the registry name for IP conversions.
func (IP) TargetModel() string { return "asset" }
