package repositories

import (
	"context"
	"iyaem/internal/domain/entities"
)

type OrganizationRepository interface {
	FindByIdentifier(ctx context.Context, identifier string) (*entities.Organization, error)
	Insert(ctx context.Context, organization *entities.Organization) error
}
