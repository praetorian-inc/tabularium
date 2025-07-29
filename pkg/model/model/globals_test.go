package model

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGobEncoding(t *testing.T) {
	t.Run("encode/decode asset", func(t *testing.T) {
		asset := NewAsset("test.example.com", "test.example.com")
		asset.Status = Active
		asset.Source = SeedSource

		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(asset)
		require.NoError(t, err)

		var decoded Asset
		err = gob.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&decoded)
		require.NoError(t, err)

		assert.Equal(t, asset.DNS, decoded.DNS)
		assert.Equal(t, asset.Name, decoded.Name)
		assert.Equal(t, asset.Status, decoded.Status)
		assert.Equal(t, asset.Source, decoded.Source)
		assert.Equal(t, asset.Key, decoded.Key)
	})

	t.Run("encode/decode risk", func(t *testing.T) {
		asset := NewAsset("test.example.com", "test.example.com")
		risk := NewRisk(&asset, "test-risk", TriageInfo)
		risk.Comment = "Test comment"
		risk.Agent = "test-agent"

		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(risk)
		require.NoError(t, err)

		var decoded Risk
		err = gob.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&decoded)
		require.NoError(t, err)

		assert.Equal(t, risk.Name, decoded.Name)
		assert.Equal(t, risk.Status, decoded.Status)
		assert.Equal(t, risk.Comment, decoded.Comment)
		assert.Equal(t, risk.Agent, decoded.Agent)
		assert.Equal(t, risk.Key, decoded.Key)
	})

	t.Run("encode/decode attribute", func(t *testing.T) {
		asset := NewAsset("test.example.com", "test.example.com")
		attr := asset.Attribute("test", "value")
		attr.Metadata = map[string]string{"key": "value"}

		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(attr)
		require.NoError(t, err)

		var decoded Attribute
		err = gob.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&decoded)
		require.NoError(t, err)

		assert.Equal(t, attr.Name, decoded.Name)
		assert.Equal(t, attr.Value, decoded.Value)
		assert.Equal(t, attr.Metadata, decoded.Metadata)
		assert.Equal(t, attr.Key, decoded.Key)
	})

	t.Run("encode/decode relationship", func(t *testing.T) {
		asset := NewAsset("test.example.com", "test.example.com")
		risk := NewRisk(&asset, "test-risk", TriageInfo)
		rel := NewHasVulnerability(&asset, &risk)

		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(rel)
		require.NoError(t, err)

		var decoded HasVulnerability
		err = gob.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&decoded)
		require.NoError(t, err)

		assert.Equal(t, rel.Base().Key, decoded.Key)
		assert.Equal(t, rel.Base().Created, decoded.Created)
		assert.Equal(t, rel.Base().Visited, decoded.Visited)
	})

	t.Run("encode/decode slice of interface", func(t *testing.T) {
		asset := NewAsset("test.example.com", "test.example.com")
		risk := NewRisk(&asset, "test-risk", TriageInfo)
		attr := asset.Attribute("test", "value")

		items := []any{&asset, &risk, &attr}

		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(&items)
		require.NoError(t, err)

		var decoded []any
		err = gob.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&decoded)
		require.NoError(t, err)

		require.Len(t, decoded, 3)

		// Check types were preserved
		decodedAsset, ok := decoded[0].(*Asset)
		require.True(t, ok)
		assert.Equal(t, asset.Key, decodedAsset.Key)

		decodedRisk, ok := decoded[1].(*Risk)
		require.True(t, ok)
		assert.Equal(t, risk.Key, decodedRisk.Key)

		decodedAttr, ok := decoded[2].(*Attribute)
		require.True(t, ok)
		assert.Equal(t, attr.Key, decodedAttr.Key)
	})
}

func TestJobEncoding(t *testing.T) {
	t.Run("encode/decode job with target", func(t *testing.T) {
		asset := NewAsset("test.example.com", "test.example.com")
		job := NewJob("test-source", &asset)
		job.Config = map[string]string{"key": "value"}

		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(job)
		require.NoError(t, err)

		var decoded Job
		err = gob.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&decoded)
		require.NoError(t, err)

		assert.Equal(t, job.Source, decoded.Source)
		assert.Equal(t, job.Config, decoded.Config)
		assert.Equal(t, job.Key, decoded.Key)

		// Verify target was properly encoded/decoded
		decodedAsset, ok := decoded.Target.Model.(*Asset)
		require.True(t, ok)
		assert.Equal(t, asset.Key, decodedAsset.Key)
	})
}

func TestGraphRelationshipEncoding(t *testing.T) {
	t.Run("encode/decode discovered relationship", func(t *testing.T) {
		source := NewAsset("source.example.com", "source.example.com")
		target := NewAsset("target.example.com", "target.example.com")
		rel := NewDiscovered(&source, &target)

		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(rel)
		require.NoError(t, err)

		var decoded Discovered
		err = gob.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&decoded)
		require.NoError(t, err)

		assert.Equal(t, rel.Label(), decoded.Label())
		assert.Equal(t, rel.Base().Key, decoded.Key)

		// Verify source and target were properly encoded/decoded
		decodedSource, decodedTarget := decoded.Nodes()
		assert.Equal(t, source.Key, decodedSource.GetKey())
		assert.Equal(t, target.Key, decodedTarget.GetKey())
	})

	t.Run("encode/decode has attribute relationship", func(t *testing.T) {
		asset := NewAsset("test.example.com", "test.example.com")
		attr := asset.Attribute("test", "value")
		rel := NewHasAttribute(&asset, &attr)

		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(rel)
		require.NoError(t, err)

		var decoded HasAttribute
		err = gob.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&decoded)
		require.NoError(t, err)

		assert.Equal(t, rel.Label(), decoded.Label())
		assert.Equal(t, rel.Base().Key, decoded.Key)

		decodedSource, decodedTarget := decoded.Nodes()
		assert.Equal(t, asset.Key, decodedSource.GetKey())
		assert.Equal(t, attr.Key, decodedTarget.GetKey())
	})
}
