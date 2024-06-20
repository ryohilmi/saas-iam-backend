package entities

import (
	vo "iyaem/internal/domain/valueobjects"
)

type Membership struct {
	id             vo.MembershipId
	userId         vo.UserId
	organizationId vo.OrganizationId
	level          vo.MembershipLevel
	roles          []Role
}

func NewMembership(
	id vo.MembershipId,
	userId vo.UserId,
	organizaitonId vo.OrganizationId,
	level vo.MembershipLevel,
	roles []Role,
) Membership {
	return Membership{id, userId, organizaitonId, level, roles}
}

func (u Membership) String() string {
	return u.userId.Value() + " " + u.organizationId.Value() + " " + string(u.level)
}

func (u Membership) Id() vo.MembershipId {
	return u.id
}

func (u Membership) UserId() vo.UserId {
	return u.userId
}

func (u Membership) OrganizationId() vo.OrganizationId {
	return u.organizationId
}

func (u Membership) Level() vo.MembershipLevel {
	return u.level
}

func (u Membership) Roles() []Role {
	return u.roles
}
