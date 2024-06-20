package postgresql

import (
	"context"
	"database/sql"
	"iyaem/internal/app/queries"
)

type OrganizationQuery struct {
	db *sql.DB
}

func NewOrganizationQuery(db *sql.DB) *OrganizationQuery {
	return &OrganizationQuery{db}
}

func (q *OrganizationQuery) GetAllAffilatedOrganizations(ctx context.Context, userId string) ([]queries.Organization, error) {
	rows, err := q.db.Query(`
		SELECT organization_id, name FROM user_organization 
		LEFT JOIN organization ON user_organization.organization_id = organization.id
		WHERE user_id = $1;`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orgs := make([]queries.Organization, 0)

	for rows.Next() {
		org := queries.Organization{}
		err := rows.Scan(&org.Id, &org.Name)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

func (q *OrganizationQuery) GetUsersInOrganization(ctx context.Context, organizationId string) ([]queries.User, error) {
	rows, err := q.db.Query(`
		SELECT uo.id, uo.user_id, u."picture", u."name", u."email", uo."level", uo.created_at as joined_at FROM user_organization uo 
		LEFT JOIN public."user" u ON u.id = uo.user_id 
		WHERE uo.organization_id=$1;`, organizationId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := make([]queries.User, 0)

	for rows.Next() {
		user := queries.User{}
		err := rows.Scan(&user.UserOrgId, &user.UserId, &user.Picture, &user.Name, &user.Email, &user.Level, &user.JoinedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
