package capmodel

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// Blank import to trigger converter registration from the public package
	_ "github.com/praetorian-inc/tabularium/pkg/capmodel"
	"github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func ptr[T any](v T) *T { return &v }

func convert[T registry.Model](t *testing.T, name string, v any) T {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	result, err := registry.Registry.Convert(name, data)
	require.NoError(t, err)
	return result.(T)
}

func TestIPConvert(t *testing.T) {
	result := convert[*model.Asset](t, "IP", IP{DNS: "192.168.1.1"})
	assert.Equal(t, "192.168.1.1", result.DNS)
	// DNS and Name share the same capmodel field ("ip"), so setting DNS propagates to both.
	assert.Equal(t, "192.168.1.1", result.Name)
	assert.Contains(t, result.Key, "#asset#")
}

func TestDomainConvert(t *testing.T) {
	result := convert[*model.Asset](t, "Domain", Domain{DNS: "example.com"})
	assert.Equal(t, "example.com", result.DNS)
	assert.Equal(t, "example.com", result.Name)
	assert.Equal(t, "#asset#example.com#example.com", result.Key)
}

func TestAssetConvert(t *testing.T) {
	result := convert[*model.Asset](t, "Asset", Asset{DNS: "example.com", Name: "10.0.0.1"})
	assert.Equal(t, "example.com", result.DNS)
	assert.Equal(t, "10.0.0.1", result.Name)
	assert.Equal(t, "#asset#example.com#10.0.0.1", result.Key)
}

func TestRiskConvert(t *testing.T) {
	result := convert[*model.Risk](t, "Risk", Risk{
		DNS:    "example.com",
		Name:   "CVE-2023-12345",
		Status: "TH",
		Source: "nessus",
		Target: Asset{DNS: "example.com", Name: "10.0.0.1"},
	})
	assert.Equal(t, "example.com", result.DNS)
	assert.Equal(t, "CVE-2023-12345", result.Name)
	assert.Equal(t, "TH", result.Status)
	assert.NotNil(t, result.Target)
	assert.NotEmpty(t, result.Key)
}

func TestPortConvert(t *testing.T) {
	result := convert[*model.Port](t, "Port", Port{
		Protocol: "tcp",
		Port:     443,
		Service:  "https",
		Parent:   Asset{DNS: "example.com", Name: "10.0.0.1"},
	})
	assert.Equal(t, "tcp", result.Protocol)
	assert.Equal(t, 443, result.Port)
	assert.Equal(t, "https", result.Service)
	assert.Contains(t, result.Key, "#port#tcp#443")
}

func TestTechnologyConvert(t *testing.T) {
	result := convert[*model.Technology](t, "Technology", Technology{
		CPE:  "cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*",
		Name: "Apache httpd",
	})
	assert.Equal(t, "cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*", result.CPE)
	assert.Equal(t, "Apache httpd", result.Name)
	assert.Equal(t, "#technology#cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*", result.Key)
}

func TestFileConvert(t *testing.T) {
	result := convert[*model.File](t, "File", File{Name: "proofs/test.txt", Bytes: []byte("hello")})
	assert.Equal(t, "proofs/test.txt", result.Name)
	assert.NotEmpty(t, result.Bytes)
	assert.Equal(t, "#file#proofs/test.txt", result.Key)
}

func TestWebApplicationConvert(t *testing.T) {
	result := convert[*model.WebApplication](t, "WebApplication", WebApplication{
		PrimaryURL: "https://example.com",
		Name:       "Example App",
		URLs:       []string{"https://api.example.com"},
	})
	assert.Contains(t, result.PrimaryURL, "https://example.com")
	assert.Equal(t, "Example App", result.Name)
	assert.NotEmpty(t, result.Key)
}

func TestWebpageConvert(t *testing.T) {
	result := convert[*model.Webpage](t, "Webpage", Webpage{
		URL: "https://example.com/login",
		Parent: WebApplication{
			PrimaryURL: "https://example.com",
			Name:       "Example",
		},
	})
	assert.Equal(t, "https://example.com/login", result.URL)
	assert.NotNil(t, result.Parent)
}

func TestPreseedConvert(t *testing.T) {
	result := convert[*model.Preseed](t, "Preseed", Preseed{
		Type:  "whois",
		Title: "registrant_email",
		Value: "admin@example.com",
	})
	assert.Equal(t, "whois", result.Type)
	assert.Equal(t, "registrant_email", result.Title)
	assert.Equal(t, "admin@example.com", result.Value)
	assert.NotEmpty(t, result.Key)
}

func TestADObjectConvert(t *testing.T) {
	result := convert[*model.ADObject](t, "ADObject", ADObject{
		Label:           "ADUser",
		SecondaryLabels: []string{"ADLocalGroup"},
		Domain:          "example.local",
		ObjectID:        "S-1-5-21-123456789",
		SID:             "S-1-5-21-123456789",
		Name:            "John Smith",
		Department:      "IT",
	})
	assert.Equal(t, "ADUser", result.Label)
	assert.Equal(t, "example.local", result.Domain)
	assert.Equal(t, "S-1-5-21-123456789", result.ObjectID)
	assert.Equal(t, "S-1-5-21-123456789", result.SID)
	assert.Equal(t, []string{"ADLocalGroup"}, result.SecondaryLabels)
	assert.Equal(t, "John Smith", result.ADProperties.Name)
	assert.Equal(t, "IT", result.ADProperties.Department)
	assert.NotEmpty(t, result.Key)
}

func TestAWSResourceConvert(t *testing.T) {
	result := convert[*model.AWSResource](t, "AWSResource", AWSResource{
		Name:         "my-ec2",
		ResourceType: "ec2",
		Region:       "us-west-2",
		AccountRef:   "123456789012",
	})
	assert.Equal(t, "my-ec2", result.Name)
	assert.Equal(t, "ec2", string(result.ResourceType))
	assert.Equal(t, "123456789012", result.AccountRef)
	assert.NotEmpty(t, result.Key)
}

func TestAzureResourceConvert(t *testing.T) {
	result := convert[*model.AzureResource](t, "AzureResource", AzureResource{
		Name:          "my-vm",
		ResourceType:  "vm",
		Region:        "eastus",
		AccountRef:    "sub-123",
		ResourceGroup: "my-rg",
	})
	assert.Equal(t, "my-vm", result.Name)
	assert.Equal(t, "vm", string(result.ResourceType))
	assert.Equal(t, "sub-123", result.AccountRef)
	assert.Equal(t, "my-rg", result.ResourceGroup)
	assert.NotEmpty(t, result.Key)
}

func TestGCPResourceConvert(t *testing.T) {
	result := convert[*model.GCPResource](t, "GCPResource", GCPResource{
		Name:         "my-instance",
		ResourceType: "compute",
		Region:       "us-central1",
		AccountRef:   "my-project",
	})
	assert.Equal(t, "my-instance", result.Name)
	assert.Equal(t, "compute", string(result.ResourceType))
	assert.Equal(t, "my-project", result.AccountRef)
	assert.NotEmpty(t, result.Key)
}

func TestPersonConvert(t *testing.T) {
	result := convert[*model.Person](t, "Person", Person{
		FirstName:        ptr("Jane"),
		LastName:         ptr("Doe"),
		Name:             ptr("Jane Doe"),
		Email:            ptr("jane@example.com"),
		Title:            ptr("Engineer"),
		Headline:         ptr("Senior Engineer at Acme"),
		Phone:            ptr("+1-555-123-4567"),
		PersonalEmails:   ptr([]string{"jane@gmail.com"}),
		WorkEmail:        ptr("jane@work.com"),
		LinkedinURL:      ptr("https://linkedin.com/in/jane"),
		TwitterURL:       ptr("https://twitter.com/jane"),
		FacebookURL:      ptr("https://facebook.com/jane"),
		GithubURL:        ptr("https://github.com/jane"),
		PhotoURL:         ptr("https://example.com/photo.jpg"),
		OrganizationName: ptr("Acme Corp"),
		Country:          ptr("United States"),
		State:            ptr("California"),
		City:             ptr("San Francisco"),
		Seniority:        ptr("Senior"),
		Departments:      ptr([]string{"Engineering"}),
		Functions:        ptr([]string{"Software Development"}),
	})
	assert.Equal(t, ptr("Jane"), result.FirstName)
	assert.Equal(t, ptr("Doe"), result.LastName)
	assert.Equal(t, ptr("jane@example.com"), result.Email)
	assert.Equal(t, ptr("Senior Engineer at Acme"), result.Headline)
	assert.Equal(t, ptr("+1-555-123-4567"), result.Phone)
	assert.Equal(t, ptr("jane@work.com"), result.WorkEmail)
	assert.Equal(t, ptr("https://linkedin.com/in/jane"), result.LinkedinURL)
	assert.Equal(t, ptr("https://twitter.com/jane"), result.TwitterURL)
	assert.Equal(t, ptr("https://facebook.com/jane"), result.FacebookURL)
	assert.Equal(t, ptr("https://github.com/jane"), result.GithubURL)
	assert.Equal(t, ptr("https://example.com/photo.jpg"), result.PhotoURL)
	assert.Equal(t, ptr("Acme Corp"), result.OrganizationName)
	assert.Equal(t, ptr("United States"), result.Country)
	assert.Equal(t, ptr("California"), result.State)
	assert.Equal(t, ptr("San Francisco"), result.City)
	assert.Equal(t, ptr("Senior"), result.Seniority)
}

func TestOrganizationConvert(t *testing.T) {
	result := convert[*model.Organization](t, "Organization", Organization{
		Name:                  ptr("Acme Corp"),
		Domain:                ptr("acme.com"),
		Website:               ptr("https://acme.com"),
		Description:           ptr("A great company"),
		Industry:              ptr("Technology"),
		SubIndustries:         ptr([]string{"SaaS"}),
		Keywords:              ptr([]string{"cloud"}),
		OrganizationType:      ptr("Public"),
		BusinessModel:         ptr("B2B"),
		EstimatedNumEmployees: ptr(5000),
		EmployeeRange:         ptr("1000-5000"),
		AnnualRevenue:         ptr(50000000.0),
		RevenueRange:          ptr("$10M-$50M"),
		MarketCapitalization:  ptr(1000000000.0),
		Country:               ptr("United States"),
		State:                 ptr("California"),
		City:                  ptr("San Francisco"),
		PostalCode:            ptr("94105"),
		StreetAddress:         ptr("123 Market St"),
		Phone:                 ptr("+1-555-123-4567"),
		Fax:                   ptr("+1-555-123-4568"),
		Email:                 ptr("contact@acme.com"),
		LinkedinURL:           ptr("https://linkedin.com/company/acme"),
		TwitterURL:            ptr("https://twitter.com/acme"),
		FacebookURL:           ptr("https://facebook.com/acme"),
		BlogURL:               ptr("https://blog.acme.com"),
		FoundedYear:           ptr(2010),
		PubliclyTraded:        ptr(true),
		TickerSymbol:          ptr("ACME"),
		Exchange:              ptr("NASDAQ"),
		Technologies:          ptr([]string{"AWS", "Docker"}),
		TechCategories:        ptr([]string{"Cloud"}),
		TechVendors:           ptr([]string{"Amazon"}),
		AlternatePhones:       ptr([]string{"+1-555-999-0000"}),
		PhoneTypes:            ptr([]string{"main"}),
		FundingRounds:         ptr([]string{"Series A"}),
		FundingAmounts:        ptr([]float64{5000000}),
		Investors:             ptr([]string{"Sequoia"}),
		AdditionalAddresses:   ptr([]string{"456 Oak St"}),
		AddressTypes:          ptr([]string{"office"}),
	})
	assert.Equal(t, ptr("Acme Corp"), result.Name)
	assert.Equal(t, ptr("acme.com"), result.Domain)
	assert.Equal(t, ptr("A great company"), result.Description)
	assert.Equal(t, ptr("Technology"), result.Industry)
	assert.Equal(t, ptr("United States"), result.Country)
	assert.Equal(t, ptr("+1-555-123-4567"), result.Phone)
	assert.Equal(t, ptr("https://linkedin.com/company/acme"), result.LinkedinURL)
	assert.Equal(t, ptr("ACME"), result.TickerSymbol)
}
