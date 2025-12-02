package model

import (
	"encoding/json"
	"fmt"
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
)

func init() {
	shared.Registry.MustRegisterModel(&Account{})
}

type Account struct {
	model.BaseModel
	baseTableModel
	Username string `dynamodbav:"username" json:"username" desc:"Chariot username associated with the account." example:"user@example.com"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the account record." example:"#account#name#member#value"`
	// Attributes
	Name     string            `dynamodbav:"name" json:"name" desc:"The owner of the account." example:"chariot.customer@example.com"`
	Member   string            `dynamodbav:"member" json:"member" desc:"The user or system granted access to this account." example:"amazon"`
	Value    string            `dynamodbav:"value,omitempty" json:"value" desc:"The identifier for this account within the context of member." example:"01234567890"`
	Updated  string            `dynamodbav:"updated" json:"updated" desc:"Timestamp when the account info was last updated (RFC3339)." example:"2023-10-27T10:00:00Z"`
	TTL      int64             `dynamodbav:"ttl" json:"ttl" desc:"Time-to-live for the account record (Unix timestamp)." example:"1706353200"`
	Secret   map[string]string `dynamodbav:"-" json:"secret,omitempty" desc:"Secret configuration map associated with the account." example:"{\"service_principal_token\": \"103713408v0871v\"}"`
	Settings json.RawMessage   `dynamodbav:"settings,omitempty" json:"settings,omitempty" desc:"Raw JSON message containing specific settings." example:"{\"notifications\": true}"`
}

// GetDescription returns a description for the Account model.
func (a *Account) GetDescription() string {
	return "Represents a Chariot account, linking a user/system (member) to a specific cloud account or service instance (value)."
}

func (a *Account) Defaulted() {
	a.Updated = Now()
}

func (a *Account) GetKey() string {
	return a.Key
}

func (a *Account) GetHooks() []model.Hook {
	return []model.Hook{
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
	model.CallHooks(&a)
	return a
}

func NewAccountWithSettings(name, member, value string, secret map[string]string, settings json.RawMessage) Account {
	account := NewAccount(name, member, value, secret)
	account.Settings = settings
	return account
}
