package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type GraphModelWrapper registry.Wrapper[GraphModel]

func NewGraphModelWrapper(model GraphModel) GraphModelWrapper {
	return GraphModelWrapper{Type: registry.Name(model), Model: model}
}

func (t *GraphModelWrapper) UnmarshalJSON(data []byte) error {
	return (*registry.Wrapper[GraphModel])(t).UnmarshalJSON(data)
}

func (t *GraphModelWrapper) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	return (*registry.Wrapper[GraphModel])(t).UnmarshalDynamoDBAttributeValue(av)
}
