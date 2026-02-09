package slim

import (
	"strings"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/collection"
	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

func TestConvertAssetTypes(t *testing.T) {
	tests := []struct {
		name            string
		slim            Converter
		expectedDNS     string
		expectedName    string
		expectedKey     string
		expectedClass   string
		expectedPrivate bool
	}{
		{
			name:          "IP with parent domain",
			slim:          IP{Address: "1.2.3.4", ParentDomain: "example.com"},
			expectedDNS:   "example.com",
			expectedName:  "1.2.3.4",
			expectedKey:   "#asset#example.com#1.2.3.4",
			expectedClass: "ipv4",
		},
		{
			name:            "standalone IP",
			slim:            IP{Address: "10.0.0.1", ParentDomain: "10.0.0.1"},
			expectedDNS:     "10.0.0.1",
			expectedName:    "10.0.0.1",
			expectedKey:     "#asset#10.0.0.1#10.0.0.1",
			expectedClass:   "ipv4",
			expectedPrivate: true,
		},
		{
			name:          "IPv6 address",
			slim:          IP{Address: "::1", ParentDomain: "example.com"},
			expectedDNS:   "example.com",
			expectedName:  "::1",
			expectedKey:   "#asset#example.com#::1",
			expectedClass: "ipv6",
		},
		{
			name:          "SlimAsset with dns and name",
			slim:          SlimAsset{DNS: "example.com", Name: "1.2.3.4"},
			expectedDNS:   "example.com",
			expectedName:  "1.2.3.4",
			expectedKey:   "#asset#example.com#1.2.3.4",
			expectedClass: "ipv4",
		},
		{
			name:          "domain via SlimAsset",
			slim:          SlimAsset{DNS: "sub.example.com", Name: "sub.example.com"},
			expectedDNS:   "sub.example.com",
			expectedName:  "sub.example.com",
			expectedKey:   "#asset#sub.example.com#sub.example.com",
			expectedClass: "domain",
		},
		{
			name:            "CIDR via SlimAsset",
			slim:            SlimAsset{DNS: "10.0.0.0/8", Name: "10.0.0.0/8"},
			expectedDNS:     "10.0.0.0/8",
			expectedName:    "10.0.0.0/8",
			expectedKey:     "#asset#10.0.0.0/8#10.0.0.0/8",
			expectedClass:   "cidr",
			expectedPrivate: true,
		},
		{
			name:            "private IP",
			slim:            IP{Address: "192.168.1.100", ParentDomain: "internal.local"},
			expectedDNS:     "internal.local",
			expectedName:    "192.168.1.100",
			expectedKey:     "#asset#internal.local#192.168.1.100",
			expectedClass:   "ipv4",
			expectedPrivate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col, err := Convert(tt.slim)
			if err != nil {
				t.Fatalf("Convert() error: %v", err)
			}

			assets := collection.Get[*model.Asset](col)
			if len(assets) != 1 {
				t.Fatalf("expected 1 asset, got %d", len(assets))
			}

			a := assets[0]
			if a.DNS != tt.expectedDNS {
				t.Errorf("DNS = %q, want %q", a.DNS, tt.expectedDNS)
			}
			if a.Name != tt.expectedName {
				t.Errorf("Name = %q, want %q", a.Name, tt.expectedName)
			}
			if a.Key != tt.expectedKey {
				t.Errorf("Key = %q, want %q", a.Key, tt.expectedKey)
			}
			if a.Class != tt.expectedClass {
				t.Errorf("Class = %q, want %q", a.Class, tt.expectedClass)
			}
			if a.Source != "self" {
				t.Errorf("Source = %q, want %q", a.Source, "self")
			}
			if a.Status != "A" {
				t.Errorf("Status = %q, want %q", a.Status, "A")
			}
			if a.Created == "" {
				t.Error("Created should be set by Defaulted()")
			}
			if a.Private != tt.expectedPrivate {
				t.Errorf("Private = %v, want %v", a.Private, tt.expectedPrivate)
			}
		})
	}
}

func TestConvertPort(t *testing.T) {
	slim := SlimPort{
		Asset:      SlimAsset{DNS: "example.com", Name: "1.2.3.4"},
		Protocol:   "tcp",
		Port:       443,
		Service:    "https",
		Capability: "portscan",
	}

	col, err := Convert(slim)
	if err != nil {
		t.Fatalf("Convert() error: %v", err)
	}

	if col.Count != 2 {
		t.Fatalf("Collection.Count = %d, want 2", col.Count)
	}

	// Verify the parent Asset.
	assets := collection.Get[*model.Asset](col)
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(assets))
	}
	expectedAssetKey := "#asset#example.com#1.2.3.4"
	if assets[0].Key != expectedAssetKey {
		t.Errorf("Asset.Key = %q, want %q", assets[0].Key, expectedAssetKey)
	}

	// Verify the Port.
	ports := collection.Get[*model.Port](col)
	if len(ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(ports))
	}
	p := ports[0]
	expectedPortKey := "#port#tcp#443" + expectedAssetKey
	if p.Key != expectedPortKey {
		t.Errorf("Port.Key = %q, want %q", p.Key, expectedPortKey)
	}
	if p.Source != expectedAssetKey {
		t.Errorf("Port.Source = %q, want %q", p.Source, expectedAssetKey)
	}
	if p.Protocol != "tcp" {
		t.Errorf("Port.Protocol = %q, want %q", p.Protocol, "tcp")
	}
	if p.Port != 443 {
		t.Errorf("Port.Port = %d, want %d", p.Port, 443)
	}
	if p.Service != "https" {
		t.Errorf("Port.Service = %q, want %q", p.Service, "https")
	}
	if p.Capability != "portscan" {
		t.Errorf("Port.Capability = %q, want %q", p.Capability, "portscan")
	}
	if p.Status != "A" {
		t.Errorf("Port.Status = %q, want %q", p.Status, "A")
	}
	if p.Created == "" {
		t.Error("Port.Created should be set by Defaulted()")
	}
	if p.TTL == 0 {
		t.Error("Port.TTL should be set by Defaulted()")
	}
}

func TestConvertRisk(t *testing.T) {
	t.Run("CVE risk", func(t *testing.T) {
		col, err := Convert(SlimRisk{
			DNS:     "example.com",
			Name:    "CVE-2024-1234",
			Comment: "test vulnerability",
		})
		if err != nil {
			t.Fatalf("Convert() error: %v", err)
		}

		if col.Count != 1 {
			t.Fatalf("Collection.Count = %d, want 1", col.Count)
		}

		risks := collection.Get[*model.Risk](col)
		if len(risks) != 1 {
			t.Fatalf("expected 1 risk, got %d", len(risks))
		}

		r := risks[0]
		if r.DNS != "example.com" {
			t.Errorf("Risk.DNS = %q, want %q", r.DNS, "example.com")
		}
		if r.Name != "CVE-2024-1234" {
			t.Errorf("Risk.Name = %q, want %q", r.Name, "CVE-2024-1234")
		}
		if r.Key != "#risk#example.com#CVE-2024-1234" {
			t.Errorf("Risk.Key = %q, want %q", r.Key, "#risk#example.com#CVE-2024-1234")
		}
		if r.Source != "provided" {
			t.Errorf("Risk.Source = %q, want %q", r.Source, "provided")
		}
		if r.Created == "" {
			t.Error("Risk.Created should be set by Defaulted()")
		}
		if r.TTL == 0 {
			t.Error("Risk.TTL should be set by Defaulted()")
		}
	})

	t.Run("non-CVE name formatting", func(t *testing.T) {
		col, err := Convert(SlimRisk{DNS: "example.com", Name: "Test Risk Name"})
		if err != nil {
			t.Fatalf("Convert() error: %v", err)
		}

		risks := collection.Get[*model.Risk](col)
		if len(risks) != 1 {
			t.Fatalf("expected 1 risk, got %d", len(risks))
		}

		r := risks[0]
		if r.Name != "test-risk-name" {
			t.Errorf("Risk.Name = %q, want %q (hooks should format non-CVE names)", r.Name, "test-risk-name")
		}
		if r.Key != "#risk#example.com#test-risk-name" {
			t.Errorf("Risk.Key = %q, want %q", r.Key, "#risk#example.com#test-risk-name")
		}
	})
}

func TestConvertTechnology(t *testing.T) {
	cpe := "cpe:2.3:a:nginx:nginx:1.25.0:*:*:*:*:*:*:*"
	col, err := Convert(SlimTechnology{CPE: cpe, Name: "nginx"})
	if err != nil {
		t.Fatalf("Convert() error: %v", err)
	}

	if col.Count != 1 {
		t.Fatalf("Collection.Count = %d, want 1", col.Count)
	}

	techs := collection.Get[*model.Technology](col)
	if len(techs) != 1 {
		t.Fatalf("expected 1 technology, got %d", len(techs))
	}

	tech := techs[0]
	if tech.CPE != cpe {
		t.Errorf("Technology.CPE = %q, want %q", tech.CPE, cpe)
	}
	if tech.Name != "nginx" {
		t.Errorf("Technology.Name = %q, want %q", tech.Name, "nginx")
	}
	if tech.Key != "#technology#"+cpe {
		t.Errorf("Technology.Key = %q, want %q", tech.Key, "#technology#"+cpe)
	}
	if tech.Created == "" {
		t.Error("Technology.Created should be set by Defaulted()")
	}
	if tech.TTL == 0 {
		t.Error("Technology.TTL should be set by Defaulted()")
	}
}

func TestConvertAttribute(t *testing.T) {
	col, err := Convert(SlimAttribute{
		Asset:      SlimAsset{DNS: "example.com", Name: "1.2.3.4"},
		Name:       "open_port",
		Value:      "443",
		Capability: "portscan",
		Metadata:   map[string]string{"tool": "masscan"},
	})
	if err != nil {
		t.Fatalf("Convert() error: %v", err)
	}

	if col.Count != 2 {
		t.Fatalf("Collection.Count = %d, want 2", col.Count)
	}

	// Verify the parent Asset.
	assets := collection.Get[*model.Asset](col)
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(assets))
	}
	expectedAssetKey := "#asset#example.com#1.2.3.4"
	if assets[0].Key != expectedAssetKey {
		t.Errorf("Asset.Key = %q, want %q", assets[0].Key, expectedAssetKey)
	}

	// Verify the Attribute.
	attrs := collection.Get[*model.Attribute](col)
	if len(attrs) != 1 {
		t.Fatalf("expected 1 attribute, got %d", len(attrs))
	}
	attr := attrs[0]
	expectedAttrKey := "#attribute#open_port#443" + expectedAssetKey
	if attr.Key != expectedAttrKey {
		t.Errorf("Attribute.Key = %q, want %q", attr.Key, expectedAttrKey)
	}
	if attr.Source != expectedAssetKey {
		t.Errorf("Attribute.Source = %q, want %q", attr.Source, expectedAssetKey)
	}
	if attr.Name != "open_port" {
		t.Errorf("Attribute.Name = %q, want %q", attr.Name, "open_port")
	}
	if attr.Value != "443" {
		t.Errorf("Attribute.Value = %q, want %q", attr.Value, "443")
	}
	if attr.Capability != "portscan" {
		t.Errorf("Attribute.Capability = %q, want %q", attr.Capability, "portscan")
	}
	if v, ok := attr.Metadata["tool"]; !ok || v != "masscan" {
		t.Errorf("Attribute.Metadata[tool] = %q, want %q", v, "masscan")
	}
	if attr.Status != "A" {
		t.Errorf("Attribute.Status = %q, want %q", attr.Status, "A")
	}
	if attr.Created == "" {
		t.Error("Attribute.Created should be set by Defaulted()")
	}
	if attr.TTL == 0 {
		t.Error("Attribute.TTL should be set by Defaulted()")
	}
}

func TestConvertFile(t *testing.T) {
	col, err := Convert(SlimFile{
		Name:  "proofs/scan.txt",
		Bytes: []byte("scan results"),
	})
	if err != nil {
		t.Fatalf("Convert() error: %v", err)
	}

	if col.Count != 1 {
		t.Fatalf("Collection.Count = %d, want 1", col.Count)
	}

	files := collection.Get[*model.File](col)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if f.Name != "proofs/scan.txt" {
		t.Errorf("File.Name = %q, want %q", f.Name, "proofs/scan.txt")
	}
	if f.Key != "#file#proofs/scan.txt" {
		t.Errorf("File.Key = %q, want %q", f.Key, "#file#proofs/scan.txt")
	}
	if string(f.Bytes) != "scan results" {
		t.Errorf("File.Bytes = %q, want %q", string(f.Bytes), "scan results")
	}
	if f.Updated == "" {
		t.Error("File.Updated should be set by Defaulted()")
	}
}

func TestConvertWebpage(t *testing.T) {
	col, err := Convert(SlimWebpage{
		Asset: SlimAsset{DNS: "example.com", Name: "1.2.3.4"},
		URL:   "https://example.com/login",
	})
	if err != nil {
		t.Fatalf("Convert() error: %v", err)
	}

	if col.Count != 2 {
		t.Fatalf("Collection.Count = %d, want 2", col.Count)
	}

	// Verify the parent Asset.
	assets := collection.Get[*model.Asset](col)
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(assets))
	}
	expectedAssetKey := "#asset#example.com#1.2.3.4"
	if assets[0].Key != expectedAssetKey {
		t.Errorf("Asset.Key = %q, want %q", assets[0].Key, expectedAssetKey)
	}

	// Verify the Webpage.
	webpages := collection.Get[*model.Webpage](col)
	if len(webpages) != 1 {
		t.Fatalf("expected 1 webpage, got %d", len(webpages))
	}
	wp := webpages[0]
	if !strings.HasPrefix(wp.Key, "#webpage#https://example.com/login") {
		t.Errorf("Webpage.Key = %q, want prefix %q", wp.Key, "#webpage#https://example.com/login")
	}
	if wp.URL != "https://example.com/login" {
		t.Errorf("Webpage.URL = %q, want %q", wp.URL, "https://example.com/login")
	}
	if wp.Status != "A" {
		t.Errorf("Webpage.Status = %q, want %q", wp.Status, "A")
	}
	if wp.Created == "" {
		t.Error("Webpage.Created should be set by Defaulted()")
	}
	if wp.TTL == 0 {
		t.Error("Webpage.TTL should be set by Defaulted()")
	}
	// Webpage.Parent should be nil because parent injection is deliberately
	// skipped for Webpage (its Parent is *WebApplication, not GraphModelWrapper).
	if wp.Parent != nil {
		t.Error("Webpage.Parent should be nil (parent injection is skipped for Webpage)")
	}
}

func TestConvertWebApplication(t *testing.T) {
	col, err := Convert(SlimWebApplication{
		PrimaryURL: "https://app.example.com",
		URLs:       []string{"https://api.example.com"},
		Name:       "Example App",
	})
	if err != nil {
		t.Fatalf("Convert() error: %v", err)
	}

	if col.Count != 1 {
		t.Fatalf("Collection.Count = %d, want 1", col.Count)
	}

	webapps := collection.Get[*model.WebApplication](col)
	if len(webapps) != 1 {
		t.Fatalf("expected 1 web application, got %d", len(webapps))
	}

	wa := webapps[0]
	if wa.PrimaryURL != "https://app.example.com/" {
		t.Errorf("WebApplication.PrimaryURL = %q, want %q", wa.PrimaryURL, "https://app.example.com/")
	}
	if wa.Name != "Example App" {
		t.Errorf("WebApplication.Name = %q, want %q", wa.Name, "Example App")
	}
	if !strings.HasPrefix(wa.Key, "#webapplication#https://app.example.com") {
		t.Errorf("WebApplication.Key = %q, want prefix %q", wa.Key, "#webapplication#https://app.example.com")
	}
	if wa.Status != "A" {
		t.Errorf("WebApplication.Status = %q, want %q", wa.Status, "A")
	}
	if wa.Source != "self" {
		t.Errorf("WebApplication.Source = %q, want %q", wa.Source, "self")
	}
	if wa.Created == "" {
		t.Error("WebApplication.Created should be set by Defaulted()")
	}
}

func TestConvertUnknownModel(t *testing.T) {
	_, err := Convert(unknownModel{})
	if err == nil {
		t.Fatal("expected error for unknown model, got nil")
	}
}

// unknownModel is a test helper that implements Converter with an unregistered model name.
type unknownModel struct{}

func (unknownModel) TargetModel() string          { return "nonexistent" }
func (unknownModel) MarshalJSON() ([]byte, error)  { return []byte(`{}`), nil }
