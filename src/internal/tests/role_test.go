package domain_test

import (
	"iyaem/internal/domain/entities"
	vo "iyaem/internal/domain/valueobjects"
	"testing"
)

func TestAddRole(t *testing.T) {
	orgId := vo.GenerateOrganizationId()
	org := entities.NewOrganization(
		orgId,
		"Test Corp",
		"test_corp",
		make([]entities.Membership, 0),
		make([]entities.Tenant, 0),
	)

	memId := vo.GenerateMembershipId()
	groups := make([]vo.UserGroup, 0)
	member := entities.NewMembership(
		memId,
		vo.GenerateUserId(),
		orgId,
		"owner",
		make([]vo.UserRole, 0),
		groups,
	)

	role := vo.NewUserRole(
		memId,
		vo.GenerateRoleId(),
		vo.GenerateTenantId(),
	)

	org.AddMember(member)
	org.AddRoleToMember(member.Id(), role.RoleId(), role.TenantId())

	found := false
	for _, m := range org.Members() {
		if m.Id().Equals(member.Id()) {
			for _, r := range m.Roles() {
				if r.Equals(role) {
					found = true
				}
			}
		}
	}

	if !found {
		t.Fatalf("AddRole() failed, role not found")
	}
}

func TestAddSameRole(t *testing.T) {
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
	err := org.AddRoleToMember(member.Id(), role.RoleId(), role.TenantId())

	if err == nil {
		t.Fatalf("AddRole() failed, error not thrown")
	}

	if err.Error() != "role already exists" {
		t.Fatalf("AddRole() failed, wrong error message")
	}
}

func TestRemoveRole(t *testing.T) {
	orgId := vo.GenerateOrganizationId()
	org := entities.NewOrganization(
		orgId,
		"Test Corp",
		"test_corp",
		make([]entities.Membership, 0),
		make([]entities.Tenant, 0),
	)

	memId := vo.GenerateMembershipId()

	role := vo.NewUserRole(
		memId,
		vo.GenerateRoleId(),
		vo.GenerateTenantId(),
	)

	member := entities.NewMembership(
		memId,
		vo.GenerateUserId(),
		orgId,
		"owner",
		[]vo.UserRole{role},
		make([]vo.UserGroup, 0),
	)

	org.AddMember(member)
	org.RemoveRoleFromMember(member.Id(), role.RoleId(), role.TenantId())

	found := false
	for _, m := range org.Members() {
		if m.Id().Equals(member.Id()) {
			for _, r := range m.Roles() {
				if r.Equals(role) {
					found = true
				}
			}
		}
	}

	if found {
		t.Fatalf("RemoveRole() failed, role found")
	}
}
