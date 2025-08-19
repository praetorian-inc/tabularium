package model

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// AD Object Type Label constants
const (
	ADObjectLabel         = "ADObject"
	ADUserLabel           = "ADUser"
	ADComputerLabel       = "ADComputer"
	ADGroupLabel          = "ADGroup"
	ADGPOLabel            = "ADGPO"
	ADOULabel             = "ADOU"
	ADContainerLabel      = "ADContainer"
	ADDomainLabel         = "ADDomain"
	ADLocalGroupLabel     = "ADLocalGroup"
	ADLocalUserLabel      = "ADLocalUser"
	ADAIACALabel          = "ADAIACA"
	ADRootCALabel         = "ADRootCA"
	ADEnterpriseCALabel   = "ADEnterpriseCA"
	ADNTAuthStoreLabel    = "ADNTAuthStore"
	ADCertTemplateLabel   = "ADCertTemplate"
	ADIssuancePolicyLabel = "ADIssuancePolicy"
)

var ADLabels = map[string]string{
	"adobject":         ADObjectLabel,
	"aduser":           ADUserLabel,
	"adcomputer":       ADComputerLabel,
	"adgroup":          ADGroupLabel,
	"adgpo":            ADGPOLabel,
	"adou":             ADOULabel,
	"adcontainer":      ADContainerLabel,
	"addomain":         ADDomainLabel,
	"adlocalgroup":     ADLocalGroupLabel,
	"adlocaluser":      ADLocalUserLabel,
	"adaiaca":          ADAIACALabel,
	"adrootca":         ADRootCALabel,
	"adenterpriseca":   ADEnterpriseCALabel,
	"adntauthstore":    ADNTAuthStoreLabel,
	"adcerttemplate":   ADCertTemplateLabel,
	"adissuancepolicy": ADIssuancePolicyLabel,
}

func GetADLabel(label string) string {
	label = strings.ToLower(label)
	check1 := ADLabels[label]
	if check1 != "" {
		return check1
	}

	label = "ad" + label
	return ADLabels[label]
}

func init() {
	labels := []string{}
	for _, label := range ADLabels {
		labels = append(labels, label)
	}
	registry.Registry.MustRegisterModel(&ADObject{}, labels...)
}

var (
	adObjectKeyPattern = regexp.MustCompile(`(?i)^#ad[a-z]+#[^#]+#[A-FS0-9-]+$`)
)

type ADObject struct {
	BaseAsset
	Label    string `neo4j:"label" json:"label" desc:"Label of the object." example:"user"`
	Domain   string `neo4j:"domain" json:"domain" desc:"AD domain this object belongs to." example:"example.local"`
	ObjectID string `neo4j:"objectid" json:"objectid" desc:"Object identifier." example:"S-1-5-21-123456789-123456789-123456789-1001"`
	SID      string `neo4j:"sid,omitempty" json:"sid,omitempty" desc:"Security identifier." example:"S-1-5-21-123456789-123456789-123456789-1001"`
	ADProperties
}

func (ad *ADObject) GetLabels() []string {
	labels := []string{ADObjectLabel, TTLLabel}
	if ad.Label != "" {
		labels = append(labels, ad.Label)
	}
	return labels
}

func (ad *ADObject) Valid() bool {
	hasObjectID := ad.ObjectID != ""
	hasDomain := ad.Domain != ""
	hasDistinguishedName := ad.DistinguishedName != ""
	keyMatches := adObjectKeyPattern.MatchString(ad.Key)

	return hasObjectID && hasDomain && hasDistinguishedName && keyMatches
}

func (ad *ADObject) Group() string {
	return ad.Domain
}

func (ad *ADObject) Identifier() string {
	return ad.DistinguishedName
}

func (ad *ADObject) Visit(o Assetlike) {
	other, ok := o.(*ADObject)
	if !ok {
		return
	}

	if ad.Key != other.Key {
		return
	}

	ad.ADProperties.Visit(other.ADProperties)

	ad.BaseAsset.Visit(other)

}

// IsClass checks if the AD object is of the specified object class
func (ad *ADObject) IsClass(objectClass string) bool {
	return strings.EqualFold(ad.ObjectClass, objectClass) || strings.EqualFold("adobject", objectClass)
}

// IsInDomain checks if the AD object belongs to the specified domain
func (ad *ADObject) IsInDomain(domain string) bool {
	return strings.EqualFold(ad.Domain, domain)
}

// GetParentDN extracts the parent distinguished name from the full DN
func (ad *ADObject) GetParentDN() string {
	if ad.DistinguishedName == "" {
		return ""
	}

	// Find the first non-escaped comma to get the parent DN
	// Handle escaped commas in CNs like "CN=O'Brien\, John"
	dn := ad.DistinguishedName
	for i := 0; i < len(dn); i++ {
		if dn[i] == ',' {
			// Check if this comma is escaped
			if i > 0 && dn[i-1] == '\\' {
				continue // Skip escaped comma
			}
			// Found non-escaped comma, return parent DN
			return strings.TrimSpace(dn[i+1:])
		}
	}

	return ""
}

// GetOU extracts the organizational unit from the distinguished name
func (ad *ADObject) GetOU() string {
	parentDN := ad.GetParentDN()
	if parentDN == "" {
		return ""
	}

	// Look for OU= in the parent DN
	parts := strings.Split(parentDN, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToUpper(part), "OU=") {
			return part[3:] // Remove "OU=" prefix
		}
	}

	return ""
}

// IsEnabled checks if the account is enabled based on common patterns
// This is a basic implementation that can be overridden by specific AD object types
func (ad *ADObject) IsEnabled() bool {
	// Default assumption is that objects are enabled unless specified otherwise
	// Specific AD object types should override this method with proper logic
	return true
}

// GetCommonName extracts the CN value from the distinguished name
func (ad *ADObject) GetCommonName() string {
	if ad.Name != "" {
		return ad.Name
	}

	// Extract CN= value from the beginning of the DN
	if strings.HasPrefix(strings.ToUpper(ad.DistinguishedName), "CN=") {
		// Find the first non-escaped comma
		dn := ad.DistinguishedName
		prefix := "CN="
		for i := len(prefix); i < len(dn); i++ {
			if dn[i] == ',' {
				if i > 0 && dn[i-1] == '\\' {
					continue // Skip escaped comma
				}
				return dn[len(prefix):i]
			}
		}
		return dn[len(prefix):]
	}

	return ""
}

// GetEffectiveDomain returns the effective domain for the object
func (ad *ADObject) GetEffectiveDomain() string {
	if ad.Domain != "" {
		return ad.Domain
	}
	if ad.NetBIOS != "" {
		return ad.NetBIOS
	}
	// Extract from DN
	if ad.DistinguishedName != "" {
		parts := strings.Split(ad.DistinguishedName, ",")
		var dcParts []string
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(strings.ToUpper(part), "DC=") {
				dcParts = append(dcParts, part[3:])
			}
		}
		if len(dcParts) > 0 {
			return strings.Join(dcParts, ".")
		}
	}
	return ""
}

// GetPrimaryIdentifier returns the primary identifier for the object
func (ad *ADObject) GetPrimaryIdentifier() string {
	if ad.ObjectID != "" {
		return ad.ObjectID
	}
	if ad.DistinguishedName != "" {
		return ad.DistinguishedName
	}
	if ad.SAMAccountName != "" {
		return ad.SAMAccountName
	}
	return ""
}

// IsPrivileged checks if the object has elevated privileges
func (ad *ADObject) IsPrivileged() bool {
	return ad.AdminCount || ad.Sensitive || ad.UnconstrainedDelegation || ad.TrustedToAuth
}

func (ad *ADObject) Attribute(name, value string) Attribute {
	attr := NewAttribute(name, value, ad)
	return attr
}

func (ad *ADObject) Seed() Seed {
	return ad.BaseAsset.Seed()
}

func (ad *ADObject) WithStatus(status string) Target {
	ret := *ad
	ret.Status = status
	return &ret
}

func (ad *ADObject) Defaulted() {
	ad.BaseAsset.Defaulted()
}

func (ad *ADObject) GetHooks() []registry.Hook {
	return []registry.Hook{
		useGroupAndIdentifier(ad, &ad.Domain, &ad.ObjectID),
		{
			Call: func() error {
				ad.Label = GetADLabel(ad.Label)
				if ad.Label == "" {
					ad.Label = ADObjectLabel
				}

				ad.Domain = strings.ToLower(ad.Domain)
				ad.ObjectID = strings.ToUpper(ad.ObjectID)

				ad.Key = fmt.Sprintf("#%s#%s#%s", strings.ToLower(ad.Label), ad.Domain, ad.ObjectID)

				ad.ObjectClass = strings.TrimPrefix(ad.Label, "AD")
				ad.Class = strings.ToLower(ad.ObjectClass)

				ad.Name = ad.GetCommonName()

				if strings.HasPrefix(ad.ObjectID, "S-") {
					ad.SID = ad.ObjectID
				}

				return nil
			},
		},
		setGroupAndIdentifier(ad, &ad.Domain, &ad.ObjectID),
	}
}

// NewADObject creates a new ADObject with the specified domain, distinguished name, and object class
func NewADObject(domain, objectID, distinguishedName, objectClass string) ADObject {
	ad := ADObject{
		Domain:   domain,
		ObjectID: objectID,
		Label:    objectClass,
		ADProperties: ADProperties{
			DistinguishedName: distinguishedName,
		},
	}

	ad.Defaulted()
	registry.CallHooks(&ad)

	return ad
}

// NewADUser creates a new AD User object
func NewADUser(domain, objectID, distinguishedName string) ADObject {
	return NewADObject(domain, objectID, distinguishedName, ADUserLabel)
}

// NewADComputer creates a new AD Computer object
func NewADComputer(domain, objectID, distinguishedName string) ADObject {
	return NewADObject(domain, objectID, distinguishedName, ADComputerLabel)
}

// NewADGroup creates a new AD Group object
func NewADGroup(domain, objectID, distinguishedName string) ADObject {
	return NewADObject(domain, objectID, distinguishedName, ADGroupLabel)
}

// NewADGPO creates a new AD GPO object
func NewADGPO(domain, objectID, distinguishedName string) ADObject {
	return NewADObject(domain, objectID, distinguishedName, ADGPOLabel)
}

// NewADOU creates a new AD OU object
func NewADOU(domain, objectID, distinguishedName string) ADObject {
	return NewADObject(domain, objectID, distinguishedName, ADOULabel)
}

// GetDescription returns a description for the ADObject model.
func (ad *ADObject) GetDescription() string {
	return "Represents an Active Directory object with properties and organizational unit information."
}

type ADProperties struct {

	// Core AD Properties
	DistinguishedName string `neo4j:"distinguishedName" json:"distinguishedName" desc:"Full distinguished name in AD." example:"CN=John Doe,CN=Users,DC=example,DC=local"`
	ObjectClass       string `neo4j:"objectClass" json:"objectClass" desc:"AD object class." example:"user"`
	Name              string `neo4j:"name" json:"name" desc:"Common name of the object." example:"John Doe"`
	DisplayName       string `neo4j:"displayName,omitempty" json:"displayName,omitempty" desc:"Display name of the object." example:"John Doe"`
	Description       string `neo4j:"description,omitempty" json:"description,omitempty" desc:"Description of the object." example:"User account for John Doe"`

	// Extended Identity Properties
	DomainSID      string `neo4j:"domainsid,omitempty" json:"domainsid,omitempty" desc:"Domain SID." example:"S-1-5-21-123456789-123456789-123456789"`
	SAMAccountName string `neo4j:"samaccountname,omitempty" json:"samaccountname,omitempty" desc:"SAM account name (lowercase field)." example:"jdoe"`
	ObjectGUID     string `neo4j:"objectguid,omitempty" json:"objectguid,omitempty" desc:"Object GUID." example:"12345678-1234-1234-1234-123456789012"`
	NetBIOS        string `neo4j:"netbios,omitempty" json:"netbios,omitempty" desc:"NetBIOS domain name." example:"CORP"`

	// Security Properties
	AdminCount              bool     `neo4j:"admincount,omitempty" json:"admincount,omitempty" desc:"AdminSDHolder protection flag." example:"true"`
	Sensitive               bool     `neo4j:"sensitive,omitempty" json:"sensitive,omitempty" desc:"Account is sensitive and cannot be delegated." example:"true"`
	HasSPN                  bool     `neo4j:"hasspn,omitempty" json:"hasspn,omitempty" desc:"Has Service Principal Name." example:"true"`
	UnconstrainedDelegation bool     `neo4j:"unconstraineddelegation,omitempty" json:"unconstraineddelegation,omitempty" desc:"Trusted for unconstrained delegation." example:"false"`
	TrustedToAuth           bool     `neo4j:"trustedtoauth,omitempty" json:"trustedtoauth,omitempty" desc:"Trusted for authentication delegation." example:"false"`
	AllowedToDelegate       []string `neo4j:"allowedtodelegate,omitempty" json:"allowedtodelegate,omitempty" desc:"Services allowed to delegate." example:"[\"HTTP/server.example.com\"]"`

	// Account Properties
	Enabled                  bool `neo4j:"enabled,omitempty" json:"enabled,omitempty" desc:"Account is enabled." example:"true"`
	PasswordNeverExpires     bool `neo4j:"pwdneverexpires,omitempty" json:"pwdneverexpires,omitempty" desc:"Password never expires." example:"true"`
	PasswordNotRequired      bool `neo4j:"passwordnotreqd,omitempty" json:"passwordnotreqd,omitempty" desc:"Password not required." example:"false"`
	DontRequirePreAuth       bool `neo4j:"dontreqpreauth,omitempty" json:"dontreqpreauth,omitempty" desc:"Pre-authentication not required (AS-REP roastable)." example:"false"`
	SmartcardRequired        bool `neo4j:"smartcardrequired,omitempty" json:"smartcardrequired,omitempty" desc:"Smart card required for interactive logon." example:"false"`
	LockedOut                bool `neo4j:"lockedout,omitempty" json:"lockedout,omitempty" desc:"Account is locked out." example:"false"`
	PasswordExpired          bool `neo4j:"passwordexpired,omitempty" json:"passwordexpired,omitempty" desc:"Password has expired." example:"false"`
	UserCannotChangePassword bool `neo4j:"passwordcantchange,omitempty" json:"passwordcantchange,omitempty" desc:"User cannot change password." example:"false"`
	IsDeleted                bool `neo4j:"isdeleted,omitempty" json:"isdeleted,omitempty" desc:"Object is deleted." example:"false"`

	// Time Properties
	LastLogon          int64 `neo4j:"lastlogon,omitempty" json:"lastlogon,omitempty" desc:"Last logon timestamp." example:"1698408000"`
	LastLogonTimestamp int64 `neo4j:"lastlogontimestamp,omitempty" json:"lastlogontimestamp,omitempty" desc:"Last logon timestamp (replicated)." example:"1698408000"`
	PasswordLastSet    int64 `neo4j:"pwdlastset,omitempty" json:"pwdlastset,omitempty" desc:"Password last set timestamp." example:"1698408000"`
	WhenCreated        int64 `neo4j:"whencreated,omitempty" json:"whencreated,omitempty" desc:"Creation timestamp." example:"1698408000"`

	// Group Properties
	GroupScope string `neo4j:"groupscope,omitempty" json:"groupscope,omitempty" desc:"Group scope (Global, Universal, DomainLocal)." example:"Global"`
	GroupType  string `neo4j:"grouptype,omitempty" json:"grouptype,omitempty" desc:"Group type." example:"Security"`

	// Computer Properties
	DNSHostname                string   `neo4j:"dnshostname,omitempty" json:"dnshostname,omitempty" desc:"DNS hostname of the computer." example:"workstation01.example.com"`
	OperatingSystem            string   `neo4j:"operatingsystem,omitempty" json:"operatingsystem,omitempty" desc:"Operating system." example:"Windows 10 Enterprise"`
	OperatingSystemVersion     string   `neo4j:"operatingsystemversion,omitempty" json:"operatingsystemversion,omitempty" desc:"OS version." example:"10.0 (19044)"`
	OperatingSystemServicePack string   `neo4j:"operatingsystemservicepack,omitempty" json:"operatingsystemservicepack,omitempty" desc:"OS service pack." example:"Service Pack 1"`
	ServicePrincipalNames      []string `neo4j:"serviceprincipalnames,omitempty" json:"serviceprincipalnames,omitempty" desc:"Service principal names." example:"[\"HOST/computer.example.com\"]"`

	// GPO Properties
	GPCFileSysPath string `neo4j:"gpcfilesyspath,omitempty" json:"gpcfilesyspath,omitempty" desc:"GPO file system path." example:"\\\\example.com\\sysvol\\example.com\\Policies\\{GUID}"`
	VersionNumber  int    `neo4j:"versionnumber,omitempty" json:"versionnumber,omitempty" desc:"GPO version number." example:"1"`

	// LAPS Properties
	HasLAPS            bool  `neo4j:"haslaps,omitempty" json:"haslaps,omitempty" desc:"LAPS is enabled for this computer." example:"true"`
	LAPSExpirationTime int64 `neo4j:"lapsexpirationtime,omitempty" json:"lapsexpirationtime,omitempty" desc:"LAPS password expiration time." example:"1698408000"`

	// Trust Properties
	TrustDirection  string `neo4j:"trustdirection,omitempty" json:"trustdirection,omitempty" desc:"Trust direction (Inbound, Outbound, Bidirectional)." example:"Bidirectional"`
	TrustType       string `neo4j:"trusttype,omitempty" json:"trusttype,omitempty" desc:"Trust type." example:"Forest"`
	TrustAttributes int    `neo4j:"trustattributes,omitempty" json:"trustattributes,omitempty" desc:"Trust attributes flags." example:"8"`
	SIDFiltering    bool   `neo4j:"sidfiltering,omitempty" json:"sidfiltering,omitempty" desc:"SID filtering enabled." example:"true"`
	IsTransitive    bool   `neo4j:"istransitive,omitempty" json:"istransitive,omitempty" desc:"Trust is transitive." example:"true"`

	// Certificate Properties
	CertThumbprint                  string   `neo4j:"certthumbprint,omitempty" json:"certthumbprint,omitempty" desc:"Certificate thumbprint." example:"ABC123DEF456"`
	CertThumbprints                 []string `neo4j:"certthumbprints,omitempty" json:"certthumbprints,omitempty" desc:"Multiple certificate thumbprints." example:"[\"ABC123\", \"DEF456\"]"`
	CertChain                       []string `neo4j:"certchain,omitempty" json:"certchain,omitempty" desc:"Certificate chain." example:"[\"root\", \"intermediate\", \"leaf\"]"`
	CertName                        string   `neo4j:"certname,omitempty" json:"certname,omitempty" desc:"Certificate name." example:"test-cert"`
	CAName                          string   `neo4j:"caname,omitempty" json:"caname,omitempty" desc:"Certificate Authority name." example:"Example-CA"`
	HasEnrollmentAgentRestrictions  bool     `neo4j:"hasenrollmentagentrestrictions,omitempty" json:"hasenrollmentagentrestrictions,omitempty" desc:"Has enrollment agent restrictions." example:"false"`
	EnrollmentAgentRestrictionsJSON string   `neo4j:"enrollmentagentrestrictionsjson,omitempty" json:"enrollmentagentrestrictionsjson,omitempty" desc:"Enrollment agent restrictions as JSON." example:"{}"`

	// Additional Properties
	Email              string   `neo4j:"email,omitempty" json:"email,omitempty" desc:"Email address." example:"john.doe@example.com"`
	Title              string   `neo4j:"title,omitempty" json:"title,omitempty" desc:"Job title." example:"Software Engineer"`
	Department         string   `neo4j:"department,omitempty" json:"department,omitempty" desc:"Department." example:"Engineering"`
	Company            string   `neo4j:"company,omitempty" json:"company,omitempty" desc:"Company name." example:"Example Corp"`
	HomeDirectory      string   `neo4j:"homedirectory,omitempty" json:"homedirectory,omitempty" desc:"Home directory path." example:"\\\\server\\homes\\jdoe"`
	UserPrincipalName  string   `neo4j:"userprincipalname,omitempty" json:"userprincipalname,omitempty" desc:"User principal name." example:"jdoe@example.com"`
	Manager            string   `neo4j:"manager,omitempty" json:"manager,omitempty" desc:"Manager's DN." example:"CN=Manager,CN=Users,DC=example,DC=local"`
	SecurityDescriptor string   `neo4j:"securitydescriptor,omitempty" json:"securitydescriptor,omitempty" desc:"Security descriptor." example:"O:DAG:DAD:PAI(A;;FA;;;DA)"`
	UserAccountControl int      `neo4j:"useraccountcontrol,omitempty" json:"useraccountcontrol,omitempty" desc:"User account control flags." example:"512"`
	SIDHistory         []string `neo4j:"sidhistory,omitempty" json:"sidhistory,omitempty" desc:"SID history." example:"[\"S-1-5-21-OLD-DOMAIN\"]"`
	IsDC               bool     `neo4j:"isdc,omitempty" json:"isdc,omitempty" desc:"Is a domain controller." example:"false"`
	IsGC               bool     `neo4j:"isgc,omitempty" json:"isgc,omitempty" desc:"Is a global catalog." example:"false"`
	IsRODC             bool     `neo4j:"isrodc,omitempty" json:"isrodc,omitempty" desc:"Is a read-only domain controller." example:"false"`
	FunctionalLevel    string   `neo4j:"functionallevel,omitempty" json:"functionallevel,omitempty" desc:"Domain/Forest functional level." example:"2016"`
	DomainFQDN         string   `neo4j:"domainfqdn,omitempty" json:"domainfqdn,omitempty" desc:"Domain FQDN." example:"example.com"`
	ForestName         string   `neo4j:"forestname,omitempty" json:"forestname,omitempty" desc:"Forest name." example:"example.com"`
}

func (ad *ADProperties) Visit(other ADProperties) {
	marshaled, _ := json.Marshal(other)
	json.Unmarshal(marshaled, ad)
}
