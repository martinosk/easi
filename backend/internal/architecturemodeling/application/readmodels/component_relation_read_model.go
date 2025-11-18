package readmodels

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

// ComponentRelationDTO represents the read model for component relations
type ComponentRelationDTO struct {
	ID                string            `json:"id"`
	SourceComponentID string            `json:"sourceComponentId"`
	TargetComponentID string            `json:"targetComponentId"`
	RelationType      string            `json:"relationType"`
	Name              string            `json:"name,omitempty"`
	Description       string            `json:"description,omitempty"`
	CreatedAt         time.Time         `json:"createdAt"`
	Links             map[string]string `json:"_links,omitempty"`
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
		"INSERT INTO component_relations (id, tenant_id, source_component_id, target_component_id, relation_type, name, description, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
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
		"UPDATE component_relations SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = $3 AND id = $4",
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
		"UPDATE component_relations SET is_deleted = TRUE, deleted_at = $1 WHERE tenant_id = $2 AND id = $3",
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
			"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM component_relations WHERE tenant_id = $1 AND id = $2 AND is_deleted = FALSE",
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
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var relations []ComponentRelationDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM component_relations WHERE tenant_id = $1 AND is_deleted = FALSE ORDER BY created_at DESC",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto ComponentRelationDTO
			var name, description sql.NullString
			if err := rows.Scan(&dto.ID, &dto.SourceComponentID, &dto.TargetComponentID, &dto.RelationType, &name, &description, &dto.CreatedAt); err != nil {
				return err
			}
			dto.Name = name.String
			dto.Description = description.String
			relations = append(relations, dto)
		}

		return rows.Err()
	})

	return relations, err
}

// GetAllPaginated retrieves relations with cursor-based pagination for the current tenant
func (rm *ComponentRelationReadModel) GetAllPaginated(ctx context.Context, limit int, afterCursor string, afterTimestamp int64) ([]ComponentRelationDTO, bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, false, err
	}

	// Query one extra to determine if there are more results
	queryLimit := limit + 1

	var relations []ComponentRelationDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		var rows *sql.Rows
		var err error

		if afterCursor == "" {
			// No cursor, get first page
			rows, err = tx.QueryContext(ctx,
				"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM component_relations WHERE tenant_id = $1 AND is_deleted = FALSE ORDER BY created_at DESC, id DESC LIMIT $2",
				tenantID.Value(), queryLimit,
			)
		} else {
			// Use cursor for pagination
			rows, err = tx.QueryContext(ctx,
				"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM component_relations WHERE tenant_id = $1 AND is_deleted = FALSE AND (created_at < to_timestamp($2) OR (created_at = to_timestamp($2) AND id < $3)) ORDER BY created_at DESC, id DESC LIMIT $4",
				tenantID.Value(), afterTimestamp, afterCursor, queryLimit,
			)
		}

		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto ComponentRelationDTO
			var name, description sql.NullString
			if err := rows.Scan(&dto.ID, &dto.SourceComponentID, &dto.TargetComponentID, &dto.RelationType, &name, &description, &dto.CreatedAt); err != nil {
				return err
			}
			dto.Name = name.String
			dto.Description = description.String
			relations = append(relations, dto)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, false, err
	}

	// Check if there are more results
	hasMore := len(relations) > limit
	if hasMore {
		// Remove the extra item
		relations = relations[:limit]
	}

	return relations, hasMore, nil
}

// GetBySourceID retrieves all relations where component is the source for the current tenant
func (rm *ComponentRelationReadModel) GetBySourceID(ctx context.Context, componentID string) ([]ComponentRelationDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var relations []ComponentRelationDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM component_relations WHERE tenant_id = $1 AND source_component_id = $2 AND is_deleted = FALSE ORDER BY created_at DESC",
			tenantID.Value(), componentID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto ComponentRelationDTO
			var name, description sql.NullString
			if err := rows.Scan(&dto.ID, &dto.SourceComponentID, &dto.TargetComponentID, &dto.RelationType, &name, &description, &dto.CreatedAt); err != nil {
				return err
			}
			dto.Name = name.String
			dto.Description = description.String
			relations = append(relations, dto)
		}

		return rows.Err()
	})

	return relations, err
}

// GetByTargetID retrieves all relations where component is the target for the current tenant
func (rm *ComponentRelationReadModel) GetByTargetID(ctx context.Context, componentID string) ([]ComponentRelationDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	var relations []ComponentRelationDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM component_relations WHERE tenant_id = $1 AND target_component_id = $2 AND is_deleted = FALSE ORDER BY created_at DESC",
			tenantID.Value(), componentID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var dto ComponentRelationDTO
			var name, description sql.NullString
			if err := rows.Scan(&dto.ID, &dto.SourceComponentID, &dto.TargetComponentID, &dto.RelationType, &name, &description, &dto.CreatedAt); err != nil {
				return err
			}
			dto.Name = name.String
			dto.Description = description.String
			relations = append(relations, dto)
		}

		return rows.Err()
	})

	return relations, err
}
