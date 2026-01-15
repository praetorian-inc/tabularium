package model_test

import (
	"encoding/json"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/collection"
	"github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTargetRelationship_Nodes(t *testing.T) {
	asset := model.NewAsset("example.com", "192.168.1.1")
	port := model.NewPort("tcp", 11434, &asset)
	tech, err := model.NewTechnology("cpe:2.3:a:ollama:ollama:*:*:*:*:*:*:*:*")
	require.NoError(t, err)

	rel := model.NewHasTechnology(&port, &tech)
	tr := model.NewTargetRelationship(rel)
	require.NotNil(t, tr)

	nodes := collection.NewCollectionFromRelationship(rel)

	ports := collection.Get[*model.Port](nodes)
	technologies := collection.Get[*model.Technology](nodes)

	assert.Len(t, ports, 1)
	assert.Len(t, technologies, 1)
	assert.Equal(t, port.Key, ports[0].Key)
	assert.Equal(t, tech.Key, technologies[0].Key)
}

func TestTargetRelationship_Label(t *testing.T) {
	asset := model.NewAsset("example.com", "192.168.1.1")
	port := model.NewPort("tcp", 11434, &asset)
	tech, err := model.NewTechnology("cpe:2.3:a:ollama:ollama:*:*:*:*:*:*:*:*")
	require.NoError(t, err)

	rel := model.NewHasTechnology(&port, &tech)
	tr := model.NewTargetRelationship(rel)

	assert.Equal(t, model.HasTechnologyLabel, tr.Label())
}

func TestTargetRelationship_ImplementsTarget(t *testing.T) {
	asset := model.NewAsset("example.com", "192.168.1.1")
	port := model.NewPort("tcp", 11434, &asset)
	port.Status = model.Active
	tech, err := model.NewTechnology("cpe:2.3:a:ollama:ollama:*:*:*:*:*:*:*:*")
	require.NoError(t, err)

	rel := model.NewHasTechnology(&port, &tech)
	tr := model.NewTargetRelationship(rel)

	// Verify it implements Target
	var target model.Target = tr
	assert.NotNil(t, target)

	// Test Target methods
	assert.Equal(t, rel.GetKey(), tr.GetKey())
	assert.Equal(t, model.Active, tr.GetStatus())
	assert.True(t, tr.IsStatus("A"))
	assert.True(t, tr.IsClass(model.HasTechnologyLabel))
	assert.True(t, tr.IsClass("has_technology"))
	assert.False(t, tr.IsClass("DISCOVERED"))
	assert.True(t, tr.Valid())
}

func TestTargetRelationship_Group(t *testing.T) {
	asset := model.NewAsset("example.com", "192.168.1.1")
	port := model.NewPort("tcp", 11434, &asset)
	tech, err := model.NewTechnology("cpe:2.3:a:ollama:ollama:*:*:*:*:*:*:*:*")
	require.NoError(t, err)

	rel := model.NewHasTechnology(&port, &tech)
	tr := model.NewTargetRelationship(rel)

	group := tr.Group()

	// Should contain source group
	assert.Contains(t, group, port.Group())
}

func TestTargetRelationship_Identifier(t *testing.T) {
	asset := model.NewAsset("example.com", "192.168.1.1")
	port := model.NewPort("tcp", 11434, &asset)
	tech, err := model.NewTechnology("cpe:2.3:a:ollama:ollama:*:*:*:*:*:*:*:*")
	require.NoError(t, err)

	rel := model.NewHasTechnology(&port, &tech)
	tr := model.NewTargetRelationship(rel)

	// Identifier should be relationship key
	assert.Equal(t, rel.GetKey(), tr.Identifier())
}

func TestTargetRelationship_IsPrivate(t *testing.T) {
	t.Run("public source and target", func(t *testing.T) {
		asset := model.NewAsset("example.com", "8.8.8.8")
		port := model.NewPort("tcp", 11434, &asset)
		tech, err := model.NewTechnology("cpe:2.3:a:ollama:ollama:*:*:*:*:*:*:*:*")
		require.NoError(t, err)

		rel := model.NewHasTechnology(&port, &tech)
		tr := model.NewTargetRelationship(rel)

		assert.False(t, tr.IsPrivate())
	})

	t.Run("private source", func(t *testing.T) {
		asset := model.NewAsset("internal.local", "192.168.1.1")
		port := model.NewPort("tcp", 11434, &asset)
		tech, err := model.NewTechnology("cpe:2.3:a:ollama:ollama:*:*:*:*:*:*:*:*")
		require.NoError(t, err)

		rel := model.NewHasTechnology(&port, &tech)
		tr := model.NewTargetRelationship(rel)

		assert.True(t, tr.IsPrivate())
	})
}

func TestTargetRelationship_JSON_RoundTrip(t *testing.T) {
	asset := model.NewAsset("example.com", "192.168.1.1")
	port := model.NewPort("tcp", 11434, &asset)
	tech, err := model.NewTechnology("cpe:2.3:a:ollama:ollama:*:*:*:*:*:*:*:*")
	require.NoError(t, err)

	rel := model.NewHasTechnology(&port, &tech)
	tr := model.NewTargetRelationship(rel)

	// Marshal
	data, err := json.Marshal(tr)
	require.NoError(t, err)

	t.Logf("Serialized: %s", string(data))

	// Unmarshal
	var tr2 model.TargetRelationship
	err = json.Unmarshal(data, &tr2)
	require.NoError(t, err)

	// Verify round-trip
	assert.Equal(t, tr.GetKey(), tr2.GetKey())
	assert.Equal(t, tr.Label(), tr2.Label())

	nodes := collection.NewCollectionFromRelationship(tr2.Relationship.Model)

	// Verify typed accessors work after deserialization
	ports := collection.Get[*model.Port](nodes)
	technologies := collection.Get[*model.Technology](nodes)

	require.Len(t, ports, 1)
	require.Len(t, technologies, 1)
	assert.Equal(t, port.Key, ports[0].Key)
	assert.Equal(t, tech.Key, technologies[0].Key)
}

func TestTargetRelationship_WithStatus(t *testing.T) {
	asset := model.NewAsset("example.com", "192.168.1.1")
	port := model.NewPort("tcp", 11434, &asset)
	tech, err := model.NewTechnology("cpe:2.3:a:ollama:ollama:*:*:*:*:*:*:*:*")
	require.NoError(t, err)

	rel := model.NewHasTechnology(&port, &tech)
	tr := model.NewTargetRelationship(rel)

	// WithStatus returns a copy with the new status
	tr2 := tr.WithStatus("F")
	assert.NotEqual(t, tr, tr2)
	assert.Equal(t, "A", tr.GetStatus())  // Original unchanged
	assert.Equal(t, "F", tr2.GetStatus()) // New has updated status
}
