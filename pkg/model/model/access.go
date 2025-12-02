package model

import (
	"fmt"
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
)

func init() {
	shared.Registry.MustRegisterModel(&Access{})
}

type ChariotAccessType string

const (
	APIKey           ChariotAccessType = "api_key"
	SignInWithGoogle ChariotAccessType = "sign_in_with_google"
	SSO              ChariotAccessType = "sso"
	UsernamePassword ChariotAccessType = "username_password"
)

type Access struct {
	model.BaseModel
	baseTableModel
	Username string `dynamodbav:"username" json:"username" desc:"Chariot username associated with the account." example:"user@example.com"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the access record." example:"#access#type#value#id"`
	// Attributes
	Sub     string            `dynamodbav:"sub" json:"sub" desc:"The Cognito sub of this user"`
	Type    ChariotAccessType `dynamodbav:"type" json:"type"`
	Value   string            `dynamodbav:"value,omitempty" json:"value,omitempty"`
	Name    string            `dynamodbav:"name,omitempty" json:"name,omitempty"`
	Email   string            `dynamodbav:"email,omitempty" json:"email,omitempty"`
	Updated string            `dynamodbav:"updated" json:"updated"`
}

func (a *Access) GetDescription() string {
	return "Represents a record of access to chariot"
}

func (a *Access) Defaulted() {
	a.Updated = Now()
}

func (a *Access) GetHooks() []model.Hook {
	return []model.Hook{
		{
			Call: func() error {
				a.Key = fmt.Sprintf("#access#%s", a.Sub)
				a.Updated = Now()
				return nil
			},
		},
	}
}

func NewAccess(sub string, tipe ChariotAccessType, value string) Access {
	a := Access{
		Sub:   sub,
		Type:  tipe,
		Value: value,
	}
	a.Defaulted()
	model.CallHooks(&a)
	return a
}
