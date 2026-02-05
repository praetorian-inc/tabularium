package external

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// Transformer is a function that converts an external type to a Tabularium model.
type Transformer func(any) (registry.Model, error)

// transformers maps external type names to their transformer functions.
var transformers = make(map[reflect.Type]Transformer)

// RegisterTransformer registers a custom transformer for an external type.
func RegisterTransformer[T any](fn func(T) (registry.Model, error)) {
	var zero T
	transformers[reflect.TypeOf(zero)] = func(v any) (registry.Model, error) {
		return fn(v.(T))
	}
}

// Convert transforms an external type into its corresponding full Tabularium type.
// It uses the following process:
// 1. Check for a custom transformer (for entity types like IP, Domain)
// 2. Otherwise, JSON marshal the external type
// 3. Create the target Tabularium model
// 4. Call Defaulted() on the model
// 5. JSON unmarshal into the model
// 6. Call hooks on the model
//
// This ensures all computed fields (Key, Class, etc.) are properly set.
func Convert[T registry.Model](external any) (T, error) {
	var zero T

	// Check for custom transformer first
	extType := reflect.TypeOf(external)
	if extType.Kind() == reflect.Ptr {
		extType = extType.Elem()
	}

	if transformer, ok := transformers[extType]; ok {
		result, err := transformer(external)
		if err != nil {
			return zero, err
		}
		if typed, ok := result.(T); ok {
			return typed, nil
		}
		return zero, fmt.Errorf("transformer returned wrong type: expected %T, got %T", zero, result)
	}

	// Standard conversion via JSON round-trip
	return convertViaJSON[T](external)
}

// convertViaJSON converts using JSON marshal/unmarshal with defaults and hooks.
func convertViaJSON[T registry.Model](external any) (T, error) {
	var zero T

	// Marshal external type to JSON
	jsonBytes, err := json.Marshal(external)
	if err != nil {
		return zero, fmt.Errorf("failed to marshal external type: %w", err)
	}

	// Create target model instance
	targetType := reflect.TypeOf(zero)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	// Create new instance
	targetPtr := reflect.New(targetType)
	target := targetPtr.Interface().(registry.Model)

	// Apply defaults first
	target.Defaulted()

	// Unmarshal JSON into target
	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return zero, fmt.Errorf("failed to unmarshal into target type: %w", err)
	}

	// Call hooks
	if err := registry.CallHooks(target); err != nil {
		return zero, fmt.Errorf("failed to call hooks: %w", err)
	}

	// Type assert back to T
	if typed, ok := target.(T); ok {
		return typed, nil
	}

	return zero, fmt.Errorf("type assertion failed: expected %T", zero)
}

// ConvertToModelByName converts an external type to a registry.Model by model name.
// The modelName should match a registered model (e.g., "asset", "risk", "job").
func ConvertToModelByName(modelName string, external any) (registry.Model, error) {
	// Get the target type from registry
	targetType, ok := registry.Registry.GetType(modelName)
	if !ok {
		return nil, fmt.Errorf("model type not registered: %s", modelName)
	}

	// Check for custom transformer first
	extType := reflect.TypeOf(external)
	if extType.Kind() == reflect.Ptr {
		extType = extType.Elem()
	}

	if transformer, ok := transformers[extType]; ok {
		return transformer(external)
	}

	// Create target instance
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}
	targetPtr := reflect.New(targetType)
	target := targetPtr.Interface().(registry.Model)

	// Marshal external type to JSON
	jsonBytes, err := json.Marshal(external)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal external type: %w", err)
	}

	// Apply defaults
	target.Defaulted()

	// Unmarshal
	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into target type: %w", err)
	}

	// Call hooks
	if err := registry.CallHooks(target); err != nil {
		return nil, fmt.Errorf("failed to call hooks: %w", err)
	}

	return target, nil
}

// --- Entity Types for Semantic Transforms ---

// IP represents an IP address for conversion to an Asset.
type IP struct {
	Address string `json:"address"` // IPv4 or IPv6 address
	Domain  string `json:"domain"`  // Optional associated domain (defaults to Address)
}

// Domain represents a domain for conversion to an Asset.
type Domain struct {
	Name string `json:"name"` // Domain name (e.g., "example.com")
}

// CIDR represents a CIDR block for conversion to an Asset.
type CIDR struct {
	Block  string `json:"block"`  // CIDR notation (e.g., "10.0.0.0/8")
	Domain string `json:"domain"` // Optional associated domain (defaults to Block)
}

func init() {
	// Register entity transformers
	RegisterTransformer(transformIP)
	RegisterTransformer(transformDomain)
	RegisterTransformer(transformCIDR)
}

// transformIP converts an IP to an Asset.
func transformIP(ip IP) (registry.Model, error) {
	// Validate IP address
	parsed := net.ParseIP(ip.Address)
	if parsed == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip.Address)
	}

	// Default domain to address if not provided
	domain := ip.Domain
	if domain == "" {
		domain = ip.Address
	}

	asset := model.NewAsset(domain, ip.Address)
	return &asset, nil
}

// transformDomain converts a Domain to an Asset.
func transformDomain(d Domain) (registry.Model, error) {
	// Basic validation
	if d.Name == "" {
		return nil, fmt.Errorf("domain name is required")
	}

	// Normalize: remove protocol prefix if present
	name := strings.TrimPrefix(d.Name, "https://")
	name = strings.TrimPrefix(name, "http://")
	name = strings.TrimSuffix(name, "/")

	asset := model.NewAsset(name, name)
	return &asset, nil
}

// transformCIDR converts a CIDR to an Asset.
// Note: For CIDR class detection to work, the CIDR must be in the DNS field.
// The Name field can be the same or a more specific identifier within the block.
func transformCIDR(c CIDR) (registry.Model, error) {
	// Validate CIDR
	_, _, err := net.ParseCIDR(c.Block)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR block: %s: %w", c.Block, err)
	}

	// For CIDR assets, DNS must contain the CIDR for class detection
	// Name can be the same or a specific IP within the range
	asset := model.NewAsset(c.Block, c.Block)
	return &asset, nil
}

// --- Convenience Functions ---

// IPToAsset converts an IP address string to an Asset.
func IPToAsset(address string, domain ...string) (*model.Asset, error) {
	ip := IP{Address: address}
	if len(domain) > 0 {
		ip.Domain = domain[0]
	}

	result, err := transformIP(ip)
	if err != nil {
		return nil, err
	}

	return result.(*model.Asset), nil
}

// DomainToAsset converts a domain string to an Asset.
func DomainToAsset(name string) (*model.Asset, error) {
	result, err := transformDomain(Domain{Name: name})
	if err != nil {
		return nil, err
	}

	return result.(*model.Asset), nil
}

// CIDRToAsset converts a CIDR block string to an Asset.
// Note: CIDR assets use the CIDR block as both DNS and Name for proper class detection.
func CIDRToAsset(block string) (*model.Asset, error) {
	cidr := CIDR{Block: block}

	result, err := transformCIDR(cidr)
	if err != nil {
		return nil, err
	}

	return result.(*model.Asset), nil
}
