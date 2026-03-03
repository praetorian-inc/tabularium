package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const IAMAWSHasInstanceProfileLabel = "IAM_AWS_HAS_INSTANCE_PROFILE"

func init() {
	registry.Registry.MustRegisterModel(&IAMAWSHasInstanceProfile{}, IAMAWSHasInstanceProfileLabel)
}

type IAMAWSHasInstanceProfile struct {
	*BaseRelationship
}

func NewIAMAWSHasInstanceProfile(source, target GraphModel) *IAMAWSHasInstanceProfile {
	return &IAMAWSHasInstanceProfile{
		BaseRelationship: NewBaseRelationship(source, target, IAMAWSHasInstanceProfileLabel),
	}
}

func (r *IAMAWSHasInstanceProfile) Label() string {
	return IAMAWSHasInstanceProfileLabel
}

func (r *IAMAWSHasInstanceProfile) GetDescription() string {
	return "Represents an instance profile association between an IAM role and an instance profile."
}

func (r *IAMAWSHasInstanceProfile) Visit(other GraphRelationship) {
	r.BaseRelationship.Visit(other)
}
