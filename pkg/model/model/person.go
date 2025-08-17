package model

import (
	"fmt"
	"regexp"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&Person{})
}

var (
	personKeyRegex = regexp.MustCompile(`^#person#[^#]+#[^#]+$`)
)

const PersonLabel = "Person"

// GetDescription returns a description for the Person model.
func (p *Person) GetDescription() string {
	return "Represents person data enriched from Apollo.io People Enrichment API, including contact details, employment history, and social profiles."
}

// Person represents enriched person data from Apollo.io People Enrichment API
type Person struct {
	registry.BaseModel
	Username string `neo4j:"username" json:"username" desc:"Chariot username associated with the person record." example:"user@example.com"`
	Key      string `neo4j:"key" json:"key" desc:"Unique key identifying the person." example:"#person#john.doe@example.com#John Doe"`

	// Core Person Information
	FirstName *string `neo4j:"first_name,omitempty" json:"first_name,omitempty" desc:"Person's first name." example:"John"`
	LastName  *string `neo4j:"last_name,omitempty" json:"last_name,omitempty" desc:"Person's last name." example:"Doe"`
	Name      *string `neo4j:"name,omitempty" json:"name,omitempty" desc:"Person's full name." example:"John Doe"`
	Email     *string `neo4j:"email,omitempty" json:"email,omitempty" desc:"Person's primary email address." example:"john.doe@example.com"`
	
	// Professional Information
	Title    *string `neo4j:"title,omitempty" json:"title,omitempty" desc:"Person's current job title." example:"Senior Software Engineer"`
	Headline *string `neo4j:"headline,omitempty" json:"headline,omitempty" desc:"Person's professional headline." example:"Senior Software Engineer at Example Corp"`
	
	// Contact Information
	Phone           *string   `neo4j:"phone,omitempty" json:"phone,omitempty" desc:"Person's phone number." example:"+1-555-123-4567"`
	PersonalEmails  *[]string `neo4j:"personal_emails,omitempty" json:"personal_emails,omitempty" desc:"List of personal email addresses." example:"[\"john.doe@gmail.com\", \"johndoe@yahoo.com\"]"`
	WorkEmail       *string   `neo4j:"work_email,omitempty" json:"work_email,omitempty" desc:"Person's work email address." example:"john.doe@company.com"`
	
	// Social Media and Online Presence
	LinkedinURL   *string `neo4j:"linkedin_url,omitempty" json:"linkedin_url,omitempty" desc:"LinkedIn profile URL." example:"https://www.linkedin.com/in/johndoe"`
	TwitterURL    *string `neo4j:"twitter_url,omitempty" json:"twitter_url,omitempty" desc:"Twitter profile URL." example:"https://twitter.com/johndoe"`
	FacebookURL   *string `neo4j:"facebook_url,omitempty" json:"facebook_url,omitempty" desc:"Facebook profile URL." example:"https://www.facebook.com/johndoe"`
	GithubURL     *string `neo4j:"github_url,omitempty" json:"github_url,omitempty" desc:"GitHub profile URL." example:"https://github.com/johndoe"`
	PhotoURL      *string `neo4j:"photo_url,omitempty" json:"photo_url,omitempty" desc:"Person's profile photo URL." example:"https://media.licdn.com/dms/image/123/profile-pic.jpg"`
	
	// Current Organization Information
	OrganizationID   *string `neo4j:"organization_id,omitempty" json:"organization_id,omitempty" desc:"Apollo.io organization ID where person currently works." example:"5e66b6381e05b4008c8331b8"`
	OrganizationName *string `neo4j:"organization_name,omitempty" json:"organization_name,omitempty" desc:"Name of organization where person currently works." example:"Example Corp"`
	
	// Employment History - JSON array of employment records
	EmploymentHistory *[]EmploymentRecord `neo4j:"employment_history,omitempty" json:"employment_history,omitempty" desc:"List of person's employment history records."`
	
	// Email Status and Validation
	EmailStatus                   *string  `neo4j:"email_status,omitempty" json:"email_status,omitempty" desc:"Email verification status." example:"verified"`
	ExtrapolatedEmailConfidence   *float64 `neo4j:"extrapolated_email_confidence,omitempty" json:"extrapolated_email_confidence,omitempty" desc:"Confidence score for extrapolated email." example:"0.95"`
	
	// Geographic Information
	Country *string `neo4j:"country,omitempty" json:"country,omitempty" desc:"Country where the person is located." example:"United States"`
	State   *string `neo4j:"state,omitempty" json:"state,omitempty" desc:"State or region where the person is located." example:"California"`
	City    *string `neo4j:"city,omitempty" json:"city,omitempty" desc:"City where the person is located." example:"San Francisco"`
	
	// Apollo.io Specific Metadata
	ApolloID            *string `neo4j:"apollo_id,omitempty" json:"apollo_id,omitempty" desc:"Apollo.io person identifier." example:"671bd2e8c2c9b5000169ba39"`
	LastEnrichedAt      *string `neo4j:"last_enriched_at,omitempty" json:"last_enriched_at,omitempty" desc:"Timestamp when data was last enriched from Apollo.io (RFC3339)." example:"2023-10-27T10:00:00Z"`
	EnrichmentSource    *string `neo4j:"enrichment_source,omitempty" json:"enrichment_source,omitempty" desc:"Source of enrichment data." example:"apollo.io"`
	DataQualityScore    *float64 `neo4j:"data_quality_score,omitempty" json:"data_quality_score,omitempty" desc:"Data quality score from Apollo.io." example:"0.92"`
	
	// Additional Professional Information
	Seniority        *string   `neo4j:"seniority,omitempty" json:"seniority,omitempty" desc:"Seniority level." example:"Senior"`
	Departments      *[]string `neo4j:"departments,omitempty" json:"departments,omitempty" desc:"List of departments person works in." example:"[\"Engineering\", \"Product\"]"`
	Functions        *[]string `neo4j:"functions,omitempty" json:"functions,omitempty" desc:"List of job functions." example:"[\"Software Development\", \"Technical Leadership\"]"`
	
	// Timestamps
	TTL     int64  `neo4j:"ttl" json:"ttl" desc:"Time-to-live for the person record (Unix timestamp)." example:"1706353200"`
	Created string `neo4j:"created" json:"created" desc:"Timestamp when the person record was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Visited string `neo4j:"visited" json:"visited" desc:"Timestamp when the person was last visited or updated (RFC3339)." example:"2023-10-27T11:00:00Z"`
	History
}

// EmploymentRecord represents a single employment history entry
type EmploymentRecord struct {
	ID               *string `json:"_id,omitempty" desc:"Employment record ID."`
	OrganizationID   *string `json:"organization_id,omitempty" desc:"Apollo.io organization ID."`
	OrganizationName *string `json:"organization_name,omitempty" desc:"Name of the organization."`
	Title            *string `json:"title,omitempty" desc:"Job title at this organization."`
	StartDate        *string `json:"start_date,omitempty" desc:"Employment start date (YYYY-MM-DD format)."`
	EndDate          *string `json:"end_date,omitempty" desc:"Employment end date (YYYY-MM-DD format). Null if current position."`
	Current          *bool   `json:"current,omitempty" desc:"Whether this is the current position."`
	Description      *string `json:"description,omitempty" desc:"Job description."`
	RawAddress       *string `json:"raw_address,omitempty" desc:"Raw address of the workplace."`
	CreatedAt        *string `json:"created_at,omitempty" desc:"Record creation timestamp (RFC3339)."`
	UpdatedAt        *string `json:"updated_at,omitempty" desc:"Record last update timestamp (RFC3339)."`
	Key              *string `json:"key,omitempty" desc:"Unique key for this employment record."`
}

func (p *Person) GetKey() string {
	return p.Key
}

func (p *Person) GetLabels() []string {
	return []string{PersonLabel, TTLLabel}
}

func (p *Person) Valid() bool {
	return personKeyRegex.MatchString(p.Key) && (p.Email != nil || p.Name != nil)
}

// Helper function to generate person key
func (p *Person) GenerateKey(email, name string) {
	if email != "" {
		p.Key = fmt.Sprintf("#person#%s#%s", email, name)
	} else {
		p.Key = fmt.Sprintf("#person#%s#%s", name, name)
	}
}

// NewPerson creates a new person record
func NewPerson(email, name, username string) *Person {
	person := &Person{
		Email:    &email,
		Name:     &name,
		Username: username,
	}
	person.GenerateKey(email, name)
	person.Created = Now()
	person.Visited = Now()
	return person
}

// NewPersonFromName creates a new person record with just a name (no email)
func NewPersonFromName(name, username string) *Person {
	person := &Person{
		Name:     &name,
		Username: username,
	}
	person.GenerateKey("", name)
	person.Created = Now()
	person.Visited = Now()
	return person
}