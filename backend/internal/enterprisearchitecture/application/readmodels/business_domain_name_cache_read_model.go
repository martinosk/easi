package readmodels

import (
	"context"
	"database/sql"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type BusinessDomainNameCacheReadModel struct {
	db *database.TenantAwareDB
}

func NewBusinessDomainNameCacheReadModel(db *database.TenantAwareDB) *BusinessDomainNameCacheReadModel {
	return &BusinessDomainNameCacheReadModel{db: db}
}

func (rm *BusinessDomainNameCacheReadModel) Upsert(ctx context.Context, businessDomainID, name string) error {
	return rm.execForTenant(ctx,
		`INSERT INTO enterprisearchitecture.business_domain_name_cache (tenant_id, business_domain_id, name)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (tenant_id, business_domain_id) DO UPDATE SET name = EXCLUDED.name`,
		businessDomainID, name,
	)
}

func (rm *BusinessDomainNameCacheReadModel) Delete(ctx context.Context, businessDomainID string) error {
	return rm.execForTenant(ctx,
		"DELETE FROM enterprisearchitecture.business_domain_name_cache WHERE tenant_id = $1 AND business_domain_id = $2",
		businessDomainID,
	)
}

func (rm *BusinessDomainNameCacheReadModel) Name(ctx context.Context, businessDomainID string) (string, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}
	var name string
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT name FROM enterprisearchitecture.business_domain_name_cache WHERE tenant_id = $1 AND business_domain_id = $2`,
			tenantID.Value(), businessDomainID,
		).Scan(&name)
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	})
	return name, err
}

func (rm *BusinessDomainNameCacheReadModel) execForTenant(ctx context.Context, query string, args ...any) error {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return err
	}
	_, err = rm.db.ExecContext(ctx, query, append([]any{tenantID.Value()}, args...)...)
	return err
}
