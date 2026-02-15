package adapters

import (
	"context"

	"easi/backend/internal/auth/application/readmodels"
)

type UserRoleCheckerAdapter struct {
	readModel *readmodels.UserReadModel
}

func NewUserRoleCheckerAdapter(rm *readmodels.UserReadModel) *UserRoleCheckerAdapter {
	return &UserRoleCheckerAdapter{readModel: rm}
}

func (a *UserRoleCheckerAdapter) IsAdmin(ctx context.Context, userID string) (bool, error) {
	user, err := a.readModel.GetByIDString(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}
	return user.Role == "admin", nil
}
