package commands

import (
	"context"
	"fmt"
	"iyaem/internal/domain/entities"
	"iyaem/internal/domain/repositories"
	"iyaem/internal/domain/valueobjects"
)

type AddOrganizationUserRequest struct {
	Email          string `json:"email"`
	OrganizationId string `json:"organization_id"`
}

type AddOrganizationUserCommand struct {
	orgRepo  repositories.OrganizationRepository
	memRepo  repositories.MembershipRepository
	userRepo repositories.UserRepository
}

func NewAddOrganizationUserCommand(
	orgRepo repositories.OrganizationRepository,
	memRepo repositories.MembershipRepository,
	userRepo repositories.UserRepository,
) *AddOrganizationUserCommand {
	return &AddOrganizationUserCommand{
		orgRepo:  orgRepo,
		memRepo:  memRepo,
		userRepo: userRepo,
	}
}

func (c *AddOrganizationUserCommand) Execute(ctx context.Context, r AddOrganizationUserRequest) (membershipId string, err error) {

	organization, err := c.orgRepo.FindById(ctx, r.OrganizationId)
	if err != nil || organization == nil {
		return "", fmt.Errorf("could not find organization")
	}

	user, err := c.userRepo.FindByEmail(ctx, r.Email)
	if err != nil || user == nil {
		return "", fmt.Errorf("could not find user")
	}

	newMemberId := valueobjects.GenerateMembershipId()

	newMember := entities.NewMembership(
		newMemberId,
		user.Id(),
		organization.Id(),
		"member",
		make([]valueobjects.UserRole, 0),
	)

	organization.AddMember(newMember)

	err = c.orgRepo.Update(ctx, organization)
	if err != nil {
		return "", fmt.Errorf("could not add user to organization")
	}

	return newMemberId.Value(), nil
}
