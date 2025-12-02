package model

import (
	"fmt"
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
)

func init() {
	shared.Registry.MustRegisterModel(&Scanner{})
}

type Scanner struct {
	model.BaseModel
	baseTableModel
	Username string `dynamodbav:"username" json:"username" desc:"Chariot username associated with the account." example:"user@example.com"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the scanner record." example:"#scanner#ip"`
	IP       string `dynamodbav:"ip" json:"ip" desc:"IP address of the scanner record." example:"127.0.0.1"`
	Created  string `dynamodbav:"created" json:"created"`
	Visited  string `dynamodbav:"visited" json:"visited"`
}

func (a *Scanner) GetDescription() string {
	return "Represents a record of access to chariot"
}

func (a *Scanner) Defaulted() {
	a.Created = Now()
	a.Visited = Now()
}

func (a *Scanner) GetHooks() []model.Hook {
	return []model.Hook{
		{
			Call: func() error {
				a.Key = fmt.Sprintf("#scanner#%s", a.IP)
				return nil
			},
		},
	}
}

func NewScanner(ip string) Scanner {
	s := Scanner{
		IP: ip,
	}
	s.Defaulted()
	model.CallHooks(&s)
	return s
}
