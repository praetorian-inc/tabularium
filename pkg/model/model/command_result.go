package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&CommandResult{})
}

// CommandResult represents the success/failure status of an Aegis capability execution
type CommandResult struct {
	registry.BaseModel
	Capability   string `json:"capability" desc:"The Aegis capability that was executed" example:"praetorian.aegis.Windows.ad.pingcastle"`
	Success      bool   `json:"success" desc:"Whether the capability executed successfully" example:"true"`
	ExitCode     int    `json:"exit_code" desc:"The exit code returned by the capability" example:"0"`
	ErrorMessage string `json:"error_message,omitempty" desc:"Error message if the capability failed" example:"Command failed with exit code 1"`
	Target       Target `json:"target" desc:"The target this capability was executed against"`
}

// GetDescription returns a description for the CommandResult model
func (c *CommandResult) GetDescription() string {
	return "Represents the success/failure status of an Aegis capability execution."
}

// NewCommandResult creates a new CommandResult for an Aegis capability
func NewCommandResult(capability string, command string, exitCode int) CommandResult {
	return CommandResult{
		Capability: capability,
		Success:    exitCode == 0,
		ExitCode:   exitCode,
	}
}

// SetOutput sets any output (currently unused for simple success/failure tracking)
func (c *CommandResult) SetOutput(output string) {
	// No-op for simple success/failure model
}

// SetErrorOutput sets the error message if the capability failed
func (c *CommandResult) SetErrorOutput(errorOutput string) {
	if !c.Success && errorOutput != "" {
		c.ErrorMessage = errorOutput
	}
}

// SetDuration sets execution duration (currently unused for simple model)
func (c *CommandResult) SetDuration(duration int64) {
	// No-op for simple success/failure model
}

// SetTargets sets the target for the capability result
func (c *CommandResult) SetTargets(parent, target Target) {
	if target != nil {
		c.Target = target
	}
	// Parent not needed for simple model
}