package entities

import (
	"fmt"
	"iyaem/internal/domain/events"
	vo "iyaem/internal/domain/valueobjects"
)

type Organization struct {
	id         vo.OrganizationId
	name       string
	identifier string
	tenants    []Tenant
	members    []Membership

	events []events.Event
}

func NewOrganization(
	id vo.OrganizationId,
	name string,
	identifier string,
	members []Membership,
	tenants []Tenant,
) Organization {
	return Organization{id, name, identifier, tenants, members, make([]events.Event, 0)}
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

func (o *Organization) Events() []events.Event {
	return o.events
}

func (o *Organization) AddMember(m Membership) {
	o.members = append(o.members, m)
	o.events = append(o.events, events.NewMemberAdded(m.id.Value(), m.UserId().Value(), string(m.level)))
}
