package model

import (
	"encoding/json"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/model/attacksurface"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// Category represents capability categories
type Category int

const (
	CategoryUnknown Category = iota
	CategoryRecon
	CategoryAD
	CategoryNetwork
	CategoryWindows
	CategoryLinux
	CategoryDNS
	CategoryWeb
	CategoryDatabase
	CategorySMB
	CategoryCICD
)

// CategoryStrings maps category IDs to their string representations
var CategoryStrings = map[Category]string{
	CategoryRecon:    "recon",
	CategoryAD:       "ad",
	CategoryNetwork:  "network",
	CategoryWindows:  "windows",
	CategoryLinux:    "linux",
	CategoryDNS:      "dns",
	CategoryWeb:      "web",
	CategoryDatabase: "database",
	CategorySMB:      "smb",
	CategoryCICD:     "cicd",
}

// CategoryNames maps string names to category IDs (reverse lookup)
var CategoryNames = map[string]Category{
	"recon":    CategoryRecon,
	"ad":       CategoryAD,
	"network":  CategoryNetwork,
	"windows":  CategoryWindows,
	"linux":    CategoryLinux,
	"dns":      CategoryDNS,
	"web":      CategoryWeb,
	"database": CategoryDatabase,
	"smb":      CategorySMB,
	"cicd":     CategoryCICD,
}

// String returns the string representation of a category
func (c Category) String() string {
	if name, exists := CategoryStrings[c]; exists {
		return name
	}
	return ""
}

// ParseCategory parses a string into a Category
func ParseCategory(name string) Category {
	if cat, exists := CategoryNames[strings.ToLower(name)]; exists {
		return cat
	}
	return 0
}

// ParseCategories parses a comma-separated string into a slice of categories
func ParseCategories(categoriesStr string) []Category {
	if categoriesStr == "" {
		return nil
	}

	var categories []Category
	parts := strings.Split(categoriesStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if cat := ParseCategory(part); cat != 0 {
			categories = append(categories, cat)
		}
	}
	return categories
}

// HasCategory checks if a category slice contains a specific category
func HasCategory(categories []Category, target Category) bool {
	for _, cat := range categories {
		if cat == target {
			return true
		}
	}
	return false
}

// CategorySlice represents a slice of categories with custom JSON marshaling
type CategorySlice []Category

// MarshalJSON implements custom JSON marshaling for CategorySlice
func (cs CategorySlice) MarshalJSON() ([]byte, error) {
	var names []string
	for _, cat := range cs {
		names = append(names, cat.String())
	}
	return json.Marshal(strings.Join(names, ","))
}

// UnmarshalJSON implements custom JSON unmarshaling for CategorySlice
func (cs *CategorySlice) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	*cs = CategorySlice(ParseCategories(str))
	return nil
}

// Platform represents the platform enum
type Platform int

const (
	PlatformAny Platform = iota
	PlatformWindows
	PlatformLinux
	PlatformMacOS
)

// PlatformStrings maps platform IDs to their string representations
var PlatformStrings = map[Platform]string{
	PlatformAny:     "any",
	PlatformWindows: "windows",
	PlatformLinux:   "linux",
	PlatformMacOS:   "macos",
}

// PlatformNames maps string names to platform IDs (reverse lookup)
var PlatformNames = map[string]Platform{
	"any":     PlatformAny,
	"":        PlatformAny, // Empty string defaults to "any"
	"windows": PlatformWindows,
	"linux":   PlatformLinux,
	"macos":   PlatformMacOS,
}

func (p Platform) String() string {
	if name, exists := PlatformStrings[p]; exists {
		return name
	}
	return ""
}

// ParsePlatform parses a string into a Platform
func ParsePlatform(name string) Platform {
	if platform, exists := PlatformNames[strings.ToLower(name)]; exists {
		return platform
	}
	return PlatformAny // Default to any if unknown
}

func (p Platform) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *Platform) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	*p = ParsePlatform(str)
	return nil
}

type AgoraCapability struct {
	registry.BaseModel
	Name          string                `json:"name" desc:"The name of the capability" example:"portscan"`
	Title         string                `json:"title" desc:"The pretty name of the capability" example:"AWS"`
	Target        string                `json:"target" desc:"The target of the capability" example:"asset"`
	Description   string                `json:"description" desc:"A description of the capability suitable for human or LLM use" example:"Identifies open ports on a target host"`
	Category      CategorySlice         `json:"category,omitempty" desc:"The categories this capability belongs to. Use Category enum constants like CategoryRecon, CategoryAD, CategoryNetwork, etc. Access string values via CategoryStrings[category]" example:"[\"recon\", \"ad\"]"`
	RunsOn        Platform              `json:"runs_on,omitempty" desc:"The platform this capability runs on. Use Platform enum constants like PlatformWindows, PlatformLinux, PlatformAny, etc. Access string values via PlatformStrings[platform]" example:"windows"`
	Version       string                `json:"version" desc:"The version of the capability (major.minor.patch)" example:"1.0.0"`
	Executor      string                `json:"executor" desc:"The task executor that can execute this capability" example:"JanusPlugin"`
	Surface       attacksurface.Surface `json:"surface" desc:"The attack surface of the capability" example:"internal"`
	Integration   bool                  `json:"integration" desc:"Whether or not this capability is an integration with an external service" example:"true"`
	Parameters    []AgoraParameter      `json:"parameters,omitempty" desc:"The parameters/options of the capability" example:"{\"input\": {\"ip\": \"1.2.3.4\"}, \"tcp\": true, \"udp\": false}"`
	Async         bool                  `json:"async" desc:"Indicates if this is an asynchronous capability" example:"false"`
	LargeArtifact bool                  `json:"largeArtifact,omitempty" desc:"If true, this capability generates large artifacts that can be stored and reviewed later" example:"false"`
}

func init() {
	registry.Registry.MustRegisterModel(&AgoraCapability{})
}

// GetDescription returns a description for the AgoraCapability model.
func (a *AgoraCapability) GetDescription() string {
	return "Describes a single capability registered with Agora and how to execute it."
}

type AgoraCapabilityOpt func(*AgoraCapability)

func WithAsync(async bool) AgoraCapabilityOpt {
	return func(a *AgoraCapability) {
		a.Async = async
	}
}
