package commands

import (
	"context"
	"fmt"
	"iyaem/internal/domain/entities"
	"iyaem/internal/domain/repositories"
	"iyaem/internal/domain/valueobjects"
)

type CreateOrganizationRequest struct {
	Name       string
	Identifier string
	UserId     string
}

type CreateOrganizationCommand struct {
	orgRepo repositories.OrganizationRepository
}

func NewCreateOrganizationCommand(
	orgRepo repositories.OrganizationRepository,
) *CreateOrganizationCommand {
	return &CreateOrganizationCommand{
		orgRepo: orgRepo,
	}
}

func (c *CreateOrganizationCommand) Execute(ctx context.Context, r CreateOrganizationRequest) (orgId string, err error) {
	organizationId := valueobjects.GenerateOrganizationId()
	if err != nil {
		return "", fmt.Errorf("generate organization id: %w", err)
	}

	ownerOrgId, err := valueobjects.NewMembershipId(r.UserId)
	if err != nil {
		return "", fmt.Errorf("new membership id: %w", err)
	}

	ownerId, err := valueobjects.NewUserId(r.UserId)
	if err != nil {
		return "", fmt.Errorf("new user id: %w", err)
	}

	owner := entities.NewMembership(
		valueobjects.MembershipId(ownerOrgId),
		ownerId,
		organizationId,
		"owner",
		make([]entities.Role, 0),
	)

	members := make([]entities.Membership, 0)
	members = append(members, owner)

	organization := entities.NewOrganization(organizationId, r.Name, r.Identifier, members, make([]entities.Tenant, 0))

	err = c.orgRepo.Insert(ctx, &organization)
	if err != nil {
		return "", fmt.Errorf("insert organization: %w", err)
	}

	return organization.Id().Value(), nil
}
