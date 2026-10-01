package readmodels

import (
	"context"
	"database/sql"
	"fmt"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type tenantStore struct {
	db *database.TenantAwareDB
}

func (s tenantStore) exec(ctx context.Context, action, query string, args ...any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, query, append([]any{tenantID.Value()}, args...)...); err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}
	return nil
}

func (s tenantStore) read(ctx context.Context, action string, fn func(q tenantQuery) error) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	if err := s.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return fn(tenantQuery{ctx: ctx, tx: tx, tenantID: tenantID.Value()})
	}); err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}
	return nil
}

type tenantQuery struct {
	ctx      context.Context
	tx       *sql.Tx
	tenantID string
}

func (q tenantQuery) each(query string, scan func(*sql.Rows) error, args ...any) error {
	rows, err := q.tx.QueryContext(q.ctx, query, append([]any{q.tenantID}, args...)...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}
