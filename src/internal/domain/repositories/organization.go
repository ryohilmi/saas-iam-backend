package repositories

import (
	"context"
	"iyaem/internal/domain/entities"
)

type OrganizationRepository interface {
	FindById(ctx context.Context, id string) (*entities.Organization, error)
	FindByIdentifier(ctx context.Context, identifier string) (*entities.Organization, error)
	Insert(ctx context.Context, organization *entities.Organization) error
	Update(ctx context.Context, organization *entities.Organization) error
}
