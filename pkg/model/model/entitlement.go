package model

// Entitlement represents a granular permission that can be checked at an endpoint.
// Roles are composed of entitlements. Endpoints require a single entitlement.
type Entitlement string

const (
	// EntitlementRead allows reading data: exports, files, recovery codes, auth, etc.
	EntitlementRead Entitlement = "read"
	// EntitlementWriteAssets allows creating/modifying assets, seeds, risks, jobs, etc.
	EntitlementWriteAssets Entitlement = "write_assets"
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
	// EntitlementManageOSINT allows OSINT operations.
	EntitlementManageOSINT Entitlement = "manage_osint"
	// EntitlementManageModels allows AI model management, functions, templates.
	EntitlementManageModels Entitlement = "manage_models"
)

// allEntitlements is the complete list of built-in entitlements.
var allEntitlements = []Entitlement{
	EntitlementRead,
	EntitlementWriteAssets,
	EntitlementManageAccounts,
	EntitlementManageSettings,
	EntitlementManageIntegrations,
	EntitlementManageAgents,
	EntitlementManageRedteam,
	EntitlementManageOSINT,
	EntitlementManageModels,
}

// roleEntitlements maps each built-in role to its entitlements.
// Admin has all entitlements. Analyst has read + write. ReadOnly has read only.
var roleEntitlements = map[Role][]Entitlement{
	RoleReadOnly: {
		EntitlementRead,
	},
	RoleAnalyst: {
		EntitlementRead,
		EntitlementWriteAssets,
	},
	RoleAdmin: allEntitlements,
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
