package model

import (
	"encoding/json"
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
	"testing"

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

	t.Run("unmarshal port", func(t *testing.T) {
		input := `{
			"key": "#port#tcp#443#asset#example.com#1.2.3.4",
			"username": "test@example.com",
			"protocol": "tcp",
			"port": 443,
			"service": "https",
			"source": "#asset#example.com#1.2.3.4",
			"status": "A"
		}`

		var event TargetWrapper
		err := json.Unmarshal([]byte(input), &event)
		require.NoError(t, err)

		port, ok := event.Model.(*Port)
		require.True(t, ok, "expected Target to be *Port")
		assert.Equal(t, "#port#tcp#443#asset#example.com#1.2.3.4", port.Key)
		assert.Equal(t, "test@example.com", port.Username)
		assert.Equal(t, "tcp", port.Protocol)
		assert.Equal(t, 443, port.Port)
		assert.Equal(t, "https", port.Service)
		assert.Equal(t, "#asset#example.com#1.2.3.4", port.Source)
		assert.Equal(t, "A", port.Status)
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
	port := NewPort("tcp", 443, &asset)
	preseed := NewPreseed("whois+company", "Chariot Systems", "Chariot Systems")
	webpage := NewWebpageFromString("https://example.com/", nil)

	testTargetInterface(t, "Asset", &asset)
	testTargetInterface(t, "Port", &port)
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

	t.Run("marshal and unmarshal port", func(t *testing.T) {
		asset := Asset{BaseAsset: BaseAsset{Key: "#asset#example.com#1.2.3.4"}}
		port := NewPort("tcp", 443, &asset)
		port.Username = "test@example.com"
		port.Service = "https"
		port.Status = "A"
		port.Created = Now()
		port.Visited = Now()
		port.TTL = 123456789
		port.Capability = "portscan"

		original := TargetWrapper{Model: &port}

		av, err := attributevalue.Marshal(original)
		require.NoError(t, err)

		var unmarshaled TargetWrapper
		err = attributevalue.Unmarshal(av, &unmarshaled)
		require.NoError(t, err)

		result, ok := unmarshaled.Model.(*Port)
		require.True(t, ok)
		assert.Equal(t, port.Key, result.Key)
		assert.Equal(t, port.Username, result.Username)
		assert.Equal(t, port.Protocol, result.Protocol)
		assert.Equal(t, port.Port, result.Port)
		assert.Equal(t, port.Service, result.Service)
		assert.Equal(t, port.Source, result.Source)
		assert.Equal(t, port.Status, result.Status)
		assert.Equal(t, port.Created, result.Created)
		assert.Equal(t, port.Visited, result.Visited)
		assert.Equal(t, port.TTL, result.TTL)
		assert.Equal(t, port.Capability, result.Capability)
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
	})

	t.Run("marshal and unmarshal webpage", func(t *testing.T) {
		webpage := NewWebpageFromString("https://example.com/", nil)
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
		assert.Equal(t, webpage.Parent, result.Parent)
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
	types := model.GetTypes[Target](shared.Registry)
	for _, tipe := range types {
		v, ok := shared.Registry.MakeType(tipe)
		require.True(t, ok)

		before, ok := v.(Target)
		require.True(t, ok)

		after := before.WithStatus("test")

		assert.IsType(t, before, after, "ensure that the WithStatus() method for %T returns an object of type %T", before, before)
	}
}

func TestTargetWithStatusDoesNotModifyOriginal(t *testing.T) {
	types := model.GetTypes[Target](shared.Registry)
	for _, tipe := range types {
		v, ok := shared.Registry.MakeType(tipe)
		require.True(t, ok)

		before, ok := v.(Target)
		require.True(t, ok)

		after := before.WithStatus("new-status")

		assert.NotEqual(t, before.GetStatus(), "new-status", "%T.WithStatus modified original object's status", before)
		assert.Equal(t, after.GetStatus(), "new-status", "%T.WithStatus failed to modify new object's object's status", after)
	}
}
