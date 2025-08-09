package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&ADDomain{})
}

var (
	ADDomainLabel = NewLabel("Addomain")
)

var (
	adDomainKey = regexp.MustCompile(`^#addomain#([^#]+)#([^#]+)$`)
)

type ADDomain struct {
	BaseAsset
	Name string `neo4j:"name" json:"name" desc:"NetBIOS of the domain." example:"example.internal"`
}

func (d *ADDomain) IsPrivate() bool {
	return true
}

func (d *ADDomain) GetKey() string {
	return d.Key
}

func (d *ADDomain) Valid() bool {
	return domain.MatchString(d.Name) && adDomainKey.MatchString(d.Key)
}

func (d *ADDomain) GetLabels() []string {
	return []string{ADDomainLabel, AssetLabel, TTLLabel}
}

func (d *ADDomain) GetClass() string {
	return "addomain"
}

func (d *ADDomain) GetStatus() string {
	return d.Status
}

func (d *ADDomain) Group() string {
	return d.Name
}

func (d *ADDomain) Identifier() string {
	return d.Name
}

func (d *ADDomain) IsStatus(value string) bool {
	return strings.HasPrefix(d.Status, value)
}

func (d *ADDomain) WithStatus(status string) Target {
	ret := *d
	ret.Status = status
	return &ret
}

func (d *ADDomain) Attribute(name, value string) Attribute {
	attr := NewAttribute(name, value, d)
	return attr
}

func (d *ADDomain) Seed() Seed {
	s := NewSeed(d.Name)
	s.SetStatus(d.Status)
	return s
}

func (d *ADDomain) Defaulted() {
	d.BaseAsset.Defaulted()
	d.Class = "addomain"
}

func (d *ADDomain) GetHooks() []registry.Hook {
	return []registry.Hook{
		useGroupAndIdentifier(d, &d.Name, &d.Name),
		{
			Call: func() error {
				d.Key = fmt.Sprintf("#addomain#%s#%s", d.Name, d.Name)
				d.BaseAsset.Identifier = d.Name
				d.BaseAsset.Group = ""
				return nil
			},
		},
		setGroupAndIdentifier(d, &d.Name, &d.Name),
	}
}

func NewADDomain(name string) ADDomain {
	d := ADDomain{
		Name: name,
	}
	d.Defaulted()
	registry.CallHooks(&d)
	return d
}

func (d *ADDomain) GetDescription() string {
	return "Represents an Active Directory domain, including its name, status, and creation/modification timestamps."
}
