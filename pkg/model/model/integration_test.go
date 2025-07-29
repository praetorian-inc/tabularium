package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegration_NewIntegration(t *testing.T) {
	ia := NewIntegration("github", "https://github.com/praetorian-inc")
	assert.Equal(t, "github", ia.Name)
	assert.Equal(t, "https://github.com/praetorian-inc", ia.Value)
	assert.Equal(t, "#integration#github#https://github.com/praetorian-inc", ia.Key)
	assert.True(t, ia.IsClass("github"))
	assert.Equal(t, "github", ia.Group())
	assert.Equal(t, "https://github.com/praetorian-inc", ia.Identifier())

	// Check BaseAsset defaults for just first repo asset
	assert.Equal(t, Active, ia.Status)
	assert.Equal(t, AccountSource, ia.Source)

	ia = NewIntegration("insightvm", "https://insightvm.praetorian.com")
	assert.Equal(t, "insightvm", ia.Name)
	assert.Equal(t, "https://insightvm.praetorian.com", ia.Value)
	assert.Equal(t, "#integration#insightvm#https://insightvm.praetorian.com", ia.Key)
	assert.True(t, ia.IsClass("insightvm"))
	assert.Equal(t, "insightvm", ia.Group())
	assert.Equal(t, "https://insightvm.praetorian.com", ia.Identifier())

	ia = NewIntegration("burp-enterprise", "https://burp-enterprise.praetorian.com")
	assert.Equal(t, "burp-enterprise", ia.Name)
	assert.Equal(t, "https://burp-enterprise.praetorian.com", ia.Value)
	assert.Equal(t, "#integration#burp-enterprise#https://burp-enterprise.praetorian.com", ia.Key)
	assert.True(t, ia.IsClass("burp-enterprise"))
	assert.Equal(t, "burp-enterprise", ia.Group())
	assert.Equal(t, "https://burp-enterprise.praetorian.com", ia.Identifier())
}

func TestIntegration_Valid(t *testing.T) {
	ia := NewIntegration("github", "https://github.com/praetorian-inc")
	assert.True(t, ia.Valid())

	ia = NewIntegration("github", "")
	assert.False(t, ia.Valid())

	ia = NewIntegration("", "https://github.com/praetorian-inc")
	assert.False(t, ia.Valid())

	ia = NewIntegration("", "")
	assert.False(t, ia.Valid())

	ia = NewIntegration("github", "https://github.com/praetorian-inc")
	ia.Key = "#integration#github#"
	assert.False(t, ia.Valid())

	ia = NewIntegration("github", "https://github.com/praetorian-inc")
	ia.Key = "#integration##https://github.com/praetorian-inc#"
	assert.False(t, ia.Valid())

	ia = NewIntegration("github", "https://github.com/praetorian-inc")
	ia.Key = "#integration##"
	assert.False(t, ia.Valid())
}

func TestNewCloudOrIntegration(t *testing.T) {
	ia := NewCloudOrIntegration("github", "https://github.com/praetorian-inc")
	assert.Equal(t, "github", ia.Group())
	assert.Equal(t, "https://github.com/praetorian-inc", ia.Identifier())
	assert.Equal(t, "#integration#github#https://github.com/praetorian-inc", ia.GetKey())
	assert.True(t, ia.IsClass("github"))
	assert.Equal(t, "github", ia.Group())
	assert.Equal(t, "https://github.com/praetorian-inc", ia.Identifier())

	ia = NewCloudOrIntegration("amazon", "123456789012")
	assert.Equal(t, "amazon", ia.Group())
	assert.Equal(t, "123456789012", ia.Identifier())
	assert.Equal(t, "#asset#amazon#123456789012", ia.GetKey())
	assert.True(t, ia.IsClass("amazon"))
	assert.Equal(t, "amazon", ia.Group())
	assert.Equal(t, "123456789012", ia.Identifier())

	ia = NewCloudOrIntegration("gcp", "123456789012")
	assert.Equal(t, "gcp", ia.Group())
	assert.Equal(t, "123456789012", ia.Identifier())
	assert.Equal(t, "#asset#gcp#123456789012", ia.GetKey())
	assert.True(t, ia.IsClass("gcp"))
	assert.Equal(t, "gcp", ia.Group())
	assert.Equal(t, "123456789012", ia.Identifier())

	ia = NewCloudOrIntegration("azure", "123456789012")
	assert.Equal(t, "azure", ia.Group())
	assert.Equal(t, "123456789012", ia.Identifier())
	assert.Equal(t, "#asset#azure#123456789012", ia.GetKey())
	assert.True(t, ia.IsClass("azure"))
	assert.Equal(t, "azure", ia.Group())
	assert.Equal(t, "123456789012", ia.Identifier())
}
