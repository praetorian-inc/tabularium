package model

import (
	"strings"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const NO_DISTINGUISHED_NAME = "<blank>" // used to make tests easier to read

// Test core ADObject creation and initialization
func TestNewADObject(t *testing.T) {
	tests := []struct {
		name              string
		domain            string
		objectID          string
		distinguishedName string
		class             string
		expectedKey       string
		expectedClass     string
		expectedName      string
	}{
		{
			name:              "create user object",
			domain:            "example.local",
			objectID:          "S-1-5-21-123456789-123456789-123456789-1001",
			distinguishedName: "CN=John Doe,CN=Users,DC=example,DC=local",
			class:             "User",
			expectedKey:       "#aduser#example.local#S-1-5-21-123456789-123456789-123456789-1001",
			expectedClass:     "user",
			expectedName:      "John Doe",
		},
		{
			name:              "create computer object",
			domain:            "corp.com",
			objectID:          "S-1-5-21-123456789-123456789-123456789-1002",
			distinguishedName: "CN=WORKSTATION01,CN=Computers,DC=corp,DC=com",
			class:             "Computer",
			expectedKey:       "#adcomputer#corp.com#S-1-5-21-123456789-123456789-123456789-1002",
			expectedClass:     "computer",
			expectedName:      "WORKSTATION01",
		},
		{
			name:              "create group object",
			domain:            "test.domain",
			objectID:          "S-1-5-21-123456789-123456789-123456789-1003",
			distinguishedName: "CN=Domain Admins,CN=Groups,DC=test,DC=domain",
			class:             "Group",
			expectedKey:       "#adgroup#test.domain#S-1-5-21-123456789-123456789-123456789-1003",
			expectedClass:     "group",
			expectedName:      "Domain Admins",
		},
		{
			name:              "create OU object",
			domain:            "example.local",
			objectID:          "51FB8637-28BC-4816-9A51-984160B207FA",
			distinguishedName: "OU=Sales,DC=example,DC=local",
			class:             "OU",
			expectedKey:       "#adou#example.local#51FB8637-28BC-4816-9A51-984160B207FA",
			expectedClass:     "ou",
			expectedName:      "",
		},
		{
			name:              "DN without CN prefix",
			domain:            "example.local",
			objectID:          "S-1-5-21-123456789-123456789-123456789-1005",
			distinguishedName: "DC=example,DC=local",
			class:             "Domain",
			expectedKey:       "#addomain#example.local#S-1-5-21-123456789-123456789-123456789-1005",
			expectedClass:     "domain",
			expectedName:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := NewADObject(tt.domain, tt.objectID, tt.distinguishedName, tt.class)

			assert.Equal(t, tt.domain, ad.Domain, "Domain should match")
			assert.Equal(t, tt.objectID, ad.ObjectID, "ObjectID should match")
			assert.Equal(t, tt.distinguishedName, ad.DistinguishedName, "DistinguishedName should match")
			assert.Equal(t, tt.expectedKey, ad.Key, "Key should be generated correctly")
			assert.Equal(t, tt.expectedName, ad.Name, "Name should be extracted from DN")
			assert.True(t, ad.Valid(), "ADObject should be valid")
			assert.NotEmpty(t, ad.Created, "Created timestamp should be set")
			assert.NotEmpty(t, ad.Visited, "Visited timestamp should be set")
		})
	}
}

func TestNewADObject_FromAlias(t *testing.T) {
	tests := []struct {
		name  string
		alias string
	}{
		{
			name:  "adobject",
			alias: "adobject",
		},
		{
			name:  "aduser",
			alias: "aduser",
		},
		{
			name:  "adcomputer",
			alias: "adcomputer",
		},
		{
			name:  "adgroup",
			alias: "adgroup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad, ok := registry.Registry.MakeType(tt.alias)
			require.True(t, ok)
			assert.NotNil(t, ad)

			adObject, ok := ad.(*ADObject)
			require.True(t, ok)
			assert.Contains(t, adObject.GetLabels(), ADObjectLabel)
			assert.Equal(t, tt.alias, adObject.Alias)
		})
	}
}

// Test GetLabels functionality
func TestADObject_GetLabels(t *testing.T) {
	ad := ADObject{}
	labels := ad.GetLabels()

	assert.Contains(t, labels, ADObjectLabel, "Should contain ADObject label")
	assert.Contains(t, labels, TTLLabel, "Should contain TTL label")
	assert.Len(t, labels, 3, "Should have exactly 2 labels")
}

// Test GetDescription functionality
func TestADObject_GetDescription(t *testing.T) {
	ad := ADObject{}
	description := ad.GetDescription()

	assert.NotEmpty(t, description, "Description should not be empty")
	assert.Contains(t, description, "Active Directory", "Description should mention Active Directory")
}

// Test Defaulted functionality
func TestADObject_Defaulted(t *testing.T) {
	ad := ADObject{}
	ad.Defaulted()

	// BaseAsset.Defaulted() should set timestamps
	assert.NotEmpty(t, ad.Created, "Created timestamp should be set")
	assert.NotEmpty(t, ad.Visited, "Visited timestamp should be set")
}

// Test GetHooks functionality
func TestADObject_GetHooks(t *testing.T) {
	tests := []struct {
		name          string
		domain        string
		objectID      string
		label         string
		expectedKey   string
		expectedClass string
	}{
		{
			name:          "hook generates correct key and class",
			domain:        "TEST.LOCAL",
			objectID:      "S-1-5-21-123456789-123456789-123456789-1001",
			label:         "ADUser",
			expectedKey:   "#aduser#test.local#S-1-5-21-123456789-123456789-123456789-1001",
			expectedClass: "user",
		},
		{
			name:          "hook handles empty values",
			domain:        "",
			objectID:      "",
			label:         "",
			expectedKey:   "#adobject##",
			expectedClass: "object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := NewADObject(tt.domain, tt.objectID, NO_DISTINGUISHED_NAME, tt.label)

			err := registry.CallHooks(&ad)
			require.NoError(t, err, "Hook should execute without error")

			assert.Equal(t, tt.expectedKey, ad.Key, "Hook should generate correct key")
			assert.Equal(t, tt.expectedClass, ad.Class, "Hook should set correct class")
		})
	}
}

// Test Visit functionality
func TestADObject_Visit(t *testing.T) {
	tests := []struct {
		name     string
		existing ADObject
		visiting Assetlike
		expected ADObject
	}{
		{
			name: "merge with valid ADObject",
			existing: ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
				ADProperties: ADProperties{
					DistinguishedName: "CN=User1,DC=example,DC=local",
					Name:              "User1",
				},
			},
			visiting: &ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
				ADProperties: ADProperties{
					SAMAccountName: "user1",
					DisplayName:    "User One",
					Description:    "Test user account",
				},
			},
			expected: ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
				ADProperties: ADProperties{
					DistinguishedName: "CN=User1,DC=example,DC=local",
					Name:              "User1",
					SAMAccountName:    "user1",
					DisplayName:       "User One",
					Description:       "Test user account",
				},
			},
		},
		{
			name: "no updates if different keys",
			existing: ADObject{
				Domain:   "example.local", // key is derived from domain and objectid
				ObjectID: "S-1-5-21-EXISTING",
				ADProperties: ADProperties{
					SAMAccountName: "existing",
					DisplayName:    "Existing Display",
					Description:    "Existing description",
				},
			},
			visiting: &ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-NEW",
				ADProperties: ADProperties{
					SAMAccountName: "new",
					DisplayName:    "New Display",
					Description:    "New description",
				},
			},
			expected: ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-EXISTING",
				ADProperties: ADProperties{
					SAMAccountName: "existing",
					DisplayName:    "Existing Display",
					Description:    "Existing description",
				},
			},
		},
		{
			name: "handle non-ADObject type",
			existing: ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
			},
			visiting: &Asset{
				Name: "1.2.3.4", DNS: "example.com",
			},
			expected: ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
			},
		},
		{
			name: "partial merge",
			existing: ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
				ADProperties: ADProperties{
					DisplayName: "Existing",
				},
			},
			visiting: &ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
				ADProperties: ADProperties{
					SAMAccountName: "newuser",
					Description:    "New description",
				},
			},
			expected: ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
				ADProperties: ADProperties{
					DisplayName:    "Existing",
					SAMAccountName: "newuser",
					Description:    "New description",
				},
			},
		},
		{
			name: "blank values don't overwrite existing values",
			existing: ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
				ADProperties: ADProperties{
					DisplayName:       "Existing",
					DistinguishedName: "CN=Existing,DC=example,DC=local",
				},
			},
			visiting: &ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
				ADProperties: ADProperties{
					DisplayName:       "New",
					DistinguishedName: "",
				},
			},
			expected: ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
				ADProperties: ADProperties{
					DisplayName:       "New",
					DistinguishedName: "CN=Existing,DC=example,DC=local",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := tt.existing

			err := registry.CallHooks(&ad)
			require.NoError(t, err, "Hook should execute without error")

			err = registry.CallHooks(tt.visiting)
			require.NoError(t, err, "Hook should execute without error")

			err = registry.CallHooks(&tt.expected)
			require.NoError(t, err, "Hook should execute without error")

			ad.Visit(tt.visiting)

			assert.Equal(t, tt.expected.ObjectID, ad.ObjectID, "ObjectID should match expected")
			assert.Equal(t, tt.expected.SID, ad.SID, "SID should match expected")
			assert.Equal(t, tt.expected.SAMAccountName, ad.SAMAccountName, "SAMAccountName should match expected")
			assert.Equal(t, tt.expected.DisplayName, ad.DisplayName, "DisplayName should match expected")
			assert.Equal(t, tt.expected.Description, ad.Description, "Description should match expected")
		})
	}
}

// Test IsClass functionality
func TestADObject_IsClass(t *testing.T) {
	tests := []struct {
		name       string
		label      string
		checkClass string
		expected   bool
	}{
		{
			name:       "exact match",
			label:      ADUserLabel,
			checkClass: "user",
			expected:   true,
		},
		{
			name:       "case insensitive match",
			label:      ADUserLabel,
			checkClass: "USER",
			expected:   true,
		},
		{
			name:       "different class",
			label:      ADComputerLabel,
			checkClass: "user",
			expected:   false,
		},
		{
			name:       "empty object class",
			label:      "",
			checkClass: "user",
			expected:   false,
		},
		{
			name:       "empty check class",
			label:      ADUserLabel,
			checkClass: "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := ADObject{Label: tt.label}
			err := registry.CallHooks(&ad)
			require.NoError(t, err, "Hook should execute without error")

			result := ad.IsClass(tt.checkClass)

			assert.Equal(t, tt.expected, result, "IsClass(\"%s\") got %v, expected %v", tt.checkClass, result, tt.expected)
		})
	}
}

// Test IsInDomain functionality
func TestADObject_IsInDomain(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		checkDomain string
		expected    bool
	}{
		{
			name:        "exact match",
			domain:      "example.local",
			checkDomain: "example.local",
			expected:    true,
		},
		{
			name:        "case insensitive match",
			domain:      "EXAMPLE.LOCAL",
			checkDomain: "example.local",
			expected:    true,
		},
		{
			name:        "different domain",
			domain:      "corp.com",
			checkDomain: "example.local",
			expected:    false,
		},
		{
			name:        "empty domain",
			domain:      "",
			checkDomain: "example.local",
			expected:    false,
		},
		{
			name:        "empty check domain",
			domain:      "example.local",
			checkDomain: "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := ADObject{Domain: tt.domain}
			result := ad.IsInDomain(tt.checkDomain)
			assert.Equal(t, tt.expected, result, "IsInDomain result should match expected")
		})
	}
}

// Test GetParentDN functionality
func TestADObject_GetParentDN(t *testing.T) {
	tests := []struct {
		name              string
		distinguishedName string
		expected          string
	}{
		{
			name:              "standard user DN",
			distinguishedName: "CN=John Doe,CN=Users,DC=example,DC=local",
			expected:          "CN=Users,DC=example,DC=local",
		},
		{
			name:              "computer DN",
			distinguishedName: "CN=COMPUTER01,CN=Computers,DC=corp,DC=com",
			expected:          "CN=Computers,DC=corp,DC=com",
		},
		{
			name:              "OU DN",
			distinguishedName: "OU=Sales,OU=Departments,DC=example,DC=local",
			expected:          "OU=Departments,DC=example,DC=local",
		},
		{
			name:              "domain root DN",
			distinguishedName: "DC=example,DC=local",
			expected:          "DC=local",
		},
		{
			name:              "single component DN",
			distinguishedName: "DC=local",
			expected:          "",
		},
		{
			name:              "empty DN",
			distinguishedName: "",
			expected:          "",
		},
		{
			name:              "DN with spaces",
			distinguishedName: "CN=John Doe, CN=Users, DC=example, DC=local",
			expected:          "CN=Users, DC=example, DC=local",
		},
		{
			name:              "DN with special characters",
			distinguishedName: "CN=O'Brien\\, John,CN=Users,DC=example,DC=local",
			expected:          "CN=Users,DC=example,DC=local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := ADObject{}
			ad.DistinguishedName = tt.distinguishedName
			result := ad.GetParentDN()
			assert.Equal(t, tt.expected, result, "GetParentDN should return correct parent DN")
		})
	}
}

// Test GetOU functionality
func TestADObject_GetOU(t *testing.T) {
	tests := []struct {
		name              string
		distinguishedName string
		expected          string
	}{
		{
			name:              "user in OU",
			distinguishedName: "CN=John Doe,OU=Sales,DC=example,DC=local",
			expected:          "Sales",
		},
		{
			name:              "nested OUs",
			distinguishedName: "CN=John Doe,OU=East,OU=Sales,DC=example,DC=local",
			expected:          "East",
		},
		{
			name:              "computer in OU",
			distinguishedName: "CN=COMP01,OU=Workstations,OU=Computers,DC=corp,DC=com",
			expected:          "Workstations",
		},
		{
			name:              "no OU in DN",
			distinguishedName: "CN=John Doe,CN=Users,DC=example,DC=local",
			expected:          "",
		},
		{
			name:              "OU itself",
			distinguishedName: "OU=Sales,DC=example,DC=local",
			expected:          "",
		},
		{
			name:              "empty DN",
			distinguishedName: "",
			expected:          "",
		},
		{
			name:              "case variations",
			distinguishedName: "CN=User,ou=Sales,DC=example,DC=local",
			expected:          "Sales",
		},
		{
			name:              "OU with spaces",
			distinguishedName: "CN=User,OU=Human Resources,DC=example,DC=local",
			expected:          "Human Resources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := ADObject{}
			ad.DistinguishedName = tt.distinguishedName
			result := ad.GetOU()
			assert.Equal(t, tt.expected, result, "GetOU should return correct OU")
		})
	}
}

// Test GetCommonName functionality
func TestADObject_GetCommonName(t *testing.T) {
	tests := []struct {
		name              string
		distinguishedName string
		nameField         string
		expected          string
	}{
		{
			name:              "extract CN from DN",
			distinguishedName: "CN=John Doe,CN=Users,DC=example,DC=local",
			nameField:         "",
			expected:          "John Doe",
		},
		{
			name:              "CN with comma",
			distinguishedName: "CN=Doe\\, John,CN=Users,DC=example,DC=local",
			nameField:         "",
			expected:          "Doe\\, John",
		},
		{
			name:              "CN only DN",
			distinguishedName: "CN=Administrator",
			nameField:         "",
			expected:          "Administrator",
		},
		{
			name:              "lowercase cn prefix",
			distinguishedName: "cn=user1,DC=example,DC=local",
			nameField:         "",
			expected:          "user1",
		},
		{
			name:              "no CN in DN",
			distinguishedName: "OU=Sales,DC=example,DC=local",
			nameField:         "SalesOU",
			expected:          "SalesOU",
		},
		{
			name:              "empty DN with name field",
			distinguishedName: "",
			nameField:         "TestName",
			expected:          "TestName",
		},
		{
			name:              "empty DN and name",
			distinguishedName: "",
			nameField:         "",
			expected:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := ADObject{}
			ad.DistinguishedName = tt.distinguishedName
			ad.Name = tt.nameField

			result := ad.GetCommonName()
			assert.Equal(t, tt.expected, result, "GetCommonName should return correct value")
		})
	}
}

// Test DN validation edge cases
func TestADObject_CommonName(t *testing.T) {
	tests := []struct {
		name              string
		distinguishedName string
		commonName        string
	}{
		{
			name:              "valid standard DN",
			distinguishedName: "CN=User,DC=example,DC=local",
			commonName:        "User",
		},
		{
			name:              "DN with escaped characters",
			distinguishedName: "CN=O'Brien\\, John Jr.,OU=Sales\\+Marketing,DC=example,DC=local",
			commonName:        "O'Brien\\, John Jr.",
		},
		{
			name:              "DN with Unicode characters",
			distinguishedName: "CN=José García,OU=España,DC=example,DC=local",
			commonName:        "José García",
		},
		{
			name:              "very long DN",
			distinguishedName: "CN=" + strings.Repeat("VeryLongName", 50) + ",DC=example,DC=local",
			commonName:        strings.Repeat("VeryLongName", 50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := NewADObject("example.local", "S-1-5-21-123456789-123456789-123456789-1001", tt.distinguishedName, "user")

			assert.Equal(t, tt.commonName, ad.GetCommonName(), "GetCommonName should return correct value")
		})
	}
}

// Test SID validation and formats
func TestADObject_SIDValidation(t *testing.T) {
	tests := []struct {
		name        string
		objectID    string
		expectedSID string
		description string
	}{
		{
			name:        "standard SID as objectID",
			objectID:    "S-1-5-21-123456789-123456789-123456789-1001",
			expectedSID: "S-1-5-21-123456789-123456789-123456789-1001",
			description: "Standard SID format",
		},
		{
			name:        "UUID as objectID",
			objectID:    "123e4567-e89b-12d3-a456-426614174000",
			expectedSID: "",
			description: "UUID format",
		},
		{
			name:        "empty SID",
			objectID:    "",
			description: "Empty SID should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := ADObject{}
			ad.ObjectID = tt.objectID

			err := registry.CallHooks(&ad)

			require.NoError(t, err, "Hook should execute without error")
			assert.Equal(t, tt.expectedSID, ad.SID, tt.description)
		})
	}
}

// Test domain validation and formats
func TestADObject_DomainValidation(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		expected    string
		description string
	}{
		{
			name:        "FQDN domain",
			domain:      "example.local",
			expected:    "example.local",
			description: "Fully qualified domain name",
		},
		{
			name:        "NetBIOS domain",
			domain:      "EXAMPLE",
			expected:    "example",
			description: "NetBIOS style domain name",
		},
		{
			name:        "multi-level domain",
			domain:      "sub.example.local",
			expected:    "sub.example.local",
			description: "Multi-level domain name",
		},
		{
			name:        "domain with numbers",
			domain:      "example123.local",
			expected:    "example123.local",
			description: "Domain with numbers",
		},
		{
			name:        "domain with hyphens",
			domain:      "example-corp.local",
			expected:    "example-corp.local",
			description: "Domain with hyphens",
		},
		{
			name:        "uppercase domain",
			domain:      "EXAMPLE.LOCAL",
			expected:    "example.local",
			description: "Uppercase domain name",
		},
		{
			name:        "empty domain",
			domain:      "",
			description: "Empty domain should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := NewADObject(tt.domain, "CN=Test,DC=example,DC=local", NO_DISTINGUISHED_NAME, "user")
			assert.Equal(t, tt.expected, ad.Domain, tt.description)
		})
	}
}

func TestADObject_SecurityBehaviors(t *testing.T) {
	t.Run("key generation prevents collision", func(t *testing.T) {
		ad1 := NewADObject("example.local", "S-1-5-21-123456789-123456789-123456789-1001", NO_DISTINGUISHED_NAME, "user")
		ad2 := NewADObject("example.local", "S-1-5-21-123456789-123456789-123456789-1002", NO_DISTINGUISHED_NAME, "user")

		assert.NotEqual(t, ad1.Key, ad2.Key, "Different DNs should generate different keys")
	})

	t.Run("key generation is case insensitive", func(t *testing.T) {
		ad1 := NewADObject("EXAMPLE.LOCAL", "s-1-5-21-123456789-123456789-123456789-1001", NO_DISTINGUISHED_NAME, "user")
		ad2 := NewADObject("example.local", "S-1-5-21-123456789-123456789-123456789-1001", NO_DISTINGUISHED_NAME, "user")

		assert.Equal(t, ad1.Key, ad2.Key, "Keys should be case-insensitive")
	})

	t.Run("domain isolation", func(t *testing.T) {
		ad1 := NewADObject("domain1.local", "S-1-5-21-123456789-123456789-123456789-1001", NO_DISTINGUISHED_NAME, "user")
		ad2 := NewADObject("domain2.local", "S-1-5-21-123456789-123456789-123456789-1001", NO_DISTINGUISHED_NAME, "user")

		assert.NotEqual(t, ad1.Key, ad2.Key, "Same DN in different domains should have different keys")
		assert.False(t, ad1.IsInDomain("domain2.local"), "Object should not be in different domain")
		assert.True(t, ad1.IsInDomain("domain1.local"), "Object should be in its own domain")
	})
}

// Test factory methods for specific AD object types (future implementation)
func TestADObject_FactoryMethods(t *testing.T) {
	t.Run("NewADUser", func(t *testing.T) {
		// Test for future NewADUser factory method
		user := NewADUser("example.local", "S-1-5-21-123456789-123456789-123456789-1001", NO_DISTINGUISHED_NAME)

		assert.Equal(t, "example.local", user.Domain)
		assert.Equal(t, "S-1-5-21-123456789-123456789-123456789-1001", user.ObjectID)
		assert.Equal(t, "S-1-5-21-123456789-123456789-123456789-1001", user.SID)
		assert.Equal(t, ADUserLabel, user.Label)
		assert.Equal(t, "user", user.Class)
	})

	t.Run("NewADComputer", func(t *testing.T) {
		// Test for future NewADComputer factory method
		computer := NewADComputer("corp.com", "S-1-5-21-123456789-123456789-123456789-1002", NO_DISTINGUISHED_NAME)

		assert.Equal(t, "corp.com", computer.Domain)
		assert.Equal(t, "S-1-5-21-123456789-123456789-123456789-1002", computer.ObjectID)
		assert.Equal(t, "S-1-5-21-123456789-123456789-123456789-1002", computer.SID)
		assert.Equal(t, ADComputerLabel, computer.Label)
		assert.Equal(t, "computer", computer.Class)
	})

	t.Run("NewADGroup", func(t *testing.T) {
		// Test for future NewADGroup factory method
		group := NewADGroup("example.local", "S-1-5-21-123456789-123456789-123456789-1003", NO_DISTINGUISHED_NAME)

		assert.Equal(t, "example.local", group.Domain)
		assert.Equal(t, "S-1-5-21-123456789-123456789-123456789-1003", group.ObjectID)
		assert.Equal(t, "S-1-5-21-123456789-123456789-123456789-1003", group.SID)
		assert.Equal(t, ADGroupLabel, group.Label)
		assert.Equal(t, "group", group.Class)
	})

	t.Run("NewADGPO", func(t *testing.T) {
		// Test for future NewADGPO factory method
		gpo := NewADGPO("example.local", "31B2F340-016D-11D2-945F-00C04FB984F9", NO_DISTINGUISHED_NAME)

		assert.Equal(t, "example.local", gpo.Domain)
		assert.Equal(t, "31B2F340-016D-11D2-945F-00C04FB984F9", gpo.ObjectID)
		assert.Equal(t, "", gpo.SID)
		assert.Equal(t, ADGPOLabel, gpo.Label)
		assert.Equal(t, "gpo", gpo.Class)
	})

	t.Run("NewADOU", func(t *testing.T) {
		// Test for future NewADOU factory method
		ou := NewADOU("example.local", "S-1-5-21-123456789-123456789-123456789-1004", NO_DISTINGUISHED_NAME)

		assert.Equal(t, "example.local", ou.Domain)
		assert.Equal(t, "S-1-5-21-123456789-123456789-123456789-1004", ou.ObjectID)
		assert.Equal(t, "S-1-5-21-123456789-123456789-123456789-1004", ou.SID)
		assert.Equal(t, ADOULabel, ou.Label)
		assert.Equal(t, "ou", ou.Class)
	})
}

// Test extensive property list (future implementation)
func TestADObject_ExtensiveProperties(t *testing.T) {
	t.Run("security properties", func(t *testing.T) {
		ad := ADObject{
			ADProperties: ADProperties{
				AdminCount:              true,
				Sensitive:               true,
				HasSPN:                  true,
				UnconstrainedDelegation: true,
				TrustedToAuth:           true,
			},
		}

		assert.True(t, ad.AdminCount)
		assert.True(t, ad.Sensitive)
		assert.True(t, ad.HasSPN)
		assert.True(t, ad.UnconstrainedDelegation)
		assert.True(t, ad.TrustedToAuth)
		assert.True(t, ad.IsPrivileged())
	})

	t.Run("account properties", func(t *testing.T) {
		ad := ADObject{
			ADProperties: ADProperties{
				PasswordNeverExpires:     true,
				PasswordNotRequired:      false,
				DontRequirePreAuth:       true,
				SmartcardRequired:        false,
				LockedOut:                true,
				PasswordExpired:          true,
				UserCannotChangePassword: true,
			},
		}

		assert.True(t, ad.PasswordNeverExpires)
		assert.False(t, ad.PasswordNotRequired)
		assert.True(t, ad.DontRequirePreAuth)
		assert.False(t, ad.SmartcardRequired)
		assert.True(t, ad.LockedOut)
		assert.True(t, ad.PasswordExpired)
		assert.True(t, ad.UserCannotChangePassword)
	})

	t.Run("LAPS properties", func(t *testing.T) {
		ad := ADObject{
			ADProperties: ADProperties{
				HasLAPS: true,
			},
		}

		assert.True(t, ad.HasLAPS)
	})

	t.Run("certificate properties", func(t *testing.T) {
		ad := ADObject{
			ADProperties: ADProperties{
				CertThumbprint:  "ABC123DEF456",
				CertThumbprints: []string{"ABC123DEF456", "789GHI012JKL"},
				CertChain:       []string{"root", "intermediate", "leaf"},
				CertName:        "test-cert",
				CAName:          "Example-CA",
			},
		}

		assert.Equal(t, "ABC123DEF456", ad.CertThumbprint)
		assert.Len(t, ad.CertThumbprints, 2)
		assert.Contains(t, ad.CertThumbprints, "ABC123DEF456")
		assert.Contains(t, ad.CertThumbprints, "789GHI012JKL")
		assert.Len(t, ad.CertChain, 3)
		assert.Equal(t, "test-cert", ad.CertName)
		assert.Equal(t, "Example-CA", ad.CAName)
	})
}

// Test helper methods (future implementation)
func TestADObject_HelperMethods(t *testing.T) {
	t.Run("GetEffectiveDomain", func(t *testing.T) {
		tests := []struct {
			name     string
			ad       ADObject
			expected string
		}{
			{
				name:     "use domain field",
				ad:       ADObject{Domain: "example.local"},
				expected: "example.local",
			},
			{
				name:     "extract from DN when domain empty",
				ad:       ADObject{ADProperties: ADProperties{DistinguishedName: "CN=User,DC=corp,DC=com"}},
				expected: "corp.com",
			},
			{
				name:     "use NetBIOS when available",
				ad:       ADObject{ADProperties: ADProperties{NetBIOS: "CORP"}},
				expected: "CORP",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.ad.GetEffectiveDomain()
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("GetPrimaryIdentifier", func(t *testing.T) {
		tests := []struct {
			name     string
			ad       ADObject
			expected string
		}{
			{
				name:     "use SID when available",
				ad:       ADObject{ObjectID: "S-1-5-21-123456789-123456789-123456789-1001"},
				expected: "S-1-5-21-123456789-123456789-123456789-1001",
			},
			{
				name:     "use DN when SID not available",
				ad:       ADObject{ADProperties: ADProperties{DistinguishedName: "CN=User,DC=example,DC=local"}},
				expected: "CN=User,DC=example,DC=local",
			},
			{
				name:     "use SAMAccountName as fallback",
				ad:       ADObject{ADProperties: ADProperties{SAMAccountName: "user1"}},
				expected: "user1",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.ad.GetPrimaryIdentifier()
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("IsPrivileged", func(t *testing.T) {
		tests := []struct {
			name     string
			ad       ADObject
			expected bool
		}{
			{
				name:     "admin count indicates privileged",
				ad:       ADObject{ADProperties: ADProperties{AdminCount: true}},
				expected: true,
			},
			{
				name:     "sensitive flag indicates privileged",
				ad:       ADObject{ADProperties: ADProperties{Sensitive: true}},
				expected: true,
			},
			{
				name:     "unconstrained delegation indicates privileged",
				ad:       ADObject{ADProperties: ADProperties{UnconstrainedDelegation: true}},
				expected: true,
			},
			{
				name:     "trusted to auth indicates privileged",
				ad:       ADObject{ADProperties: ADProperties{TrustedToAuth: true}},
				expected: true,
			},
			{
				name:     "non-privileged object",
				ad:       ADObject{},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.ad.IsPrivileged()
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// Table-driven tests for complex scenarios
func TestADObject_ComplexScenarios(t *testing.T) {
	scenarios := []struct {
		name        string
		setupFunc   func() *ADObject
		testFunc    func(t *testing.T, ad *ADObject)
		description string
	}{
		{
			name: "domain admin user",
			setupFunc: func() *ADObject {
				ad := NewADObject("corp.com", "S-1-5-21-123456789-123456789-123456789-1001", NO_DISTINGUISHED_NAME, "user")
				ad.DistinguishedName = "CN=Administrator,CN=Users,DC=corp,DC=com"
				ad.SID = "S-1-5-21-123456789-123456789-123456789-500"
				ad.SAMAccountName = "Administrator"
				ad.AdminCount = true
				ad.Sensitive = true
				registry.CallHooks(&ad)
				return &ad
			},
			testFunc: func(t *testing.T, ad *ADObject) {
				assert.True(t, ad.IsClass("user"))
				assert.True(t, ad.IsInDomain("corp.com"))
				assert.Equal(t, "Administrator", ad.GetCommonName())
				assert.True(t, ad.AdminCount)
				assert.True(t, ad.Sensitive)
			},
			description: "Domain Administrator account should have elevated privileges",
		},
		{
			name: "computer with LAPS",
			setupFunc: func() *ADObject {
				ad := NewADObject("example.local", "S-1-5-21-123456789-123456789-123456789-1001", NO_DISTINGUISHED_NAME, "computer")
				ad.DistinguishedName = "CN=WORKSTATION01,OU=Computers,DC=example,DC=local"
				ad.SAMAccountName = "WORKSTATION01$"
				ad.HasLAPS = true
				registry.CallHooks(&ad)
				return &ad
			},
			testFunc: func(t *testing.T, ad *ADObject) {
				assert.True(t, ad.IsClass("computer"))
				assert.Equal(t, "Computers", ad.GetOU())
				assert.True(t, ad.HasLAPS)
				assert.True(t, strings.HasSuffix(ad.SAMAccountName, "$"))
			},
			description: "Computer with LAPS enabled",
		},
		{
			name: "service account with SPN",
			setupFunc: func() *ADObject {
				ad := NewADObject("example.local", "S-1-5-21-123456789-123456789-123456789-1001", NO_DISTINGUISHED_NAME, "user")
				ad.DistinguishedName = "CN=svc_sql,CN=Users,DC=example,DC=local"
				ad.SAMAccountName = "svc_sql"
				ad.HasSPN = true
				ad.TrustedToAuth = true
				registry.CallHooks(&ad)
				return &ad
			},
			testFunc: func(t *testing.T, ad *ADObject) {
				assert.True(t, ad.IsClass("user"))
				assert.True(t, ad.HasSPN)
				assert.True(t, ad.TrustedToAuth)
			},
			description: "Service account with constrained delegation",
		},
		{
			name: "nested group",
			setupFunc: func() *ADObject {
				ad := NewADObject("corp.com", "S-1-5-21-123456789-123456789-123456789-1001", NO_DISTINGUISHED_NAME, "group")
				ad.DistinguishedName = "CN=Finance Admins,OU=Groups,OU=Finance,DC=corp,DC=com"
				ad.SAMAccountName = "Finance Admins"
				registry.CallHooks(&ad)
				return &ad
			},
			testFunc: func(t *testing.T, ad *ADObject) {
				assert.True(t, ad.IsClass("group"))
				assert.Equal(t, "Groups", ad.GetOU())
				assert.Equal(t, "OU=Groups,OU=Finance,DC=corp,DC=com", ad.GetParentDN())
			},
			description: "Group in nested OU structure",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			ad := scenario.setupFunc()
			scenario.testFunc(t, ad)
		})
	}
}

// Test concurrent access patterns
func TestADObject_ConcurrentAccess(t *testing.T) {
	ad := NewADObject("example.local", "CN=Test,DC=example,DC=local", NO_DISTINGUISHED_NAME, "user")

	// Test concurrent reads
	t.Run("concurrent reads", func(t *testing.T) {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				_ = ad.GetCommonName()
				_ = ad.GetParentDN()
				_ = ad.GetOU()
				_ = ad.IsClass("user")
				_ = ad.IsInDomain("example.local")
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestADDomain_SeedModels(t *testing.T) {
	seed := NewADDomainSeed("example.local", "S-1-5-21-123456789-123456789-123456789-1001", "CN=example.local,DC=example,DC=local")
	seedModels := seed.SeedModels()

	assert.Equal(t, 1, len(seedModels))
	assert.Equal(t, &seed, seedModels[0])
	assert.Contains(t, seed.GetLabels(), SeedLabel)
}
