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

func (o *Organization) FindMemberById(membershipId vo.MembershipId) *Membership {
	for _, member := range o.members {
		if member.id == membershipId {
			return &member
		}
	}

	return nil
}

func (o *Organization) PromoteMember(m Membership) error {
	for i, member := range o.members {
		if member.id == m.id {
			o.members[i].level = vo.MembershipLevel("manager")
			o.events = append(o.events, events.NewMemberPromoted(m.id.Value()))
			return nil
		}
	}

	return fmt.Errorf("could not find member with id %v", m.id)
}

func (o *Organization) DemoteMember(m Membership) error {
	for i, member := range o.members {
		if member.id == m.id {
			o.members[i].level = vo.MembershipLevel("member")
			o.events = append(o.events, events.NewMemberDemoted(m.id.Value()))
			return nil
		}
	}

	return fmt.Errorf("could not find member with id %v", m.id)
}

func (o *Organization) AddRoleToMember(membershipId vo.MembershipId, roleId vo.RoleId, tenantId vo.TenantId) error {
	for i, member := range o.members {
		if member.id == membershipId {
			userRole := vo.NewUserRole(membershipId, roleId, tenantId)

			for _, role := range o.members[i].roles {
				if roleId == role.RoleId() && tenantId == role.TenantId() {
					return fmt.Errorf("role already exists")
				}
			}

			o.members[i].roles = append(o.members[i].roles, userRole)

			o.events = append(o.events, events.NewRoleAddedToMember(membershipId.Value(), roleId.Value(), tenantId.Value()))
			return nil
		}
	}

	return fmt.Errorf("could not find member with id %v", membershipId)
}

func (o *Organization) RemoveRoleFromMember(membershipId vo.MembershipId, roleId vo.RoleId, tenantId vo.TenantId) error {
	for i, member := range o.members {
		if member.id == membershipId {
			for j, role := range member.roles {
				if role.RoleId() == roleId && role.TenantId() == tenantId {
					o.members[i].roles = append(o.members[i].roles[:j], o.members[i].roles[j+1:]...)
					o.events = append(o.events, events.NewRoleRemovedFromMember(membershipId.Value(), roleId.Value(), tenantId.Value()))
					return nil
				}
			}
		}
	}

	return fmt.Errorf("could not find member with id %v", membershipId)
}
