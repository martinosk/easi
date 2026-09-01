package readmodels

import (
	"context"
	"database/sql"
	"fmt"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type UserNameCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewUserNameCacheReadModel(db *database.TenantAwareDB) *UserNameCacheReadModel {
	return &UserNameCacheReadModel{db: db}
}

func (rm *UserNameCacheReadModel) Upsert(ctx context.Context, id, name, email string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx, `
		INSERT INTO architecturemodeling.user_names (tenant_id, user_id, name, email)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, user_id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email
	`, tenantID.Value(), id, name, email)
	if err != nil {
		return fmt.Errorf("upsert user name cache entry for user %s: %w", id, err)
	}
	return nil
}

func (rm *UserNameCacheReadModel) Exists(ctx context.Context, id string) (bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return false, err
	}

	var exists bool
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM architecturemodeling.user_names WHERE tenant_id = $1 AND user_id = $2)",
			tenantID.Value(), id,
		).Scan(&exists)
	})
	if err != nil {
		return false, fmt.Errorf("check user %s in name cache for tenant %s: %w", id, tenantID.Value(), err)
	}
	return exists, nil
}
