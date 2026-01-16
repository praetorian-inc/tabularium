package model

import (
	"time"

	"github.com/praetorian-inc/tabularium/pkg/model/attacksurface"
)

type Capability interface {
	// Name returns the name of the capability
	Name() string
	// Title returns the pretty name of the capability
	Title() string
	// Description returns a description of the capability
	Description() string
	// Surface returns the attack surface of the capability
	Surface() attacksurface.Surface
	// Parameters returns the parameters of the capability
	Parameters() []AgoraParameter
	// GetJob returns the capability's job
	GetJob() Job
	// Match returns whether the attributes of the capability match this capability
	Match() error
	// Accepts returns whether the status of the target and job are acceptable to this capability
	Accepts() bool
	// Invoke executes the capability
	Invoke() error
	// Timeout returns the maximum amount of time this capability should be allowed to run
	Timeout() int
	// HasGlobalConfig indicates if the capability's configuration file is shared across all instances of the capability
	HasGlobalConfig() bool
	// Full returns how often the capability should execute a full scan. This is used to determine if job.Full is set to true.
	Full() time.Duration
	// Rescan indicates whether this capability should be executed during risk rescanning
	Rescan() bool
	// Retries returns the maximum number of retries allowed by this capability, and the delay time that should be used
	Retries() (int, time.Duration)
	// CredentialType returns the type of credential this capability expects
	CredentialType() CredentialType
	// CredentialFormat returns the formats the capability expects (e.g. [CredentialFormatEnv, CredentialFormatFile] for environment variable and file formats together)
	CredentialFormat() []CredentialFormat
	// CredentialParams return any additional credential params needed
	CredentialParams() AdditionalCredParams
	// ValidateCredentials checks the credentials provided to the capability on construction, and returns an error if not valid
	ValidateCredentials() error
	// Static returns whether this capability should always be executed from static IPs if available, or nil if no particular behavior is desired
	Static() *bool
	// BypassFrozen returns whether this capability should run even if the account is frozen, or nil if the default behavior is desired
	BypassFrozen() *bool
	// CheckAffiliation checks if the supplied asset is currently enumerated within the integrated account
	CheckAffiliation(Asset) (bool, error)
	// Integration returns whether this capability is an integration
	Integration() bool
	// LargeArtifact returns whether this capability should support downloading a large artifact output
	LargeArtifact() bool
	Lifecycle
}

// ParseParameters accepts a capability's agora parameters and a job config, and parses the values from the job config into the agora parameters
func ParseCapabilityParameters(params []AgoraParameter, config map[string]string) error {
	for _, param := range params {
		// we must use a string pointer to distinguish between an unset parameter (nil) and a parameter set to the empty string
		var strPtr *string
		if strVal, ok := config[param.Name]; ok {
			strPtr = &strVal
		}

		if err := param.Parse(strPtr); err != nil {
			return err
		}
	}

	return nil
}
