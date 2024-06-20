package entities

import (
	vo "iyaem/internal/domain/valueobjects"
)

type Permission struct {
	id            vo.PermissionId
	applicationId vo.ApplicationId
	name          string
}

func NewPermission(id vo.PermissionId, applicationId vo.ApplicationId, name string) Permission {
	return Permission{id, applicationId, name}
}

func (u Permission) String() string {
	return u.id.Value() + " " + u.name
}

func (u *Permission) Id() vo.PermissionId {
	return u.id
}

func (u *Permission) Name() string {
	return u.name
}
