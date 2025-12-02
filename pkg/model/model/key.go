package model

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
)

func init() {
	shared.Registry.MustRegisterModel(&Key{})
}

type Key struct {
	model.BaseModel
	baseTableModel
	Username string `dynamodbav:"username" json:"username" desc:"Chariot username associated with the account." example:"user@example.com"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the key record." example:"#key#550e8400-e29b-41d4-a716-446655440000"`
	ID       string `dynamodbav:"id" json:"id" desc:"ID of this key" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name     string `dynamodbav:"name" json:"name" desc:"Name of this key" example:"user"`
	Created  string `dynamodbav:"created" json:"created"`
	Creator  string `dynamodbav:"creator" json:"creator"`
	Visited  string `dynamodbav:"visited" json:"visited"`
	Deleted  string `dynamodbav:"deleted" json:"deleted"`
	Deleter  string `dynamodbav:"deleter" json:"deleter"`
	Status   string `dynamodbav:"status" json:"status"`
	Secret   string `dynamodbav:"-" json:"secret"`
	Expires  string `dynamodbav:"expires" json:"expires" desc:"Optional expiration timestamp in RFC3339 format" example:"2024-12-31T23:59:59Z"`
}

func (k *Key) GetDescription() string {
	return "An API key for the Chariot platform"
}

func (k *Key) GetHooks() []model.Hook {
	return []model.Hook{
		{
			Call: func() error {
				k.Key = fmt.Sprintf("#key#%s", k.ID)
				return nil
			},
		},
	}
}

func (k *Key) Defaulted() {
	k.Created = Now()
	k.Status = Active
	bites := make([]byte, 16)
	_, _ = rand.Read(bites)
	k.ID = base32.StdEncoding.EncodeToString(bites)
}

func NewKey(name string) Key {
	k := Key{
		Name: name,
	}
	k.Defaulted()
	model.CallHooks(&k)
	return k
}
