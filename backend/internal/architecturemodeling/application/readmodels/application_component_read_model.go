package readmodels

import (
	"context"
	"database/sql"
	"time"
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
	db *sql.DB
}

// NewApplicationComponentReadModel creates a new read model
func NewApplicationComponentReadModel(db *sql.DB) *ApplicationComponentReadModel {
	return &ApplicationComponentReadModel{db: db}
}

// InitializeSchema creates the read model table
func (rm *ApplicationComponentReadModel) InitializeSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS application_components (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(500) NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_application_components_name ON application_components(name);
		CREATE INDEX IF NOT EXISTS idx_application_components_created_at ON application_components(created_at);
	`

	_, err := rm.db.Exec(schema)
	return err
}

// Insert adds a new component to the read model
func (rm *ApplicationComponentReadModel) Insert(ctx context.Context, dto ApplicationComponentDTO) error {
	_, err := rm.db.ExecContext(ctx,
		"INSERT INTO application_components (id, name, description, created_at) VALUES ($1, $2, $3, $4)",
		dto.ID, dto.Name, dto.Description, dto.CreatedAt,
	)
	return err
}

// Update updates an existing component in the read model
func (rm *ApplicationComponentReadModel) Update(ctx context.Context, id, name, description string) error {
	_, err := rm.db.ExecContext(ctx,
		"UPDATE application_components SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3",
		name, description, id,
	)
	return err
}

// GetByID retrieves a component by ID
func (rm *ApplicationComponentReadModel) GetByID(ctx context.Context, id string) (*ApplicationComponentDTO, error) {
	var dto ApplicationComponentDTO
	err := rm.db.QueryRowContext(ctx,
		"SELECT id, name, description, created_at FROM application_components WHERE id = $1",
		id,
	).Scan(&dto.ID, &dto.Name, &dto.Description, &dto.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &dto, nil
}

// GetAll retrieves all components
func (rm *ApplicationComponentReadModel) GetAll(ctx context.Context) ([]ApplicationComponentDTO, error) {
	rows, err := rm.db.QueryContext(ctx,
		"SELECT id, name, description, created_at FROM application_components ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

// GetAllPaginated retrieves components with cursor-based pagination
func (rm *ApplicationComponentReadModel) GetAllPaginated(ctx context.Context, limit int, afterCursor string, afterTimestamp int64) ([]ApplicationComponentDTO, bool, error) {
	// Query one extra to determine if there are more results
	queryLimit := limit + 1

	var rows *sql.Rows
	var err error

	if afterCursor == "" {
		// No cursor, get first page
		rows, err = rm.db.QueryContext(ctx,
			"SELECT id, name, description, created_at FROM application_components ORDER BY created_at DESC, id DESC LIMIT $1",
			queryLimit,
		)
	} else {
		// Use cursor for pagination
		rows, err = rm.db.QueryContext(ctx,
			"SELECT id, name, description, created_at FROM application_components WHERE created_at < to_timestamp($1) OR (created_at = to_timestamp($1) AND id < $2) ORDER BY created_at DESC, id DESC LIMIT $3",
			afterTimestamp, afterCursor, queryLimit,
		)
	}

	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var components []ApplicationComponentDTO
	for rows.Next() {
		var dto ApplicationComponentDTO
		if err := rows.Scan(&dto.ID, &dto.Name, &dto.Description, &dto.CreatedAt); err != nil {
			return nil, false, err
		}
		components = append(components, dto)
	}

	if err := rows.Err(); err != nil {
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
