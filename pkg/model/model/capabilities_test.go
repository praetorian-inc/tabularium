package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"slices"
	"testing"
)

func TestAssetlike_Implementations(t *testing.T) {
	for name := range shared.Registry.GetAllTypes() {
		item, ok := shared.Registry.MakeType(name)
		require.True(t, ok)

		graphModel, ok := item.(GraphModel)
		if !ok {
			continue
		}

		if !slices.Contains(graphModel.GetLabels(), AssetLabel) {
			continue
		}

		_, ok = item.(Assetlike)
		require.True(t, ok, "failed to cast %q to Assetlike", name)
	}
}

func TestObjectLabelsMatchNames(t *testing.T) {
	for name := range shared.Registry.GetAllTypes() {
		model, ok := shared.Registry.MakeType(name)
		require.True(t, ok)

		tipe := reflect.TypeOf(model)
		if tipe.Kind() == reflect.Ptr {
			tipe = tipe.Elem()
		}

		graphModel, ok := model.(GraphModel)
		if !ok {
			continue
		}

		name := tipe.Name()
		if name == "NoInput" {
			continue // NoInput does not actually get put into Neo4j, it just needs to satisfy the GraphModel interface to work with Chariot
		}

		model.CallHooks(graphModel)

		assert.Contains(t, graphModel.GetLabels(), name, "%s does not have a label that matches its name", name)
	}
}
