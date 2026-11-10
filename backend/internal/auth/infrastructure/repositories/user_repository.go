package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, "SELECT set_config('app.current_tenant', $1, true)", tenantID); err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	var user User
	var name sql.NullString

	err = tx.QueryRowContext(ctx,
		`SELECT id, email, name, role, status
		 FROM auth.users
		 WHERE tenant_id = $1 AND email = $2`,
		tenantID, email,
	).Scan(&user.ID, &user.Email, &name, &user.Role, &user.Status)

	if errors.Is(err, sql.ErrNoRows) {
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
