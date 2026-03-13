package model

// WebAuthStatus represents the status of a web auth credential validation.
type WebAuthStatus string

const (
	WebAuthStatusActive     WebAuthStatus = "active"
	WebAuthStatusFailed     WebAuthStatus = "failed"
	WebAuthStatusValidating WebAuthStatus = "validating"
)

// ValidationConfig defines how to verify that authentication was successful.
// Stored alongside every web auth credential in SSM.
type ValidationConfig struct {
	URL              string  `json:"url"`                          // Endpoint to hit to verify auth (e.g. "/api/me")
	ExpectStatus     *int    `json:"expect_status,omitempty"`      // Expected HTTP status code
	ExpectBodyMatch  *string `json:"expect_body_match,omitempty"`  // Substring/regex to match in response body
	RejectRedirectTo *string `json:"reject_redirect_to,omitempty"` // If redirected here, auth failed (e.g. "/login")
}
