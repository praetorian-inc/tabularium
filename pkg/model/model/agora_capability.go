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
	CategoryRecon Category = iota + 1
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

// String returns the string representation of a category
func (c Category) String() string {
	switch c {
	case CategoryRecon:
		return "recon"
	case CategoryAD:
		return "ad"
	case CategoryNetwork:
		return "network"
	case CategoryWindows:
		return "windows"
	case CategoryLinux:
		return "linux"
	case CategoryDNS:
		return "dns"
	case CategoryWeb:
		return "web"
	case CategoryDatabase:
		return "database"
	case CategorySMB:
		return "smb"
	case CategoryCICD:
		return "cicd"
	default:
		return ""
	}
}

// ParseCategory parses a string into a Category
func ParseCategory(name string) Category {
	switch strings.ToLower(name) {
	case "recon":
		return CategoryRecon
	case "ad":
		return CategoryAD
	case "network":
		return CategoryNetwork
	case "windows":
		return CategoryWindows
	case "linux":
		return CategoryLinux
	case "dns":
		return CategoryDNS
	case "web":
		return CategoryWeb
	case "database":
		return CategoryDatabase
	case "smb":
		return CategorySMB
	case "cicd":
		return CategoryCICD
	default:
		return 0
	}
}

// Platform represents the platform enum
type Platform int

const (
	PlatformAny Platform = iota
	PlatformWindows
	PlatformLinux
	PlatformMacOS
)

func (p Platform) String() string {
	switch p {
	case PlatformWindows:
		return "windows"
	case PlatformLinux:
		return "linux"
	case PlatformMacOS:
		return "macos"
	default:
		return ""
	}
}

func (p Platform) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *Platform) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	switch strings.ToLower(str) {
	case "windows":
		*p = PlatformWindows
	case "linux":
		*p = PlatformLinux
	case "macos":
		*p = PlatformMacOS
	default:
		*p = PlatformAny
	}

	return nil
}

type AgoraCapability struct {
	registry.BaseModel
	Name          string                `json:"name" desc:"The name of the capability" example:"portscan"`
	Title         string                `json:"title" desc:"The pretty name of the capability" example:"AWS"`
	Target        string                `json:"target" desc:"The target of the capability" example:"asset"`
	Description   string                `json:"description" desc:"A description of the capability suitable for human or LLM use" example:"Identifies open ports on a target host"`
	Category      string                `json:"category" desc:"The categories of the capability" example:"recon,ad"`
	RunsOn        Platform              `json:"runs_on" desc:"The platform the capability runs on" example:"windows"`
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
