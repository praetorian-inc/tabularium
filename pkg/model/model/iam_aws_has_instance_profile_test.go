package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMAWSHasInstanceProfile_Label(t *testing.T) {
	role, err := NewAWSResource(
		"arn:aws:iam::123456789012:role/TestRole",
		"123456789012",
		AWSRole,
		nil,
	)
	require.NoError(t, err)

	ip, err := NewAWSResource(
		"arn:aws:iam::123456789012:instance-profile/EC2-Profile",
		"123456789012",
		AWSInstanceProfile,
		nil,
	)
	require.NoError(t, err)

	rel := NewIAMAWSHasInstanceProfile(&role, &ip)

	assert.Equal(t, IAMAWSHasInstanceProfileLabel, rel.Label())
	assert.True(t, rel.Valid())
}

func TestIAMAWSHasInstanceProfile_Key(t *testing.T) {
	role, err := NewAWSResource(
		"arn:aws:iam::123456789012:role/TestRole",
		"123456789012",
		AWSRole,
		nil,
	)
	require.NoError(t, err)

	ip, err := NewAWSResource(
		"arn:aws:iam::123456789012:instance-profile/EC2-Profile",
		"123456789012",
		AWSInstanceProfile,
		nil,
	)
	require.NoError(t, err)

	rel := NewIAMAWSHasInstanceProfile(&role, &ip)

	expectedKey := role.GetKey() + "#" + IAMAWSHasInstanceProfileLabel + ip.GetKey()
	assert.Equal(t, expectedKey, rel.GetKey())
}

func TestIAMAWSHasInstanceProfile_Visit(t *testing.T) {
	role, err := NewAWSResource(
		"arn:aws:iam::123456789012:role/TestRole",
		"123456789012",
		AWSRole,
		nil,
	)
	require.NoError(t, err)

	ip, err := NewAWSResource(
		"arn:aws:iam::123456789012:instance-profile/EC2-Profile",
		"123456789012",
		AWSInstanceProfile,
		nil,
	)
	require.NoError(t, err)

	existing := NewIAMAWSHasInstanceProfile(&role, &ip)
	update := NewIAMAWSHasInstanceProfile(&role, &ip)

	existing.Visit(update)
	assert.Equal(t, update.Base().Visited, existing.Base().Visited)
}

func TestIAMAWSHasInstanceProfile_GetDescription(t *testing.T) {
	role, err := NewAWSResource(
		"arn:aws:iam::123456789012:role/TestRole",
		"123456789012",
		AWSRole,
		nil,
	)
	require.NoError(t, err)

	ip, err := NewAWSResource(
		"arn:aws:iam::123456789012:instance-profile/EC2-Profile",
		"123456789012",
		AWSInstanceProfile,
		nil,
	)
	require.NoError(t, err)

	rel := NewIAMAWSHasInstanceProfile(&role, &ip)
	assert.NotEmpty(t, rel.GetDescription())
}

func TestIAMAWSHasInstanceProfile_Nodes(t *testing.T) {
	role, err := NewAWSResource(
		"arn:aws:iam::123456789012:role/TestRole",
		"123456789012",
		AWSRole,
		nil,
	)
	require.NoError(t, err)

	ip, err := NewAWSResource(
		"arn:aws:iam::123456789012:instance-profile/EC2-Profile",
		"123456789012",
		AWSInstanceProfile,
		nil,
	)
	require.NoError(t, err)

	rel := NewIAMAWSHasInstanceProfile(&role, &ip)

	source, target := rel.Nodes()
	assert.Equal(t, role.GetKey(), source.GetKey())
	assert.Equal(t, ip.GetKey(), target.GetKey())
}
