package commands

import (
	"context"
	"fmt"
	"iyaem/internal/domain/repositories"
	"iyaem/internal/domain/valueobjects"
)

type DemoteUserRequest struct {
	MembershipId   string `json:"user_org_id"`
	OrganizationId string `json:"organization_id"`
}

type DemoteUserCommand struct {
	orgRepo repositories.OrganizationRepository
	memRepo repositories.MembershipRepository
}

func NewDemoteUserCommand(
	orgRepo repositories.OrganizationRepository,
	memRepo repositories.MembershipRepository,
) *DemoteUserCommand {
	return &DemoteUserCommand{
		orgRepo: orgRepo,
		memRepo: memRepo,
	}
}

func (c *DemoteUserCommand) Execute(ctx context.Context, r DemoteUserRequest) (membershipId string, err error) {

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

	err = organization.DemoteMember(*member)
	if err != nil {
		return "", fmt.Errorf("could not demote user")
	}

	err = c.orgRepo.Update(ctx, organization)
	if err != nil {
		return "", fmt.Errorf("could not demote user")
	}

	return member.Id().Value(), nil
}
