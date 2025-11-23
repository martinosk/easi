package repositories

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type Position struct {
	X float64
	Y float64
}

type ComponentPositionData struct {
	ComponentID string
	X           float64
	Y           float64
}

type ElementType string

const (
	ElementTypeComponent  ElementType = "component"
	ElementTypeCapability ElementType = "capability"
)

type ViewLayoutRepository struct {
	db *database.TenantAwareDB
}

func NewViewLayoutRepository(db *database.TenantAwareDB) *ViewLayoutRepository {
	return &ViewLayoutRepository{db: db}
}

func (r *ViewLayoutRepository) UpdateComponentPosition(ctx context.Context, viewID, componentID string, x, y float64) error {
	return r.upsertElementPosition(ctx, viewID, componentID, ElementTypeComponent, Position{X: x, Y: y})
}

func (r *ViewLayoutRepository) upsertElementPosition(ctx context.Context, viewID, elementID string, elementType ElementType, pos Position) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO view_element_positions (view_id, tenant_id, element_id, element_type, x, y, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT (tenant_id, view_id, element_id, element_type)
		DO UPDATE SET x = $5, y = $6, updated_at = $7`,
		viewID, tenantID.Value(), elementID, string(elementType), pos.X, pos.Y, time.Now().UTC(),
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
			`INSERT INTO view_element_positions (view_id, tenant_id, element_id, element_type, x, y, created_at, updated_at)
			VALUES ($1, $2, $3, 'component', $4, $5, $6, $6)
			ON CONFLICT (tenant_id, view_id, element_id, element_type)
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
			"SELECT element_id, x, y FROM view_element_positions WHERE tenant_id = $1 AND view_id = $2 AND element_type = 'component'",
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
	return r.deleteElementPosition(ctx, viewID, componentID, ElementTypeComponent)
}

func (r *ViewLayoutRepository) DeleteLayout(ctx context.Context, viewID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		"DELETE FROM view_element_positions WHERE tenant_id = $1 AND view_id = $2",
		tenantID.Value(), viewID,
	)
	return err
}

func (r *ViewLayoutRepository) deleteElementPosition(ctx context.Context, viewID, elementID string, elementType ElementType) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		"DELETE FROM view_element_positions WHERE tenant_id = $1 AND view_id = $2 AND element_id = $3 AND element_type = $4",
		tenantID.Value(), viewID, elementID, string(elementType),
	)
	return err
}

func (r *ViewLayoutRepository) UpdateEdgeType(ctx context.Context, viewID, edgeType string) error {
	return r.updateViewPreference(ctx, viewID, "edge_type", edgeType)
}

func (r *ViewLayoutRepository) UpdateLayoutDirection(ctx context.Context, viewID, layoutDirection string) error {
	return r.updateViewPreference(ctx, viewID, "layout_direction", layoutDirection)
}

func (r *ViewLayoutRepository) updateViewPreference(ctx context.Context, viewID, column, value string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	query := `INSERT INTO view_preferences (tenant_id, view_id, ` + column + `, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, view_id)
		DO UPDATE SET ` + column + ` = $3, updated_at = $4`
	_, err = r.db.ExecContext(ctx, query, tenantID.Value(), viewID, value, time.Now().UTC())
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

type CapabilityPositionData struct {
	CapabilityID string
	X            float64
	Y            float64
}

func (r *ViewLayoutRepository) AddCapabilityToView(ctx context.Context, viewID, capabilityID string, x, y float64) error {
	return r.upsertElementPosition(ctx, viewID, capabilityID, ElementTypeCapability, Position{X: x, Y: y})
}

func (r *ViewLayoutRepository) UpdateCapabilityPosition(ctx context.Context, viewID, capabilityID string, x, y float64) error {
	return r.upsertElementPosition(ctx, viewID, capabilityID, ElementTypeCapability, Position{X: x, Y: y})
}

func (r *ViewLayoutRepository) RemoveCapabilityFromView(ctx context.Context, viewID, capabilityID string) error {
	return r.deleteElementPosition(ctx, viewID, capabilityID, ElementTypeCapability)
}
