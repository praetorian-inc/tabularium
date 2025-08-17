package model

import (
	"fmt"
	"regexp"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&Organisation{})
}

var (
	organizationKeyRegex = regexp.MustCompile(`^#organization#[^#]+#[^#]+$`)
)

const OrganizationLabel = "Organization"

// GetDescription returns a description for the Organisation model.
func (o *Organisation) GetDescription() string {
	return "Represents organization data enriched from Apollo.io API, including company details, industry, revenue, employees, and contact information."
}

// Organisation represents enriched organization data from Apollo.io
type Organisation struct {
	registry.BaseModel
	Username string `neo4j:"username" json:"username" desc:"Chariot username associated with the organization record." example:"user@example.com"`
	Key      string `neo4j:"key" json:"key" desc:"Unique key identifying the organization." example:"#organization#example.com#Example Corp"`

	// Core Organization Information
	Name        *string `neo4j:"name,omitempty" json:"name,omitempty" desc:"Organization name." example:"Example Corporation"`
	Domain      *string `neo4j:"domain,omitempty" json:"domain,omitempty" desc:"Primary domain associated with the organization." example:"example.com"`
	Website     *string `neo4j:"website,omitempty" json:"website,omitempty" desc:"Organization website URL." example:"https://www.example.com"`
	Description *string `neo4j:"description,omitempty" json:"description,omitempty" desc:"Organization description." example:"Leading technology company providing innovative solutions."`

	// Industry and Classification
	Industry           *string   `neo4j:"industry,omitempty" json:"industry,omitempty" desc:"Primary industry classification." example:"Software & Technology"`
	SubIndustries      *[]string `neo4j:"sub_industries,omitempty" json:"sub_industries,omitempty" desc:"List of sub-industry classifications." example:"[\"Enterprise Software\", \"Cloud Computing\"]"`
	Keywords           *[]string `neo4j:"keywords,omitempty" json:"keywords,omitempty" desc:"Keywords associated with the organization." example:"[\"SaaS\", \"cloud\", \"enterprise\"]"`
	OrganizationType   *string   `neo4j:"organization_type,omitempty" json:"organization_type,omitempty" desc:"Type of organization." example:"Public Company"`
	BusinessModel      *string   `neo4j:"business_model,omitempty" json:"business_model,omitempty" desc:"Business model description." example:"B2B SaaS"`

	// Size and Financial Information
	EstimatedNumEmployees  *int     `neo4j:"estimated_num_employees,omitempty" json:"estimated_num_employees,omitempty" desc:"Estimated number of employees." example:"5000"`
	EmployeeRange          *string  `neo4j:"employee_range,omitempty" json:"employee_range,omitempty" desc:"Employee count range." example:"1000-5000"`
	AnnualRevenue          *float64 `neo4j:"annual_revenue,omitempty" json:"annual_revenue,omitempty" desc:"Annual revenue in USD." example:"50000000"`
	RevenueRange           *string  `neo4j:"revenue_range,omitempty" json:"revenue_range,omitempty" desc:"Revenue range description." example:"$10M-$50M"`
	MarketCapitalization   *float64 `neo4j:"market_capitalization,omitempty" json:"market_capitalization,omitempty" desc:"Market capitalization in USD." example:"1000000000"`

	// Geographic Information
	Country      *string `neo4j:"country,omitempty" json:"country,omitempty" desc:"Country where the organization is based." example:"United States"`
	State        *string `neo4j:"state,omitempty" json:"state,omitempty" desc:"State or region." example:"California"`
	City         *string `neo4j:"city,omitempty" json:"city,omitempty" desc:"City where the organization is headquartered." example:"San Francisco"`
	PostalCode   *string `neo4j:"postal_code,omitempty" json:"postal_code,omitempty" desc:"Postal or ZIP code." example:"94105"`
	StreetAddress *string `neo4j:"street_address,omitempty" json:"street_address,omitempty" desc:"Street address of headquarters." example:"123 Market Street"`

	// Contact Information
	Phone       *string `neo4j:"phone,omitempty" json:"phone,omitempty" desc:"Primary phone number." example:"+1-555-123-4567"`
	Fax         *string `neo4j:"fax,omitempty" json:"fax,omitempty" desc:"Fax number." example:"+1-555-123-4568"`
	Email       *string `neo4j:"email,omitempty" json:"email,omitempty" desc:"Primary contact email." example:"contact@example.com"`

	// Company Identifiers
	LinkedinURL  *string `neo4j:"linkedin_url,omitempty" json:"linkedin_url,omitempty" desc:"LinkedIn company page URL." example:"https://www.linkedin.com/company/example-corp"`
	TwitterURL   *string `neo4j:"twitter_url,omitempty" json:"twitter_url,omitempty" desc:"Twitter profile URL." example:"https://twitter.com/examplecorp"`
	FacebookURL  *string `neo4j:"facebook_url,omitempty" json:"facebook_url,omitempty" desc:"Facebook page URL." example:"https://www.facebook.com/examplecorp"`
	BlogURL      *string `neo4j:"blog_url,omitempty" json:"blog_url,omitempty" desc:"Company blog URL." example:"https://blog.example.com"`

	// Founded and Status
	FoundedYear    *int    `neo4j:"founded_year,omitempty" json:"founded_year,omitempty" desc:"Year the organization was founded." example:"2010"`
	PubliclyTraded *bool   `neo4j:"publicly_traded,omitempty" json:"publicly_traded,omitempty" desc:"Whether the organization is publicly traded." example:"true"`
	TickerSymbol   *string `neo4j:"ticker_symbol,omitempty" json:"ticker_symbol,omitempty" desc:"Stock ticker symbol if publicly traded." example:"EXMP"`
	Exchange       *string `neo4j:"exchange,omitempty" json:"exchange,omitempty" desc:"Stock exchange where shares are traded." example:"NASDAQ"`

	// Apollo.io Specific
	ApollioID           *string `neo4j:"apollio_id,omitempty" json:"apollio_id,omitempty" desc:"Apollo.io organization identifier." example:"apollo123456"`
	LastEnrichedAt      *string `neo4j:"last_enriched_at,omitempty" json:"last_enriched_at,omitempty" desc:"Timestamp when data was last enriched from Apollo.io (RFC3339)." example:"2023-10-27T10:00:00Z"`
	EnrichmentSource    *string `neo4j:"enrichment_source,omitempty" json:"enrichment_source,omitempty" desc:"Source of enrichment data." example:"apollo.io"`
	DataQualityScore    *float64 `neo4j:"data_quality_score,omitempty" json:"data_quality_score,omitempty" desc:"Data quality score from Apollo.io." example:"0.95"`

	// Technology Stack Information
	Technologies *[]string `neo4j:"technologies,omitempty" json:"technologies,omitempty" desc:"List of technologies used by the organization." example:"[\"Salesforce\", \"AWS\", \"Docker\"]"`
	TechCategories *[]string `neo4j:"tech_categories,omitempty" json:"tech_categories,omitempty" desc:"List of technology categories." example:"[\"CRM\", \"Cloud\", \"DevOps\"]"`
	TechVendors *[]string `neo4j:"tech_vendors,omitempty" json:"tech_vendors,omitempty" desc:"List of technology vendors." example:"[\"Salesforce.com\", \"Amazon\", \"Docker Inc\"]"`

	// Additional Contact Information
	AlternatePhones *[]string `neo4j:"alternate_phones,omitempty" json:"alternate_phones,omitempty" desc:"List of additional phone numbers." example:"[\"+1-555-123-4567\", \"+1-555-987-6543\"]"`
	PhoneTypes *[]string `neo4j:"phone_types,omitempty" json:"phone_types,omitempty" desc:"List of phone types corresponding to alternate phones." example:"[\"main\", \"fax\"]"`

	// Funding Information
	FundingRounds *[]string `neo4j:"funding_rounds,omitempty" json:"funding_rounds,omitempty" desc:"List of funding round types." example:"[\"Seed\", \"Series A\", \"Series B\"]"`
	FundingAmounts *[]float64 `neo4j:"funding_amounts,omitempty" json:"funding_amounts,omitempty" desc:"List of funding amounts in USD." example:"[1000000, 5000000, 15000000]"`
	Investors *[]string `neo4j:"investors,omitempty" json:"investors,omitempty" desc:"List of investors." example:"[\"Accel Partners\", \"Sequoia Capital\", \"Greylock Partners\"]"`

	// Additional Address Information
	AdditionalAddresses *[]string `neo4j:"additional_addresses,omitempty" json:"additional_addresses,omitempty" desc:"List of additional office addresses." example:"[\"456 Oak St, New York, NY\", \"789 Pine Ave, Austin, TX\"]"`
	AddressTypes *[]string `neo4j:"address_types,omitempty" json:"address_types,omitempty" desc:"List of address types." example:"[\"headquarters\", \"office\", \"branch\"]"`

	// Timestamps
	TTL     int64  `neo4j:"ttl" json:"ttl" desc:"Time-to-live for the organization record (Unix timestamp)." example:"1706353200"`
	Created string `neo4j:"created" json:"created" desc:"Timestamp when the organization record was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Visited string `neo4j:"visited" json:"visited" desc:"Timestamp when the organization was last visited or updated (RFC3339)." example:"2023-10-27T11:00:00Z"`
	History
}


func (o *Organisation) GetKey() string {
	return o.Key
}

func (o *Organisation) GetLabels() []string {
	return []string{OrganizationLabel, TTLLabel}
}

func (o *Organisation) Valid() bool {
	return organizationKeyRegex.MatchString(o.Key) && o.Domain != nil
}

// Helper function to generate organization key
func (o *Organisation) GenerateKey(domain, name string) {
	o.Key = fmt.Sprintf("#organization#%s#%s", domain, name)
}

// NewOrganisation creates a new organization record
func NewOrganisation(domain, name, username string) *Organisation {
	org := &Organisation{
		Domain:   &domain,
		Name:     &name,
		Username: username,
	}
	org.GenerateKey(domain, name)
	org.Created = Now()
	org.Visited = Now()
	return org
}