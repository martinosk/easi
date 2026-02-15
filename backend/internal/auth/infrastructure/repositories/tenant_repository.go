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
		`SELECT id, name FROM platform.tenants WHERE id = $1`,
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

func (r *TenantRepository) GetDomains(ctx context.Context, tenantID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT domain FROM platform.tenant_domains WHERE tenant_id = $1 ORDER BY domain`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []string
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return nil, err
		}
		domains = append(domains, domain)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return domains, nil
}
