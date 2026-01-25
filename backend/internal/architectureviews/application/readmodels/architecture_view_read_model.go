package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type ViewID string
type ComponentID string
type CapabilityID string

type ElementType string

const (
	ElementTypeComponent    ElementType = "component"
	ElementTypeCapability   ElementType = "capability"
	ElementTypeOriginEntity ElementType = "origin_entity"
)

type Position struct {
	X float64
	Y float64
}

type ElementPosition struct {
	ViewID      ViewID
	ElementID   string
	ElementType ElementType
	Position    Position
}

func NewElementPositionForComponent(viewID ViewID, componentID ComponentID, pos Position) ElementPosition {
	return ElementPosition{
		ViewID:      viewID,
		ElementID:   string(componentID),
		ElementType: ElementTypeComponent,
		Position:    pos,
	}
}

func NewElementPositionForCapability(viewID ViewID, capabilityID CapabilityID, pos Position) ElementPosition {
	return ElementPosition{
		ViewID:      viewID,
		ElementID:   string(capabilityID),
		ElementType: ElementTypeCapability,
		Position:    pos,
	}
}

type ViewFieldUpdate struct {
	ViewID ViewID
	Field  string
	Value  interface{}
}

type VisibilityUpdate struct {
	ViewID      string
	IsPrivate   bool
	OwnerUserID string
	OwnerEmail  string
}

type ComponentPositionDTO struct {
	ComponentID string      `json:"componentId"`
	X           float64     `json:"x"`
	Y           float64     `json:"y"`
	CustomColor *string     `json:"customColor,omitempty"`
	Links       types.Links `json:"_links,omitempty"`
}

type CapabilityPositionDTO struct {
	CapabilityID string      `json:"capabilityId"`
	X            float64     `json:"x"`
	Y            float64     `json:"y"`
	CustomColor  *string     `json:"customColor,omitempty"`
	Links        types.Links `json:"_links,omitempty"`
}

type OriginEntityPositionDTO struct {
	OriginEntityID string      `json:"originEntityId"`
	X              float64     `json:"x"`
	Y              float64     `json:"y"`
	Links          types.Links `json:"_links,omitempty"`
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
	ID              string                    `json:"id"`
	Name            string                    `json:"name"`
	Description     string                    `json:"description,omitempty"`
	IsDefault       bool                      `json:"isDefault"`
	IsPrivate       bool                      `json:"isPrivate"`
	OwnerUserID     *string                   `json:"ownerUserId,omitempty"`
	OwnerEmail      *string                   `json:"ownerEmail,omitempty"`
	Components      []ComponentPositionDTO    `json:"components"`
	Capabilities    []CapabilityPositionDTO   `json:"capabilities"`
	OriginEntities  []OriginEntityPositionDTO `json:"originEntities"`
	CreatedAt       time.Time                 `json:"createdAt"`
	EdgeType        string                    `json:"edgeType,omitempty"`
	LayoutDirection string                    `json:"layoutDirection,omitempty"`
	ColorScheme     string                    `json:"colorScheme,omitempty"`
	Links           types.Links               `json:"_links,omitempty"`
}

// ArchitectureViewReadModel handles queries for architecture views
type ArchitectureViewReadModel struct {
	db *database.TenantAwareDB
}

// NewArchitectureViewReadModel creates a new read model
func NewArchitectureViewReadModel(db *database.TenantAwareDB) *ArchitectureViewReadModel {
	return &ArchitectureViewReadModel{db: db}
}

func (rm *ArchitectureViewReadModel) getTenantID(ctx context.Context) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}
	return tenantID.Value(), nil
}

// InsertView adds a new view to the read model
func (rm *ArchitectureViewReadModel) InsertView(ctx context.Context, dto ArchitectureViewDTO) error {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO architecture_views (id, tenant_id, name, description, is_default, is_private, owner_user_id, owner_email, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		dto.ID, tenantID, dto.Name, dto.Description, dto.IsDefault, dto.IsPrivate, dto.OwnerUserID, dto.OwnerEmail, dto.CreatedAt,
	)
	return err
}

func (rm *ArchitectureViewReadModel) AddComponent(ctx context.Context, viewID ViewID, componentID ComponentID, pos Position) error {
	elem := NewElementPositionForComponent(viewID, componentID, pos)
	return rm.addElement(ctx, elem)
}

func (rm *ArchitectureViewReadModel) UpdateComponentPosition(ctx context.Context, viewID ViewID, componentID ComponentID, pos Position) error {
	elem := NewElementPositionForComponent(viewID, componentID, pos)
	return rm.updateElementPosition(ctx, elem)
}

func (rm *ArchitectureViewReadModel) RemoveComponent(ctx context.Context, viewID, componentID string) error {
	return rm.removeElement(ctx, ViewID(viewID), componentID, ElementTypeComponent)
}

func (rm *ArchitectureViewReadModel) AddCapability(ctx context.Context, viewID, capabilityID string, pos Position) error {
	elem := NewElementPositionForCapability(ViewID(viewID), CapabilityID(capabilityID), pos)
	return rm.addElement(ctx, elem)
}

func (rm *ArchitectureViewReadModel) UpdateCapabilityPosition(ctx context.Context, viewID, capabilityID string, pos Position) error {
	elem := NewElementPositionForCapability(ViewID(viewID), CapabilityID(capabilityID), pos)
	return rm.updateElementPosition(ctx, elem)
}

func (rm *ArchitectureViewReadModel) RemoveCapability(ctx context.Context, viewID, capabilityID string) error {
	return rm.removeElement(ctx, ViewID(viewID), capabilityID, ElementTypeCapability)
}

func (rm *ArchitectureViewReadModel) addElement(ctx context.Context, elem ElementPosition) error {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO view_element_positions (view_id, tenant_id, element_id, element_type, x, y, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		string(elem.ViewID), tenantID, elem.ElementID, string(elem.ElementType), elem.Position.X, elem.Position.Y, time.Now().UTC(),
	)
	return err
}

func (rm *ArchitectureViewReadModel) updateElementPosition(ctx context.Context, elem ElementPosition) error {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE view_element_positions SET x = $1, y = $2, updated_at = $3 WHERE tenant_id = $4 AND view_id = $5 AND element_id = $6 AND element_type = $7",
		elem.Position.X, elem.Position.Y, time.Now().UTC(), tenantID, string(elem.ViewID), elem.ElementID, string(elem.ElementType),
	)
	return err
}

func (rm *ArchitectureViewReadModel) removeElement(ctx context.Context, viewID ViewID, elementID string, elementType ElementType) error {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM view_element_positions WHERE tenant_id = $1 AND view_id = $2 AND element_id = $3 AND element_type = $4",
		tenantID, string(viewID), elementID, string(elementType),
	)
	return err
}

// UpdateViewName updates a view's name
func (rm *ArchitectureViewReadModel) UpdateViewName(ctx context.Context, viewID, newName string) error {
	update := ViewFieldUpdate{ViewID: ViewID(viewID), Field: "name", Value: newName}
	return rm.updateViewField(ctx, update)
}

// MarkViewAsDeleted marks a view as deleted
func (rm *ArchitectureViewReadModel) MarkViewAsDeleted(ctx context.Context, viewID string) error {
	update := ViewFieldUpdate{ViewID: ViewID(viewID), Field: "is_deleted", Value: true}
	return rm.updateViewField(ctx, update)
}

// SetViewAsDefault sets a view as the default
func (rm *ArchitectureViewReadModel) SetViewAsDefault(ctx context.Context, viewID string, isDefault bool) error {
	update := ViewFieldUpdate{ViewID: ViewID(viewID), Field: "is_default", Value: isDefault}
	return rm.updateViewField(ctx, update)
}

func (rm *ArchitectureViewReadModel) UpdateEdgeType(ctx context.Context, viewID, edgeType string) error {
	update := ViewFieldUpdate{ViewID: ViewID(viewID), Field: "edge_type", Value: edgeType}
	return rm.updateViewField(ctx, update)
}

func (rm *ArchitectureViewReadModel) UpdateLayoutDirection(ctx context.Context, viewID, layoutDirection string) error {
	update := ViewFieldUpdate{ViewID: ViewID(viewID), Field: "layout_direction", Value: layoutDirection}
	return rm.updateViewField(ctx, update)
}

func (rm *ArchitectureViewReadModel) UpdateVisibility(ctx context.Context, update VisibilityUpdate) error {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE architecture_views SET is_private = $1, owner_user_id = $2, owner_email = $3, updated_at = $4 WHERE tenant_id = $5 AND id = $6",
		update.IsPrivate, update.OwnerUserID, update.OwnerEmail, time.Now().UTC(), tenantID, update.ViewID,
	)
	return err
}

func (rm *ArchitectureViewReadModel) updateViewField(ctx context.Context, update ViewFieldUpdate) error {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return err
	}

	query := "UPDATE architecture_views SET " + update.Field + " = $1, updated_at = $2 WHERE tenant_id = $3 AND id = $4"
	_, err = rm.db.ExecContext(ctx, query, update.Value, time.Now().UTC(), tenantID, string(update.ViewID))
	return err
}

// GetDefaultView retrieves the default view for the current tenant
func (rm *ArchitectureViewReadModel) GetDefaultView(ctx context.Context) (*ArchitectureViewDTO, error) {
	return rm.getViewByQuery(ctx,
		`SELECT av.id, av.name, av.description, av.is_default, av.is_private, av.owner_user_id, av.owner_email, av.created_at, vp.edge_type, vp.layout_direction, vp.color_scheme
		FROM architecture_views av
		LEFT JOIN view_preferences vp ON av.id = vp.view_id AND av.tenant_id = vp.tenant_id
		WHERE av.tenant_id = $1 AND av.is_default = true AND av.is_deleted = false LIMIT 1`,
		func(tenantID string) []interface{} { return []interface{}{tenantID} },
	)
}

// GetByID retrieves a view by ID with all component positions for the current tenant
func (rm *ArchitectureViewReadModel) GetByID(ctx context.Context, id string) (*ArchitectureViewDTO, error) {
	return rm.getViewByQuery(ctx,
		`SELECT av.id, av.name, av.description, av.is_default, av.is_private, av.owner_user_id, av.owner_email, av.created_at, vp.edge_type, vp.layout_direction, vp.color_scheme
		FROM architecture_views av
		LEFT JOIN view_preferences vp ON av.id = vp.view_id AND av.tenant_id = vp.tenant_id
		WHERE av.tenant_id = $1 AND av.id = $2 AND av.is_deleted = false`,
		func(tenantID string) []interface{} { return []interface{}{tenantID, id} },
	)
}

func (rm *ArchitectureViewReadModel) getViewByQuery(ctx context.Context, query string, argsBuilder func(tenantID string) []interface{}) (*ArchitectureViewDTO, error) {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var dto ArchitectureViewDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		found, err := rm.scanSingleView(ctx, tx, query, argsBuilder(tenantID), &dto)
		if err != nil {
			return err
		}
		if !found {
			notFound = true
			return nil
		}

		dto.Components, err = rm.getComponentsForViewTx(ctx, tx, tenantID, dto.ID)
		if err != nil {
			return err
		}

		dto.Capabilities, err = rm.getCapabilitiesForViewTx(ctx, tx, tenantID, dto.ID)
		if err != nil {
			return err
		}

		dto.OriginEntities, err = rm.getOriginEntitiesForViewTx(ctx, tx, tenantID, dto.ID)
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

type viewScanFields struct {
	ownerUserID     sql.NullString
	ownerEmail      sql.NullString
	edgeType        sql.NullString
	layoutDirection sql.NullString
	colorScheme     sql.NullString
}

func (f *viewScanFields) applyTo(dto *ArchitectureViewDTO) {
	if f.ownerUserID.Valid {
		dto.OwnerUserID = &f.ownerUserID.String
	}
	if f.ownerEmail.Valid {
		dto.OwnerEmail = &f.ownerEmail.String
	}
	if f.edgeType.Valid {
		dto.EdgeType = f.edgeType.String
	}
	if f.layoutDirection.Valid {
		dto.LayoutDirection = f.layoutDirection.String
	}
	if f.colorScheme.Valid {
		dto.ColorScheme = f.colorScheme.String
	}
}

func (rm *ArchitectureViewReadModel) scanSingleView(ctx context.Context, tx *sql.Tx, query string, args []interface{}, dto *ArchitectureViewDTO) (bool, error) {
	var fields viewScanFields
	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&dto.ID, &dto.Name, &dto.Description, &dto.IsDefault, &dto.IsPrivate, &fields.ownerUserID, &fields.ownerEmail, &dto.CreatedAt, &fields.edgeType, &fields.layoutDirection, &fields.colorScheme,
	)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	fields.applyTo(dto)
	return true, nil
}

func (rm *ArchitectureViewReadModel) GetAll(ctx context.Context) ([]ArchitectureViewDTO, error) {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var views []ArchitectureViewDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		views, err = rm.queryViews(ctx, tx, tenantID)
		if err != nil {
			return err
		}

		return rm.populateViewComponents(ctx, tx, tenantID, views)
	})

	return views, err
}

func (rm *ArchitectureViewReadModel) queryViews(ctx context.Context, tx *sql.Tx, tenantID string) ([]ArchitectureViewDTO, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT av.id, av.name, av.description, av.is_default, av.is_private, av.owner_user_id, av.owner_email, av.created_at, vp.edge_type, vp.layout_direction, vp.color_scheme
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
	var fields viewScanFields

	err := rows.Scan(&dto.ID, &dto.Name, &dto.Description, &dto.IsDefault, &dto.IsPrivate, &fields.ownerUserID, &fields.ownerEmail, &dto.CreatedAt, &fields.edgeType, &fields.layoutDirection, &fields.colorScheme)
	if err != nil {
		return dto, err
	}

	fields.applyTo(&dto)
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

		originEntities, err := rm.getOriginEntitiesForViewTx(ctx, tx, tenantID, views[i].ID)
		if err != nil {
			return err
		}
		views[i].OriginEntities = originEntities
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

func (rm *ArchitectureViewReadModel) getOriginEntitiesForViewTx(ctx context.Context, tx *sql.Tx, tenantID, viewID string) ([]OriginEntityPositionDTO, error) {
	return getElementsForViewTx(ctx, tx, tenantID, viewID, ElementTypeOriginEntity, func(rows *sql.Rows) (OriginEntityPositionDTO, error) {
		var dto OriginEntityPositionDTO
		var customColor sql.NullString
		err := rows.Scan(&dto.OriginEntityID, &dto.X, &dto.Y, &customColor)
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

type ViewAuthInfo struct {
	IsPrivate   bool
	OwnerUserID *string
}

func (rm *ArchitectureViewReadModel) GetAuthInfo(ctx context.Context, viewID string) (*ViewAuthInfo, error) {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var info *ViewAuthInfo
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		var ownerUserID sql.NullString
		var isPrivate bool

		err := tx.QueryRowContext(ctx,
			"SELECT is_private, owner_user_id FROM architecture_views WHERE tenant_id = $1 AND id = $2 AND is_deleted = false",
			tenantID, viewID,
		).Scan(&isPrivate, &ownerUserID)

		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return err
		}

		info = &ViewAuthInfo{IsPrivate: isPrivate}
		if ownerUserID.Valid {
			info.OwnerUserID = &ownerUserID.String
		}

		return nil
	})

	return info, err
}

func (rm *ArchitectureViewReadModel) GetViewsContainingComponent(ctx context.Context, componentID string) ([]string, error) {
	tenantID, err := rm.getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var viewIDs []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT DISTINCT view_id FROM view_element_positions WHERE tenant_id = $1 AND element_id = $2 AND element_type = 'component'",
			tenantID, componentID,
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
