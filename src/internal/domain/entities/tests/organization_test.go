package entities

import (
	"iyaem/internal/domain/entities"
	vo "iyaem/internal/domain/valueobjects"
	"testing"
)

func TestAddMember(t *testing.T) {
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
	)

	org.AddMember(member)

	found := false
	for _, m := range org.Members() {
		if m.Id().Equals(member.Id()) {
			found = true
		}
	}

	if !found {
		t.Fatalf("AddMember() failed, member not found")
	}
}

func TestPromoteMember(t *testing.T) {
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
	)

	org.AddMember(member)
	org.PromoteMember(member)

	found := false
	for _, m := range org.Members() {
		if m.Id().Equals(member.Id()) && m.Level() == "manager" {
			found = true
		}
	}

	if !found {
		t.Fatalf("PromoteMember() failed, member not found")
	}
}

func TestDemoteMember(t *testing.T) {
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
		"manager",
		make([]vo.UserRole, 0),
	)

	org.AddMember(member)
	org.DemoteMember(member)

	found := false
	for _, m := range org.Members() {
		if m.Id().Equals(member.Id()) && m.Level() == "member" {
			found = true
		}
	}

	if !found {
		t.Fatalf("DemoteMember() failed, member not found")
	}
}

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
	member := entities.NewMembership(
		memId,
		vo.GenerateUserId(),
		orgId,
		"owner",
		make([]vo.UserRole, 0),
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
