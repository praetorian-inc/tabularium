package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// AD Relationship type labels (matching BloodHound exactly)
const (
	// Domain Trusts
	ADSameForestTrust  = "SameForestTrust"
	ADCrossForestTrust = "CrossForestTrust"

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

var ADRelationshipLabels = []string{
	ADOwnsLabel,
	ADGenericAllLabel,
	ADGenericWriteLabel,
	ADWriteOwnerLabel,
	ADWriteDACLLabel,
	ADMemberOfLabel,
	ADForceChangePasswordLabel,
	ADAllExtendedRightsLabel,
	ADAddMemberLabel,
	ADHasSessionLabel,
	ADContainsLabel,
	ADGPLinkLabel,
	ADAllowedToDelegateLabel,
	ADCoerceToTGTLabel,
	ADGetChangesLabel,
	ADGetChangesAllLabel,
	ADGetChangesInFilteredSetLabel,
	ADCrossForestTrustLabel,
	ADSameForestTrustLabel,
	ADSpoofSIDHistoryLabel,
	ADAbuseTGTDelegationLabel,
	ADAllowedToActLabel,
	ADAdminToLabel,
	ADCanPSRemoteLabel,
	ADCanRDPLabel,
	ADExecuteDCOMLabel,
	ADHasSIDHistoryLabel,
	ADAddSelfLabel,
	ADDCSyncLabel,
	ADReadLAPSPasswordLabel,
	ADReadGMSAPasswordLabel,
	ADDumpSMSAPasswordLabel,
	ADSQLAdminLabel,
	ADAddAllowedToActLabel,
	ADWriteSPNLabel,
	ADAddKeyCredentialLinkLabel,
	ADLocalToComputerLabel,
	ADMemberOfLocalGroupLabel,
	ADRemoteInteractiveLogonRightLabel,
	ADSyncLAPSPasswordLabel,
	ADWriteAccountRestrictionsLabel,
	ADWriteGPLinkLabel,
	ADRootCAForLabel,
	ADDCForLabel,
	ADPublishedToLabel,
	ADManageCertificatesLabel,
	ADManageCALabel,
	ADDelegatedEnrollmentAgentLabel,
	ADEnrollLabel,
	ADHostsCAServiceLabel,
	ADWritePKIEnrollmentFlagLabel,
	ADWritePKINameFlagLabel,
	ADNTAuthStoreForLabel,
	ADTrustedForNTAuthLabel,
	ADEnterpriseCAForLabel,
	ADIssuedSignedByLabel,
	ADGoldenCertLabel,
	ADEnrollOnBehalfOfLabel,
	ADOIDGroupLinkLabel,
	ADExtendedByPolicyLabel,
	ADADCSESC1Label,
	ADADCSESC3Label,
	ADADCSESC4Label,
	ADADCSESC6aLabel,
	ADADCSESC6bLabel,
	ADADCSESC9aLabel,
	ADADCSESC9bLabel,
	ADADCSESC10aLabel,
	ADADCSESC10bLabel,
	ADADCSESC13Label,
	ADSyncedToEntraUserLabel,
	ADCoerceAndRelayNTLMToSMBLabel,
	ADCoerceAndRelayNTLMToADCSLabel,
	ADCoerceAndRelayNTLMToLDAPLabel,
	ADCoerceAndRelayNTLMToLDAPSLabel,
	ADWriteOwnerLimitedRightsLabel,
	ADWriteOwnerRawLabel,
	ADOwnsLimitedRightsLabel,
	ADOwnsRawLabel,
	ADClaimSpecialIdentityLabel,
	ADContainsIdentityLabel,
	ADPropagatesACEsToLabel,
	ADGPOAppliesToLabel,
	ADCanApplyGPOLabel,
	ADHasTrustKeysLabel,
}

func init() {
	registry.Registry.MustRegisterModel(&ADRelationship{}, ADRelationshipLabels...)
}

type ADRelationship struct {
	*BaseRelationship
	RelationshipType string       `neo4j:"relationshipType" json:"relationshipType"`
	Enforced         *GobSafeBool `neo4j:"enforced" json:"enforced,omitempty" desc:"Whether GPO link is enforced (no override). Only applicable to GPLink relationships" example:"true"`
}

func (ar *ADRelationship) GetDescription() string {
	return "Represents an Active Directory relationship between two AD objects, supporting all BloodHound relationship types."
}

func (ar *ADRelationship) Label() string {
	return ar.RelationshipType
}

func (ar *ADRelationship) Visit(o GraphRelationship) {
	other, ok := o.(*ADRelationship)
	if !ok {
		return
	}

	if other.Enforced != nil {
		ar.Enforced = other.Enforced
	}

	if ar.RelationshipType == "" && other.RelationshipType != "" {
		ar.RelationshipType = other.RelationshipType
	}

	ar.BaseRelationship.Visit(other)
}

func (ar *ADRelationship) SetEnforced(value bool) {
	gsb := GobSafeBool(value)
	ar.Enforced = &gsb
}

func NewADRelationship(source, target GraphModel, label string) GraphRelationship {
	return &ADRelationship{
		RelationshipType: label,
		BaseRelationship: NewBaseRelationship(source, target, label),
	}
}
