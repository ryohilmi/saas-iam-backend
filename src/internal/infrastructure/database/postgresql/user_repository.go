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

func (r *UserRepository) Insert(ctx context.Context, user *entities.User) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO public.user (id, email, picture, name) VALUES ($1, $2, $3, $4);`,
		user.Id().Value(), user.Email(), user.Picture(), user.Name(),
	)

	if err != nil {
		return err
	}

	for _, identity := range user.Identities() {
		_, err = tx.Exec(
			`INSERT INTO user_identity (idp_id, user_id) VALUES ($1, $2);`,
			identity.IdpId(), user.Id().Value(),
		)

		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) FindById(ctx context.Context, userId vo.UserId) (*entities.User, error) {
	return nil, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {

	var user entities.User
	var userRecord struct {
		Id      string
		Name    string
		Email   string
		Picture string
	}

	row := r.db.QueryRow(`	
		SELECT id, name, email, picture
		FROM public.user
		WHERE email=$1`, email,
	)
	err := row.Scan(&userRecord.Id, &userRecord.Name, &userRecord.Email, &userRecord.Picture)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	userId, _ := vo.NewUserId(userRecord.Id)

	var identityRecord struct {
		IdpId  string
		UserId string
	}

	rows, err := r.db.Query(`
		SELECT idp_id, user_id
		FROM user_identity
		WHERE user_id=$1`, userRecord.Id,
	)

	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	identities := make([]vo.Identity, 0)
	for rows.Next() {
		err = rows.Scan(&identityRecord.IdpId, &identityRecord.UserId)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		identity := vo.NewIdentity(identityRecord.IdpId, userId)
		identities = append(identities, identity)
	}

	user = entities.NewUser(
		userId,
		userRecord.Name,
		userRecord.Email,
		userRecord.Picture,
		identities,
		make([]entities.Membership, 0),
	)

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	return nil
}
