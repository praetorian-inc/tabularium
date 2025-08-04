package model

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// PeopleDiscoveryService demonstrates enumeration of people from various sources
// for security assessments targeting human attack surfaces
type PeopleDiscoveryService struct {
	People    map[string]*Person      // Keyed by normalized name + company
	Companies map[string][]string     // Maps company to person names
	Sources   map[string]PersonSource // Available data sources
}

// PersonSource represents a data source for people enumeration
type PersonSource interface {
	GetName() string
	GetDescription() string
	EnumeratePeople(organization string) ([]*Person, error)
	GetPersonDetails(person *Person) (*Person, error)
	IsAvailable() bool
	GetRateLimit() time.Duration
}

// LinkedInSource simulates LinkedIn enumeration for people discovery
type LinkedInSource struct {
	Available bool
	RateLimit time.Duration
}

// ZoomInfoSource simulates ZoomInfo B2B database for people enumeration
type ZoomInfoSource struct {
	Available bool
	RateLimit time.Duration
}

// GitHubSource discovers people through GitHub organization membership
type GitHubSource struct {
	Available bool
	RateLimit time.Duration
}

// CompanyWebsiteSource discovers people through company websites and about pages
type CompanyWebsiteSource struct {
	Available bool
	RateLimit time.Duration
}

// PersonDiscoveryResult represents the result of a people enumeration operation
type PersonDiscoveryResult struct {
	Person     *Person
	Source     string
	Confidence int
	Assets     []string // Related assets discovered (email domains, etc.)
	Timestamp  string
}

// NewPeopleDiscoveryService creates a new service for people enumeration
func NewPeopleDiscoveryService() *PeopleDiscoveryService {
	service := &PeopleDiscoveryService{
		People:    make(map[string]*Person),
		Companies: make(map[string][]string),
		Sources:   make(map[string]PersonSource),
	}

	// Initialize mock sources
	service.Sources["linkedin"] = &LinkedInSource{Available: true, RateLimit: 2 * time.Second}
	service.Sources["zoominfo"] = &ZoomInfoSource{Available: true, RateLimit: 1 * time.Second}
	service.Sources["github"] = &GitHubSource{Available: true, RateLimit: 500 * time.Millisecond}
	service.Sources["website"] = &CompanyWebsiteSource{Available: true, RateLimit: 3 * time.Second}

	return service
}

// DiscoverPeople performs comprehensive people enumeration for a target organization
func (pds *PeopleDiscoveryService) DiscoverPeople(organization string, sources []string) ([]PersonDiscoveryResult, error) {
	var allResults []PersonDiscoveryResult

	if organization == "" {
		return allResults, fmt.Errorf("organization name is required")
	}

	// If no sources specified, use all available sources
	if len(sources) == 0 {
		for sourceName := range pds.Sources {
			if pds.Sources[sourceName].IsAvailable() {
				sources = append(sources, sourceName)
			}
		}
	}

	for _, sourceName := range sources {
		source, exists := pds.Sources[sourceName]
		if !exists || !source.IsAvailable() {
			continue
		}

		// Respect rate limiting
		time.Sleep(source.GetRateLimit())

		people, err := source.EnumeratePeople(organization)
		if err != nil {
			continue // Skip failed sources
		}

		for _, person := range people {
			// Enhance person with additional details
			enhancedPerson, err := source.GetPersonDetails(person)
			if err != nil {
				enhancedPerson = person // Use basic info if enhancement fails
			}

			result := PersonDiscoveryResult{
				Person:     enhancedPerson,
				Source:     sourceName,
				Confidence: pds.calculateConfidence(enhancedPerson, sourceName),
				Assets:     pds.extractAssets(enhancedPerson),
				Timestamp:  Now(),
			}

			allResults = append(allResults, result)
			pds.addPersonToCache(enhancedPerson)
		}
	}

	return allResults, nil
}

// LinkedIn Source Implementation

func (ls *LinkedInSource) GetName() string {
	return "LinkedIn"
}

func (ls *LinkedInSource) GetDescription() string {
	return "Professional networking platform for discovering employees and organizational structure"
}

func (ls *LinkedInSource) IsAvailable() bool {
	return ls.Available
}

func (ls *LinkedInSource) GetRateLimit() time.Duration {
	return ls.RateLimit
}

func (ls *LinkedInSource) EnumeratePeople(organization string) ([]*Person, error) {
	// Simulate LinkedIn company page scraping and employee discovery
	mockEmployees := generateMockLinkedInEmployees(organization)

	var people []*Person
	for _, emp := range mockEmployees {
		person := NewPerson(emp.Name)
		person.SetCurrentCompany(organization)
		person.JobTitle = emp.Title
		person.Department = emp.Department
		person.Location = emp.Location
		person.Industry = "Technology" // Default for demo
		person.LinkedInURL = fmt.Sprintf("https://linkedin.com/in/%s", strings.ToLower(strings.ReplaceAll(emp.Name, " ", "")))
		person.Bio = emp.Bio
		person.NetworkSize = emp.NetworkSize
		person.Skills = emp.Skills
		person.YearsExperience = emp.Experience
		person.SeniorityLevel = determineSeniorityFromTitle(emp.Title)

		// Add LinkedIn-specific attachments
		if emp.Email != "" {
			person.AddEmail(emp.Email, EmailTypeWork, "linkedin")
		}
		if emp.Phone != "" {
			person.AddPhone(emp.Phone, PhoneTypeWork, "US", "linkedin")
		}

		people = append(people, &person)
	}

	return people, nil
}

func (ls *LinkedInSource) GetPersonDetails(person *Person) (*Person, error) {
	// Simulate enhanced LinkedIn profile data
	enhanced := *person

	// Add more detailed LinkedIn information
	enhanced.Bio = generateLinkedInBio(person.JobTitle, person.Department)
	enhanced.NetworkSize = 500 + rand.Intn(2000)
	enhanced.Skills = generateTechSkills(person.JobTitle)
	enhanced.Languages = []string{"English"}

	// Add more contact methods found on LinkedIn
	if person.GetCurrentCompany() != "" && person.GetCurrentCompany() != "unknown" {
		domain := strings.ToLower(strings.ReplaceAll(person.GetCurrentCompany(), " ", "")) + ".com"
		firstLast := strings.Fields(person.FullName)
		if len(firstLast) >= 2 {
			email := fmt.Sprintf("%s.%s@%s", strings.ToLower(firstLast[0]), strings.ToLower(firstLast[1]), domain)
			enhanced.AddEmail(email, EmailTypeWork, "linkedin-inference")
		}
	}

	return &enhanced, nil
}

// ZoomInfo Source Implementation

func (zs *ZoomInfoSource) GetName() string {
	return "ZoomInfo"
}

func (zs *ZoomInfoSource) GetDescription() string {
	return "B2B contact database for comprehensive employee and company intelligence"
}

func (zs *ZoomInfoSource) IsAvailable() bool {
	return zs.Available
}

func (zs *ZoomInfoSource) GetRateLimit() time.Duration {
	return zs.RateLimit
}

func (zs *ZoomInfoSource) EnumeratePeople(organization string) ([]*Person, error) {
	// Simulate ZoomInfo database lookup
	mockContacts := generateMockZoomInfoContacts(organization)

	var people []*Person
	for _, contact := range mockContacts {
		person := NewPerson(contact.Name)
		person.SetCurrentCompany(organization)
		person.JobTitle = contact.Title
		person.Department = contact.Department
		person.Location = contact.Location
		person.CompanySize = contact.CompanySize
		person.Industry = contact.Industry
		person.YearsExperience = contact.Experience
		person.SeniorityLevel = determineSeniorityFromTitle(contact.Title)
		person.IsDecisionMaker = contact.HasBudgetAuthority

		// ZoomInfo typically has more complete contact information
		if contact.Email != "" {
			person.AddEmail(contact.Email, EmailTypeWork, "zoominfo")
		}
		if contact.DirectPhone != "" {
			person.AddPhone(contact.DirectPhone, PhoneTypeWork, "US", "zoominfo")
		}
		if contact.MobilePhone != "" {
			person.AddPhone(contact.MobilePhone, PhoneTypeMobile, "US", "zoominfo")
		}

		people = append(people, &person)
	}

	return people, nil
}

func (zs *ZoomInfoSource) GetPersonDetails(person *Person) (*Person, error) {
	// Simulate ZoomInfo detailed contact enhancement
	enhanced := *person

	// Add ZoomInfo-specific data
	enhanced.SecurityClearance = ""        // Usually not available in commercial databases
	enhanced.AccessLevel = AccessLevelUser // Default assumption

	// Generate additional contact methods
	if person.GetCurrentCompany() != "" && person.GetCurrentCompany() != "unknown" {
		domain := inferEmailDomain(person.GetCurrentCompany())
		firstLast := strings.Fields(person.FullName)
		if len(firstLast) >= 2 {
			// Multiple email format possibilities
			variations := []string{
				fmt.Sprintf("%s@%s", strings.ToLower(firstLast[0]), domain),
				fmt.Sprintf("%s%s@%s", strings.ToLower(firstLast[0][:1]), strings.ToLower(firstLast[1]), domain),
				fmt.Sprintf("%s_%s@%s", strings.ToLower(firstLast[0]), strings.ToLower(firstLast[1]), domain),
			}

			for i, email := range variations {
				if i == 0 {
					enhanced.AddEmail(email, EmailTypeWork, "zoominfo-primary")
				} else {
					enhanced.AddEmail(email, EmailTypeWork, "zoominfo-variation")
				}
				if i >= 2 { // Limit variations
					break
				}
			}
		}
	}

	return &enhanced, nil
}

// GitHub Source Implementation

func (gs *GitHubSource) GetName() string {
	return "GitHub"
}

func (gs *GitHubSource) GetDescription() string {
	return "Code repository platform for discovering technical team members and their skills"
}

func (gs *GitHubSource) IsAvailable() bool {
	return gs.Available
}

func (gs *GitHubSource) GetRateLimit() time.Duration {
	return gs.RateLimit
}

func (gs *GitHubSource) EnumeratePeople(organization string) ([]*Person, error) {
	// Simulate GitHub organization member discovery
	mockMembers := generateMockGitHubMembers(organization)

	var people []*Person
	for _, member := range mockMembers {
		person := NewPerson(member.Name)
		person.SetCurrentCompany(organization)
		person.JobTitle = inferTitleFromGitHub(member.Role, member.Contributions)
		person.Department = "Engineering" // Default for GitHub users
		person.Skills = member.Languages
		person.YearsExperience = member.Experience
		person.SeniorityLevel = determineSeniorityFromTitle(person.JobTitle)

		// GitHub-specific information
		if member.PublicEmail != "" {
			person.AddEmail(member.PublicEmail, EmailTypeWork, "github")
		}

		githubURL := fmt.Sprintf("https://github.com/%s", member.Username)
		person.AddWebsite(githubURL, WebsiteGitHub, "github")

		people = append(people, &person)
	}

	return people, nil
}

func (gs *GitHubSource) GetPersonDetails(person *Person) (*Person, error) {
	// Simulate GitHub profile enhancement
	enhanced := *person

	// Add GitHub-specific technical details
	enhanced.Skills = append(enhanced.Skills, "Git", "Version Control", "Open Source")

	return &enhanced, nil
}

// Company Website Source Implementation

func (cws *CompanyWebsiteSource) GetName() string {
	return "Company Website"
}

func (cws *CompanyWebsiteSource) GetDescription() string {
	return "Company websites, about pages, and team directories for discovering leadership and staff"
}

func (cws *CompanyWebsiteSource) IsAvailable() bool {
	return cws.Available
}

func (cws *CompanyWebsiteSource) GetRateLimit() time.Duration {
	return cws.RateLimit
}

func (cws *CompanyWebsiteSource) EnumeratePeople(organization string) ([]*Person, error) {
	// Simulate website scraping for team/about pages
	mockStaff := generateMockWebsiteStaff(organization)

	var people []*Person
	for _, staff := range mockStaff {
		person := NewPerson(staff.Name)
		person.SetCurrentCompany(organization)
		person.JobTitle = staff.Title
		person.Department = staff.Department
		person.Bio = staff.Bio
		person.SeniorityLevel = determineSeniorityFromTitle(staff.Title)
		person.IsDecisionMaker = strings.Contains(strings.ToLower(staff.Title), "ceo") ||
			strings.Contains(strings.ToLower(staff.Title), "cto") ||
			strings.Contains(strings.ToLower(staff.Title), "president")

		// Website contact information (often limited)
		if staff.Email != "" {
			person.AddEmail(staff.Email, EmailTypeWork, "website")
		}

		people = append(people, &person)
	}

	return people, nil
}

func (cws *CompanyWebsiteSource) GetPersonDetails(person *Person) (*Person, error) {
	// Company websites typically have limited details
	return person, nil
}

// Helper functions and mock data generators

type MockLinkedInEmployee struct {
	Name        string
	Title       string
	Department  string
	Location    string
	Bio         string
	NetworkSize int
	Skills      []string
	Experience  int
	Email       string
	Phone       string
}

type MockZoomInfoContact struct {
	Name               string
	Title              string
	Department         string
	Location           string
	CompanySize        string
	Industry           string
	Experience         int
	HasBudgetAuthority bool
	Email              string
	DirectPhone        string
	MobilePhone        string
}

type MockGitHubMember struct {
	Name          string
	Username      string
	Role          string
	Contributions int
	Languages     []string
	Experience    int
	PublicEmail   string
}

type MockWebsiteStaff struct {
	Name       string
	Title      string
	Department string
	Bio        string
	Email      string
}

func generateMockLinkedInEmployees(organization string) []MockLinkedInEmployee {
	employees := []MockLinkedInEmployee{
		{
			Name:        "Sarah Johnson",
			Title:       "Chief Technology Officer",
			Department:  "Engineering",
			Location:    "San Francisco, CA",
			Bio:         "Experienced technology leader with 15+ years in software engineering and team building.",
			NetworkSize: 1500,
			Skills:      []string{"Leadership", "Software Architecture", "Team Building", "Strategic Planning"},
			Experience:  15,
			Email:       fmt.Sprintf("sarah.johnson@%s", inferEmailDomain(organization)),
		},
		{
			Name:        "Michael Chen",
			Title:       "Senior Software Engineer",
			Department:  "Engineering",
			Location:    "Seattle, WA",
			Bio:         "Full-stack developer passionate about scalable systems and clean code.",
			NetworkSize: 750,
			Skills:      []string{"Go", "Python", "React", "Kubernetes", "AWS"},
			Experience:  8,
			Email:       fmt.Sprintf("michael.chen@%s", inferEmailDomain(organization)),
		},
		{
			Name:        "Jennifer Martinez",
			Title:       "VP of Marketing",
			Department:  "Marketing",
			Location:    "New York, NY",
			Bio:         "Digital marketing strategist focused on growth and brand development.",
			NetworkSize: 2000,
			Skills:      []string{"Digital Marketing", "Growth Strategy", "Brand Management", "Analytics"},
			Experience:  12,
			Email:       fmt.Sprintf("jennifer.martinez@%s", inferEmailDomain(organization)),
		},
		{
			Name:        "David Kim",
			Title:       "Security Engineer",
			Department:  "Security",
			Location:    "Austin, TX",
			Bio:         "Cybersecurity specialist with expertise in threat detection and incident response.",
			NetworkSize: 600,
			Skills:      []string{"Cybersecurity", "Threat Detection", "Incident Response", "Penetration Testing"},
			Experience:  6,
			Email:       fmt.Sprintf("david.kim@%s", inferEmailDomain(organization)),
		},
	}

	return employees
}

func generateMockZoomInfoContacts(organization string) []MockZoomInfoContact {
	contacts := []MockZoomInfoContact{
		{
			Name:               "Robert Thompson",
			Title:              "CEO",
			Department:         "Executive",
			Location:           "San Francisco, CA",
			CompanySize:        "201-500",
			Industry:           "Technology",
			Experience:         20,
			HasBudgetAuthority: true,
			Email:              fmt.Sprintf("robert.thompson@%s", inferEmailDomain(organization)),
			DirectPhone:        "+14155551001",
			MobilePhone:        "+14155551002",
		},
		{
			Name:               "Lisa Wang",
			Title:              "Head of Sales",
			Department:         "Sales",
			Location:           "Chicago, IL",
			CompanySize:        "201-500",
			Industry:           "Technology",
			Experience:         10,
			HasBudgetAuthority: true,
			Email:              fmt.Sprintf("lisa.wang@%s", inferEmailDomain(organization)),
			DirectPhone:        "+13125551003",
			MobilePhone:        "+13125551004",
		},
		{
			Name:               "James Wilson",
			Title:              "Software Developer",
			Department:         "Engineering",
			Location:           "Remote",
			CompanySize:        "201-500",
			Industry:           "Technology",
			Experience:         4,
			HasBudgetAuthority: false,
			Email:              fmt.Sprintf("james.wilson@%s", inferEmailDomain(organization)),
			DirectPhone:        "+15035551005",
		},
	}

	return contacts
}

func generateMockGitHubMembers(organization string) []MockGitHubMember {
	members := []MockGitHubMember{
		{
			Name:          "Alex Rodriguez",
			Username:      "alexrod",
			Role:          "Owner",
			Contributions: 1500,
			Languages:     []string{"Go", "Python", "JavaScript", "Rust"},
			Experience:    12,
			PublicEmail:   fmt.Sprintf("alex@%s", inferEmailDomain(organization)),
		},
		{
			Name:          "Emily Davis",
			Username:      "emilyd",
			Role:          "Member",
			Contributions: 800,
			Languages:     []string{"TypeScript", "React", "Node.js", "PostgreSQL"},
			Experience:    6,
			PublicEmail:   fmt.Sprintf("emily.davis@%s", inferEmailDomain(organization)),
		},
		{
			Name:          "Marcus Brown",
			Username:      "mbrown",
			Role:          "Member",
			Contributions: 400,
			Languages:     []string{"Python", "Django", "Docker", "AWS"},
			Experience:    3,
		},
	}

	return members
}

func generateMockWebsiteStaff(organization string) []MockWebsiteStaff {
	staff := []MockWebsiteStaff{
		{
			Name:       "Catherine Miller",
			Title:      "President & Co-Founder",
			Department: "Executive",
			Bio:        "Catherine founded the company in 2015 with a vision to revolutionize the industry.",
			Email:      fmt.Sprintf("catherine@%s", inferEmailDomain(organization)),
		},
		{
			Name:       "Thomas Anderson",
			Title:      "Chief Operating Officer",
			Department: "Operations",
			Bio:        "Thomas oversees daily operations and ensures efficient business processes.",
			Email:      fmt.Sprintf("thomas@%s", inferEmailDomain(organization)),
		},
		{
			Name:       "Maria Garcia",
			Title:      "Head of Customer Success",
			Department: "Customer Success",
			Bio:        "Maria leads our customer success team to ensure client satisfaction and growth.",
			Email:      fmt.Sprintf("maria@%s", inferEmailDomain(organization)),
		},
	}

	return staff
}

// Utility functions

func (pds *PeopleDiscoveryService) calculateConfidence(person *Person, source string) int {
	confidence := 50 // Base confidence

	// Increase confidence based on data completeness
	if person.GetPrimaryEmail() != "" {
		confidence += 15
	}
	if person.GetPrimaryPhone() != "" {
		confidence += 15
	}
	if person.JobTitle != "" {
		confidence += 10
	}
	if person.Department != "" {
		confidence += 5
	}
	if person.LinkedInURL != "" {
		confidence += 10
	}

	// Source-specific confidence adjustments
	switch source {
	case "zoominfo":
		confidence += 10 // ZoomInfo generally has high-quality B2B data
	case "linkedin":
		confidence += 5 // LinkedIn is reliable but may have outdated info
	case "github":
		confidence -= 5 // Technical focus, may not have complete business info
	case "website":
		confidence += 15 // Official company information
	}

	// Cap at 100
	if confidence > 100 {
		confidence = 100
	}

	return confidence
}

func (pds *PeopleDiscoveryService) extractAssets(person *Person) []string {
	var assets []string

	// Extract email domains as assets
	for _, email := range person.GetActiveEmails() {
		if strings.Contains(email, "@") {
			domain := strings.Split(email, "@")[1]
			assets = append(assets, domain)
		}
	}

	// Extract phone number patterns
	for _, phone := range person.GetActivePhones() {
		if strings.HasPrefix(phone, "+1") {
			assets = append(assets, "US-phone-numbers")
		}
	}

	return assets
}

func (pds *PeopleDiscoveryService) addPersonToCache(person *Person) {
	key := fmt.Sprintf("%s:%s", NormalizePersonName(person.FullName), NormalizeCompanyName(person.GetCurrentCompany()))
	pds.People[key] = person

	// Update company index
	company := person.GetCurrentCompany()
	if _, exists := pds.Companies[company]; !exists {
		pds.Companies[company] = make([]string, 0)
	}
	pds.Companies[company] = append(pds.Companies[company], person.FullName)
}

func inferEmailDomain(organization string) string {
	// Simple domain inference - in reality this would be more sophisticated
	normalized := strings.ToLower(strings.ReplaceAll(organization, " ", ""))
	return normalized + ".com"
}

func determineSeniorityFromTitle(title string) string {
	title = strings.ToLower(title)

	if strings.Contains(title, "intern") || strings.Contains(title, "trainee") {
		return SeniorityEntry
	}
	if strings.Contains(title, "junior") || strings.Contains(title, "associate") {
		return SeniorityEntry
	}
	if strings.Contains(title, "senior") || strings.Contains(title, "lead") {
		return SenioritSenior
	}
	if strings.Contains(title, "principal") || strings.Contains(title, "staff") {
		return SenioritSenior
	}
	if strings.Contains(title, "manager") || strings.Contains(title, "director") || strings.Contains(title, "head") {
		return SeniorityExecutive
	}
	if strings.Contains(title, "vp") || strings.Contains(title, "vice") {
		return SeniorityExecutive
	}
	if strings.Contains(title, "chief") || strings.Contains(title, "ceo") ||
		strings.Contains(title, "cto") || strings.Contains(title, "cfo") ||
		strings.Contains(title, "president") {
		return SeniorityCLevel
	}

	return SeniorityMid
}

func generateLinkedInBio(jobTitle, department string) string {
	templates := []string{
		"Experienced %s professional with expertise in %s and team leadership.",
		"Passionate %s focused on innovation and growth in the %s space.",
		"Results-driven %s with a track record of success in %s operations.",
		"Strategic %s leader specializing in %s transformation and optimization.",
	}

	template := templates[rand.Intn(len(templates))]
	return fmt.Sprintf(template, strings.ToLower(jobTitle), strings.ToLower(department))
}

func generateTechSkills(jobTitle string) []string {
	title := strings.ToLower(jobTitle)

	var skills []string

	if strings.Contains(title, "engineer") || strings.Contains(title, "developer") {
		skills = []string{"Go", "Python", "JavaScript", "AWS", "Docker", "Kubernetes"}
	} else if strings.Contains(title, "security") {
		skills = []string{"Cybersecurity", "Penetration Testing", "SIEM", "Incident Response"}
	} else if strings.Contains(title, "data") {
		skills = []string{"Python", "SQL", "Machine Learning", "Analytics", "Tableau"}
	} else if strings.Contains(title, "marketing") {
		skills = []string{"Digital Marketing", "Analytics", "SEO", "Content Strategy"}
	} else if strings.Contains(title, "sales") {
		skills = []string{"CRM", "Lead Generation", "B2B Sales", "Account Management"}
	} else {
		skills = []string{"Leadership", "Project Management", "Strategic Planning"}
	}

	return skills
}

func inferTitleFromGitHub(role string, contributions int) string {
	if role == "Owner" {
		return "Lead Engineer"
	}

	if contributions > 1000 {
		return "Senior Software Engineer"
	} else if contributions > 500 {
		return "Software Engineer"
	} else {
		return "Junior Developer"
	}
}

// DemonstratePeopleEnumeration shows comprehensive people discovery across multiple sources
func DemonstratePeopleEnumeration(organization string) {
	fmt.Printf("🔍 People Enumeration Demonstration for: %s\n", organization)
	fmt.Println("=" + strings.Repeat("=", 60))

	service := NewPeopleDiscoveryService()

	// Discover people from all sources
	results, err := service.DiscoverPeople(organization, []string{})
	if err != nil {
		fmt.Printf("Error during enumeration: %v\n", err)
		return
	}

	fmt.Printf("📊 Discovered %d people across %d sources\n\n", len(results), len(service.Sources))

	// Group results by source
	sourceResults := make(map[string][]PersonDiscoveryResult)
	for _, result := range results {
		sourceResults[result.Source] = append(sourceResults[result.Source], result)
	}

	// Display results by source
	for sourceName, sourceData := range sourceResults {
		fmt.Printf("📱 %s (%d people)\n", strings.ToUpper(sourceName), len(sourceData))
		fmt.Println(strings.Repeat("-", 40))

		for _, result := range sourceData {
			person := result.Person
			fmt.Printf("   👤 %s\n", person.GetCanonicalName())
			fmt.Printf("      Title: %s\n", person.JobTitle)
			fmt.Printf("      Department: %s\n", person.Department)
			if person.Location != "" {
				fmt.Printf("      Location: %s\n", person.Location)
			}
			fmt.Printf("      Seniority: %s\n", person.SeniorityLevel)
			fmt.Printf("      Confidence: %d%%\n", result.Confidence)

			// Display contact information
			emails := person.GetActiveEmails()
			if len(emails) > 0 {
				fmt.Printf("      Emails: %v\n", emails)
			}

			phones := person.GetActivePhones()
			if len(phones) > 0 {
				fmt.Printf("      Phones: %v\n", phones)
			}

			if len(person.Skills) > 0 {
				fmt.Printf("      Skills: %v\n", person.Skills)
			}

			if len(result.Assets) > 0 {
				fmt.Printf("      Assets: %v\n", result.Assets)
			}

			fmt.Println()
		}
		fmt.Println()
	}

	// Security assessment summary
	fmt.Println("🛡️  SECURITY ASSESSMENT SUMMARY")
	fmt.Println(strings.Repeat("=", 40))

	decisionMakers := 0
	techStaff := 0
	contactsFound := 0
	uniqueDomains := make(map[string]bool)

	for _, result := range results {
		person := result.Person

		if person.IsDecisionMaker {
			decisionMakers++
		}

		if strings.Contains(strings.ToLower(person.Department), "engineer") ||
			strings.Contains(strings.ToLower(person.Department), "security") ||
			strings.Contains(strings.ToLower(person.JobTitle), "developer") {
			techStaff++
		}

		if len(person.GetActiveEmails()) > 0 || len(person.GetActivePhones()) > 0 {
			contactsFound++
		}

		for _, email := range person.GetActiveEmails() {
			if strings.Contains(email, "@") {
				domain := strings.Split(email, "@")[1]
				uniqueDomains[domain] = true
			}
		}
	}

	fmt.Printf("📋 Total People Discovered: %d\n", len(results))
	fmt.Printf("👑 Decision Makers: %d\n", decisionMakers)
	fmt.Printf("💻 Technical Staff: %d\n", techStaff)
	fmt.Printf("📞 Contacts with Info: %d\n", contactsFound)
	fmt.Printf("🌐 Unique Email Domains: %d\n", len(uniqueDomains))

	// Attack surface insights
	fmt.Println("\n🎯 ATTACK SURFACE INSIGHTS")
	fmt.Println(strings.Repeat("-", 30))

	fmt.Printf("• %d potential phishing targets\n", contactsFound)
	fmt.Printf("• %d high-value targets (decision makers)\n", decisionMakers)
	fmt.Printf("• %d technical staff (privileged access likely)\n", techStaff)

	domains := make([]string, 0, len(uniqueDomains))
	for domain := range uniqueDomains {
		domains = append(domains, domain)
	}
	fmt.Printf("• Email domains for enumeration: %v\n", domains)

	fmt.Println("\n✅ People enumeration complete!")
}
