package commands

import (
	"context"
	"iyaem/internal/domain/repositories"
	"iyaem/internal/domain/valueobjects"
)

type AddRoleToMemberRequest struct {
	OrganizationId string `json:"organization_id"`
	MembershipId   string `json:"user_org_id"`
	RoleId         string `json:"role_id"`
	TenantId       string `json:"tenant_id"`
}

type AddRoleToMemberCommand struct {
	orgRepo repositories.OrganizationRepository
	memRepo repositories.MembershipRepository
}

func NewAddRoleToMemberCommand(
	orgRepo repositories.OrganizationRepository,
	memRepo repositories.MembershipRepository,
) *AddRoleToMemberCommand {
	return &AddRoleToMemberCommand{
		orgRepo: orgRepo,
		memRepo: memRepo,
	}
}

func (c *AddRoleToMemberCommand) Execute(ctx context.Context, r AddRoleToMemberRequest) (membershipId string, err error) {

	organization, err := c.orgRepo.FindById(ctx, r.OrganizationId)
	if err != nil || organization == nil {
		return "could not find organization", err
	}

	memberId, err := valueobjects.NewMembershipId(r.MembershipId)
	if err != nil {
		return "", err
	}

	roleId, err := valueobjects.NewRoleId(r.RoleId)
	if err != nil {
		return "", err
	}

	tenantId, err := valueobjects.NewTenantId(r.TenantId)
	if err != nil {
		return "", err
	}

	err = organization.AddRoleToMember(memberId, roleId, tenantId)
	if err != nil {
		return "", err
	}

	err = c.orgRepo.Update(ctx, organization)
	if err != nil {
		return "", err
	}

	return memberId.Value(), nil
}
