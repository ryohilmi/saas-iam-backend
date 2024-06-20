package postgresql

import (
	"context"

	"iyaem/internal/domain/entities"
	"iyaem/internal/domain/repositories"
	vo "iyaem/internal/domain/valueobjects"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &UserRepository{
		db: db,
	}
}

// func (r *UserRepository) Insert(ctx context.Context, user *entities.User) error {

// }

func (r *UserRepository) FindById(ctx context.Context, userId vo.UserId) (*entities.User, error) {
	var userRecord struct {
		Id       string
		Username string
		Email    string
	}

	var membershipRecord struct {
		UserId         string
		OrganizationId string
	}

	var membership []entities.Membership

	result := r.db.Raw(`SELECT id, username, email FROM public.user WHERE id = ?`, userId.Value()).
		Take(&userRecord)

	if result.Error != nil {
		return nil, result.Error
	}

	rows, err := r.db.Raw("SELECT user_id, organization_id FROM public.user_organization WHERE user_id = ?", userId.Value()).
		Rows()
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		rows.Scan(&membershipRecord)
		orgId, _ := vo.NewOrganizationId(membershipRecord.OrganizationId)
		membership = append(membership, entities.NewMembership(userId, orgId, "owner", make([]entities.Role, 0)))
	}

	if result.Error != nil {
		return nil, result.Error
	}

	user := entities.NewUser(userId, userRecord.Username, userRecord.Email, membership)

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	return nil
}
