package postgresql

import (
	"context"
	"database/sql"
	"iyaem/internal/app/queries"
)

type UserQuery struct {
	db *sql.DB
}

func NewUserQuery(db *sql.DB) *UserQuery {
	return &UserQuery{db}
}

func (q *UserQuery) UserLevel(ctx context.Context, userId string, organizationId string) (queries.Level, error) {
	var level queries.Level

	row := q.db.QueryRow(`
		SELECT level FROM user_organization
		LEFT JOIN public.user u ON user_organization.user_id = u.id 
		WHERE email=$1 AND organization_id=$2;`, userId, organizationId)

	err := row.Scan(&level)
	if err != nil {
		return "", err
	}

	return level, nil
}
