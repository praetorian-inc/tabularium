package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const IAMAWSAttachedPermissionLabel = "IAM_AWS_ATTACHED_PERMISSION"

func init() {
	registry.Registry.MustRegisterModel(&IAMAWSAttachedPermission{}, IAMAWSAttachedPermissionLabel)
}

type IAMAWSAttachedPermission struct {
	*BaseRelationship
}

func NewIAMAWSAttachedPermission(source, target GraphModel) *IAMAWSAttachedPermission {
	return &IAMAWSAttachedPermission{
		BaseRelationship: NewBaseRelationship(source, target, IAMAWSAttachedPermissionLabel),
	}
}

func (r *IAMAWSAttachedPermission) Label() string {
	return IAMAWSAttachedPermissionLabel
}

func (r *IAMAWSAttachedPermission) GetDescription() string {
	return "Represents an attached managed policy relationship between an IAM principal and a managed policy."
}

func (r *IAMAWSAttachedPermission) Visit(other GraphRelationship) {
	r.BaseRelationship.Visit(other)
}
