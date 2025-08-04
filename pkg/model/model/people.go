package model

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/nyaruka/phonenumbers"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&Person{})
	registry.Registry.MustRegisterModel(&PersonEmail{})
	registry.Registry.MustRegisterModel(&PersonPhone{})
	registry.Registry.MustRegisterModel(&PersonUsername{})
	registry.Registry.MustRegisterModel(&PersonWebsite{})
}

const (
	PersonLabel = "Person"

	// Email types
	EmailTypeWork     = "work"
	EmailTypePersonal = "personal"
	EmailTypeOther    = "other"

	// Phone types
	PhoneTypeWork   = "work"
	PhoneTypeMobile = "mobile"
	PhoneTypeHome   = "home"
	PhoneTypeOther  = "other"

	// Username platforms
	PlatformDomain  = "domain"  // Active Directory/Corporate domain
	PlatformIdP     = "idp"     // Identity Provider (SAML, OIDC)
	PlatformWindows = "windows" // Windows domain
	PlatformEmail   = "email"   // Email system username
	PlatformOther   = "other"

	// Website types
	WebsiteLinkedIn  = "linkedin"
	WebsitePersonal  = "personal"
	WebsitePortfolio = "portfolio"
	WebsiteGitHub    = "github"
	WebsiteTwitter   = "twitter"
	WebsiteOther     = "other"

	// Seniority levels
	SeniorityEntry     = "entry"
	SeniorityMid       = "mid"
	SenioritSenior     = "senior"
	SeniorityExecutive = "executive"
	SeniorityCLevel    = "c-level"

	// Access levels
	AccessLevelUser       = "user"
	AccessLevelAdmin      = "admin"
	AccessLevelPrivileged = "privileged"
	AccessLevelReadOnly   = "readonly"
)

var (
	personKey     = regexp.MustCompile(`^#person#([^#]+)#([^#]+)$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	validEmailTypes = map[string]bool{
		EmailTypeWork:     true,
		EmailTypePersonal: true,
		EmailTypeOther:    true,
	}

	validPhoneTypes = map[string]bool{
		PhoneTypeWork:   true,
		PhoneTypeMobile: true,
		PhoneTypeHome:   true,
		PhoneTypeOther:  true,
	}

	validUsernamePlatforms = map[string]bool{
		PlatformDomain:  true,
		PlatformIdP:     true,
		PlatformWindows: true,
		PlatformEmail:   true,
		PlatformOther:   true,
	}

	validWebsiteTypes = map[string]bool{
		WebsiteLinkedIn:  true,
		WebsitePersonal:  true,
		WebsitePortfolio: true,
		WebsiteGitHub:    true,
		WebsiteTwitter:   true,
		WebsiteOther:     true,
	}

	validSeniorityLevels = map[string]bool{
		SeniorityEntry:     true,
		SeniorityMid:       true,
		SenioritSenior:     true,
		SeniorityExecutive: true,
		SeniorityCLevel:    true,
	}

	validAccessLevels = map[string]bool{
		AccessLevelUser:       true,
		AccessLevelAdmin:      true,
		AccessLevelPrivileged: true,
		AccessLevelReadOnly:   true,
	}
)

// Person represents a person asset with comprehensive CRM-style information
type Person struct {
	BaseAsset
	// Core identity (required field: FullName)
	FullName   string `neo4j:"fullName" json:"fullName" desc:"Full name of the person." example:"John Michael Smith"`
	FirstName  string `neo4j:"firstName" json:"firstName,omitempty" desc:"First name." example:"John"`
	MiddleName string `neo4j:"middleName" json:"middleName,omitempty" desc:"Middle name." example:"Michael"`
	LastName   string `neo4j:"lastName" json:"lastName,omitempty" desc:"Last name." example:"Smith"`

	// Professional context
	JobTitle        string `neo4j:"jobTitle" json:"jobTitle,omitempty" desc:"Job title or position." example:"Senior Software Engineer"`
	JobDescription  string `neo4j:"jobDescription" json:"jobDescription,omitempty" desc:"Detailed job description." example:"Responsible for developing secure applications..."`
	Department      string `neo4j:"department" json:"department,omitempty" desc:"Department or division." example:"Engineering"`
	Location        string `neo4j:"location" json:"location,omitempty" desc:"Geographic location." example:"San Francisco, CA"`
	Industry        string `neo4j:"industry" json:"industry,omitempty" desc:"Industry sector." example:"Technology"`
	SeniorityLevel  string `neo4j:"seniorityLevel" json:"seniorityLevel,omitempty" desc:"Seniority level." example:"senior"`
	YearsExperience int    `neo4j:"yearsExperience" json:"yearsExperience,omitempty" desc:"Years of professional experience." example:"8"`
	CompanySize     string `neo4j:"companySize" json:"companySize,omitempty" desc:"Size of current company." example:"201-500"`

	// Social & Professional Presence
	LinkedInURL     string   `neo4j:"linkedInURL" json:"linkedInURL,omitempty" desc:"LinkedIn profile URL." example:"https://linkedin.com/in/johnsmith"`
	ProfileImageURL string   `neo4j:"profileImageURL" json:"profileImageURL,omitempty" desc:"Profile image URL." example:"https://media.licdn.com/..."`
	Bio             string   `neo4j:"bio" json:"bio,omitempty" desc:"Professional biography or summary."`
	NetworkSize     int      `neo4j:"networkSize" json:"networkSize,omitempty" desc:"Size of professional network." example:"500"`
	Skills          []string `neo4j:"skills" json:"skills,omitempty" desc:"Technical and professional skills." example:"[\"Go\", \"Python\", \"Security\"]"`
	Languages       []string `neo4j:"languages" json:"languages,omitempty" desc:"Spoken languages." example:"[\"English\", \"Spanish\"]"`

	// Security Assessment Context
	SecurityClearance string `neo4j:"securityClearance" json:"securityClearance,omitempty" desc:"Security clearance level." example:"secret"`
	AccessLevel       string `neo4j:"accessLevel" json:"accessLevel,omitempty" desc:"Access level in organization." example:"admin"`
	IsDecisionMaker   bool   `neo4j:"isDecisionMaker" json:"isDecisionMaker,omitempty" desc:"Has budget or purchasing authority." example:"true"`
	LastSeenActive    string `neo4j:"lastSeenActive" json:"lastSeenActive,omitempty" desc:"Last known activity timestamp (RFC3339)." example:"2023-10-27T10:00:00Z"`

	// Structured attachments (CRM-style data)
	Emails    []PersonEmail    `neo4j:"-" json:"emails,omitempty" desc:"Email addresses associated with the person."`
	Phones    []PersonPhone    `neo4j:"-" json:"phones,omitempty" desc:"Phone numbers associated with the person."`
	Usernames []PersonUsername `neo4j:"-" json:"usernames,omitempty" desc:"Usernames on various platforms."`
	Websites  []PersonWebsite  `neo4j:"-" json:"websites,omitempty" desc:"Personal and professional websites."`

	// Temporary compatibility field (will be removed when PersonWorksFor relationships are fully integrated)
	tempCompany string `neo4j:"-" json:"-"`
}

// PersonAttachment interface for extensible data items
type PersonAttachment interface {
	GetType() string
	GetValue() string
	IsCurrentlyActive() bool
	GetSource() string
	Valid() bool
}

// PersonEmail represents an email address attachment
type PersonEmail struct {
	registry.BaseModel
	Email     string `json:"email" desc:"Email address." example:"john.smith@company.com"`
	Type      string `json:"type" desc:"Type of email address." example:"work"`
	IsPrimary bool   `json:"isPrimary" desc:"Whether this is the primary email." example:"true"`
	IsActive  bool   `json:"isActive" desc:"Whether this email is currently active." example:"true"`
	Source    string `json:"source" desc:"Source where this email was discovered." example:"linkedin"`
	DateAdded string `json:"dateAdded" desc:"When this email was added (RFC3339)." example:"2023-10-27T10:00:00Z"`
}

// PersonPhone represents a phone number attachment
type PersonPhone struct {
	registry.BaseModel
	Number    string `json:"number" desc:"Phone number in international format." example:"+14155551234"`
	Type      string `json:"type" desc:"Type of phone number." example:"work"`
	Country   string `json:"country" desc:"Country code (ISO 3166-1 alpha-2)." example:"US"`
	IsPrimary bool   `json:"isPrimary" desc:"Whether this is the primary phone." example:"true"`
	IsActive  bool   `json:"isActive" desc:"Whether this phone is currently active." example:"true"`
	Source    string `json:"source" desc:"Source where this phone was discovered." example:"zoominfo"`
	DateAdded string `json:"dateAdded" desc:"When this phone was added (RFC3339)." example:"2023-10-27T10:00:00Z"`
}

// PersonUsername represents a username on a platform
type PersonUsername struct {
	registry.BaseModel
	Username  string `json:"username" desc:"Username or account identifier." example:"jsmith"`
	Platform  string `json:"platform" desc:"Platform or system type." example:"domain"`
	Domain    string `json:"domain" desc:"Domain or organization identifier." example:"company.com"`
	IsActive  bool   `json:"isActive" desc:"Whether this username is currently active." example:"true"`
	Source    string `json:"source" desc:"Source where this username was discovered." example:"manual"`
	DateAdded string `json:"dateAdded" desc:"When this username was added (RFC3339)." example:"2023-10-27T10:00:00Z"`
}

// PersonWebsite represents a website or social media profile
type PersonWebsite struct {
	registry.BaseModel
	URL       string `json:"url" desc:"Website or profile URL." example:"https://linkedin.com/in/johnsmith"`
	Type      string `json:"type" desc:"Type of website." example:"linkedin"`
	IsActive  bool   `json:"isActive" desc:"Whether this website is currently accessible." example:"true"`
	Source    string `json:"source" desc:"Source where this website was discovered." example:"linkedin"`
	DateAdded string `json:"dateAdded" desc:"When this website was added (RFC3339)." example:"2023-10-27T10:00:00Z"`
}

// Interface implementations for PersonAttachment

func (pe *PersonEmail) GetType() string         { return pe.Type }
func (pe *PersonEmail) GetValue() string        { return pe.Email }
func (pe *PersonEmail) IsCurrentlyActive() bool { return pe.IsActive }
func (pe *PersonEmail) GetSource() string       { return pe.Source }

func (pp *PersonPhone) GetType() string         { return pp.Type }
func (pp *PersonPhone) GetValue() string        { return pp.Number }
func (pp *PersonPhone) IsCurrentlyActive() bool { return pp.IsActive }
func (pp *PersonPhone) GetSource() string       { return pp.Source }

func (pu *PersonUsername) GetType() string         { return pu.Platform }
func (pu *PersonUsername) GetValue() string        { return pu.Username }
func (pu *PersonUsername) IsCurrentlyActive() bool { return pu.IsActive }
func (pu *PersonUsername) GetSource() string       { return pu.Source }

func (pw *PersonWebsite) GetType() string         { return pw.Type }
func (pw *PersonWebsite) GetValue() string        { return pw.URL }
func (pw *PersonWebsite) IsCurrentlyActive() bool { return pw.IsActive }
func (pw *PersonWebsite) GetSource() string       { return pw.Source }

// Validation methods

func (pe *PersonEmail) Valid() bool {
	if pe.Email == "" || !emailRegex.MatchString(pe.Email) {
		return false
	}
	if !validEmailTypes[pe.Type] {
		return false
	}
	if pe.DateAdded == "" {
		return false
	}
	return true
}

func (pp *PersonPhone) Valid() bool {
	if pp.Number == "" {
		return false
	}

	// Validate phone number using libphonenumber
	parsedNumber, err := phonenumbers.Parse(pp.Number, pp.Country)
	if err != nil {
		return false
	}
	if !phonenumbers.IsValidNumber(parsedNumber) {
		return false
	}

	if !validPhoneTypes[pp.Type] {
		return false
	}
	if pp.Country == "" || len(pp.Country) != 2 {
		return false
	}
	if pp.DateAdded == "" {
		return false
	}
	return true
}

func (pu *PersonUsername) Valid() bool {
	if pu.Username == "" || !usernameRegex.MatchString(pu.Username) {
		return false
	}
	if !validUsernamePlatforms[pu.Platform] {
		return false
	}
	if pu.DateAdded == "" {
		return false
	}
	return true
}

func (pw *PersonWebsite) Valid() bool {
	if pw.URL == "" {
		return false
	}

	// Validate URL format - must have scheme and host
	parsedURL, err := url.Parse(pw.URL)
	if err != nil {
		return false
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	if !validWebsiteTypes[pw.Type] {
		return false
	}
	if pw.DateAdded == "" {
		return false
	}
	return true
}

// Person model methods

func (p *Person) IsPrivate() bool {
	// People are typically considered private/sensitive data
	return true
}

func (p *Person) GetKey() string {
	return p.Key
}

func (p *Person) Valid() bool {
	if p.FullName == "" {
		return false
	}
	if !personKey.MatchString(p.Key) {
		return false
	}

	// Validate seniority level if provided
	if p.SeniorityLevel != "" && !validSeniorityLevels[p.SeniorityLevel] {
		return false
	}

	// Validate access level if provided
	if p.AccessLevel != "" && !validAccessLevels[p.AccessLevel] {
		return false
	}

	// Validate all attachments
	for _, email := range p.Emails {
		if !email.Valid() {
			return false
		}
	}
	for _, phone := range p.Phones {
		if !phone.Valid() {
			return false
		}
	}
	for _, username := range p.Usernames {
		if !username.Valid() {
			return false
		}
	}
	for _, website := range p.Websites {
		if !website.Valid() {
			return false
		}
	}

	return true
}

func (p *Person) GetLabels() []string {
	return []string{PersonLabel, AssetLabel, TTLLabel}
}

func (p *Person) GetClass() string {
	return "person"
}

func (p *Person) GetStatus() string {
	return p.Status
}

func (p *Person) Group() string {
	return p.GetCurrentCompany()
}

// GetCurrentCompany returns the company from the most recent active PersonWorksFor relationship
// Returns "unknown" if no current employment relationship exists
func (p *Person) GetCurrentCompany() string {
	// Note: This would typically query PersonWorksFor relationships from the graph
	// For now, returning "unknown" as a placeholder since relationships are stored separately
	// In a full implementation, this would query the graph for active PersonWorksFor relationships

	// Temporary compatibility: check if we have a temporary company set
	if p.tempCompany != "" {
		return p.tempCompany
	}
	return "unknown"
}

// SetCurrentCompany is a temporary helper method for compatibility during transition
// In production, company information should be set via PersonWorksFor relationships
func (p *Person) SetCurrentCompany(company string) {
	p.tempCompany = company
}

func (p *Person) Identifier() string {
	return p.FullName
}

func (p *Person) IsStatus(value string) bool {
	return strings.HasPrefix(p.Status, value)
}

func (p *Person) WithStatus(status string) Target {
	ret := *p
	ret.Status = status
	return &ret
}

func (p *Person) Attribute(name, value string) Attribute {
	attr := NewAttribute(name, value, p)
	return attr
}

func (p *Person) Seed() Seed {
	s := NewSeed(p.FullName)
	s.SetStatus(p.Status)
	return s
}

func (p *Person) Defaulted() {
	p.BaseAsset.Defaulted()
	p.Class = "person"

	// Set default timestamp for attachments that don't have one
	now := Now()
	for i := range p.Emails {
		if p.Emails[i].DateAdded == "" {
			p.Emails[i].DateAdded = now
		}
		if p.Emails[i].IsActive == false {
			p.Emails[i].IsActive = true
		}
	}
	for i := range p.Phones {
		if p.Phones[i].DateAdded == "" {
			p.Phones[i].DateAdded = now
		}
		if p.Phones[i].IsActive == false {
			p.Phones[i].IsActive = true
		}
	}
	for i := range p.Usernames {
		if p.Usernames[i].DateAdded == "" {
			p.Usernames[i].DateAdded = now
		}
		if p.Usernames[i].IsActive == false {
			p.Usernames[i].IsActive = true
		}
	}
	for i := range p.Websites {
		if p.Websites[i].DateAdded == "" {
			p.Websites[i].DateAdded = now
		}
		if p.Websites[i].IsActive == false {
			p.Websites[i].IsActive = true
		}
	}
}

func (p *Person) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				// Generate key based on normalized name and company
				normalizedName := NormalizePersonName(p.FullName)
				currentCompany := p.GetCurrentCompany()
				normalizedCompany := NormalizeCompanyName(currentCompany)
				p.Key = fmt.Sprintf("#person#%s#%s", normalizedName, normalizedCompany)
				p.BaseAsset.Identifier = p.FullName
				p.BaseAsset.Group = currentCompany

				// Parse structured name if not already set
				if p.FirstName == "" || p.LastName == "" {
					p.parseStructuredName()
				}

				return nil
			},
		},
	}
}

func (p *Person) GetDescription() string {
	return "Represents a person asset with comprehensive CRM-style information for security assessments and human attack surface mapping."
}

// Helper methods

func (p *Person) parseStructuredName() {
	parts := strings.Fields(strings.TrimSpace(p.FullName))
	if len(parts) == 0 {
		return
	}

	if len(parts) == 1 {
		p.FirstName = parts[0]
	} else if len(parts) == 2 {
		p.FirstName = parts[0]
		p.LastName = parts[1]
	} else if len(parts) >= 3 {
		p.FirstName = parts[0]
		p.LastName = parts[len(parts)-1]
		// Everything in between is middle name
		if len(parts) > 2 {
			p.MiddleName = strings.Join(parts[1:len(parts)-1], " ")
		}
	}
}

func (p *Person) GetCanonicalName() string {
	if p.LastName == "" {
		return p.FullName
	}

	canonical := p.LastName
	if p.FirstName != "" {
		canonical += ", " + p.FirstName
	}
	if p.MiddleName != "" {
		canonical += " " + p.MiddleName
	}
	return canonical
}

// Attachment management methods

func (p *Person) AddEmail(email, emailType, source string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}
	if !validEmailTypes[emailType] {
		return fmt.Errorf("invalid email type: %s", emailType)
	}

	// Check for duplicates
	for _, existing := range p.Emails {
		if existing.Email == email {
			return fmt.Errorf("email already exists: %s", email)
		}
	}

	p.Emails = append(p.Emails, PersonEmail{
		Email:     email,
		Type:      emailType,
		IsPrimary: len(p.Emails) == 0, // First email is primary
		IsActive:  true,
		Source:    source,
		DateAdded: Now(),
	})

	return nil
}

func (p *Person) AddPhone(number, phoneType, country, source string) error {
	if number == "" {
		return fmt.Errorf("phone number cannot be empty")
	}
	if !validPhoneTypes[phoneType] {
		return fmt.Errorf("invalid phone type: %s", phoneType)
	}
	if len(country) != 2 {
		return fmt.Errorf("country must be 2-letter ISO code")
	}

	// Validate phone number
	parsedNumber, err := phonenumbers.Parse(number, country)
	if err != nil {
		return fmt.Errorf("invalid phone number: %v", err)
	}
	if !phonenumbers.IsValidNumber(parsedNumber) {
		return fmt.Errorf("phone number is not valid")
	}

	// Format to international format
	formattedNumber := phonenumbers.Format(parsedNumber, phonenumbers.E164)

	// Check for duplicates
	for _, existing := range p.Phones {
		if existing.Number == formattedNumber {
			return fmt.Errorf("phone number already exists: %s", formattedNumber)
		}
	}

	p.Phones = append(p.Phones, PersonPhone{
		Number:    formattedNumber,
		Type:      phoneType,
		Country:   country,
		IsPrimary: len(p.Phones) == 0, // First phone is primary
		IsActive:  true,
		Source:    source,
		DateAdded: Now(),
	})

	return nil
}

func (p *Person) AddUsername(username, platform, domain, source string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("invalid username format: %s", username)
	}
	if !validUsernamePlatforms[platform] {
		return fmt.Errorf("invalid platform: %s", platform)
	}

	// Check for duplicates
	for _, existing := range p.Usernames {
		if existing.Username == username && existing.Platform == platform && existing.Domain == domain {
			return fmt.Errorf("username already exists: %s@%s:%s", username, domain, platform)
		}
	}

	p.Usernames = append(p.Usernames, PersonUsername{
		Username:  username,
		Platform:  platform,
		Domain:    domain,
		IsActive:  true,
		Source:    source,
		DateAdded: Now(),
	})

	return nil
}

func (p *Person) AddWebsite(websiteURL, websiteType, source string) error {
	if websiteURL == "" {
		return fmt.Errorf("website URL cannot be empty")
	}

	// Validate URL format - must have scheme and host
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid URL format: missing scheme or host")
	}

	if !validWebsiteTypes[websiteType] {
		return fmt.Errorf("invalid website type: %s", websiteType)
	}

	// Check for duplicates
	for _, existing := range p.Websites {
		if existing.URL == websiteURL {
			return fmt.Errorf("website already exists: %s", websiteURL)
		}
	}

	p.Websites = append(p.Websites, PersonWebsite{
		URL:       websiteURL,
		Type:      websiteType,
		IsActive:  true,
		Source:    source,
		DateAdded: Now(),
	})

	return nil
}

// Query methods

func (p *Person) GetActiveEmails() []string {
	var emails []string
	for _, email := range p.Emails {
		if email.IsActive {
			emails = append(emails, email.Email)
		}
	}
	sort.Strings(emails)
	return emails
}

func (p *Person) GetActivePhones() []string {
	var phones []string
	for _, phone := range p.Phones {
		if phone.IsActive {
			phones = append(phones, phone.Number)
		}
	}
	sort.Strings(phones)
	return phones
}

func (p *Person) GetActiveUsernames() []string {
	var usernames []string
	for _, username := range p.Usernames {
		if username.IsActive {
			if username.Domain != "" {
				usernames = append(usernames, fmt.Sprintf("%s@%s", username.Username, username.Domain))
			} else {
				usernames = append(usernames, username.Username)
			}
		}
	}
	sort.Strings(usernames)
	return usernames
}

func (p *Person) GetPrimaryEmail() string {
	for _, email := range p.Emails {
		if email.IsPrimary && email.IsActive {
			return email.Email
		}
	}
	// Return first active email if no primary set
	for _, email := range p.Emails {
		if email.IsActive {
			return email.Email
		}
	}
	return ""
}

func (p *Person) GetPrimaryPhone() string {
	for _, phone := range p.Phones {
		if phone.IsPrimary && phone.IsActive {
			return phone.Number
		}
	}
	// Return first active phone if no primary set
	for _, phone := range p.Phones {
		if phone.IsActive {
			return phone.Number
		}
	}
	return ""
}

// Normalization functions

func NormalizePersonName(fullName string) string {
	// Convert to lowercase and remove extra spaces
	normalized := strings.ToLower(strings.TrimSpace(fullName))

	// Basic transliteration of common accented characters
	replacements := map[string]string{
		"á": "a", "à": "a", "ä": "a", "â": "a", "ã": "a", "å": "a",
		"é": "e", "è": "e", "ë": "e", "ê": "e",
		"í": "i", "ì": "i", "ï": "i", "î": "i",
		"ó": "o", "ò": "o", "ö": "o", "ô": "o", "õ": "o", "ø": "o",
		"ú": "u", "ù": "u", "ü": "u", "û": "u",
		"ñ": "n", "ç": "c",
		"ý": "y", "ÿ": "y",
	}

	for accented, plain := range replacements {
		normalized = strings.ReplaceAll(normalized, accented, plain)
	}

	// Remove special characters, keep only letters, numbers, spaces
	normalized = regexp.MustCompile(`[^a-z0-9\s]`).ReplaceAllString(normalized, "")
	// Collapse multiple spaces
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, "")
	return normalized
}

func NormalizeCompanyName(company string) string {
	if company == "" {
		return "unknown"
	}
	// Convert to lowercase and remove spaces
	normalized := strings.ToLower(strings.TrimSpace(company))
	normalized = regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(normalized, "")
	return normalized
}

// Constructor

func NewPerson(fullName string) Person {
	person := Person{
		FullName: fullName,
	}
	person.Defaulted()
	registry.CallHooks(&person)
	return person
}

// PersonSearchService provides search and discovery capabilities
type PersonSearchService struct {
	People map[string]*Person // Keyed by normalized full name + company
}

func NewPersonSearchService() *PersonSearchService {
	return &PersonSearchService{
		People: make(map[string]*Person),
	}
}

func (pss *PersonSearchService) AddPerson(person *Person) {
	key := fmt.Sprintf("%s:%s", NormalizePersonName(person.FullName), NormalizeCompanyName(person.GetCurrentCompany()))
	pss.People[key] = person
}

func (pss *PersonSearchService) FindByFullName(fullName, company string) *Person {
	key := fmt.Sprintf("%s:%s", NormalizePersonName(fullName), NormalizeCompanyName(company))
	return pss.People[key]
}

func (pss *PersonSearchService) FindByEmail(email string) []*Person {
	var results []*Person
	email = strings.ToLower(email)

	for _, person := range pss.People {
		for _, personEmail := range person.Emails {
			if strings.ToLower(personEmail.Email) == email && personEmail.IsActive {
				results = append(results, person)
				break
			}
		}
	}

	return results
}

func (pss *PersonSearchService) FindByUsername(username, domain string) []*Person {
	var results []*Person
	username = strings.ToLower(username)
	domain = strings.ToLower(domain)

	for _, person := range pss.People {
		for _, personUsername := range person.Usernames {
			if strings.ToLower(personUsername.Username) == username &&
				strings.ToLower(personUsername.Domain) == domain &&
				personUsername.IsActive {
				results = append(results, person)
				break
			}
		}
	}

	return results
}

func (pss *PersonSearchService) FindByCompany(company string) []*Person {
	var results []*Person
	normalizedCompany := NormalizeCompanyName(company)

	for _, person := range pss.People {
		if NormalizeCompanyName(person.GetCurrentCompany()) == normalizedCompany {
			results = append(results, person)
		}
	}

	return results
}

func (pss *PersonSearchService) FindByJobTitle(jobTitle string) []*Person {
	var results []*Person
	normalizedTitle := strings.ToLower(jobTitle)

	for _, person := range pss.People {
		if strings.Contains(strings.ToLower(person.JobTitle), normalizedTitle) {
			results = append(results, person)
		}
	}

	return results
}

func (pss *PersonSearchService) GetAllPeople() []*Person {
	var results []*Person
	for _, person := range pss.People {
		results = append(results, person)
	}
	return results
}

// GetDescription methods for attachment models

func (pe *PersonEmail) GetDescription() string {
	return "Represents an email address associated with a person, including type and source information."
}

func (pp *PersonPhone) GetDescription() string {
	return "Represents a phone number associated with a person, with international format validation."
}

func (pu *PersonUsername) GetDescription() string {
	return "Represents a username or account identifier for a person on various platforms and systems."
}

func (pw *PersonWebsite) GetDescription() string {
	return "Represents a website or social media profile associated with a person."
}

// GetHooks for attachment models (empty implementations)

func (pe *PersonEmail) GetHooks() []registry.Hook    { return []registry.Hook{} }
func (pp *PersonPhone) GetHooks() []registry.Hook    { return []registry.Hook{} }
func (pu *PersonUsername) GetHooks() []registry.Hook { return []registry.Hook{} }
func (pw *PersonWebsite) GetHooks() []registry.Hook  { return []registry.Hook{} }

// Defaulted for attachment models

func (pe *PersonEmail) Defaulted() {
	if pe.DateAdded == "" {
		pe.DateAdded = Now()
	}
	if pe.IsActive == false {
		pe.IsActive = true
	}
}

func (pp *PersonPhone) Defaulted() {
	if pp.DateAdded == "" {
		pp.DateAdded = Now()
	}
	if pp.IsActive == false {
		pp.IsActive = true
	}
}

func (pu *PersonUsername) Defaulted() {
	if pu.DateAdded == "" {
		pu.DateAdded = Now()
	}
	if pu.IsActive == false {
		pu.IsActive = true
	}
}

func (pw *PersonWebsite) Defaulted() {
	if pw.DateAdded == "" {
		pw.DateAdded = Now()
	}
	if pw.IsActive == false {
		pw.IsActive = true
	}
}

// GetKey methods for attachment models (these are not typically used as standalone entities)

func (pe *PersonEmail) GetKey() string    { return "" }
func (pp *PersonPhone) GetKey() string    { return "" }
func (pu *PersonUsername) GetKey() string { return "" }
func (pw *PersonWebsite) GetKey() string  { return "" }
