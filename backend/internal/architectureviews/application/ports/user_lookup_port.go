package ports

import "context"

type UserRoleChecker interface {
	IsAdmin(ctx context.Context, userID string) (bool, error)
}
