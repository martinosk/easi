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

func (rm *ApplicationComponentReadModel) MarkAsDeleted(ctx context.Context, id string, deletedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE application_components SET is_deleted = TRUE, deleted_at = $1 WHERE tenant_id = $2 AND id = $3",
		deletedAt, tenantID.Value(), id,
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
			"SELECT id, name, description, created_at FROM application_components WHERE tenant_id = $1 AND id = $2 AND is_deleted = FALSE",
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
			"SELECT id, name, description, created_at FROM application_components WHERE tenant_id = $1 AND is_deleted = FALSE ORDER BY created_at DESC",
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

type paginationQuery struct {
	tenantID       string
	afterCursor    string
	afterTimestamp int64
	limit          int
}

func (rm *ApplicationComponentReadModel) GetAllPaginated(ctx context.Context, limit int, afterCursor string, afterTimestamp int64) ([]ApplicationComponentDTO, bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, false, err
	}

	query := paginationQuery{
		tenantID:       tenantID.Value(),
		afterCursor:    afterCursor,
		afterTimestamp: afterTimestamp,
		limit:          limit + 1,
	}

	var components []ApplicationComponentDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := rm.queryPaginatedComponents(ctx, tx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		components, err = rm.scanComponents(rows)
		return err
	})

	if err != nil {
		return nil, false, err
	}

	return rm.trimAndCheckMore(components, limit)
}

func (rm *ApplicationComponentReadModel) queryPaginatedComponents(ctx context.Context, tx *sql.Tx, query paginationQuery) (*sql.Rows, error) {
	if query.afterCursor == "" {
		return tx.QueryContext(ctx,
			"SELECT id, name, description, created_at FROM application_components WHERE tenant_id = $1 AND is_deleted = FALSE ORDER BY created_at DESC, id DESC LIMIT $2",
			query.tenantID, query.limit,
		)
	}
	return tx.QueryContext(ctx,
		"SELECT id, name, description, created_at FROM application_components WHERE tenant_id = $1 AND is_deleted = FALSE AND (created_at < to_timestamp($2) OR (created_at = to_timestamp($2) AND id < $3)) ORDER BY created_at DESC, id DESC LIMIT $4",
		query.tenantID, query.afterTimestamp, query.afterCursor, query.limit,
	)
}

func (rm *ApplicationComponentReadModel) scanComponents(rows *sql.Rows) ([]ApplicationComponentDTO, error) {
	var components []ApplicationComponentDTO
	for rows.Next() {
		var dto ApplicationComponentDTO
		if err := rows.Scan(&dto.ID, &dto.Name, &dto.Description, &dto.CreatedAt); err != nil {
			return nil, err
		}
		components = append(components, dto)
	}
	return components, rows.Err()
}

func (rm *ApplicationComponentReadModel) trimAndCheckMore(components []ApplicationComponentDTO, limit int) ([]ApplicationComponentDTO, bool, error) {
	hasMore := len(components) > limit
	if hasMore {
		components = components[:limit]
	}
	return components, hasMore, nil
}
