package readmodels

import (
	"context"
	"database/sql"
	"fmt"

	sharedctx "easi/backend/internal/shared/context"

	"github.com/lib/pq"
)

func (rm *OnePagerFactsReadModel) FilledFieldCounts(ctx context.Context, subjectType string, subjectIDs, fieldIDs []string) (map[string]int, error) {
	counts := make(map[string]int)
	if len(subjectIDs) == 0 || len(fieldIDs) == 0 {
		return counts, nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT subject_id, COUNT(*) FROM onepagers.one_pager_facts
			WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = ANY($3) AND field_id = ANY($4) AND value IS NOT NULL
			GROUP BY subject_id`,
			tenantID.Value(), subjectType, pq.Array(subjectIDs), pq.Array(fieldIDs),
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		return scanFilledFieldCounts(rows, counts)
	})
	if err != nil {
		return nil, fmt.Errorf("query filled field counts for subject type %s: %w", subjectType, err)
	}
	return counts, nil
}

func scanFilledFieldCounts(rows *sql.Rows, counts map[string]int) error {
	for rows.Next() {
		var subjectID string
		var count int
		if err := rows.Scan(&subjectID, &count); err != nil {
			return err
		}
		counts[subjectID] = count
	}
	return rows.Err()
}

func (rm *OnePagerFactsReadModel) CountSubjectsWithValue(ctx context.Context, subjectType, fieldID string) (int, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return 0, err
	}

	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM onepagers.one_pager_facts
			WHERE tenant_id = $1 AND subject_type = $2 AND field_id = $3 AND value IS NOT NULL`,
			tenantID.Value(), subjectType, fieldID,
		).Scan(&count)
	})
	if err != nil {
		return 0, fmt.Errorf("count subjects with value for field %s: %w", fieldID, err)
	}
	return count, nil
}
