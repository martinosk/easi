package readmodels

import (
	"context"
	"database/sql"
	"fmt"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type tenantScopedDB struct {
	db *database.TenantAwareDB
}

func (s tenantScopedDB) exec(ctx context.Context, action, query string, args ...any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, query, append([]any{tenantID.Value()}, args...)...); err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}
	return nil
}

func (s tenantScopedDB) read(ctx context.Context, action string, fn func(tx *sql.Tx, tenantID string) error) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	if err := s.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error { return fn(tx, tenantID.Value()) }); err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}
	return nil
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
