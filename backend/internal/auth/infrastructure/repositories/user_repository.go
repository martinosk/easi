package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID     uuid.UUID
	Email  string
	Name   string
	Role   string
	Status string
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByEmail(ctx context.Context, tenantID, email string) (*User, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Close()

	escapedTenantID := strings.ReplaceAll(tenantID, "'", "''")
	_, err = conn.ExecContext(ctx, fmt.Sprintf("SET app.current_tenant = '%s'", escapedTenantID))
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	var user User
	var name sql.NullString

	err = conn.QueryRowContext(ctx,
		`SELECT id, email, name, role, status
		 FROM users
		 WHERE tenant_id = $1 AND email = $2`,
		tenantID, email,
	).Scan(&user.ID, &user.Email, &name, &user.Role, &user.Status)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if name.Valid {
		user.Name = name.String
	}

	return &user, nil
}
