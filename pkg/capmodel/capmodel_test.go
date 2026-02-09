package capmodel

import (
	"strings"
	"testing"

	_ "github.com/praetorian-inc/tabularium/pkg/model/model"
)

func TestIPConvert(t *testing.T) {
	ip := IP{DNS: "192.168.1.1"}
	result, err := ip.Convert()
	if err != nil {
		t.Fatalf("IP.Convert() error: %v", err)
	}

	if result.DNS != "192.168.1.1" {
		t.Errorf("expected DNS=192.168.1.1, got %q", result.DNS)
	}
	if result.Name != "192.168.1.1" {
		t.Errorf("expected Name=192.168.1.1, got %q", result.Name)
	}
	if result.Key == "" {
		t.Error("expected Key to be set by hooks")
	}
	if !strings.HasPrefix(result.Key, "#asset#") {
		t.Errorf("expected Key to start with #asset#, got %q", result.Key)
	}
}

func TestDomainConvert(t *testing.T) {
	d := Domain{DNS: "example.com"}
	result, err := d.Convert()
	if err != nil {
		t.Fatalf("Domain.Convert() error: %v", err)
	}

	if result.DNS != "example.com" {
		t.Errorf("expected DNS=example.com, got %q", result.DNS)
	}
	if result.Name != "example.com" {
		t.Errorf("expected Name=example.com, got %q", result.Name)
	}
	if result.Key != "#asset#example.com#example.com" {
		t.Errorf("expected Key=#asset#example.com#example.com, got %q", result.Key)
	}
}

func TestAssetConvert(t *testing.T) {
	a := Asset{DNS: "example.com", Name: "10.0.0.1"}
	result, err := a.Convert()
	if err != nil {
		t.Fatalf("Asset.Convert() error: %v", err)
	}

	if result.DNS != "example.com" {
		t.Errorf("expected DNS=example.com, got %q", result.DNS)
	}
	if result.Name != "10.0.0.1" {
		t.Errorf("expected Name=10.0.0.1, got %q", result.Name)
	}
	if result.Key != "#asset#example.com#10.0.0.1" {
		t.Errorf("expected Key=#asset#example.com#10.0.0.1, got %q", result.Key)
	}
}

func TestRiskConvert(t *testing.T) {
	r := Risk{
		DNS:    "example.com",
		Name:   "CVE-2023-12345",
		Status: "TH",
		Source: "nessus",
		Target: Asset{DNS: "example.com", Name: "10.0.0.1"},
	}
	result, err := r.Convert()
	if err != nil {
		t.Fatalf("Risk.Convert() error: %v", err)
	}

	if result.DNS != "example.com" {
		t.Errorf("expected DNS=example.com, got %q", result.DNS)
	}
	if result.Name != "CVE-2023-12345" {
		t.Errorf("expected Name=CVE-2023-12345, got %q", result.Name)
	}
	if result.Status != "TH" {
		t.Errorf("expected Status=TH, got %q", result.Status)
	}
	if result.Target == nil {
		t.Fatal("expected Target to be set")
	}
	if result.Key == "" {
		t.Error("expected Key to be set by hooks")
	}
}

func TestPortConvert(t *testing.T) {
	p := Port{
		Protocol: "tcp",
		Port:     443,
		Service:  "https",
		Parent:   Asset{DNS: "example.com", Name: "10.0.0.1"},
	}
	result, err := p.Convert()
	if err != nil {
		t.Fatalf("Port.Convert() error: %v", err)
	}

	if result.Protocol != "tcp" {
		t.Errorf("expected Protocol=tcp, got %q", result.Protocol)
	}
	if result.Port != 443 {
		t.Errorf("expected Port=443, got %d", result.Port)
	}
	if result.Service != "https" {
		t.Errorf("expected Service=https, got %q", result.Service)
	}
	if result.Key == "" {
		t.Error("expected Key to be set by hooks")
	}
	if !strings.HasPrefix(result.Key, "#port#tcp#443") {
		t.Errorf("expected Key to start with #port#tcp#443, got %q", result.Key)
	}
}

func TestTechnologyConvert(t *testing.T) {
	tech := Technology{
		CPE:  "cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*",
		Name: "Apache httpd",
	}
	result, err := tech.Convert()
	if err != nil {
		t.Fatalf("Technology.Convert() error: %v", err)
	}

	if result.CPE != "cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*" {
		t.Errorf("expected CPE to match, got %q", result.CPE)
	}
	if result.Name != "Apache httpd" {
		t.Errorf("expected Name=Apache httpd, got %q", result.Name)
	}
	if result.Key != "#technology#cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*" {
		t.Errorf("unexpected Key: %q", result.Key)
	}
}

func TestFileConvert(t *testing.T) {
	f := File{Name: "proofs/test.txt", Bytes: []byte("hello")}
	result, err := f.Convert()
	if err != nil {
		t.Fatalf("File.Convert() error: %v", err)
	}

	if result.Name != "proofs/test.txt" {
		t.Errorf("expected Name=proofs/test.txt, got %q", result.Name)
	}
	if len(result.Bytes) == 0 {
		t.Error("expected Bytes to be non-empty")
	}
	if result.Key != "#file#proofs/test.txt" {
		t.Errorf("unexpected Key: %q", result.Key)
	}
}

func TestWebApplicationConvert(t *testing.T) {
	wa := WebApplication{
		PrimaryURL: "https://example.com",
		Name:       "Example App",
		URLs:       []string{"https://api.example.com"},
	}
	result, err := wa.Convert()
	if err != nil {
		t.Fatalf("WebApplication.Convert() error: %v", err)
	}

	if !strings.HasPrefix(result.PrimaryURL, "https://example.com") {
		t.Errorf("expected PrimaryURL to start with https://example.com, got %q", result.PrimaryURL)
	}
	if result.Name != "Example App" {
		t.Errorf("expected Name=Example App, got %q", result.Name)
	}
	if result.Key == "" {
		t.Error("expected Key to be set by hooks")
	}
}

func TestWebpageConvert(t *testing.T) {
	wp := Webpage{
		URL: "https://example.com/login",
		Parent: WebApplication{
			PrimaryURL: "https://example.com",
			Name:       "Example",
		},
	}
	result, err := wp.Convert()
	if err != nil {
		t.Fatalf("Webpage.Convert() error: %v", err)
	}

	if result.URL != "https://example.com/login" {
		t.Errorf("expected URL=https://example.com/login, got %q", result.URL)
	}
	if result.Parent == nil {
		t.Fatal("expected Parent to be set")
	}
}

func TestPreseedConvert(t *testing.T) {
	p := Preseed{
		Type:  "whois",
		Title: "registrant_email",
		Value: "admin@example.com",
	}
	result, err := p.Convert()
	if err != nil {
		t.Fatalf("Preseed.Convert() error: %v", err)
	}

	if result.Type != "whois" {
		t.Errorf("expected Type=whois, got %q", result.Type)
	}
	if result.Title != "registrant_email" {
		t.Errorf("expected Title=registrant_email, got %q", result.Title)
	}
	if result.Value != "admin@example.com" {
		t.Errorf("expected Value=admin@example.com, got %q", result.Value)
	}
	if result.Key == "" {
		t.Error("expected Key to be set by hooks")
	}
}
