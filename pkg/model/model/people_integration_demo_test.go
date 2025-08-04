package model

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeopleDiscoveryService_NewService(t *testing.T) {
	service := NewPeopleDiscoveryService()

	assert.NotNil(t, service)
	assert.NotNil(t, service.People)
	assert.NotNil(t, service.Companies)
	assert.NotNil(t, service.Sources)

	// Check that all expected sources are initialized
	expectedSources := []string{"linkedin", "zoominfo", "github", "website"}
	for _, sourceName := range expectedSources {
		source, exists := service.Sources[sourceName]
		assert.True(t, exists, "Source %s should exist", sourceName)
		assert.True(t, source.IsAvailable(), "Source %s should be available", sourceName)
		assert.NotEmpty(t, source.GetName())
		assert.NotEmpty(t, source.GetDescription())
	}
}

func TestPeopleDiscoveryService_DiscoverPeople(t *testing.T) {
	service := NewPeopleDiscoveryService()

	results, err := service.DiscoverPeople("Acme Corporation", []string{"linkedin", "zoominfo"})
	require.NoError(t, err)

	assert.NotEmpty(t, results, "Should discover some people")

	// Verify we have results from both sources
	sources := make(map[string]bool)
	for _, result := range results {
		sources[result.Source] = true

		// Validate result structure
		assert.NotNil(t, result.Person)
		assert.NotEmpty(t, result.Source)
		assert.GreaterOrEqual(t, result.Confidence, 0)
		assert.LessOrEqual(t, result.Confidence, 100)
		assert.NotEmpty(t, result.Timestamp)

		// Validate person data
		person := result.Person
		assert.NotEmpty(t, person.FullName)
		assert.Equal(t, "Acme Corporation", person.GetCurrentCompany())
		assert.NotEmpty(t, person.JobTitle)
		assert.NotEmpty(t, person.SeniorityLevel)
		assert.True(t, person.Valid())
	}

	assert.True(t, sources["linkedin"], "Should have LinkedIn results")
	assert.True(t, sources["zoominfo"], "Should have ZoomInfo results")
}

func TestPeopleDiscoveryService_DiscoverPeople_AllSources(t *testing.T) {
	service := NewPeopleDiscoveryService()

	// Test with empty sources array (should use all available sources)
	results, err := service.DiscoverPeople("Tech Startup Inc", []string{})
	require.NoError(t, err)

	assert.NotEmpty(t, results)

	// Should have results from all 4 sources
	sources := make(map[string]int)
	for _, result := range results {
		sources[result.Source]++
	}

	assert.Contains(t, sources, "linkedin")
	assert.Contains(t, sources, "zoominfo")
	assert.Contains(t, sources, "github")
	assert.Contains(t, sources, "website")
}

func TestPeopleDiscoveryService_ErrorHandling(t *testing.T) {
	service := NewPeopleDiscoveryService()

	// Test with empty organization
	results, err := service.DiscoverPeople("", []string{"linkedin"})
	assert.Error(t, err)
	assert.Empty(t, results)
	assert.Contains(t, err.Error(), "organization name is required")
}

func TestLinkedInSource(t *testing.T) {
	source := &LinkedInSource{Available: true, RateLimit: 0} // No rate limit for testing

	assert.Equal(t, "LinkedIn", source.GetName())
	assert.NotEmpty(t, source.GetDescription())
	assert.True(t, source.IsAvailable())

	// Test enumeration
	people, err := source.EnumeratePeople("Test Company")
	require.NoError(t, err)
	assert.NotEmpty(t, people)

	for _, person := range people {
		assert.NotEmpty(t, person.FullName)
		assert.Equal(t, "Test Company", person.GetCurrentCompany())
		assert.NotEmpty(t, person.JobTitle)
		assert.NotEmpty(t, person.LinkedInURL)
		assert.Contains(t, person.LinkedInURL, "linkedin.com/in/")
		assert.GreaterOrEqual(t, person.NetworkSize, 0)
		assert.NotEmpty(t, person.Skills)

		// Should have structured name parsing
		assert.NotEmpty(t, person.FirstName)
		assert.True(t, person.Valid())
	}

	// Test person enhancement
	basePerson := people[0]
	enhanced, err := source.GetPersonDetails(basePerson)
	require.NoError(t, err)

	assert.NotEmpty(t, enhanced.Bio)
	assert.GreaterOrEqual(t, enhanced.NetworkSize, 500)
	assert.NotEmpty(t, enhanced.Skills)
	assert.Contains(t, enhanced.Languages, "English")
}

func TestZoomInfoSource(t *testing.T) {
	source := &ZoomInfoSource{Available: true, RateLimit: 0}

	assert.Equal(t, "ZoomInfo", source.GetName())
	assert.NotEmpty(t, source.GetDescription())
	assert.True(t, source.IsAvailable())

	// Test enumeration
	people, err := source.EnumeratePeople("Enterprise Corp")
	require.NoError(t, err)
	assert.NotEmpty(t, people)

	for _, person := range people {
		assert.NotEmpty(t, person.FullName)
		assert.Equal(t, "Enterprise Corp", person.GetCurrentCompany())
		assert.NotEmpty(t, person.JobTitle)
		assert.NotEmpty(t, person.CompanySize)
		assert.NotEmpty(t, person.Industry)

		// ZoomInfo typically has more complete contact info
		emails := person.GetActiveEmails()
		phones := person.GetActivePhones()
		assert.GreaterOrEqual(t, len(emails)+len(phones), 1, "Should have at least one contact method")

		assert.True(t, person.Valid())
	}

	// Test enhancement with multiple email variations
	basePerson := people[0]
	enhanced, err := source.GetPersonDetails(basePerson)
	require.NoError(t, err)

	emails := enhanced.GetActiveEmails()
	assert.GreaterOrEqual(t, len(emails), 1, "Should have at least one email after enhancement")

	// Check for email format variations
	hasVariations := false
	for _, email := range emails {
		if strings.Contains(email, "_") || len(strings.Split(email, "@")[0]) <= 3 {
			hasVariations = true
			break
		}
	}
	assert.True(t, hasVariations, "Should generate email format variations")
}

func TestGitHubSource(t *testing.T) {
	source := &GitHubSource{Available: true, RateLimit: 0}

	assert.Equal(t, "GitHub", source.GetName())
	assert.NotEmpty(t, source.GetDescription())
	assert.True(t, source.IsAvailable())

	// Test enumeration
	people, err := source.EnumeratePeople("Open Source Co")
	require.NoError(t, err)
	assert.NotEmpty(t, people)

	for _, person := range people {
		assert.NotEmpty(t, person.FullName)
		assert.Equal(t, "Open Source Co", person.GetCurrentCompany())
		assert.Equal(t, "Engineering", person.Department)
		assert.NotEmpty(t, person.Skills)
		assert.Contains(t, person.Skills, "Git") // Should be added during enhancement

		// Should have GitHub website
		websites := []string{}
		for _, website := range person.Websites {
			websites = append(websites, website.URL)
		}
		hasGitHub := false
		for _, url := range websites {
			if strings.Contains(url, "github.com") {
				hasGitHub = true
				break
			}
		}
		assert.True(t, hasGitHub, "Should have GitHub URL")

		assert.True(t, person.Valid())
	}

	// Test enhancement
	basePerson := people[0]
	enhanced, err := source.GetPersonDetails(basePerson)
	require.NoError(t, err)

	assert.Contains(t, enhanced.Skills, "Git")
	assert.Contains(t, enhanced.Skills, "Version Control")
	assert.Contains(t, enhanced.Skills, "Open Source")
}

func TestCompanyWebsiteSource(t *testing.T) {
	source := &CompanyWebsiteSource{Available: true, RateLimit: 0}

	assert.Equal(t, "Company Website", source.GetName())
	assert.NotEmpty(t, source.GetDescription())
	assert.True(t, source.IsAvailable())

	// Test enumeration
	people, err := source.EnumeratePeople("Professional Services LLC")
	require.NoError(t, err)
	assert.NotEmpty(t, people)

	for _, person := range people {
		assert.NotEmpty(t, person.FullName)
		assert.Equal(t, "Professional Services LLC", person.GetCurrentCompany())
		assert.NotEmpty(t, person.JobTitle)
		assert.NotEmpty(t, person.Bio)

		// Check for decision maker identification
		title := strings.ToLower(person.JobTitle)
		expectedDecisionMaker := strings.Contains(title, "ceo") ||
			strings.Contains(title, "cto") ||
			strings.Contains(title, "president")
		assert.Equal(t, expectedDecisionMaker, person.IsDecisionMaker)

		assert.True(t, person.Valid())
	}
}

func TestDetermineSeniorityFromTitle(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Software Engineering Intern", SeniorityEntry},
		{"Junior Developer", SeniorityEntry},
		{"Associate Product Manager", SeniorityEntry},
		{"Software Engineer", SeniorityMid},
		{"Product Manager", SeniorityMid},
		{"Senior Software Engineer", SenioritSenior},
		{"Lead Developer", SenioritSenior},
		{"Principal Engineer", SenioritSenior},
		{"Staff Engineer", SenioritSenior},
		{"Engineering Manager", SeniorityExecutive},
		{"Director of Engineering", SeniorityExecutive},
		{"VP of Engineering", SeniorityExecutive},
		{"Head of Product", SeniorityExecutive},
		{"Chief Technology Officer", SeniorityCLevel},
		{"CEO", SeniorityCLevel},
		{"President", SeniorityCLevel},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := determineSeniorityFromTitle(tt.title)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInferEmailDomain(t *testing.T) {
	tests := []struct {
		organization string
		expected     string
	}{
		{"Acme Corporation", "acmecorporation.com"},
		{"Tech Startup Inc", "techstartupinc.com"},
		{"Professional Services LLC", "professionalservicesllc.com"},
		{"Simple Name", "simplename.com"},
	}

	for _, tt := range tests {
		t.Run(tt.organization, func(t *testing.T) {
			result := inferEmailDomain(tt.organization)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateTechSkills(t *testing.T) {
	tests := []struct {
		jobTitle       string
		expectedSkills []string
	}{
		{
			jobTitle:       "Software Engineer",
			expectedSkills: []string{"Go", "Python", "JavaScript", "AWS", "Docker", "Kubernetes"},
		},
		{
			jobTitle:       "Security Analyst",
			expectedSkills: []string{"Cybersecurity", "Penetration Testing", "SIEM", "Incident Response"},
		},
		{
			jobTitle:       "Data Scientist",
			expectedSkills: []string{"Python", "SQL", "Machine Learning", "Analytics", "Tableau"},
		},
		{
			jobTitle:       "Marketing Manager",
			expectedSkills: []string{"Digital Marketing", "Analytics", "SEO", "Content Strategy"},
		},
		{
			jobTitle:       "Sales Director",
			expectedSkills: []string{"CRM", "Lead Generation", "B2B Sales", "Account Management"},
		},
		{
			jobTitle:       "Operations Manager",
			expectedSkills: []string{"Leadership", "Project Management", "Strategic Planning"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.jobTitle, func(t *testing.T) {
			skills := generateTechSkills(tt.jobTitle)
			assert.Equal(t, tt.expectedSkills, skills)
		})
	}
}

func TestPeopleDiscoveryService_ConfidenceCalculation(t *testing.T) {
	service := NewPeopleDiscoveryService()

	// Test person with minimal info
	minimalPerson := NewPerson("John Smith")
	minimalPerson.SetCurrentCompany("Test Company")
	confidence := service.calculateConfidence(&minimalPerson, "linkedin")
	assert.GreaterOrEqual(t, confidence, 50) // Base confidence

	// Test person with complete info
	completePerson := NewPerson("Jane Doe")
	completePerson.SetCurrentCompany("Test Company")
	completePerson.JobTitle = "Software Engineer"
	completePerson.Department = "Engineering"
	completePerson.LinkedInURL = "https://linkedin.com/in/janedoe"
	completePerson.AddEmail("jane@test.com", EmailTypeWork, "test")
	completePerson.AddPhone("+14155551234", PhoneTypeWork, "US", "test")

	completeConfidence := service.calculateConfidence(&completePerson, "website")
	assert.Greater(t, completeConfidence, confidence, "Complete person should have higher confidence")
	assert.LessOrEqual(t, completeConfidence, 100, "Confidence should not exceed 100")
}

func TestPeopleDiscoveryService_ExtractAssets(t *testing.T) {
	service := NewPeopleDiscoveryService()

	person := NewPerson("Test Person")
	person.AddEmail("test@company.com", EmailTypeWork, "test")
	person.AddEmail("test@subsidiary.com", EmailTypeWork, "test")
	person.AddPhone("+14155551234", PhoneTypeWork, "US", "test")
	person.AddPhone("+442071234567", PhoneTypeWork, "GB", "test")

	assets := service.extractAssets(&person)

	assert.Contains(t, assets, "company.com")
	assert.Contains(t, assets, "subsidiary.com")
	assert.Contains(t, assets, "US-phone-numbers")
}

func TestPeopleDiscoveryService_AddPersonToCache(t *testing.T) {
	service := NewPeopleDiscoveryService()

	person := NewPerson("Alice Johnson")
	person.SetCurrentCompany("Tech Corp")

	service.addPersonToCache(&person)

	// Check person was added to cache
	key := fmt.Sprintf("%s:%s", NormalizePersonName(person.FullName), NormalizeCompanyName(person.GetCurrentCompany()))
	cachedPerson, exists := service.People[key]
	assert.True(t, exists)
	assert.Equal(t, person.FullName, cachedPerson.FullName)

	// Check company index was updated
	companyPeople, exists := service.Companies[person.GetCurrentCompany()]
	assert.True(t, exists)
	assert.Contains(t, companyPeople, person.FullName)
}

func TestInferTitleFromGitHub(t *testing.T) {
	tests := []struct {
		role          string
		contributions int
		expected      string
	}{
		{"Owner", 500, "Lead Engineer"},
		{"Member", 1500, "Senior Software Engineer"},
		{"Member", 800, "Software Engineer"},
		{"Member", 300, "Junior Developer"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%d", tt.role, tt.contributions), func(t *testing.T) {
			result := inferTitleFromGitHub(tt.role, tt.contributions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPeopleDiscoveryService_Integration(t *testing.T) {
	service := NewPeopleDiscoveryService()

	// Test comprehensive discovery
	results, err := service.DiscoverPeople("Cybersecurity Firm", []string{})
	require.NoError(t, err)

	// Should have people from multiple sources
	assert.GreaterOrEqual(t, len(results), 4, "Should have people from all sources")

	// Verify data quality
	highConfidenceCount := 0
	decisionMakerCount := 0
	techStaffCount := 0

	for _, result := range results {
		if result.Confidence >= 80 {
			highConfidenceCount++
		}

		person := result.Person
		if person.IsDecisionMaker {
			decisionMakerCount++
		}

		if strings.Contains(strings.ToLower(person.JobTitle), "engineer") ||
			strings.Contains(strings.ToLower(person.JobTitle), "developer") ||
			strings.Contains(strings.ToLower(person.Department), "engineering") {
			techStaffCount++
		}
	}

	assert.GreaterOrEqual(t, highConfidenceCount, 1, "Should have at least one high-confidence result")
	assert.GreaterOrEqual(t, decisionMakerCount, 1, "Should identify at least one decision maker")
	assert.GreaterOrEqual(t, techStaffCount, 1, "Should identify at least one technical staff member")

	// Verify security assessment data
	uniqueDomains := make(map[string]bool)
	contactsWithInfo := 0

	for _, result := range results {
		person := result.Person

		if len(person.GetActiveEmails()) > 0 || len(person.GetActivePhones()) > 0 {
			contactsWithInfo++
		}

		for _, email := range person.GetActiveEmails() {
			if strings.Contains(email, "@") {
				domain := strings.Split(email, "@")[1]
				uniqueDomains[domain] = true
			}
		}
	}

	assert.GreaterOrEqual(t, contactsWithInfo, len(results)/2, "Most people should have contact info")
	assert.GreaterOrEqual(t, len(uniqueDomains), 1, "Should discover at least one email domain")
}

func TestDemonstratePeopleEnumeration_Integration(t *testing.T) {
	// This test captures the output to verify the demonstration function works
	// In a real scenario, this would be run manually to see the full output

	t.Run("people enumeration demo", func(t *testing.T) {
		// Just verify it doesn't panic
		assert.NotPanics(t, func() {
			DemonstratePeopleEnumeration("Security Consulting LLC")
		})
	})
}

func TestPeopleDiscoveryService_EdgeCases(t *testing.T) {
	service := NewPeopleDiscoveryService()

	// Test with non-existent source
	results, err := service.DiscoverPeople("Test Corp", []string{"nonexistent"})
	require.NoError(t, err)
	assert.Empty(t, results, "Should return empty results for non-existent source")

	// Test with unavailable source
	service.Sources["test"] = &LinkedInSource{Available: false, RateLimit: 0}
	results, err = service.DiscoverPeople("Test Corp", []string{"test"})
	require.NoError(t, err)
	assert.Empty(t, results, "Should return empty results for unavailable source")
}

func TestPersonSource_Interface(t *testing.T) {
	sources := []PersonSource{
		&LinkedInSource{Available: true, RateLimit: 0},
		&ZoomInfoSource{Available: true, RateLimit: 0},
		&GitHubSource{Available: true, RateLimit: 0},
		&CompanyWebsiteSource{Available: true, RateLimit: 0},
	}

	for i, source := range sources {
		t.Run(fmt.Sprintf("source_%d", i), func(t *testing.T) {
			assert.NotEmpty(t, source.GetName())
			assert.NotEmpty(t, source.GetDescription())
			assert.True(t, source.IsAvailable())
			assert.GreaterOrEqual(t, source.GetRateLimit(), time.Duration(0))

			// Test enumeration
			people, err := source.EnumeratePeople("Test Organization")
			assert.NoError(t, err)
			assert.NotEmpty(t, people)

			// Test details enhancement
			if len(people) > 0 {
				enhanced, err := source.GetPersonDetails(people[0])
				assert.NoError(t, err)
				assert.NotNil(t, enhanced)
			}
		})
	}
}

func TestMockDataGeneration(t *testing.T) {
	t.Run("LinkedIn employees", func(t *testing.T) {
		employees := generateMockLinkedInEmployees("Test Company")
		assert.NotEmpty(t, employees)

		for _, emp := range employees {
			assert.NotEmpty(t, emp.Name)
			assert.NotEmpty(t, emp.Title)
			assert.NotEmpty(t, emp.Department)
			assert.NotEmpty(t, emp.Skills)
			assert.GreaterOrEqual(t, emp.Experience, 0)
		}
	})

	t.Run("ZoomInfo contacts", func(t *testing.T) {
		contacts := generateMockZoomInfoContacts("Test Company")
		assert.NotEmpty(t, contacts)

		for _, contact := range contacts {
			assert.NotEmpty(t, contact.Name)
			assert.NotEmpty(t, contact.Title)
			assert.NotEmpty(t, contact.Department)
			assert.NotEmpty(t, contact.Industry)
			assert.GreaterOrEqual(t, contact.Experience, 0)
		}
	})

	t.Run("GitHub members", func(t *testing.T) {
		members := generateMockGitHubMembers("Test Company")
		assert.NotEmpty(t, members)

		for _, member := range members {
			assert.NotEmpty(t, member.Name)
			assert.NotEmpty(t, member.Username)
			assert.NotEmpty(t, member.Role)
			assert.NotEmpty(t, member.Languages)
			assert.GreaterOrEqual(t, member.Contributions, 0)
		}
	})

	t.Run("Website staff", func(t *testing.T) {
		staff := generateMockWebsiteStaff("Test Company")
		assert.NotEmpty(t, staff)

		for _, person := range staff {
			assert.NotEmpty(t, person.Name)
			assert.NotEmpty(t, person.Title)
			assert.NotEmpty(t, person.Department)
			assert.NotEmpty(t, person.Bio)
		}
	})
}

// Benchmark tests for performance with thousands of people
func BenchmarkPeopleDiscoveryService_DiscoverPeople(b *testing.B) {
	service := NewPeopleDiscoveryService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := service.DiscoverPeople("Benchmark Corp", []string{"linkedin"})
		if err != nil {
			b.Fatal(err)
		}
		if len(results) == 0 {
			b.Fatal("Expected results but got none")
		}
	}
}

func BenchmarkPeopleDiscoveryService_CalculateConfidence(b *testing.B) {
	service := NewPeopleDiscoveryService()
	person := NewPerson("Benchmark Person")
	person.SetCurrentCompany("Benchmark Corp")
	person.JobTitle = "Software Engineer"
	person.AddEmail("benchmark@corp.com", EmailTypeWork, "test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		confidence := service.calculateConfidence(&person, "linkedin")
		if confidence < 0 || confidence > 100 {
			b.Fatal("Invalid confidence value")
		}
	}
}
