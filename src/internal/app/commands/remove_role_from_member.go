package commands

import (
	"context"
	"iyaem/internal/domain/repositories"
	"iyaem/internal/domain/valueobjects"
	"log"
)

type RemoveRoleFromMemberRequest struct {
	OrganizationId string `json:"organization_id"`
	MembershipId   string `json:"user_org_id"`
	RoleId         string `json:"role_id"`
	TenantId       string `json:"tenant_id"`
}

type RemoveRoleFromMemberCommand struct {
	orgRepo repositories.OrganizationRepository
	memRepo repositories.MembershipRepository
}

func NewRemoveRoleFromMemberCommand(
	orgRepo repositories.OrganizationRepository,
	memRepo repositories.MembershipRepository,
) *RemoveRoleFromMemberCommand {
	return &RemoveRoleFromMemberCommand{
		orgRepo: orgRepo,
		memRepo: memRepo,
	}
}

func (c *RemoveRoleFromMemberCommand) Execute(ctx context.Context, r RemoveRoleFromMemberRequest) (membershipId string, err error) {

	organization, err := c.orgRepo.FindById(ctx, r.OrganizationId)
	if err != nil || organization == nil {
		log.Printf("Error: %v", err)
		return "could not find organization", err
	}

	memberId, err := valueobjects.NewMembershipId(r.MembershipId)
	if err != nil {
		return "", err
	}

	roleId, err := valueobjects.NewRoleId(r.RoleId)
	if err != nil {
		log.Printf("Error: %v", err)
		return "", err
	}

	tenantId, err := valueobjects.NewTenantId(r.TenantId)
	if err != nil {
		log.Printf("Error: %v", err)
		return "", err
	}

	err = organization.RemoveRoleFromMember(memberId, roleId, tenantId)
	if err != nil {
		log.Printf("Error: %v", err)
		return "", err
	}

	err = c.orgRepo.Update(ctx, organization)
	if err != nil {
		log.Printf("Error: %v", err)
		return "", err
	}

	return memberId.Value(), nil
}
