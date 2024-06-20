package postgresql

import (
	"context"
	"database/sql"
	"log"

	"iyaem/internal/domain/entities"
	"iyaem/internal/domain/repositories"
	vo "iyaem/internal/domain/valueobjects"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repositories.UserRepository {
	return &UserRepository{
		db: db,
	}
}

// func (r *UserRepository) Insert(ctx context.Context, user *entities.User) error {

// }

func (r *UserRepository) FindById(ctx context.Context, userId vo.UserId) (*entities.User, error) {
	return nil, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {

	var user entities.User
	var userRecord struct {
		Id    string
		Name  string
		Email string
	}

	row := r.db.QueryRow(`	
		SELECT id, name, email
		FROM public.user
		WHERE email=$1`, email,
	)
	err := row.Scan(&userRecord.Id, &userRecord.Name, &userRecord.Email)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	userId, _ := vo.NewUserId(userRecord.Id)

	user = entities.NewUser(
		userId,
		userRecord.Name,
		userRecord.Email,
		make([]entities.Membership, 0),
	)

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	return nil
}
