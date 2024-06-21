package queries

import "context"

type Organization struct {
	Id   string `json:"organization_id"`
	Name string `json:"name"`
}

type User struct {
	UserOrgId string `json:"user_org_id"`
	UserId    string `json:"user_id"`
	Picture   string `json:"picture"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Level     string `json:"level"`
	JoinedAt  string `json:"joined_at"`
}

type OrganizationQuery interface {
	AllAffilatedOrganizations(ctx context.Context, userId string) ([]Organization, error)
	UsersInOrganization(ctx context.Context, organizationId string) ([]User, error)
	RecentUsersInOrganization(ctx context.Context, organizationId string) ([]User, error)
	FindById(ctx context.Context, organizationId string) (Organization, error)
}
