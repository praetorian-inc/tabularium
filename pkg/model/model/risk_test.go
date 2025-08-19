package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRisk_StateSeverity(t *testing.T) {
	tests := []struct {
		status   string
		state    string
		severity string
	}{
		{"TI", "T", "I"},
		{"D", "D", ""},
		{"", "", ""},
	}

	for _, test := range tests {
		t.Run(test.status, func(t *testing.T) {
			risk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test", test.status)
			assert.Equal(t, test.state, risk.State())
			assert.Equal(t, test.severity, risk.Severity())
			assert.True(t, risk.Is(test.state), "expected Is(%s) to return true for %s", test.state, test.status)
		})
	}
}

func TestRisk_Set(t *testing.T) {
	tests := []struct {
		initial          string
		state            string
		expected         string
		expectedPriority int
	}{
		{OpenCritical, Remediated, RemediatedCritical, 0},
		{DeletedCriticalDuplicate, Open, OpenCritical, 0},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			risk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test", test.initial)
			risk.Set(test.state)
			assert.Equal(t, test.expected, risk.Status)
			assert.Equal(t, test.expectedPriority, risk.Priority)
		})
	}
}

func TestRisk_MergePriority(t *testing.T) {
	tests := []struct {
		initial          string
		update           string
		expected         string
		expectedPriority int
	}{
		{OpenCritical, OpenLow, OpenLow, 30},
		{DeletedLowDuplicate, OpenHigh, OpenHigh, 10},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			risk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test", test.initial)
			update := Risk{Status: test.update}
			risk.Merge(update)
			assert.Equal(t, test.expected, risk.Status)
			assert.Equal(t, test.expectedPriority, risk.Priority)
		})
	}
}

func TestRiskConstructors(t *testing.T) {
	testAsset := NewAsset("example.com", "Example Asset")
	testAttribute := NewAttribute("test", "test", &testAsset)
	testWebpage := NewWebpageFromString("https://gladiator.systems", &testAttribute)
	tests := []struct {
		name         string
		target       Target
		riskName     string
		expectedName string
		dns          string
	}{
		{
			name:     "Same DNS",
			target:   &testAsset,
			riskName: "test-risk",
			dns:      "example.com",
		},
		{
			name:     "Different DNS",
			target:   &testAsset,
			riskName: "test-risk",
			dns:      "subdomain.example.com",
		},
		{
			name:     "Same DNS",
			target:   &testAttribute,
			riskName: "test-risk",
			dns:      "example.com",
		},
		{
			name:     "Different DNS",
			target:   &testAttribute,
			riskName: "test-risk",
			dns:      "subdomain.example.com",
		},
		{
			name:     "Same DNS",
			target:   &testWebpage,
			riskName: "test-risk",
			dns:      "example.com",
		},
		{
			name:     "Different DNS",
			target:   &testWebpage,
			riskName: "test-risk",
			dns:      "subdomain.example.com",
		},
		{
			name:         "Format Name",
			target:       &testAsset,
			riskName:     "Test Risk",
			expectedName: "test-risk",
			dns:          "example.com",
		},
		{
			name:         "Format Name (CVE)",
			target:       &testAsset,
			riskName:     "CVE-2023-12345",
			expectedName: "CVE-2023-12345",
			dns:          "example.com",
		},
		{
			name:         "Format Name (CVE should be uppercase)",
			target:       &testAsset,
			riskName:     "cve-2023-12345",
			expectedName: "CVE-2023-12345",
			dns:          "example.com",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			risk1 := NewRisk(test.target, test.riskName, TriageInfo)
			assert.Equal(t, test.target.Group(), risk1.DNS, "NewRisk: DNS should match target group")

			risk2 := NewRiskWithDNS(test.target, test.riskName, test.dns, TriageInfo)
			assert.Equal(t, test.dns, risk2.DNS, "NewRiskWithDNS: DNS should match provided DNS")
			assert.Equal(t, risk1.Name, risk2.Name, "Names should match")
			assert.Equal(t, risk1.Status, risk2.Status, "Status should match") 
			assert.Equal(t, risk1.Source, risk2.Source, "Source should match")
			assert.Equal(t, risk1.Target, risk2.Target, "Target should match")
		})
	}
}

func TestRisk_PendingAsset(t *testing.T) {
	originalAsset := NewAsset("example.com", "Example Asset")
	risk := NewRisk(&originalAsset, "test-risk", TriageInfo)

	pendingAsset, ok := risk.PendingAsset()
	if !ok {
		t.Errorf("expected PendingAsset to return a valid asset")
	}

	assert.Equal(t, Pending, pendingAsset.Status, "Status should be Pending")
	assert.Equal(t, originalAsset.Key, pendingAsset.Key, "Key should not change")
	assert.Equal(t, originalAsset.DNS, pendingAsset.DNS, "DNS should not change")
	assert.Equal(t, originalAsset.Name, pendingAsset.Name, "Name should not change")
	assert.Equal(t, originalAsset.Source, pendingAsset.Source, "Source should not change")
	assert.Equal(t, originalAsset.Created, pendingAsset.Created, "Created should not change")
	assert.Equal(t, originalAsset.Visited, pendingAsset.Visited, "Visited should not change")
	assert.Equal(t, originalAsset.TTL, pendingAsset.TTL, "TTL should not change")

	attr := NewAttribute("test-attr", "test-value", &originalAsset)
	attrRisk := NewRisk(&attr, "test-risk", TriageInfo)

	// True negative
	pendingAsset, ok = attrRisk.PendingAsset()
	if ok {
		t.Errorf("expected PendingAsset to return false for attribute-based risk")
	}
}

func TestRisk_Valid(t *testing.T) {
	validRisk := NewRisk(&Asset{DNS: "test", Name: "test"}, "test", TriageInfo)
	missingKey := NewRisk(&Asset{DNS: "test", Name: "test"}, "test", TriageInfo)
	missingKey.Key = ""
	missingStatus := NewRisk(&Asset{DNS: "test", Name: "test"}, "test", "")
	missingName := NewRisk(&Asset{DNS: "test", Name: "test"}, "", TriageInfo)
	missingDNS := NewRiskWithDNS(&Asset{DNS: "test", Name: "test"}, "test", "", TriageInfo)

	assert.True(t, validRisk.Valid())
	assert.False(t, missingKey.Valid())
	assert.False(t, missingStatus.Valid())
	assert.False(t, missingName.Valid())
	assert.False(t, missingDNS.Valid())
}

func TestRisk_Proof(t *testing.T) {
	tests := []struct {
		name         string
		dns          string
		riskName     string
		proofData    []byte
		expectedPath string
		description  string
	}{
		{
			name:         "JSON with valid port",
			dns:          "example.com",
			riskName:     "test-risk",
			proofData:    []byte(`{"port": "8080", "service": "http"}`),
			expectedPath: "proofs/example.com/test-risk/8080",
			description:  "Should use enhanced path format with port",
		},
		{
			name:         "JSON with numeric port as string",
			dns:          "example.com",
			riskName:     "ssh-vuln",
			proofData:    []byte(`{"port": "22", "protocol": "ssh"}`),
			expectedPath: "proofs/example.com/ssh-vuln/22",
			description:  "Should handle numeric port values as strings",
		},
		{
			name:         "JSON with complex port string",
			dns:          "test.example.com",
			riskName:     "complex-service",
			proofData:    []byte(`{"port": "8080-8090", "range": true}`),
			expectedPath: "proofs/test.example.com/complex-service/8080-8090",
			description:  "Should handle complex port strings",
		},
		{
			name:         "JSON without port field",
			dns:          "example.com",
			riskName:     "no-port-risk",
			proofData:    []byte(`{"service": "http", "method": "GET"}`),
			expectedPath: "proofs/example.com/no-port-risk",
			description:  "Should fall back to original format when port field is missing",
		},
		{
			name:         "JSON with empty port string",
			dns:          "example.com",
			riskName:     "empty-port",
			proofData:    []byte(`{"port": "", "service": "unknown"}`),
			expectedPath: "proofs/example.com/empty-port",
			description:  "Should fall back to original format when port is empty string",
		},
		{
			name:         "JSON with port as number",
			dns:          "example.com",
			riskName:     "numeric-port",
			proofData:    []byte(`{"port": 443, "service": "https"}`),
			expectedPath: "proofs/example.com/numeric-port",
			description:  "Should fall back to original format when port is not a string",
		},
		{
			name:         "JSON with port as boolean",
			dns:          "example.com",
			riskName:     "bool-port",
			proofData:    []byte(`{"port": true, "service": "test"}`),
			expectedPath: "proofs/example.com/bool-port",
			description:  "Should fall back to original format when port is not a string",
		},
		{
			name:         "JSON with port as null",
			dns:          "example.com",
			riskName:     "null-port",
			proofData:    []byte(`{"port": null, "service": "test"}`),
			expectedPath: "proofs/example.com/null-port",
			description:  "Should fall back to original format when port is null",
		},
		{
			name:         "JSON with whitespace-only port",
			dns:          "example.com",
			riskName:     "whitespace-port",
			proofData:    []byte(`{"port": "   ", "service": "test"}`),
			expectedPath: "proofs/example.com/whitespace-port/   ",
			description:  "Should use whitespace port as valid (non-empty string)",
		},
		{
			name:         "Non-JSON data",
			dns:          "example.com",
			riskName:     "plain-text",
			proofData:    []byte("This is plain text proof data"),
			expectedPath: "proofs/example.com/plain-text",
			description:  "Should fall back to original format for non-JSON data",
		},
		{
			name:         "Empty proof data",
			dns:          "example.com",
			riskName:     "empty-proof",
			proofData:    []byte{},
			expectedPath: "proofs/example.com/empty-proof",
			description:  "Should fall back to original format for empty data",
		},
		{
			name:         "Nil proof data",
			dns:          "example.com",
			riskName:     "nil-proof",
			proofData:    nil,
			expectedPath: "proofs/example.com/nil-proof",
			description:  "Should fall back to original format for nil data",
		},
		{
			name:         "Invalid JSON",
			dns:          "example.com",
			riskName:     "invalid-json",
			proofData:    []byte(`{"port": "8080", invalid json`),
			expectedPath: "proofs/example.com/invalid-json",
			description:  "Should fall back to original format for malformed JSON",
		},
		{
			name:         "JSON with multiple fields including port",
			dns:          "api.example.com",
			riskName:     "api-vulnerability",
			proofData:    []byte(`{"host": "api.example.com", "port": "3000", "path": "/api/v1", "method": "POST", "vulnerable": true}`),
			expectedPath: "proofs/api.example.com/api-vulnerability/3000",
			description:  "Should extract port from complex JSON structure",
		},
		{
			name:         "Edge case: complex DNS and risk name",
			dns:          "sub-domain.example-site.co.uk",
			riskName:     "CVE-2023-12345",
			proofData:    []byte(`{"port": "443", "ssl": true}`),
			expectedPath: "proofs/sub-domain.example-site.co.uk/CVE-2023-12345/443",
			description:  "Should handle complex DNS and CVE names with port",
		},
		{
			name:         "JSON with port containing special characters",
			dns:          "example.com",
			riskName:     "special-chars",
			proofData:    []byte(`{"port": "8080/tcp", "protocol": "http"}`),
			expectedPath: "proofs/example.com/special-chars/8080/tcp",
			description:  "Should handle port strings with special characters",
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a risk with the specified DNS and name
			risk := Risk{
				DNS:  test.dns,
				Name: test.riskName,
			}
			
			// Call the Proof method
			proofFile := risk.Proof(test.proofData)
			
			// Verify the file name matches expected path
			assert.Equal(t, test.expectedPath, proofFile.Name, test.description)
			
			// Verify the proof data is preserved
			assert.Equal(t, test.proofData, []byte(proofFile.Bytes), "Proof data should be preserved unchanged")
		})
	}
}

func TestRisk_Proof_PortFieldVariations(t *testing.T) {
	// Test various edge cases for port field extraction
	tests := []struct {
		name         string
		proofJSON    string
		expectedPath string
		usePort      bool
	}{
		{
			name:         "port as nested object",
			proofJSON:    `{"port": {"number": 8080}, "service": "http"}`,
			expectedPath: "proofs/example.com/test-risk", // should fall back
			usePort:      false,
		},
		{
			name:         "port as array",
			proofJSON:    `{"port": ["8080", "8081"], "service": "http"}`,
			expectedPath: "proofs/example.com/test-risk", // should fall back
			usePort:      false,
		},
		{
			name:         "port with zero value",
			proofJSON:    `{"port": "0", "service": "reserved"}`,
			expectedPath: "proofs/example.com/test-risk/0",
			usePort:      true,
		},
		{
			name:         "port with large number",
			proofJSON:    `{"port": "65535", "service": "max-port"}`,
			expectedPath: "proofs/example.com/test-risk/65535",
			usePort:      true,
		},
		{
			name:         "multiple port fields (first one wins)",
			proofJSON:    `{"port": "8080", "Port": "9090", "PORT": "3000"}`,
			expectedPath: "proofs/example.com/test-risk/8080",
			usePort:      true,
		},
		{
			name:         "unicode in port field",
			proofJSON:    `{"port": "８０８０", "service": "unicode"}`, // full-width characters
			expectedPath: "proofs/example.com/test-risk/８０８０",
			usePort:      true,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			risk := Risk{
				DNS:  "example.com",
				Name: "test-risk",
			}
			
			proofFile := risk.Proof([]byte(test.proofJSON))
			
			assert.Equal(t, test.expectedPath, proofFile.Name, "Port field parsing should handle edge cases correctly")
			assert.Equal(t, []byte(test.proofJSON), []byte(proofFile.Bytes), "Proof data should be preserved")
		})
	}
}

func TestRisk_Proof_PreservesFileProperties(t *testing.T) {
	// Verify that the Proof method creates a proper File object
	risk := Risk{
		DNS:  "test.example.com",
		Name: "file-props-test",
	}
	
	testData := []byte(`{"port": "9000", "test": "data"}`)
	proofFile := risk.Proof(testData)
	
	// Test File object properties
	assert.NotNil(t, proofFile, "Proof file should not be nil")
	assert.Equal(t, "proofs/test.example.com/file-props-test/9000", proofFile.Name, "File name should be set correctly")
	assert.Equal(t, testData, []byte(proofFile.Bytes), "File bytes should match input data")
	
	// Test with non-JSON data to ensure File properties are still set
	nonJSONData := []byte("plain text proof")
	proofFile2 := risk.Proof(nonJSONData)
	
	assert.NotNil(t, proofFile2, "Proof file should not be nil for non-JSON data")
	assert.Equal(t, "proofs/test.example.com/file-props-test", proofFile2.Name, "File name should use original format for non-JSON")
	assert.Equal(t, nonJSONData, []byte(proofFile2.Bytes), "File bytes should match input data for non-JSON")
}

func TestRisk_Proof_IntegrationWithNewRiskAndProof(t *testing.T) {
	// Test integration between separate NewRisk + Proof calls and the enhanced Proof method
	testAsset := NewAsset("example.com", "Test Asset")
	
	tests := []struct {
		name         string
		proofData    []byte
		expectedPath string
		description  string
	}{
		{
			name:         "JSON proof with port creates enhanced path",
			proofData:    []byte(`{"port": "8080", "vulnerability": "exposed service"}`),
			expectedPath: "proofs/subdomain.example.com/test-vulnerability/8080",
			description:  "Integration should use enhanced path format",
		},
		{
			name:         "Non-JSON proof uses original path",
			proofData:    []byte("Plain text proof data"),
			expectedPath: "proofs/subdomain.example.com/test-vulnerability",
			description:  "Integration should fall back to original path format",
		},
		{
			name:         "JSON proof without port uses original path",
			proofData:    []byte(`{"service": "web", "method": "GET"}`),
			expectedPath: "proofs/subdomain.example.com/test-vulnerability",
			description:  "Integration should fall back when no port in JSON",
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Use the new pattern: separate NewRisk + risk.Proof() calls
			risk := NewRiskWithDNS(&testAsset, "test-vulnerability", "subdomain.example.com", TriageHigh)
			proofFile := risk.Proof(test.proofData)
			relationship := NewHasVulnerability(&testAsset, &risk)
			
			// Verify the ProofPath field is set on the risk
			assert.Equal(t, test.expectedPath, risk.ProofPath, "Risk.ProofPath should be set correctly")
			
			// Verify the proof file path uses the enhanced or fallback logic
			assert.Equal(t, test.expectedPath, proofFile.Name, test.description)
			
			// Verify the proof data is preserved
			assert.Equal(t, test.proofData, []byte(proofFile.Bytes), "Proof data should be preserved")
			
			// Verify risk properties are set correctly
			assert.Equal(t, "test-vulnerability", risk.Name, "Risk name should match")
			assert.Equal(t, "subdomain.example.com", risk.DNS, "Risk DNS should match")
			assert.Equal(t, TriageHigh, risk.Status, "Risk status should match")
			
			// Verify relationship can be created (though AttachmentPath would be set by LocalProcessor)
			hasVulnRel, ok := relationship.(*HasVulnerability)
			assert.True(t, ok, "Relationship should be HasVulnerability type")
			assert.NotNil(t, hasVulnRel, "Relationship should be created successfully")
		})
	}
}
