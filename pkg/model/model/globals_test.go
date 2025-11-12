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

// Hydratable follows these 2 flows for Hydrating and Dehydrating
// To Dehydrate
// We first check to see if the HydratableFilepath() returns a non-empty string
// If so, we get the HydratedFile() and insert it into S3
// Then we call Dehydrate() on the model to remove what we just saved to S3
//
// To Hydrate
// We first check to see if the HydratableFilepath() returns a non-empty string
// If so, we get the file from s3 using that filepath
// Then we call Hydrate() on the model to load the data from the file
func TestHydratableModels(t *testing.T) {
	t.Run("webpage dehydration and hydration flow", func(t *testing.T) {
		// Webpages always will be hydratable regardless of if req/responses exist
		webpage := NewWebpageFromString("https://example.com", nil)
		assert.NotEqual(t, SKIP_HYDRATION, webpage.HydratableFilepath())

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
		assert.Equal(t, webpage.DetailsFilePath(), webpage.HydratableFilepath())

		file := webpage.HydratedFile()
		assert.Equal(t, webpage.DetailsFilepath, file.Name)
		assert.NotEmpty(t, file.Bytes)
		assert.NotEmpty(t, webpage.Requests)
		assert.Equal(t, webpage.Requests[0], request)

		dehydrated := webpage.Dehydrate()
		dehydratedWebpage, ok := dehydrated.(*Webpage)
		assert.True(t, ok)
		assert.Empty(t, dehydratedWebpage.Requests)

		newWebpageInstance := NewWebpageFromString("https://example.com", nil)
		newWebpageInstance.DetailsFilepath = dehydratedWebpage.DetailsFilepath

		err := newWebpageInstance.Hydrate(file.Bytes)
		assert.NoError(t, err)
		assert.NotEmpty(t, newWebpageInstance.Requests)
		assert.Equal(t, newWebpageInstance.Requests[0], request)
	})

	t.Run("web application hydration flow", func(t *testing.T) {
		//Web Applications only hydrate if they are a Web Service (contain a API Definition)
		webapp := NewWebApplication("https://example.com", "Example's Website")
		assert.Equal(t, SKIP_HYDRATION, webapp.HydratableFilepath())

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
		assert.Equal(t, webapp.GetHydratableFilepath(), webapp.HydratableFilepath())
		assert.Empty(t, webapp.ApiDefinitionContentPath)

		file := webapp.HydratedFile()
		assert.Equal(t, webapp.ApiDefinitionContentPath, file.Name)
		assert.NotEmpty(t, file.Bytes)
		assert.Equal(t, webapp.ApiDefinitionContent, apiContent)

		dehydrated := webapp.Dehydrate()
		dehydratedWebapp, ok := dehydrated.(*WebApplication)
		assert.True(t, ok)
		assert.Empty(t, dehydratedWebapp.WebApplicationDetails)

		newInstance := NewWebApplication("https://example.com", "Example's Website")
		newInstance.Hydrate(file.Bytes)
		assert.Equal(t, newInstance.ApiDefinitionContentPath, file.Name)
		assert.NotEmpty(t, newInstance.ApiDefinitionContent)
		assert.Equal(t, newInstance.ApiDefinitionContent, apiContent)
	})

	t.Run("aws resource hydration flow", func(t *testing.T) {
		name := "arn:aws:s3:::example-bucket"
		account := "123456789012"

		resource, err := NewAWSResource(name, account, AWSS3Bucket, nil)
		require.NoError(t, err)

		assert.Equal(t, SKIP_HYDRATION, resource.HydratableFilepath())

		policyBytes := []byte(`{"Statement":[{"Action":"s3:*"}]}`)
		err = resource.Hydrate(policyBytes)
		require.NoError(t, err)

		expectedPath := resource.GetOrgPolicyFilename()
		assert.Equal(t, expectedPath, resource.HydratableFilepath())

		file := resource.HydratedFile()
		assert.Equal(t, expectedPath, file.Name)
		assert.Equal(t, policyBytes, []byte(file.Bytes))

		dehydrated := resource.Dehydrate()
		dehydratedResource, ok := dehydrated.(*AWSResource)
		assert.True(t, ok)
		assert.Nil(t, dehydratedResource.OrgPolicy)

		newInstance, err := NewAWSResource(name, account, AWSS3Bucket, nil)
		require.NoError(t, err)

		err = newInstance.Hydrate(file.Bytes)
		require.NoError(t, err)

		assert.Equal(t, policyBytes, newInstance.OrgPolicy)
		assert.Equal(t, newInstance.GetOrgPolicyFilename(), newInstance.HydratableFilepath())
	})

	t.Run("file hydration flow", func(t *testing.T) {
		file := NewFile("example.json")
		file.Bytes = []byte(`{"key": "value"}`)

		assert.Equal(t, file.Name, file.HydratableFilepath())

		hFile := file.HydratedFile()
		assert.Equal(t, file.Name, hFile.Name)
		assert.Equal(t, file.Bytes, hFile.Bytes)

		dehydrated := file.Dehydrate()
		assert.Nil(t, dehydrated)

		newInstance := NewFile("example.json")
		newInstance.Hydrate(file.Bytes)
		assert.Equal(t, newInstance.Name, file.Name)
		assert.Equal(t, newInstance.Bytes, file.Bytes)
	})
}
