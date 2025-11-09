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
	return err
}

// InsertView adds a new view to the read model
func (rm *ArchitectureViewReadModel) InsertView(ctx context.Context, dto ArchitectureViewDTO) error {
	_, err := rm.db.ExecContext(ctx,
		"INSERT INTO architecture_views (id, name, description, created_at) VALUES ($1, $2, $3, $4)",
		dto.ID, dto.Name, dto.Description, dto.CreatedAt,
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

// GetByID retrieves a view by ID with all component positions
func (rm *ArchitectureViewReadModel) GetByID(ctx context.Context, id string) (*ArchitectureViewDTO, error) {
	var dto ArchitectureViewDTO
	err := rm.db.QueryRowContext(ctx,
		"SELECT id, name, description, created_at FROM architecture_views WHERE id = $1",
		id,
	).Scan(&dto.ID, &dto.Name, &dto.Description, &dto.CreatedAt)

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

// GetAll retrieves all views
func (rm *ArchitectureViewReadModel) GetAll(ctx context.Context) ([]ArchitectureViewDTO, error) {
	rows, err := rm.db.QueryContext(ctx,
		"SELECT id, name, description, created_at FROM architecture_views ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []ArchitectureViewDTO
	for rows.Next() {
		var dto ArchitectureViewDTO
		if err := rows.Scan(&dto.ID, &dto.Name, &dto.Description, &dto.CreatedAt); err != nil {
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
