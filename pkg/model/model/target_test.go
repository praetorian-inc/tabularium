package model

import (
	"encoding/json"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTargetEvent_UnmarshalJSON(t *testing.T) {
	t.Run("unmarshal asset", func(t *testing.T) {
		input := `{
			"key": "#asset#example.com#1.2.3.4",
			"username": "test@example.com",
			"dns": "example.com",
			"name": "1.2.3.4",
			"status": "P"
		}`

		var event TargetWrapper
		err := json.Unmarshal([]byte(input), &event)
		require.NoError(t, err)

		asset, ok := event.Model.(*Asset)
		require.True(t, ok, "expected Target to be *Asset")
		assert.Equal(t, "#asset#example.com#1.2.3.4", asset.Key)
		assert.Equal(t, "test@example.com", asset.Username)
		assert.Equal(t, "example.com", asset.DNS)
		assert.Equal(t, "1.2.3.4", asset.Name)
		assert.Equal(t, "P", asset.Status)
	})

	t.Run("unmarshal attribute", func(t *testing.T) {
		input := `{
			"key": "#attribute#https#443#asset#example.com#1.2.3.4",
			"username": "test@example.com",
			"name": "https",
			"value": "443",
			"source": "#asset#example.com#1.2.3.4",
			"status": "A"
		}`

		var event TargetWrapper
		err := json.Unmarshal([]byte(input), &event)
		require.NoError(t, err)

		attr, ok := event.Model.(*Attribute)
		require.True(t, ok, "expected Target to be *Attribute")
		assert.Equal(t, "#attribute#https#443#asset#example.com#1.2.3.4", attr.Key)
		assert.Equal(t, "test@example.com", attr.Username)
		assert.Equal(t, "https", attr.Name)
		assert.Equal(t, "443", attr.Value)
		assert.Equal(t, "#asset#example.com#1.2.3.4", attr.Source)
		assert.Equal(t, "A", attr.Status)
	})

	t.Run("unmarshal preseed", func(t *testing.T) {
		input := `{
			"key": "#preseed#whois+company#Chariot Systems#Chariot Systems",
			"username": "test@example.com",
			"type": "whois+company",
			"title": "Chariot Systems",
			"value": "Chariot Systems",
			"status": "P"
		}`

		var event TargetWrapper
		err := json.Unmarshal([]byte(input), &event)
		require.NoError(t, err)

		preseed, ok := event.Model.(*Preseed)
		require.True(t, ok, "expected Target to be *Preseed")
		assert.Equal(t, "#preseed#whois+company#Chariot Systems#Chariot Systems", preseed.Key)
		assert.Equal(t, "test@example.com", preseed.Username)
		assert.Equal(t, "whois+company", preseed.Type)
		assert.Equal(t, "Chariot Systems", preseed.Title)
		assert.Equal(t, "Chariot Systems", preseed.Value)
		assert.Equal(t, "P", preseed.Status)
	})

	t.Run("unmarshal webpage", func(t *testing.T) {
		input := `{
			"key": "#webpage#https://example.com/#1.2.3.4",
			"username": "test@example.com",
			"url": "https://example.com/",
			"status": "A",
			"source": ["crawler"],
			"state": "interesting",
			"created": "2023-10-27T10:00:00Z",
			"visited": "2023-10-27T11:00:00Z",
			"ttl": 123456789,
			"metadata": {
				"test": "value"
			}
		}`

		var event TargetWrapper
		err := json.Unmarshal([]byte(input), &event)
		require.NoError(t, err)

		webpage, ok := event.Model.(*Webpage)
		require.True(t, ok, "expected Target to be *Webpage")
		assert.Equal(t, "#webpage#https://example.com/#1.2.3.4", webpage.Key)
		assert.Equal(t, "test@example.com", webpage.Username)
		assert.Equal(t, "https://example.com/", webpage.URL)
		assert.Equal(t, "A", webpage.Status)
		assert.Equal(t, []string{"crawler"}, webpage.Source)
		assert.Equal(t, "2023-10-27T10:00:00Z", webpage.Created)
		assert.Equal(t, "2023-10-27T11:00:00Z", webpage.Visited)
		assert.Equal(t, int64(123456789), webpage.TTL)
		assert.Equal(t, "value", webpage.Metadata["test"])
	})

	t.Run("error on missing key", func(t *testing.T) {
		input := `{
			"username": "test@example.com",
			"name": "example.com"
		}`

		var event TargetWrapper
		err := json.Unmarshal([]byte(input), &event)
		assert.Error(t, err)
	})

	t.Run("error on invalid key type", func(t *testing.T) {
		input := `{
			"key": 123,
			"username": "test@example.com"
		}`

		var event TargetWrapper
		err := json.Unmarshal([]byte(input), &event)
		assert.Error(t, err)
	})

	t.Run("error on unknown target type", func(t *testing.T) {
		input := `{
			"key": "#unknown#example.com#1.2.3.4",
			"username": "test@example.com"
		}`

		var event TargetWrapper
		err := json.Unmarshal([]byte(input), &event)
		assert.Error(t, err)
	})

	t.Run("error on invalid JSON", func(t *testing.T) {
		input := `{invalid json`

		var event TargetWrapper
		err := json.Unmarshal([]byte(input), &event)
		assert.Error(t, err)
	})
}

func TestTargetEvent_Interface(t *testing.T) {
	asset := NewAsset("example.com", "1.2.3.4")
	attribute := asset.Attribute("https", "443")
	preseed := NewPreseed("whois+company", "Chariot Systems", "Chariot Systems")
	webpage := NewWebpageFromString("https://example.com/", &attribute)

	testTargetInterface(t, "Asset", &asset)
	testTargetInterface(t, "Attribute", &attribute)
	testTargetInterface(t, "Preseed", &preseed)
	testTargetInterface(t, "Webpage", &webpage)
}

func testTargetInterface(t *testing.T, name string, target Target) {
	t.Run(name+" implements Target interface", func(t *testing.T) {
		assert.NotEmpty(t, target.GetKey(), "GetKey() should return non-empty string")

		oldStatus := target.GetStatus()
		newTarget := target.WithStatus("newstatus")
		assert.NotEqual(t, oldStatus, newTarget.GetStatus(), "Status should be updated")

		assert.NotEmpty(t, target.Group(), "Group() should return non-empty string")
		assert.NotEmpty(t, target.Identifier(), "Identifier() should return non-empty string")

		assert.NotPanics(t, func() {
			target.IsStatus("somestatus")
		}, "Is() should not panic")
	})
}

func TestTargetEvent_DynamoDBMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal asset", func(t *testing.T) {
		asset := NewAsset("example.com", "1.2.3.4")
		asset.Username = "test@example.com"
		asset.Status = "P"
		asset.Created = Now()
		asset.Visited = Now()
		asset.TTL = 123456789

		original := TargetWrapper{Model: &asset}

		av, err := attributevalue.Marshal(original)
		require.NoError(t, err)

		var unmarshaled TargetWrapper
		err = attributevalue.Unmarshal(av, &unmarshaled)
		require.NoError(t, err)

		result, ok := unmarshaled.Model.(*Asset)
		require.True(t, ok)
		assert.Equal(t, asset.Key, result.Key)
		assert.Equal(t, asset.Username, result.Username)
		assert.Equal(t, asset.DNS, result.DNS)
		assert.Equal(t, asset.Name, result.Name)
		assert.Equal(t, asset.Status, result.Status)
		assert.Equal(t, asset.Created, result.Created)
		assert.Equal(t, asset.Visited, result.Visited)
		assert.Equal(t, asset.TTL, result.TTL)
	})

	t.Run("marshal and unmarshal attribute", func(t *testing.T) {
		attribute := NewAttribute("https", "443", &Asset{BaseAsset: BaseAsset{Key: "#asset#example.com#1.2.3.4"}})
		attribute.Username = "test@example.com"
		attribute.Status = "A"
		attribute.Created = Now()
		attribute.Visited = Now()
		attribute.TTL = 123456789
		attribute.Capability = "portscan"
		attribute.Metadata = map[string]string{"test": "value"}

		original := TargetWrapper{Model: &attribute}

		av, err := attributevalue.Marshal(original)
		require.NoError(t, err)

		var unmarshaled TargetWrapper
		err = attributevalue.Unmarshal(av, &unmarshaled)
		require.NoError(t, err)

		result, ok := unmarshaled.Model.(*Attribute)
		require.True(t, ok)
		assert.Equal(t, attribute.Key, result.Key)
		assert.Equal(t, attribute.Username, result.Username)
		assert.Equal(t, attribute.Name, result.Name)
		assert.Equal(t, attribute.Value, result.Value)
		assert.Equal(t, attribute.Source, result.Source)
		assert.Equal(t, attribute.Status, result.Status)
		assert.Equal(t, attribute.Created, result.Created)
		assert.Equal(t, attribute.Visited, result.Visited)
		assert.Equal(t, attribute.TTL, result.TTL)
		assert.Equal(t, attribute.Capability, result.Capability)
		assert.Equal(t, attribute.Metadata, result.Metadata)
	})

	t.Run("marshal and unmarshal preseed", func(t *testing.T) {
		preseed := NewPreseed("whois+company", "Chariot Systems", "Chariot Systems")
		preseed.Username = "test@example.com"
		preseed.Status = "P"
		preseed.Created = Now()
		preseed.Visited = Now()
		preseed.TTL = 123456789
		preseed.Capability = "whois"

		original := TargetWrapper{Model: &preseed}

		av, err := attributevalue.Marshal(original)
		require.NoError(t, err)

		var unmarshaled TargetWrapper
		err = attributevalue.Unmarshal(av, &unmarshaled)
		require.NoError(t, err)

		result, ok := unmarshaled.Model.(*Preseed)
		require.True(t, ok)
		assert.Equal(t, preseed.Key, result.Key)
		assert.Equal(t, preseed.Username, result.Username)
		assert.Equal(t, preseed.Type, result.Type)
		assert.Equal(t, preseed.Title, result.Title)
		assert.Equal(t, preseed.Value, result.Value)
		assert.Equal(t, preseed.Status, result.Status)
		assert.Equal(t, preseed.Created, result.Created)
		assert.Equal(t, preseed.Visited, result.Visited)
		assert.Equal(t, preseed.TTL, result.TTL)
		assert.Equal(t, preseed.Capability, result.Capability)
	})

	t.Run("marshal and unmarshal webapplication", func(t *testing.T) {
		webapplication := NewWebApplication("https://example.com/", "Example App")
		webapplication.Username = "test@example.com"
		webapplication.Status = "A"
		webapplication.Created = Now()
		webapplication.Visited = Now()
		webapplication.TTL = 123456789
		webapplication.Source = "seed"
		webapplication.URLs = []string{"https://example.com/api", "https://example.com/admin"}
		webapplication.BurpSiteID = "1234"
		webapplication.BurpFolderID = "42"
		webapplication.BurpScheduleID = "abcd"

		original := TargetWrapper{Model: &webapplication}

		av, err := attributevalue.Marshal(original)
		require.NoError(t, err)

		var unmarshaled TargetWrapper
		err = attributevalue.Unmarshal(av, &unmarshaled)
		require.NoError(t, err)

		result, ok := unmarshaled.Model.(*WebApplication)
		require.True(t, ok)
		assert.Equal(t, webapplication.Key, result.Key)
		assert.Equal(t, webapplication.Username, result.Username)
		assert.Equal(t, webapplication.PrimaryURL, result.PrimaryURL)
		assert.Equal(t, webapplication.URLs, result.URLs)
		assert.Equal(t, webapplication.Status, result.Status)
		assert.Equal(t, webapplication.Source, result.Source)
		assert.Equal(t, webapplication.BurpSiteID, result.BurpSiteID)
		assert.Equal(t, webapplication.BurpFolderID, result.BurpFolderID)
		assert.Equal(t, webapplication.BurpScheduleID, result.BurpScheduleID)
	})

	t.Run("marshal and unmarshal webpage", func(t *testing.T) {
		asset := NewAsset("example.com", "1.2.3.4")
		attribute := NewAttribute("https", "443", &asset)
		webpage := NewWebpageFromString("https://example.com/", &attribute)
		webpage.Username = "test@example.com"
		webpage.Status = "A"
		webpage.Created = Now()
		webpage.Visited = Now()
		webpage.TTL = 123456789
		webpage.Source = []string{"crawler"}
		webpage.Metadata = map[string]any{"test": "value"}
		webpage.DetailsFilepath = "path/to/details.json"

		original := TargetWrapper{Model: &webpage}

		av, err := attributevalue.Marshal(original)
		require.NoError(t, err)

		var unmarshaled TargetWrapper
		err = attributevalue.Unmarshal(av, &unmarshaled)
		require.NoError(t, err)

		result, ok := unmarshaled.Model.(*Webpage)

		require.True(t, ok)
		assert.Equal(t, webpage.Key, result.Key)
		assert.Equal(t, webpage.Username, result.Username)
		assert.Equal(t, webpage.URL, result.URL)
		assert.Equal(t, webpage.Status, result.Status)
		assert.Equal(t, webpage.Source, result.Source)
		assert.Equal(t, webpage.Parent.Model, result.Parent.Model)
		assert.Equal(t, webpage.Metadata, result.Metadata)
		assert.Equal(t, webpage.Created, result.Created)
		assert.Equal(t, webpage.Visited, result.Visited)
		assert.Equal(t, webpage.TTL, result.TTL)
		assert.Equal(t, webpage.DetailsFilepath, result.DetailsFilepath)
	})

	t.Run("error on missing key", func(t *testing.T) {
		av, err := attributevalue.Marshal(map[string]string{
			"username": "test@example.com",
		})
		require.NoError(t, err)

		var event TargetWrapper
		err = attributevalue.Unmarshal(av, &event)
		assert.Error(t, err)
	})

	t.Run("error on invalid key type", func(t *testing.T) {
		av, err := attributevalue.Marshal(map[string]any{
			"key": 123,
		})
		require.NoError(t, err)

		var event TargetWrapper
		err = attributevalue.Unmarshal(av, &event)
		assert.Error(t, err)
	})

	t.Run("error on unknown target type", func(t *testing.T) {
		av, err := attributevalue.Marshal(map[string]string{
			"key": "#unknown#test",
		})
		require.NoError(t, err)

		var event TargetWrapper
		err = attributevalue.Unmarshal(av, &event)
		assert.Error(t, err)
	})
}

func TestTargetMarshal_Unmarshal(t *testing.T) {
	wrapper := TargetWrapper{
		Model: &Asset{},
	}
	data, err := json.Marshal(wrapper)
	assert.NoError(t, err)
	out := TargetWrapper{}
	err = json.Unmarshal(data, &out)
	assert.NoError(t, err)
	assert.Equal(t, "asset", out.Type)
}

// attempts to prevent a footgun when using a type embedded in other types - accidentally return the base type from
// WithStatus instead of the outer type
func TestTargetWithStatusTyping(t *testing.T) {
	types := registry.GetTypes[Target](registry.Registry)
	for _, tipe := range types {
		v, ok := registry.Registry.MakeType(tipe)
		assert.True(t, ok)

		before, ok := v.(Target)
		assert.True(t, ok)

		after := before.WithStatus("test")

		assert.IsType(t, before, after, "ensure that the WithStatus() method for %T returns an object of type %T", before, before)
	}
}
