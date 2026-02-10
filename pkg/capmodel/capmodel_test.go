package capmodel

import (
	"strings"
	"testing"

	_ "github.com/praetorian-inc/tabularium/pkg/model/model"
)

func assertEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", field, got, want)
	}
}

func assertPrefix(t *testing.T, field, got, prefix string) {
	t.Helper()
	if !strings.HasPrefix(got, prefix) {
		t.Errorf("%s: got %q, want prefix %q", field, got, prefix)
	}
}

func assertNonEmpty(t *testing.T, field, got string) {
	t.Helper()
	if got == "" {
		t.Errorf("%s: expected non-empty", field)
	}
}

func assertPtrEqual(t *testing.T, field string, got *string, want string) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: got nil, want %q", field, want)
	} else if *got != want {
		t.Errorf("%s: got %q, want %q", field, *got, want)
	}
}

func strPtr(s string) *string { return &s }

func intPtr(i int) *int { return &i }

func float64Ptr(f float64) *float64 { return &f }

func boolPtr(b bool) *bool { return &b }

func strSlicePtr(ss []string) *[]string { return &ss }

func float64SlicePtr(fs []float64) *[]float64 { return &fs }

func TestIPConvert(t *testing.T) {
	result, err := IP{DNS: "192.168.1.1"}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "DNS", result.DNS, "192.168.1.1")
	// DNS and Name share the same capmodel field ("ip"), so setting DNS propagates to both.
	assertEqual(t, "Name", result.Name, "192.168.1.1")
	assertPrefix(t, "Key", result.Key, "#asset#")
}

func TestDomainConvert(t *testing.T) {
	result, err := Domain{DNS: "example.com"}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "DNS", result.DNS, "example.com")
	assertEqual(t, "Name", result.Name, "example.com")
	assertEqual(t, "Key", result.Key, "#asset#example.com#example.com")
}

func TestAssetConvert(t *testing.T) {
	result, err := Asset{DNS: "example.com", Name: "10.0.0.1"}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "DNS", result.DNS, "example.com")
	assertEqual(t, "Name", result.Name, "10.0.0.1")
	assertEqual(t, "Key", result.Key, "#asset#example.com#10.0.0.1")
}

func TestRiskConvert(t *testing.T) {
	result, err := Risk{
		DNS:    "example.com",
		Name:   "CVE-2023-12345",
		Status: "TH",
		Source: "nessus",
		Target: Asset{DNS: "example.com", Name: "10.0.0.1"},
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "DNS", result.DNS, "example.com")
	assertEqual(t, "Name", result.Name, "CVE-2023-12345")
	assertEqual(t, "Status", result.Status, "TH")
	if result.Target == nil {
		t.Fatal("expected Target to be set")
	}
	assertNonEmpty(t, "Key", result.Key)
}

func TestPortConvert(t *testing.T) {
	result, err := Port{
		Protocol: "tcp",
		Port:     443,
		Service:  "https",
		Parent:   Asset{DNS: "example.com", Name: "10.0.0.1"},
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "Protocol", result.Protocol, "tcp")
	if result.Port != 443 {
		t.Errorf("Port: got %d, want 443", result.Port)
	}
	assertEqual(t, "Service", result.Service, "https")
	assertPrefix(t, "Key", result.Key, "#port#tcp#443")
}

func TestTechnologyConvert(t *testing.T) {
	result, err := Technology{
		CPE:  "cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*",
		Name: "Apache httpd",
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "CPE", result.CPE, "cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*")
	assertEqual(t, "Name", result.Name, "Apache httpd")
	assertEqual(t, "Key", result.Key, "#technology#cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*")
}

func TestFileConvert(t *testing.T) {
	result, err := File{Name: "proofs/test.txt", Bytes: []byte("hello")}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "Name", result.Name, "proofs/test.txt")
	if len(result.Bytes) == 0 {
		t.Error("expected Bytes to be non-empty")
	}
	assertEqual(t, "Key", result.Key, "#file#proofs/test.txt")
}

func TestWebApplicationConvert(t *testing.T) {
	result, err := WebApplication{
		PrimaryURL: "https://example.com",
		Name:       "Example App",
		URLs:       []string{"https://api.example.com"},
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertPrefix(t, "PrimaryURL", result.PrimaryURL, "https://example.com")
	assertEqual(t, "Name", result.Name, "Example App")
	assertNonEmpty(t, "Key", result.Key)
}

func TestWebpageConvert(t *testing.T) {
	result, err := Webpage{
		URL: "https://example.com/login",
		Parent: WebApplication{
			PrimaryURL: "https://example.com",
			Name:       "Example",
		},
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "URL", result.URL, "https://example.com/login")
	if result.Parent == nil {
		t.Fatal("expected Parent to be set")
	}
}

func TestPreseedConvert(t *testing.T) {
	result, err := Preseed{
		Type:  "whois",
		Title: "registrant_email",
		Value: "admin@example.com",
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "Type", result.Type, "whois")
	assertEqual(t, "Title", result.Title, "registrant_email")
	assertEqual(t, "Value", result.Value, "admin@example.com")
	assertNonEmpty(t, "Key", result.Key)
}

func TestADObjectConvert(t *testing.T) {
	result, err := ADObject{
		Label:           "ADUser",
		SecondaryLabels: []string{"ADLocalGroup"},
		Domain:          "example.local",
		ObjectID:        "S-1-5-21-123456789",
		SID:             "S-1-5-21-123456789",
		ADProperties:    map[string]any{"name": "John Smith", "department": "IT"},
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "Label", result.Label, "ADUser")
	assertEqual(t, "Domain", result.Domain, "example.local")
	assertEqual(t, "ObjectID", result.ObjectID, "S-1-5-21-123456789")
	assertEqual(t, "SID", result.SID, "S-1-5-21-123456789")
	if len(result.SecondaryLabels) != 1 || result.SecondaryLabels[0] != "ADLocalGroup" {
		t.Errorf("SecondaryLabels: got %v, want [ADLocalGroup]", result.SecondaryLabels)
	}
	assertEqual(t, "ADProperties.Name", result.ADProperties.Name, "John Smith")
	assertEqual(t, "ADProperties.Department", result.ADProperties.Department, "IT")
	assertNonEmpty(t, "Key", result.Key)
}

func TestAWSResourceConvert(t *testing.T) {
	result, err := AWSResource{
		Name:         "my-ec2",
		ResourceType: "ec2",
		Region:       "us-west-2",
		AccountRef:   "123456789012",
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "Name", result.Name, "my-ec2")
	assertEqual(t, "ResourceType", string(result.ResourceType), "ec2")
	assertEqual(t, "AccountRef", result.AccountRef, "123456789012")
	assertNonEmpty(t, "Key", result.Key)
}

func TestAzureResourceConvert(t *testing.T) {
	result, err := AzureResource{
		Name:          "my-vm",
		ResourceType:  "vm",
		Region:        "eastus",
		AccountRef:    "sub-123",
		ResourceGroup: "my-rg",
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "Name", result.Name, "my-vm")
	assertEqual(t, "ResourceType", string(result.ResourceType), "vm")
	assertEqual(t, "AccountRef", result.AccountRef, "sub-123")
	assertEqual(t, "ResourceGroup", result.ResourceGroup, "my-rg")
	assertNonEmpty(t, "Key", result.Key)
}

func TestGCPResourceConvert(t *testing.T) {
	result, err := GCPResource{
		Name:         "my-instance",
		ResourceType: "compute",
		Region:       "us-central1",
		AccountRef:   "my-project",
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "Name", result.Name, "my-instance")
	assertEqual(t, "ResourceType", string(result.ResourceType), "compute")
	assertEqual(t, "AccountRef", result.AccountRef, "my-project")
	assertNonEmpty(t, "Key", result.Key)
}

func TestPersonConvert(t *testing.T) {
	result, err := Person{
		FirstName:        strPtr("Jane"),
		LastName:         strPtr("Doe"),
		Name:             strPtr("Jane Doe"),
		Email:            strPtr("jane@example.com"),
		Title:            strPtr("Engineer"),
		Headline:         strPtr("Senior Engineer at Acme"),
		Phone:            strPtr("+1-555-123-4567"),
		PersonalEmails:   strSlicePtr([]string{"jane@gmail.com"}),
		WorkEmail:        strPtr("jane@work.com"),
		LinkedinURL:      strPtr("https://linkedin.com/in/jane"),
		TwitterURL:       strPtr("https://twitter.com/jane"),
		FacebookURL:      strPtr("https://facebook.com/jane"),
		GithubURL:        strPtr("https://github.com/jane"),
		PhotoURL:         strPtr("https://example.com/photo.jpg"),
		OrganizationName: strPtr("Acme Corp"),
		Country:          strPtr("United States"),
		State:            strPtr("California"),
		City:             strPtr("San Francisco"),
		Seniority:        strPtr("Senior"),
		Departments:      strSlicePtr([]string{"Engineering"}),
		Functions:        strSlicePtr([]string{"Software Development"}),
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertPtrEqual(t, "FirstName", result.FirstName, "Jane")
	assertPtrEqual(t, "LastName", result.LastName, "Doe")
	assertPtrEqual(t, "Email", result.Email, "jane@example.com")
	assertPtrEqual(t, "Headline", result.Headline, "Senior Engineer at Acme")
	assertPtrEqual(t, "Phone", result.Phone, "+1-555-123-4567")
	assertPtrEqual(t, "WorkEmail", result.WorkEmail, "jane@work.com")
	assertPtrEqual(t, "LinkedinURL", result.LinkedinURL, "https://linkedin.com/in/jane")
	assertPtrEqual(t, "TwitterURL", result.TwitterURL, "https://twitter.com/jane")
	assertPtrEqual(t, "FacebookURL", result.FacebookURL, "https://facebook.com/jane")
	assertPtrEqual(t, "GithubURL", result.GithubURL, "https://github.com/jane")
	assertPtrEqual(t, "PhotoURL", result.PhotoURL, "https://example.com/photo.jpg")
	assertPtrEqual(t, "OrganizationName", result.OrganizationName, "Acme Corp")
	assertPtrEqual(t, "Country", result.Country, "United States")
	assertPtrEqual(t, "State", result.State, "California")
	assertPtrEqual(t, "City", result.City, "San Francisco")
	assertPtrEqual(t, "Seniority", result.Seniority, "Senior")
}

func TestOrganizationConvert(t *testing.T) {
	result, err := Organization{
		Name:                  strPtr("Acme Corp"),
		Domain:                strPtr("acme.com"),
		Website:               strPtr("https://acme.com"),
		Description:           strPtr("A great company"),
		Industry:              strPtr("Technology"),
		SubIndustries:         strSlicePtr([]string{"SaaS"}),
		Keywords:              strSlicePtr([]string{"cloud"}),
		OrganizationType:      strPtr("Public"),
		BusinessModel:         strPtr("B2B"),
		EstimatedNumEmployees: intPtr(5000),
		EmployeeRange:         strPtr("1000-5000"),
		AnnualRevenue:         float64Ptr(50000000),
		RevenueRange:          strPtr("$10M-$50M"),
		MarketCapitalization:  float64Ptr(1000000000),
		Country:               strPtr("United States"),
		State:                 strPtr("California"),
		City:                  strPtr("San Francisco"),
		PostalCode:            strPtr("94105"),
		StreetAddress:         strPtr("123 Market St"),
		Phone:                 strPtr("+1-555-123-4567"),
		Fax:                   strPtr("+1-555-123-4568"),
		Email:                 strPtr("contact@acme.com"),
		LinkedinURL:           strPtr("https://linkedin.com/company/acme"),
		TwitterURL:            strPtr("https://twitter.com/acme"),
		FacebookURL:           strPtr("https://facebook.com/acme"),
		BlogURL:               strPtr("https://blog.acme.com"),
		FoundedYear:           intPtr(2010),
		PubliclyTraded:        boolPtr(true),
		TickerSymbol:          strPtr("ACME"),
		Exchange:              strPtr("NASDAQ"),
		Technologies:          strSlicePtr([]string{"AWS", "Docker"}),
		TechCategories:        strSlicePtr([]string{"Cloud"}),
		TechVendors:           strSlicePtr([]string{"Amazon"}),
		AlternatePhones:       strSlicePtr([]string{"+1-555-999-0000"}),
		PhoneTypes:            strSlicePtr([]string{"main"}),
		FundingRounds:         strSlicePtr([]string{"Series A"}),
		FundingAmounts:        float64SlicePtr([]float64{5000000}),
		Investors:             strSlicePtr([]string{"Sequoia"}),
		AdditionalAddresses:   strSlicePtr([]string{"456 Oak St"}),
		AddressTypes:          strSlicePtr([]string{"office"}),
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertPtrEqual(t, "Name", result.Name, "Acme Corp")
	assertPtrEqual(t, "Domain", result.Domain, "acme.com")
	assertPtrEqual(t, "Description", result.Description, "A great company")
	assertPtrEqual(t, "Industry", result.Industry, "Technology")
	assertPtrEqual(t, "Country", result.Country, "United States")
	assertPtrEqual(t, "Phone", result.Phone, "+1-555-123-4567")
	assertPtrEqual(t, "LinkedinURL", result.LinkedinURL, "https://linkedin.com/company/acme")
	assertPtrEqual(t, "TickerSymbol", result.TickerSymbol, "ACME")
}
