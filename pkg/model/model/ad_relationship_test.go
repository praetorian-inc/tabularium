package model

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestADRelationship(t *testing.T) {
	user := NewADUser("example.local", "S-1-5-21-123456789-123456789-123456789-1001", "CN=John Doe,CN=Users,DC=example,DC=local")
	computer := NewADComputer("example.local", "S-1-5-21-123456789-123456789-123456789-1002", "CN=WORKSTATION01,OU=Computers,DC=example,DC=local")
	relationship := NewADRelationship(&user, &computer, ADGenericAllLabel)

	require.Equal(t, ADGenericAllLabel, relationship.Label())
	require.True(t, relationship.Valid())
}

func TestADRelationship_Visit(t *testing.T) {
	ou := NewADOU("example.local", "53F6B870-D2A4-4E83-94EB-41B9C325F26C", "CN=Big Container,DC=example,DC=local")
	domain := NewADComputer("example.local", "S-1-5-21-123456789-123456789-123456789", "DC=example,DC=local")
	existing := NewADRelationship(&ou, &domain, ADGPLinkLabel).(*ADRelationship)
	update := NewADRelationship(&ou, &domain, ADGPLinkLabel).(*ADRelationship)

	falseVal := false
	trueVal := true

	existing.Enforced = &falseVal
	update.Enforced = &trueVal

	existing.Visit(update)

	require.NotNil(t, existing.Enforced)
	assert.True(t, *existing.Enforced)
}
