package beta

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Beta is a marker type that can be embedded anonymously in structs
// to mark them as beta features. When embedded, it automatically adds
// "beta": true to various marshaled output formats.
type Beta struct {
	Beta betaValue `neo4j:"beta" json:"beta" dynamodbav:"beta"`
}

func (b Beta) IsBeta() bool {
	return true
}

// betaValue is the actual type that implements the marshaling logic.
// when a Marshaler type is anonymously embedded, its marshaling logic
// hijacks the parent's marshaling logic. This will erase all of the 
// parent's other fields. Therefore, we must embed the actual marshaling
// logic one layer deeper than Beta.
type betaValue struct{}

func (b betaValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(true)
}

func (b betaValue) MarshalMap(out map[string]any) error {
	out["beta"] = true
	return nil
}

func (b betaValue) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberBOOL{Value: true}, nil
}
