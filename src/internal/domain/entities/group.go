package entities

import (
	"fmt"
	vo "iyaem/internal/domain/valueobjects"
)

type Group struct {
	id            vo.GroupId
	name          string
	description   string
	applicationId vo.ApplicationId
	roles         []Role
}

func NewGroup(id vo.GroupId, name string, description string, applicationId vo.ApplicationId, roles []Role) Group {
	return Group{id, name, description, applicationId, roles}
}

func (u Group) String() string {
	return fmt.Sprint(u.id.Value()+" "+u.name, "\n", "Roles: ", u.roles, "\n")
}

func (u *Group) Id() vo.GroupId {
	return u.id
}

func (u *Group) Name() string {
	return u.name
}

func (u *Group) Roles() []Permission {
	return u.Roles()
}
