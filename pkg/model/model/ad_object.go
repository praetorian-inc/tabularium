package model

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
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

var ADLabels = []string{
	ADObjectLabel,
	ADUserLabel,
	ADComputerLabel,
	ADGroupLabel,
	ADGPOLabel,
	ADOULabel,
	ADContainerLabel,
	ADDomainLabel,
	ADLocalGroupLabel,
	ADLocalUserLabel,
	ADAIACALabel,
	ADRootCALabel,
	ADEnterpriseCALabel,
	ADNTAuthStoreLabel,
	ADCertTemplateLabel,
	ADIssuancePolicyLabel,
}

func init() {
	for _, label := range ADLabels {
		MustRegisterLabel(label)
	}

	registry.Registry.MustRegisterModel(&ADObject{}, ADLabels...)
}

var (
	adObjectKeyPattern = regexp.MustCompile(`(?i)^#ad[a-z]+#[^#]+#[A-FS0-9-]+$`)
)

const TierZeroTag = "tier-zero"

var TierZeroSIDSuffixes = []string{
	"S-1-5-9", // Enterprise Domain Controllers
	"-500",    // Administrator Account
	"-512",    // Domain Admins
	"-516",    // Domain Controllers
	"-518",    // Schema Admins
	"-519",    // Enterprise Admins
	"-526",    // Key Admins
	"-527",    // Enterprise Key Admins
	"-544",    // Administrators
	"-551",    // Backup Operators
}

type ADObject struct {
	BaseAsset
	registry.ModelAlias
	Label           string   `neo4j:"label" json:"label" desc:"Primary label of the object." example:"ADUser" capmodel:"ADObject"`
	SecondaryLabels []string `neo4j:"-" json:"labels" desc:"Secondary labels of the object." example:"ADLocalGroup" capmodel:"ADObject"`
	Domain          string   `neo4j:"domain" json:"domain" desc:"AD domain this object belongs to." example:"example.local" capmodel:"ADObject"`
	ObjectID        string   `neo4j:"objectid" json:"objectid" desc:"Object identifier." example:"S-1-5-21-123456789-123456789-123456789-1001" capmodel:"ADObject"`
	SID             string   `neo4j:"sid" json:"sid,omitempty" desc:"Security identifier." example:"S-1-5-21-123456789-123456789-123456789-1001" capmodel:"ADObject"`
	ADProperties
}

func (ad *ADObject) GetLabels() []string {
	labels := []string{ADObjectLabel, AssetLabel, TTLLabel}
	if ad.Label != "" {
		labels = append(labels, ad.Label)
	}
	labels = append(labels, ad.SecondaryLabels...)
	if ad.Source == SeedSource {
		labels = append(labels, SeedLabel)
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
	ad.TTL = 0
}

func (d *ADObject) SeedModels() []Seedable {
	copy := *d
	return []Seedable{&copy}
}

func (ad *ADObject) IsClass(class string) bool {
	return strings.EqualFold(ad.Class, class) || strings.EqualFold("adobject", class)
}

func (ad *ADObject) Attribute(name, value string) Attribute {
	attr := NewAttribute(name, value, ad)
	return attr
}

func (ad *ADObject) WithStatus(status string) Target {
	ret := *ad
	ret.Status = status
	return &ret
}

func (ad *ADObject) Defaulted() {
	ad.BaseAsset.Defaulted()
	ad.TTL = 0
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

				if strings.HasPrefix(ad.ObjectID, "S-") {
					ad.SID = ad.ObjectID
				}

				ad.TTL = 0

				ad.tagIfTierZero()
				return nil
			},
		},
		setGroupAndIdentifier(ad, &ad.Domain, &ad.ObjectID),
	}
}

func (ad *ADObject) tagIfTierZero() {
	if slices.Contains(ad.Tags.Tags, TierZeroTag) {
		return
	}

	if ad.SID == "" {
		return
	}
	for _, suffix := range TierZeroSIDSuffixes {
		if strings.HasSuffix(ad.SID, suffix) {
			ad.Tags.Tags = append(ad.Tags.Tags, TierZeroTag)
			return
		}
	}
}

func (ad *ADObject) getADLabel() string {
	get := func(label string) (string, bool) {
		for _, l := range ADLabels {
			if strings.EqualFold(label, l) {
				return l, true
			}
		}

		return "", false
	}

	if l, ok := get(ad.Label); ok {
		return l
	}

	if l, ok := get(ad.Alias); ok {
		return l
	}

	return ADObjectLabel
}

// NewADObject creates a new ADObject with the specified domain, distinguished name, and object label
func NewADObject(domain, objectID, distinguishedName, objectLabel string) ADObject {
	ad := ADObject{
		Domain:   domain,
		ObjectID: objectID,
		Label:    objectLabel,
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

func NewADDomainSeed(domain, objectID, distinguishedName string) ADObject {
	object := NewADDomain(domain, objectID, distinguishedName)
	object.SetSource(SeedSource)
	return object
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
	Name        string `neo4j:"name" json:"name,omitempty" desc:"Common name of the AD object" example:"John Smith" capmodel:"ADObject"`
	Description string `neo4j:"description" json:"description,omitempty" desc:"Descriptive text for the AD object" example:"IT Department Administrator" capmodel:"ADObject"`
	DisplayName string `neo4j:"displayname" json:"displayname,omitempty" desc:"Display name of the AD object" example:"Smith, John (IT)" capmodel:"ADObject"`
	IsDeleted   bool   `neo4j:"isdeleted" json:"isdeleted,omitempty" desc:"Whether the object has been deleted from AD" example:"false" capmodel:"ADObject"`

	// Remaining properties
	AdminCount                              bool     `neo4j:"admincount" json:"admincount,omitempty" desc:"Indicates if object is protected by AdminSDHolder" example:"true" capmodel:"ADObject"`
	CASecurityCollected                     bool     `neo4j:"casecuritycollected" json:"casecuritycollected,omitempty" desc:"Whether Certificate Authority security information has been collected" example:"true" capmodel:"ADObject"`
	CAName                                  string   `neo4j:"caname" json:"caname,omitempty" desc:"Name of the Certificate Authority" example:"CORP-CA-01" capmodel:"ADObject"`
	CertChain                               []string `neo4j:"certchain" json:"certchain,omitempty" desc:"Certificate chain for the certificate" example:"[\"CN=Root CA\", \"CN=Intermediate CA\", \"CN=Issuing CA\"]" capmodel:"ADObject"`
	CertName                                string   `neo4j:"certname" json:"certname,omitempty" desc:"Common name of the certificate" example:"UserAuthentication" capmodel:"ADObject"`
	CertThumbprint                          string   `neo4j:"certthumbprint" json:"certthumbprint,omitempty" desc:"SHA1 thumbprint of the certificate" example:"1234567890ABCDEF1234567890ABCDEF12345678" capmodel:"ADObject"`
	CertThumbprints                         []string `neo4j:"certthumbprints" json:"certthumbprints,omitempty" desc:"List of certificate thumbprints associated with the object" example:"[\"1234567890ABCDEF1234567890ABCDEF12345678\", \"ABCDEF1234567890ABCDEF1234567890ABCDEF12\"]" capmodel:"ADObject"`
	HasEnrollmentAgentRestrictions          bool     `neo4j:"hasenrollmentagentrestrictions" json:"hasenrollmentagentrestrictions,omitempty" desc:"Whether enrollment agent restrictions are configured" example:"true" capmodel:"ADObject"`
	EnrollmentAgentRestrictionsCollected    bool     `neo4j:"enrollmentagentrestrictionscollected" json:"enrollmentagentrestrictionscollected,omitempty" desc:"Whether enrollment agent restrictions data has been collected" example:"true" capmodel:"ADObject"`
	IsUserSpecifiesSanEnabled               bool     `neo4j:"isuserspecifiessanenabled" json:"isuserspecifiessanenabled,omitempty" desc:"Whether users can specify Subject Alternative Name in certificate requests" example:"false" capmodel:"ADObject"`
	IsUserSpecifiesSanEnabledCollected      bool     `neo4j:"isuserspecifiessanenabledcollected" json:"isuserspecifiessanenabledcollected,omitempty" desc:"Whether SAN enablement data has been collected" example:"true" capmodel:"ADObject"`
	RoleSeparationEnabled                   string   `neo4j:"roleseparationenabled" json:"roleseparationenabled,omitempty" desc:"Whether CA role separation is enforced" example:"true" capmodel:"ADObject"`
	RoleSeparationEnabledCollected          bool     `neo4j:"roleseparationenabledcollected" json:"roleseparationenabledcollected,omitempty" desc:"Whether role separation data has been collected" example:"true" capmodel:"ADObject"`
	HasBasicConstraints                     bool     `neo4j:"hasbasicconstraints" json:"hasbasicconstraints,omitempty" desc:"Whether certificate has basic constraints extension" example:"true" capmodel:"ADObject"`
	BasicConstraintPathLength               int      `neo4j:"basicconstraintpathlength" json:"basicconstraintpathlength,omitempty" desc:"Maximum number of CA certificates in certification path" example:"2" capmodel:"ADObject"`
	UnresolvedPublishedTemplates            []string `neo4j:"unresolvedpublishedtemplates" json:"unresolvedpublishedtemplates,omitempty" desc:"List of certificate templates that could not be resolved" example:"[\"CustomTemplate1\", \"LegacyTemplate2\"]" capmodel:"ADObject"`
	DNSHostname                             string   `neo4j:"dnshostname" json:"dnshostname,omitempty" desc:"DNS hostname of the computer object" example:"srv01.contoso.local" capmodel:"ADObject"`
	CrossCertificatePair                    []string `neo4j:"crosscertificatepair" json:"crosscertificatepair,omitempty" desc:"Cross-certificates for establishing trust between CAs" example:"[\"MIIDXTCCAkWgAwIBAgIJAKs...\"]" capmodel:"ADObject"`
	DistinguishedName                       string   `neo4j:"distinguishedname" json:"distinguishedname,omitempty" desc:"Full distinguished name of the AD object" example:"CN=John Smith,OU=Users,DC=contoso,DC=local" capmodel:"ADObject"`
	DomainSID                               string   `neo4j:"domainsid" json:"domainsid,omitempty" desc:"Security identifier of the domain" example:"S-1-5-21-3623811015-3361044348-30300820" capmodel:"ADObject"`
	Sensitive                               bool     `neo4j:"sensitive" json:"sensitive,omitempty" desc:"Account is marked as sensitive and cannot be delegated" example:"true" capmodel:"ADObject"`
	BlocksInheritance                       bool     `neo4j:"blocksinheritance" json:"blocksinheritance,omitempty" desc:"Whether GPO inheritance is blocked at this container" example:"false" capmodel:"ADObject"`
	IsACL                                   string   `neo4j:"isacl" json:"isacl,omitempty" desc:"Whether ACL data is available for this object" example:"true" capmodel:"ADObject"`
	IsACLProtected                          bool     `neo4j:"isaclprotected" json:"isaclprotected,omitempty" desc:"Whether ACL inheritance is disabled" example:"false" capmodel:"ADObject"`
	InheritanceHash                         string   `neo4j:"inheritancehash" json:"inheritancehash,omitempty" desc:"Hash of the inheritance chain for GPO processing" example:"A1B2C3D4E5F6" capmodel:"ADObject"`
	InheritanceHashes                       []string `neo4j:"inheritancehashes" json:"inheritancehashes,omitempty" desc:"Collection of inheritance hashes for the object" example:"[\"A1B2C3D4E5F6\", \"F6E5D4C3B2A1\"]" capmodel:"ADObject"`
	Enforced                                string   `neo4j:"enforced" json:"enforced,omitempty" desc:"Whether GPO link is enforced (no override)" example:"true" capmodel:"ADObject"`
	Department                              string   `neo4j:"department" json:"department,omitempty" desc:"Department the user belongs to" example:"Information Technology" capmodel:"ADObject"`
	HasCrossCertificatePair                 bool     `neo4j:"hascrosscertificatepair" json:"hascrosscertificatepair,omitempty" desc:"Whether object has cross-certificate pairs" example:"false" capmodel:"ADObject"`
	HasSPN                                  bool     `neo4j:"hasspn" json:"hasspn,omitempty" desc:"Whether object has Service Principal Names registered" example:"true" capmodel:"ADObject"`
	UnconstrainedDelegation                 bool     `neo4j:"unconstraineddelegation" json:"unconstraineddelegation,omitempty" desc:"Account is trusted for unconstrained Kerberos delegation" example:"false" capmodel:"ADObject"`
	LastLogon                               int64    `neo4j:"lastlogon" json:"lastlogon,omitempty" desc:"Last logon time in Windows NT time format" example:"132514789200000000" capmodel:"ADObject"`
	LastLogonTimestamp                      int64    `neo4j:"lastlogontimestamp" json:"lastlogontimestamp,omitempty" desc:"Replicated last logon timestamp" example:"132514789200000000" capmodel:"ADObject"`
	IsPrimaryGroup                          string   `neo4j:"isprimarygroup" json:"isprimarygroup,omitempty" desc:"Whether this is the primary group for any users" example:"true" capmodel:"ADObject"`
	HasLAPS                                 bool     `neo4j:"haslaps" json:"haslaps,omitempty" desc:"Whether Local Administrator Password Solution is enabled" example:"true" capmodel:"ADObject"`
	DontRequirePreAuth                      bool     `neo4j:"dontreqpreauth" json:"dontreqpreauth,omitempty" desc:"Kerberos pre-authentication is not required" example:"false" capmodel:"ADObject"`
	LogonType                               string   `neo4j:"logontype" json:"logontype,omitempty" desc:"Type of logon allowed for the account" example:"Interactive" capmodel:"ADObject"`
	HasURA                                  bool     `neo4j:"hasura" json:"hasura,omitempty" desc:"Whether User Rights Assignments are configured" example:"true" capmodel:"ADObject"`
	PasswordNeverExpires                    bool     `neo4j:"pwdneverexpires" json:"pwdneverexpires,omitempty" desc:"Password is set to never expire" example:"false" capmodel:"ADObject"`
	PasswordNotRequired                     bool     `neo4j:"passwordnotreqd" json:"passwordnotreqd,omitempty" desc:"No password is required for the account" example:"false" capmodel:"ADObject"`
	FunctionalLevel                         string   `neo4j:"functionallevel" json:"functionallevel,omitempty" desc:"Domain or forest functional level" example:"2016" capmodel:"ADObject"`
	TrustType                               string   `neo4j:"trusttype" json:"trusttype,omitempty" desc:"Type of AD trust relationship" example:"ParentChild" capmodel:"ADObject"`
	SpoofSIDHistoryBlocked                  string   `neo4j:"spoofsidhistoryblocked" json:"spoofsidhistoryblocked,omitempty" desc:"Whether SID history spoofing is blocked" example:"true" capmodel:"ADObject"`
	TrustedToAuth                           bool     `neo4j:"trustedtoauth" json:"trustedtoauth,omitempty" desc:"Account is trusted for constrained delegation with protocol transition" example:"false" capmodel:"ADObject"`
	SAMAccountName                          string   `neo4j:"samaccountname" json:"samaccountname,omitempty" desc:"Pre-Windows 2000 logon name" example:"jsmith" capmodel:"ADObject"`
	CertificateMappingMethodsRaw            int      `neo4j:"certificatemappingmethodsraw" json:"certificatemappingmethodsraw,omitempty" desc:"Raw certificate mapping methods value" example:"0x1F" capmodel:"ADObject"`
	CertificateMappingMethods               []string `neo4j:"certificatemappingmethods" json:"certificatemappingmethods,omitempty" desc:"Certificate to account mapping methods" example:"Subject,Issuer,SAN" capmodel:"ADObject"`
	StrongCertificateBindingEnforcementRaw  int      `neo4j:"strongcertificatebindingenforcementraw" json:"strongcertificatebindingenforcementraw,omitempty" desc:"Raw strong certificate binding enforcement value" example:"2" capmodel:"ADObject"`
	StrongCertificateBindingEnforcement     string   `neo4j:"strongcertificatebindingenforcement" json:"strongcertificatebindingenforcement,omitempty" desc:"Level of strong certificate binding enforcement" example:"Full" capmodel:"ADObject"`
	EKUs                                    []string `neo4j:"ekus" json:"ekus,omitempty" desc:"Extended Key Usage OIDs for certificates" example:"[\"1.3.6.1.5.5.7.3.2\", \"1.3.6.1.5.5.7.3.4\"]" capmodel:"ADObject"`
	SubjectAltRequireUPN                    bool     `neo4j:"subjectaltrequireupn" json:"subjectaltrequireupn,omitempty" desc:"Certificate requires UPN in Subject Alternative Name" example:"true" capmodel:"ADObject"`
	SubjectAltRequireDNS                    bool     `neo4j:"subjectaltrequiredns" json:"subjectaltrequiredns,omitempty" desc:"Certificate requires DNS name in Subject Alternative Name" example:"false" capmodel:"ADObject"`
	SubjectAltRequireDomainDNS              bool     `neo4j:"subjectaltrequiredomaindns" json:"subjectaltrequiredomaindns,omitempty" desc:"Certificate requires domain DNS in Subject Alternative Name" example:"false" capmodel:"ADObject"`
	SubjectAltRequireEmail                  bool     `neo4j:"subjectaltrequireemail" json:"subjectaltrequireemail,omitempty" desc:"Certificate requires email in Subject Alternative Name" example:"true" capmodel:"ADObject"`
	SubjectAltRequireSPN                    bool     `neo4j:"subjectaltrequirespn" json:"subjectaltrequirespn,omitempty" desc:"Certificate requires SPN in Subject Alternative Name" example:"false" capmodel:"ADObject"`
	SubjectRequireEmail                     bool     `neo4j:"subjectrequireemail" json:"subjectrequireemail,omitempty" desc:"Certificate requires email in subject" example:"false" capmodel:"ADObject"`
	AuthorizedSignatures                    int      `neo4j:"authorizedsignatures" json:"authorizedsignatures,omitempty" desc:"Number of authorized signatures required" example:"1" capmodel:"ADObject"`
	ApplicationPolicies                     []string `neo4j:"applicationpolicies" json:"applicationpolicies,omitempty" desc:"Application policy OIDs for certificates" example:"[\"1.3.6.1.5.5.7.3.2\"]" capmodel:"ADObject"`
	IssuancePolicies                        []string `neo4j:"issuancepolicies" json:"issuancepolicies,omitempty" desc:"Certificate issuance policy OIDs" example:"[\"1.3.6.1.4.1.311.21.8.1\"]" capmodel:"ADObject"`
	SchemaVersion                           int      `neo4j:"schemaversion" json:"schemaversion,omitempty" desc:"Certificate template schema version" example:"2" capmodel:"ADObject"`
	RequiresManagerApproval                 bool     `neo4j:"requiresmanagerapproval" json:"requiresmanagerapproval,omitempty" desc:"Certificate enrollment requires manager approval" example:"true" capmodel:"ADObject"`
	AuthenticationEnabled                   bool     `neo4j:"authenticationenabled" json:"authenticationenabled,omitempty" desc:"Authentication is enabled for the certificate template" example:"true" capmodel:"ADObject"`
	SchannelAuthenticationEnabled           bool     `neo4j:"schannelauthenticationenabled" json:"schannelauthenticationenabled,omitempty" desc:"SChannel authentication is enabled" example:"false" capmodel:"ADObject"`
	EnrolleeSuppliesSubject                 bool     `neo4j:"enrolleesuppliessubject" json:"enrolleesuppliessubject,omitempty" desc:"Enrollee can supply subject information in certificate request" example:"false" capmodel:"ADObject"`
	CertificateApplicationPolicy            []string `neo4j:"certificateapplicationpolicy" json:"certificateapplicationpolicy,omitempty" desc:"Certificate application policy extensions" example:"[\"1.3.6.1.5.5.7.3.2\"]" capmodel:"ADObject"`
	CertificateNameFlag                     string   `neo4j:"certificatenameflag" json:"certificatenameflag,omitempty" desc:"Certificate name flags configuration" example:"SubjectRequireDirectoryPath" capmodel:"ADObject"`
	EffectiveEKUs                           []string `neo4j:"effectiveekus" json:"effectiveekus,omitempty" desc:"Effective Extended Key Usage OIDs after policy application" example:"[\"1.3.6.1.5.5.7.3.2\", \"1.3.6.1.5.5.7.3.4\"]" capmodel:"ADObject"`
	EnrollmentFlag                          string   `neo4j:"enrollmentflag" json:"enrollmentflag,omitempty" desc:"Certificate enrollment flags" example:"AutoEnrollment" capmodel:"ADObject"`
	Flags                                   string   `neo4j:"flags" json:"flags,omitempty" desc:"General purpose flags for the object" example:"0x00000001" capmodel:"ADObject"`
	NoSecurityExtension                     bool     `neo4j:"nosecurityextension" json:"nosecurityextension,omitempty" desc:"Certificate template has no security extension" example:"false" capmodel:"ADObject"`
	RenewalPeriod                           string   `neo4j:"renewalperiod" json:"renewalperiod,omitempty" desc:"Certificate renewal period" example:"6 weeks" capmodel:"ADObject"`
	ValidityPeriod                          string   `neo4j:"validityperiod" json:"validityperiod,omitempty" desc:"Certificate validity period" example:"1 year" capmodel:"ADObject"`
	OID                                     string   `neo4j:"oid" json:"oid,omitempty" desc:"Object identifier for the certificate template" example:"1.3.6.1.4.1.311.21.8.1234567.1234567.1.1.1" capmodel:"ADObject"`
	HomeDirectory                           string   `neo4j:"homedirectory" json:"homedirectory,omitempty" desc:"User's home directory path" example:"\\\\fileserver\\users\\jsmith" capmodel:"ADObject"`
	CertificatePolicy                       []string `neo4j:"certificatepolicy" json:"certificatepolicy,omitempty" desc:"Certificate policy OIDs" example:"[\"1.3.6.1.4.1.311.21.8.1\", \"1.3.6.1.5.5.7.2.1\"]" capmodel:"ADObject"`
	CertTemplateOID                         string   `neo4j:"certtemplateoid" json:"certtemplateoid,omitempty" desc:"Certificate template object identifier" example:"1.3.6.1.4.1.311.21.8.1234567.1234567.1.1.1" capmodel:"ADObject"`
	GroupLinkID                             string   `neo4j:"grouplinkid" json:"grouplinkid,omitempty" desc:"Link ID for group policy objects" example:"{31B2F340-016D-11D2-945F-00C04FB984F9}" capmodel:"ADObject"`
	ObjectGUID                              string   `neo4j:"objectguid" json:"objectguid,omitempty" desc:"Globally unique identifier for the AD object" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890" capmodel:"ADObject"`
	ExpirePasswordsOnSmartCardOnlyAccounts  bool     `neo4j:"expirepasswordsonsmartcardonlyaccounts" json:"expirepasswordsonsmartcardonlyaccounts,omitempty" desc:"Whether passwords expire for smart card only accounts" example:"false" capmodel:"ADObject"`
	MachineAccountQuota                     int      `neo4j:"machineaccountquota" json:"machineaccountquota,omitempty" desc:"Number of computer accounts a user can create" example:"10" capmodel:"ADObject"`
	SupportedKerberosEncryptionTypes        []string `neo4j:"supportedencryptiontypes" json:"supportedencryptiontypes,omitempty" desc:"Supported Kerberos encryption types" example:"[\"RC4_HMAC_MD5\", \"AES128_CTS_HMAC_SHA1_96\", \"AES256_CTS_HMAC_SHA1_96\"]" capmodel:"ADObject"`
	TGTDelegation                           string   `neo4j:"tgtdelegation" json:"tgtdelegation,omitempty" desc:"TGT delegation configuration" example:"Enabled" capmodel:"ADObject"`
	PasswordStoredUsingReversibleEncryption bool     `neo4j:"encryptedtextpwdallowed" json:"encryptedtextpwdallowed,omitempty" desc:"Password is stored using reversible encryption" example:"false" capmodel:"ADObject"`
	SmartcardRequired                       bool     `neo4j:"smartcardrequired" json:"smartcardrequired,omitempty" desc:"Smart card is required for interactive logon" example:"false" capmodel:"ADObject"`
	UseDESKeyOnly                           bool     `neo4j:"usedeskeyonly" json:"usedeskeyonly,omitempty" desc:"Use only DES encryption keys for Kerberos" example:"false" capmodel:"ADObject"`
	LogonScriptEnabled                      bool     `neo4j:"logonscriptenabled" json:"logonscriptenabled,omitempty" desc:"Logon script is enabled for the account" example:"true" capmodel:"ADObject"`
	LockedOut                               bool     `neo4j:"lockedout" json:"lockedout,omitempty" desc:"Account is currently locked out" example:"false" capmodel:"ADObject"`
	UserCannotChangePassword                bool     `neo4j:"passwordcantchange" json:"passwordcantchange,omitempty" desc:"User cannot change their password" example:"false" capmodel:"ADObject"`
	PasswordExpired                         bool     `neo4j:"passwordexpired" json:"passwordexpired,omitempty" desc:"Password has expired" example:"false" capmodel:"ADObject"`
	DSHeuristics                            string   `neo4j:"dsheuristics" json:"dsheuristics,omitempty" desc:"Directory Service heuristics configuration" example:"0000000001" capmodel:"ADObject"`
	UserAccountControl                      int      `neo4j:"useraccountcontrol" json:"useraccountcontrol,omitempty" desc:"User account control flags bitmask" example:"512" capmodel:"ADObject"`
	TrustAttributesInbound                  string   `neo4j:"trustattributesinbound" json:"trustattributesinbound,omitempty" desc:"Inbound trust attributes" example:"0x00000020" capmodel:"ADObject"`
	TrustAttributesOutbound                 string   `neo4j:"trustattributesoutbound" json:"trustattributesoutbound,omitempty" desc:"Outbound trust attributes" example:"0x00000020" capmodel:"ADObject"`
	MinPwdLength                            int      `neo4j:"minpwdlength" json:"minpwdlength,omitempty" desc:"Minimum password length requirement" example:"8" capmodel:"ADObject"`
	PwdProperties                           int      `neo4j:"pwdproperties" json:"pwdproperties,omitempty" desc:"Password policy properties bitmask" example:"1" capmodel:"ADObject"`
	PwdHistoryLength                        int      `neo4j:"pwdhistorylength" json:"pwdhistorylength,omitempty" desc:"Number of passwords remembered in history" example:"24" capmodel:"ADObject"`
	LockoutThreshold                        int      `neo4j:"lockoutthreshold" json:"lockoutthreshold,omitempty" desc:"Number of failed logon attempts before lockout" example:"5" capmodel:"ADObject"`
	MinPwdAge                               string   `neo4j:"minpwdage" json:"minpwdage,omitempty" desc:"Minimum password age" example:"1d" capmodel:"ADObject"`
	MaxPwdAge                               string   `neo4j:"maxpwdage" json:"maxpwdage,omitempty" desc:"Maximum password age" example:"90d" capmodel:"ADObject"`
	LockoutDuration                         string   `neo4j:"lockoutduration" json:"lockoutduration,omitempty" desc:"Account lockout duration" example:"30m" capmodel:"ADObject"`
	LockoutObservationWindow                int      `neo4j:"lockoutobservationwindow" json:"lockoutobservationwindow,omitempty" desc:"Time window in minutes for observing failed logon attempts" example:"30" capmodel:"ADObject"`
	OwnerSid                                string   `neo4j:"ownersid" json:"ownersid,omitempty" desc:"Security identifier of the object owner" example:"S-1-5-21-3623811015-3361044348-30300820-1001" capmodel:"ADObject"`
	SMBSigning                              bool     `neo4j:"smbsigning" json:"smbsigning,omitempty" desc:"SMB signing is required" example:"true" capmodel:"ADObject"`
	WebClientRunning                        bool     `neo4j:"webclientrunning" json:"webclientrunning,omitempty" desc:"Whether WebDAV client service is running" example:"true" capmodel:"ADObject"`
	RestrictOutboundNTLM                    bool     `neo4j:"restrictoutboundntlm" json:"restrictoutboundntlm,omitempty" desc:"Outbound NTLM authentication is restricted" example:"false" capmodel:"ADObject"`
	GMSA                                    bool     `neo4j:"gmsa" json:"gmsa,omitempty" desc:"Group Managed Service Account" example:"true" capmodel:"ADObject"`
	MSA                                     bool     `neo4j:"msa" json:"msa,omitempty" desc:"Managed Service Account" example:"false" capmodel:"ADObject"`
	DoesAnyAceGrantOwnerRights              bool     `neo4j:"doesanyacegrantownerrights" json:"doesanyacegrantownerrights,omitempty" desc:"Whether any ACE grants owner rights" example:"true" capmodel:"ADObject"`
	DoesAnyInheritedAceGrantOwnerRights     bool     `neo4j:"doesanyinheritedacegrantownerrights" json:"doesanyinheritedacegrantownerrights,omitempty" desc:"Whether any inherited ACE grants owner rights" example:"false" capmodel:"ADObject"`
	ADCSWebEnrollmentHTTP                   string   `neo4j:"adcswebenrollmenthttp" json:"adcswebenrollmenthttp,omitempty" desc:"ADCS web enrollment HTTP endpoint availability" example:"http://ca.contoso.local/certsrv" capmodel:"ADObject"`
	ADCSWebEnrollmentHTTPS                  string   `neo4j:"adcswebenrollmenthttps" json:"adcswebenrollmenthttps,omitempty" desc:"ADCS web enrollment HTTPS endpoint availability" example:"https://ca.contoso.local/certsrv" capmodel:"ADObject"`
	ADCSWebEnrollmentHTTPSEPA               string   `neo4j:"adcswebenrollmenthttpsepa" json:"adcswebenrollmenthttpsepa,omitempty" desc:"ADCS web enrollment HTTPS with Extended Protection" example:"https://ca.contoso.local/certsrv" capmodel:"ADObject"`
	LDAPSigning                             bool     `neo4j:"ldapsigning" json:"ldapsigning,omitempty" desc:"LDAP signing requirement" example:"Required" capmodel:"ADObject"`
	LDAPAvailable                           bool     `neo4j:"ldapavailable" json:"ldapavailable,omitempty" desc:"Whether LDAP service is available" example:"true" capmodel:"ADObject"`
	LDAPSAvailable                          bool     `neo4j:"ldapsavailable" json:"ldapsavailable,omitempty" desc:"Whether LDAPS (secure LDAP) is available" example:"true" capmodel:"ADObject"`
	LDAPSEPA                                bool     `neo4j:"ldapsepa" json:"ldapsepa,omitempty" desc:"LDAPS with Extended Protection for Authentication" example:"Enabled" capmodel:"ADObject"`
	IsDC                                    bool     `neo4j:"isdc" json:"isdc,omitempty" desc:"Whether computer is a Domain Controller" example:"true" capmodel:"ADObject"`
	IsReadOnlyDC                            bool     `neo4j:"isreadonlydc" json:"isreadonlydc,omitempty" desc:"Whether computer is a Read-Only Domain Controller" example:"false" capmodel:"ADObject"`
	HTTPEnrollmentEndpoints                 string   `neo4j:"httpenrollmentendpoints" json:"httpenrollmentendpoints,omitempty" desc:"List of HTTP certificate enrollment endpoints" example:"[\"http://ca1.contoso.local/certsrv\", \"http://ca2.contoso.local/certsrv\"]" capmodel:"ADObject"`
	HTTPSEnrollmentEndpoints                string   `neo4j:"httpsenrollmentendpoints" json:"httpsenrollmentendpoints,omitempty" desc:"List of HTTPS certificate enrollment endpoints" example:"[\"https://ca1.contoso.local/certsrv\", \"https://ca2.contoso.local/certsrv\"]" capmodel:"ADObject"`
	HasVulnerableEndpoint                   bool     `neo4j:"hasvulnerableendpoint" json:"hasvulnerableendpoint,omitempty" desc:"Whether object has vulnerable enrollment endpoints" example:"true" capmodel:"ADObject"`
	RequireSecuritySignature                bool     `neo4j:"requiresecuritysignature" json:"requiresecuritysignature,omitempty" desc:"Whether security signature is required" example:"true" capmodel:"ADObject"`
	EnableSecuritySignature                 bool     `neo4j:"enablesecuritysignature" json:"enablesecuritysignature,omitempty" desc:"Whether security signature is enabled" example:"true" capmodel:"ADObject"`
	RestrictReceivingNTLMTraffic            bool     `neo4j:"restrictreceivingntmltraffic" json:"restrictreceivingntmltraffic,omitempty" desc:"Restriction policy for receiving NTLM traffic" example:"DenyAll" capmodel:"ADObject"`
	NTLMMinServerSec                        int      `neo4j:"ntlmminserversec" json:"ntlmminserversec,omitempty" desc:"Minimum security level for NTLM SSP server" example:"537395200" capmodel:"ADObject"`
	NTLMMinClientSec                        int      `neo4j:"ntlmminclientsec" json:"ntlmminclientsec,omitempty" desc:"Minimum security level for NTLM SSP client" example:"537395200" capmodel:"ADObject"`
	LMCompatibilityLevel                    string   `neo4j:"lmcompatibilitylevel" json:"lmcompatibilitylevel,omitempty" desc:"LAN Manager authentication compatibility level" example:"5" capmodel:"ADObject"`
	UseMachineID                            string   `neo4j:"usemachineid" json:"usemachineid,omitempty" desc:"Whether to use machine identity for authentication" example:"true" capmodel:"ADObject"`
	ClientAllowedNTLMServers                string   `neo4j:"clientallowedntlmservers" json:"clientallowedntlmservers,omitempty" desc:"List of servers allowed to use NTLM authentication" example:"*.contoso.local" capmodel:"ADObject"`
	Transitive                              string   `neo4j:"transitive" json:"transitive,omitempty" desc:"Whether trust relationship is transitive" example:"true" capmodel:"ADObject"`
	GroupScope                              string   `neo4j:"groupscope" json:"groupscope,omitempty" desc:"Scope of the AD group" example:"Global" capmodel:"ADObject"`
	NetBIOS                                 string   `neo4j:"netbios" json:"netbios,omitempty" desc:"NetBIOS name of the domain" example:"CONTOSO" capmodel:"ADObject"`
	AdminSDHolderProtected                  string   `neo4j:"adminsdholderprotected" json:"adminsdholderprotected,omitempty" desc:"Whether object is protected by AdminSDHolder process" example:"true" capmodel:"ADObject"`
	ServicePrincipalNames                   []string `neo4j:"serviceprincipalnames" json:"serviceprincipalnames,omitempty" desc:"The service principal name(s) associated with this account" example:"WSMAN/database" capmodel:"ADObject"`
	OperatingSystem                         string   `neo4j:"operatingsystem" json:"operatingsystem,omitempty" desc:"The operating system associated with this computer" example:"Windows Server 2019 SE" capmodel:"ADObject"`
}

func (ad *ADProperties) Visit(other ADProperties) {
	marshaled, _ := json.Marshal(other)
	json.Unmarshal(marshaled, ad)
}
