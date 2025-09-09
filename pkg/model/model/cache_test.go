package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CacheConstruction(t *testing.T) {
	cache := NewCache("this", "is", "a", "test", "key")
	assert.Equal(t, "#cache#this#is#a#test#key", cache.Key)
}
