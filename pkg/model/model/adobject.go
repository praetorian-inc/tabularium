package model

import (
	"strings"
)

// ADObject serves as a base type for Active Directory objects.
// This type provides common AD properties and methods but does not
// satisfy registry.Model as it's intended for embedding in other types.
type ADObject struct {
	// Core AD Properties
	Domain            string `neo4j:"domain" json:"domain" desc:"AD domain this object belongs to." example:"example.local"`
	DistinguishedName string `neo4j:"distinguishedName" json:"distinguishedName" desc:"Full distinguished name in AD." example:"CN=John Doe,CN=Users,DC=example,DC=local"`
	SID               string `neo4j:"sid" json:"sid" desc:"Security identifier." example:"S-1-5-21-123456789-123456789-123456789-1001"`
	ObjectClass       string `neo4j:"objectClass" json:"objectClass" desc:"AD object class." example:"user"`
	Name              string `neo4j:"name" json:"name" desc:"Common name of the object." example:"John Doe"`
	SAMAccountName    string `neo4j:"samAccountName,omitempty" json:"samAccountName,omitempty" desc:"SAM account name (for users/computers)." example:"jdoe"`
	DisplayName       string `neo4j:"displayName,omitempty" json:"displayName,omitempty" desc:"Display name of the object." example:"John Doe"`
	Description       string `neo4j:"description,omitempty" json:"description,omitempty" desc:"Description of the object." example:"User account for John Doe"`
}

// NewADObject creates a new ADObject with the specified domain, distinguished name, and object class
func NewADObject(domain, distinguishedName, objectClass string) ADObject {
	ad := ADObject{
		Domain:            domain,
		DistinguishedName: distinguishedName,
		ObjectClass:       objectClass,
	}
	
	// Extract the common name from the DN if available
	ad.Name = ad.GetCommonName()
	
	return ad
}

// IsClass checks if the AD object is of the specified object class
func (ad *ADObject) IsClass(objectClass string) bool {
	return strings.EqualFold(ad.ObjectClass, objectClass)
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
	
	// Find the first comma to get the parent DN
	if idx := strings.Index(ad.DistinguishedName, ","); idx != -1 {
		return strings.TrimSpace(ad.DistinguishedName[idx+1:])
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
	if ad.DistinguishedName == "" {
		return ad.Name
	}
	
	// Extract CN= value from the beginning of the DN
	if strings.HasPrefix(strings.ToUpper(ad.DistinguishedName), "CN=") {
		if idx := strings.Index(ad.DistinguishedName, ","); idx != -1 {
			return ad.DistinguishedName[3:idx] // Remove "CN=" prefix
		}
		return ad.DistinguishedName[3:] // No comma found, return everything after CN=
	}
	
	return ad.Name
}