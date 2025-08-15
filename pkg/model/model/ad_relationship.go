package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&ADRelationship{})
}

// AD Relationship type labels (matching BloodHound exactly)
const (
	// Ownership and Control
	ADOwnsLabel         = "Owns"
	ADGenericAllLabel   = "GenericAll"
	ADGenericWriteLabel = "GenericWrite"
	ADWriteOwnerLabel   = "WriteOwner"
	ADWriteDACLLabel    = "WriteDacl"

	// Group Membership
	ADMemberOfLabel = "MemberOf"

	// Password Control
	ADForceChangePasswordLabel = "ForceChangePassword"

	// Extended Rights
	ADAllExtendedRightsLabel = "AllExtendedRights"
	ADAddMemberLabel         = "AddMember"

	// Sessions and Access
	ADHasSessionLabel = "HasSession"

	// Container Relationships
	ADContainsLabel = "Contains"

	// GPO Relationships
	ADGPLinkLabel = "GPLink"

	// Delegation
	ADAllowedToDelegateLabel = "AllowedToDelegate"
	ADCoerceToTGTLabel       = "CoerceToTGT"

	// Replication Rights
	ADGetChangesLabel              = "GetChanges"
	ADGetChangesAllLabel           = "GetChangesAll"
	ADGetChangesInFilteredSetLabel = "GetChangesInFilteredSet"

	// Trust Relationships
	ADCrossForestTrustLabel   = "CrossForestTrust"
	ADSameForestTrustLabel    = "SameForestTrust"
	ADSpoofSIDHistoryLabel    = "SpoofSIDHistory"
	ADAbuseTGTDelegationLabel = "AbuseTGTDelegation"

	// Resource-Based Constrained Delegation
	ADAllowedToActLabel = "AllowedToAct"

	// Administrative Access
	ADAdminToLabel     = "AdminTo"
	ADCanPSRemoteLabel = "CanPSRemote"
	ADCanRDPLabel      = "CanRDP"
	ADExecuteDCOMLabel = "ExecuteDCOM"

	// SID History
	ADHasSIDHistoryLabel = "HasSIDHistory"

	// Self Rights
	ADAddSelfLabel = "AddSelf"

	// DCSync
	ADDCSyncLabel = "DCSync"

	// Password Reading
	ADReadLAPSPasswordLabel = "ReadLAPSPassword"
	ADReadGMSAPasswordLabel = "ReadGMSAPassword"
	ADDumpSMSAPasswordLabel = "DumpSMSAPassword"

	// SQL
	ADSQLAdminLabel = "SQLAdmin"

	// Specific Write Rights
	ADAddAllowedToActLabel      = "AddAllowedToAct"
	ADWriteSPNLabel             = "WriteSPN"
	ADAddKeyCredentialLinkLabel = "AddKeyCredentialLink"

	// Local Group Membership
	ADLocalToComputerLabel             = "LocalToComputer"
	ADMemberOfLocalGroupLabel          = "MemberOfLocalGroup"
	ADRemoteInteractiveLogonRightLabel = "RemoteInteractiveLogonRight"

	// LAPS
	ADSyncLAPSPasswordLabel = "SyncLAPSPassword"

	// Write Permissions
	ADWriteAccountRestrictionsLabel = "WriteAccountRestrictions"
	ADWriteGPLinkLabel              = "WriteGPLink"

	// Certificate Authority Relationships
	ADRootCAForLabel                = "RootCAFor"
	ADDCForLabel                    = "DCFor"
	ADPublishedToLabel              = "PublishedTo"
	ADManageCertificatesLabel       = "ManageCertificates"
	ADManageCALabel                 = "ManageCA"
	ADDelegatedEnrollmentAgentLabel = "DelegatedEnrollmentAgent"
	ADEnrollLabel                   = "Enroll"
	ADHostsCAServiceLabel           = "HostsCAService"
	ADWritePKIEnrollmentFlagLabel   = "WritePKIEnrollmentFlag"
	ADWritePKINameFlagLabel         = "WritePKINameFlag"
	ADNTAuthStoreForLabel           = "NTAuthStoreFor"
	ADTrustedForNTAuthLabel         = "TrustedForNTAuth"
	ADEnterpriseCAForLabel          = "EnterpriseCAFor"
	ADIssuedSignedByLabel           = "IssuedSignedBy"
	ADGoldenCertLabel               = "GoldenCert"
	ADEnrollOnBehalfOfLabel         = "EnrollOnBehalfOf"

	// Certificate Template Links
	ADOIDGroupLinkLabel     = "OIDGroupLink"
	ADExtendedByPolicyLabel = "ExtendedByPolicy"

	// ADCS Attack Paths
	ADADCSESC1Label   = "ADCSESC1"
	ADADCSESC2Label   = "ADCSESC2"
	ADADCSESC3Label   = "ADCSESC3"
	ADADCSESC4Label   = "ADCSESC4"
	ADADCSESC6aLabel  = "ADCSESC6a"
	ADADCSESC6bLabel  = "ADCSESC6b"
	ADADCSESC9aLabel  = "ADCSESC9a"
	ADADCSESC9bLabel  = "ADCSESC9b"
	ADADCSESC10aLabel = "ADCSESC10a"
	ADADCSESC10bLabel = "ADCSESC10b"
	ADADCSESC13Label  = "ADCSESC13"

	// Azure Sync
	ADSyncedToEntraUserLabel = "SyncedToEntraUser"

	// NTLM Coercion and Relay
	ADCoerceAndRelayNTLMToSMBLabel   = "CoerceAndRelayNTLMToSMB"
	ADCoerceAndRelayNTLMToADCSLabel  = "CoerceAndRelayNTLMToADCS"
	ADCoerceAndRelayNTLMToLDAPLabel  = "CoerceAndRelayNTLMToLDAP"
	ADCoerceAndRelayNTLMToLDAPSLabel = "CoerceAndRelayNTLMToLDAPS"

	// Limited Rights Variants
	ADWriteOwnerLimitedRightsLabel = "WriteOwnerLimitedRights"
	ADWriteOwnerRawLabel           = "WriteOwnerRaw"
	ADOwnsLimitedRightsLabel       = "OwnsLimitedRights"
	ADOwnsRawLabel                 = "OwnsRaw"

	// Special Identity
	ADClaimSpecialIdentityLabel = "ClaimSpecialIdentity"

	// Identity and ACE Propagation
	ADContainsIdentityLabel = "ContainsIdentity"
	ADPropagatesACEsToLabel = "PropagatesACEsTo"

	// GPO Application
	ADGPOAppliesToLabel = "GPOAppliesTo"
	ADCanApplyGPOLabel  = "CanApplyGPO"

	// Trust Keys
	ADHasTrustKeysLabel = "HasTrustKeys"
)

// ADRelationship represents an Active Directory relationship between two AD objects
type ADRelationship struct {
	*BaseRelationship

	// RelationshipType determines the Neo4j relationship label
	RelationshipType string `neo4j:"-" json:"relationshipType"`
}

// GetDescription returns a description for the ADRelationship model.
func (ar *ADRelationship) GetDescription() string {
	return "Represents an Active Directory relationship between two AD objects, supporting all BloodHound relationship types."
}

// Label returns the Neo4j relationship label based on the relationship type
func (ar *ADRelationship) Label() string {
	return ar.RelationshipType
}

func NewADRelationship(source, target GraphModel, label string) GraphRelationship {
	return &ADRelationship{
		RelationshipType: label,
		BaseRelationship: NewBaseRelationship(source, target, label),
	}
}
