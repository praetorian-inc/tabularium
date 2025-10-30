package model

import "github.com/praetorian-inc/tabularium/pkg/model/attacksurface"

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
	// Accepts returns whether the status of this asset allows it to be accepted by this capability
	Accepts() bool
	// Invoke executes the capability
	Invoke() error
	// Timeout returns the maximum amount of time this capability should be allowed to run
	Timeout() int
	// Cleanup cleans up the capability after execution
	HasGlobalConfig() bool
	// Full returns whether the capability should execute a full scan
	Full() bool
	// Rescan indicates whether this capability should be executed during risk rescanning
	Rescan() bool
	// CredentialType returns the type of credential this capability expects
	CredentialType() CredentialType
	// CredentialFormat returns the formats the capability expects (e.g. [CredentialFormatEnv, CredentialFormatFile] for environment variable and file formats together)
	CredentialFormat() []CredentialFormat
	// CredentialParams return any additional credential params needed
	CredentialParams() AdditionalCredParams
	// ValidateCredentials checks the credentials provided to the capability on construction, and returns an error if not valid
	ValidateCredentials() error
	// Static returns whether this capability should always be executed from static IPs if available
	Static() bool
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
