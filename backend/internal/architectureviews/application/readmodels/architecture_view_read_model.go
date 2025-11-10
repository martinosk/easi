package readmodels

import (
	"context"
	"database/sql"
	"time"
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
	db *sql.DB
}

// NewArchitectureViewReadModel creates a new read model
func NewArchitectureViewReadModel(db *sql.DB) *ArchitectureViewReadModel {
	return &ArchitectureViewReadModel{db: db}
}

// InitializeSchema creates the read model tables
func (rm *ArchitectureViewReadModel) InitializeSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS architecture_views (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(500) NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS view_component_positions (
			view_id VARCHAR(255) NOT NULL,
			component_id VARCHAR(255) NOT NULL,
			x DOUBLE PRECISION NOT NULL,
			y DOUBLE PRECISION NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (view_id, component_id),
			FOREIGN KEY (view_id) REFERENCES architecture_views(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_architecture_views_name ON architecture_views(name);
		CREATE INDEX IF NOT EXISTS idx_architecture_views_created_at ON architecture_views(created_at);
		CREATE INDEX IF NOT EXISTS idx_view_component_positions_view_id ON view_component_positions(view_id);
	`

	_, err := rm.db.Exec(schema)
	if err != nil {
		return err
	}

	// Add new columns if they don't exist (migration for existing databases)
	migrations := `
		ALTER TABLE architecture_views ADD COLUMN IF NOT EXISTS is_default BOOLEAN NOT NULL DEFAULT false;
		ALTER TABLE architecture_views ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT false;
		CREATE INDEX IF NOT EXISTS idx_architecture_views_is_default ON architecture_views(is_default);
	`

	_, err = rm.db.Exec(migrations)
	return err
}

// InsertView adds a new view to the read model
func (rm *ArchitectureViewReadModel) InsertView(ctx context.Context, dto ArchitectureViewDTO) error {
	_, err := rm.db.ExecContext(ctx,
		"INSERT INTO architecture_views (id, name, description, is_default, created_at) VALUES ($1, $2, $3, $4, $5)",
		dto.ID, dto.Name, dto.Description, dto.IsDefault, dto.CreatedAt,
	)
	return err
}

// AddComponent adds a component position to a view
func (rm *ArchitectureViewReadModel) AddComponent(ctx context.Context, viewID, componentID string, x, y float64) error {
	_, err := rm.db.ExecContext(ctx,
		"INSERT INTO view_component_positions (view_id, component_id, x, y, created_at) VALUES ($1, $2, $3, $4, $5)",
		viewID, componentID, x, y, time.Now().UTC(),
	)
	return err
}

// UpdateComponentPosition updates a component's position in a view
func (rm *ArchitectureViewReadModel) UpdateComponentPosition(ctx context.Context, viewID, componentID string, x, y float64) error {
	_, err := rm.db.ExecContext(ctx,
		"UPDATE view_component_positions SET x = $1, y = $2, updated_at = $3 WHERE view_id = $4 AND component_id = $5",
		x, y, time.Now().UTC(), viewID, componentID,
	)
	return err
}

// RemoveComponent removes a component from a view
func (rm *ArchitectureViewReadModel) RemoveComponent(ctx context.Context, viewID, componentID string) error {
	_, err := rm.db.ExecContext(ctx,
		"DELETE FROM view_component_positions WHERE view_id = $1 AND component_id = $2",
		viewID, componentID,
	)
	return err
}

// UpdateViewName updates a view's name
func (rm *ArchitectureViewReadModel) UpdateViewName(ctx context.Context, viewID, newName string) error {
	_, err := rm.db.ExecContext(ctx,
		"UPDATE architecture_views SET name = $1, updated_at = $2 WHERE id = $3",
		newName, time.Now().UTC(), viewID,
	)
	return err
}

// MarkViewAsDeleted marks a view as deleted
func (rm *ArchitectureViewReadModel) MarkViewAsDeleted(ctx context.Context, viewID string) error {
	_, err := rm.db.ExecContext(ctx,
		"UPDATE architecture_views SET is_deleted = true, updated_at = $1 WHERE id = $2",
		time.Now().UTC(), viewID,
	)
	return err
}

// SetViewAsDefault sets a view as the default
func (rm *ArchitectureViewReadModel) SetViewAsDefault(ctx context.Context, viewID string, isDefault bool) error {
	_, err := rm.db.ExecContext(ctx,
		"UPDATE architecture_views SET is_default = $1, updated_at = $2 WHERE id = $3",
		isDefault, time.Now().UTC(), viewID,
	)
	return err
}

// GetDefaultView retrieves the default view
func (rm *ArchitectureViewReadModel) GetDefaultView(ctx context.Context) (*ArchitectureViewDTO, error) {
	var dto ArchitectureViewDTO
	err := rm.db.QueryRowContext(ctx,
		"SELECT id, name, description, is_default, created_at FROM architecture_views WHERE is_default = true AND is_deleted = false LIMIT 1",
	).Scan(&dto.ID, &dto.Name, &dto.Description, &dto.IsDefault, &dto.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Get component positions
	dto.Components, _ = rm.getComponentsForView(ctx, dto.ID)

	return &dto, nil
}

// GetByID retrieves a view by ID with all component positions
func (rm *ArchitectureViewReadModel) GetByID(ctx context.Context, id string) (*ArchitectureViewDTO, error) {
	var dto ArchitectureViewDTO
	err := rm.db.QueryRowContext(ctx,
		"SELECT id, name, description, is_default, created_at FROM architecture_views WHERE id = $1 AND is_deleted = false",
		id,
	).Scan(&dto.ID, &dto.Name, &dto.Description, &dto.IsDefault, &dto.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Get component positions
	rows, err := rm.db.QueryContext(ctx,
		"SELECT component_id, x, y FROM view_component_positions WHERE view_id = $1",
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dto.Components = make([]ComponentPositionDTO, 0)
	for rows.Next() {
		var comp ComponentPositionDTO
		if err := rows.Scan(&comp.ComponentID, &comp.X, &comp.Y); err != nil {
			return nil, err
		}
		dto.Components = append(dto.Components, comp)
	}

	return &dto, rows.Err()
}

// GetAll retrieves all views (excluding deleted ones)
func (rm *ArchitectureViewReadModel) GetAll(ctx context.Context) ([]ArchitectureViewDTO, error) {
	rows, err := rm.db.QueryContext(ctx,
		"SELECT id, name, description, is_default, created_at FROM architecture_views WHERE is_deleted = false ORDER BY is_default DESC, created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []ArchitectureViewDTO
	for rows.Next() {
		var dto ArchitectureViewDTO
		if err := rows.Scan(&dto.ID, &dto.Name, &dto.Description, &dto.IsDefault, &dto.CreatedAt); err != nil {
			return nil, err
		}

		// Get component positions for this view
		dto.Components, _ = rm.getComponentsForView(ctx, dto.ID)
		views = append(views, dto)
	}

	return views, rows.Err()
}

// getComponentsForView is a helper to fetch components for a view
func (rm *ArchitectureViewReadModel) getComponentsForView(ctx context.Context, viewID string) ([]ComponentPositionDTO, error) {
	rows, err := rm.db.QueryContext(ctx,
		"SELECT component_id, x, y FROM view_component_positions WHERE view_id = $1",
		viewID,
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
