package repositories

import (
	"context"
	"database/sql"
)

type Tenant struct {
	ID   string
	Name string
}

type TenantRepository struct {
	db *sql.DB
}

func NewTenantRepository(db *sql.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) GetByID(ctx context.Context, tenantID string) (*Tenant, error) {
	var tenant Tenant

	err := r.db.QueryRowContext(ctx,
		`SELECT id, name FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&tenant.ID, &tenant.Name)

	if err == sql.ErrNoRows {
		return nil, ErrTenantNotFound
	}
	if err != nil {
		return nil, err
	}

	return &tenant, nil
}
