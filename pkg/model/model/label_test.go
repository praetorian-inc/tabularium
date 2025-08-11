package model

import (
	"github.com/praetorian-inc/tabularium/pkg/model/label"
	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestLabels_Registered(t *testing.T) {
	modelRegistry := registry.Registry
	labelRegistry := label.GetRegistry()

	for name, modelType := range modelRegistry.GetAllTypes() {
		instance := reflect.New(modelType.Elem()).Interface()

		graphModel, ok := instance.(GraphModel)
		if !ok {
			continue
		}

		labels := graphModel.GetLabels()

		for _, label := range labels {
			if label == "" {
				continue
			}

			registeredLabel, exists := labelRegistry.Get(label)
			require.True(t, exists, "Label %q from model %q should be registered", label, name)

			assert.Equal(t, label, registeredLabel, "Registered label should match exactly for model %q", name)
		}
	}
}
