package entities

import (
	"fmt"
	vo "iyaem/internal/domain/valueobjects"
)

type Role struct {
	id            vo.RoleId
	name          string
	applicationId vo.ApplicationId
	permissions   []Permission
}

func NewRole(id vo.RoleId, name string, applicationId vo.ApplicationId, permissions []Permission) Role {
	return Role{id, name, applicationId, permissions}
}

func (u Role) String() string {
	return fmt.Sprint(u.id.Value()+" "+u.name, "\n", "Permissions: ", u.permissions, "\n")
}

func (u *Role) Id() vo.RoleId {
	return u.id
}

func (u *Role) Name() string {
	return u.name
}

func (u *Role) Permissions() []Permission {
	return u.permissions
}

func (u *Role) AddPermission(p Permission) {
	u.permissions = append(u.permissions, p)
}
