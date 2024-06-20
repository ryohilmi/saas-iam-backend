package repositories

import (
	"context"
	"iyaem/internal/domain/entities"
	vo "iyaem/internal/domain/valueobjects"
)

type UserRepository interface {
	// Insert(ctx context.Context, user *entities.User) error
	FindById(ctx context.Context, userId vo.UserId) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	// Update(ctx context.Context, user *entities.User) error
}
