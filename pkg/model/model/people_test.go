package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerson_NewPerson(t *testing.T) {
	person := NewPerson("John Smith")

	assert.Equal(t, "John Smith", person.FullName)
	assert.Equal(t, "person", person.Class)
	assert.Equal(t, "John", person.FirstName)
	assert.Equal(t, "Smith", person.LastName)
	assert.Equal(t, "", person.MiddleName)
	assert.Contains(t, person.Key, "#person#johnsmith#")
}

func TestPerson_ParseStructuredName(t *testing.T) {
	tests := []struct {
		name       string
		fullName   string
		firstName  string
		middleName string
		lastName   string
	}{
		{
			name:      "single name",
			fullName:  "John",
			firstName: "John",
			lastName:  "",
		},
		{
			name:      "first and last",
			fullName:  "John Smith",
			firstName: "John",
			lastName:  "Smith",
		},
		{
			name:       "first middle last",
			fullName:   "John Michael Smith",
			firstName:  "John",
			middleName: "Michael",
			lastName:   "Smith",
		},
		{
			name:       "multiple middle names",
			fullName:   "John Michael David Smith",
			firstName:  "John",
			middleName: "Michael David",
			lastName:   "Smith",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			person := NewPerson(tt.fullName)
			assert.Equal(t, tt.firstName, person.FirstName)
			assert.Equal(t, tt.middleName, person.MiddleName)
			assert.Equal(t, tt.lastName, person.LastName)
		})
	}
}

func TestPerson_Valid(t *testing.T) {
	tests := []struct {
		name     string
		person   Person
		expected bool
	}{
		{
			name:     "valid person",
			person:   NewPerson("John Smith"),
			expected: true,
		},
		{
			name: "missing full name",
			person: Person{
				BaseAsset: BaseAsset{Key: "#person#test#company"},
			},
			expected: false,
		},
		{
			name: "invalid key format",
			person: func() Person {
				p := NewPerson("John Smith")
				p.BaseAsset.Key = "invalid-key"
				return p
			}(),
			expected: false,
		},
		{
			name: "invalid seniority level",
			person: func() Person {
				p := NewPerson("John Smith")
				p.SeniorityLevel = "invalid"
				return p
			}(),
			expected: false,
		},
		{
			name: "invalid access level",
			person: func() Person {
				p := NewPerson("John Smith")
				p.AccessLevel = "invalid"
				return p
			}(),
			expected: false,
		},
		{
			name: "invalid email attachment",
			person: func() Person {
				p := NewPerson("John Smith")
				p.Emails = []PersonEmail{
					{Email: "invalid-email", Type: EmailTypeWork, IsActive: true, DateAdded: Now()},
				}
				return p
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.person.Valid())
		})
	}
}

func TestPerson_GetCanonicalName(t *testing.T) {
	tests := []struct {
		name     string
		person   Person
		expected string
	}{
		{
			name:     "full name with all parts",
			person:   Person{FullName: "John Michael Smith", FirstName: "John", MiddleName: "Michael", LastName: "Smith"},
			expected: "Smith, John Michael",
		},
		{
			name:     "first and last only",
			person:   Person{FullName: "John Smith", FirstName: "John", LastName: "Smith"},
			expected: "Smith, John",
		},
		{
			name:     "no structured name",
			person:   Person{FullName: "John Smith"},
			expected: "John Smith",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.person.GetCanonicalName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPersonEmail_Valid(t *testing.T) {
	tests := []struct {
		name     string
		email    PersonEmail
		expected bool
	}{
		{
			name: "valid work email",
			email: PersonEmail{
				Email:     "john.smith@company.com",
				Type:      EmailTypeWork,
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: true,
		},
		{
			name: "invalid email format",
			email: PersonEmail{
				Email:     "invalid-email",
				Type:      EmailTypeWork,
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: false,
		},
		{
			name: "invalid email type",
			email: PersonEmail{
				Email:     "john@company.com",
				Type:      "invalid",
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: false,
		},
		{
			name: "missing date added",
			email: PersonEmail{
				Email:    "john@company.com",
				Type:     EmailTypeWork,
				IsActive: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.email.Valid())
		})
	}
}

func TestPersonPhone_Valid(t *testing.T) {
	tests := []struct {
		name     string
		phone    PersonPhone
		expected bool
	}{
		{
			name: "valid US phone",
			phone: PersonPhone{
				Number:    "+14155551234",
				Type:      PhoneTypeWork,
				Country:   "US",
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: true,
		},
		{
			name: "valid UK phone",
			phone: PersonPhone{
				Number:    "+442071838750",
				Type:      PhoneTypeMobile,
				Country:   "GB",
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: true,
		},
		{
			name: "invalid phone format",
			phone: PersonPhone{
				Number:    "555-1234",
				Type:      PhoneTypeWork,
				Country:   "US",
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: false,
		},
		{
			name: "invalid phone type",
			phone: PersonPhone{
				Number:    "+14155551234",
				Type:      "invalid",
				Country:   "US",
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: false,
		},
		{
			name: "invalid country code",
			phone: PersonPhone{
				Number:    "+14155551234",
				Type:      PhoneTypeWork,
				Country:   "USA", // Should be 2-letter
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.phone.Valid())
		})
	}
}

func TestPersonUsername_Valid(t *testing.T) {
	tests := []struct {
		name     string
		username PersonUsername
		expected bool
	}{
		{
			name: "valid domain username",
			username: PersonUsername{
				Username:  "jsmith",
				Platform:  PlatformDomain,
				Domain:    "company.com",
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: true,
		},
		{
			name: "valid email username",
			username: PersonUsername{
				Username:  "john.smith",
				Platform:  PlatformEmail,
				Domain:    "company.com",
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: true,
		},
		{
			name: "invalid username format",
			username: PersonUsername{
				Username:  "john@smith", // @ not allowed in username
				Platform:  PlatformDomain,
				Domain:    "company.com",
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: false,
		},
		{
			name: "invalid platform",
			username: PersonUsername{
				Username:  "jsmith",
				Platform:  "invalid",
				Domain:    "company.com",
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.username.Valid())
		})
	}
}

func TestPersonWebsite_Valid(t *testing.T) {
	tests := []struct {
		name     string
		website  PersonWebsite
		expected bool
	}{
		{
			name: "valid LinkedIn URL",
			website: PersonWebsite{
				URL:       "https://linkedin.com/in/johnsmith",
				Type:      WebsiteLinkedIn,
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: true,
		},
		{
			name: "valid personal website",
			website: PersonWebsite{
				URL:       "https://johnsmith.dev",
				Type:      WebsitePersonal,
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: true,
		},
		{
			name: "invalid URL format",
			website: PersonWebsite{
				URL:       "not-a-url",
				Type:      WebsitePersonal,
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: false,
		},
		{
			name: "invalid website type",
			website: PersonWebsite{
				URL:       "https://example.com",
				Type:      "invalid",
				IsActive:  true,
				DateAdded: Now(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.website.Valid())
		})
	}
}

func TestPerson_AddEmail(t *testing.T) {
	person := NewPerson("John Smith")

	// Test adding valid email
	err := person.AddEmail("john.smith@company.com", EmailTypeWork, "manual")
	assert.NoError(t, err)
	assert.Len(t, person.Emails, 1)
	assert.Equal(t, "john.smith@company.com", person.Emails[0].Email)
	assert.Equal(t, EmailTypeWork, person.Emails[0].Type)
	assert.True(t, person.Emails[0].IsPrimary) // First email should be primary
	assert.Equal(t, "manual", person.Emails[0].Source)

	// Test adding duplicate email
	err = person.AddEmail("john.smith@company.com", EmailTypePersonal, "manual")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test adding second email (should not be primary)
	err = person.AddEmail("j.smith@company.com", EmailTypeWork, "linkedin")
	assert.NoError(t, err)
	assert.Len(t, person.Emails, 2)
	assert.False(t, person.Emails[1].IsPrimary)

	// Test invalid email format
	err = person.AddEmail("invalid-email", EmailTypeWork, "manual")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")

	// Test invalid email type
	err = person.AddEmail("test@company.com", "invalid", "manual")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email type")
}

func TestPerson_AddPhone(t *testing.T) {
	person := NewPerson("John Smith")

	// Test adding valid US phone
	err := person.AddPhone("+14155551234", PhoneTypeWork, "US", "manual")
	assert.NoError(t, err)
	assert.Len(t, person.Phones, 1)
	assert.Equal(t, "+14155551234", person.Phones[0].Number)
	assert.True(t, person.Phones[0].IsPrimary)

	// Test adding valid UK phone
	err = person.AddPhone("+442071838750", PhoneTypeMobile, "GB", "zoominfo")
	assert.NoError(t, err)
	assert.Len(t, person.Phones, 2)
	assert.False(t, person.Phones[1].IsPrimary)

	// Test invalid phone format
	err = person.AddPhone("555-1234", PhoneTypeWork, "US", "manual")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "phone number")

	// Test invalid country code
	err = person.AddPhone("+14155551234", PhoneTypeWork, "USA", "manual")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "country must be 2-letter")
}

func TestPerson_AddUsername(t *testing.T) {
	person := NewPerson("John Smith")

	// Test adding valid username
	err := person.AddUsername("jsmith", PlatformDomain, "company.com", "manual")
	assert.NoError(t, err)
	assert.Len(t, person.Usernames, 1)
	assert.Equal(t, "jsmith", person.Usernames[0].Username)
	assert.Equal(t, PlatformDomain, person.Usernames[0].Platform)
	assert.Equal(t, "company.com", person.Usernames[0].Domain)

	// Test adding duplicate username
	err = person.AddUsername("jsmith", PlatformDomain, "company.com", "manual")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test adding same username on different platform (should work)
	err = person.AddUsername("jsmith", PlatformIdP, "company.com", "manual")
	assert.NoError(t, err)
	assert.Len(t, person.Usernames, 2)

	// Test invalid username format
	err = person.AddUsername("john@smith", PlatformDomain, "company.com", "manual")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid username format")
}

func TestPerson_AddWebsite(t *testing.T) {
	person := NewPerson("John Smith")

	// Test adding valid website
	err := person.AddWebsite("https://linkedin.com/in/johnsmith", WebsiteLinkedIn, "linkedin")
	assert.NoError(t, err)
	assert.Len(t, person.Websites, 1)
	assert.Equal(t, "https://linkedin.com/in/johnsmith", person.Websites[0].URL)
	assert.Equal(t, WebsiteLinkedIn, person.Websites[0].Type)

	// Test adding duplicate website
	err = person.AddWebsite("https://linkedin.com/in/johnsmith", WebsiteLinkedIn, "manual")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test invalid URL
	err = person.AddWebsite("not-a-url", WebsitePersonal, "manual")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid URL format")

	// Test invalid website type
	err = person.AddWebsite("https://example.com", "invalid", "manual")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid website type")
}

func TestPerson_QueryMethods(t *testing.T) {
	person := NewPerson("John Smith")

	// Add test data
	person.AddEmail("john@company.com", EmailTypeWork, "manual")
	person.AddEmail("john.personal@gmail.com", EmailTypePersonal, "manual")
	person.AddPhone("+14155551234", PhoneTypeWork, "US", "manual")
	person.AddPhone("+14155559999", PhoneTypeMobile, "US", "manual")
	person.AddUsername("jsmith", PlatformDomain, "company.com", "manual")
	person.AddUsername("john.smith", PlatformEmail, "company.com", "manual")

	// Test GetActiveEmails
	emails := person.GetActiveEmails()
	assert.Len(t, emails, 2)
	assert.Contains(t, emails, "john@company.com")
	assert.Contains(t, emails, "john.personal@gmail.com")

	// Test GetActivePhones
	phones := person.GetActivePhones()
	assert.Len(t, phones, 2)
	assert.Contains(t, phones, "+14155551234")
	assert.Contains(t, phones, "+14155559999")

	// Test GetActiveUsernames
	usernames := person.GetActiveUsernames()
	assert.Len(t, usernames, 2)
	assert.Contains(t, usernames, "jsmith@company.com")
	assert.Contains(t, usernames, "john.smith@company.com")

	// Test GetPrimaryEmail
	primaryEmail := person.GetPrimaryEmail()
	assert.Equal(t, "john@company.com", primaryEmail) // First added should be primary

	// Test GetPrimaryPhone
	primaryPhone := person.GetPrimaryPhone()
	assert.Equal(t, "+14155551234", primaryPhone) // First added should be primary
}

func TestNormalizePersonName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"John Smith", "johnsmith"},
		{"John Michael Smith", "johnmichaelsmith"},
		{"John O'Connor", "johnoconnor"},
		{"Mary-Jane Watson", "maryjanewatson"},
		{"José García", "josegarcia"},
		{"  John   Smith  ", "johnsmith"},
		{"JOHN SMITH", "johnsmith"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizePersonName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeCompanyName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Acme Corp", "acmecorp"},
		{"Google Inc.", "googleinc"},
		{"", "unknown"},
		{"  Company Name  ", "companyname"},
		{"COMPANY-NAME", "companyname"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeCompanyName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPersonSearchService(t *testing.T) {
	service := NewPersonSearchService()

	// Create test people
	person1 := NewPerson("John Smith")
	person1.SetCurrentCompany("Acme Corp")
	person1.JobTitle = "Software Engineer"
	person1.AddEmail("john@acme.com", EmailTypeWork, "manual")
	person1.AddUsername("jsmith", PlatformDomain, "acme.com", "manual")

	person2 := NewPerson("Jane Doe")
	person2.SetCurrentCompany("Tech Inc")
	person2.JobTitle = "Senior Developer"
	person2.AddEmail("jane@tech.com", EmailTypeWork, "manual")
	person2.AddUsername("jdoe", PlatformDomain, "tech.com", "manual")

	person3 := NewPerson("Bob Johnson")
	person3.SetCurrentCompany("Acme Corp")
	person3.JobTitle = "Manager"
	person3.AddEmail("bob@acme.com", EmailTypeWork, "manual")

	// Add to service
	service.AddPerson(&person1)
	service.AddPerson(&person2)
	service.AddPerson(&person3)

	// Test FindByFullName
	found := service.FindByFullName("John Smith", "Acme Corp")
	assert.NotNil(t, found)
	assert.Equal(t, "John Smith", found.FullName)

	// Test FindByEmail
	results := service.FindByEmail("john@acme.com")
	assert.Len(t, results, 1)
	assert.Equal(t, "John Smith", results[0].FullName)

	// Test FindByUsername
	results = service.FindByUsername("jsmith", "acme.com")
	assert.Len(t, results, 1)
	assert.Equal(t, "John Smith", results[0].FullName)

	// Test FindByCompany
	results = service.FindByCompany("Acme Corp")
	assert.Len(t, results, 2)
	names := []string{results[0].FullName, results[1].FullName}
	assert.Contains(t, names, "John Smith")
	assert.Contains(t, names, "Bob Johnson")

	// Test FindByJobTitle
	results = service.FindByJobTitle("Engineer")
	assert.Len(t, results, 1)
	assert.Equal(t, "John Smith", results[0].FullName)

	// Test GetAllPeople
	all := service.GetAllPeople()
	assert.Len(t, all, 3)
}

func TestPerson_IsClass(t *testing.T) {
	person := NewPerson("John Smith")
	assert.True(t, person.IsClass("person"))
	assert.False(t, person.IsClass("asset"))
}

func TestPerson_Unmarshall(t *testing.T) {
	tests := []struct {
		name  string
		data  string
		valid bool
	}{
		{
			name:  "valid person - full name only",
			data:  `{"type": "person", "fullName": "John Smith"}`,
			valid: true,
		},
		{
			name:  "valid person - with company",
			data:  `{"type": "person", "fullName": "John Smith", "company": "Acme Corp"}`,
			valid: true,
		},
		{
			name: "valid person - with attachments",
			data: `{
				"type": "person", 
				"fullName": "John Smith",
				"emails": [{"email": "john@company.com", "type": "work", "isActive": true, "dateAdded": "2023-10-27T10:00:00Z"}]
			}`,
			valid: true,
		},
		{
			name:  "invalid person - missing full name",
			data:  `{"type": "person"}`,
			valid: false,
		},
		{
			name:  "invalid person - empty full name",
			data:  `{"type": "person", "fullName": ""}`,
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a registry.Wrapper[Assetlike]
			err := json.Unmarshal([]byte(tt.data), &a)
			require.NoError(t, err)

			registry.CallHooks(a.Model)
			assert.Equal(t, tt.valid, a.Model.Valid())
		})
	}
}

func TestPerson_JSONSerialization(t *testing.T) {
	person := NewPerson("John Michael Smith")
	person.SetCurrentCompany("Acme Corp")
	person.JobTitle = "Senior Software Engineer"
	person.Department = "Engineering"
	person.Industry = "Technology"
	person.SeniorityLevel = SenioritSenior
	person.IsDecisionMaker = true
	person.Skills = []string{"Go", "Python", "Security"}
	person.Languages = []string{"English", "Spanish"}

	person.AddEmail("john@acme.com", EmailTypeWork, "manual")
	person.AddPhone("+14155551234", PhoneTypeWork, "US", "manual")
	person.AddUsername("jsmith", PlatformDomain, "acme.com", "manual")
	person.AddWebsite("https://linkedin.com/in/johnsmith", WebsiteLinkedIn, "linkedin")

	// Marshal to JSON
	data, err := json.Marshal(person)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled Person
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify key fields
	assert.Equal(t, person.FullName, unmarshaled.FullName)
	assert.Equal(t, person.FirstName, unmarshaled.FirstName)
	assert.Equal(t, person.MiddleName, unmarshaled.MiddleName)
	assert.Equal(t, person.LastName, unmarshaled.LastName)
	// Note: After removing Company field, GetCurrentCompany() returns "unknown" unless set
	assert.Equal(t, "unknown", unmarshaled.GetCurrentCompany())
	assert.Equal(t, person.JobTitle, unmarshaled.JobTitle)
	assert.Equal(t, person.SeniorityLevel, unmarshaled.SeniorityLevel)
	assert.Equal(t, person.IsDecisionMaker, unmarshaled.IsDecisionMaker)
	assert.Equal(t, person.Skills, unmarshaled.Skills)
	assert.Equal(t, person.Languages, unmarshaled.Languages)
	assert.Len(t, unmarshaled.Emails, len(person.Emails))
	assert.Len(t, unmarshaled.Phones, len(person.Phones))
	assert.Len(t, unmarshaled.Usernames, len(person.Usernames))
	assert.Len(t, unmarshaled.Websites, len(person.Websites))
}

func TestPersonAttachment_InterfaceImplementation(t *testing.T) {
	// Test that all attachment types implement PersonAttachment interface
	var attachments []PersonAttachment

	email := PersonEmail{Email: "test@example.com", Type: EmailTypeWork, IsActive: true, Source: "manual", DateAdded: Now()}
	phone := PersonPhone{Number: "+14155551234", Type: PhoneTypeWork, Country: "US", IsActive: true, Source: "manual", DateAdded: Now()}
	username := PersonUsername{Username: "testuser", Platform: PlatformDomain, Domain: "example.com", IsActive: true, Source: "manual", DateAdded: Now()}
	website := PersonWebsite{URL: "https://example.com", Type: WebsitePersonal, IsActive: true, Source: "manual", DateAdded: Now()}

	attachments = append(attachments, &email, &phone, &username, &website)

	for i, attachment := range attachments {
		t.Run(fmt.Sprintf("attachment_%d", i), func(t *testing.T) {
			assert.NotEmpty(t, attachment.GetType())
			assert.NotEmpty(t, attachment.GetValue())
			assert.True(t, attachment.IsCurrentlyActive())
			assert.NotEmpty(t, attachment.GetSource())
			assert.True(t, attachment.Valid())
		})
	}
}

func TestPerson_PrivacyAndSecurity(t *testing.T) {
	person := NewPerson("John Smith")

	// Test that person is considered private (sensitive data)
	assert.True(t, person.IsPrivate())

	// Test security-related fields
	person.SecurityClearance = "secret"
	person.AccessLevel = AccessLevelAdmin
	person.IsDecisionMaker = true

	assert.True(t, person.Valid())
	assert.Equal(t, "secret", person.SecurityClearance)
	assert.Equal(t, AccessLevelAdmin, person.AccessLevel)
	assert.True(t, person.IsDecisionMaker)
}

// Benchmark tests for performance with thousands of people
func BenchmarkPersonSearchService_AddPerson(b *testing.B) {
	service := NewPersonSearchService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		person := NewPerson(fmt.Sprintf("Person %d", i))
		person.SetCurrentCompany(fmt.Sprintf("Company %d", i%100)) // 100 companies
		service.AddPerson(&person)
	}
}

func BenchmarkPersonSearchService_FindByEmail(b *testing.B) {
	service := NewPersonSearchService()

	// Setup test data
	for i := 0; i < 1000; i++ {
		person := NewPerson(fmt.Sprintf("Person %d", i))
		person.SetCurrentCompany(fmt.Sprintf("Company %d", i%100))
		person.AddEmail(fmt.Sprintf("person%d@company.com", i), EmailTypeWork, "setup")
		service.AddPerson(&person)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.FindByEmail(fmt.Sprintf("person%d@company.com", i%1000))
	}
}

func BenchmarkPersonSearchService_FindByCompany(b *testing.B) {
	service := NewPersonSearchService()

	// Setup test data with 1000 people across 100 companies
	for i := 0; i < 1000; i++ {
		person := NewPerson(fmt.Sprintf("Person %d", i))
		person.SetCurrentCompany(fmt.Sprintf("Company %d", i%100))
		service.AddPerson(&person)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.FindByCompany(fmt.Sprintf("Company %d", i%100))
	}
}

func TestPerson_EdgeCases(t *testing.T) {
	// Test with minimal data
	person := NewPerson("X")
	assert.True(t, person.Valid())
	assert.Equal(t, "X", person.FirstName)
	assert.Equal(t, "", person.LastName)

	// Test with very long name
	longName := strings.Repeat("VeryLongName ", 20)
	person2 := NewPerson(longName)
	assert.True(t, person2.Valid())

	// Test with Unicode names
	person3 := NewPerson("José María García-Hernández")
	assert.True(t, person3.Valid())
	assert.Contains(t, person3.Key, "josemariagarciahernandez")
}

func TestPerson_CompleteWorkflow(t *testing.T) {
	// Create a comprehensive person record
	person := NewPerson("Sarah Michelle Johnson")
	person.SetCurrentCompany("Acme Corporation")
	person.JobTitle = "Chief Technology Officer"
	person.JobDescription = "Responsible for technology strategy and engineering teams"
	person.Department = "Engineering"
	person.Location = "San Francisco, CA"
	person.Industry = "Technology"
	person.SeniorityLevel = SeniorityCLevel
	person.YearsExperience = 15
	person.CompanySize = "201-500"
	person.LinkedInURL = "https://linkedin.com/in/sarahjohnson"
	person.Bio = "Experienced technology leader with 15+ years in software engineering"
	person.NetworkSize = 1500
	person.Skills = []string{"Go", "Python", "Kubernetes", "Leadership", "Strategy"}
	person.Languages = []string{"English", "French"}
	person.SecurityClearance = "secret"
	person.AccessLevel = AccessLevelAdmin
	person.IsDecisionMaker = true

	// Add contact information
	err := person.AddEmail("sarah.johnson@acme.com", EmailTypeWork, "manual")
	assert.NoError(t, err)
	err = person.AddEmail("sarah@personal.com", EmailTypePersonal, "linkedin")
	assert.NoError(t, err)

	err = person.AddPhone("+14155551234", PhoneTypeWork, "US", "manual")
	assert.NoError(t, err)
	err = person.AddPhone("+14155559999", PhoneTypeMobile, "US", "zoominfo")
	assert.NoError(t, err)

	err = person.AddUsername("sjohnson", PlatformDomain, "acme.com", "manual")
	assert.NoError(t, err)
	err = person.AddUsername("sarah.johnson", PlatformIdP, "acme.com", "manual")
	assert.NoError(t, err)

	err = person.AddWebsite("https://linkedin.com/in/sarahjohnson", WebsiteLinkedIn, "linkedin")
	assert.NoError(t, err)
	err = person.AddWebsite("https://github.com/sjohnson", WebsiteGitHub, "github")
	assert.NoError(t, err)

	// Validate the complete record
	assert.True(t, person.Valid())
	assert.Equal(t, "person", person.GetClass())
	assert.Equal(t, "Acme Corporation", person.Group())
	assert.Equal(t, "Sarah Michelle Johnson", person.Identifier())
	assert.Equal(t, "Johnson, Sarah Michelle", person.GetCanonicalName())

	// Test search capabilities
	service := NewPersonSearchService()
	service.AddPerson(&person)

	found := service.FindByFullName("Sarah Michelle Johnson", "Acme Corporation")
	assert.NotNil(t, found)
	assert.Equal(t, person.FullName, found.FullName)

	emailResults := service.FindByEmail("sarah.johnson@acme.com")
	assert.Len(t, emailResults, 1)

	companyResults := service.FindByCompany("Acme Corporation")
	assert.Len(t, companyResults, 1)

	titleResults := service.FindByJobTitle("Technology Officer")
	assert.Len(t, titleResults, 1)

	// Test attachment queries
	assert.Len(t, person.GetActiveEmails(), 2)
	assert.Len(t, person.GetActivePhones(), 2)
	assert.Len(t, person.GetActiveUsernames(), 2)
	assert.Equal(t, "sarah.johnson@acme.com", person.GetPrimaryEmail())
	assert.Equal(t, "+14155551234", person.GetPrimaryPhone())

	t.Logf("✅ Complete person workflow successful for: %s", person.GetCanonicalName())
	t.Logf("   - Company: %s", person.GetCurrentCompany())
	t.Logf("   - Title: %s", person.JobTitle)
	t.Logf("   - Seniority: %s", person.SeniorityLevel)
	t.Logf("   - Emails: %v", person.GetActiveEmails())
	t.Logf("   - Phones: %v", person.GetActivePhones())
	t.Logf("   - Skills: %v", person.Skills)
}
