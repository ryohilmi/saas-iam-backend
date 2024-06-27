package entities

import (
	vo "iyaem/internal/domain/valueobjects"
)

type Tenant struct {
	id             vo.TenantId
	organizationId vo.OrganizationId
	applicationId  vo.ApplicationId
}

func NewTenant(id vo.TenantId, orgId vo.OrganizationId, appId vo.ApplicationId) Tenant {
	return Tenant{id, orgId, appId}
}

func (u *Tenant) Id() vo.TenantId {
	return u.id
}

func (u *Tenant) OrganizationId() vo.OrganizationId {
	return u.organizationId
}

func (u *Tenant) ApplicationId() vo.ApplicationId {
	return u.applicationId
}
