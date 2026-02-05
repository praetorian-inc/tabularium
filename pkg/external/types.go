package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// Asset is a simplified asset for external tool writers.
// It contains only the fields needed to identify an asset.
type Asset struct {
	DNS  string `json:"dns"`  // Domain or group identifier
	Name string `json:"name"` // Asset name or IP address
}

// Group implements Target interface.
func (a Asset) Group() string { return a.DNS }

// Identifier implements Target interface.
func (a Asset) Identifier() string { return a.Name }

// ToTarget converts to a full Tabularium Asset.
func (a Asset) ToTarget() (model.Target, error) {
	if a.DNS == "" && a.Name == "" {
		return nil, fmt.Errorf("asset requires dns or name")
	}
	asset := model.NewAsset(a.DNS, a.Name)
	return &asset, nil
}

// ToModel converts to a full Tabularium Asset (convenience method).
func (a Asset) ToModel() (*model.Asset, error) {
	target, err := a.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.Asset), nil
}

// Port is a simplified port for external tool writers.
// It contains essential fields needed to identify an open port on an asset.
type Port struct {
	Protocol string `json:"protocol"` // Protocol: "tcp" or "udp"
	Port     int    `json:"port"`     // Port number (1-65535)
	Service  string `json:"service"`  // Service name (e.g., "https", "ssh")
	Parent   Asset  `json:"parent"`   // Parent asset this port belongs to
}

// Group implements Target interface.
func (p Port) Group() string { return p.Parent.DNS }

// Identifier implements Target interface.
func (p Port) Identifier() string {
	return fmt.Sprintf("%s:%d", p.Parent.Name, p.Port)
}

// ToTarget converts to a full Tabularium Port.
func (p Port) ToTarget() (model.Target, error) {
	if p.Port <= 0 || p.Port > 65535 {
		return nil, fmt.Errorf("port number must be between 1 and 65535")
	}
	if p.Protocol == "" {
		return nil, fmt.Errorf("port requires protocol (tcp or udp)")
	}
	if p.Protocol != "tcp" && p.Protocol != "udp" {
		return nil, fmt.Errorf("port protocol must be tcp or udp")
	}

	parentAsset, err := p.Parent.ToModel()
	if err != nil {
		return nil, fmt.Errorf("invalid parent asset: %w", err)
	}

	port := model.NewPort(p.Protocol, p.Port, parentAsset)
	if p.Service != "" {
		port.Service = p.Service
	}

	return &port, nil
}

// ToModel converts to a full Tabularium Port (convenience method).
func (p Port) ToModel() (*model.Port, error) {
	target, err := p.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.Port), nil
}

// PortFromModel converts a Tabularium Port to an external Port.
func PortFromModel(m *model.Port) Port {
	asset := m.Asset()
	return Port{
		Protocol: m.Protocol,
		Port:     m.Port,
		Service:  m.Service,
		Parent:   Asset{DNS: asset.DNS, Name: asset.Name},
	}
}

// AWSResource is a simplified AWS resource for external tool writers.
type AWSResource struct {
	ARN               string                  `json:"arn"`               // AWS ARN (Amazon Resource Name)
	AccountRef        string                  `json:"accountRef"`        // AWS account reference
	ResourceType      model.CloudResourceType `json:"resourceType"`      // Type of AWS resource
	Properties        map[string]any          `json:"properties"`        // Resource-specific properties
	OrgPolicyFilename string                  `json:"orgPolicyFilename"` // Organization policy filename (optional)
}

// Group implements Target interface.
func (a AWSResource) Group() string { return a.AccountRef }

// Identifier implements Target interface.
func (a AWSResource) Identifier() string { return a.ARN }

// ToTarget converts to a full Tabularium AWSResource.
func (a AWSResource) ToTarget() (model.Target, error) {
	if a.ARN == "" {
		return nil, fmt.Errorf("aws resource requires arn")
	}
	if a.AccountRef == "" {
		return nil, fmt.Errorf("aws resource requires accountRef")
	}

	resource, err := model.NewAWSResource(a.ARN, a.AccountRef, a.ResourceType, a.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to create aws resource: %w", err)
	}

	if a.OrgPolicyFilename != "" {
		resource.OrgPolicyFilename = a.OrgPolicyFilename
	}

	return &resource, nil
}

// ToModel converts to a full Tabularium AWSResource (convenience method).
func (a AWSResource) ToModel() (*model.AWSResource, error) {
	target, err := a.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.AWSResource), nil
}

// AWSResourceFromModel converts a Tabularium AWSResource to an external AWSResource.
func AWSResourceFromModel(m *model.AWSResource) AWSResource {
	return AWSResource{
		ARN:               m.Name,
		AccountRef:        m.AccountRef,
		ResourceType:      m.ResourceType,
		Properties:        m.Properties,
		OrgPolicyFilename: m.OrgPolicyFilename,
	}
}

// AzureResource is a simplified Azure resource for external tool writers.
type AzureResource struct {
	Name          string                  `json:"name"`          // Azure resource name/ID
	AccountRef    string                  `json:"accountRef"`    // Azure account reference
	ResourceType  model.CloudResourceType `json:"resourceType"`  // Type of Azure resource
	Properties    map[string]any          `json:"properties"`    // Resource-specific properties
	ResourceGroup string                  `json:"resourceGroup"` // Azure resource group
}

// Group implements Target interface.
func (a AzureResource) Group() string { return a.AccountRef }

// Identifier implements Target interface.
func (a AzureResource) Identifier() string { return a.Name }

// ToTarget converts to a full Tabularium AzureResource.
func (a AzureResource) ToTarget() (model.Target, error) {
	if a.Name == "" {
		return nil, fmt.Errorf("azure resource requires name")
	}
	if a.AccountRef == "" {
		return nil, fmt.Errorf("azure resource requires accountRef")
	}

	properties := a.Properties
	if properties == nil {
		properties = make(map[string]any)
	}
	if a.ResourceGroup != "" {
		properties["resourceGroup"] = a.ResourceGroup
	}

	resource, err := model.NewAzureResource(a.Name, a.AccountRef, a.ResourceType, properties)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure resource: %w", err)
	}

	return &resource, nil
}

// ToModel converts to a full Tabularium AzureResource (convenience method).
func (a AzureResource) ToModel() (*model.AzureResource, error) {
	target, err := a.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.AzureResource), nil
}

// AzureResourceFromModel converts a Tabularium AzureResource to an external AzureResource.
func AzureResourceFromModel(m *model.AzureResource) AzureResource {
	return AzureResource{
		Name:          m.Name,
		AccountRef:    m.AccountRef,
		ResourceType:  m.ResourceType,
		Properties:    m.Properties,
		ResourceGroup: m.ResourceGroup,
	}
}

// GCPResource is a simplified GCP resource for external tool writers.
type GCPResource struct {
	Name         string                  `json:"name"`         // GCP resource name/ID
	AccountRef   string                  `json:"accountRef"`   // GCP account reference
	ResourceType model.CloudResourceType `json:"resourceType"` // Type of GCP resource
	Properties   map[string]any          `json:"properties"`   // Resource-specific properties
}

// Group implements Target interface.
func (g GCPResource) Group() string { return g.AccountRef }

// Identifier implements Target interface.
func (g GCPResource) Identifier() string { return g.Name }

// ToTarget converts to a full Tabularium GCPResource.
func (g GCPResource) ToTarget() (model.Target, error) {
	if g.Name == "" {
		return nil, fmt.Errorf("gcp resource requires name")
	}
	if g.AccountRef == "" {
		return nil, fmt.Errorf("gcp resource requires accountRef")
	}

	properties := g.Properties
	if properties == nil {
		properties = make(map[string]any)
	}

	resource, err := model.NewGCPResource(g.Name, g.AccountRef, g.ResourceType, properties)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcp resource: %w", err)
	}

	return &resource, nil
}

// ToModel converts to a full Tabularium GCPResource (convenience method).
func (g GCPResource) ToModel() (*model.GCPResource, error) {
	target, err := g.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.GCPResource), nil
}

// GCPResourceFromModel converts a Tabularium GCPResource to an external GCPResource.
func GCPResourceFromModel(m *model.GCPResource) GCPResource {
	return GCPResource{
		Name:         m.Name,
		AccountRef:   m.AccountRef,
		ResourceType: m.ResourceType,
		Properties:   m.Properties,
	}
}

// Risk is a simplified risk/vulnerability for external tool writers.
type Risk struct {
	Name   string `json:"name"`   // Vulnerability name (e.g., "CVE-2023-1234")
	Status string `json:"status"` // Status code (e.g., "TH", "OH", "OC")
	Target Target `json:"target"` // The target this risk is associated with
}

// ToModel converts to a full Tabularium Risk.
func (r Risk) ToModel() (*model.Risk, error) {
	if r.Name == "" {
		return nil, fmt.Errorf("risk requires name")
	}
	if r.Target == nil {
		return nil, fmt.Errorf("risk requires target")
	}

	target, err := r.Target.ToTarget()
	if err != nil {
		return nil, fmt.Errorf("invalid target: %w", err)
	}

	status := r.Status
	if status == "" {
		status = model.TriageHigh // Default to "TH" (Triage High)
	}

	risk := model.NewRisk(target, r.Name, status)
	return &risk, nil
}

// Preseed is a simplified preseed for external tool writers.
// It contains essential fields for creating preseed records.
type Preseed struct {
	Type       string            `json:"type"`                 // Type of preseed data (e.g., "whois", "edgar")
	Title      string            `json:"title"`                // Title or category within type (e.g., "registrant_email")
	Value      string            `json:"value"`                // The actual preseed value (REQUIRED)
	Display    string            `json:"display,omitempty"`    // Display hint (e.g., "text", "image", "base64")
	Metadata   map[string]string `json:"metadata,omitempty"`   // Additional metadata
	Status     string            `json:"status,omitempty"`     // Status code (defaults to "P" for Pending)
	Capability string            `json:"capability,omitempty"` // Associated capability
}

// Group implements Target interface.
func (p Preseed) Group() string { return p.Type }

// Identifier implements Target interface.
func (p Preseed) Identifier() string { return p.Value }

// ToTarget converts to a full Tabularium Preseed.
func (p Preseed) ToTarget() (model.Target, error) {
	if p.Value == "" {
		return nil, fmt.Errorf("preseed requires value")
	}
	if p.Type == "" {
		return nil, fmt.Errorf("preseed requires type")
	}
	if p.Title == "" {
		return nil, fmt.Errorf("preseed requires title")
	}

	preseed := model.NewPreseed(p.Type, p.Title, p.Value)

	// Apply optional fields if provided
	if p.Display != "" {
		preseed.Display = p.Display
	}
	if p.Status != "" {
		preseed.Status = p.Status
	}
	if p.Capability != "" {
		preseed.Capability = p.Capability
	}
	if p.Metadata != nil {
		preseed.Metadata = p.Metadata
	}

	return &preseed, nil
}

// ToModel converts to a full Tabularium Preseed (convenience method).
func (p Preseed) ToModel() (*model.Preseed, error) {
	target, err := p.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.Preseed), nil
}

// PreseedFromModel creates an external Preseed from a model Preseed.
func PreseedFromModel(p *model.Preseed) Preseed {
	return Preseed{
		Type:       p.Type,
		Title:      p.Title,
		Value:      p.Value,
		Display:    p.Display,
		Metadata:   p.Metadata,
		Status:     p.Status,
		Capability: p.Capability,
	}
}

// Organization is a simplified organization for external tool writers.
// It contains fields useful for company intelligence and enrichment use cases.
type Organization struct {
	// Required fields
	Domain string `json:"domain"` // Primary domain (REQUIRED)
	Name   string `json:"name"`   // Organization name

	// Core fields
	Website     string `json:"website,omitempty"`     // Organization website URL
	Description string `json:"description,omitempty"` // Organization description
	Industry    string `json:"industry,omitempty"`    // Primary industry classification

	// Size and financial information
	EstimatedNumEmployees int     `json:"estimated_num_employees,omitempty"` // Estimated number of employees
	EmployeeRange         string  `json:"employee_range,omitempty"`          // Employee count range (e.g., "1000-5000")
	AnnualRevenue         float64 `json:"annual_revenue,omitempty"`          // Annual revenue in USD
	RevenueRange          string  `json:"revenue_range,omitempty"`           // Revenue range (e.g., "$10M-$50M")

	// Geographic information
	Country       string `json:"country,omitempty"`        // Country where organization is based
	State         string `json:"state,omitempty"`          // State or region
	City          string `json:"city,omitempty"`           // City of headquarters
	StreetAddress string `json:"street_address,omitempty"` // Street address

	// Contact information
	Phone string `json:"phone,omitempty"` // Primary phone number
	Email string `json:"email,omitempty"` // Primary contact email

	// Social and web presence
	LinkedinURL string `json:"linkedin_url,omitempty"` // LinkedIn company page URL
	TwitterURL  string `json:"twitter_url,omitempty"`  // Twitter profile URL

	// Company details
	FoundedYear    int    `json:"founded_year,omitempty"`    // Year the organization was founded
	PubliclyTraded bool   `json:"publicly_traded,omitempty"` // Whether publicly traded
	TickerSymbol   string `json:"ticker_symbol,omitempty"`   // Stock ticker symbol

	// Enrichment metadata
	ApollioID        string  `json:"apollio_id,omitempty"`         // Apollo.io organization identifier
	EnrichmentSource string  `json:"enrichment_source,omitempty"`  // Source of enrichment data
	DataQualityScore float64 `json:"data_quality_score,omitempty"` // Data quality score (0-1)

	// Technology information
	Technologies []string `json:"technologies,omitempty"` // List of technologies used

	// Additional classification
	SubIndustries    []string `json:"sub_industries,omitempty"`    // Sub-industry classifications
	Keywords         []string `json:"keywords,omitempty"`          // Associated keywords
	OrganizationType string   `json:"organization_type,omitempty"` // Type of organization
}

// ToModel converts to a full Tabularium Organization.
func (o Organization) ToModel(username string) (*model.Organization, error) {
	if o.Domain == "" {
		return nil, fmt.Errorf("organization requires domain")
	}
	if o.Name == "" {
		return nil, fmt.Errorf("organization requires name")
	}

	org := model.NewOrganization(o.Domain, o.Name, username)

	// Set optional fields if provided
	if o.Website != "" {
		org.Website = &o.Website
	}
	if o.Description != "" {
		org.Description = &o.Description
	}
	if o.Industry != "" {
		org.Industry = &o.Industry
	}
	if o.EstimatedNumEmployees > 0 {
		org.EstimatedNumEmployees = &o.EstimatedNumEmployees
	}
	if o.EmployeeRange != "" {
		org.EmployeeRange = &o.EmployeeRange
	}
	if o.AnnualRevenue > 0 {
		org.AnnualRevenue = &o.AnnualRevenue
	}
	if o.RevenueRange != "" {
		org.RevenueRange = &o.RevenueRange
	}
	if o.Country != "" {
		org.Country = &o.Country
	}
	if o.State != "" {
		org.State = &o.State
	}
	if o.City != "" {
		org.City = &o.City
	}
	if o.StreetAddress != "" {
		org.StreetAddress = &o.StreetAddress
	}
	if o.Phone != "" {
		org.Phone = &o.Phone
	}
	if o.Email != "" {
		org.Email = &o.Email
	}
	if o.LinkedinURL != "" {
		org.LinkedinURL = &o.LinkedinURL
	}
	if o.TwitterURL != "" {
		org.TwitterURL = &o.TwitterURL
	}
	if o.FoundedYear > 0 {
		org.FoundedYear = &o.FoundedYear
	}
	if o.PubliclyTraded {
		org.PubliclyTraded = &o.PubliclyTraded
	}
	if o.TickerSymbol != "" {
		org.TickerSymbol = &o.TickerSymbol
	}
	if o.ApollioID != "" {
		org.ApollioID = &o.ApollioID
	}
	if o.EnrichmentSource != "" {
		org.EnrichmentSource = &o.EnrichmentSource
	}
	if o.DataQualityScore > 0 {
		org.DataQualityScore = &o.DataQualityScore
	}
	if len(o.Technologies) > 0 {
		org.Technologies = &o.Technologies
	}
	if len(o.SubIndustries) > 0 {
		org.SubIndustries = &o.SubIndustries
	}
	if len(o.Keywords) > 0 {
		org.Keywords = &o.Keywords
	}
	if o.OrganizationType != "" {
		org.OrganizationType = &o.OrganizationType
	}

	return org, nil
}

// OrganizationFromModel converts a full Tabularium Organization to the simplified external type.
func OrganizationFromModel(org *model.Organization) Organization {
	ext := Organization{
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

// derefString safely dereferences a string pointer, returning empty string if nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Webpage is a simplified webpage for external tool writers.
type Webpage struct {
	URL string `json:"url"` // The webpage URL
}

// Group implements Target interface.
func (w Webpage) Group() string {
	return w.URL
}

// Identifier implements Target interface.
func (w Webpage) Identifier() string {
	return w.URL
}

// ToTarget converts to a full Tabularium Webpage.
func (w Webpage) ToTarget() (model.Target, error) {
	if w.URL == "" {
		return nil, fmt.Errorf("webpage requires url")
	}

	webpage := model.NewWebpageFromString(w.URL, nil)
	if !webpage.Valid() {
		return nil, fmt.Errorf("invalid webpage url: %s", w.URL)
	}

	return &webpage, nil
}

// ToModel converts to a full Tabularium Webpage (convenience method).
func (w Webpage) ToModel() (*model.Webpage, error) {
	target, err := w.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.Webpage), nil
}

// WebpageFromModel converts a Tabularium Webpage to an external Webpage.
func WebpageFromModel(m *model.Webpage) Webpage {
	return Webpage{
		URL: m.URL,
	}
}

// ADObject is a simplified Active Directory object for external tool writers.
// It contains only the essential fields needed to identify an AD object.
type ADObject struct {
	Label             string `json:"label"`             // Primary label (ADUser, ADComputer, ADGroup, etc.)
	Domain            string `json:"domain"`            // AD domain
	ObjectID          string `json:"objectid"`          // Object identifier (SID or GUID)
	DistinguishedName string `json:"distinguishedname"` // DN path
}

// Group implements Target interface.
func (a ADObject) Group() string { return a.Domain }

// Identifier implements Target interface.
func (a ADObject) Identifier() string { return a.ObjectID }

// ToTarget converts to a full Tabularium ADObject.
func (a ADObject) ToTarget() (model.Target, error) {
	if a.Domain == "" {
		return nil, fmt.Errorf("adobject requires domain")
	}
	if a.ObjectID == "" {
		return nil, fmt.Errorf("adobject requires objectid")
	}

	label := a.Label
	if label == "" {
		label = model.ADObjectLabel
	}

	adObject := model.NewADObject(a.Domain, a.ObjectID, a.DistinguishedName, label)
	return &adObject, nil
}

// ToModel converts to a full Tabularium ADObject (convenience method).
func (a ADObject) ToModel() (*model.ADObject, error) {
	target, err := a.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.ADObject), nil
}

// ADObjectFromModel converts a Tabularium ADObject to an external ADObject.
func ADObjectFromModel(m *model.ADObject) ADObject {
	return ADObject{
		Label:             m.Label,
		Domain:            m.Domain,
		ObjectID:          m.ObjectID,
		DistinguishedName: m.DistinguishedName,
	}
}

// Technology is a simplified technology for external tool writers.
// It represents a specific technology (software, library, framework) identified on an asset.
type Technology struct {
	CPE  string `json:"cpe"`            // The full CPE string (e.g., "cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*")
	Name string `json:"name,omitempty"` // Optional common name for the technology (e.g., "Apache httpd")
}

// ToModel converts to a full Tabularium Technology.
func (t Technology) ToModel() (*model.Technology, error) {
	if t.CPE == "" {
		return nil, fmt.Errorf("technology requires cpe")
	}

	tech, err := model.NewTechnology(t.CPE)
	if err != nil {
		return nil, fmt.Errorf("invalid cpe: %w", err)
	}

	if t.Name != "" {
		tech.Name = t.Name
	}

	return &tech, nil
}

// TechnologyFromModel converts a Tabularium Technology to an external Technology.
func TechnologyFromModel(m *model.Technology) Technology {
	return Technology{
		CPE:  m.CPE,
		Name: m.Name,
	}
}

// Person is a simplified person for external tool writers.
// It contains essential fields for identifying and enriching person data.
type Person struct {
	Email            string `json:"email"`                       // Person's email address
	Name             string `json:"name"`                        // Person's full name
	Title            string `json:"title,omitempty"`             // Job title
	OrganizationName string `json:"organization_name,omitempty"` // Organization they work for
	LinkedinURL      string `json:"linkedin_url,omitempty"`      // LinkedIn profile URL
}

// Group implements Target interface.
func (p Person) Group() string { return p.Email }

// Identifier implements Target interface.
func (p Person) Identifier() string {
	if p.Email != "" {
		return fmt.Sprintf("#person#%s#%s", p.Email, p.Name)
	}
	return fmt.Sprintf("#person#%s#%s", p.Name, p.Name)
}

// ToTarget converts to a full Tabularium Person.
func (p Person) ToTarget() (model.Target, error) {
	if p.Email == "" && p.Name == "" {
		return nil, fmt.Errorf("person requires email or name")
	}

	var person *model.Person
	if p.Email != "" {
		person = model.NewPerson(p.Email, p.Name, "")
	} else {
		person = model.NewPersonFromName(p.Name, "")
	}

	if p.Title != "" {
		person.Title = &p.Title
	}
	if p.OrganizationName != "" {
		person.OrganizationName = &p.OrganizationName
	}
	if p.LinkedinURL != "" {
		person.LinkedinURL = &p.LinkedinURL
	}

	return person, nil
}

// ToModel converts to a full Tabularium Person (convenience method).
func (p Person) ToModel() (*model.Person, error) {
	target, err := p.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.Person), nil
}

// PersonFromModel converts a Tabularium Person to an external Person.
func PersonFromModel(m *model.Person) Person {
	ext := Person{}

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
