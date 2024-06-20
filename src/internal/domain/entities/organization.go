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

func (o *Organization) PromoteMember(m Membership) {
	for i, member := range o.members {
		if member.id == m.id {
			o.members[i].level = vo.MembershipLevel("manager")
			o.events = append(o.events, events.NewMemberPromoted(m.id.Value()))
			return
		}
	}
}

func (o *Organization) DemoteMember(m Membership) {
	for i, member := range o.members {
		if member.id == m.id {
			o.members[i].level = vo.MembershipLevel("member")
			o.events = append(o.events, events.NewMemberDemoted(m.id.Value()))
			return
		}
	}
}

func (o *Organization) AddRoleToMember(m *Membership, roleId vo.RoleId, tenantId vo.TenantId) {
	for i, member := range o.members {
		if member.id == m.id {
			userRole := vo.NewUserRole(m.id, roleId, tenantId)

			o.members[i].roles = append(o.members[i].roles, userRole)
			o.events = append(o.events, events.NewRoleAddedToMember(m.id.Value(), roleId.Value(), tenantId.Value()))
			return
		}
	}
}

func (o *Organization) RemoveRoleFromMember(m *Membership, roleId vo.RoleId, tenantId vo.TenantId) {

	for i, member := range o.members {
		if member.id == m.id {
			fmt.Printf("\nMember %v: %v\n", m.id, m.Roles())

			for j, role := range member.roles {
				if role.RoleId() == roleId && role.TenantId() == tenantId {
					o.members[i].roles = append(o.members[i].roles[:j], o.members[i].roles[j+1:]...)
					o.events = append(o.events, events.NewRoleRemovedFromMember(m.id.Value(), roleId.Value(), tenantId.Value()))
					return
				}
			}

			fmt.Printf("\nMember %v: %v\n", member.id, member.Roles())
		}
	}
}
