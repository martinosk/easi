package readmodels

import (
	"context"
	"database/sql"
	"errors"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/stewardship/domain/valueobjects"
)

type CachedUser struct {
	ID     string
	Name   string
	Email  string
	Active bool
}

type UserCacheReadModel struct {
	store tenantScopedDB
}

func NewUserCacheReadModel(db *database.TenantAwareDB) *UserCacheReadModel {
	return &UserCacheReadModel{store: tenantScopedDB{db: db}}
}

func (rm *UserCacheReadModel) Upsert(ctx context.Context, user CachedUser) error {
	return rm.store.exec(ctx, "cache user "+user.ID, `
		INSERT INTO stewardship.user_cache (tenant_id, user_id, name, email, active)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id, user_id) DO UPDATE SET
			name = EXCLUDED.name, email = EXCLUDED.email, active = EXCLUDED.active`,
		user.ID, user.Name, user.Email, user.Active)
}

func (rm *UserCacheReadModel) SetActive(ctx context.Context, userID string, active bool) error {
	return rm.store.exec(ctx, "set cached status of user "+userID, `
		UPDATE stewardship.user_cache SET active = $3 WHERE tenant_id = $1 AND user_id = $2`,
		userID, active)
}

func (rm *UserCacheReadModel) UserStanding(ctx context.Context, userID string) (valueobjects.UserStanding, error) {
	var active bool
	err := rm.store.read(ctx, "read cached standing of user "+userID, func(tx *sql.Tx, tenantID string) error {
		return tx.QueryRowContext(ctx,
			`SELECT active FROM stewardship.user_cache WHERE tenant_id = $1 AND user_id = $2`,
			tenantID, userID,
		).Scan(&active)
	})
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return valueobjects.UserUnknown, nil
	case err != nil:
		return valueobjects.UserUnknown, err
	case active:
		return valueobjects.UserActive, nil
	default:
		return valueobjects.UserDisabled, nil
	}
}
