package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type CommandResult struct {
	registry.BaseModel
	ArtifactName string `json:"artifact_name" desc:"The name of the artifact that generated this command result" example:"portscan"`
	Command      string `json:"command" desc:"The command that was executed" example:"nmap -sS 192.168.1.1"`
	ExitCode     int    `json:"exit_code" desc:"The exit code returned by the command" example:"0"`
	Stdout       string `json:"stdout,omitempty" desc:"Standard output from the command execution"`
	Stderr       string `json:"stderr,omitempty" desc:"Standard error from the command execution"`
	Duration     int64  `json:"duration,omitempty" desc:"Duration of command execution in milliseconds"`
	Success      bool   `json:"success" desc:"Whether the command executed successfully (exit code 0)"`
	ErrorMessage string `json:"error_message,omitempty" desc:"Error message if command failed"`
	ParentTarget Target `json:"parent_target,omitempty" desc:"Parent target for the command execution"`
	Target       Target `json:"target,omitempty" desc:"Primary target for the command execution"`
}

func init() {
	registry.Registry.MustRegisterModel(&CommandResult{})
}

// NewCommandResult creates a new CommandResult with the basic required fields
func NewCommandResult(artifactName, command string, exitCode int) *CommandResult {
	return &CommandResult{
		ArtifactName: artifactName,
		Command:      command,
		ExitCode:     exitCode,
		Success:      exitCode == 0,
	}
}

// SetOutput sets the stdout output for the command result
func (cr *CommandResult) SetOutput(output string) {
	cr.Stdout = output
}

// SetError sets the stderr output for the command result
func (cr *CommandResult) SetError(errorOutput string) {
	cr.Stderr = errorOutput
}

// SetErrorOutput sets the stderr output for the command result (alias for SetError)
func (cr *CommandResult) SetErrorOutput(errorOutput string) {
	cr.Stderr = errorOutput
	if errorOutput != "" && cr.ExitCode != 0 {
		cr.ErrorMessage = errorOutput
	}
}

// SetDuration sets the execution duration for the command result
func (cr *CommandResult) SetDuration(duration int64) {
	cr.Duration = duration
}

// SetTargets sets the parent and primary targets for the command result
func (cr *CommandResult) SetTargets(parent, target Target) {
	if parent != nil {
		cr.ParentTarget = parent
	}
	if target != nil {
		cr.Target = target
	}
}

// GetDescription returns a description for the CommandResult model.
func (cr *CommandResult) GetDescription() string {
	return "Represents the result of executing a command, including exit code and output."
}