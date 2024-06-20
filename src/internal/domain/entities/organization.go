package entities

import (
	"fmt"
	vo "iyaem/internal/domain/valueobjects"
)

type Organization struct {
	id         vo.OrganizationId
	name       string
	identifier string
	tenants    []Tenant
	members    []Membership
}

func NewOrganization(
	id vo.OrganizationId,
	name string,
	identifier string,
	members []Membership,
	tenants []Tenant,
) Organization {
	return Organization{id, name, identifier, tenants, members}
}

func (o Organization) String() string {
	return fmt.Sprint(o.id.Value(), " ", o.name, "\nTenants: ", o.tenants)
}

func (o *Organization) Id() vo.OrganizationId {
	return o.id
}

func (o *Organization) Name() string {
	return o.name
}

func (o *Organization) Identifier() string {
	return o.identifier
}

func (o *Organization) Tenants() []Tenant {
	return o.tenants
}

func (o *Organization) Members() []Membership {
	return o.members
}
