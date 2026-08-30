package readmodels

import (
	"context"
	"database/sql"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type EnterpriseCapabilityCacheDTO struct {
	ID             string
	Name           string
	Category       string
	Active         bool
	TargetMaturity *int
}

type EnterpriseCapabilityCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewEnterpriseCapabilityCacheReadModel(db *database.TenantAwareDB) *EnterpriseCapabilityCacheReadModel {
	return &EnterpriseCapabilityCacheReadModel{db: db}
}

func (rm *EnterpriseCapabilityCacheReadModel) Insert(ctx context.Context, dto EnterpriseCapabilityCacheDTO) error {
	return rm.execForTenant(ctx,
		`INSERT INTO architecturedirection.enterprise_capability_cache (tenant_id, id, name, category, active)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (tenant_id, id) DO UPDATE SET
		 name = EXCLUDED.name, category = EXCLUDED.category, active = EXCLUDED.active`,
		dto.ID, dto.Name, nullIfEmpty(dto.Category), dto.Active,
	)
}

func (rm *EnterpriseCapabilityCacheReadModel) UpdateDetails(ctx context.Context, dto EnterpriseCapabilityCacheDTO) error {
	return rm.execForTenant(ctx,
		`UPDATE architecturedirection.enterprise_capability_cache SET name = $2, category = $3 WHERE tenant_id = $1 AND id = $4`,
		dto.Name, nullIfEmpty(dto.Category), dto.ID,
	)
}

func (rm *EnterpriseCapabilityCacheReadModel) Delete(ctx context.Context, id string) error {
	return rm.execForTenant(ctx,
		`DELETE FROM architecturedirection.enterprise_capability_cache WHERE tenant_id = $1 AND id = $2`,
		id,
	)
}

func (rm *EnterpriseCapabilityCacheReadModel) UpdateTargetMaturity(ctx context.Context, id string, targetMaturity int) error {
	return rm.execForTenant(ctx,
		`UPDATE architecturedirection.enterprise_capability_cache SET target_maturity = $2 WHERE tenant_id = $1 AND id = $3`,
		targetMaturity, id,
	)
}

func (rm *EnterpriseCapabilityCacheReadModel) GetByID(ctx context.Context, id string) (*EnterpriseCapabilityCacheDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	var dto EnterpriseCapabilityCacheDTO
	var category sql.NullString
	var targetMaturity sql.NullInt64
	var notFound bool
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT id, name, category, active, target_maturity FROM architecturedirection.enterprise_capability_cache
			 WHERE tenant_id = $1 AND id = $2`,
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.Name, &category, &dto.Active, &targetMaturity)
		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
		return err
	})
	if err != nil || notFound {
		return nil, err
	}
	dto.Category = category.String
	if targetMaturity.Valid {
		value := int(targetMaturity.Int64)
		dto.TargetMaturity = &value
	}
	return &dto, nil
}

func (rm *EnterpriseCapabilityCacheReadModel) ActiveEnterpriseCapabilityNames(ctx context.Context) (map[string]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	names := map[string]string{}
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT id, name FROM architecturedirection.enterprise_capability_cache WHERE tenant_id = $1 AND active = true`,
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var id, name string
			if err := rows.Scan(&id, &name); err != nil {
				return err
			}
			names[id] = name
		}
		return rows.Err()
	})
	return names, err
}

func (rm *EnterpriseCapabilityCacheReadModel) execForTenant(ctx context.Context, query string, args ...any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, append([]any{tenantID.Value()}, args...)...)
	return err
}
