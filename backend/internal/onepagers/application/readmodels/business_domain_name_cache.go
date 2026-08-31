package readmodels

import (
	"context"
	"database/sql"
	"fmt"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type BusinessDomainNameCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewBusinessDomainNameCacheReadModel(db *database.TenantAwareDB) *BusinessDomainNameCacheReadModel {
	return &BusinessDomainNameCacheReadModel{db: db}
}

func (rm *BusinessDomainNameCacheReadModel) Save(ctx context.Context, businessDomainID, name string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx,
		`INSERT INTO onepagers.business_domain_name_cache (tenant_id, business_domain_id, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id, business_domain_id) DO UPDATE SET name = EXCLUDED.name`,
		tenantID.Value(), businessDomainID, name,
	)
	if err != nil {
		return fmt.Errorf("cache business domain name for %s: %w", businessDomainID, err)
	}
	return nil
}

func (rm *BusinessDomainNameCacheReadModel) Delete(ctx context.Context, businessDomainID string) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx,
		`DELETE FROM onepagers.business_domain_name_cache WHERE tenant_id = $1 AND business_domain_id = $2`,
		tenantID.Value(), businessDomainID,
	)
	if err != nil {
		return fmt.Errorf("remove cached business domain name for %s: %w", businessDomainID, err)
	}
	return nil
}

func (rm *BusinessDomainNameCacheReadModel) Name(ctx context.Context, businessDomainID string) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}
	var name string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		scanErr := tx.QueryRowContext(ctx,
			`SELECT name FROM onepagers.business_domain_name_cache WHERE tenant_id = $1 AND business_domain_id = $2`,
			tenantID.Value(), businessDomainID,
		).Scan(&name)
		if scanErr == sql.ErrNoRows {
			return nil
		}
		return scanErr
	})
	if err != nil {
		return "", fmt.Errorf("read cached business domain name for %s: %w", businessDomainID, err)
	}
	return name, nil
}
