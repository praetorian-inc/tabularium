package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

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
