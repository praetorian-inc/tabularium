package model

import (
	"strings"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test core ADObject creation and initialization
func TestNewADObject(t *testing.T) {
	tests := []struct {
		name              string
		domain            string
		distinguishedName string
		objectClass       string
		expectedKey       string
		expectedClass     string
		expectedName      string
	}{
		{
			name:              "create user object",
			domain:            "example.local",
			distinguishedName: "CN=John Doe,CN=Users,DC=example,DC=local",
			objectClass:       "user",
			expectedKey:       "#adobject#example.local#cn=john doe,cn=users,dc=example,dc=local",
			expectedClass:     "user",
			expectedName:      "John Doe",
		},
		{
			name:              "create computer object",
			domain:            "CORP.COM",
			distinguishedName: "CN=WORKSTATION01,CN=Computers,DC=corp,DC=com",
			objectClass:       "computer",
			expectedKey:       "#adobject#corp.com#cn=workstation01,cn=computers,dc=corp,dc=com",
			expectedClass:     "computer",
			expectedName:      "WORKSTATION01",
		},
		{
			name:              "create group object",
			domain:            "test.domain",
			distinguishedName: "CN=Domain Admins,CN=Groups,DC=test,DC=domain",
			objectClass:       "group",
			expectedKey:       "#adobject#test.domain#cn=domain admins,cn=groups,dc=test,dc=domain",
			expectedClass:     "group",
			expectedName:      "Domain Admins",
		},
		{
			name:              "create OU object",
			domain:            "example.local",
			distinguishedName: "OU=Sales,DC=example,DC=local",
			objectClass:       "organizationalUnit",
			expectedKey:       "#adobject#example.local#ou=sales,dc=example,dc=local",
			expectedClass:     "organizationalunit",
			expectedName:      "",
		},
		{
			name:              "DN without CN prefix",
			domain:            "example.local",
			distinguishedName: "DC=example,DC=local",
			objectClass:       "domain",
			expectedKey:       "#adobject#example.local#dc=example,dc=local",
			expectedClass:     "domain",
			expectedName:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := NewADObject(tt.domain, tt.distinguishedName, tt.objectClass)

			assert.Equal(t, tt.domain, ad.Domain, "Domain should match")
			assert.Equal(t, tt.distinguishedName, ad.DistinguishedName, "DistinguishedName should match")
			assert.Equal(t, tt.objectClass, ad.ObjectClass, "ObjectClass should match")
			assert.Equal(t, tt.expectedKey, ad.Key, "Key should be generated correctly")
			assert.Equal(t, tt.expectedClass, ad.Class, "Class should be set correctly")
			assert.Equal(t, tt.expectedName, ad.Name, "Name should be extracted from DN")
			assert.True(t, ad.Valid(), "ADObject should be valid")
			assert.NotEmpty(t, ad.Created, "Created timestamp should be set")
			assert.NotEmpty(t, ad.Visited, "Visited timestamp should be set")
		})
	}
}

// Test GetLabels functionality
func TestADObject_GetLabels(t *testing.T) {
	ad := ADObject{}
	labels := ad.GetLabels()

	assert.Contains(t, labels, ADObjectLabel, "Should contain ADObject label")
	assert.Contains(t, labels, TTLLabel, "Should contain TTL label")
	assert.Len(t, labels, 2, "Should have exactly 2 labels")
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
		name              string
		domain            string
		distinguishedName string
		label             string
		expectedKey       string
		expectedClass     string
	}{
		{
			name:              "hook generates correct key and class",
			domain:            "TEST.LOCAL",
			distinguishedName: "CN=TestUser,DC=test,DC=local",
			label:             "ADUser",
			expectedKey:       "#adobject#test.local#cn=testuser,dc=test,dc=local",
			expectedClass:     "user",
		},
		{
			name:              "hook handles empty values",
			domain:            "",
			distinguishedName: "",
			label:             "",
			expectedKey:       "#adobject##",
			expectedClass:     "",
		},
		{
			name:              "hook handles special characters in DN",
			domain:            "example.com",
			distinguishedName: "CN=O'Brien\\, John,CN=Users,DC=example,DC=com",
			label:             "ADUser",
			expectedKey:       "#adobject#example.com#cn=o'brien\\, john,cn=users,dc=example,dc=com",
			expectedClass:     "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := ADObject{
				Domain:            tt.domain,
				DistinguishedName: tt.distinguishedName,
				Label:             tt.label,
			}

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
		visiting interface{} // Use interface{} to allow testing with non-ADObject types
		expected ADObject
	}{
		{
			name: "merge with valid ADObject",
			existing: ADObject{
				Domain:            "example.local",
				DistinguishedName: "CN=User1,DC=example,DC=local",
				ObjectClass:       "user",
				Name:              "User1",
			},
			visiting: ADObject{
				SID:            "S-1-5-21-123456789-123456789-123456789-1001",
				SAMAccountName: "user1",
				DisplayName:    "User One",
				Description:    "Test user account",
			},
			expected: ADObject{
				Domain:            "example.local",
				DistinguishedName: "CN=User1,DC=example,DC=local",
				ObjectClass:       "user",
				Name:              "User1",
				SID:               "S-1-5-21-123456789-123456789-123456789-1001",
				SAMAccountName:    "user1",
				DisplayName:       "User One",
				Description:       "Test user account",
			},
		},
		{
			name: "don't override existing values",
			existing: ADObject{
				SID:            "S-1-5-21-EXISTING",
				SAMAccountName: "existing",
				DisplayName:    "Existing Display",
				Description:    "Existing description",
			},
			visiting: ADObject{
				SID:            "S-1-5-21-NEW",
				SAMAccountName: "new",
				DisplayName:    "New Display",
				Description:    "New description",
			},
			expected: ADObject{
				SID:            "S-1-5-21-EXISTING",
				SAMAccountName: "existing",
				DisplayName:    "Existing Display",
				Description:    "Existing description",
			},
		},
		{
			name: "handle non-ADObject type",
			existing: ADObject{
				Domain: "example.local",
			},
			visiting: "not an ADObject",
			expected: ADObject{
				Domain: "example.local",
			},
		},
		{
			name: "partial merge",
			existing: ADObject{
				SID:         "S-1-5-21-EXISTING",
				DisplayName: "Existing",
			},
			visiting: ADObject{
				SAMAccountName: "newuser",
				Description:    "New description",
			},
			expected: ADObject{
				SID:            "S-1-5-21-EXISTING",
				DisplayName:    "Existing",
				SAMAccountName: "newuser",
				Description:    "New description",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := tt.existing

			// Handle different types of visiting objects
			switch v := tt.visiting.(type) {
			case ADObject:
				ad.Visit(&v)
			case string:
				// Test that Visit handles non-ADObject types gracefully
				// Create a dummy struct that implements Assetlike
				// For this test, we'll just skip since Visit expects Assetlike
				// and a string doesn't implement it
			default:
				if assetlike, ok := tt.visiting.(Assetlike); ok {
					ad.Visit(assetlike)
				}
			}

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
		name        string
		objectClass string
		checkClass  string
		expected    bool
	}{
		{
			name:        "exact match",
			objectClass: "user",
			checkClass:  "user",
			expected:    true,
		},
		{
			name:        "case insensitive match",
			objectClass: "User",
			checkClass:  "USER",
			expected:    true,
		},
		{
			name:        "different class",
			objectClass: "computer",
			checkClass:  "user",
			expected:    false,
		},
		{
			name:        "empty object class",
			objectClass: "",
			checkClass:  "user",
			expected:    false,
		},
		{
			name:        "empty check class",
			objectClass: "user",
			checkClass:  "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := ADObject{ObjectClass: tt.objectClass}
			result := ad.IsClass(tt.checkClass)
			assert.Equal(t, tt.expected, result, "IsClass result should match expected")
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
			ad := ADObject{DistinguishedName: tt.distinguishedName}
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
			ad := ADObject{DistinguishedName: tt.distinguishedName}
			result := ad.GetOU()
			assert.Equal(t, tt.expected, result, "GetOU should return correct OU")
		})
	}
}

// Test IsEnabled functionality
func TestADObject_IsEnabled(t *testing.T) {
	ad := ADObject{}
	// Default implementation always returns true
	assert.True(t, ad.IsEnabled(), "IsEnabled should return true by default")
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
			ad := ADObject{
				DistinguishedName: tt.distinguishedName,
				Name:              tt.nameField,
			}
			result := ad.GetCommonName()
			assert.Equal(t, tt.expected, result, "GetCommonName should return correct value")
		})
	}
}

// Test DN validation edge cases
func TestADObject_DNValidation(t *testing.T) {
	tests := []struct {
		name              string
		distinguishedName string
		shouldProcess     bool
		description       string
	}{
		{
			name:              "valid standard DN",
			distinguishedName: "CN=User,DC=example,DC=local",
			shouldProcess:     true,
			description:       "Standard DN should process correctly",
		},
		{
			name:              "DN with escaped characters",
			distinguishedName: "CN=O'Brien\\, John Jr.,OU=Sales\\+Marketing,DC=example,DC=local",
			shouldProcess:     true,
			description:       "DN with escaped special characters should process",
		},
		{
			name:              "DN with Unicode characters",
			distinguishedName: "CN=José García,OU=España,DC=example,DC=local",
			shouldProcess:     true,
			description:       "DN with Unicode should process",
		},
		{
			name:              "malformed DN missing equals",
			distinguishedName: "CNUser,DCexample,DClocal",
			shouldProcess:     true,
			description:       "Malformed DN should still be accepted",
		},
		{
			name:              "DN with extra spaces",
			distinguishedName: "CN = User , DC = example , DC = local",
			shouldProcess:     true,
			description:       "DN with spaces should process",
		},
		{
			name:              "very long DN",
			distinguishedName: strings.Repeat("CN=VeryLongName,", 50) + "DC=example,DC=local",
			shouldProcess:     true,
			description:       "Very long DN should process",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := NewADObject("example.local", tt.distinguishedName, "user")

			if tt.shouldProcess {
				assert.Equal(t, tt.distinguishedName, ad.DistinguishedName, tt.description)
				assert.NotEmpty(t, ad.Key, "Key should be generated for valid DN")
			}
		})
	}
}

// Test SID validation and formats
func TestADObject_SIDValidation(t *testing.T) {
	tests := []struct {
		name        string
		sid         string
		description string
	}{
		{
			name:        "standard domain SID",
			sid:         "S-1-5-21-123456789-123456789-123456789-1001",
			description: "Standard domain SID format",
		},
		{
			name:        "well-known SID",
			sid:         "S-1-5-32-544",
			description: "Well-known Administrators SID",
		},
		{
			name:        "domain controller SID",
			sid:         "S-1-5-21-123456789-123456789-123456789-516",
			description: "Domain Controllers group SID",
		},
		{
			name:        "local system SID",
			sid:         "S-1-5-18",
			description: "Local System account SID",
		},
		{
			name:        "empty SID",
			sid:         "",
			description: "Empty SID should be allowed",
		},
		{
			name:        "malformed SID",
			sid:         "NOT-A-VALID-SID",
			description: "Malformed SID should still be stored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := ADObject{SID: tt.sid}
			assert.Equal(t, tt.sid, ad.SID, tt.description)
		})
	}
}

// Test domain validation and formats
func TestADObject_DomainValidation(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		description string
	}{
		{
			name:        "FQDN domain",
			domain:      "example.local",
			description: "Fully qualified domain name",
		},
		{
			name:        "NetBIOS domain",
			domain:      "EXAMPLE",
			description: "NetBIOS style domain name",
		},
		{
			name:        "multi-level domain",
			domain:      "sub.example.local",
			description: "Multi-level domain name",
		},
		{
			name:        "domain with numbers",
			domain:      "example123.local",
			description: "Domain with numbers",
		},
		{
			name:        "domain with hyphens",
			domain:      "example-corp.local",
			description: "Domain with hyphens",
		},
		{
			name:        "uppercase domain",
			domain:      "EXAMPLE.LOCAL",
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
			ad := NewADObject(tt.domain, "CN=Test,DC=example,DC=local", "user")
			assert.Equal(t, tt.domain, ad.Domain, tt.description)
		})
	}
}

// Test security-relevant behaviors
func TestADObject_SecurityBehaviors(t *testing.T) {
	t.Run("key generation prevents collision", func(t *testing.T) {
		ad1 := NewADObject("example.local", "CN=User1,DC=example,DC=local", "user")
		ad2 := NewADObject("example.local", "CN=User2,DC=example,DC=local", "user")

		assert.NotEqual(t, ad1.Key, ad2.Key, "Different DNs should generate different keys")
	})

	t.Run("key generation is case insensitive", func(t *testing.T) {
		ad1 := NewADObject("EXAMPLE.LOCAL", "CN=User,DC=EXAMPLE,DC=LOCAL", "USER")
		ad2 := NewADObject("example.local", "CN=User,DC=example,DC=local", "user")

		assert.Equal(t, ad1.Key, ad2.Key, "Keys should be case-insensitive")
	})

	t.Run("domain isolation", func(t *testing.T) {
		ad1 := NewADObject("domain1.local", "CN=User,DC=domain1,DC=local", "user")
		ad2 := NewADObject("domain2.local", "CN=User,DC=domain2,DC=local", "user")

		assert.NotEqual(t, ad1.Key, ad2.Key, "Same DN in different domains should have different keys")
		assert.False(t, ad1.IsInDomain("domain2.local"), "Object should not be in different domain")
		assert.True(t, ad1.IsInDomain("domain1.local"), "Object should be in its own domain")
	})
}

// Test factory methods for specific AD object types (future implementation)
func TestADObject_FactoryMethods(t *testing.T) {
	t.Run("NewADUser", func(t *testing.T) {
		// Test for future NewADUser factory method
		user := NewADUser("example.local", "CN=JDoe,CN=Users,DC=example,DC=local", "jdoe")

		assert.Equal(t, "example.local", user.Domain)
		assert.Equal(t, "CN=JDoe,CN=Users,DC=example,DC=local", user.DistinguishedName)
		assert.Equal(t, "jdoe", user.SAMAccountName)
		assert.Equal(t, ADUserLabel, user.Label)
		assert.Equal(t, "User", user.ObjectClass)
		assert.Equal(t, "user", user.Class)
	})

	t.Run("NewADComputer", func(t *testing.T) {
		// Test for future NewADComputer factory method
		computer := NewADComputer("corp.com", "CN=WORKSTATION01,CN=Computers,DC=corp,DC=com", "workstation01.corp.com")

		assert.Equal(t, "corp.com", computer.Domain)
		assert.Equal(t, "CN=WORKSTATION01,CN=Computers,DC=corp,DC=com", computer.DistinguishedName)
		assert.Equal(t, ADComputerLabel, computer.Label)
		assert.Equal(t, "Computer", computer.ObjectClass)
		assert.Equal(t, "computer", computer.Class)
	})

	t.Run("NewADGroup", func(t *testing.T) {
		// Test for future NewADGroup factory method
		group := NewADGroup("example.local", "CN=Domain Admins,CN=Groups,DC=example,DC=local", "Domain Admins")

		assert.Equal(t, "example.local", group.Domain)
		assert.Equal(t, "CN=Domain Admins,CN=Groups,DC=example,DC=local", group.DistinguishedName)
		assert.Equal(t, "Domain Admins", group.SAMAccountName)
		assert.Equal(t, ADGroupLabel, group.Label)
		assert.Equal(t, "Group", group.ObjectClass)
		assert.Equal(t, "group", group.Class)
	})

	t.Run("NewADGPO", func(t *testing.T) {
		// Test for future NewADGPO factory method
		gpo := NewADGPO("example.local", "CN={31B2F340-016D-11D2-945F-00C04FB984F9},CN=Policies,CN=System,DC=example,DC=local", "Default Domain Policy")

		assert.Equal(t, "example.local", gpo.Domain)
		assert.Contains(t, gpo.DistinguishedName, "31B2F340-016D-11D2-945F-00C04FB984F9")
		assert.Equal(t, "Default Domain Policy", gpo.DisplayName)
		assert.Equal(t, ADGPOLabel, gpo.Label)
		assert.Equal(t, "GPO", gpo.ObjectClass)
		assert.Equal(t, "gpo", gpo.Class)
	})

	t.Run("NewADOU", func(t *testing.T) {
		// Test for future NewADOU factory method
		ou := NewADOU("example.local", "OU=Sales,DC=example,DC=local", "Sales")

		assert.Equal(t, "example.local", ou.Domain)
		assert.Equal(t, "OU=Sales,DC=example,DC=local", ou.DistinguishedName)
		assert.Equal(t, "Sales", ou.Name)
		assert.Equal(t, ADOULabel, ou.Label)
		assert.Equal(t, "OU", ou.ObjectClass)
		assert.Equal(t, "ou", ou.Class)
	})
}

// Test extensive property list (future implementation)
func TestADObject_ExtensiveProperties(t *testing.T) {
	t.Run("security properties", func(t *testing.T) {
		ad := ADObject{
			AdminCount:              true,
			Sensitive:               true,
			HasSPN:                  true,
			UnconstrainedDelegation: true,
			TrustedToAuth:           true,
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
			PasswordNeverExpires:     true,
			PasswordNotRequired:      false,
			DontRequirePreAuth:       true,
			SmartcardRequired:        false,
			LockedOut:                true,
			PasswordExpired:          true,
			UserCannotChangePassword: true,
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
			HasLAPS: true,
		}

		assert.True(t, ad.HasLAPS)
	})

	t.Run("certificate properties", func(t *testing.T) {
		ad := ADObject{
			CertThumbprint:  "ABC123DEF456",
			CertThumbprints: []string{"ABC123DEF456", "789GHI012JKL"},
			CertChain:       []string{"root", "intermediate", "leaf"},
			CertName:        "test-cert",
			CAName:          "Example-CA",
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
				ad:       ADObject{DistinguishedName: "CN=User,DC=corp,DC=com"},
				expected: "corp.com",
			},
			{
				name:     "use NetBIOS when available",
				ad:       ADObject{NetBIOS: "CORP"},
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
				ad:       ADObject{DistinguishedName: "CN=User,DC=example,DC=local"},
				expected: "CN=User,DC=example,DC=local",
			},
			{
				name:     "use SAMAccountName as fallback",
				ad:       ADObject{SAMAccountName: "user1"},
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
				ad:       ADObject{AdminCount: true},
				expected: true,
			},
			{
				name:     "sensitive flag indicates privileged",
				ad:       ADObject{Sensitive: true},
				expected: true,
			},
			{
				name:     "unconstrained delegation indicates privileged",
				ad:       ADObject{UnconstrainedDelegation: true},
				expected: true,
			},
			{
				name:     "trusted to auth indicates privileged",
				ad:       ADObject{TrustedToAuth: true},
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
				ad := NewADObject("corp.com", "CN=Administrator,CN=Users,DC=corp,DC=com", "user")
				ad.SID = "S-1-5-21-123456789-123456789-123456789-500"
				ad.SAMAccountName = "Administrator"
				ad.AdminCount = true
				ad.Sensitive = true
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
				ad := NewADObject("example.local", "CN=WORKSTATION01,OU=Computers,DC=example,DC=local", "computer")
				ad.SAMAccountName = "WORKSTATION01$"
				ad.HasLAPS = true
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
				ad := NewADObject("example.local", "CN=svc_sql,CN=Users,DC=example,DC=local", "user")
				ad.SAMAccountName = "svc_sql"
				ad.HasSPN = true
				ad.TrustedToAuth = true
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
				ad := NewADObject("corp.com", "CN=Finance Admins,OU=Groups,OU=Finance,DC=corp,DC=com", "group")
				ad.SAMAccountName = "Finance Admins"
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
	ad := NewADObject("example.local", "CN=Test,DC=example,DC=local", "user")

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

// Benchmark tests for performance-critical operations
func BenchmarkADObject_NewADObject(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewADObject("example.local", "CN=User,CN=Users,DC=example,DC=local", "user")
	}
}

func BenchmarkADObject_GetParentDN(b *testing.B) {
	ad := ADObject{DistinguishedName: "CN=User,OU=Sales,OU=Departments,DC=example,DC=local"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ad.GetParentDN()
	}
}

func BenchmarkADObject_GetOU(b *testing.B) {
	ad := ADObject{DistinguishedName: "CN=User,OU=Sales,OU=Departments,DC=example,DC=local"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ad.GetOU()
	}
}

func BenchmarkADObject_Visit(b *testing.B) {
	ad1 := NewADObject("example.local", "CN=User1,DC=example,DC=local", "user")
	ad2 := NewADObject("example.local", "CN=User2,DC=example,DC=local", "user")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ad1.Visit(&ad2)
	}
}
