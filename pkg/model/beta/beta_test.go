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
