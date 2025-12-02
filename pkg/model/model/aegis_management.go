package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
)

func init() {
	shared.Registry.MustRegisterModel(&AegisManagement{})
	shared.Registry.MustRegisterModel(&AegisParameter{})
}

// AegisManagement represents a management capability for Aegis infrastructure operations
// This closely mirrors AgoraCapability but is specifically for management tasks
type AegisManagement struct {
	model.BaseModel
	Name          string           `json:"name" desc:"The name of the management capability" example:"tunnel_management"`
	Title         string           `json:"title" desc:"The pretty name of the management capability" example:"Cloudflare Tunnel Management"`
	Target        string           `json:"target" desc:"The target of the management capability" example:"agent"`
	Description   string           `json:"description" desc:"A description of the management capability suitable for human or LLM use" example:"Manages Cloudflare tunnels for secure agent connectivity"`
	RunsOn        Platform         `json:"runs_on,omitempty" desc:"The platform this management capability runs on" example:"linux"`
	Version       string           `json:"version" desc:"The version of the management capability (major.minor.patch)" example:"1.0.0"`
	Executor      string           `json:"executor" desc:"The task executor that can execute this management capability" example:"aegis"`
	Integration   bool             `json:"integration" desc:"Whether or not this management capability is an integration with an external service" example:"false"`
	Parameters    []AegisParameter `json:"parameters,omitempty" desc:"The parameters/options of the management capability"`
	Async         bool             `json:"async" desc:"Indicates if this is an asynchronous management capability" example:"true"`
	HealthCheck   bool             `json:"healthCheck,omitempty" desc:"If true, triggers a healthcheck after the management task completes" example:"true"`
	LargeArtifact bool             `json:"largeArtifact,omitempty" desc:"If true, this management capability generates large artifacts that can be stored and reviewed later" example:"false"`
}

// GetDescription returns a description for the AegisManagement model
func (a *AegisManagement) GetDescription() string {
	return "Describes a management capability for Aegis infrastructure operations and how to execute it."
}

// AegisParameter represents a parameter for Aegis management capabilities
// This closely mirrors AgoraParameter but is specifically for management tasks
type AegisParameter struct {
	model.BaseModel
	Name        string `json:"name" desc:"The name of the parameter" example:"tunnel_name"`
	Description string `json:"description" desc:"A description of the parameter suitable for human or LLM use" example:"The name of the Cloudflare tunnel to manage"`
	Default     string `json:"default,omitempty" desc:"A string representation of the default value of the parameter" example:"default-tunnel"`
	Required    bool   `json:"required" desc:"Whether the parameter is required" example:"true"`
	Type        string `json:"type" desc:"The type of the parameter" example:"string"`
	Sensitive   bool   `json:"sensitive,omitempty" desc:"Whether the parameter contains sensitive information that should be handled securely" example:"false"`
}

// GetDescription returns a description for the AegisParameter model
func (a *AegisParameter) GetDescription() string {
	return "Describes a single parameter for an Aegis management capability."
}
