package model

import (
	"encoding/json"
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&Account{})
}

type Account struct {
	registry.BaseModel
	baseTableModel
	Username string `dynamodbav:"username" json:"username" desc:"Chariot username associated with the account." example:"user@example.com"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the account record." example:"#account#name#member#value"`
	// Attributes
	Name     string            `dynamodbav:"name" json:"name" desc:"The owner of the account." example:"chariot.customer@example.com"`
	Member   string            `dynamodbav:"member" json:"member" desc:"The user or system granted access to this account." example:"amazon"`
	Value    string            `dynamodbav:"value,omitempty" json:"value" desc:"The identifier for this account within the context of member." example:"01234567890"`
	// Hierarchy fields for parent/child tenant relationships
	ParentTenant     string   `dynamodbav:"parent_tenant,omitempty" json:"parent_tenant,omitempty" desc:"Email of the parent tenant if this is a child tenant." example:"parent@example.com"`
	TenantType       string   `dynamodbav:"tenant_type,omitempty" json:"tenant_type,omitempty" desc:"Tenant type: 'parent', 'child', or 'standalone'." example:"child"`
	InheritedDomains []string `dynamodbav:"inherited_domains,omitempty" json:"inherited_domains,omitempty" desc:"Domains inherited from parent tenant." example:"[\"example.com\", \"acme.com\"]"`
	AllowedDomains   []string `dynamodbav:"allowed_domains,omitempty" json:"allowed_domains,omitempty" desc:"Domains allowed for this tenant (in addition to inherited)." example:"[\"subsidiary.com\"]"`
	SharedFromParent bool     `dynamodbav:"shared_from_parent,omitempty" json:"shared_from_parent,omitempty" desc:"True if this user account was shared from a parent tenant." example:"true"`
	Updated  string            `dynamodbav:"updated" json:"updated" desc:"Timestamp when the account info was last updated (RFC3339)." example:"2023-10-27T10:00:00Z"`
	TTL      int64             `dynamodbav:"ttl" json:"ttl" desc:"Time-to-live for the account record (Unix timestamp)." example:"1706353200"`
	Secret   map[string]string `dynamodbav:"-" json:"secret,omitempty" desc:"Secret configuration map associated with the account." example:"{\"service_principal_token\": \"103713408v0871v\"}"`
	Settings json.RawMessage   `dynamodbav:"settings,omitempty" json:"settings,omitempty" desc:"Raw JSON message containing specific settings." example:"{\"notifications\": true}"`
}

// GetDescription returns a description for the Account model.
func (a *Account) GetDescription() string {
	return "Represents a Chariot account (linked account), linking a user/system (member) to a specific cloud account or service instance (value). Supports parent/child tenant hierarchies with domain inheritance and user sharing."
}

func (a *Account) Defaulted() {
	a.Updated = Now()
}

func (a *Account) GetKey() string {
	return a.Key
}

func (a *Account) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				a.Key = fmt.Sprintf("#account#%s#%s#%s", a.Name, a.Member, a.Value)
				return nil
			},
		},
	}
}

func NewAccount(name, member, value string, secret map[string]string) Account {
	a := Account{
		Name:   name,
		Member: member,
		Value:  value,
		Secret: secret,
	}
	a.Defaulted()
	registry.CallHooks(&a)
	return a
}

func NewAccountWithSettings(name, member, value string, secret map[string]string, settings json.RawMessage) Account {
	account := NewAccount(name, member, value, secret)
	account.Settings = settings
	return account
}
