package model

// Entitlement represents a granular permission that can be checked at an endpoint.
// Roles are composed of entitlements. Endpoints require a single entitlement.
type Entitlement string

const (
	// EntitlementRead allows reading data: exports, files, recovery codes, auth, etc.
	EntitlementRead Entitlement = "read"
	// EntitlementWriteAssets allows creating/modifying assets, seeds, risks, etc.
	EntitlementWriteAssets Entitlement = "write_assets"
	// EntitlementWriteJobs allows capability scheduling and job execution.
	EntitlementWriteJobs Entitlement = "write_jobs"
	// EntitlementWriteFiles allows uploading, deleting, and managing files.
	EntitlementWriteFiles Entitlement = "write_files"
	// EntitlementConversationAI allows access to conversation AI endpoints.
	EntitlementConversationAI Entitlement = "conversation_ai"
	// EntitlementManageAccounts allows account CRUD, purge, user management.
	EntitlementManageAccounts Entitlement = "manage_accounts"
	// EntitlementManageSettings allows settings, configurations, flags, keys, tokens.
	EntitlementManageSettings Entitlement = "manage_settings"
	// EntitlementManageIntegrations allows integration validation, configuration, broker.
	EntitlementManageIntegrations Entitlement = "manage_integrations"
	// EntitlementManageAgents allows aegis management, cloud initialization.
	EntitlementManageAgents Entitlement = "manage_agents"
	// EntitlementManageRedteam allows red team deployments, domain parking, payloads.
	EntitlementManageRedteam Entitlement = "manage_redteam"
	// EntitlementPraetorian gates all Praetorian-only endpoints.
	// This entitlement is NOT assigned to any role statically.
	// It is dynamically granted to Praetorian users who are also admins.
	EntitlementPraetorian Entitlement = "praetorian"
)

// allEntitlements is the complete list of built-in entitlements.
var allEntitlements = []Entitlement{
	EntitlementRead,
	EntitlementWriteAssets,
	EntitlementWriteJobs,
	EntitlementWriteFiles,
	EntitlementConversationAI,
	EntitlementManageAccounts,
	EntitlementManageSettings,
	EntitlementManageIntegrations,
	EntitlementManageAgents,
	EntitlementManageRedteam,
	EntitlementPraetorian,
}

// roleEntitlements maps each built-in role to its entitlements.
// Admin has all entitlements except praetorian (which is granted dynamically).
// Analyst has read + write_assets + write_jobs + write_files + conversation_ai.
// ReadOnly has read only.
var roleEntitlements = map[Role][]Entitlement{
	RoleReadOnly: {
		EntitlementRead,
	},
	RoleAnalyst: {
		EntitlementRead,
		EntitlementWriteAssets,
		EntitlementWriteJobs,
		EntitlementWriteFiles,
		EntitlementConversationAI,
	},
	RoleAdmin: {
		EntitlementRead,
		EntitlementWriteAssets,
		EntitlementWriteJobs,
		EntitlementWriteFiles,
		EntitlementConversationAI,
		EntitlementManageAccounts,
		EntitlementManageSettings,
		EntitlementManageIntegrations,
		EntitlementManageAgents,
		EntitlementManageRedteam,
	},
}

// Valid returns true if e is a recognized built-in entitlement.
func (e Entitlement) Valid() bool {
	for _, ent := range allEntitlements {
		if e == ent {
			return true
		}
	}
	return false
}

// EntitlementsForRole returns the entitlements associated with the given role.
// Returns nil for unrecognized roles.
func EntitlementsForRole(role Role) []Entitlement {
	return roleEntitlements[role]
}

// RoleHasEntitlement returns true if the given role includes the specified entitlement.
func RoleHasEntitlement(role Role, entitlement Entitlement) bool {
	for _, e := range roleEntitlements[role] {
		if e == entitlement {
			return true
		}
	}
	return false
}

// AllEntitlements returns the complete list of built-in entitlements.
func AllEntitlements() []Entitlement {
	result := make([]Entitlement, len(allEntitlements))
	copy(result, allEntitlements)
	return result
}
