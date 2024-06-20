package commands

import (
	"context"
	"fmt"
	"iyaem/internal/domain/repositories"
)

type PromoteUserRequest struct {
	MembershipId   string `json:"user_org_id"`
	OrganizationId string `json:"organization_id"`
}

type PromoteUserCommand struct {
	orgRepo repositories.OrganizationRepository
	memRepo repositories.MembershipRepository
}

func NewPromoteUserCommand(
	orgRepo repositories.OrganizationRepository,
	memRepo repositories.MembershipRepository,
) *PromoteUserCommand {
	return &PromoteUserCommand{
		orgRepo: orgRepo,
		memRepo: memRepo,
	}
}

func (c *PromoteUserCommand) Execute(ctx context.Context, r PromoteUserRequest) (membershipId string, err error) {

	organization, err := c.orgRepo.FindById(ctx, r.OrganizationId)
	if err != nil || organization == nil {
		return "", fmt.Errorf("could not find organization")
	}

	member, err := c.memRepo.FindById(ctx, r.MembershipId)
	if err != nil || member == nil {
		return "", fmt.Errorf("could not find user")
	}

	organization.PromoteMember(*member)

	err = c.orgRepo.Update(ctx, organization)
	if err != nil {
		return "", fmt.Errorf("could not promote user")
	}

	return member.Id().Value(), nil
}
