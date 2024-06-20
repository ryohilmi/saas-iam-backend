package repositories

import (
	"context"
	"iyaem/internal/domain/entities"
)

type MembershipRepository interface {
	// Insert(ctx context.Context, user *entities.User) error
	FindById(ctx context.Context, email string) (*entities.Membership, error)
	FindByEmail(ctx context.Context, email string) (*entities.Membership, error)
	// Update(ctx context.Context, user *entities.User) error
}
