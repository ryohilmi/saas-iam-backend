package domain_test

import (
	"iyaem/internal/domain/entities"
	"iyaem/internal/domain/events"
	vo "iyaem/internal/domain/valueobjects"
	"testing"
)

func TestMemberAddedEvent(t *testing.T) {
	orgId := vo.GenerateOrganizationId()
	org := entities.NewOrganization(
		orgId,
		"Test Corp",
		"test_corp",
		make([]entities.Membership, 0),
		make([]entities.Tenant, 0),
	)

	userId := vo.GenerateUserId()
	memId := vo.GenerateMembershipId()
	member := entities.NewMembership(
		memId,
		userId,
		orgId,
		"owner",
		make([]vo.UserRole, 0),
		make([]vo.UserGroup, 0),
	)

	org.AddMember(member)

	found := false
	for _, e := range org.Events() {
		switch eventType := e.(type) {
		case events.MemberAdded:
			memberAdded := e.(events.MemberAdded)

			memberIdMatched := memberAdded.MembershipId == member.Id().Value()
			userIdMatched := memberAdded.UserId == userId.Value()

			if memberIdMatched && userIdMatched {
				found = true
			}

			_ = eventType
		}
	}

	if !found {
		t.Fatalf("AddMember() failed, member not found")
	}
}

func TestMemberPromotedEvent(t *testing.T) {
	orgId := vo.GenerateOrganizationId()
	org := entities.NewOrganization(
		orgId,
		"Test Corp",
		"test_corp",
		make([]entities.Membership, 0),
		make([]entities.Tenant, 0),
	)

	memId := vo.GenerateMembershipId()
	member := entities.NewMembership(
		memId,
		vo.GenerateUserId(),
		orgId,
		"member",
		make([]vo.UserRole, 0),
		make([]vo.UserGroup, 0),
	)

	org.AddMember(member)
	org.PromoteMember(member)

	found := false
	for _, e := range org.Events() {
		switch eventType := e.(type) {
		case events.MemberPromoted:
			memberPromoted := e.(events.MemberPromoted)

			memberIdMatched := memberPromoted.MembershipId == member.Id().Value()

			if memberIdMatched {
				found = true
			}

			_ = eventType
		}
	}

	if !found {
		t.Fatalf("PromoteMember() failed, member not found")
	}
}

func TestMemberDemotedEvent(t *testing.T) {
	orgId := vo.GenerateOrganizationId()
	org := entities.NewOrganization(
		orgId,
		"Test Corp",
		"test_corp",
		make([]entities.Membership, 0),
		make([]entities.Tenant, 0),
	)

	memId := vo.GenerateMembershipId()
	member := entities.NewMembership(
		memId,
		vo.GenerateUserId(),
		orgId,
		"owner",
		make([]vo.UserRole, 0),
		make([]vo.UserGroup, 0),
	)

	org.AddMember(member)
	org.DemoteMember(member)

	found := false
	for _, e := range org.Events() {
		switch eventType := e.(type) {
		case events.MemberDemoted:
			memberDemoted := e.(events.MemberDemoted)

			memberIdMatched := memberDemoted.MembershipId == member.Id().Value()

			if memberIdMatched {
				found = true
			}

			_ = eventType
		}
	}

	if !found {
		t.Fatalf("DemoteMember() failed, member not found")
	}
}

func TestRoleAddedToMemberEvent(t *testing.T) {
	orgId := vo.GenerateOrganizationId()
	org := entities.NewOrganization(
		orgId,
		"Test Corp",
		"test_corp",
		make([]entities.Membership, 0),
		make([]entities.Tenant, 0),
	)

	memId := vo.GenerateMembershipId()
	member := entities.NewMembership(
		memId,
		vo.GenerateUserId(),
		orgId,
		"owner",
		make([]vo.UserRole, 0),
		make([]vo.UserGroup, 0),
	)

	role := vo.NewUserRole(
		memId,
		vo.GenerateRoleId(),
		vo.GenerateTenantId(),
	)

	org.AddMember(member)
	org.AddRoleToMember(member.Id(), role.RoleId(), role.TenantId())

	found := false
	for _, e := range org.Events() {
		switch eventType := e.(type) {
		case events.RoleAddedToMember:
			roleAdded := e.(events.RoleAddedToMember)

			memberIdMatched := roleAdded.MembershipId == member.Id().Value()
			roleIdMatched := roleAdded.RoleId == role.RoleId().Value()

			if memberIdMatched && roleIdMatched {
				found = true
			}

			_ = eventType
		}
	}

	if !found {
		t.Fatalf("AddRole() failed, role not found")
	}
}

func TestRoleRemovedFromMemberEvent(t *testing.T) {
	orgId := vo.GenerateOrganizationId()
	org := entities.NewOrganization(
		orgId,
		"Test Corp",
		"test_corp",
		make([]entities.Membership, 0),
		make([]entities.Tenant, 0),
	)

	memId := vo.GenerateMembershipId()
	member := entities.NewMembership(
		memId,
		vo.GenerateUserId(),
		orgId,
		"owner",
		make([]vo.UserRole, 0),
		make([]vo.UserGroup, 0),
	)

	role := vo.NewUserRole(
		memId,
		vo.GenerateRoleId(),
		vo.GenerateTenantId(),
	)

	org.AddMember(member)
	org.AddRoleToMember(member.Id(), role.RoleId(), role.TenantId())
	org.RemoveRoleFromMember(member.Id(), role.RoleId(), role.TenantId())

	found := false
	for _, e := range org.Events() {
		switch eventType := e.(type) {
		case events.RoleRemovedFromMember:
			roleRemoved := e.(events.RoleRemovedFromMember)

			memberIdMatched := roleRemoved.MembershipId == member.Id().Value()
			roleIdMatched := roleRemoved.RoleId == role.RoleId().Value()

			if memberIdMatched && roleIdMatched {
				found = true
			}

			_ = eventType
		}
	}

	if !found {
		t.Fatalf("RemoveRole() failed, role not found")
	}
}

func TestGroupAddedToMemberEvent(t *testing.T) {
	orgId := vo.GenerateOrganizationId()
	org := entities.NewOrganization(
		orgId,
		"Test Corp",
		"test_corp",
		make([]entities.Membership, 0),
		make([]entities.Tenant, 0),
	)

	memId := vo.GenerateMembershipId()
	member := entities.NewMembership(
		memId,
		vo.GenerateUserId(),
		orgId,
		"owner",
		make([]vo.UserRole, 0),
		make([]vo.UserGroup, 0),
	)

	group := vo.NewUserGroup(
		memId,
		vo.GenerateGroupId(),
		vo.GenerateTenantId(),
	)

	org.AddMember(member)
	org.AddGroupToMember(member.Id(), group.GroupId(), group.TenantId())

	found := false
	for _, e := range org.Events() {
		switch eventType := e.(type) {
		case events.GroupAddedToMember:
			groupAdded := e.(events.GroupAddedToMember)

			memberIdMatched := groupAdded.MembershipId == member.Id().Value()
			groupIdMatched := groupAdded.GroupId == group.GroupId().Value()

			if memberIdMatched && groupIdMatched {
				found = true
			}

			_ = eventType
		}
	}

	if !found {
		t.Fatalf("AddGroup() failed, group not found")
	}
}

func TestGroupRemovedFromMemberEvent(t *testing.T) {
	orgId := vo.GenerateOrganizationId()
	org := entities.NewOrganization(
		orgId,
		"Test Corp",
		"test_corp",
		make([]entities.Membership, 0),
		make([]entities.Tenant, 0),
	)

	memId := vo.GenerateMembershipId()
	member := entities.NewMembership(
		memId,
		vo.GenerateUserId(),
		orgId,
		"owner",
		make([]vo.UserRole, 0),
		make([]vo.UserGroup, 0),
	)

	group := vo.NewUserGroup(
		memId,
		vo.GenerateGroupId(),
		vo.GenerateTenantId(),
	)

	org.AddMember(member)
	org.AddGroupToMember(member.Id(), group.GroupId(), group.TenantId())
	org.RemoveGroupFromMember(member.Id(), group.GroupId(), group.TenantId())

	found := false
	for _, e := range org.Events() {
		switch eventType := e.(type) {
		case events.GroupRemovedFromMember:
			groupRemoved := e.(events.GroupRemovedFromMember)

			memberIdMatched := groupRemoved.MembershipId == member.Id().Value()
			groupIdMatched := groupRemoved.GroupId == group.GroupId().Value()

			if memberIdMatched && groupIdMatched {
				found = true
			}

			_ = eventType
		}
	}

	if !found {
		t.Fatalf("RemoveGroup() failed, group not found")
	}
}
