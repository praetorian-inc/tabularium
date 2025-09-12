package model

import (
	"github.com/praetorian-inc/tabularium/pkg/model/attacksurface"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type AegisManagement struct {
	registry.BaseModel
	Name          string                `json:"name" desc:"The name of the Aegis management capability" example:"system-info"`
	Title         string                `json:"title" desc:"The pretty name of the management capability" example:"System Information"`
	Target        string                `json:"target" desc:"The target of the management capability" example:"agent"`
	Description   string                `json:"description" desc:"A description of the management capability suitable for human or LLM use" example:"Gathers system information from Aegis agents"`
	Category      CategorySlice         `json:"category,omitempty" desc:"The categories this management capability belongs to. Use Category enum constants like CategoryRecon, CategoryAD, CategoryNetwork, etc. Access string values via CategoryStrings[category]" example:"[\"recon\", \"windows\"]"`
	RunsOn        Platform              `json:"runs_on,omitempty" desc:"The platform this management capability runs on. Use Platform enum constants like PlatformWindows, PlatformLinux, PlatformAny, etc. Access string values via PlatformStrings[platform]" example:"windows"`
	Version       string                `json:"version" desc:"The version of the management capability (major.minor.patch)" example:"1.0.0"`
	Executor      string                `json:"executor" desc:"The task executor that can execute this management capability" example:"AegisAgent"`
	Surface       attacksurface.Surface `json:"surface" desc:"The attack surface of the management capability" example:"internal"`
	Integration   bool                  `json:"integration" desc:"Whether or not this capability is an integration with an external service" example:"false"`
	Parameters    []AegisParameter      `json:"parameters,omitempty" desc:"The parameters/options of the management capability" example:"{\"input\": {\"system\": \"windows\"}, \"detailed\": true}"`
	Async         bool                  `json:"async" desc:"Indicates if this is an asynchronous management capability" example:"false"`
	LargeArtifact bool                  `json:"largeArtifact,omitempty" desc:"If true, this management capability generates large artifacts that can be stored and reviewed later" example:"false"`
}

func init() {
	registry.Registry.MustRegisterModel(&AegisManagement{})
}

// GetDescription returns a description for the AegisManagement model.
func (a *AegisManagement) GetDescription() string {
	return "Describes a single Aegis management capability registered with the system and how to execute it."
}