package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTags_GetTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     Tags
		expected []string
	}{
		{
			name:     "empty tags",
			tags:     Tags{},
			expected: nil,
		},
		{
			name:     "single tag",
			tags:     Tags{Tags: []string{"production"}},
			expected: []string{"production"},
		},
		{
			name:     "multiple tags",
			tags:     Tags{Tags: []string{"production", "critical", "web"}},
			expected: []string{"production", "critical", "web"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tags.GetTags()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTags_AppendTags(t *testing.T) {
	tests := []struct {
		name     string
		initial  Tags
		toAppend []string
		expected []string
	}{
		{
			name:     "append to empty",
			initial:  Tags{},
			toAppend: []string{"production"},
			expected: []string{"production"},
		},
		{
			name:     "append single tag",
			initial:  Tags{Tags: []string{"production"}},
			toAppend: []string{"critical"},
			expected: []string{"production", "critical"},
		},
		{
			name:     "append multiple tags",
			initial:  Tags{Tags: []string{"production"}},
			toAppend: []string{"critical", "web", "database"},
			expected: []string{"production", "critical", "web", "database"},
		},
		{
			name:     "append duplicate tags (no deduplication in AppendTags)",
			initial:  Tags{Tags: []string{"production"}},
			toAppend: []string{"production", "critical"},
			expected: []string{"production", "production", "critical"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.AppendTags(tt.toAppend...)
			assert.Equal(t, tt.expected, tt.initial.Tags)
		})
	}
}

func TestTags_Merge(t *testing.T) {
	tests := []struct {
		name     string
		existing Tags
		update   Tags
		expected []string
	}{
		{
			name:     "merge into empty",
			existing: Tags{},
			update:   Tags{Tags: []string{"production", "critical"}},
			expected: []string{"production", "critical"},
		},
		{
			name:     "merge with empty",
			existing: Tags{Tags: []string{"production"}},
			update:   Tags{},
			expected: []string{"production"},
		},
		{
			name:     "merge replaces existing",
			existing: Tags{Tags: []string{"staging", "test"}},
			update:   Tags{Tags: []string{"production", "critical"}},
			expected: []string{"production", "critical"},
		},
		{
			name:     "merge with nil",
			existing: Tags{Tags: []string{"production"}},
			update:   Tags{Tags: nil},
			expected: []string{"production"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.existing.Merge(tt.update)
			assert.Equal(t, tt.expected, tt.existing.Tags)
		})
	}
}

func TestTags_Visit(t *testing.T) {
	tests := []struct {
		name     string
		existing Tags
		update   Tags
		expected []string
	}{
		{
			name:     "visit with empty tags",
			existing: Tags{Tags: []string{"production"}},
			update:   Tags{},
			expected: []string{"production"},
		},
		{
			name:     "visit empty with tags",
			existing: Tags{},
			update:   Tags{Tags: []string{"production", "critical"}},
			expected: []string{"production", "critical"},
		},
		{
			name:     "visit with new tags",
			existing: Tags{Tags: []string{"production"}},
			update:   Tags{Tags: []string{"critical", "web"}},
			expected: []string{"production", "critical", "web"},
		},
		{
			name:     "visit with duplicate tags",
			existing: Tags{Tags: []string{"production", "web"}},
			update:   Tags{Tags: []string{"production", "critical"}},
			expected: []string{"production", "web", "critical"},
		},
		{
			name:     "visit with all duplicate tags",
			existing: Tags{Tags: []string{"production", "critical"}},
			update:   Tags{Tags: []string{"production", "critical"}},
			expected: []string{"production", "critical"},
		},
		{
			name:     "visit preserves order",
			existing: Tags{Tags: []string{"a", "b", "c"}},
			update:   Tags{Tags: []string{"d", "e", "f"}},
			expected: []string{"a", "b", "c", "d", "e", "f"},
		},
		{
			name:     "visit with nil",
			existing: Tags{Tags: []string{"production"}},
			update:   Tags{Tags: nil},
			expected: []string{"production"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.existing.Visit(tt.update)
			assert.Equal(t, tt.expected, tt.existing.Tags)
		})
	}
}

func TestTags_VisitDeduplication(t *testing.T) {
	// Test that Visit properly deduplicates tags
	existing := Tags{Tags: []string{"tag1", "tag2"}}

	// First visit
	existing.Visit(Tags{Tags: []string{"tag3", "tag1"}})
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, existing.Tags, "should not duplicate tag1")

	// Second visit
	existing.Visit(Tags{Tags: []string{"tag4", "tag2", "tag5"}})
	assert.Equal(t, []string{"tag1", "tag2", "tag3", "tag4", "tag5"}, existing.Tags, "should not duplicate tag2")

	// Third visit with all duplicates
	existing.Visit(Tags{Tags: []string{"tag1", "tag2", "tag3"}})
	assert.Equal(t, []string{"tag1", "tag2", "tag3", "tag4", "tag5"}, existing.Tags, "should not add any duplicates")
}

func TestTaggableInterface(t *testing.T) {
	// Test that Tags implements Taggable interface
	var taggable Taggable = &Tags{}

	taggable.AppendTags("test1", "test2")
	tags := taggable.GetTags()

	assert.Equal(t, []string{"test1", "test2"}, tags)
}
