package repositories

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type ComponentPositionData struct {
	ComponentID string
	X           float64
	Y           float64
}

type ViewLayoutRepository struct {
	db *database.TenantAwareDB
}

func NewViewLayoutRepository(db *database.TenantAwareDB) *ViewLayoutRepository {
	return &ViewLayoutRepository{db: db}
}

func (r *ViewLayoutRepository) UpdateComponentPosition(ctx context.Context, viewID, componentID string, x, y float64) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO view_component_positions (view_id, tenant_id, component_id, x, y, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		ON CONFLICT (tenant_id, view_id, component_id)
		DO UPDATE SET x = $4, y = $5, updated_at = $6`,
		viewID, tenantID.Value(), componentID, x, y, time.Now().UTC(),
	)
	return err
}

func (r *ViewLayoutRepository) UpdateMultiplePositions(ctx context.Context, viewID string, positions []ComponentPositionData) error {
	if len(positions) == 0 {
		return nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTxWithTenant(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	for _, pos := range positions {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO view_component_positions (view_id, tenant_id, component_id, x, y, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $6)
			ON CONFLICT (tenant_id, view_id, component_id)
			DO UPDATE SET x = $4, y = $5, updated_at = $6`,
			viewID, tenantID.Value(), pos.ComponentID, pos.X, pos.Y, now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ViewLayoutRepository) GetLayout(ctx context.Context, viewID string) ([]ComponentPositionData, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var positions []ComponentPositionData

	err = r.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT component_id, x, y FROM view_component_positions WHERE tenant_id = $1 AND view_id = $2",
			tenantID.Value(), viewID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var pos ComponentPositionData
			if err := rows.Scan(&pos.ComponentID, &pos.X, &pos.Y); err != nil {
				return err
			}
			positions = append(positions, pos)
		}

		return rows.Err()
	})

	return positions, err
}

func (r *ViewLayoutRepository) DeleteComponentPosition(ctx context.Context, viewID, componentID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		"DELETE FROM view_component_positions WHERE tenant_id = $1 AND view_id = $2 AND component_id = $3",
		tenantID.Value(), viewID, componentID,
	)
	return err
}

func (r *ViewLayoutRepository) DeleteLayout(ctx context.Context, viewID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		"DELETE FROM view_component_positions WHERE tenant_id = $1 AND view_id = $2",
		tenantID.Value(), viewID,
	)
	return err
}

func (r *ViewLayoutRepository) UpdateEdgeType(ctx context.Context, viewID, edgeType string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO view_preferences (tenant_id, view_id, edge_type, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, view_id)
		DO UPDATE SET edge_type = $3, updated_at = $4`,
		tenantID.Value(), viewID, edgeType, time.Now().UTC(),
	)
	return err
}

func (r *ViewLayoutRepository) UpdateLayoutDirection(ctx context.Context, viewID, layoutDirection string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO view_preferences (tenant_id, view_id, layout_direction, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, view_id)
		DO UPDATE SET layout_direction = $3, updated_at = $4`,
		tenantID.Value(), viewID, layoutDirection, time.Now().UTC(),
	)
	return err
}

func (r *ViewLayoutRepository) GetPreferences(ctx context.Context, viewID string) (edgeType, layoutDirection string, err error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", "", err
	}

	var et, ld sql.NullString
	err = r.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			"SELECT edge_type, layout_direction FROM view_preferences WHERE tenant_id = $1 AND view_id = $2",
			tenantID.Value(), viewID,
		).Scan(&et, &ld)
	})

	if err == sql.ErrNoRows {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}

	if et.Valid {
		edgeType = et.String
	}
	if ld.Valid {
		layoutDirection = ld.String
	}

	return edgeType, layoutDirection, nil
}
