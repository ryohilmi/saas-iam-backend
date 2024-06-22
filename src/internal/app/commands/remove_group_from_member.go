package commands

import (
	"context"
	"iyaem/internal/domain/repositories"
	"iyaem/internal/domain/valueobjects"
	"log"
)

type RemoveGroupFromMemberRequest struct {
	OrganizationId string `json:"organization_id"`
	MembershipId   string `json:"user_org_id"`
	GroupId        string `json:"group_id"`
	TenantId       string `json:"tenant_id"`
}

type RemoveGroupFromMemberCommand struct {
	orgRepo repositories.OrganizationRepository
}

func NewRemoveGroupFromMemberCommand(
	orgRepo repositories.OrganizationRepository,
) *RemoveGroupFromMemberCommand {
	return &RemoveGroupFromMemberCommand{
		orgRepo: orgRepo,
	}
}

func (c *RemoveGroupFromMemberCommand) Execute(ctx context.Context, r RemoveGroupFromMemberRequest) (membershipId string, err error) {

	organization, err := c.orgRepo.FindById(ctx, r.OrganizationId)
	if err != nil || organization == nil {
		log.Printf("Error: %v", err)
		return "could not find organization", err
	}

	memberId, err := valueobjects.NewMembershipId(r.MembershipId)
	if err != nil {
		return "", err
	}

	groupId, err := valueobjects.NewGroupId(r.GroupId)
	if err != nil {
		log.Printf("Error: %v", err)
		return "", err
	}

	tenantId, err := valueobjects.NewTenantId(r.TenantId)
	if err != nil {
		log.Printf("Error: %v", err)
		return "", err
	}

	err = organization.RemoveGroupFromMember(memberId, groupId, tenantId)
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
