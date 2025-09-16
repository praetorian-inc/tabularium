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
		label             string
		expectedKey       string
		expectedClass     string
		expectedName      string
	}{
		{
			name:              "create user object",
			domain:            "example.local",
			objectID:          "S-1-5-21-123456789-123456789-123456789-1001",
			distinguishedName: "CN=John Doe,CN=Users,DC=example,DC=local",
			label:             ADUserLabel,
			expectedKey:       "#aduser#example.local#S-1-5-21-123456789-123456789-123456789-1001",
			expectedClass:     "user",
			expectedName:      "John Doe",
		},
		{
			name:              "create computer object",
			domain:            "corp.com",
			objectID:          "S-1-5-21-123456789-123456789-123456789-1002",
			distinguishedName: "CN=WORKSTATION01,CN=Computers,DC=corp,DC=com",
			label:             ADComputerLabel,
			expectedKey:       "#adcomputer#corp.com#S-1-5-21-123456789-123456789-123456789-1002",
			expectedClass:     "computer",
			expectedName:      "WORKSTATION01",
		},
		{
			name:              "create group object",
			domain:            "test.domain",
			objectID:          "S-1-5-21-123456789-123456789-123456789-1003",
			distinguishedName: "CN=Domain Admins,CN=Groups,DC=test,DC=domain",
			label:             ADGroupLabel,
			expectedKey:       "#adgroup#test.domain#S-1-5-21-123456789-123456789-123456789-1003",
			expectedClass:     "group",
			expectedName:      "Domain Admins",
		},
		{
			name:              "create OU object",
			domain:            "example.local",
			objectID:          "51FB8637-28BC-4816-9A51-984160B207FA",
			distinguishedName: "OU=Sales,DC=example,DC=local",
			label:             ADOULabel,
			expectedKey:       "#adou#example.local#51FB8637-28BC-4816-9A51-984160B207FA",
			expectedClass:     "ou",
			expectedName:      "",
		},
		{
			name:              "DN without CN prefix",
			domain:            "example.local",
			objectID:          "S-1-5-21-123456789-123456789-123456789-1005",
			distinguishedName: "DC=example,DC=local",
			label:             ADDomainLabel,
			expectedKey:       "#addomain#example.local#S-1-5-21-123456789-123456789-123456789-1005",
			expectedClass:     "domain",
			expectedName:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := NewADObject(tt.domain, tt.objectID, tt.distinguishedName, tt.label)

			assert.Equal(t, tt.domain, ad.Domain, "Domain should match")
			assert.Equal(t, tt.objectID, ad.ObjectID, "ObjectID should match")
			assert.Equal(t, tt.distinguishedName, ad.DistinguishedName, "DistinguishedName should match")
			assert.Equal(t, tt.expectedKey, ad.Key, "Key should be generated correctly")
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
			name:          "hook generates correct key and label",
			domain:        "TEST.LOCAL",
			objectID:      "S-1-5-21-123456789-123456789-123456789-1001",
			label:         "ADUser",
			expectedKey:   "#aduser#test.local#S-1-5-21-123456789-123456789-123456789-1001",
			expectedClass: "user",
		},
		{
			name:          "hook handles empty values - should fail",
			domain:        "",
			objectID:      "",
			label:         "",
			expectedKey:   "",
			expectedClass: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := NewADObject(tt.domain, tt.objectID, NO_DISTINGUISHED_NAME, tt.label)

			err := registry.CallHooks(&ad)

			if tt.name == "hook handles empty values - should fail" {
				require.Error(t, err, "Hook should fail when neither ObjectID nor DN is provided")
				assert.Contains(t, err.Error(), "requires either ObjectID or DistinguishedName")
			} else {
				require.NoError(t, err, "Hook should execute without error")
				assert.Equal(t, tt.expectedKey, ad.Key, "Hook should generate correct key")
				assert.Equal(t, tt.expectedClass, ad.Class, "Hook should set correct label")
			}
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
			name:       "different label",
			label:      ADComputerLabel,
			checkClass: "user",
			expected:   false,
		},
		{
			name:       "empty object label",
			label:      "",
			checkClass: "user",
			expected:   false,
		},
		{
			name:       "empty check label",
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
		assert.NotEqual(t, ad1.Domain, ad2.Domain, "Same DN in different domains should have different domains")
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

func TestADDomain_SeedModels(t *testing.T) {
	seed := NewADDomainSeed("example.local", "S-1-5-21-123456789-123456789-123456789-1001", "CN=example.local,DC=example,DC=local")
	seedModels := seed.SeedModels()

	assert.Equal(t, 1, len(seedModels))
	assert.Equal(t, &seed, seedModels[0])
	assert.Contains(t, seed.GetLabels(), SeedLabel)
}

// Test object creation failure when neither ObjectID nor DN is provided
func TestADObject_CreationFailure(t *testing.T) {
	tests := []struct {
		name              string
		domain            string
		objectID          string
		distinguishedName string
		expectError       bool
		description       string
	}{
		{
			name:              "should fail with empty ObjectID and empty DN",
			domain:            "example.local",
			objectID:          "",
			distinguishedName: "",
			expectError:       true,
			description:       "Object creation should fail when both ObjectID and DN are empty",
		},
		{
			name:              "should fail with empty ObjectID and blank DN constant",
			domain:            "example.local",
			objectID:          "",
			distinguishedName: "<blank>",
			expectError:       true,
			description:       "Object creation should fail when ObjectID is empty and DN is the test blank constant",
		},
		{
			name:              "should succeed with ObjectID",
			domain:            "example.local",
			objectID:          "S-1-5-21-123456789-123456789-123456789-1001",
			distinguishedName: "",
			expectError:       false,
			description:       "Object creation should succeed with ObjectID",
		},
		{
			name:              "should succeed with DN",
			domain:            "example.local",
			objectID:          "",
			distinguishedName: "CN=User,DC=example,DC=local",
			expectError:       false,
			description:       "Object creation should succeed with DN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := NewADObject(tt.domain, tt.objectID, tt.distinguishedName, ADUserLabel)
			err := registry.CallHooks(&ad)

			if tt.expectError {
				require.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), "requires either ObjectID or DistinguishedName")
			} else {
				require.NoError(t, err, tt.description)
				assert.True(t, ad.Valid(), "Created object should be valid")
			}
		})
	}
}

// Test DN validation (allowing DN as substitute for ObjectID)
func TestADObject_DNValidation(t *testing.T) {
	tests := []struct {
		name              string
		domain            string
		objectID          string
		distinguishedName string
		valid             bool
		description       string
	}{
		{
			name:              "valid with ObjectID only",
			domain:            "example.local",
			objectID:          "S-1-5-21-123456789-123456789-123456789-1001",
			distinguishedName: "",
			valid:             true,
			description:       "Valid with ObjectID and domain",
		},
		{
			name:              "valid with DN only",
			domain:            "example.local",
			objectID:          "",
			distinguishedName: "CN=User,DC=example,DC=local",
			valid:             true,
			description:       "Valid with DN and domain",
		},
		{
			name:              "valid with both ObjectID and DN",
			domain:            "example.local",
			objectID:          "S-1-5-21-123456789-123456789-123456789-1001",
			distinguishedName: "CN=User,DC=example,DC=local",
			valid:             true,
			description:       "Valid with both ObjectID and DN",
		},
		{
			name:              "invalid with neither ObjectID nor DN",
			domain:            "example.local",
			objectID:          "",
			distinguishedName: "",
			valid:             false,
			description:       "Invalid without ObjectID or DN",
		},
		{
			name:              "invalid without domain",
			domain:            "",
			objectID:          "S-1-5-21-123456789-123456789-123456789-1001",
			distinguishedName: "",
			valid:             false,
			description:       "Invalid without domain",
		},
		{
			name:              "invalid with DN but no domain",
			domain:            "",
			objectID:          "",
			distinguishedName: "CN=User,DC=example,DC=local",
			valid:             false,
			description:       "Invalid with DN but no domain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := NewADObject(tt.domain, tt.objectID, tt.distinguishedName, ADUserLabel)
			result := ad.Valid()
			assert.Equal(t, tt.valid, result, tt.description)
		})
	}
}

// Test extractCNFromDN helper function
func TestExtractCNFromDN(t *testing.T) {
	tests := []struct {
		name     string
		dn       string
		expected string
	}{
		{
			name:     "standard DN with CN first",
			dn:       "CN=John Doe,OU=Users,DC=example,DC=local",
			expected: "John Doe",
		},
		{
			name:     "DN with spaces around CN",
			dn:       "CN=Test User, OU=IT, DC=example, DC=local",
			expected: "Test User",
		},
		{
			name:     "DN with lowercase cn",
			dn:       "cn=service account,ou=Services,dc=example,dc=local",
			expected: "service account",
		},
		{
			name:     "DN with CN in middle",
			dn:       "OU=Users,CN=Middle User,DC=example,DC=local",
			expected: "Middle User",
		},
		{
			name:     "DN without CN",
			dn:       "OU=Organizational Unit,DC=example,DC=local",
			expected: "",
		},
		{
			name:     "empty DN",
			dn:       "",
			expected: "",
		},
		{
			name:     "DN with special characters in CN",
			dn:       "CN=User-Name.123,OU=Users,DC=example,DC=local",
			expected: "User-Name.123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCNFromDN(tt.dn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test key generation for DN-based objects
func TestADObject_DNKeyGeneration(t *testing.T) {
	tests := []struct {
		name              string
		domain            string
		objectID          string
		distinguishedName string
		expectedKey       string
		description       string
	}{
		{
			name:              "ObjectID takes precedence over DN",
			domain:            "example.local",
			objectID:          "S-1-5-21-123456789-123456789-123456789-1001",
			distinguishedName: "CN=User,DC=example,DC=local",
			expectedKey:       "#aduser#example.local#S-1-5-21-123456789-123456789-123456789-1001",
			description:       "ObjectID should be used when both are present",
		},
		{
			name:              "DN used when ObjectID is empty",
			domain:            "example.local",
			objectID:          "",
			distinguishedName: "CN=Test User,DC=example,DC=local",
			expectedKey:       "#aduser#example.local#Test User",
			description:       "DN CN should be used when ObjectID is empty",
		},
		{
			name:              "DN fallback for DN without CN",
			domain:            "example.local",
			objectID:          "",
			distinguishedName: "OU=OrganizationalUnit,DC=example,DC=local",
			expectedKey:       "#aduser#example.local#OU=OrganizationalUnit,DC=example,DC=local",
			description:       "Full DN should be used when CN cannot be extracted",
		},
		{
			name:              "DN truncation for long DN",
			domain:            "example.local",
			objectID:          "",
			distinguishedName: "OU=Very-Long-Organizational-Unit-Name-That-Exceeds-Normal-Length-Limits-For-Testing-Purposes,DC=example,DC=local",
			expectedKey:       "#aduser#example.local#OU=Very-Long-Organizational-Unit-Name-That-Exceeds-Normal-Length-Limits-For-Testing-Purposes,DC=exam",
			description:       "Long DN should be truncated when used as key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := NewADObject(tt.domain, tt.objectID, tt.distinguishedName, ADUserLabel)
			assert.Equal(t, tt.expectedKey, ad.Key, tt.description)
		})
	}
}

// Test reconciliation functionality
func TestADObject_Reconciliation(t *testing.T) {
	tests := []struct {
		name        string
		existing    ADObject
		visiting    ADObject
		shouldMatch bool
		description string
	}{
		{
			name: "reconcile ObjectID+DN with DN-only",
			existing: ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
				Label:    ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=UserTemplate,CN=Certificate Templates,CN=Public Key Services,CN=Services,CN=Configuration,DC=example,DC=local",
					Name:              "UserTemplate",
				},
			},
			visiting: ADObject{
				Domain:   "example.local",
				ObjectID: "",
				Label:    ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=UserTemplate,CN=Certificate Templates,CN=Public Key Services,CN=Services,CN=Configuration,DC=example,DC=local",
					Description:       "Template for user certificates",
				},
			},
			shouldMatch: true,
			description: "Objects with same domain, label, and DN should reconcile",
		},
		{
			name: "reconcile DN-only with ObjectID+DN",
			existing: ADObject{
				Domain:   "example.local",
				ObjectID: "",
				Label:    ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=UserTemplate,CN=Certificate Templates,CN=Public Key Services,CN=Services,CN=Configuration,DC=example,DC=local",
					Description:       "Template for user certificates",
				},
			},
			visiting: ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
				Label:    ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=UserTemplate,CN=Certificate Templates,CN=Public Key Services,CN=Services,CN=Configuration,DC=example,DC=local",
					Name:              "UserTemplate",
				},
			},
			shouldMatch: true,
			description: "Objects with same domain, label, and DN should reconcile (reverse case)",
		},
		{
			name: "no reconciliation for different domains",
			existing: ADObject{
				Domain:   "domain1.local",
				ObjectID: "",
				Label:    ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=UserTemplate,CN=Certificate Templates,CN=Public Key Services,CN=Services,CN=Configuration,DC=domain1,DC=local",
				},
			},
			visiting: ADObject{
				Domain:   "domain2.local",
				ObjectID: "",
				Label:    ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=UserTemplate,CN=Certificate Templates,CN=Public Key Services,CN=Services,CN=Configuration,DC=domain2,DC=local",
				},
			},
			shouldMatch: false,
			description: "Objects from different domains should not reconcile",
		},
		{
			name: "no reconciliation for different labels",
			existing: ADObject{
				Domain:   "example.local",
				ObjectID: "",
				Label:    ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=Template,DC=example,DC=local",
				},
			},
			visiting: ADObject{
				Domain:   "example.local",
				ObjectID: "",
				Label:    ADUserLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=Template,DC=example,DC=local",
				},
			},
			shouldMatch: false,
			description: "Objects with different labels should not reconcile",
		},
		{
			name: "no reconciliation for different DNs",
			existing: ADObject{
				Domain:   "example.local",
				ObjectID: "",
				Label:    ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=Template1,DC=example,DC=local",
				},
			},
			visiting: ADObject{
				Domain:   "example.local",
				ObjectID: "",
				Label:    ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=Template2,DC=example,DC=local",
				},
			},
			shouldMatch: false,
			description: "Objects with different DNs should not reconcile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize objects
			existing := tt.existing
			visiting := tt.visiting

			err := registry.CallHooks(&existing)
			require.NoError(t, err)

			err = registry.CallHooks(&visiting)
			require.NoError(t, err)

			// Test canReconcileWith
			result := existing.canReconcileWith(&visiting)
			assert.Equal(t, tt.shouldMatch, result, tt.description)

			// Test actual reconciliation via Visit
			if tt.shouldMatch {
				originalKey := existing.Key
				existing.Visit(&visiting)

				// If existing had no ObjectID but visiting did, existing should now have ObjectID
				if tt.existing.ObjectID == "" && tt.visiting.ObjectID != "" {
					assert.Equal(t, tt.visiting.ObjectID, existing.ObjectID, "ObjectID should be copied during reconciliation")
					// Key should be regenerated with ObjectID
					expectedKey := "#" + strings.ToLower(existing.Label) + "#" + existing.Domain + "#" + existing.ObjectID
					assert.Equal(t, expectedKey, existing.Key, "Key should be regenerated with ObjectID")
				} else {
					assert.Equal(t, originalKey, existing.Key, "Key should remain unchanged")
				}

				// Properties should be merged
				if tt.visiting.Description != "" {
					assert.Equal(t, tt.visiting.Description, existing.Description, "Description should be merged")
				}
				if tt.visiting.Name != "" {
					assert.Equal(t, tt.visiting.Name, existing.Name, "Name should be merged")
				}
			}
		})
	}
}

// Test canReconcileWith method directly
func TestADObject_CanReconcileWith(t *testing.T) {
	tests := []struct {
		name     string
		obj1     ADObject
		obj2     ADObject
		expected bool
	}{
		{
			name: "matching domain, label, and DN",
			obj1: ADObject{
				Domain: "example.local",
				Label:  ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=Template,DC=example,DC=local",
				},
			},
			obj2: ADObject{
				Domain: "example.local",
				Label:  ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=Template,DC=example,DC=local",
				},
			},
			expected: true,
		},
		{
			name: "case-insensitive DN matching",
			obj1: ADObject{
				Domain: "example.local",
				Label:  ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=Template,DC=example,DC=local",
				},
			},
			obj2: ADObject{
				Domain: "example.local",
				Label:  ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "cn=template,dc=example,dc=local",
				},
			},
			expected: true,
		},
		{
			name: "one with ObjectID, other with DN",
			obj1: ADObject{
				Domain:   "example.local",
				ObjectID: "S-1-5-21-123456789-123456789-123456789-1001",
				Label:    ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "",
				},
			},
			obj2: ADObject{
				Domain:   "example.local",
				ObjectID: "",
				Label:    ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=Template,DC=example,DC=local",
				},
			},
			expected: true,
		},
		{
			name: "different domains",
			obj1: ADObject{
				Domain: "domain1.local",
				Label:  ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=Template,DC=domain1,DC=local",
				},
			},
			obj2: ADObject{
				Domain: "domain2.local",
				Label:  ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=Template,DC=domain2,DC=local",
				},
			},
			expected: false,
		},
		{
			name: "different labels",
			obj1: ADObject{
				Domain: "example.local",
				Label:  ADCertTemplateLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=Template,DC=example,DC=local",
				},
			},
			obj2: ADObject{
				Domain: "example.local",
				Label:  ADUserLabel,
				ADProperties: ADProperties{
					DistinguishedName: "CN=Template,DC=example,DC=local",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.obj1.canReconcileWith(&tt.obj2)
			assert.Equal(t, tt.expected, result)
		})
	}
}
