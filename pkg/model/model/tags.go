package model

import "slices"

type Tags struct {
	Tags []string `json:"tags,omitempty" neo4j:"tags,omitempty"`
}

func (t *Tags) Merge(other Tags) {
	if other.Tags != nil {
		t.Tags = other.Tags
	}
}

func (t *Tags) Visit(other Tags) {
	for _, tag := range other.Tags {
		if !slices.Contains(t.Tags, tag) {
			t.Tags = append(t.Tags, tag)
		}
	}
}
