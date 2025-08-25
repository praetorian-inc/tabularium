package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestADRelationship(t *testing.T) {
	t.Run("basic relationship creation", func(t *testing.T) {
		user := NewADUser("example.local", "S-1-5-21-123456789-123456789-123456789-1001", "CN=John Doe,CN=Users,DC=example,DC=local")
		computer := NewADComputer("example.local", "S-1-5-21-123456789-123456789-123456789-1002", "CN=WORKSTATION01,OU=Computers,DC=example,DC=local")
		relationship := NewADRelationship(&user, &computer, ADGenericAllLabel)

		require.Equal(t, ADGenericAllLabel, relationship.Label())
		require.True(t, relationship.Valid())
	})

	t.Run("relationship with nil source", func(t *testing.T) {
		computer := NewADComputer("example.local", "S-1-5-21-123456789-123456789-123456789-1002", "CN=WORKSTATION01,OU=Computers,DC=example,DC=local")
		relationship := NewADRelationship(nil, &computer, ADGenericAllLabel)
		require.False(t, relationship.Valid())
	})

	t.Run("relationship with empty label", func(t *testing.T) {
		user := NewADUser("example.local", "S-1-5-21-123456789-123456789-123456789-1001", "CN=John Doe,CN=Users,DC=example,DC=local")
		computer := NewADComputer("example.local", "S-1-5-21-123456789-123456789-123456789-1002", "CN=WORKSTATION01,OU=Computers,DC=example,DC=local")
		relationship := NewADRelationship(&user, &computer, "")
		require.Equal(t, "", relationship.Label())
	})
}
