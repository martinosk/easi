package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantAlreadyExists = errors.New("tenant already exists")
	ErrDomainAlreadyExists = errors.New("domain already registered to another tenant")
)

type TenantRepository struct {
	db *sql.DB
}

func NewTenantRepository(db *sql.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

type TenantRecord struct {
	ID              string
	Name            string
	Status          string
	Domains         []string
	DiscoveryURL    string
	ClientID        string
	AuthMethod      string
	Scopes          string
	FirstAdminEmail string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (r *TenantRepository) Create(ctx context.Context, record TenantRecord) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := ensureTenantAbsent(ctx, tx, record.ID); err != nil {
		return err
	}
	if err := ensureDomainsAvailable(ctx, tx, record.Domains); err != nil {
		return err
	}
	if err := insertTenant(ctx, tx, record); err != nil {
		return err
	}

	return tx.Commit()
}

func ensureTenantAbsent(ctx context.Context, tx *sql.Tx, id string) error {
	var exists bool
	err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM auth.tenants WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrTenantAlreadyExists
	}
	return nil
}

func ensureDomainsAvailable(ctx context.Context, tx *sql.Tx, domains []string) error {
	for _, domain := range domains {
		var existingTenantID string
		err := tx.QueryRowContext(ctx,
			"SELECT tenant_id FROM auth.tenant_domains WHERE domain = $1",
			domain,
		).Scan(&existingTenantID)
		if err == nil {
			return ErrDomainAlreadyExists
		}
		if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

func insertTenant(ctx context.Context, tx *sql.Tx, record TenantRecord) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO auth.tenants (id, name, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		record.ID, record.Name, record.Status, record.CreatedAt, record.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if err := insertTenantDomains(ctx, tx, record); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO auth.tenant_oidc_configs (tenant_id, discovery_url, client_id, auth_method, scopes, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		record.ID, record.DiscoveryURL, record.ClientID, record.AuthMethod, record.Scopes, record.CreatedAt, record.UpdatedAt,
	)
	return err
}

func insertTenantDomains(ctx context.Context, tx *sql.Tx, record TenantRecord) error {
	for _, domain := range record.Domains {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO auth.tenant_domains (domain, tenant_id, created_at)
			 VALUES ($1, $2, $3)`,
			domain, record.ID, record.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *TenantRepository) GetByID(ctx context.Context, id string) (*TenantRecord, error) {
	record := &TenantRecord{}

	err := r.db.QueryRowContext(ctx,
		`SELECT t.id, t.name, t.status, t.created_at, t.updated_at,
		        oc.discovery_url, oc.client_id, oc.auth_method, oc.scopes
		 FROM auth.tenants t
		 JOIN auth.tenant_oidc_configs oc ON t.id = oc.tenant_id
		 WHERE t.id = $1`,
		id,
	).Scan(
		&record.ID, &record.Name, &record.Status, &record.CreatedAt, &record.UpdatedAt,
		&record.DiscoveryURL, &record.ClientID, &record.AuthMethod, &record.Scopes,
	)
	if err == sql.ErrNoRows {
		return nil, ErrTenantNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx,
		"SELECT domain FROM auth.tenant_domains WHERE tenant_id = $1",
		id,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return nil, err
		}
		record.Domains = append(record.Domains, domain)
	}

	return record, nil
}

func (r *TenantRepository) List(ctx context.Context, status string, domain string) ([]*TenantRecord, error) {
	query, args := buildListQuery(status, domain)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var records []*TenantRecord
	for rows.Next() {
		record := &TenantRecord{}
		if err := rows.Scan(&record.ID, &record.Name, &record.Status, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return nil, err
		}
		if err := r.loadDomains(ctx, record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

func buildListQuery(status string, domain string) (string, []interface{}) {
	query := `SELECT t.id, t.name, t.status, t.created_at, t.updated_at
			  FROM auth.tenants t`
	var args []interface{}
	var conditions []string

	if status != "" {
		args = append(args, status)
		conditions = append(conditions, "t.status = $"+string(rune('0'+len(args))))
	}

	if domain != "" {
		query += " JOIN auth.tenant_domains td ON t.id = td.tenant_id"
		args = append(args, domain)
		conditions = append(conditions, "td.domain = $"+string(rune('0'+len(args))))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	return query + " ORDER BY t.created_at DESC", args
}

func (r *TenantRepository) loadDomains(ctx context.Context, record *TenantRecord) error {
	domainRows, err := r.db.QueryContext(ctx,
		"SELECT domain FROM auth.tenant_domains WHERE tenant_id = $1",
		record.ID,
	)
	if err != nil {
		return err
	}
	defer func() { _ = domainRows.Close() }()

	for domainRows.Next() {
		var d string
		if err := domainRows.Scan(&d); err != nil {
			return err
		}
		record.Domains = append(record.Domains, d)
	}
	return nil
}

func (r *TenantRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM auth.tenants WHERE id = $1)",
		id,
	).Scan(&exists)
	return exists, err
}

func (r *TenantRepository) DomainExists(ctx context.Context, domain string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM auth.tenant_domains WHERE domain = $1)",
		domain,
	).Scan(&exists)
	return exists, err
}
