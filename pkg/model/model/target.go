package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// Target represents any entity that can be the subject of a security capability or job.
// This includes assets (domains, IPs), attributes (ports, services), preseeds, webpages,
// cloud resources, and other scannable entities in the Chariot system.
//
// The Target interface provides a standard way to identify, classify, and manage the
// lifecycle of entities that capabilities can operate on.
type Target interface {
	GraphModel

	// GetStatus returns the current operational status of the target.
	// Common values include:
	//   - "A" (Active): Target is active and ready for scanning
	//   - "P" (Pending): Target is pending initial processing
	//   - "AL", "AH" (ActiveLow, ActiveHigh): Different scanning intensity levels
	//   - "D" (Deleted): Target has been marked for deletion
	//
	// The status determines which capabilities will accept and process this target.
	GetStatus() string

	// WithStatus creates a new copy of the target with the specified status.
	// This method must preserve the concrete type to avoid type erasure issues
	// when the result is used in capability matching and job processing.
	//
	// Example: target.WithStatus("AH") changes status to ActiveHigh for intensive scanning.
	WithStatus(string) Target

	// Group returns a grouping identifier used with Identifier() to determine job uniqueness.
	// The combination of Group() and Identifier() creates a unique job key.
	//
	// Examples:
	//   - Asset: Returns DNS field (e.g., "example.com")
	//   - Attribute: Returns parent asset's DNS (e.g., "example.com")
	//   - Preseed: Returns Type field (e.g., "whois")
	//   - CloudResource: Returns a resource-specific grouping
	//
	// This is used in job key generation: "#job#{Group()}#{Identifier()}#{capability}"
	Group() string

	// Identifier returns a unique identifier within the Group() for job deduplication.
	// The combination of Group() and Identifier() ensures each job is unique.
	//
	// Examples:
	//   - Asset: Returns Name field (e.g., "example.com" or "192.168.1.1")
	//   - Attribute: Returns formatted target URL (e.g., "https://example.com:443")
	//   - Preseed: Returns Value field (e.g., "admin@example.com")
	//   - CloudResource: Returns the resource name/ARN
	//
	// This prevents duplicate jobs from being created for the same target+capability.
	Identifier() string

	// IsStatus checks if the target's status starts with the given prefix.
	// This allows hierarchical status matching where "A" matches "AL", "AH", etc.
	//
	// Examples:
	//   - IsStatus("A") matches "A", "AL", "AH", "AP" (all Active variants)
	//   - IsStatus("P") matches "P" (Pending)
	//   - IsStatus("Active") matches "ActiveHigh", "ActiveLow"
	//
	// Used extensively in capability acceptance logic and status filtering.
	IsStatus(string) bool

	// IsClass checks if the target belongs to a specific classification or type.
	// The classification system varies by target type but enables capability matching.
	//
	// Examples:
	//   - Asset: Checks Class field - IsClass("ipv4"), IsClass("domain"), IsClass("s3")
	//   - Attribute: Checks Name field - IsClass("http"), IsClass("ssh"), IsClass("port")
	//   - CloudResource: Checks ResourceType - IsClass("ec2"), IsClass("s3")
	//   - Preseed: Checks Type field - IsClass("whois"), IsClass("edgar")
	//
	// Capabilities use this to determine if they can process a target type.
	IsClass(string) bool

	// IsPrivate determines if the target represents an internal/private resource.
	// Private targets are typically excluded from external scanning capabilities
	// unless specifically configured to include internal resources.
	//
	// Examples:
	//   - Asset: true for private IP ranges (10.x.x.x, 192.168.x.x) or internal domains (.local)
	//   - Attribute: delegates to parent asset's IsPrivate() method
	//   - CloudResource: typically false (cloud resources are considered accessible)
	//   - Webpage: true if parent target is private
	//
	// Used in capability matching to prevent scanning internal infrastructure.
	IsPrivate() bool
}

type TargetWrapper registry.Wrapper[Target]

func (t TargetWrapper) MarshalJSON() ([]byte, error) {
	return (registry.Wrapper[Target])(t).MarshalJSON()
}

func (t *TargetWrapper) UnmarshalJSON(data []byte) error {
	return (*registry.Wrapper[Target])(t).UnmarshalJSON(data)
}

func (t TargetWrapper) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return (registry.Wrapper[Target])(t).MarshalDynamoDBAttributeValue()
}

func (t *TargetWrapper) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	return (*registry.Wrapper[Target])(t).UnmarshalDynamoDBAttributeValue(av)
}
