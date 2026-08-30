package readmodels

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type MaturityScaleSection struct {
	Name     string `json:"name"`
	MinValue int    `json:"minValue"`
	MaxValue int    `json:"maxValue"`
}

type MaturityScaleCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewMaturityScaleCacheReadModel(db *database.TenantAwareDB) *MaturityScaleCacheReadModel {
	return &MaturityScaleCacheReadModel{db: db}
}

func (rm *MaturityScaleCacheReadModel) Save(ctx context.Context, sections []MaturityScaleSection) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(sections)
	if err != nil {
		return fmt.Errorf("encode maturity scale sections: %w", err)
	}
	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO onepagers.maturity_scale_cache (tenant_id, sections) VALUES ($1, $2::jsonb)
		ON CONFLICT (tenant_id) DO UPDATE SET sections = EXCLUDED.sections`,
		tenantID.Value(), encoded,
	)
	if err != nil {
		return fmt.Errorf("cache maturity scale sections: %w", err)
	}
	return nil
}

func (rm *MaturityScaleCacheReadModel) Sections(ctx context.Context) ([]MaturityScaleSection, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	var raw []byte
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		scanErr := tx.QueryRowContext(ctx,
			`SELECT sections FROM onepagers.maturity_scale_cache WHERE tenant_id = $1`,
			tenantID.Value(),
		).Scan(&raw)
		if scanErr == sql.ErrNoRows {
			return nil
		}
		return scanErr
	})
	if err != nil {
		return nil, fmt.Errorf("read cached maturity scale sections: %w", err)
	}
	if len(raw) == 0 {
		return nil, nil
	}
	var sections []MaturityScaleSection
	if err := json.Unmarshal(raw, &sections); err != nil {
		return nil, fmt.Errorf("decode cached maturity scale sections: %w", err)
	}
	return sections, nil
}
