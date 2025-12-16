package model

import "slices"

type Taggable interface {
	GetTags() []string
	AppendTags(...string)
}

type Tags struct {
	Tags []string `json:"tags,omitempty" neo4j:"tags"`
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

func (t *Tags) GetTags() []string {
	return t.Tags
}

func (t *Tags) AppendTags(tags ...string) {
	t.Tags = append(t.Tags, tags...)
}
