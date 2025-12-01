package model

import (
	"fmt"
	"strings"

	"slices"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&ParkedDomain{})
	MustRegisterLabel(ParkedDomainLabel)
}

type ParkedDomain struct {
	registry.BaseModel

	Username       string `neo4j:"username" json:"username"`
	Key            string `neo4j:"key" json:"key"`
	Domain         string `neo4j:"domain" json:"domain"`
	ParkedCategory string `neo4j:"parked_category" json:"parked_category"`
	Status         string `neo4j:"status" json:"status"`
	CheckoutStart  string `neo4j:"checkout_start" json:"checkout_start"`
	CheckoutEnd    string `neo4j:"checkout_end" json:"checkout_end"`
	CheckoutUser   string `neo4j:"checkout_user" json:"checkout_user"`
	CheckoutNote   string `neo4j:"checkout_note" json:"checkout_note"`

	// Registration information
	Registrar  string `neo4j:"registrar" json:"registrar"`
	Expires    string `neo4j:"expires" json:"expires"`
	Registered string `neo4j:"registered" json:"registered"`

	// Cloudflare information
	CloudflareStatus string `neo4j:"cloudflare_status" json:"cloudflare_status"`
	AutoRenew        bool   `neo4j:"auto_renew" json:"auto_renew"`
	ZoneId           string `neo4j:"zone_id" json:"zone_id"`

	Updated string `neo4j:"updated" json:"updated"`
}

var ValidParkingCategories = []string{"none", "legal", "business", "finance", "health"}

// GetDescription returns a description for the ParkedDomain model.
func (p *ParkedDomain) GetDescription() string {
	return "Represents a domain name parked in Cloudflare for red team operations."
}

func (p *ParkedDomain) Defaulted() {
	p.Updated = Now()
}

func (p *ParkedDomain) GetKey() string {
	return p.Key
}

const ParkedDomainLabel = "ParkedDomain"

func (p *ParkedDomain) GetLabels() []string {
	return []string{ParkedDomainLabel}
}

func (p *ParkedDomain) Valid() bool {
	return p.Domain != "" && p.Key != ""
}

func (p *ParkedDomain) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if p.Domain == "" {
					return fmt.Errorf("domain is required")
				}
				p.Key = fmt.Sprintf("#parkeddomain#%s", p.Domain)
				return nil
			},
		},
	}
}

func NewParkedDomain(domain string) ParkedDomain {
	p := ParkedDomain{
		Domain: strings.ToLower(domain),
	}
	p.Defaulted()
	registry.CallHooks(&p)
	return p
}

// Merge updates if the field is to be updated
func (p *ParkedDomain) Merge(update ParkedDomain, fieldsToUpdate []string) {
	if slices.Contains(fieldsToUpdate, "parked_category") {

		// TODO:
		// perform the actions needed to update Netlify to park it in the new category, or
		// delete the site if the category is none

		p.ParkedCategory = update.ParkedCategory
	}

	if slices.Contains(fieldsToUpdate, "cloudflare_status") {
		p.CloudflareStatus = update.CloudflareStatus
	}

	if slices.Contains(fieldsToUpdate, "status") {
		p.Status = update.Status

		// Updating to in-use would have these other fields set by the user
		if p.Status == "in-use" {
			p.CheckoutUser = update.CheckoutUser
			p.CheckoutStart = update.CheckoutStart
			p.CheckoutEnd = update.CheckoutEnd
			p.CheckoutNote = update.CheckoutNote
		}
	}

	if slices.Contains(fieldsToUpdate, "auto_renew") {
		p.AutoRenew = update.AutoRenew
	}

	if len(fieldsToUpdate) > 0 {
		p.Updated = Now()
	}
}
