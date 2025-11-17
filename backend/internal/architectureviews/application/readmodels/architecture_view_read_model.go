package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

// ComponentPositionDTO represents a component's position on a view
type ComponentPositionDTO struct {
	ComponentID string  `json:"componentId"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
}

// ArchitectureViewDTO represents the read model for architecture views
type ArchitectureViewDTO struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	IsDefault   bool                    `json:"isDefault"`
	Components  []ComponentPositionDTO  `json:"components"`
	CreatedAt   time.Time               `json:"createdAt"`
	Links       map[string]string       `json:"_links,omitempty"`
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

// AddComponent adds a component position to a view
func (rm *ArchitectureViewReadModel) AddComponent(ctx context.Context, viewID, componentID string, x, y float64) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO view_component_positions (view_id, tenant_id, component_id, x, y, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		viewID, tenantID.Value(), componentID, x, y, time.Now().UTC(),
	)
	return err
}

// UpdateComponentPosition updates a component's position in a view
func (rm *ArchitectureViewReadModel) UpdateComponentPosition(ctx context.Context, viewID, componentID string, x, y float64) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE view_component_positions SET x = $1, y = $2, updated_at = $3 WHERE tenant_id = $4 AND view_id = $5 AND component_id = $6",
		x, y, time.Now().UTC(), tenantID.Value(), viewID, componentID,
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
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE architecture_views SET name = $1, updated_at = $2 WHERE tenant_id = $3 AND id = $4",
		newName, time.Now().UTC(), tenantID.Value(), viewID,
	)
	return err
}

// MarkViewAsDeleted marks a view as deleted
func (rm *ArchitectureViewReadModel) MarkViewAsDeleted(ctx context.Context, viewID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE architecture_views SET is_deleted = true, updated_at = $1 WHERE tenant_id = $2 AND id = $3",
		time.Now().UTC(), tenantID.Value(), viewID,
	)
	return err
}

// SetViewAsDefault sets a view as the default
func (rm *ArchitectureViewReadModel) SetViewAsDefault(ctx context.Context, viewID string, isDefault bool) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE architecture_views SET is_default = $1, updated_at = $2 WHERE tenant_id = $3 AND id = $4",
		isDefault, time.Now().UTC(), tenantID.Value(), viewID,
	)
	return err
}

// GetDefaultView retrieves the default view for the current tenant
func (rm *ArchitectureViewReadModel) GetDefaultView(ctx context.Context) (*ArchitectureViewDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto ArchitectureViewDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, name, description, is_default, created_at FROM architecture_views WHERE tenant_id = $1 AND is_default = true AND is_deleted = false LIMIT 1",
			tenantID.Value(),
		).Scan(&dto.ID, &dto.Name, &dto.Description, &dto.IsDefault, &dto.CreatedAt)

		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
		if err != nil {
			return err
		}

		// Get component positions within same transaction
		components, err := rm.getComponentsForViewTx(ctx, tx, tenantID.Value(), dto.ID)
		if err != nil {
			return err
		}
		dto.Components = components

		return nil
	})

	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, nil
	}

	return &dto, nil
}

// GetByID retrieves a view by ID with all component positions for the current tenant
func (rm *ArchitectureViewReadModel) GetByID(ctx context.Context, id string) (*ArchitectureViewDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto ArchitectureViewDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, name, description, is_default, created_at FROM architecture_views WHERE tenant_id = $1 AND id = $2 AND is_deleted = false",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.Name, &dto.Description, &dto.IsDefault, &dto.CreatedAt)

		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
		if err != nil {
			return err
		}

		// Get component positions within same transaction
		rows, err := tx.QueryContext(ctx,
			"SELECT component_id, x, y FROM view_component_positions WHERE tenant_id = $1 AND view_id = $2",
			tenantID.Value(), id,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		dto.Components = make([]ComponentPositionDTO, 0)
		for rows.Next() {
			var comp ComponentPositionDTO
			if err := rows.Scan(&comp.ComponentID, &comp.X, &comp.Y); err != nil {
				return err
			}
			dto.Components = append(dto.Components, comp)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, nil
	}

	return &dto, nil
}

// GetAll retrieves all views (excluding deleted ones) for the current tenant
func (rm *ArchitectureViewReadModel) GetAll(ctx context.Context) ([]ArchitectureViewDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var views []ArchitectureViewDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, name, description, is_default, created_at FROM architecture_views WHERE tenant_id = $1 AND is_deleted = false ORDER BY is_default DESC, created_at DESC",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto ArchitectureViewDTO
			if err := rows.Scan(&dto.ID, &dto.Name, &dto.Description, &dto.IsDefault, &dto.CreatedAt); err != nil {
				return err
			}

			// Get component positions for this view within same transaction
			components, err := rm.getComponentsForViewTx(ctx, tx, tenantID.Value(), dto.ID)
			if err != nil {
				return err
			}
			dto.Components = components
			views = append(views, dto)
		}

		return rows.Err()
	})

	return views, err
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
