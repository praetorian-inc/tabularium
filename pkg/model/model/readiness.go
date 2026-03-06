package model

const (
	// CredentialCaptureOnly means the integration was not implemented yet. It only
	// implements the logic to store customer-provided credentials in SSM.
	CredentialCaptureOnly = "CREDENTIAL_CAPTURE_ONLY"

	// NotTestedWithProductionCredentials means the integration was written with
	// trial accounts or other non-production workloads and hasn't been tested with
	// production customer credentials yet.
	NotTestedWithProductionCredentials = "NOT_TESTED_WITH_PRODUCTION_CREDENTIALS"

	// GeneralAvailability means the integration has been tested with at least one real customer
	// and the implementation is considered complete.
	GeneralAvailability = "GENERAL_AVAILABILITY"
)
