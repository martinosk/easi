package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
		`SELECT tenant_id, name FROM auth.tenant_cache WHERE tenant_id = $1`,
		tenantID,
	).Scan(&tenant.ID, &tenant.Name)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTenantNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load cached tenant %s: %w", tenantID, err)
	}

	return &tenant, nil
}

func (r *TenantRepository) GetDomains(ctx context.Context, tenantID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT domain FROM auth.tenant_domain_cache WHERE tenant_id = $1 ORDER BY domain`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("load cached domains of tenant %s: %w", tenantID, err)
	}
	defer func() { _ = rows.Close() }()

	var domains []string
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return nil, fmt.Errorf("scan cached domain of tenant %s: %w", tenantID, err)
		}
		domains = append(domains, domain)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read cached domains of tenant %s: %w", tenantID, err)
	}

	return domains, nil
}
