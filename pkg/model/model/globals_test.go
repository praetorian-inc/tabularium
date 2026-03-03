package model

import (
	"bytes"
	"encoding/gob"
	"fmt"
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

func TestHydratableModels(t *testing.T) {
	t.Run("webpage dehydration and hydration flow", func(t *testing.T) {
		// Webpages always will be hydratable regardless of if req/responses exist
		webpage := NewWebpageFromString("https://example.com", nil)
		assert.True(t, webpage.CanHydrate())

		response := WebpageResponse{
			StatusCode: 200,
			Headers: map[string][]string{
				"Content-Type": {"text/html"},
			},
			Body: "<html><body><h1>Hello World</h1></body></html>",
		}
		request := WebpageRequest{
			RawURL:   "https://example.com/page",
			Method:   "GET",
			Response: &response,
		}
		webpage.AddRequest(request)
		assert.Empty(t, webpage.DetailsFilepath)

		files, dehydrated := webpage.Dehydrate()
		assert.Equal(t, webpage.DetailsFilepath, files[0].Name)
		assert.NotEmpty(t, files[0].Bytes)
		assert.NotEmpty(t, webpage.Requests)
		assert.Equal(t, webpage.Requests[0], request)

		dehydratedWebpage, ok := dehydrated.(*Webpage)
		assert.True(t, ok)
		assert.Empty(t, dehydratedWebpage.Requests)

		newWebpageInstance := NewWebpageFromString("https://example.com", nil)
		newWebpageInstance.DetailsFilepath = dehydratedWebpage.DetailsFilepath

		err := newWebpageInstance.Hydrate(func(path string) ([]byte, error) {
			for _, f := range files {
				if f.Name == path {
					return f.Bytes, nil
				}
			}
			return nil, fmt.Errorf("file not found: %s", path)
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, newWebpageInstance.Requests)
		assert.Equal(t, newWebpageInstance.Requests[0], request)
	})

	t.Run("web application hydration flow", func(t *testing.T) {
		//Web Applications only hydrate if they are a Web Service (contain a API Definition)
		webapp := NewWebApplication("https://example.com", "Example's Website")
		assert.False(t, webapp.CanHydrate())

		fileBased := FileBasedAPIDefinition{
			Filename:         "api.json",
			Contents:         "{}",
			EnabledEndpoints: []EnabledEndpoint{EnabledEndpoint{ID: "endpoint1"}, EnabledEndpoint{ID: "endpoint2"}},
		}
		urlBased := URLBasedAPIDefinition{
			URL: "https://example.com/api/swagger.json",
		}
		apiContent := APIDefinitionResult{
			PrimaryURL:          "https://example.com/api",
			FileBasedDefinition: &fileBased,
			URLBasedDefinition:  &urlBased,
		}
		webapp.ApiDefinitionContent = apiContent
		webapp.BurpType = "webservice"
		assert.True(t, webapp.CanHydrate())
		assert.Empty(t, webapp.ApiDefinitionContentPath)

		files, dehydrated := webapp.Dehydrate()
		assert.Equal(t, webapp.ApiDefinitionContentPath, files[0].Name)
		assert.NotEmpty(t, files[0].Bytes)
		assert.Equal(t, webapp.ApiDefinitionContent, apiContent)

		dehydratedWebapp, ok := dehydrated.(*WebApplication)
		assert.True(t, ok)
		assert.Empty(t, dehydratedWebapp.WebApplicationDetails)

		newInstance := NewWebApplication("https://example.com", "Example's Website")
		newInstance.Hydrate(func(path string) ([]byte, error) {
			for _, f := range files {
				if f.Name == path {
					return f.Bytes, nil
				}
			}
			return nil, fmt.Errorf("file not found: %s", path)
		})
		assert.Equal(t, newInstance.ApiDefinitionContentPath, files[0].Name)
		assert.NotEmpty(t, newInstance.ApiDefinitionContent)
		assert.Equal(t, newInstance.ApiDefinitionContent, apiContent)
	})

	t.Run("aws resource hydration flow", func(t *testing.T) {
		name := "arn:aws:s3:::example-bucket"
		account := "123456789012"

		resource, err := NewAWSResource(name, account, AWSS3Bucket, nil)
		require.NoError(t, err)

		assert.False(t, resource.CanHydrate())

		policyBytes := []byte(`{"Statement":[{"Action":"s3:*"}]}`)
		resource.SetOrgPolicy(policyBytes)

		assert.True(t, resource.CanHydrate())

		expectedPath := resource.OrgPolicyFilename()
		files, dehydrated := resource.Dehydrate()
		assert.Equal(t, expectedPath, files[0].Name)
		assert.Equal(t, policyBytes, []byte(files[0].Bytes))

		dehydratedResource, ok := dehydrated.(*AWSResource)
		assert.True(t, ok)
		assert.Nil(t, dehydratedResource.OrgPolicy)

		newInstance, err := NewAWSResource(name, account, AWSS3Bucket, nil)
		require.NoError(t, err)
		newInstance.HasOrgPolicy = true

		err = newInstance.Hydrate(func(path string) ([]byte, error) {
			for _, f := range files {
				if f.Name == path {
					return f.Bytes, nil
				}
			}
			return nil, fmt.Errorf("file not found: %s", path)
		})
		require.NoError(t, err)

		assert.Equal(t, policyBytes, newInstance.OrgPolicy)
		assert.True(t, newInstance.CanHydrate())
	})

	t.Run("file hydration flow", func(t *testing.T) {
		file := NewFile("example.json")
		file.Bytes = []byte(`{"key": "value"}`)

		assert.True(t, file.CanHydrate())

		files, dehydrated := file.Dehydrate()
		assert.Equal(t, file.Name, files[0].Name)
		assert.Equal(t, file.Bytes, files[0].Bytes)
		assert.Nil(t, dehydrated)

		newInstance := NewFile("example.json")
		newInstance.Hydrate(func(path string) ([]byte, error) {
			return file.Bytes, nil
		})
		assert.Equal(t, newInstance.Name, file.Name)
		assert.Equal(t, newInstance.Bytes, file.Bytes)
	})
}
