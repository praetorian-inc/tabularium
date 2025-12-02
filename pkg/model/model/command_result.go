package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
)

func init() {
	shared.Registry.MustRegisterModel(&CommandResult{})
}

// CommandResult represents the success/failure status of an Aegis capability execution
type CommandResult struct {
	model.BaseModel
	Capability   string `json:"capability" desc:"The Aegis capability that was executed" example:"praetorian.aegis.Windows.ad.pingcastle"`
	Command      string `json:"command,omitempty" desc:"The actual command that was executed" example:"adduser testuser"`
	Success      bool   `json:"success" desc:"Whether the capability executed successfully" example:"true"`
	ExitCode     int    `json:"exit_code" desc:"The exit code returned by the capability" example:"0"`
	Output       string `json:"output,omitempty" desc:"Standard output from the command" example:"User added successfully"`
	ErrorOutput  string `json:"error_output,omitempty" desc:"Error output from the command" example:"fatal: The user 'newuser123' does not exist."`
	ErrorMessage string `json:"error_message,omitempty" desc:"Error message if the capability failed" example:"Command failed with exit code 1"`
	Duration     int64  `json:"duration,omitempty" desc:"Execution duration in milliseconds" example:"1500"`
	Target       Target `json:"target,omitempty" desc:"The target this capability was executed against (optional for management capabilities)"`
}

// GetDescription returns a description for the CommandResult model
func (c *CommandResult) GetDescription() string {
	return "Represents the success/failure status of an Aegis capability execution."
}

// NewCommandResult creates a new CommandResult for an Aegis capability
func NewCommandResult(capability string, command string, exitCode int) CommandResult {
	return CommandResult{
		Capability: capability,
		Command:    command,
		Success:    exitCode == 0,
		ExitCode:   exitCode,
	}
}

// SetOutput sets the standard output from the command
func (c *CommandResult) SetOutput(output string) {
	c.Output = output
}

// SetErrorOutput sets the error output from the command
func (c *CommandResult) SetErrorOutput(errorOutput string) {
	c.ErrorOutput = errorOutput
	// Also set ErrorMessage for backwards compatibility
	if !c.Success && errorOutput != "" {
		c.ErrorMessage = errorOutput
	}
}

// SetDuration sets execution duration in milliseconds
func (c *CommandResult) SetDuration(duration int64) {
	c.Duration = duration
}

// SetTargets sets the target for the capability result
func (c *CommandResult) SetTargets(parent, target Target) {
	if target != nil {
		c.Target = target
	}
	// Parent not needed for simple model
}
