package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/viewlayouts/domain/aggregates"
	"easi/backend/internal/viewlayouts/domain/valueobjects"
)

var (
	ErrContainerNotFound = errors.New("layout container not found")
	ErrVersionConflict   = errors.New("version conflict: container was modified")
)

type LayoutContainerRepository struct {
	db *database.TenantAwareDB
}

func NewLayoutContainerRepository(db *database.TenantAwareDB) *LayoutContainerRepository {
	return &LayoutContainerRepository{db: db}
}

type containerRow struct {
	id          string
	contextType string
	contextRef  string
	prefsJSON   string
	version     int
	createdAt   time.Time
	updatedAt   time.Time
}

func (r *LayoutContainerRepository) reconstituteContainer(
	ctx context.Context,
	tx *sql.Tx,
	row containerRow,
	containerID valueobjects.LayoutContainerID,
) (*aggregates.LayoutContainer, error) {
	contextType, err := valueobjects.NewLayoutContextType(row.contextType)
	if err != nil {
		return nil, err
	}

	contextRef, err := valueobjects.NewContextRef(row.contextRef)
	if err != nil {
		return nil, err
	}

	var prefsData map[string]interface{}
	if err := json.Unmarshal([]byte(row.prefsJSON), &prefsData); err != nil {
		prefsData = make(map[string]interface{})
	}

	container := aggregates.NewLayoutContainerWithState(
		containerID,
		contextType,
		contextRef,
		valueobjects.NewLayoutPreferences(prefsData),
		row.version,
		row.createdAt,
		row.updatedAt,
	)

	elements, err := r.loadElements(ctx, tx, containerID)
	if err != nil {
		return nil, err
	}
	container.SetElements(elements)

	return container, nil
}

func (r *LayoutContainerRepository) GetByContext(
	ctx context.Context,
	contextType valueobjects.LayoutContextType,
	contextRef valueobjects.ContextRef,
) (*aggregates.LayoutContainer, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var container *aggregates.LayoutContainer

	err = r.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		var row containerRow
		err := tx.QueryRowContext(ctx,
			`SELECT id, context_type, context_ref, preferences, version, created_at, updated_at
			FROM layout_containers
			WHERE tenant_id = $1 AND context_type = $2 AND context_ref = $3`,
			tenantID.Value(), contextType.Value(), contextRef.Value(),
		).Scan(&row.id, &row.contextType, &row.contextRef, &row.prefsJSON, &row.version, &row.createdAt, &row.updatedAt)

		if err == sql.ErrNoRows {
			return ErrContainerNotFound
		}
		if err != nil {
			return err
		}

		containerID, err := valueobjects.NewLayoutContainerIDFromString(row.id)
		if err != nil {
			return err
		}

		container, err = r.reconstituteContainer(ctx, tx, row, containerID)
		return err
	})

	if err != nil {
		return nil, err
	}

	return container, nil
}

func (r *LayoutContainerRepository) GetByID(
	ctx context.Context,
	id valueobjects.LayoutContainerID,
) (*aggregates.LayoutContainer, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var container *aggregates.LayoutContainer

	err = r.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := containerRow{id: id.Value()}
		err := tx.QueryRowContext(ctx,
			`SELECT context_type, context_ref, preferences, version, created_at, updated_at
			FROM layout_containers
			WHERE tenant_id = $1 AND id = $2`,
			tenantID.Value(), id.Value(),
		).Scan(&row.contextType, &row.contextRef, &row.prefsJSON, &row.version, &row.createdAt, &row.updatedAt)

		if err == sql.ErrNoRows {
			return ErrContainerNotFound
		}
		if err != nil {
			return err
		}

		container, err = r.reconstituteContainer(ctx, tx, row, id)
		return err
	})

	if err != nil {
		return nil, err
	}

	return container, nil
}

type elementRow struct {
	elementID   string
	x, y        float64
	width       sql.NullFloat64
	height      sql.NullFloat64
	customColor sql.NullString
	sortOrder   sql.NullInt32
}

func (row elementRow) toElementPosition() (valueobjects.ElementPosition, error) {
	elementID, err := valueobjects.NewElementID(row.elementID)
	if err != nil {
		return valueobjects.ElementPosition{}, err
	}

	var widthPtr, heightPtr *float64
	var colorPtr *valueobjects.HexColor
	var sortOrderPtr *int

	if row.width.Valid {
		widthPtr = &row.width.Float64
	}
	if row.height.Valid {
		heightPtr = &row.height.Float64
	}
	if row.customColor.Valid {
		if color, err := valueobjects.NewHexColor(row.customColor.String); err == nil {
			colorPtr = &color
		}
	}
	if row.sortOrder.Valid {
		so := int(row.sortOrder.Int32)
		sortOrderPtr = &so
	}

	pos, _ := valueobjects.NewElementPositionWithOptions(elementID, row.x, row.y, widthPtr, heightPtr, colorPtr, sortOrderPtr)
	return pos, nil
}

func (r *LayoutContainerRepository) loadElements(ctx context.Context, tx *sql.Tx, containerID valueobjects.LayoutContainerID) ([]valueobjects.ElementPosition, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx,
		`SELECT element_id, x, y, width, height, custom_color, sort_order
		FROM element_positions
		WHERE tenant_id = $1 AND container_id = $2`,
		tenantID.Value(), containerID.Value(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var elements []valueobjects.ElementPosition
	for rows.Next() {
		var row elementRow
		if err := rows.Scan(&row.elementID, &row.x, &row.y, &row.width, &row.height, &row.customColor, &row.sortOrder); err != nil {
			return nil, err
		}

		pos, err := row.toElementPosition()
		if err != nil {
			return nil, err
		}
		elements = append(elements, pos)
	}

	return elements, rows.Err()
}

func (r *LayoutContainerRepository) Save(ctx context.Context, container *aggregates.LayoutContainer) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	prefsJSON, err := json.Marshal(container.Preferences().ToMap())
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO layout_containers (id, tenant_id, context_type, context_ref, preferences, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tenant_id, context_type, context_ref)
		DO UPDATE SET preferences = $5, version = layout_containers.version + 1, updated_at = $8`,
		container.ID().Value(),
		tenantID.Value(),
		container.ContextType().Value(),
		container.ContextRef().Value(),
		string(prefsJSON),
		container.Version(),
		container.CreatedAt(),
		now,
	)

	return err
}

func (r *LayoutContainerRepository) Delete(ctx context.Context, id valueobjects.LayoutContainerID) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTxWithTenant(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		"DELETE FROM element_positions WHERE tenant_id = $1 AND container_id = $2",
		tenantID.Value(), id.Value(),
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		"DELETE FROM layout_containers WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), id.Value(),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

type positionSQLParams struct {
	width       sql.NullFloat64
	height      sql.NullFloat64
	customColor sql.NullString
	sortOrder   sql.NullInt32
}

func buildPositionParams(position valueobjects.ElementPosition) positionSQLParams {
	var params positionSQLParams
	if position.Width() != nil {
		params.width = sql.NullFloat64{Float64: *position.Width(), Valid: true}
	}
	if position.Height() != nil {
		params.height = sql.NullFloat64{Float64: *position.Height(), Valid: true}
	}
	if position.CustomColor() != nil {
		params.customColor = sql.NullString{String: position.CustomColor().Value(), Valid: true}
	}
	if position.SortOrder() != nil {
		params.sortOrder = sql.NullInt32{Int32: int32(*position.SortOrder()), Valid: true}
	}
	return params
}

func (r *LayoutContainerRepository) UpsertElementPosition(
	ctx context.Context,
	containerID valueobjects.LayoutContainerID,
	position valueobjects.ElementPosition,
) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	params := buildPositionParams(position)
	now := time.Now().UTC()

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO element_positions (container_id, tenant_id, element_id, x, y, width, height, custom_color, sort_order, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tenant_id, container_id, element_id)
		DO UPDATE SET x = $4, y = $5, width = $6, height = $7, custom_color = $8, sort_order = $9, updated_at = $10`,
		containerID.Value(),
		tenantID.Value(),
		position.ElementID().Value(),
		position.X(),
		position.Y(),
		params.width,
		params.height,
		params.customColor,
		params.sortOrder,
		now,
	)

	return err
}

func (r *LayoutContainerRepository) DeleteElementPosition(
	ctx context.Context,
	containerID valueobjects.LayoutContainerID,
	elementID valueobjects.ElementID,
) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		"DELETE FROM element_positions WHERE tenant_id = $1 AND container_id = $2 AND element_id = $3",
		tenantID.Value(), containerID.Value(), elementID.Value(),
	)

	return err
}

func (r *LayoutContainerRepository) BatchUpdatePositions(
	ctx context.Context,
	containerID valueobjects.LayoutContainerID,
	positions []valueobjects.ElementPosition,
) error {
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
	for _, position := range positions {
		params := buildPositionParams(position)
		_, err := tx.ExecContext(ctx,
			`INSERT INTO element_positions (container_id, tenant_id, element_id, x, y, width, height, custom_color, sort_order, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (tenant_id, container_id, element_id)
			DO UPDATE SET x = $4, y = $5, width = $6, height = $7, custom_color = $8, sort_order = $9, updated_at = $10`,
			containerID.Value(),
			tenantID.Value(),
			position.ElementID().Value(),
			position.X(),
			position.Y(),
			params.width,
			params.height,
			params.customColor,
			params.sortOrder,
			now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *LayoutContainerRepository) DeleteByContextRef(
	ctx context.Context,
	contextType valueobjects.LayoutContextType,
	contextRef valueobjects.ContextRef,
) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTxWithTenant(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var containerID string
	err = tx.QueryRowContext(ctx,
		"SELECT id FROM layout_containers WHERE tenant_id = $1 AND context_type = $2 AND context_ref = $3",
		tenantID.Value(), contextType.Value(), contextRef.Value(),
	).Scan(&containerID)

	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		"DELETE FROM element_positions WHERE tenant_id = $1 AND container_id = $2",
		tenantID.Value(), containerID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		"DELETE FROM layout_containers WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), containerID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *LayoutContainerRepository) DeleteElementFromAllLayouts(
	ctx context.Context,
	elementID valueobjects.ElementID,
) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		"DELETE FROM element_positions WHERE tenant_id = $1 AND element_id = $2",
		tenantID.Value(), elementID.Value(),
	)

	return err
}
