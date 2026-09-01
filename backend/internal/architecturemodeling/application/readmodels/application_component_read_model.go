package readmodels

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type ApplicationComponentDTO struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Description    string              `json:"description,omitempty"`
	CreatedAt      time.Time           `json:"createdAt"`
	OwnershipState string              `json:"ownershipState"`
	Owner          *ComponentOwnerDTO  `json:"owner,omitempty"`
	Experts        []ExpertDTO         `json:"experts,omitempty"`
	Links          types.Links         `json:"_links,omitempty"`
	XRelated       []types.RelatedLink `json:"-"`
}

type ComponentOwnerDTO struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

const componentSelect = `SELECT c.id, c.name, c.description, c.created_at, c.ownership_state, c.owner_kind, c.owner_id,
	CASE
		WHEN c.owner_kind = 'user' THEN (SELECT COALESCE(NULLIF(u.name, ''), u.email) FROM architecturemodeling.user_names u WHERE u.tenant_id = c.tenant_id AND u.user_id = c.owner_id)
		WHEN c.owner_kind = 'team' THEN (SELECT t.name FROM architecturemodeling.internal_teams t WHERE t.tenant_id = c.tenant_id AND t.id = c.owner_id AND t.is_deleted = FALSE)
	END AS owner_name
	FROM architecturemodeling.application_components c`

func (d ApplicationComponentDTO) MarshalJSON() ([]byte, error) {
	type alias ApplicationComponentDTO
	base, err := json.Marshal(alias(d))
	if err != nil {
		return nil, err
	}
	return types.SpliceXRelated(base, d.XRelated)
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
		return fmt.Errorf("resolve tenant for insert application component %s: %w", dto.ID, err)
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM architecturemodeling.application_components WHERE tenant_id = $1 AND id = $2",
		tenantID.Value(), dto.ID,
	)
	if err != nil {
		return fmt.Errorf("delete existing application component %s before insert: %w", dto.ID, err)
	}

	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO architecturemodeling.application_components
		(id, tenant_id, name, description, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		dto.ID, tenantID.Value(), dto.Name, dto.Description, dto.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert application component %s for tenant %s: %w", dto.ID, tenantID.Value(), err)
	}
	return nil
}

func (rm *ApplicationComponentReadModel) execByID(ctx context.Context, query string, id string, extraArgs ...any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("resolve tenant for application component %s: %w", id, err)
	}
	args := append([]any{tenantID.Value(), id}, extraArgs...)
	_, err = rm.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("application component %s mutation for tenant %s: %w", id, tenantID.Value(), err)
	}
	return nil
}

func (rm *ApplicationComponentReadModel) Update(ctx context.Context, id, name, description string) error {
	return rm.execByID(ctx,
		"UPDATE architecturemodeling.application_components SET name = $3, description = $4, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $1 AND id = $2",
		id, name, description,
	)
}

func (rm *ApplicationComponentReadModel) MarkAsDeleted(ctx context.Context, id string, deletedAt time.Time) error {
	return rm.execByID(ctx,
		"UPDATE architecturemodeling.application_components SET is_deleted = TRUE, deleted_at = $3 WHERE tenant_id = $1 AND id = $2",
		id, deletedAt,
	)
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
		row := tx.QueryRowContext(ctx,
			componentSelect+" WHERE c.tenant_id = $1 AND c.id = $2 AND c.is_deleted = FALSE",
			tenantID.Value(), id,
		)
		var err error
		dto, err = scanComponent(row.Scan)

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

func (rm *ApplicationComponentReadModel) Count(ctx context.Context) (int, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return 0, fmt.Errorf("resolve tenant for count application components: %w", err)
	}

	var count int
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM architecturemodeling.application_components WHERE tenant_id = $1 AND is_deleted = FALSE",
			tenantID.Value(),
		).Scan(&count)
	})
	if err != nil {
		return 0, fmt.Errorf("count application components for tenant %s: %w", tenantID.Value(), err)
	}

	return count, nil
}

// GetAll retrieves all components for the current tenant
// RLS policies automatically filter, but we add explicit filter for defense-in-depth
func (rm *ApplicationComponentReadModel) GetAll(ctx context.Context) ([]ApplicationComponentDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	return rm.queryComponents(ctx, tenantID.Value(),
		componentSelect+" WHERE c.tenant_id = $1 AND c.is_deleted = FALSE ORDER BY LOWER(c.name) ASC, c.id ASC",
		tenantID.Value(),
	)
}

func (rm *ApplicationComponentReadModel) GetByIDs(ctx context.Context, ids []string) ([]ApplicationComponentDTO, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	return rm.queryComponents(ctx, tenantID.Value(),
		componentSelect+" WHERE c.tenant_id = $1 AND c.id = ANY($2) AND c.is_deleted = FALSE ORDER BY LOWER(c.name) ASC, c.id ASC",
		tenantID.Value(), pq.Array(ids),
	)
}

func (rm *ApplicationComponentReadModel) queryComponents(ctx context.Context, tenantID, query string, args ...any) ([]ApplicationComponentDTO, error) {
	var components []ApplicationComponentDTO
	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		components, err = rm.scanComponents(rows)
		if err != nil {
			return err
		}
		return rm.loadExpertsForComponents(ctx, tx, tenantID, components)
	})

	return components, err
}

type ComponentQuery struct {
	Limit       int
	AfterCursor string
	AfterName   string
	NameFilter  string
}

type resolvedQuery struct {
	tenantID    string
	afterCursor string
	afterName   string
	nameFilter  string
	limit       int
}

func (rm *ApplicationComponentReadModel) GetAllPaginated(ctx context.Context, q ComponentQuery) ([]ApplicationComponentDTO, bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, false, err
	}

	query := resolvedQuery{
		tenantID:    tenantID.Value(),
		afterCursor: q.AfterCursor,
		afterName:   q.AfterName,
		nameFilter:  q.NameFilter,
		limit:       q.Limit + 1,
	}

	var components []ApplicationComponentDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := rm.queryPaginatedComponents(ctx, tx, query)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		components, err = rm.scanComponents(rows)
		if err != nil {
			return err
		}

		return rm.loadExpertsForComponents(ctx, tx, tenantID.Value(), components)
	})

	if err != nil {
		return nil, false, err
	}

	return rm.trimAndCheckMore(components, q.Limit)
}

func (rm *ApplicationComponentReadModel) queryPaginatedComponents(ctx context.Context, tx *sql.Tx, query resolvedQuery) (*sql.Rows, error) {
	if query.nameFilter != "" && query.afterCursor != "" {
		return tx.QueryContext(ctx,
			componentSelect+" WHERE c.tenant_id = $1 AND c.is_deleted = FALSE AND LOWER(c.name) LIKE '%' || LOWER($2) || '%' AND (LOWER(c.name) > LOWER($3) OR (LOWER(c.name) = LOWER($3) AND c.id > $4)) ORDER BY LOWER(c.name) ASC, c.id ASC LIMIT $5",
			query.tenantID, query.nameFilter, query.afterName, query.afterCursor, query.limit,
		)
	}
	if query.nameFilter != "" {
		return tx.QueryContext(ctx,
			componentSelect+" WHERE c.tenant_id = $1 AND c.is_deleted = FALSE AND LOWER(c.name) LIKE '%' || LOWER($2) || '%' ORDER BY LOWER(c.name) ASC, c.id ASC LIMIT $3",
			query.tenantID, query.nameFilter, query.limit,
		)
	}
	if query.afterCursor == "" {
		return tx.QueryContext(ctx,
			componentSelect+" WHERE c.tenant_id = $1 AND c.is_deleted = FALSE ORDER BY LOWER(c.name) ASC, c.id ASC LIMIT $2",
			query.tenantID, query.limit,
		)
	}
	return tx.QueryContext(ctx,
		componentSelect+" WHERE c.tenant_id = $1 AND c.is_deleted = FALSE AND (LOWER(c.name) > LOWER($2) OR (LOWER(c.name) = LOWER($2) AND c.id > $3)) ORDER BY LOWER(c.name) ASC, c.id ASC LIMIT $4",
		query.tenantID, query.afterName, query.afterCursor, query.limit,
	)
}

func scanComponent(scan func(dest ...any) error) (ApplicationComponentDTO, error) {
	var dto ApplicationComponentDTO
	var ownerKind, ownerID, ownerName sql.NullString
	if err := scan(&dto.ID, &dto.Name, &dto.Description, &dto.CreatedAt, &dto.OwnershipState, &ownerKind, &ownerID, &ownerName); err != nil {
		return dto, err
	}
	if ownerKind.Valid && ownerID.Valid {
		dto.Owner = &ComponentOwnerDTO{Kind: ownerKind.String, ID: ownerID.String, Name: ownerName.String}
	}
	return dto, nil
}

func (rm *ApplicationComponentReadModel) scanComponents(rows *sql.Rows) ([]ApplicationComponentDTO, error) {
	var components []ApplicationComponentDTO
	for rows.Next() {
		dto, err := scanComponent(rows.Scan)
		if err != nil {
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
		return fmt.Errorf("resolve tenant for add expert to application component %s: %w", info.ComponentID, err)
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM architecturemodeling.application_component_experts WHERE tenant_id = $1 AND component_id = $2 AND expert_name = $3 AND expert_role = $4",
		tenantID.Value(), info.ComponentID, info.Name, info.Role,
	)
	if err != nil {
		return fmt.Errorf("delete existing expert %s/%s for application component %s: %w", info.Name, info.Role, info.ComponentID, err)
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO architecturemodeling.application_component_experts (component_id, tenant_id, expert_name, expert_role, contact_info, added_at) VALUES ($1, $2, $3, $4, $5, $6)",
		info.ComponentID, tenantID.Value(), info.Name, info.Role, info.Contact, info.AddedAt,
	)
	if err != nil {
		return fmt.Errorf("insert expert %s/%s for application component %s tenant %s: %w", info.Name, info.Role, info.ComponentID, tenantID.Value(), err)
	}
	return nil
}

func (rm *ApplicationComponentReadModel) RemoveExpert(ctx context.Context, info ExpertInfo) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("resolve tenant for remove expert from application component %s: %w", info.ComponentID, err)
	}

	_, err = rm.db.ExecContext(ctx,
		"DELETE FROM architecturemodeling.application_component_experts WHERE tenant_id = $1 AND component_id = $2 AND expert_name = $3 AND expert_role = $4 AND contact_info = $5",
		tenantID.Value(), info.ComponentID, info.Name, info.Role, info.Contact,
	)
	if err != nil {
		return fmt.Errorf("remove expert %s/%s from application component %s tenant %s: %w", info.Name, info.Role, info.ComponentID, tenantID.Value(), err)
	}
	return nil
}

func (rm *ApplicationComponentReadModel) GetDistinctExpertRoles(ctx context.Context) ([]string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var roles []string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT DISTINCT expert_role FROM architecturemodeling.application_component_experts WHERE tenant_id = $1 ORDER BY expert_role",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

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
		"SELECT expert_name, expert_role, contact_info, added_at FROM architecturemodeling.application_component_experts WHERE tenant_id = $1 AND component_id = $2",
		tenantID, componentID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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
		"SELECT component_id, expert_name, expert_role, contact_info, added_at FROM architecturemodeling.application_component_experts WHERE tenant_id = $1 AND component_id = ANY($2)",
		tenantID, pq.Array(componentIDs),
	)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

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
