package capmodel

import "encoding/json"

// Domain is a convenience type for creating domain assets. The underlying
// model requires both DNS and Name to be set to the same value; this type
// handles that automatically.
//
//	col, err := capmodel.Convert(capmodel.Domain{Name: "example.com"})
type Domain struct {
	Name string `json:"name"`
}

// TargetModel returns the registry name for Domain conversions.
func (Domain) TargetModel() string { return "asset" }

// ConvertJSON maps the Domain to the underlying asset JSON shape,
// setting both dns and name to the domain value.
func (d Domain) ConvertJSON() ([]byte, error) {
	return json.Marshal(struct {
		DNS  string `json:"dns"`
		Name string `json:"name"`
	}{
		DNS:  d.Name,
		Name: d.Name,
	})
}
