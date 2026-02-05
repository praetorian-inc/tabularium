package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/praetorian-inc/tabularium/pkg/registry"
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
type Port struct {
	Protocol string `json:"protocol"` // tcp or udp
	Port     int    `json:"port"`     // Port number
	Service  string `json:"service"`  // Service name (e.g., "https", "ssh")
	Parent   Asset  `json:"parent"`   // Parent asset
}

// Group implements Target interface.
func (p Port) Group() string { return p.Parent.DNS }

// Identifier implements Target interface.
func (p Port) Identifier() string {
	return fmt.Sprintf("%s:%d", p.Parent.Name, p.Port)
}

// ToTarget converts to a full Tabularium Port.
func (p Port) ToTarget() (model.Target, error) {
	if p.Protocol == "" {
		return nil, fmt.Errorf("port requires protocol")
	}
	if p.Port <= 0 || p.Port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535")
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

// Account is a simplified account for external tool writers.
// It represents cloud provider credentials and account information.
type Account struct {
	Name     string            `json:"name"`               // The owner of the account (e.g., "chariot.customer@example.com")
	Member   string            `json:"member"`             // The user or system granted access (e.g., "amazon", "azure")
	Value    string            `json:"value"`              // The identifier for this account within the context of member (e.g., "01234567890")
	Secret   map[string]string `json:"secret,omitempty"`   // Secret configuration map (e.g., credentials, tokens)
	Settings []byte            `json:"settings,omitempty"` // Raw JSON settings
}

// ToModel converts to a full Tabularium Account.
func (a Account) ToModel() (*model.Account, error) {
	if a.Name == "" {
		return nil, fmt.Errorf("account requires name")
	}
	if a.Member == "" {
		return nil, fmt.Errorf("account requires member")
	}
	if a.Value == "" {
		return nil, fmt.Errorf("account requires value")
	}

	var account model.Account
	if len(a.Settings) > 0 {
		account = model.NewAccountWithSettings(a.Name, a.Member, a.Value, a.Secret, a.Settings)
	} else {
		account = model.NewAccount(a.Name, a.Member, a.Value, a.Secret)
	}

	return &account, nil
}

// User is a simplified user for external tool writers.
type User struct {
	Name     string    `json:"name"`     // User email address
	Accounts []Account `json:"accounts"` // Accounts associated with the user
}

// ToModel converts to a full Tabularium User.
func (u User) ToModel() (*model.User, error) {
	if u.Name == "" {
		return nil, fmt.Errorf("user requires name")
	}

	accounts := make([]model.Account, 0, len(u.Accounts))
	for _, acc := range u.Accounts {
		modelAccount, err := acc.ToModel()
		if err != nil {
			return nil, fmt.Errorf("invalid account: %w", err)
		}
		accounts = append(accounts, *modelAccount)
	}

	user := model.NewUser(u.Name, accounts)
	return &user, nil
}

// FromUser creates an external User from a Tabularium User model.
func FromUser(u *model.User) User {
	if u == nil {
		return User{}
	}

	accounts := make([]Account, 0, len(u.Accounts))
	for _, acc := range u.Accounts {
		accounts = append(accounts, Account{
			Name:   acc.Name,
			Member: acc.Member,
			Value:  acc.Value,
		})
	}

	return User{
		Name:     u.Name,
		Accounts: accounts,
	}
}

// CloudResource is a simplified cloud resource for external tool writers.
// Note: CloudResource does not implement the Target interface directly since
// model.CloudResource is a base type. Use specific cloud resource types
// (AWSResource, GCPResource, AzureResource) if you need Target interface support.
type CloudResource struct {
	Name         string         `json:"name"`         // Resource name/ARN
	Provider     string         `json:"provider"`     // Cloud provider (e.g., "aws", "azure", "gcp")
	ResourceType string         `json:"resourceType"` // Type of resource (e.g., "AWS::EC2::Instance")
	DisplayName  string         `json:"displayName"`  // Human-readable display name
	Region       string         `json:"region"`       // Region where resource is located
	AccountRef   string         `json:"accountRef"`   // Account reference/ID
	Properties   map[string]any `json:"properties"`   // Additional resource properties
	IPs          []string       `json:"ips"`          // Associated IP addresses
	URLs         []string       `json:"urls"`         // Associated URLs
	Labels       []string       `json:"labels"`       // Resource labels
}

// Group returns the account reference (grouping identifier).
func (c CloudResource) Group() string { return c.AccountRef }

// Identifier returns the resource name.
func (c CloudResource) Identifier() string { return c.Name }

// ToModel converts to a full Tabularium CloudResource.
func (c CloudResource) ToModel() (*model.CloudResource, error) {
	if c.Name == "" {
		return nil, fmt.Errorf("cloud resource requires name")
	}
	if c.Provider == "" {
		return nil, fmt.Errorf("cloud resource requires provider")
	}
	if c.AccountRef == "" {
		return nil, fmt.Errorf("cloud resource requires accountRef")
	}

	cloudResource := &model.CloudResource{
		Name:         c.Name,
		Provider:     c.Provider,
		ResourceType: model.CloudResourceType(c.ResourceType),
		DisplayName:  c.DisplayName,
		Region:       c.Region,
		AccountRef:   c.AccountRef,
		Properties:   c.Properties,
		IPs:          c.IPs,
		URLs:         c.URLs,
		Labels:       c.Labels,
	}

	cloudResource.Defaulted()
	return cloudResource, nil
}

// CloudResourceFromModel creates an external CloudResource from a model CloudResource.
func CloudResourceFromModel(m *model.CloudResource) CloudResource {
	return CloudResource{
		Name:         m.Name,
		Provider:     m.Provider,
		ResourceType: m.ResourceType.String(),
		DisplayName:  m.DisplayName,
		Region:       m.Region,
		AccountRef:   m.AccountRef,
		Properties:   m.Properties,
		IPs:          m.IPs,
		URLs:         m.URLs,
		Labels:       m.Labels,
	}
}

// Integration is a simplified integration for external tool writers.
type Integration struct {
	Name   string `json:"name"`   // Integration name (e.g., "github", "slack")
	Value  string `json:"value"`  // Integration identifier/value
	Status string `json:"status"` // Integration status
}

// ToModel converts to a full Tabularium Integration.
func (i Integration) ToModel() (*model.Integration, error) {
	if i.Name == "" {
		return nil, fmt.Errorf("integration requires name")
	}
	if i.Value == "" {
		return nil, fmt.Errorf("integration requires value")
	}

	integration := model.NewIntegration(i.Name, i.Value)
	if i.Status != "" {
		integration.Status = i.Status
	}

	return &integration, nil
}

// FromIntegration creates an external Integration from a Tabularium Integration model.
func FromIntegration(i *model.Integration) Integration {
	if i == nil {
		return Integration{}
	}

	return Integration{
		Name:   i.Name,
		Value:  i.Value,
		Status: i.Status,
	}
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

// WebApplication is a simplified web application for external tool writers.
type WebApplication struct {
	PrimaryURL string   `json:"primary_url"` // The primary/canonical URL of the web application
	URLs       []string `json:"urls"`        // Additional URLs associated with this web application
	Name       string   `json:"name"`        // Name of the web application
	Status     string   `json:"status"`      // Status code (e.g., "A", "I", "P")
}

// Group implements Target interface.
func (w WebApplication) Group() string {
	return w.Name
}

// Identifier implements Target interface.
func (w WebApplication) Identifier() string {
	return w.PrimaryURL
}

// ToTarget converts to a full Tabularium WebApplication.
func (w WebApplication) ToTarget() (model.Target, error) {
	if w.PrimaryURL == "" {
		return nil, fmt.Errorf("webapplication requires primary_url")
	}

	name := w.Name
	if name == "" {
		name = w.PrimaryURL
	}

	// Create the webapp without calling NewWebApplication to avoid double normalization
	webapp := model.WebApplication{
		PrimaryURL: w.PrimaryURL,
		Name:       name,
		URLs:       w.URLs,
	}

	if w.Status != "" {
		webapp.Status = w.Status
	}

	// Apply defaults and run hooks (including URL normalization)
	webapp.Defaulted()
	if err := registry.CallHooks(&webapp); err != nil {
		return nil, fmt.Errorf("failed to initialize webapplication: %w", err)
	}

	return &webapp, nil
}

// ToModel converts to a full Tabularium WebApplication (convenience method).
func (w WebApplication) ToModel() (*model.WebApplication, error) {
	target, err := w.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.WebApplication), nil
}

// WebApplicationFromModel converts a Tabularium WebApplication to an external WebApplication.
func WebApplicationFromModel(m *model.WebApplication) WebApplication {
	return WebApplication{
		PrimaryURL: m.PrimaryURL,
		URLs:       m.URLs,
		Name:       m.Name,
		Status:     m.Status,
	}
}
