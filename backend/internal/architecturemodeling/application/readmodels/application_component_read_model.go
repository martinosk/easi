package readmodels

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

// ApplicationComponentDTO represents the read model for application components
type ApplicationComponentDTO struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	Experts     []ExpertDTO `json:"experts,omitempty"`
	Links       types.Links `json:"_links,omitempty"`
}

type ExpertDTO struct {
	Name    string      `json:"name"`
	Role    string      `json:"role"`
	Contact string      `json:"contact"`
	AddedAt time.Time   `json:"addedAt"`
	Links   types.Links `json:"_links,omitempty"`
}

type ExpertInfo struct {
	ComponentID string
	Name        string
	Role        string
	Contact     string
	AddedAt     time.Time
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
		if err != nil {
			return err
		}

		dto.Experts, err = rm.fetchExperts(ctx, tx, tenantID.Value(), id)
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
			"SELECT id, name, description, created_at FROM application_components WHERE tenant_id = $1 AND is_deleted = FALSE ORDER BY LOWER(name) ASC, id ASC",
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
	tenantID    string
	afterCursor string
	afterName   string
	limit       int
}

func (rm *ApplicationComponentReadModel) GetAllPaginated(ctx context.Context, limit int, afterCursor string, afterName string) ([]ApplicationComponentDTO, bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, false, err
	}

	query := paginationQuery{
		tenantID:    tenantID.Value(),
		afterCursor: afterCursor,
		afterName:   afterName,
		limit:       limit + 1,
	}

	var components []ApplicationComponentDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := rm.queryPaginatedComponents(ctx, tx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		components, err = rm.scanComponents(rows)
		if err != nil {
			return err
		}

		return rm.loadExpertsForComponents(ctx, tx, tenantID.Value(), components)
	})

	if err != nil {
		return nil, false, err
	}

	return rm.trimAndCheckMore(components, limit)
}

func (rm *ApplicationComponentReadModel) queryPaginatedComponents(ctx context.Context, tx *sql.Tx, query paginationQuery) (*sql.Rows, error) {
	if query.afterCursor == "" {
		return tx.QueryContext(ctx,
			"SELECT id, name, description, created_at FROM application_components WHERE tenant_id = $1 AND is_deleted = FALSE ORDER BY LOWER(name) ASC, id ASC LIMIT $2",
			query.tenantID, query.limit,
		)
	}
	return tx.QueryContext(ctx,
		"SELECT id, name, description, created_at FROM application_components WHERE tenant_id = $1 AND is_deleted = FALSE AND (LOWER(name) > LOWER($2) OR (LOWER(name) = LOWER($2) AND id > $3)) ORDER BY LOWER(name) ASC, id ASC LIMIT $4",
		query.tenantID, query.afterName, query.afterCursor, query.limit,
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

func (rm *ApplicationComponentReadModel) AddExpert(ctx context.Context, info ExpertInfo) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO application_component_experts (component_id, tenant_id, expert_name, expert_role, contact_info, added_at) VALUES ($1, $2, $3, $4, $5, $6)",
		info.ComponentID, tenantID.Value(), info.Name, info.Role, info.Contact, info.AddedAt,
	)
	return err
}

func (rm *ApplicationComponentReadModel) RemoveExpert(ctx context.Context, info ExpertInfo) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM application_component_experts WHERE tenant_id = $1 AND component_id = $2 AND expert_name = $3 AND expert_role = $4 AND contact_info = $5",
		tenantID.Value(), info.ComponentID, info.Name, info.Role, info.Contact,
	)
	return err
}

func (rm *ApplicationComponentReadModel) GetDistinctExpertRoles(ctx context.Context) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var roles []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT DISTINCT expert_role FROM application_component_experts WHERE tenant_id = $1 ORDER BY expert_role",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var role string
			if err := rows.Scan(&role); err != nil {
				return err
			}
			roles = append(roles, role)
		}
		return rows.Err()
	})
	return roles, err
}

func (rm *ApplicationComponentReadModel) fetchExperts(ctx context.Context, tx *sql.Tx, tenantID, componentID string) ([]ExpertDTO, error) {
	rows, err := tx.QueryContext(ctx,
		"SELECT expert_name, expert_role, contact_info, added_at FROM application_component_experts WHERE tenant_id = $1 AND component_id = $2",
		tenantID, componentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var experts []ExpertDTO
	for rows.Next() {
		var expert ExpertDTO
		if err := rows.Scan(&expert.Name, &expert.Role, &expert.Contact, &expert.AddedAt); err != nil {
			return nil, err
		}
		experts = append(experts, expert)
	}
	return experts, rows.Err()
}

func (rm *ApplicationComponentReadModel) loadExpertsForComponents(ctx context.Context, tx *sql.Tx, tenantID string, components []ApplicationComponentDTO) error {
	if len(components) == 0 {
		return nil
	}

	componentIDs := make([]string, len(components))
	componentIndex := make(map[string]int)
	for i, c := range components {
		componentIDs[i] = c.ID
		componentIndex[c.ID] = i
	}

	rows, err := tx.QueryContext(ctx,
		"SELECT component_id, expert_name, expert_role, contact_info, added_at FROM application_component_experts WHERE tenant_id = $1 AND component_id = ANY($2)",
		tenantID, pq.Array(componentIDs),
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var componentID string
		var expert ExpertDTO
		if err := rows.Scan(&componentID, &expert.Name, &expert.Role, &expert.Contact, &expert.AddedAt); err != nil {
			return err
		}
		if idx, ok := componentIndex[componentID]; ok {
			components[idx].Experts = append(components[idx].Experts, expert)
		}
	}
	return rows.Err()
}
