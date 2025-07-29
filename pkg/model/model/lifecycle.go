package model

type Lifecycle interface {
	// Setup sets up the capability for execution
	Setup() error
	// Cleanup cleans up the capability after execution
	Cleanup() error
}
