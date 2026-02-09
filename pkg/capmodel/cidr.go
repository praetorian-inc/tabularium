package capmodel

import "encoding/json"

// CIDR is a convenience type for creating CIDR range assets. The underlying
// model requires both DNS and Name to be set to the same value; this type
// handles that automatically.
//
//	col, err := capmodel.Convert(capmodel.CIDR{Range: "10.0.0.0/8"})
type CIDR struct {
	Range string `json:"range"`
}

// TargetModel returns the registry name for CIDR conversions.
func (CIDR) TargetModel() string { return "asset" }

// ConvertJSON maps the CIDR to the underlying asset JSON shape,
// setting both dns and name to the CIDR range value.
func (c CIDR) ConvertJSON() ([]byte, error) {
	return json.Marshal(struct {
		DNS  string `json:"dns"`
		Name string `json:"name"`
	}{
		DNS:  c.Range,
		Name: c.Range,
	})
}
