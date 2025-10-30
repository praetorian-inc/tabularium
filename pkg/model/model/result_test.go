package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNoInput(t *testing.T) {
	testContext := ResultContext{
		Parent: TargetWrapper{
			Model: NewNoInput("test"),
		},
	}
	assert.True(t, IsNoInput(testContext.Parent.Model))
	assert.Nil(t, testContext.GetParent())
}
