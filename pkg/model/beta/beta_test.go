package beta_test

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/praetorian-inc/tabularium/pkg/model/beta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestBetaType struct {
	beta.Beta
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestBeta_MarshalJSON(t *testing.T) {
	t.Run("standalone beta marshals correctly", func(t *testing.T) {
		b := beta.Beta{}
		data, err := json.Marshal(b)
		require.NoError(t, err)

		expected := `{"beta":true}`
		assert.JSONEq(t, expected, string(data))
	})

	t.Run("embedded beta adds beta field to JSON", func(t *testing.T) {
		testObj := TestBetaType{
			ID:   "test-id",
			Name: "Test Name",
		}

		data, err := json.Marshal(testObj)
		require.NoError(t, err)
		assert.JSONEq(t, `{"beta":true,"id":"test-id","name":"Test Name"}`, string(data))
	})
}

func TestBeta_MarshalDynamoDBAttributeValue(t *testing.T) {
	t.Run("MarshalDynamoDBAttributeValue returns boolean true", func(t *testing.T) {
		b := beta.Beta{}

		attributes, err := attributevalue.MarshalMap(b)
		require.NoError(t, err)

		attribute, ok := attributes["beta"].(*types.AttributeValueMemberBOOL)

		require.True(t, ok)
		assert.Equal(t, true, attribute.Value)
	})

	t.Run("embedded beta adds beta field to DynamoDB", func(t *testing.T) {
		testObj := TestBetaType{
			ID:   "test-id",
			Name: "Test Name",
		}

		attributes, err := attributevalue.MarshalMap(testObj)
		require.NoError(t, err)

		attribute, ok := attributes["beta"].(*types.AttributeValueMemberBOOL)
		require.True(t, ok)
		assert.Equal(t, true, attribute.Value)
	})
}

func TestBeta_UnmarshalJSON(t *testing.T) {
	t.Run("simple unmarshal", func(t *testing.T) {
		b := beta.Beta{}

		data := []byte(`{}`)
		err := json.Unmarshal(data, &b)
		require.NoError(t, err)
		assert.True(t, b.IsBeta())
	})

	t.Run("simple unmarshal should ignore true beta value", func(t *testing.T) {
		b := beta.Beta{}

		data := []byte(`{"beta":true}`)
		err := json.Unmarshal(data, &b)
		require.NoError(t, err)
		assert.True(t, b.IsBeta())
	})

	t.Run("simple unmarshal should ignore false beta value", func(t *testing.T) {
		b := beta.Beta{}

		data := []byte(`{"beta":false}`)
		err := json.Unmarshal(data, &b)
		require.NoError(t, err)
		assert.True(t, b.IsBeta())
	})

	t.Run("embedded unmarshal", func(t *testing.T) {
		b := TestBetaType{}

		data := []byte(`{"id":"test-id","name":"Test Name"}`)
		err := json.Unmarshal(data, &b)
		require.NoError(t, err)
		assert.True(t, b.IsBeta())
	})

	t.Run("embedded unmarshal should ignore true beta value", func(t *testing.T) {
		b := TestBetaType{}

		data := []byte(`{"beta":true,"id":"test-id","name":"Test Name"}`)
		err := json.Unmarshal(data, &b)
		require.NoError(t, err)
		assert.True(t, b.IsBeta())
	})

	t.Run("embedded unmarshal should ignore false beta value", func(t *testing.T) {
		b := TestBetaType{}

		data := []byte(`{"beta":false,"id":"test-id","name":"Test Name"}`)
		err := json.Unmarshal(data, &b)
		require.NoError(t, err)
		assert.True(t, b.IsBeta())
	})
}

func TestBeta_UnmarshalDynamoDBAttributeValue(t *testing.T) {
	t.Run("simple unmarshal", func(t *testing.T) {
		avs := map[string]types.AttributeValue{}

		b := beta.Beta{}
		err := attributevalue.UnmarshalMap(avs, &b)
		require.NoError(t, err)
		assert.True(t, b.IsBeta())
	})

	t.Run("simple unmarshal should ignore true beta value", func(t *testing.T) {
		avs := map[string]types.AttributeValue{
			"beta": &types.AttributeValueMemberBOOL{Value: true},
		}

		b := beta.Beta{}
		err := attributevalue.UnmarshalMap(avs, &b)
		require.NoError(t, err)
		assert.True(t, b.IsBeta())
	})

	t.Run("simple unmarshal should ignore false beta value", func(t *testing.T) {
		avs := map[string]types.AttributeValue{
			"beta": &types.AttributeValueMemberBOOL{Value: false},
		}

		b := beta.Beta{}
		err := attributevalue.UnmarshalMap(avs, &b)
		require.NoError(t, err)
		assert.True(t, b.IsBeta())
	})

	t.Run("embedded unmarshal", func(t *testing.T) {
		avs := map[string]types.AttributeValue{
			"id":   &types.AttributeValueMemberS{Value: "test-id"},
			"name": &types.AttributeValueMemberS{Value: "Test Name"},
		}

		b := TestBetaType{}
		err := attributevalue.UnmarshalMap(avs, &b)
		require.NoError(t, err)
		assert.True(t, b.IsBeta())
	})

	t.Run("embedded unmarshal should ignore true beta value", func(t *testing.T) {
		avs := map[string]types.AttributeValue{
			"beta": &types.AttributeValueMemberBOOL{Value: true},
			"id":   &types.AttributeValueMemberS{Value: "test-id"},
			"name": &types.AttributeValueMemberS{Value: "Test Name"},
		}

		b := TestBetaType{}
		err := attributevalue.UnmarshalMap(avs, &b)
		require.NoError(t, err)
		assert.True(t, b.IsBeta())
	})

	t.Run("embedded unmarshal should ignore false beta value", func(t *testing.T) {
		avs := map[string]types.AttributeValue{
			"beta": &types.AttributeValueMemberBOOL{Value: false},
			"id":   &types.AttributeValueMemberS{Value: "test-id"},
			"name": &types.AttributeValueMemberS{Value: "Test Name"},
		}

		b := TestBetaType{}
		err := attributevalue.UnmarshalMap(avs, &b)
		require.NoError(t, err)
		assert.True(t, b.IsBeta())
	})
}
