package domain_test

import (
	"iyaem/internal/domain/entities"
	vo "iyaem/internal/domain/valueobjects"
	"testing"
)

func TestAddGroup(t *testing.T) {
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
	for _, m := range org.Members() {
		if m.Id().Equals(member.Id()) {
			for _, r := range m.Groups() {
				if r.Equals(group) {
					found = true
				}
			}
		}
	}

	if !found {
		t.Fatalf("AddGroup() failed, group not found")
	}
}

func TestAddSameGroup(t *testing.T) {
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
	err := org.AddGroupToMember(member.Id(), group.GroupId(), group.TenantId())

	if err == nil {
		t.Fatalf("AddGroup() failed, error not thrown")
	}

	if err.Error() != "group already exists" {
		t.Fatalf("AddGroup() failed, wrong error message")
	}
}

func TestRemoveGroup(t *testing.T) {
	orgId := vo.GenerateOrganizationId()
	org := entities.NewOrganization(
		orgId,
		"Test Corp",
		"test_corp",
		make([]entities.Membership, 0),
		make([]entities.Tenant, 0),
	)

	memId := vo.GenerateMembershipId()

	group := vo.NewUserGroup(
		memId,
		vo.GenerateGroupId(),
		vo.GenerateTenantId(),
	)

	member := entities.NewMembership(
		memId,
		vo.GenerateUserId(),
		orgId,
		"owner",
		make([]vo.UserRole, 0),
		[]vo.UserGroup{group},
	)

	org.AddMember(member)
	org.RemoveGroupFromMember(member.Id(), group.GroupId(), group.TenantId())

	found := false
	for _, m := range org.Members() {
		if m.Id().Equals(member.Id()) {
			for _, g := range m.Groups() {
				if g.Equals(group) {
					found = true
				}
			}
		}
	}

	if found {
		t.Fatalf("RemoveRole() failed, role found")
	}
}
