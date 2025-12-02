package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/wrapper"
)

type GraphModelWrapper wrapper.Wrapper[GraphModel]

func NewGraphModelWrapper(m GraphModel) GraphModelWrapper {
	return GraphModelWrapper{Type: model.Name(m), Model: m}
}

func (t *GraphModelWrapper) UnmarshalJSON(data []byte) error {
	return (*wrapper.Wrapper[GraphModel])(t).UnmarshalJSON(data)
}

func (t *GraphModelWrapper) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	return (*wrapper.Wrapper[GraphModel])(t).UnmarshalDynamoDBAttributeValue(av)
}
