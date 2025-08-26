package model

import "github.com/praetorian-inc/tabularium/pkg/registry"

// NoInput represents a capability that requires no input target.
// This is a sentinel type used for capabilities that operate solely on parameters.
type NoInput struct {
	registry.BaseModel
	Status string `json:"status"`
	Key    string `json:"key"`
}

func init() {
	registry.Registry.MustRegisterModel(&NoInput{})
}

func NewNoInput() *NoInput {
	n := &NoInput{}
	registry.CallHooks(n)
	return n
}

func (n *NoInput) GetStatus() string          { return n.Status }
func (n *NoInput) WithStatus(s string) Target { n.Status = s; return n }
func (n *NoInput) Group() string              { return "" }
func (n *NoInput) Identifier() string         { return "" }
func (n *NoInput) IsStatus(string) bool       { return true }
func (n *NoInput) IsClass(string) bool        { return false }
func (n *NoInput) IsPrivate() bool            { return false }
func (n *NoInput) GetLabels() []string        { return []string{} }
func (n *NoInput) Valid() bool                { return false }
func (n *NoInput) GetDescription() string {
	return "Sentinel target for capabilities that require no input target."
}
func (n *NoInput) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				n.Key = "noinput"
				n.Status = Active
				return nil
			},
		},
	}
}
