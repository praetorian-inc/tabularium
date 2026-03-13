package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewK8sResource(t *testing.T) {
	props := map[string]any{"asset_name": "my-deploy"}
	res, err := NewK8sResource("ade9fcf6-1dfa-461f-993f-ad160576f728", "173736951668", K8sDeployment, props)
	require.NoError(t, err)

	assert.Equal(t, "ade9fcf6-1dfa-461f-993f-ad160576f728", res.Name)
	assert.Equal(t, "173736951668", res.AccountRef)
	assert.Equal(t, K8sDeployment, res.ResourceType)
	assert.Equal(t, K8sProvider, res.Provider)
	assert.Equal(t, "#k8sresource#173736951668#ade9fcf6-1dfa-461f-993f-ad160576f728", res.Key)
	assert.Contains(t, res.Labels, K8sResourceLabel)
	assert.Equal(t, "my-deploy", res.Properties["asset_name"])
}

func TestK8sResource_Key(t *testing.T) {
	res, err := NewK8sResource("uuid-123", "account-456", K8sSecret, nil)
	require.NoError(t, err)
	assert.Equal(t, "#k8sresource#account-456#uuid-123", res.Key)
}

func TestK8sResource_NewAssets(t *testing.T) {
	res, err := NewK8sResource("uuid-123", "account-456", K8sPod, nil)
	require.NoError(t, err)

	assets := res.NewAssets()
	require.Len(t, assets, 1)
	assert.Equal(t, "uuid-123", assets[0].CloudId)
	assert.Equal(t, "account-456", assets[0].CloudAccount)
}

func TestK8sResource_ClusterAndNamespace(t *testing.T) {
	res, err := NewK8sResource("uuid-123", "account-456", K8sDeployment, nil)
	require.NoError(t, err)

	res.Cluster = "prod-cluster"
	res.Namespace = "default"

	assert.Equal(t, "prod-cluster", res.Cluster)
	assert.Equal(t, "default", res.Namespace)
}

func TestK8sResource_Visit(t *testing.T) {
	res1, _ := NewK8sResource("uuid-123", "account-456", K8sDeployment, nil)
	res2, _ := NewK8sResource("uuid-123", "account-456", K8sDeployment, nil)
	res2.Cluster = "prod-cluster"
	res2.Namespace = "kube-system"

	res1.Visit(&res2)
	assert.Equal(t, "prod-cluster", res1.Cluster)
	assert.Equal(t, "kube-system", res1.Namespace)
}

func TestK8sResource_IsPrivate(t *testing.T) {
	res, _ := NewK8sResource("uuid-123", "account-456", K8sDeployment, nil)
	assert.True(t, res.IsPrivate())
}

func TestK8sResource_GroupAndIdentifier(t *testing.T) {
	res, _ := NewK8sResource("uuid-123", "account-456", K8sJob, nil)
	assert.Equal(t, "account-456", res.Group())
	assert.Equal(t, "uuid-123", res.Identifier())
}
