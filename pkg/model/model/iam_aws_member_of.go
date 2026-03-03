package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const IAMAWSMemberOfLabel = "IAM_AWS_MEMBER_OF"

func init() {
	registry.Registry.MustRegisterModel(&IAMAWSMemberOf{}, IAMAWSMemberOfLabel)
}

type IAMAWSMemberOf struct {
	*BaseRelationship
}

func NewIAMAWSMemberOf(source, target GraphModel) *IAMAWSMemberOf {
	return &IAMAWSMemberOf{
		BaseRelationship: NewBaseRelationship(source, target, IAMAWSMemberOfLabel),
	}
}

func (r *IAMAWSMemberOf) Label() string {
	return IAMAWSMemberOfLabel
}

func (r *IAMAWSMemberOf) GetDescription() string {
	return "Represents a group membership relationship between an IAM user and an IAM group."
}

func (r *IAMAWSMemberOf) Visit(other GraphRelationship) {
	r.BaseRelationship.Visit(other)
}
