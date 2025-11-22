package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type ViewID string
type ComponentID string

type Position struct {
	X float64
	Y float64
}

type ComponentPositionDTO struct {
	ComponentID string  `json:"componentId"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
}

func NewComponentPosition(componentID ComponentID, pos Position) ComponentPositionDTO {
	return ComponentPositionDTO{
		ComponentID: string(componentID),
		X:           pos.X,
		Y:           pos.Y,
	}
}

// ArchitectureViewDTO represents the read model for architecture views
type ArchitectureViewDTO struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	IsDefault       bool                   `json:"isDefault"`
	Components      []ComponentPositionDTO `json:"components"`
	CreatedAt       time.Time              `json:"createdAt"`
	EdgeType        string                 `json:"edgeType,omitempty"`
	LayoutDirection string                 `json:"layoutDirection,omitempty"`
	Links           map[string]string      `json:"_links,omitempty"`
}

// ArchitectureViewReadModel handles queries for architecture views
type ArchitectureViewReadModel struct {
	db *database.TenantAwareDB
}

// NewArchitectureViewReadModel creates a new read model
func NewArchitectureViewReadModel(db *database.TenantAwareDB) *ArchitectureViewReadModel {
	return &ArchitectureViewReadModel{db: db}
}

// InsertView adds a new view to the read model
func (rm *ArchitectureViewReadModel) InsertView(ctx context.Context, dto ArchitectureViewDTO) error {
	// Extract tenant from context - infrastructure concern
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	// Use tenant-aware exec that sets app.current_tenant for RLS
	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO architecture_views (id, tenant_id, name, description, is_default, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		dto.ID, tenantID.Value(), dto.Name, dto.Description, dto.IsDefault, dto.CreatedAt,
	)
	return err
}

func (rm *ArchitectureViewReadModel) AddComponent(ctx context.Context, viewID ViewID, componentID ComponentID, pos Position) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO view_component_positions (view_id, tenant_id, component_id, x, y, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		viewID, tenantID.Value(), componentID, pos.X, pos.Y, time.Now().UTC(),
	)
	return err
}

func (rm *ArchitectureViewReadModel) UpdateComponentPosition(ctx context.Context, viewID ViewID, componentID ComponentID, pos Position) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE view_component_positions SET x = $1, y = $2, updated_at = $3 WHERE tenant_id = $4 AND view_id = $5 AND component_id = $6",
		pos.X, pos.Y, time.Now().UTC(), tenantID.Value(), viewID, componentID,
	)
	return err
}

// RemoveComponent removes a component from a view
func (rm *ArchitectureViewReadModel) RemoveComponent(ctx context.Context, viewID, componentID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM view_component_positions WHERE tenant_id = $1 AND view_id = $2 AND component_id = $3",
		tenantID.Value(), viewID, componentID,
	)
	return err
}

// UpdateViewName updates a view's name
func (rm *ArchitectureViewReadModel) UpdateViewName(ctx context.Context, viewID, newName string) error {
	return rm.updateViewField(ctx, viewID, "name", newName)
}

// MarkViewAsDeleted marks a view as deleted
func (rm *ArchitectureViewReadModel) MarkViewAsDeleted(ctx context.Context, viewID string) error {
	return rm.updateViewField(ctx, viewID, "is_deleted", true)
}

// SetViewAsDefault sets a view as the default
func (rm *ArchitectureViewReadModel) SetViewAsDefault(ctx context.Context, viewID string, isDefault bool) error {
	return rm.updateViewField(ctx, viewID, "is_default", isDefault)
}

func (rm *ArchitectureViewReadModel) UpdateEdgeType(ctx context.Context, viewID, edgeType string) error {
	return rm.updateViewField(ctx, viewID, "edge_type", edgeType)
}

func (rm *ArchitectureViewReadModel) UpdateLayoutDirection(ctx context.Context, viewID, layoutDirection string) error {
	return rm.updateViewField(ctx, viewID, "layout_direction", layoutDirection)
}

func (rm *ArchitectureViewReadModel) updateViewField(ctx context.Context, viewID, field string, value interface{}) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	query := "UPDATE architecture_views SET " + field + " = $1, updated_at = $2 WHERE tenant_id = $3 AND id = $4"
	_, err = rm.db.ExecContext(ctx, query, value, time.Now().UTC(), tenantID.Value(), viewID)
	return err
}

// GetDefaultView retrieves the default view for the current tenant
func (rm *ArchitectureViewReadModel) GetDefaultView(ctx context.Context) (*ArchitectureViewDTO, error) {
	return rm.getViewByQuery(ctx,
		`SELECT av.id, av.name, av.description, av.is_default, av.created_at, vp.edge_type, vp.layout_direction
		FROM architecture_views av
		LEFT JOIN view_preferences vp ON av.id = vp.view_id AND av.tenant_id = vp.tenant_id
		WHERE av.tenant_id = $1 AND av.is_default = true AND av.is_deleted = false LIMIT 1`,
		func(tenantID string) []interface{} { return []interface{}{tenantID} },
	)
}

// GetByID retrieves a view by ID with all component positions for the current tenant
func (rm *ArchitectureViewReadModel) GetByID(ctx context.Context, id string) (*ArchitectureViewDTO, error) {
	return rm.getViewByQuery(ctx,
		`SELECT av.id, av.name, av.description, av.is_default, av.created_at, vp.edge_type, vp.layout_direction
		FROM architecture_views av
		LEFT JOIN view_preferences vp ON av.id = vp.view_id AND av.tenant_id = vp.tenant_id
		WHERE av.tenant_id = $1 AND av.id = $2 AND av.is_deleted = false`,
		func(tenantID string) []interface{} { return []interface{}{tenantID, id} },
	)
}

func (rm *ArchitectureViewReadModel) getViewByQuery(ctx context.Context, query string, argsBuilder func(tenantID string) []interface{}) (*ArchitectureViewDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto ArchitectureViewDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		found, err := rm.scanSingleView(ctx, tx, query, argsBuilder(tenantID.Value()), &dto)
		if err != nil {
			return err
		}
		if !found {
			notFound = true
			return nil
		}

		dto.Components, err = rm.getComponentsForViewTx(ctx, tx, tenantID.Value(), dto.ID)
		return err
	})

	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, nil
	}

	return &dto, nil
}

func (rm *ArchitectureViewReadModel) scanSingleView(ctx context.Context, tx *sql.Tx, query string, args []interface{}, dto *ArchitectureViewDTO) (bool, error) {
	var edgeType, layoutDirection sql.NullString
	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&dto.ID, &dto.Name, &dto.Description, &dto.IsDefault, &dto.CreatedAt, &edgeType, &layoutDirection,
	)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if edgeType.Valid {
		dto.EdgeType = edgeType.String
	}
	if layoutDirection.Valid {
		dto.LayoutDirection = layoutDirection.String
	}

	return true, nil
}

func (rm *ArchitectureViewReadModel) GetAll(ctx context.Context) ([]ArchitectureViewDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var views []ArchitectureViewDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		views, err = rm.queryViews(ctx, tx, tenantID.Value())
		if err != nil {
			return err
		}

		return rm.populateViewComponents(ctx, tx, tenantID.Value(), views)
	})

	return views, err
}

func (rm *ArchitectureViewReadModel) queryViews(ctx context.Context, tx *sql.Tx, tenantID string) ([]ArchitectureViewDTO, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT av.id, av.name, av.description, av.is_default, av.created_at, vp.edge_type, vp.layout_direction
		FROM architecture_views av
		LEFT JOIN view_preferences vp ON av.id = vp.view_id AND av.tenant_id = vp.tenant_id
		WHERE av.tenant_id = $1 AND av.is_deleted = false ORDER BY av.is_default DESC, av.created_at DESC`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []ArchitectureViewDTO
	for rows.Next() {
		dto, err := rm.scanViewRow(rows)
		if err != nil {
			return nil, err
		}
		views = append(views, dto)
	}

	return views, rows.Err()
}

func (rm *ArchitectureViewReadModel) scanViewRow(rows *sql.Rows) (ArchitectureViewDTO, error) {
	var dto ArchitectureViewDTO
	var edgeType, layoutDirection sql.NullString

	err := rows.Scan(&dto.ID, &dto.Name, &dto.Description, &dto.IsDefault, &dto.CreatedAt, &edgeType, &layoutDirection)
	if err != nil {
		return dto, err
	}

	if edgeType.Valid {
		dto.EdgeType = edgeType.String
	}
	if layoutDirection.Valid {
		dto.LayoutDirection = layoutDirection.String
	}

	return dto, nil
}

func (rm *ArchitectureViewReadModel) populateViewComponents(ctx context.Context, tx *sql.Tx, tenantID string, views []ArchitectureViewDTO) error {
	for i := range views {
		components, err := rm.getComponentsForViewTx(ctx, tx, tenantID, views[i].ID)
		if err != nil {
			return err
		}
		views[i].Components = components
	}
	return nil
}

// getComponentsForViewTx is a helper to fetch components for a view within a transaction
func (rm *ArchitectureViewReadModel) getComponentsForViewTx(ctx context.Context, tx *sql.Tx, tenantID, viewID string) ([]ComponentPositionDTO, error) {
	rows, err := tx.QueryContext(ctx,
		"SELECT component_id, x, y FROM view_component_positions WHERE tenant_id = $1 AND view_id = $2",
		tenantID, viewID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	components := make([]ComponentPositionDTO, 0)
	for rows.Next() {
		var comp ComponentPositionDTO
		if err := rows.Scan(&comp.ComponentID, &comp.X, &comp.Y); err != nil {
			return nil, err
		}
		components = append(components, comp)
	}

	return components, rows.Err()
}

func (rm *ArchitectureViewReadModel) GetViewsContainingComponent(ctx context.Context, componentID string) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var viewIDs []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT DISTINCT view_id FROM view_component_positions WHERE tenant_id = $1 AND component_id = $2",
			tenantID.Value(), componentID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var viewID string
			if err := rows.Scan(&viewID); err != nil {
				return err
			}
			viewIDs = append(viewIDs, viewID)
		}

		return rows.Err()
	})

	return viewIDs, err
}
