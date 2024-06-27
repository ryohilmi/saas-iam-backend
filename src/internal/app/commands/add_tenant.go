package commands

import (
	"context"
	"fmt"
	"iyaem/internal/domain/entities"
	"iyaem/internal/domain/repositories"
	"iyaem/internal/domain/valueobjects"
)

type AddTenantRequest struct {
	OrganizationId string `json:"organization_id"`
	TenantId       string `json:"tenant_id"`
	ApplicationId  string `json:"application_id"`
}

type AddTenantCommand struct {
	orgRepo repositories.OrganizationRepository
}

func NewAddTenantCommand(
	orgRepo repositories.OrganizationRepository,
) *AddTenantCommand {
	return &AddTenantCommand{
		orgRepo: orgRepo,
	}
}

func (c *AddTenantCommand) Execute(ctx context.Context, r AddTenantRequest) (membershipId string, err error) {

	organization, err := c.orgRepo.FindById(ctx, r.OrganizationId)
	if err != nil || organization == nil {
		return "", fmt.Errorf("could not find organization")
	}

	newTenantId, err := valueobjects.NewTenantId(r.TenantId)
	if err != nil {
		return "", err
	}

	newApplicationId, err := valueobjects.NewApplicationId(r.ApplicationId)
	if err != nil {
		return "", err
	}

	newTenant := entities.NewTenant(
		newTenantId,
		organization.Id(),
		newApplicationId,
	)

	organization.AddTenant(newTenant)

	err = c.orgRepo.Update(ctx, organization)
	if err != nil {
		return "", fmt.Errorf("could not update organization: %s", err)
	}

	return newTenantId.Value(), nil
}
