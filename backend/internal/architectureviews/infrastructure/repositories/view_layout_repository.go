package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

var ErrPlacementNotFound = errors.New("element is not placed in view")

type Position struct {
	X float64
	Y float64
}

type ComponentPositionData struct {
	ComponentID string
	X           float64
	Y           float64
}

type ElementRef struct {
	ViewID      string
	ElementID   string
	ElementType ElementType
}

type ElementType string

const (
	ElementTypeComponent    ElementType = "component"
	ElementTypeCapability   ElementType = "capability"
	ElementTypeOriginEntity ElementType = "origin_entity"
)

type ViewLayoutRepository struct {
	db *database.TenantAwareDB
}

func NewViewLayoutRepository(db *database.TenantAwareDB) *ViewLayoutRepository {
	return &ViewLayoutRepository{db: db}
}

func (r *ViewLayoutRepository) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.RunInTx(ctx, fn)
}

func (r *ViewLayoutRepository) PlaceElement(ctx context.Context, ref ElementRef, pos Position) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO architectureviews.view_element_positions (view_id, tenant_id, element_id, element_type, x, y, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT (tenant_id, view_id, element_id, element_type)
		DO UPDATE SET x = $5, y = $6, updated_at = $7`,
		ref.ViewID, tenantID.Value(), ref.ElementID, string(ref.ElementType), pos.X, pos.Y, time.Now().UTC(),
	)
	return err
}

func (r *ViewLayoutRepository) MoveElement(ctx context.Context, ref ElementRef, pos Position) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, moveElementSQL,
		pos.X, pos.Y, time.Now().UTC(), tenantID.Value(), ref.ViewID, ref.ElementID, string(ref.ElementType),
	)
	if err != nil {
		return err
	}
	moved, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if moved == 0 {
		return ErrPlacementNotFound
	}
	return nil
}

const moveElementSQL = `UPDATE architectureviews.view_element_positions
	SET x = $1, y = $2, updated_at = $3
	WHERE tenant_id = $4 AND view_id = $5 AND element_id = $6 AND element_type = $7`

func (r *ViewLayoutRepository) UpdateMultiplePositions(ctx context.Context, viewID string, positions []ComponentPositionData) error {
	if len(positions) == 0 {
		return nil
	}

	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	return r.db.RunInTx(ctx, func(ctx context.Context) error {
		for _, pos := range positions {
			if _, err := r.db.ExecContext(ctx, moveElementSQL,
				pos.X, pos.Y, now, tenantID.Value(), viewID, pos.ComponentID, string(ElementTypeComponent),
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ViewLayoutRepository) DeleteElementFromAllViews(ctx context.Context, elementID string, elementType ElementType) error {
	return r.execForTenant(ctx,
		"DELETE FROM architectureviews.view_element_positions WHERE tenant_id = $1 AND element_id = $2 AND element_type = $3",
		elementID, string(elementType),
	)
}

func (r *ViewLayoutRepository) execForTenant(ctx context.Context, query string, args ...interface{}) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, append([]interface{}{tenantID.Value()}, args...)...)
	return err
}

func (r *ViewLayoutRepository) GetLayout(ctx context.Context, viewID string) ([]ComponentPositionData, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var positions []ComponentPositionData

	err = r.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT element_id, x, y FROM architectureviews.view_element_positions WHERE tenant_id = $1 AND view_id = $2 AND element_type = 'component'",
			tenantID.Value(), viewID,
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

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

func (r *ViewLayoutRepository) DeleteElementPosition(ctx context.Context, ref ElementRef) error {
	return r.execForTenant(ctx,
		"DELETE FROM architectureviews.view_element_positions WHERE tenant_id = $1 AND view_id = $2 AND element_id = $3 AND element_type = $4",
		ref.ViewID, ref.ElementID, string(ref.ElementType),
	)
}

type PreferenceKey string

const (
	PreferenceKeyEdgeType        PreferenceKey = "edge_type"
	PreferenceKeyLayoutDirection PreferenceKey = "layout_direction"
	PreferenceKeyColorScheme     PreferenceKey = "color_scheme"
)

func (r *ViewLayoutRepository) UpdatePreference(ctx context.Context, viewID string, key PreferenceKey, value string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	column := string(key)
	query := `INSERT INTO architectureviews.view_preferences (tenant_id, view_id, ` + column + `, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, view_id)
		DO UPDATE SET ` + column + ` = $3, updated_at = $4`
	_, err = r.db.ExecContext(ctx, query, tenantID.Value(), viewID, value, time.Now().UTC())
	return err
}

func (r *ViewLayoutRepository) UpdateElementColor(ctx context.Context, ref ElementRef, color string) error {
	return r.setElementColor(ctx, ref, &color)
}

func (r *ViewLayoutRepository) ClearElementColor(ctx context.Context, ref ElementRef) error {
	return r.setElementColor(ctx, ref, nil)
}

func (r *ViewLayoutRepository) setElementColor(ctx context.Context, ref ElementRef, color *string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	query := `UPDATE architectureviews.view_element_positions
		SET custom_color = $1, updated_at = $2
		WHERE tenant_id = $3 AND view_id = $4 AND element_id = $5 AND element_type = $6`
	_, err = r.db.ExecContext(ctx, query, color, time.Now().UTC(), tenantID.Value(), ref.ViewID, ref.ElementID, string(ref.ElementType))
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
			"SELECT edge_type, layout_direction FROM architectureviews.view_preferences WHERE tenant_id = $1 AND view_id = $2",
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
