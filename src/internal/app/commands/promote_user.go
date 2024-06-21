package commands

import (
	"context"
	"fmt"
	"iyaem/internal/domain/repositories"
	"iyaem/internal/domain/valueobjects"
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

	memberId, err := valueobjects.NewMembershipId(r.MembershipId)
	if err != nil {
		return "", err
	}

	member := organization.FindMemberById(memberId)
	if member == nil {
		return "", fmt.Errorf("could not find user")
	}

	err = organization.PromoteMember(*member)
	if err != nil {
		return "", fmt.Errorf("could not promote user")
	}

	err = c.orgRepo.Update(ctx, organization)
	if err != nil {
		return "", fmt.Errorf("could not promote user")
	}

	return member.Id().Value(), nil
}
