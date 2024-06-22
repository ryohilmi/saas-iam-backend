package commands

import (
	"context"
	"iyaem/internal/domain/repositories"
	"iyaem/internal/domain/valueobjects"
)

type AddGroupToMemberRequest struct {
	OrganizationId string `json:"organization_id"`
	MembershipId   string `json:"user_org_id"`
	GroupId        string `json:"group_id"`
	TenantId       string `json:"tenant_id"`
}

type AddGroupToMemberCommand struct {
	orgRepo repositories.OrganizationRepository
}

func NewAddGroupToMemberCommand(
	orgRepo repositories.OrganizationRepository,
) *AddGroupToMemberCommand {
	return &AddGroupToMemberCommand{
		orgRepo: orgRepo,
	}
}

func (c *AddGroupToMemberCommand) Execute(ctx context.Context, r AddGroupToMemberRequest) (membershipId string, err error) {

	organization, err := c.orgRepo.FindById(ctx, r.OrganizationId)
	if err != nil || organization == nil {
		return "could not find organization", err
	}

	memberId, err := valueobjects.NewMembershipId(r.MembershipId)
	if err != nil {
		return "", err
	}

	groupId, err := valueobjects.NewGroupId(r.GroupId)
	if err != nil {
		return "", err
	}

	tenantId, err := valueobjects.NewTenantId(r.TenantId)
	if err != nil {
		return "", err
	}

	err = organization.AddGroupToMember(memberId, groupId, tenantId)
	if err != nil {
		return "", err
	}

	err = c.orgRepo.Update(ctx, organization)
	if err != nil {
		return "", err
	}

	return memberId.Value(), nil
}
