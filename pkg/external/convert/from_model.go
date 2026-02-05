package convert

import (
	"github.com/praetorian-inc/tabularium/pkg/external"
	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// PortFromModel converts a Tabularium Port to an external Port.
func PortFromModel(m *model.Port) external.Port {
	asset := m.Asset()
	return external.Port{
		Protocol: m.Protocol,
		Port:     m.Port,
		Service:  m.Service,
		Parent:   external.Asset{DNS: asset.DNS, Name: asset.Name},
	}
}

// AWSResourceFromModel converts a Tabularium AWSResource to an external AWSResource.
func AWSResourceFromModel(m *model.AWSResource) external.AWSResource {
	return external.AWSResource{
		ARN:               m.Name,
		AccountRef:        m.AccountRef,
		ResourceType:      m.ResourceType,
		Properties:        m.Properties,
		OrgPolicyFilename: m.OrgPolicyFilename,
	}
}

// AzureResourceFromModel converts a Tabularium AzureResource to an external AzureResource.
func AzureResourceFromModel(m *model.AzureResource) external.AzureResource {
	return external.AzureResource{
		Name:          m.Name,
		AccountRef:    m.AccountRef,
		ResourceType:  m.ResourceType,
		Properties:    m.Properties,
		ResourceGroup: m.ResourceGroup,
	}
}

// GCPResourceFromModel converts a Tabularium GCPResource to an external GCPResource.
func GCPResourceFromModel(m *model.GCPResource) external.GCPResource {
	return external.GCPResource{
		Name:         m.Name,
		AccountRef:   m.AccountRef,
		ResourceType: m.ResourceType,
		Properties:   m.Properties,
	}
}

// OrganizationFromModel converts a full Tabularium Organization to the simplified external type.
func OrganizationFromModel(org *model.Organization) external.Organization {
	ext := external.Organization{
		Domain: derefString(org.Domain),
		Name:   derefString(org.Name),
	}

	// Core fields
	if org.Website != nil {
		ext.Website = *org.Website
	}
	if org.Description != nil {
		ext.Description = *org.Description
	}
	if org.Industry != nil {
		ext.Industry = *org.Industry
	}

	// Size and financial
	if org.EstimatedNumEmployees != nil {
		ext.EstimatedNumEmployees = *org.EstimatedNumEmployees
	}
	if org.EmployeeRange != nil {
		ext.EmployeeRange = *org.EmployeeRange
	}
	if org.AnnualRevenue != nil {
		ext.AnnualRevenue = *org.AnnualRevenue
	}
	if org.RevenueRange != nil {
		ext.RevenueRange = *org.RevenueRange
	}

	// Geographic
	if org.Country != nil {
		ext.Country = *org.Country
	}
	if org.State != nil {
		ext.State = *org.State
	}
	if org.City != nil {
		ext.City = *org.City
	}
	if org.StreetAddress != nil {
		ext.StreetAddress = *org.StreetAddress
	}

	// Contact
	if org.Phone != nil {
		ext.Phone = *org.Phone
	}
	if org.Email != nil {
		ext.Email = *org.Email
	}

	// Social
	if org.LinkedinURL != nil {
		ext.LinkedinURL = *org.LinkedinURL
	}
	if org.TwitterURL != nil {
		ext.TwitterURL = *org.TwitterURL
	}

	// Company details
	if org.FoundedYear != nil {
		ext.FoundedYear = *org.FoundedYear
	}
	if org.PubliclyTraded != nil {
		ext.PubliclyTraded = *org.PubliclyTraded
	}
	if org.TickerSymbol != nil {
		ext.TickerSymbol = *org.TickerSymbol
	}

	// Enrichment
	if org.ApollioID != nil {
		ext.ApollioID = *org.ApollioID
	}
	if org.EnrichmentSource != nil {
		ext.EnrichmentSource = *org.EnrichmentSource
	}
	if org.DataQualityScore != nil {
		ext.DataQualityScore = *org.DataQualityScore
	}

	// Technology
	if org.Technologies != nil {
		ext.Technologies = *org.Technologies
	}

	// Classification
	if org.SubIndustries != nil {
		ext.SubIndustries = *org.SubIndustries
	}
	if org.Keywords != nil {
		ext.Keywords = *org.Keywords
	}
	if org.OrganizationType != nil {
		ext.OrganizationType = *org.OrganizationType
	}

	return ext
}

// PreseedFromModel creates an external Preseed from a model Preseed.
func PreseedFromModel(p *model.Preseed) external.Preseed {
	return external.Preseed{
		Type:       p.Type,
		Title:      p.Title,
		Value:      p.Value,
		Display:    p.Display,
		Metadata:   p.Metadata,
		Status:     p.Status,
		Capability: p.Capability,
	}
}

// ADObjectFromModel converts a Tabularium ADObject to an external ADObject.
func ADObjectFromModel(m *model.ADObject) external.ADObject {
	return external.ADObject{
		Label:             m.Label,
		Domain:            m.Domain,
		ObjectID:          m.ObjectID,
		DistinguishedName: m.DistinguishedName,
	}
}

// TechnologyFromModel converts a Tabularium Technology to an external Technology.
func TechnologyFromModel(m *model.Technology) external.Technology {
	return external.Technology{
		CPE:  m.CPE,
		Name: m.Name,
	}
}

// PersonFromModel converts a Tabularium Person to an external Person.
func PersonFromModel(m *model.Person) external.Person {
	ext := external.Person{}

	if m.Email != nil {
		ext.Email = *m.Email
	}
	if m.Name != nil {
		ext.Name = *m.Name
	}
	if m.Title != nil {
		ext.Title = *m.Title
	}
	if m.OrganizationName != nil {
		ext.OrganizationName = *m.OrganizationName
	}
	if m.LinkedinURL != nil {
		ext.LinkedinURL = *m.LinkedinURL
	}

	return ext
}

// WebpageFromModel converts a Tabularium Webpage to an external Webpage.
func WebpageFromModel(m *model.Webpage) external.Webpage {
	return external.Webpage{
		URL: m.URL,
	}
}

// derefString safely dereferences a string pointer.
// Returns empty string if the pointer is nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
