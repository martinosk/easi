package readmodels

import (
	"context"
	"database/sql"

	sharedctx "easi/backend/internal/shared/context"
)

func queryList[T any](rm *ValueStreamReadModel, ctx context.Context, query string, arg interface{}, scan func(scanner) (T, error)) ([]T, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var results []T
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, tenantID.Value(), arg)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			item, err := scan(rows)
			if err != nil {
				return err
			}
			results = append(results, item)
		}
		return rows.Err()
	})
	return results, err
}

func (rm *ValueStreamReadModel) idempotentInsert(ctx context.Context, deleteQuery, insertQuery string, deleteArgs, insertArgs func(string) []interface{}) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	tid := tenantID.Value()
	if _, err = rm.db.ExecContext(ctx, deleteQuery, deleteArgs(tid)...); err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, insertQuery, insertArgs(tid)...)
	return err
}

func (rm *ValueStreamReadModel) execTenantQuery(ctx context.Context, query string, buildArgs func(tenantID string) []interface{}) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, buildArgs(tenantID.Value())...)
	return err
}

func (rm *ValueStreamReadModel) nameExistsForTenant(ctx context.Context, baseQuery, excludeQuery, excludeID string, extraArgs ...interface{}) (bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return false, err
	}
	args := append([]interface{}{tenantID.Value()}, extraArgs...)
	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		if excludeID != "" {
			return tx.QueryRowContext(ctx, excludeQuery, append(args, excludeID)...).Scan(&count)
		}
		return tx.QueryRowContext(ctx, baseQuery, args...).Scan(&count)
	})
	return count > 0, err
}
