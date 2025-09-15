package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetLabel(t *testing.T) {
	// active directory
	assert.Equal(t, ADObjectLabel, GetLabel("adobject"))
	assert.Equal(t, ADObjectLabel, GetLabel("ADObject"))
	assert.Equal(t, ADObjectLabel, GetLabel("AdObjEcT"))
	assert.Equal(t, ADDomainLabel, GetLabel("addomain"))
	assert.Equal(t, ADDomainLabel, GetLabel("ADDomain"))

	// cloud
	assert.Equal(t, CloudResourceLabel, GetLabel("cloudresource"))
	assert.Equal(t, CloudResourceLabel, GetLabel("CloudResource"))
	assert.Equal(t, AWSResourceLabel, GetLabel("awsresource"))
	assert.Equal(t, AWSResourceLabel, GetLabel("AWSResource"))
	assert.Equal(t, GCPResourceLabel, GetLabel("gcpresource"))
	assert.Equal(t, GCPResourceLabel, GetLabel("GCPResource"))
	assert.Equal(t, AzureResourceLabel, GetLabel("azureresource"))
	assert.Equal(t, AzureResourceLabel, GetLabel("AzureResource"))

	// network
	assert.Equal(t, AssetLabel, GetLabel("asset"))
	assert.Equal(t, AssetLabel, GetLabel("Asset"))
	assert.Equal(t, AssetLabel, GetLabel("ASSET"))
	assert.Equal(t, RepositoryLabel, GetLabel("repository"))
	assert.Equal(t, RepositoryLabel, GetLabel("Repository"))
	assert.Equal(t, IntegrationLabel, GetLabel("integration"))
	assert.Equal(t, IntegrationLabel, GetLabel("Integration"))

	// appsec
	assert.Equal(t, WebApplicationLabel, GetLabel("webapplication"))
	assert.Equal(t, WebApplicationLabel, GetLabel("Webapplication"))
	assert.Equal(t, WebApplicationLabel, GetLabel("WebApplication"))
	assert.Equal(t, WebApplicationLabel, GetLabel("WeBappliCatiOn"))
}
