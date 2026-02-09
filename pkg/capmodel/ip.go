package capmodel

import "encoding/json"

// IP is a convenience type for creating IP address assets. It provides
// ergonomic field names (Address, ParentDomain) instead of the underlying
// model's DNS/Name fields.
//
// For tool writers who prefer the low-level Asset type, NewIPAsset provides
// equivalent functionality.
//
//	col, err := capmodel.Convert(capmodel.IP{Address: "1.2.3.4", ParentDomain: "example.com"})
//	col, err := capmodel.Convert(capmodel.NewIP("1.2.3.4", "example.com"))
type IP struct {
	Address      string `json:"address"`
	ParentDomain string `json:"parent_domain"`
}

// TargetModel returns the registry name for IP conversions.
func (IP) TargetModel() string { return "asset" }

// ConvertJSON maps the ergonomic IP fields to the underlying asset JSON shape.
func (ip IP) ConvertJSON() ([]byte, error) {
	return json.Marshal(struct {
		DNS  string `json:"dns"`
		Name string `json:"name"`
	}{
		DNS:  ip.ParentDomain,
		Name: ip.Address,
	})
}

// NewIP creates an IP for an IP address discovery.
// For standalone IPs (no parent domain), set parentDomain to the IP address itself.
func NewIP(address, parentDomain string) IP {
	return IP{Address: address, ParentDomain: parentDomain}
}
