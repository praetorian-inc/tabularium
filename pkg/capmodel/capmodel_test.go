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
		Label:    "ADUser",
		Domain:   "example.local",
		ObjectID: "S-1-5-21-123456789",
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "Label", result.Label, "ADUser")
	assertEqual(t, "Domain", result.Domain, "example.local")
	assertEqual(t, "ObjectID", result.ObjectID, "S-1-5-21-123456789")
	assertNonEmpty(t, "Key", result.Key)
}

func TestCloudResourceConvert(t *testing.T) {
	result, err := CloudResource{
		Name:         "my-bucket",
		ResourceType: "s3",
		Region:       "us-east-1",
		IPs:          []string{"10.0.0.1"},
		URLs:         []string{"https://my-bucket.s3.amazonaws.com"},
		Properties:   map[string]any{"versioning": true},
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "Name", result.Name, "my-bucket")
	assertEqual(t, "ResourceType", string(result.ResourceType), "s3")
	assertEqual(t, "Region", result.Region, "us-east-1")
	if len(result.IPs) != 1 {
		t.Errorf("IPs: got %d, want 1", len(result.IPs))
	}
}

func TestAWSResourceConvert(t *testing.T) {
	result, err := AWSResource{
		Name:         "my-ec2",
		ResourceType: "ec2",
		Region:       "us-west-2",
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "Name", result.Name, "my-ec2")
	assertEqual(t, "ResourceType", string(result.ResourceType), "ec2")
	assertNonEmpty(t, "Key", result.Key)
}

func TestAzureResourceConvert(t *testing.T) {
	result, err := AzureResource{
		Name:         "my-vm",
		ResourceType: "vm",
		Region:       "eastus",
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "Name", result.Name, "my-vm")
	assertEqual(t, "ResourceType", string(result.ResourceType), "vm")
	assertNonEmpty(t, "Key", result.Key)
}

func TestGCPResourceConvert(t *testing.T) {
	result, err := GCPResource{
		Name:         "my-instance",
		ResourceType: "compute",
		Region:       "us-central1",
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "Name", result.Name, "my-instance")
	assertEqual(t, "ResourceType", string(result.ResourceType), "compute")
	assertNonEmpty(t, "Key", result.Key)
}

func TestPersonConvert(t *testing.T) {
	result, err := Person{
		FirstName: strPtr("Jane"),
		LastName:  strPtr("Doe"),
		Name:      strPtr("Jane Doe"),
		Email:     strPtr("jane@example.com"),
		Title:     strPtr("Engineer"),
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertPtrEqual(t, "FirstName", result.FirstName, "Jane")
	assertPtrEqual(t, "LastName", result.LastName, "Doe")
	assertPtrEqual(t, "Email", result.Email, "jane@example.com")
}

func TestOrganizationConvert(t *testing.T) {
	result, err := Organization{
		Name:    strPtr("Acme Corp"),
		Domain:  strPtr("acme.com"),
		Website: strPtr("https://acme.com"),
	}.Convert()
	if err != nil {
		t.Fatal(err)
	}
	assertPtrEqual(t, "Name", result.Name, "Acme Corp")
	assertPtrEqual(t, "Domain", result.Domain, "acme.com")
}
