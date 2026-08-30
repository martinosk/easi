package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrDomainNotFound = errors.New("domain not registered")
	ErrTenantNotFound = errors.New("tenant not found")
	ErrTenantInactive = errors.New("tenant is not active")
)

const tenantOIDCColumns = `t.tenant_id, t.status, oc.discovery_url, oc.issuer_url, oc.client_id, oc.auth_method, oc.scopes`

type TenantOIDCConfig struct {
	TenantID     string
	DiscoveryURL string
	IssuerURL    string
	ClientID     string
	AuthMethod   string
	Scopes       string
}

type TenantOIDCRepository struct {
	db *sql.DB
}

func NewTenantOIDCRepository(db *sql.DB) *TenantOIDCRepository {
	return &TenantOIDCRepository{db: db}
}

func (r *TenantOIDCRepository) GetByEmailDomain(ctx context.Context, emailDomain string) (*TenantOIDCConfig, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+tenantOIDCColumns+`
		 FROM auth.tenant_domain_cache td
		 JOIN auth.tenant_cache t ON td.tenant_id = t.tenant_id
		 JOIN auth.tenant_oidc_cache oc ON t.tenant_id = oc.tenant_id
		 WHERE td.domain = $1`,
		emailDomain,
	)

	config, err := scanActiveTenantOIDC(row, ErrDomainNotFound)
	if err != nil {
		return nil, fmt.Errorf("resolve OIDC configuration for domain %s: %w", emailDomain, err)
	}
	return config, nil
}

func (r *TenantOIDCRepository) GetByTenantID(ctx context.Context, tenantID string) (*TenantOIDCConfig, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+tenantOIDCColumns+`
		 FROM auth.tenant_cache t
		 JOIN auth.tenant_oidc_cache oc ON t.tenant_id = oc.tenant_id
		 WHERE t.tenant_id = $1`,
		tenantID,
	)

	config, err := scanActiveTenantOIDC(row, ErrTenantNotFound)
	if err != nil {
		return nil, fmt.Errorf("resolve OIDC configuration for tenant %s: %w", tenantID, err)
	}
	return config, nil
}

func scanActiveTenantOIDC(row *sql.Row, missingErr error) (*TenantOIDCConfig, error) {
	var config TenantOIDCConfig
	var status string
	var issuerURL sql.NullString

	err := row.Scan(&config.TenantID, &status, &config.DiscoveryURL, &issuerURL,
		&config.ClientID, &config.AuthMethod, &config.Scopes)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, missingErr
	}
	if err != nil {
		return nil, err
	}

	if status != "active" {
		return nil, ErrTenantInactive
	}

	config.IssuerURL = issuerURL.String
	return &config, nil
}
