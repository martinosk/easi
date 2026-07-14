package api

import (
	"context"
	"database/sql"
	"fmt"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/ports"
	sharedctx "easi/backend/internal/shared/context"
)

type onePagerAuditAdapter struct {
	db *database.TenantAwareDB
}

func newOnePagerAuditAdapter(db *database.TenantAwareDB) ports.SubjectAuditReader {
	return onePagerAuditAdapter{db: db}
}

func (a onePagerAuditAdapter) Created(ctx context.Context, aggregateID string) (ports.SubjectAudit, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return ports.SubjectAudit{}, err
	}

	var audit ports.SubjectAudit
	err = a.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			`SELECT actor_id, actor_email, occurred_at
			FROM infrastructure.events
			WHERE tenant_id = $1 AND aggregate_id = $2 AND version = 1
			LIMIT 1`,
			tenantID.Value(), aggregateID,
		)
		scanErr := row.Scan(&audit.ActorID, &audit.ActorEmail, &audit.CreatedAt)
		if scanErr == sql.ErrNoRows {
			return nil
		}
		if scanErr == nil {
			audit.Found = true
		}
		return scanErr
	})
	if err != nil {
		return ports.SubjectAudit{}, fmt.Errorf("read creation audit for aggregate %s: %w", aggregateID, err)
	}
	return audit, nil
}
