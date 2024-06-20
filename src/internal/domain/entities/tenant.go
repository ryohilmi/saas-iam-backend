package entities

import (
	"fmt"
	vo "iyaem/internal/domain/valueobjects"
)

type Tenant struct {
	id             vo.TenantId
	name           string
	organizationId vo.OrganizationId
	applicationId  vo.ApplicationId
	roles          []Role
}

func NewTenant(id vo.TenantId, name string, orgId vo.OrganizationId, appId vo.ApplicationId, roles []Role) Tenant {
	return Tenant{id, name, orgId, appId, roles}
}

func (u Tenant) String() string {
	return fmt.Sprint(u.id.Value(), " ", u.name, "\n", "Roles: ", u.roles, "\n")
}

func (u *Tenant) Id() vo.TenantId {
	return u.id
}

func (u *Tenant) Name() string {
	return u.name
}

func (u *Tenant) OrganizationId() vo.OrganizationId {
	return u.organizationId
}

func (u *Tenant) ApplicationId() vo.ApplicationId {
	return u.applicationId
}

func (u *Tenant) Roles() []Role {
	return u.roles
}

func (u *Tenant) AddRole(r Role) {
	u.roles = append(u.roles, r)
}
