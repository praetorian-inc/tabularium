package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMAWSAttachedPermission_Label(t *testing.T) {
	role, err := NewAWSResource(
		"arn:aws:iam::123456789012:role/TestRole",
		"123456789012",
		AWSRole,
		nil,
	)
	require.NoError(t, err)

	policy, err := NewAWSResource(
		"arn:aws:iam::aws:policy/AdministratorAccess",
		"123456789012",
		AWSManagedPolicy,
		nil,
	)
	require.NoError(t, err)

	rel := NewIAMAWSAttachedPermission(&role, &policy)

	assert.Equal(t, IAMAWSAttachedPermissionLabel, rel.Label())
	assert.True(t, rel.Valid())
}

func TestIAMAWSAttachedPermission_Key(t *testing.T) {
	role, err := NewAWSResource(
		"arn:aws:iam::123456789012:role/TestRole",
		"123456789012",
		AWSRole,
		nil,
	)
	require.NoError(t, err)

	policy, err := NewAWSResource(
		"arn:aws:iam::aws:policy/AdministratorAccess",
		"123456789012",
		AWSManagedPolicy,
		nil,
	)
	require.NoError(t, err)

	rel := NewIAMAWSAttachedPermission(&role, &policy)

	expectedKey := role.GetKey() + "#" + IAMAWSAttachedPermissionLabel + policy.GetKey()
	assert.Equal(t, expectedKey, rel.GetKey())
}

func TestIAMAWSAttachedPermission_Visit(t *testing.T) {
	role, err := NewAWSResource(
		"arn:aws:iam::123456789012:role/TestRole",
		"123456789012",
		AWSRole,
		nil,
	)
	require.NoError(t, err)

	policy, err := NewAWSResource(
		"arn:aws:iam::aws:policy/AdministratorAccess",
		"123456789012",
		AWSManagedPolicy,
		nil,
	)
	require.NoError(t, err)

	existing := NewIAMAWSAttachedPermission(&role, &policy)
	update := NewIAMAWSAttachedPermission(&role, &policy)

	existing.Visit(update)
	assert.Equal(t, update.Base().Visited, existing.Base().Visited)
}

func TestIAMAWSAttachedPermission_GetDescription(t *testing.T) {
	role, err := NewAWSResource(
		"arn:aws:iam::123456789012:role/TestRole",
		"123456789012",
		AWSRole,
		nil,
	)
	require.NoError(t, err)

	policy, err := NewAWSResource(
		"arn:aws:iam::aws:policy/AdministratorAccess",
		"123456789012",
		AWSManagedPolicy,
		nil,
	)
	require.NoError(t, err)

	rel := NewIAMAWSAttachedPermission(&role, &policy)
	assert.NotEmpty(t, rel.GetDescription())
}

func TestIAMAWSAttachedPermission_Nodes(t *testing.T) {
	role, err := NewAWSResource(
		"arn:aws:iam::123456789012:role/TestRole",
		"123456789012",
		AWSRole,
		nil,
	)
	require.NoError(t, err)

	policy, err := NewAWSResource(
		"arn:aws:iam::aws:policy/AdministratorAccess",
		"123456789012",
		AWSManagedPolicy,
		nil,
	)
	require.NoError(t, err)

	rel := NewIAMAWSAttachedPermission(&role, &policy)

	source, target := rel.Nodes()
	assert.Equal(t, role.GetKey(), source.GetKey())
	assert.Equal(t, policy.GetKey(), target.GetKey())
}
