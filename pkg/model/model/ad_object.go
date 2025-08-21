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
	registry.ModelAlias
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
	keyMatches := adObjectKeyPattern.MatchString(ad.Key)

	return hasObjectID && hasDomain && keyMatches
}

func (ad *ADObject) Group() string {
	return ad.Domain
}

func (ad *ADObject) Identifier() string {
	return ad.ObjectID
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
func (ad *ADObject) IsClass(class string) bool {
	return strings.EqualFold(ad.Class, class) || strings.EqualFold("adobject", class)
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
				ad.Label = ad.getADLabel()
				ad.Domain = strings.ToLower(ad.Domain)
				ad.ObjectID = strings.ToUpper(ad.ObjectID)

				ad.Key = fmt.Sprintf("#%s#%s#%s", strings.ToLower(ad.Label), ad.Domain, ad.ObjectID)

				ad.Class = strings.ToLower(strings.TrimPrefix(ad.Label, "AD"))

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

func (ad *ADObject) getADLabel() string {
	label := GetADLabel(ad.Label)
	if label != "" {
		return label
	}

	label = GetADLabel(ad.Alias)
	if label != "" {
		return label
	}

	return ADObjectLabel
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

// NewADDomain creates a new AD Domain object
func NewADDomain(domain, objectID, distinguishedName string) ADObject {
	return NewADObject(domain, objectID, distinguishedName, ADDomainLabel)
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
	Name        string `neo4j:"name,omitempty" json:"name,omitempty" desc:"Common name of the AD object" example:"John Smith"`
	Description string `neo4j:"description,omitempty" json:"description,omitempty" desc:"Descriptive text for the AD object" example:"IT Department Administrator"`
	DisplayName string `neo4j:"displayname,omitempty" json:"displayname,omitempty" desc:"Display name of the AD object" example:"Smith, John (IT)"`
	IsDeleted   bool   `neo4j:"isdeleted,omitempty" json:"isdeleted,omitempty" desc:"Whether the object has been deleted from AD" example:"false"`

	// Remaining properties
	AdminCount                              bool     `neo4j:"admincount,omitempty" json:"admincount,omitempty" desc:"Indicates if object is protected by AdminSDHolder" example:"true"`
	CASecurityCollected                     bool     `neo4j:"casecuritycollected,omitempty" json:"casecuritycollected,omitempty" desc:"Whether Certificate Authority security information has been collected" example:"true"`
	CAName                                  string   `neo4j:"caname,omitempty" json:"caname,omitempty" desc:"Name of the Certificate Authority" example:"CORP-CA-01"`
	CertChain                               []string `neo4j:"certchain,omitempty" json:"certchain,omitempty" desc:"Certificate chain for the certificate" example:"[\"CN=Root CA\", \"CN=Intermediate CA\", \"CN=Issuing CA\"]"`
	CertName                                string   `neo4j:"certname,omitempty" json:"certname,omitempty" desc:"Common name of the certificate" example:"UserAuthentication"`
	CertThumbprint                          string   `neo4j:"certthumbprint,omitempty" json:"certthumbprint,omitempty" desc:"SHA1 thumbprint of the certificate" example:"1234567890ABCDEF1234567890ABCDEF12345678"`
	CertThumbprints                         []string `neo4j:"certthumbprints,omitempty" json:"certthumbprints,omitempty" desc:"List of certificate thumbprints associated with the object" example:"[\"1234567890ABCDEF1234567890ABCDEF12345678\", \"ABCDEF1234567890ABCDEF1234567890ABCDEF12\"]"`
	HasEnrollmentAgentRestrictions          string   `neo4j:"hasenrollmentagentrestrictions,omitempty" json:"hasenrollmentagentrestrictions,omitempty" desc:"Whether enrollment agent restrictions are configured" example:"true"`
	EnrollmentAgentRestrictionsCollected    bool     `neo4j:"enrollmentagentrestrictionscollected,omitempty" json:"enrollmentagentrestrictionscollected,omitempty" desc:"Whether enrollment agent restrictions data has been collected" example:"true"`
	IsUserSpecifiesSanEnabled               string   `neo4j:"isuserspecifiessanenabled,omitempty" json:"isuserspecifiessanenabled,omitempty" desc:"Whether users can specify Subject Alternative Name in certificate requests" example:"false"`
	IsUserSpecifiesSanEnabledCollected      bool     `neo4j:"isuserspecifiessanenabledcollected,omitempty" json:"isuserspecifiessanenabledcollected,omitempty" desc:"Whether SAN enablement data has been collected" example:"true"`
	RoleSeparationEnabled                   string   `neo4j:"roleseparationenabled,omitempty" json:"roleseparationenabled,omitempty" desc:"Whether CA role separation is enforced" example:"true"`
	RoleSeparationEnabledCollected          bool     `neo4j:"roleseparationenabledcollected,omitempty" json:"roleseparationenabledcollected,omitempty" desc:"Whether role separation data has been collected" example:"true"`
	HasBasicConstraints                     bool     `neo4j:"hasbasicconstraints,omitempty" json:"hasbasicconstraints,omitempty" desc:"Whether certificate has basic constraints extension" example:"true"`
	BasicConstraintPathLength               int      `neo4j:"basicconstraintpathlength,omitempty" json:"basicconstraintpathlength,omitempty" desc:"Maximum number of CA certificates in certification path" example:"2"`
	UnresolvedPublishedTemplates            []string `neo4j:"unresolvedpublishedtemplates,omitempty" json:"unresolvedpublishedtemplates,omitempty" desc:"List of certificate templates that could not be resolved" example:"[\"CustomTemplate1\", \"LegacyTemplate2\"]"`
	DNSHostname                             string   `neo4j:"dnshostname,omitempty" json:"dnshostname,omitempty" desc:"DNS hostname of the computer object" example:"srv01.contoso.local"`
	CrossCertificatePair                    []string `neo4j:"crosscertificatepair,omitempty" json:"crosscertificatepair,omitempty" desc:"Cross-certificates for establishing trust between CAs" example:"[\"MIIDXTCCAkWgAwIBAgIJAKs...\"]"`
	DistinguishedName                       string   `neo4j:"distinguishedname,omitempty" json:"distinguishedname,omitempty" desc:"Full distinguished name of the AD object" example:"CN=John Smith,OU=Users,DC=contoso,DC=local"`
	DomainFQDN                              string   `neo4j:"domain,omitempty" json:"domain,omitempty" desc:"Fully qualified domain name" example:"contoso.local"`
	DomainSID                               string   `neo4j:"domainsid,omitempty" json:"domainsid,omitempty" desc:"Security identifier of the domain" example:"S-1-5-21-3623811015-3361044348-30300820"`
	Sensitive                               bool     `neo4j:"sensitive,omitempty" json:"sensitive,omitempty" desc:"Account is marked as sensitive and cannot be delegated" example:"true"`
	BlocksInheritance                       bool     `neo4j:"blocksinheritance,omitempty" json:"blocksinheritance,omitempty" desc:"Whether GPO inheritance is blocked at this container" example:"false"`
	IsACL                                   string   `neo4j:"isacl,omitempty" json:"isacl,omitempty" desc:"Whether ACL data is available for this object" example:"true"`
	IsACLProtected                          bool     `neo4j:"isaclprotected,omitempty" json:"isaclprotected,omitempty" desc:"Whether ACL inheritance is disabled" example:"false"`
	InheritanceHash                         string   `neo4j:"inheritancehash,omitempty" json:"inheritancehash,omitempty" desc:"Hash of the inheritance chain for GPO processing" example:"A1B2C3D4E5F6"`
	InheritanceHashes                       string   `neo4j:"inheritancehashes,omitempty" json:"inheritancehashes,omitempty" desc:"Collection of inheritance hashes for the object" example:"[\"A1B2C3D4E5F6\", \"F6E5D4C3B2A1\"]"`
	Enforced                                string   `neo4j:"enforced,omitempty" json:"enforced,omitempty" desc:"Whether GPO link is enforced (no override)" example:"true"`
	Department                              string   `neo4j:"department,omitempty" json:"department,omitempty" desc:"Department the user belongs to" example:"Information Technology"`
	HasCrossCertificatePair                 bool     `neo4j:"hascrosscertificatepair,omitempty" json:"hascrosscertificatepair,omitempty" desc:"Whether object has cross-certificate pairs" example:"false"`
	HasSPN                                  bool     `neo4j:"hasspn,omitempty" json:"hasspn,omitempty" desc:"Whether object has Service Principal Names registered" example:"true"`
	UnconstrainedDelegation                 bool     `neo4j:"unconstraineddelegation,omitempty" json:"unconstraineddelegation,omitempty" desc:"Account is trusted for unconstrained Kerberos delegation" example:"false"`
	LastLogon                               int64    `neo4j:"lastlogon,omitempty" json:"lastlogon,omitempty" desc:"Last logon time in Windows NT time format" example:"132514789200000000"`
	LastLogonTimestamp                      int64    `neo4j:"lastlogontimestamp,omitempty" json:"lastlogontimestamp,omitempty" desc:"Replicated last logon timestamp" example:"132514789200000000"`
	IsPrimaryGroup                          string   `neo4j:"isprimarygroup,omitempty" json:"isprimarygroup,omitempty" desc:"Whether this is the primary group for any users" example:"true"`
	HasLAPS                                 bool     `neo4j:"haslaps,omitempty" json:"haslaps,omitempty" desc:"Whether Local Administrator Password Solution is enabled" example:"true"`
	DontRequirePreAuth                      bool     `neo4j:"dontreqpreauth,omitempty" json:"dontreqpreauth,omitempty" desc:"Kerberos pre-authentication is not required" example:"false"`
	LogonType                               string   `neo4j:"logontype,omitempty" json:"logontype,omitempty" desc:"Type of logon allowed for the account" example:"Interactive"`
	HasURA                                  string   `neo4j:"hasura,omitempty" json:"hasura,omitempty" desc:"Whether User Rights Assignments are configured" example:"true"`
	PasswordNeverExpires                    bool     `neo4j:"pwdneverexpires,omitempty" json:"pwdneverexpires,omitempty" desc:"Password is set to never expire" example:"false"`
	PasswordNotRequired                     bool     `neo4j:"passwordnotreqd,omitempty" json:"passwordnotreqd,omitempty" desc:"No password is required for the account" example:"false"`
	FunctionalLevel                         string   `neo4j:"functionallevel,omitempty" json:"functionallevel,omitempty" desc:"Domain or forest functional level" example:"2016"`
	TrustType                               string   `neo4j:"trusttype,omitempty" json:"trusttype,omitempty" desc:"Type of AD trust relationship" example:"ParentChild"`
	SpoofSIDHistoryBlocked                  string   `neo4j:"spoofsidhistoryblocked,omitempty" json:"spoofsidhistoryblocked,omitempty" desc:"Whether SID history spoofing is blocked" example:"true"`
	TrustedToAuth                           bool     `neo4j:"trustedtoauth,omitempty" json:"trustedtoauth,omitempty" desc:"Account is trusted for constrained delegation with protocol transition" example:"false"`
	SAMAccountName                          string   `neo4j:"samaccountname,omitempty" json:"samaccountname,omitempty" desc:"Pre-Windows 2000 logon name" example:"jsmith"`
	CertificateMappingMethodsRaw            string   `neo4j:"certificatemappingmethodsraw,omitempty" json:"certificatemappingmethodsraw,omitempty" desc:"Raw certificate mapping methods value" example:"0x1F"`
	CertificateMappingMethods               string   `neo4j:"certificatemappingmethods,omitempty" json:"certificatemappingmethods,omitempty" desc:"Certificate to account mapping methods" example:"Subject,Issuer,SAN"`
	StrongCertificateBindingEnforcementRaw  string   `neo4j:"strongcertificatebindingenforcementraw,omitempty" json:"strongcertificatebindingenforcementraw,omitempty" desc:"Raw strong certificate binding enforcement value" example:"2"`
	StrongCertificateBindingEnforcement     string   `neo4j:"strongcertificatebindingenforcement,omitempty" json:"strongcertificatebindingenforcement,omitempty" desc:"Level of strong certificate binding enforcement" example:"Full"`
	EKUs                                    []string `neo4j:"ekus,omitempty" json:"ekus,omitempty" desc:"Extended Key Usage OIDs for certificates" example:"[\"1.3.6.1.5.5.7.3.2\", \"1.3.6.1.5.5.7.3.4\"]"`
	SubjectAltRequireUPN                    bool     `neo4j:"subjectaltrequireupn,omitempty" json:"subjectaltrequireupn,omitempty" desc:"Certificate requires UPN in Subject Alternative Name" example:"true"`
	SubjectAltRequireDNS                    bool     `neo4j:"subjectaltrequiredns,omitempty" json:"subjectaltrequiredns,omitempty" desc:"Certificate requires DNS name in Subject Alternative Name" example:"false"`
	SubjectAltRequireDomainDNS              bool     `neo4j:"subjectaltrequiredomaindns,omitempty" json:"subjectaltrequiredomaindns,omitempty" desc:"Certificate requires domain DNS in Subject Alternative Name" example:"false"`
	SubjectAltRequireEmail                  bool     `neo4j:"subjectaltrequireemail,omitempty" json:"subjectaltrequireemail,omitempty" desc:"Certificate requires email in Subject Alternative Name" example:"true"`
	SubjectAltRequireSPN                    bool     `neo4j:"subjectaltrequirespn,omitempty" json:"subjectaltrequirespn,omitempty" desc:"Certificate requires SPN in Subject Alternative Name" example:"false"`
	SubjectRequireEmail                     bool     `neo4j:"subjectrequireemail,omitempty" json:"subjectrequireemail,omitempty" desc:"Certificate requires email in subject" example:"false"`
	AuthorizedSignatures                    int      `neo4j:"authorizedsignatures,omitempty" json:"authorizedsignatures,omitempty" desc:"Number of authorized signatures required" example:"1"`
	ApplicationPolicies                     []string `neo4j:"applicationpolicies,omitempty" json:"applicationpolicies,omitempty" desc:"Application policy OIDs for certificates" example:"[\"1.3.6.1.5.5.7.3.2\"]"`
	IssuancePolicies                        []string `neo4j:"issuancepolicies,omitempty" json:"issuancepolicies,omitempty" desc:"Certificate issuance policy OIDs" example:"[\"1.3.6.1.4.1.311.21.8.1\"]"`
	SchemaVersion                           int      `neo4j:"schemaversion,omitempty" json:"schemaversion,omitempty" desc:"Certificate template schema version" example:"2"`
	RequiresManagerApproval                 bool     `neo4j:"requiresmanagerapproval,omitempty" json:"requiresmanagerapproval,omitempty" desc:"Certificate enrollment requires manager approval" example:"true"`
	AuthenticationEnabled                   bool     `neo4j:"authenticationenabled,omitempty" json:"authenticationenabled,omitempty" desc:"Authentication is enabled for the certificate template" example:"true"`
	SchannelAuthenticationEnabled           bool     `neo4j:"schannelauthenticationenabled,omitempty" json:"schannelauthenticationenabled,omitempty" desc:"SChannel authentication is enabled" example:"false"`
	EnrolleeSuppliesSubject                 bool     `neo4j:"enrolleesuppliessubject,omitempty" json:"enrolleesuppliessubject,omitempty" desc:"Enrollee can supply subject information in certificate request" example:"false"`
	CertificateApplicationPolicy            []string `neo4j:"certificateapplicationpolicy,omitempty" json:"certificateapplicationpolicy,omitempty" desc:"Certificate application policy extensions" example:"[\"1.3.6.1.5.5.7.3.2\"]"`
	CertificateNameFlag                     string   `neo4j:"certificatenameflag,omitempty" json:"certificatenameflag,omitempty" desc:"Certificate name flags configuration" example:"SubjectRequireDirectoryPath"`
	EffectiveEKUs                           []string `neo4j:"effectiveekus,omitempty" json:"effectiveekus,omitempty" desc:"Effective Extended Key Usage OIDs after policy application" example:"[\"1.3.6.1.5.5.7.3.2\", \"1.3.6.1.5.5.7.3.4\"]"`
	EnrollmentFlag                          string   `neo4j:"enrollmentflag,omitempty" json:"enrollmentflag,omitempty" desc:"Certificate enrollment flags" example:"AutoEnrollment"`
	Flags                                   string   `neo4j:"flags,omitempty" json:"flags,omitempty" desc:"General purpose flags for the object" example:"0x00000001"`
	NoSecurityExtension                     bool     `neo4j:"nosecurityextension,omitempty" json:"nosecurityextension,omitempty" desc:"Certificate template has no security extension" example:"false"`
	RenewalPeriod                           string   `neo4j:"renewalperiod,omitempty" json:"renewalperiod,omitempty" desc:"Certificate renewal period" example:"6 weeks"`
	ValidityPeriod                          string   `neo4j:"validityperiod,omitempty" json:"validityperiod,omitempty" desc:"Certificate validity period" example:"1 year"`
	OID                                     string   `neo4j:"oid,omitempty" json:"oid,omitempty" desc:"Object identifier for the certificate template" example:"1.3.6.1.4.1.311.21.8.1234567.1234567.1.1.1"`
	HomeDirectory                           string   `neo4j:"homedirectory,omitempty" json:"homedirectory,omitempty" desc:"User's home directory path" example:"\\\\fileserver\\users\\jsmith"`
	CertificatePolicy                       []string `neo4j:"certificatepolicy,omitempty" json:"certificatepolicy,omitempty" desc:"Certificate policy OIDs" example:"[\"1.3.6.1.4.1.311.21.8.1\", \"1.3.6.1.5.5.7.2.1\"]"`
	CertTemplateOID                         string   `neo4j:"certtemplateoid,omitempty" json:"certtemplateoid,omitempty" desc:"Certificate template object identifier" example:"1.3.6.1.4.1.311.21.8.1234567.1234567.1.1.1"`
	GroupLinkID                             string   `neo4j:"grouplinkid,omitempty" json:"grouplinkid,omitempty" desc:"Link ID for group policy objects" example:"{31B2F340-016D-11D2-945F-00C04FB984F9}"`
	ObjectGUID                              string   `neo4j:"objectguid,omitempty" json:"objectguid,omitempty" desc:"Globally unique identifier for the AD object" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	ExpirePasswordsOnSmartCardOnlyAccounts  bool     `neo4j:"expirepasswordsonsmartcardonlyaccounts,omitempty" json:"expirepasswordsonsmartcardonlyaccounts,omitempty" desc:"Whether passwords expire for smart card only accounts" example:"false"`
	MachineAccountQuota                     int      `neo4j:"machineaccountquota,omitempty" json:"machineaccountquota,omitempty" desc:"Number of computer accounts a user can create" example:"10"`
	SupportedKerberosEncryptionTypes        []string `neo4j:"supportedencryptiontypes,omitempty" json:"supportedencryptiontypes,omitempty" desc:"Supported Kerberos encryption types" example:"[\"RC4_HMAC_MD5\", \"AES128_CTS_HMAC_SHA1_96\", \"AES256_CTS_HMAC_SHA1_96\"]"`
	TGTDelegation                           string   `neo4j:"tgtdelegation,omitempty" json:"tgtdelegation,omitempty" desc:"TGT delegation configuration" example:"Enabled"`
	PasswordStoredUsingReversibleEncryption bool     `neo4j:"encryptedtextpwdallowed,omitempty" json:"encryptedtextpwdallowed,omitempty" desc:"Password is stored using reversible encryption" example:"false"`
	SmartcardRequired                       bool     `neo4j:"smartcardrequired,omitempty" json:"smartcardrequired,omitempty" desc:"Smart card is required for interactive logon" example:"false"`
	UseDESKeyOnly                           bool     `neo4j:"usedeskeyonly,omitempty" json:"usedeskeyonly,omitempty" desc:"Use only DES encryption keys for Kerberos" example:"false"`
	LogonScriptEnabled                      bool     `neo4j:"logonscriptenabled,omitempty" json:"logonscriptenabled,omitempty" desc:"Logon script is enabled for the account" example:"true"`
	LockedOut                               bool     `neo4j:"lockedout,omitempty" json:"lockedout,omitempty" desc:"Account is currently locked out" example:"false"`
	UserCannotChangePassword                bool     `neo4j:"passwordcantchange,omitempty" json:"passwordcantchange,omitempty" desc:"User cannot change their password" example:"false"`
	PasswordExpired                         bool     `neo4j:"passwordexpired,omitempty" json:"passwordexpired,omitempty" desc:"Password has expired" example:"false"`
	DSHeuristics                            string   `neo4j:"dsheuristics,omitempty" json:"dsheuristics,omitempty" desc:"Directory Service heuristics configuration" example:"0000000001"`
	UserAccountControl                      int      `neo4j:"useraccountcontrol,omitempty" json:"useraccountcontrol,omitempty" desc:"User account control flags bitmask" example:"512"`
	TrustAttributesInbound                  string   `neo4j:"trustattributesinbound,omitempty" json:"trustattributesinbound,omitempty" desc:"Inbound trust attributes" example:"0x00000020"`
	TrustAttributesOutbound                 string   `neo4j:"trustattributesoutbound,omitempty" json:"trustattributesoutbound,omitempty" desc:"Outbound trust attributes" example:"0x00000020"`
	MinPwdLength                            int      `neo4j:"minpwdlength,omitempty" json:"minpwdlength,omitempty" desc:"Minimum password length requirement" example:"8"`
	PwdProperties                           int      `neo4j:"pwdproperties,omitempty" json:"pwdproperties,omitempty" desc:"Password policy properties bitmask" example:"1"`
	PwdHistoryLength                        int      `neo4j:"pwdhistorylength,omitempty" json:"pwdhistorylength,omitempty" desc:"Number of passwords remembered in history" example:"24"`
	LockoutThreshold                        int      `neo4j:"lockoutthreshold,omitempty" json:"lockoutthreshold,omitempty" desc:"Number of failed logon attempts before lockout" example:"5"`
	MinPwdAge                               string   `neo4j:"minpwdage,omitempty" json:"minpwdage,omitempty" desc:"Minimum password age" example:"1d"`
	MaxPwdAge                               string   `neo4j:"maxpwdage,omitempty" json:"maxpwdage,omitempty" desc:"Maximum password age" example:"90d"`
	LockoutDuration                         string   `neo4j:"lockoutduration,omitempty" json:"lockoutduration,omitempty" desc:"Account lockout duration" example:"30m"`
	LockoutObservationWindow                int      `neo4j:"lockoutobservationwindow,omitempty" json:"lockoutobservationwindow,omitempty" desc:"Time window in minutes for observing failed logon attempts" example:"30"`
	OwnerSid                                string   `neo4j:"ownersid,omitempty" json:"ownersid,omitempty" desc:"Security identifier of the object owner" example:"S-1-5-21-3623811015-3361044348-30300820-1001"`
	SMBSigning                              bool     `neo4j:"smbsigning,omitempty" json:"smbsigning,omitempty" desc:"SMB signing is required" example:"true"`
	WebClientRunning                        bool     `neo4j:"webclientrunning,omitempty" json:"webclientrunning,omitempty" desc:"Whether WebDAV client service is running" example:"true"`
	RestrictOutboundNTLM                    bool     `neo4j:"restrictoutboundntlm,omitempty" json:"restrictoutboundntlm,omitempty" desc:"Outbound NTLM authentication is restricted" example:"false"`
	GMSA                                    bool     `neo4j:"gmsa,omitempty" json:"gmsa,omitempty" desc:"Group Managed Service Account" example:"true"`
	MSA                                     bool     `neo4j:"msa,omitempty" json:"msa,omitempty" desc:"Managed Service Account" example:"false"`
	DoesAnyAceGrantOwnerRights              bool     `neo4j:"doesanyacegrantownerrights,omitempty" json:"doesanyacegrantownerrights,omitempty" desc:"Whether any ACE grants owner rights" example:"true"`
	DoesAnyInheritedAceGrantOwnerRights     bool     `neo4j:"doesanyinheritedacegrantownerrights,omitempty" json:"doesanyinheritedacegrantownerrights,omitempty" desc:"Whether any inherited ACE grants owner rights" example:"false"`
	ADCSWebEnrollmentHTTP                   string   `neo4j:"adcswebenrollmenthttp,omitempty" json:"adcswebenrollmenthttp,omitempty" desc:"ADCS web enrollment HTTP endpoint availability" example:"http://ca.contoso.local/certsrv"`
	ADCSWebEnrollmentHTTPS                  string   `neo4j:"adcswebenrollmenthttps,omitempty" json:"adcswebenrollmenthttps,omitempty" desc:"ADCS web enrollment HTTPS endpoint availability" example:"https://ca.contoso.local/certsrv"`
	ADCSWebEnrollmentHTTPSEPA               string   `neo4j:"adcswebenrollmenthttpsepa,omitempty" json:"adcswebenrollmenthttpsepa,omitempty" desc:"ADCS web enrollment HTTPS with Extended Protection" example:"https://ca.contoso.local/certsrv"`
	LDAPSigning                             bool     `neo4j:"ldapsigning,omitempty" json:"ldapsigning,omitempty" desc:"LDAP signing requirement" example:"Required"`
	LDAPAvailable                           bool     `neo4j:"ldapavailable,omitempty" json:"ldapavailable,omitempty" desc:"Whether LDAP service is available" example:"true"`
	LDAPSAvailable                          bool     `neo4j:"ldapsavailable,omitempty" json:"ldapsavailable,omitempty" desc:"Whether LDAPS (secure LDAP) is available" example:"true"`
	LDAPSEPA                                bool     `neo4j:"ldapsepa,omitempty" json:"ldapsepa,omitempty" desc:"LDAPS with Extended Protection for Authentication" example:"Enabled"`
	IsDC                                    bool     `neo4j:"isdc,omitempty" json:"isdc,omitempty" desc:"Whether computer is a Domain Controller" example:"true"`
	IsReadOnlyDC                            bool     `neo4j:"isreadonlydc,omitempty" json:"isreadonlydc,omitempty" desc:"Whether computer is a Read-Only Domain Controller" example:"false"`
	HTTPEnrollmentEndpoints                 string   `neo4j:"httpenrollmentendpoints,omitempty" json:"httpenrollmentendpoints,omitempty" desc:"List of HTTP certificate enrollment endpoints" example:"[\"http://ca1.contoso.local/certsrv\", \"http://ca2.contoso.local/certsrv\"]"`
	HTTPSEnrollmentEndpoints                string   `neo4j:"httpsenrollmentendpoints,omitempty" json:"httpsenrollmentendpoints,omitempty" desc:"List of HTTPS certificate enrollment endpoints" example:"[\"https://ca1.contoso.local/certsrv\", \"https://ca2.contoso.local/certsrv\"]"`
	HasVulnerableEndpoint                   string   `neo4j:"hasvulnerableendpoint,omitempty" json:"hasvulnerableendpoint,omitempty" desc:"Whether object has vulnerable enrollment endpoints" example:"true"`
	RequireSecuritySignature                string   `neo4j:"requiresecuritysignature,omitempty" json:"requiresecuritysignature,omitempty" desc:"Whether security signature is required" example:"true"`
	EnableSecuritySignature                 string   `neo4j:"enablesecuritysignature,omitempty" json:"enablesecuritysignature,omitempty" desc:"Whether security signature is enabled" example:"true"`
	RestrictReceivingNTLMTraffic            string   `neo4j:"restrictreceivingntmltraffic,omitempty" json:"restrictreceivingntmltraffic,omitempty" desc:"Restriction policy for receiving NTLM traffic" example:"DenyAll"`
	NTLMMinServerSec                        string   `neo4j:"ntlmminserversec,omitempty" json:"ntlmminserversec,omitempty" desc:"Minimum security level for NTLM SSP server" example:"537395200"`
	NTLMMinClientSec                        string   `neo4j:"ntlmminclientsec,omitempty" json:"ntlmminclientsec,omitempty" desc:"Minimum security level for NTLM SSP client" example:"537395200"`
	LMCompatibilityLevel                    string   `neo4j:"lmcompatibilitylevel,omitempty" json:"lmcompatibilitylevel,omitempty" desc:"LAN Manager authentication compatibility level" example:"5"`
	UseMachineID                            string   `neo4j:"usemachineid,omitempty" json:"usemachineid,omitempty" desc:"Whether to use machine identity for authentication" example:"true"`
	ClientAllowedNTLMServers                string   `neo4j:"clientallowedntlmservers,omitempty" json:"clientallowedntlmservers,omitempty" desc:"List of servers allowed to use NTLM authentication" example:"*.contoso.local"`
	Transitive                              string   `neo4j:"transitive,omitempty" json:"transitive,omitempty" desc:"Whether trust relationship is transitive" example:"true"`
	GroupScope                              string   `neo4j:"groupscope,omitempty" json:"groupscope,omitempty" desc:"Scope of the AD group" example:"Global"`
	NetBIOS                                 string   `neo4j:"netbios,omitempty" json:"netbios,omitempty" desc:"NetBIOS name of the domain" example:"CONTOSO"`
	AdminSDHolderProtected                  string   `neo4j:"adminsdholderprotected,omitempty" json:"adminsdholderprotected,omitempty" desc:"Whether object is protected by AdminSDHolder process" example:"true"`
}

func (ad *ADProperties) Visit(other ADProperties) {
	marshaled, _ := json.Marshal(other)
	json.Unmarshal(marshaled, ad)
}
