package readmodels

import (
	"context"
	"database/sql"
	"fmt"
)

type TenantCacheEntry struct {
	TenantID     string
	Name         string
	Status       string
	Domains      []string
	DiscoveryURL string
	IssuerURL    string
	ClientID     string
	AuthMethod   string
	Scopes       string
}

type TenantCacheReadModel struct {
	db *sql.DB
}

func NewTenantCacheReadModel(db *sql.DB) *TenantCacheReadModel {
	return &TenantCacheReadModel{db: db}
}

func (rm *TenantCacheReadModel) Save(ctx context.Context, entry TenantCacheEntry) error {
	tx, err := rm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tenant cache transaction %s: %w", entry.TenantID, err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := upsertTenant(ctx, tx, entry); err != nil {
		return err
	}
	if err := upsertTenantDomains(ctx, tx, entry); err != nil {
		return err
	}
	if err := upsertTenantOIDC(ctx, tx, entry); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tenant cache %s: %w", entry.TenantID, err)
	}
	return nil
}

func upsertTenant(ctx context.Context, tx *sql.Tx, entry TenantCacheEntry) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO auth.tenant_cache (tenant_id, name, status)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (tenant_id) DO UPDATE SET name = EXCLUDED.name, status = EXCLUDED.status`,
		entry.TenantID, entry.Name, entry.Status,
	)
	if err != nil {
		return fmt.Errorf("cache tenant %s: %w", entry.TenantID, err)
	}
	return nil
}

func upsertTenantDomains(ctx context.Context, tx *sql.Tx, entry TenantCacheEntry) error {
	for _, domain := range entry.Domains {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO auth.tenant_domain_cache (domain, tenant_id)
			 VALUES ($1, $2)
			 ON CONFLICT (domain) DO UPDATE SET tenant_id = EXCLUDED.tenant_id`,
			domain, entry.TenantID,
		)
		if err != nil {
			return fmt.Errorf("cache tenant domain %s: %w", domain, err)
		}
	}
	return nil
}

func upsertTenantOIDC(ctx context.Context, tx *sql.Tx, entry TenantCacheEntry) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO auth.tenant_oidc_cache (tenant_id, discovery_url, issuer_url, client_id, auth_method, scopes)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (tenant_id) DO UPDATE SET
		     discovery_url = EXCLUDED.discovery_url,
		     issuer_url = EXCLUDED.issuer_url,
		     client_id = EXCLUDED.client_id,
		     auth_method = EXCLUDED.auth_method,
		     scopes = EXCLUDED.scopes`,
		entry.TenantID, entry.DiscoveryURL, nullIfEmpty(entry.IssuerURL),
		entry.ClientID, entry.AuthMethod, entry.Scopes,
	)
	if err != nil {
		return fmt.Errorf("cache tenant OIDC configuration %s: %w", entry.TenantID, err)
	}
	return nil
}

func (rm *TenantCacheReadModel) Exists(ctx context.Context, tenantID string) (bool, error) {
	var exists bool
	err := rm.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM auth.tenant_cache WHERE tenant_id = $1)`,
		tenantID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("look up cached tenant %s: %w", tenantID, err)
	}
	return exists, nil
}

func nullIfEmpty(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}
