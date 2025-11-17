package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

// ApplicationComponentDTO represents the read model for application components
type ApplicationComponentDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	Links       map[string]string `json:"_links,omitempty"`
}

// ApplicationComponentReadModel handles queries for application components
type ApplicationComponentReadModel struct {
	db *database.TenantAwareDB
}

// NewApplicationComponentReadModel creates a new read model
func NewApplicationComponentReadModel(db *database.TenantAwareDB) *ApplicationComponentReadModel {
	return &ApplicationComponentReadModel{db: db}
}

// Insert adds a new component to the read model
func (rm *ApplicationComponentReadModel) Insert(ctx context.Context, dto ApplicationComponentDTO) error {
	// Extract tenant from context - infrastructure concern
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO application_components (id, tenant_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		dto.ID, tenantID.Value(), dto.Name, dto.Description, dto.CreatedAt,
	)
	return err
}

// Update updates an existing component in the read model
// RLS policies ensure we can only update our tenant's rows
func (rm *ApplicationComponentReadModel) Update(ctx context.Context, id, name, description string) error {
	// Extract tenant for defense-in-depth filtering
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE application_components SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $3 AND id = $4",
		name, description, tenantID.Value(), id,
	)
	return err
}

// GetByID retrieves a component by ID
// RLS policies automatically filter by tenant, but we add explicit filter for defense-in-depth
func (rm *ApplicationComponentReadModel) GetByID(ctx context.Context, id string) (*ApplicationComponentDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto ApplicationComponentDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			"SELECT id, name, description, created_at FROM application_components WHERE tenant_id = $1 AND id = $2",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.Name, &dto.Description, &dto.CreatedAt)

		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
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

// GetAll retrieves all components for the current tenant
// RLS policies automatically filter, but we add explicit filter for defense-in-depth
func (rm *ApplicationComponentReadModel) GetAll(ctx context.Context) ([]ApplicationComponentDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var components []ApplicationComponentDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, name, description, created_at FROM application_components WHERE tenant_id = $1 ORDER BY created_at DESC",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto ApplicationComponentDTO
			if err := rows.Scan(&dto.ID, &dto.Name, &dto.Description, &dto.CreatedAt); err != nil {
				return err
			}
			components = append(components, dto)
		}

		return rows.Err()
	})

	return components, err
}

// GetAllPaginated retrieves components with cursor-based pagination for the current tenant
func (rm *ApplicationComponentReadModel) GetAllPaginated(ctx context.Context, limit int, afterCursor string, afterTimestamp int64) ([]ApplicationComponentDTO, bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, false, err
	}

	// Query one extra to determine if there are more results
	queryLimit := limit + 1

	var components []ApplicationComponentDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		var rows *sql.Rows
		var err error

		if afterCursor == "" {
			// No cursor, get first page
			rows, err = tx.QueryContext(ctx,
				"SELECT id, name, description, created_at FROM application_components WHERE tenant_id = $1 ORDER BY created_at DESC, id DESC LIMIT $2",
				tenantID.Value(), queryLimit,
			)
		} else {
			// Use cursor for pagination
			rows, err = tx.QueryContext(ctx,
				"SELECT id, name, description, created_at FROM application_components WHERE tenant_id = $1 AND (created_at < to_timestamp($2) OR (created_at = to_timestamp($2) AND id < $3)) ORDER BY created_at DESC, id DESC LIMIT $4",
				tenantID.Value(), afterTimestamp, afterCursor, queryLimit,
			)
		}

		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto ApplicationComponentDTO
			if err := rows.Scan(&dto.ID, &dto.Name, &dto.Description, &dto.CreatedAt); err != nil {
				return err
			}
			components = append(components, dto)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, false, err
	}

	// Check if there are more results
	hasMore := len(components) > limit
	if hasMore {
		// Remove the extra item
		components = components[:limit]
	}

	return components, hasMore, nil
}
