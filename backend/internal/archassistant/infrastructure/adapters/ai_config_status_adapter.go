package adapters

import (
	"context"
	"database/sql"

	"easi/backend/internal/infrastructure/database"
)

type AIConfigStatusAdapter struct {
	db *database.TenantAwareDB
}

func NewAIConfigStatusAdapter(db *database.TenantAwareDB) *AIConfigStatusAdapter {
	return &AIConfigStatusAdapter{db: db}
}

func (a *AIConfigStatusAdapter) IsConfigured(ctx context.Context) (bool, error) {
	var configured bool
	err := a.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			"SELECT status = 'configured' FROM archassistant.ai_configurations LIMIT 1",
		)
		err := row.Scan(&configured)
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	})
	return configured, err
}
