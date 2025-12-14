package repositories

import (
	"context"
	"database/sql"
	"errors"
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

	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1)", record.ID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrTenantAlreadyExists
	}

	for _, domain := range record.Domains {
		var existingTenantID string
		err = tx.QueryRowContext(ctx,
			"SELECT tenant_id FROM tenant_domains WHERE domain = $1",
			domain,
		).Scan(&existingTenantID)
		if err == nil {
			return ErrDomainAlreadyExists
		}
		if err != sql.ErrNoRows {
			return err
		}
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO tenants (id, name, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		record.ID, record.Name, record.Status, record.CreatedAt, record.UpdatedAt,
	)
	if err != nil {
		return err
	}

	for _, domain := range record.Domains {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO tenant_domains (domain, tenant_id, created_at)
			 VALUES ($1, $2, $3)`,
			domain, record.ID, record.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO tenant_oidc_configs (tenant_id, discovery_url, client_id, auth_method, scopes, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		record.ID, record.DiscoveryURL, record.ClientID, record.AuthMethod, record.Scopes, record.CreatedAt, record.UpdatedAt,
	)
	if err != nil {
		return err
	}

	expiresAt := record.CreatedAt.Add(7 * 24 * time.Hour)
	_, err = tx.ExecContext(ctx,
		`INSERT INTO invitations (tenant_id, email, role, status, created_at, expires_at)
		 VALUES ($1, $2, 'admin', 'pending', $3, $4)`,
		record.ID, record.FirstAdminEmail, record.CreatedAt, expiresAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TenantRepository) GetByID(ctx context.Context, id string) (*TenantRecord, error) {
	record := &TenantRecord{}

	err := r.db.QueryRowContext(ctx,
		`SELECT t.id, t.name, t.status, t.created_at, t.updated_at,
		        oc.discovery_url, oc.client_id, oc.auth_method, oc.scopes
		 FROM tenants t
		 JOIN tenant_oidc_configs oc ON t.id = oc.tenant_id
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
		"SELECT domain FROM tenant_domains WHERE tenant_id = $1",
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
	query := `SELECT t.id, t.name, t.status, t.created_at, t.updated_at
			  FROM tenants t`
	var args []interface{}
	var conditions []string
	argIndex := 1

	if status != "" {
		conditions = append(conditions, "t.status = $"+string(rune('0'+argIndex)))
		args = append(args, status)
		argIndex++
	}

	if domain != "" {
		query += " JOIN tenant_domains td ON t.id = td.tenant_id"
		conditions = append(conditions, "td.domain = $"+string(rune('0'+argIndex)))
		args = append(args, domain)
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY t.created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*TenantRecord
	for rows.Next() {
		record := &TenantRecord{}
		if err := rows.Scan(&record.ID, &record.Name, &record.Status, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return nil, err
		}

		domainRows, err := r.db.QueryContext(ctx,
			"SELECT domain FROM tenant_domains WHERE tenant_id = $1",
			record.ID,
		)
		if err != nil {
			return nil, err
		}

		for domainRows.Next() {
			var d string
			if err := domainRows.Scan(&d); err != nil {
				domainRows.Close()
				return nil, err
			}
			record.Domains = append(record.Domains, d)
		}
		domainRows.Close()

		records = append(records, record)
	}

	return records, nil
}

func (r *TenantRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1)",
		id,
	).Scan(&exists)
	return exists, err
}

func (r *TenantRepository) DomainExists(ctx context.Context, domain string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM tenant_domains WHERE domain = $1)",
		domain,
	).Scan(&exists)
	return exists, err
}
