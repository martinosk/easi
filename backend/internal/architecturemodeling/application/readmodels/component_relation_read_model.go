package readmodels

import (
	"context"
	"database/sql"
	"time"
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
	db *sql.DB
}

// NewComponentRelationReadModel creates a new read model
func NewComponentRelationReadModel(db *sql.DB) *ComponentRelationReadModel {
	return &ComponentRelationReadModel{db: db}
}

// InitializeSchema creates the read model table
func (rm *ComponentRelationReadModel) InitializeSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS component_relations (
			id VARCHAR(255) PRIMARY KEY,
			source_component_id VARCHAR(255) NOT NULL,
			target_component_id VARCHAR(255) NOT NULL,
			relation_type VARCHAR(50) NOT NULL,
			name VARCHAR(500),
			description TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_component_relations_source ON component_relations(source_component_id);
		CREATE INDEX IF NOT EXISTS idx_component_relations_target ON component_relations(target_component_id);
		CREATE INDEX IF NOT EXISTS idx_component_relations_type ON component_relations(relation_type);
		CREATE INDEX IF NOT EXISTS idx_component_relations_created_at ON component_relations(created_at);
	`

	_, err := rm.db.Exec(schema)
	return err
}

// Insert adds a new relation to the read model
func (rm *ComponentRelationReadModel) Insert(ctx context.Context, dto ComponentRelationDTO) error {
	_, err := rm.db.ExecContext(ctx,
		"INSERT INTO component_relations (id, source_component_id, target_component_id, relation_type, name, description, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		dto.ID, dto.SourceComponentID, dto.TargetComponentID, dto.RelationType, dto.Name, dto.Description, dto.CreatedAt,
	)
	return err
}

// GetByID retrieves a relation by ID
func (rm *ComponentRelationReadModel) GetByID(ctx context.Context, id string) (*ComponentRelationDTO, error) {
	var dto ComponentRelationDTO
	err := rm.db.QueryRowContext(ctx,
		"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM component_relations WHERE id = $1",
		id,
	).Scan(&dto.ID, &dto.SourceComponentID, &dto.TargetComponentID, &dto.RelationType, &dto.Name, &dto.Description, &dto.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &dto, nil
}

// GetAll retrieves all relations
func (rm *ComponentRelationReadModel) GetAll(ctx context.Context) ([]ComponentRelationDTO, error) {
	rows, err := rm.db.QueryContext(ctx,
		"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM component_relations ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relations []ComponentRelationDTO
	for rows.Next() {
		var dto ComponentRelationDTO
		if err := rows.Scan(&dto.ID, &dto.SourceComponentID, &dto.TargetComponentID, &dto.RelationType, &dto.Name, &dto.Description, &dto.CreatedAt); err != nil {
			return nil, err
		}
		relations = append(relations, dto)
	}

	return relations, rows.Err()
}

// GetAllPaginated retrieves relations with cursor-based pagination
func (rm *ComponentRelationReadModel) GetAllPaginated(ctx context.Context, limit int, afterCursor string, afterTimestamp int64) ([]ComponentRelationDTO, bool, error) {
	queryLimit := limit + 1

	var rows *sql.Rows
	var err error

	if afterCursor == "" {
		rows, err = rm.db.QueryContext(ctx,
			"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM component_relations ORDER BY created_at DESC, id DESC LIMIT $1",
			queryLimit,
		)
	} else {
		rows, err = rm.db.QueryContext(ctx,
			"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM component_relations WHERE created_at < to_timestamp($1) OR (created_at = to_timestamp($1) AND id < $2) ORDER BY created_at DESC, id DESC LIMIT $3",
			afterTimestamp, afterCursor, queryLimit,
		)
	}

	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var relations []ComponentRelationDTO
	for rows.Next() {
		var dto ComponentRelationDTO
		if err := rows.Scan(&dto.ID, &dto.SourceComponentID, &dto.TargetComponentID, &dto.RelationType, &dto.Name, &dto.Description, &dto.CreatedAt); err != nil {
			return nil, false, err
		}
		relations = append(relations, dto)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(relations) > limit
	if hasMore {
		relations = relations[:limit]
	}

	return relations, hasMore, nil
}

// GetBySourceID retrieves all relations where component is the source
func (rm *ComponentRelationReadModel) GetBySourceID(ctx context.Context, componentID string) ([]ComponentRelationDTO, error) {
	rows, err := rm.db.QueryContext(ctx,
		"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM component_relations WHERE source_component_id = $1 ORDER BY created_at DESC",
		componentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relations []ComponentRelationDTO
	for rows.Next() {
		var dto ComponentRelationDTO
		if err := rows.Scan(&dto.ID, &dto.SourceComponentID, &dto.TargetComponentID, &dto.RelationType, &dto.Name, &dto.Description, &dto.CreatedAt); err != nil {
			return nil, err
		}
		relations = append(relations, dto)
	}

	return relations, rows.Err()
}

// GetByTargetID retrieves all relations where component is the target
func (rm *ComponentRelationReadModel) GetByTargetID(ctx context.Context, componentID string) ([]ComponentRelationDTO, error) {
	rows, err := rm.db.QueryContext(ctx,
		"SELECT id, source_component_id, target_component_id, relation_type, name, description, created_at FROM component_relations WHERE target_component_id = $1 ORDER BY created_at DESC",
		componentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relations []ComponentRelationDTO
	for rows.Next() {
		var dto ComponentRelationDTO
		if err := rows.Scan(&dto.ID, &dto.SourceComponentID, &dto.TargetComponentID, &dto.RelationType, &dto.Name, &dto.Description, &dto.CreatedAt); err != nil {
			return nil, err
		}
		relations = append(relations, dto)
	}

	return relations, rows.Err()
}
