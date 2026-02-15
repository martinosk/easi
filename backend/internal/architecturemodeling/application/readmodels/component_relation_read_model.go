package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

// ComponentRelationDTO represents the read model for component relations
type ComponentRelationDTO struct {
	ID                string      `json:"id"`
	SourceComponentID string      `json:"sourceComponentId"`
	TargetComponentID string      `json:"targetComponentId"`
	RelationType      string      `json:"relationType"`
	Name              string      `json:"name,omitempty"`
	Description       string      `json:"description,omitempty"`
	CreatedAt         time.Time   `json:"createdAt"`
	Links             types.Links `json:"_links,omitempty"`
}

// ComponentRelationReadModel handles queries for component relations
type ComponentRelationReadModel struct {
	db *database.TenantAwareDB
}

// NewComponentRelationReadModel creates a new read model
func NewComponentRelationReadModel(db *database.TenantAwareDB) *ComponentRelationReadModel {
	return &ComponentRelationReadModel{db: db}
}

// Insert adds a new relation to the read model
func (rm *ComponentRelationReadModel) Insert(ctx context.Context, dto ComponentRelationDTO) error {
	// Extract tenant from context - infrastructure concern
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"INSERT INTO architecturemodeling.component_relations (id, tenant_id, source_component_id, target_component_id, relation_type, name, description, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		dto.ID, tenantID.Value(), dto.SourceComponentID, dto.TargetComponentID, dto.RelationType, dto.Name, dto.Description, dto.CreatedAt,
	)
	return err
}

// Update updates an existing relation in the read model
// RLS policies ensure we can only update our tenant's rows
func (rm *ComponentRelationReadModel) Update(ctx context.Context, id, name, description string) error {
	// Extract tenant for defense-in-depth filtering
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE architecturemodeling.component_relations SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $3 AND id = $4",
		name, description, tenantID.Value(), id,
	)
	return err
}

func (rm *ComponentRelationReadModel) MarkAsDeleted(ctx context.Context, id string, deletedAt time.Time) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}

	_, err = rm.db.ExecContext(ctx,
		"UPDATE architecturemodeling.component_relations SET is_deleted = TRUE, deleted_at = $1 WHERE tenant_id = $2 AND id = $3",
		deletedAt, tenantID.Value(), id,
	)
	return err
}

// GetByID retrieves a relation by ID
// RLS policies automatically filter by tenant, but we add explicit filter for defense-in-depth
func (rm *ComponentRelationReadModel) GetByID(ctx context.Context, id string) (*ComponentRelationDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var dto ComponentRelationDTO
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		var name, description sql.NullString
		err := tx.QueryRowContext(ctx,
			"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM architecturemodeling.component_relations WHERE tenant_id = $1 AND id = $2 AND is_deleted = FALSE",
			tenantID.Value(), id,
		).Scan(&dto.ID, &dto.SourceComponentID, &dto.TargetComponentID, &dto.RelationType, &name, &description, &dto.CreatedAt)

		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
		if err == nil {
			dto.Name = name.String
			dto.Description = description.String
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

// GetAll retrieves all relations for the current tenant
// RLS policies automatically filter, but we add explicit filter for defense-in-depth
func (rm *ComponentRelationReadModel) GetAll(ctx context.Context) ([]ComponentRelationDTO, error) {
	return rm.queryRelations(ctx, "SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM architecturemodeling.component_relations WHERE tenant_id = $1 AND is_deleted = FALSE ORDER BY created_at DESC")
}

type paginationParams struct {
	tenantID       string
	afterCursor    string
	afterTimestamp int64
	limit          int
}

// GetAllPaginated retrieves relations with cursor-based pagination for the current tenant
func (rm *ComponentRelationReadModel) GetAllPaginated(ctx context.Context, limit int, afterCursor string, afterTimestamp int64) ([]ComponentRelationDTO, bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, false, err
	}

	params := paginationParams{
		tenantID:       tenantID.Value(),
		afterCursor:    afterCursor,
		afterTimestamp: afterTimestamp,
		limit:          limit + 1,
	}

	var relations []ComponentRelationDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := rm.selectPaginatedRows(ctx, tx, params)
		if err != nil {
			return err
		}
		defer rows.Close()

		relations, err = rm.collectRelations(rows)
		return err
	})

	if err != nil {
		return nil, false, err
	}

	return rm.extractPageResults(relations, limit)
}

func (rm *ComponentRelationReadModel) selectPaginatedRows(ctx context.Context, tx *sql.Tx, params paginationParams) (*sql.Rows, error) {
	if params.afterCursor == "" {
		return tx.QueryContext(ctx,
			"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM architecturemodeling.component_relations WHERE tenant_id = $1 AND is_deleted = FALSE ORDER BY created_at DESC, id DESC LIMIT $2",
			params.tenantID, params.limit,
		)
	}
	return tx.QueryContext(ctx,
		"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM architecturemodeling.component_relations WHERE tenant_id = $1 AND is_deleted = FALSE AND (created_at < to_timestamp($2) OR (created_at = to_timestamp($2) AND id < $3)) ORDER BY created_at DESC, id DESC LIMIT $4",
		params.tenantID, params.afterTimestamp, params.afterCursor, params.limit,
	)
}

func (rm *ComponentRelationReadModel) extractPageResults(relations []ComponentRelationDTO, limit int) ([]ComponentRelationDTO, bool, error) {
	hasMore := len(relations) > limit
	if hasMore {
		relations = relations[:limit]
	}
	return relations, hasMore, nil
}

// GetBySourceID retrieves all relations where component is the source for the current tenant
func (rm *ComponentRelationReadModel) GetBySourceID(ctx context.Context, componentID string) ([]ComponentRelationDTO, error) {
	return rm.queryRelationsWithParam(ctx, "SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM architecturemodeling.component_relations WHERE tenant_id = $1 AND source_component_id = $2 AND is_deleted = FALSE ORDER BY created_at DESC", componentID)
}

// GetByTargetID retrieves all relations where component is the target for the current tenant
func (rm *ComponentRelationReadModel) GetByTargetID(ctx context.Context, componentID string) ([]ComponentRelationDTO, error) {
	return rm.queryRelationsWithParam(ctx, "SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM architecturemodeling.component_relations WHERE tenant_id = $1 AND target_component_id = $2 AND is_deleted = FALSE ORDER BY created_at DESC", componentID)
}

func (rm *ComponentRelationReadModel) queryRelations(ctx context.Context, query string) ([]ComponentRelationDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	return rm.executeRelationQuery(ctx, query, tenantID.Value())
}

func (rm *ComponentRelationReadModel) queryRelationsWithParam(ctx context.Context, query string, param string) ([]ComponentRelationDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	return rm.executeRelationQuery(ctx, query, tenantID.Value(), param)
}

func (rm *ComponentRelationReadModel) executeRelationQuery(ctx context.Context, query string, args ...interface{}) ([]ComponentRelationDTO, error) {
	var relations []ComponentRelationDTO
	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		relations, err = rm.collectRelations(rows)
		return err
	})

	return relations, err
}

func (rm *ComponentRelationReadModel) collectRelations(rows *sql.Rows) ([]ComponentRelationDTO, error) {
	var relations []ComponentRelationDTO
	for rows.Next() {
		dto, err := rm.scanRelationRow(rows)
		if err != nil {
			return nil, err
		}
		relations = append(relations, dto)
	}
	return relations, rows.Err()
}

func (rm *ComponentRelationReadModel) scanRelationRow(rows *sql.Rows) (ComponentRelationDTO, error) {
	var dto ComponentRelationDTO
	var name, description sql.NullString
	err := rows.Scan(&dto.ID, &dto.SourceComponentID, &dto.TargetComponentID, &dto.RelationType, &name, &description, &dto.CreatedAt)
	if err != nil {
		return ComponentRelationDTO{}, err
	}
	dto.Name = name.String
	dto.Description = description.String
	return dto, nil
}
