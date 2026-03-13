package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebAuthCredentialTypes(t *testing.T) {
	assert.Equal(t, CredentialType("web-idp"), WebIdPCredential)
	assert.Equal(t, CredentialType("web-login"), WebLoginCredential)
	assert.Equal(t, CredentialType("web-static"), WebStaticCredential)
}

func TestValidationConfig_Defaults(t *testing.T) {
	vc := ValidationConfig{
		URL:          "/api/me",
		ExpectStatus: intPtr(200),
	}
	assert.Equal(t, "/api/me", vc.URL)
	assert.Equal(t, 200, *vc.ExpectStatus)
	assert.Nil(t, vc.ExpectBodyMatch)
	assert.Nil(t, vc.RejectRedirectTo)
}

func TestWebAuthStatus_Constants(t *testing.T) {
	assert.Equal(t, WebAuthStatus("active"), WebAuthStatusActive)
	assert.Equal(t, WebAuthStatus("failed"), WebAuthStatusFailed)
	assert.Equal(t, WebAuthStatus("validating"), WebAuthStatusValidating)
}

func intPtr(i int) *int {
	return &i
}
