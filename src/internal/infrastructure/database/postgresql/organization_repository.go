package postgresql

import (
	"context"
	"database/sql"
	"iyaem/internal/domain/entities"
	"iyaem/internal/domain/repositories"
)

type OrganizationRepository struct {
	db *sql.DB
}

func NewOrganizationRepository(db *sql.DB) repositories.OrganizationRepository {
	return &OrganizationRepository{
		db: db,
	}
}

func (r *OrganizationRepository) Insert(ctx context.Context, org *entities.Organization) error {

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO organization (id, name, identifier) VALUES ($1, $2, $3);`,
		org.Id().Value(), org.Name(), org.Identifier(),
	)

	if err != nil {
		return err
	}

	for _, member := range org.Members() {
		_, err = tx.Exec(
			`INSERT INTO user_organization (id, organization_id, user_id, level) VALUES ($1, $2, $3, $4);`,
			member.Id().Value(), org.Id().Value(), member.UserId().Value(), member.Level(),
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
