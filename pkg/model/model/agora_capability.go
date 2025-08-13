package model

import (
	"github.com/praetorian-inc/tabularium/pkg/model/attacksurface"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type AgoraCapability struct {
	registry.BaseModel
	Name          string                `json:"name" desc:"The name of the capability" example:"portscan"`
	Title         string                `json:"title" desc:"The pretty name of the capability" example:"AWS"`
	Target        string                `json:"target" desc:"The target of the capability" example:"asset"`
	Description   string                `json:"description" desc:"A description of the capability suitable for human or LLM use" example:"Identifies open ports on a target host"`
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
