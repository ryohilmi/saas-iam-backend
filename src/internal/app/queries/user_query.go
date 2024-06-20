package queries

import "context"

type Level string

type UserQuery interface {
	UserLevel(ctx context.Context, userId string, organizationId string) (Level, error)
}
