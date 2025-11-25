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

type ElementType string

const (
	ElementTypeComponent  ElementType = "component"
	ElementTypeCapability ElementType = "capability"
)

type Position struct {
	X float64
	Y float64
}

type ComponentPositionDTO struct {
	ComponentID string            `json:"componentId"`
	X           float64           `json:"x"`
	Y           float64           `json:"y"`
	CustomColor *string           `json:"customColor,omitempty"`
	Links       map[string]string `json:"_links,omitempty"`
}

type CapabilityPositionDTO struct {
	CapabilityID string            `json:"capabilityId"`
	X            float64           `json:"x"`
	Y            float64           `json:"y"`
	CustomColor  *string           `json:"customColor,omitempty"`
	Links        map[string]string `json:"_links,omitempty"`
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
	ID              string                  `json:"id"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description,omitempty"`
	IsDefault       bool                    `json:"isDefault"`
	Components      []ComponentPositionDTO  `json:"components"`
	Capabilities    []CapabilityPositionDTO `json:"capabilities"`
	CreatedAt       time.Time               `json:"createdAt"`
	EdgeType        string                  `json:"edgeType,omitempty"`
	LayoutDirection string                  `json:"layoutDirection,omitempty"`
	ColorScheme     string                  `json:"colorScheme,omitempty"`
	Links           map[string]string       `json:"_links,omitempty"`
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
	return rm.addElement(ctx, string(viewID), string(componentID), ElementTypeComponent, pos)
}

func (rm *ArchitectureViewReadModel) UpdateComponentPosition(ctx context.Context, viewID ViewID, componentID ComponentID, pos Position) error {
	return rm.updateElementPosition(ctx, string(viewID), string(componentID), ElementTypeComponent, pos)
}

func (rm *ArchitectureViewReadModel) RemoveComponent(ctx context.Context, viewID, componentID string) error {
	return rm.removeElement(ctx, viewID, componentID, ElementTypeComponent)
}

func (rm *ArchitectureViewReadModel) AddCapability(ctx context.Context, viewID, capabilityID string, pos Position) error {
	return rm.addElement(ctx, viewID, capabilityID, ElementTypeCapability, pos)
}

func (rm *ArchitectureViewReadModel) UpdateCapabilityPosition(ctx context.Context, viewID, capabilityID string, pos Position) error {
	return rm.updateElementPosition(ctx, viewID, capabilityID, ElementTypeCapability, pos)
}

func (rm *ArchitectureViewReadModel) RemoveCapability(ctx context.Context, viewID, capabilityID string) error {
	return rm.removeElement(ctx, viewID, capabilityID, ElementTypeCapability)
}

func (rm *ArchitectureViewReadModel) addElement(ctx context.Context, viewID, elementID string, elementType ElementType, pos Position) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO view_element_positions (view_id, tenant_id, element_id, element_type, x, y, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		viewID, tenantID.Value(), elementID, string(elementType), pos.X, pos.Y, time.Now().UTC(),
	)
	return err
}

func (rm *ArchitectureViewReadModel) updateElementPosition(ctx context.Context, viewID, elementID string, elementType ElementType, pos Position) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE view_element_positions SET x = $1, y = $2, updated_at = $3 WHERE tenant_id = $4 AND view_id = $5 AND element_id = $6 AND element_type = $7",
		pos.X, pos.Y, time.Now().UTC(), tenantID.Value(), viewID, elementID, string(elementType),
	)
	return err
}

func (rm *ArchitectureViewReadModel) removeElement(ctx context.Context, viewID, elementID string, elementType ElementType) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM view_element_positions WHERE tenant_id = $1 AND view_id = $2 AND element_id = $3 AND element_type = $4",
		tenantID.Value(), viewID, elementID, string(elementType),
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
		`SELECT av.id, av.name, av.description, av.is_default, av.created_at, vp.edge_type, vp.layout_direction, vp.color_scheme
		FROM architecture_views av
		LEFT JOIN view_preferences vp ON av.id = vp.view_id AND av.tenant_id = vp.tenant_id
		WHERE av.tenant_id = $1 AND av.is_default = true AND av.is_deleted = false LIMIT 1`,
		func(tenantID string) []interface{} { return []interface{}{tenantID} },
	)
}

// GetByID retrieves a view by ID with all component positions for the current tenant
func (rm *ArchitectureViewReadModel) GetByID(ctx context.Context, id string) (*ArchitectureViewDTO, error) {
	return rm.getViewByQuery(ctx,
		`SELECT av.id, av.name, av.description, av.is_default, av.created_at, vp.edge_type, vp.layout_direction, vp.color_scheme
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
		if err != nil {
			return err
		}

		dto.Capabilities, err = rm.getCapabilitiesForViewTx(ctx, tx, tenantID.Value(), dto.ID)
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
	var edgeType, layoutDirection, colorScheme sql.NullString
	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&dto.ID, &dto.Name, &dto.Description, &dto.IsDefault, &dto.CreatedAt, &edgeType, &layoutDirection, &colorScheme,
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
	if colorScheme.Valid {
		dto.ColorScheme = colorScheme.String
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
		`SELECT av.id, av.name, av.description, av.is_default, av.created_at, vp.edge_type, vp.layout_direction, vp.color_scheme
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
	var edgeType, layoutDirection, colorScheme sql.NullString

	err := rows.Scan(&dto.ID, &dto.Name, &dto.Description, &dto.IsDefault, &dto.CreatedAt, &edgeType, &layoutDirection, &colorScheme)
	if err != nil {
		return dto, err
	}

	if edgeType.Valid {
		dto.EdgeType = edgeType.String
	}
	if layoutDirection.Valid {
		dto.LayoutDirection = layoutDirection.String
	}
	if colorScheme.Valid {
		dto.ColorScheme = colorScheme.String
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

		capabilities, err := rm.getCapabilitiesForViewTx(ctx, tx, tenantID, views[i].ID)
		if err != nil {
			return err
		}
		views[i].Capabilities = capabilities
	}
	return nil
}

func (rm *ArchitectureViewReadModel) getComponentsForViewTx(ctx context.Context, tx *sql.Tx, tenantID, viewID string) ([]ComponentPositionDTO, error) {
	return getElementsForViewTx(ctx, tx, tenantID, viewID, ElementTypeComponent, func(rows *sql.Rows) (ComponentPositionDTO, error) {
		var dto ComponentPositionDTO
		var customColor sql.NullString
		err := rows.Scan(&dto.ComponentID, &dto.X, &dto.Y, &customColor)
		if customColor.Valid {
			dto.CustomColor = &customColor.String
		}
		return dto, err
	})
}

func (rm *ArchitectureViewReadModel) getCapabilitiesForViewTx(ctx context.Context, tx *sql.Tx, tenantID, viewID string) ([]CapabilityPositionDTO, error) {
	return getElementsForViewTx(ctx, tx, tenantID, viewID, ElementTypeCapability, func(rows *sql.Rows) (CapabilityPositionDTO, error) {
		var dto CapabilityPositionDTO
		var customColor sql.NullString
		err := rows.Scan(&dto.CapabilityID, &dto.X, &dto.Y, &customColor)
		if customColor.Valid {
			dto.CustomColor = &customColor.String
		}
		return dto, err
	})
}

func getElementsForViewTx[T any](ctx context.Context, tx *sql.Tx, tenantID, viewID string, elementType ElementType, scan func(*sql.Rows) (T, error)) ([]T, error) {
	rows, err := tx.QueryContext(ctx,
		"SELECT element_id, x, y, custom_color FROM view_element_positions WHERE tenant_id = $1 AND view_id = $2 AND element_type = $3",
		tenantID, viewID, string(elementType),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	elements := make([]T, 0)
	for rows.Next() {
		elem, err := scan(rows)
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
	}

	return elements, rows.Err()
}

func (rm *ArchitectureViewReadModel) GetViewsContainingComponent(ctx context.Context, componentID string) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var viewIDs []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT DISTINCT view_id FROM view_element_positions WHERE tenant_id = $1 AND element_id = $2 AND element_type = 'component'",
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
