package repositories

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrDomainNotFound = errors.New("domain not registered")
	ErrTenantNotFound = errors.New("tenant not found")
	ErrTenantInactive = errors.New("tenant is not active")
)

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
	var config TenantOIDCConfig
	var status string
	var issuerURL sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT t.id, t.status, oc.discovery_url, oc.issuer_url, oc.client_id, oc.auth_method, oc.scopes
		 FROM platform.tenant_domains td
		 JOIN platform.tenants t ON td.tenant_id = t.id
		 JOIN platform.tenant_oidc_configs oc ON t.id = oc.tenant_id
		 WHERE td.domain = $1`,
		emailDomain,
	).Scan(&config.TenantID, &status, &config.DiscoveryURL, &issuerURL, &config.ClientID, &config.AuthMethod, &config.Scopes)

	if err == sql.ErrNoRows {
		return nil, ErrDomainNotFound
	}
	if err != nil {
		return nil, err
	}

	if status != "active" {
		return nil, ErrTenantInactive
	}

	if issuerURL.Valid {
		config.IssuerURL = issuerURL.String
	}

	return &config, nil
}

func (r *TenantOIDCRepository) GetByTenantID(ctx context.Context, tenantID string) (*TenantOIDCConfig, error) {
	var config TenantOIDCConfig
	var status string
	var issuerURL sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT t.id, t.status, oc.discovery_url, oc.issuer_url, oc.client_id, oc.auth_method, oc.scopes
		 FROM platform.tenants t
		 JOIN platform.tenant_oidc_configs oc ON t.id = oc.tenant_id
		 WHERE t.id = $1`,
		tenantID,
	).Scan(&config.TenantID, &status, &config.DiscoveryURL, &issuerURL, &config.ClientID, &config.AuthMethod, &config.Scopes)

	if err == sql.ErrNoRows {
		return nil, ErrTenantNotFound
	}
	if err != nil {
		return nil, err
	}

	if status != "active" {
		return nil, ErrTenantInactive
	}

	if issuerURL.Valid {
		config.IssuerURL = issuerURL.String
	}

	return &config, nil
}
