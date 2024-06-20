package repositories

import (
	"context"
	"iyaem/internal/domain/entities"
)

type OrganizationRepository interface {
	Insert(ctx context.Context, organization *entities.Organization) error
}
