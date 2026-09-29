package readmodels

import (
	"context"
	"database/sql"
	"errors"

	"easi/backend/internal/infrastructure/database"
)

type CachedDomain struct {
	ID                string
	Name              string
	DomainArchitectID string
}

type cachedDomainWithArchitect struct {
	ID                  string
	Name                string
	DomainArchitectID   *string
	DomainArchitectName *string
}

type DomainCacheReadModel struct {
	store tenantScopedDB
}

func NewDomainCacheReadModel(db *database.TenantAwareDB) *DomainCacheReadModel {
	return &DomainCacheReadModel{store: tenantScopedDB{db: db}}
}

func (rm *DomainCacheReadModel) Upsert(ctx context.Context, domain CachedDomain) error {
	return rm.store.exec(ctx, "cache business domain "+domain.ID, `
		INSERT INTO stewardship.domain_cache (tenant_id, domain_id, name, domain_architect_id)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		ON CONFLICT (tenant_id, domain_id) DO UPDATE SET
			name = EXCLUDED.name, domain_architect_id = EXCLUDED.domain_architect_id`,
		domain.ID, domain.Name, domain.DomainArchitectID)
}

func (rm *DomainCacheReadModel) Delete(ctx context.Context, domainID string) error {
	return rm.store.exec(ctx, "forget cached business domain "+domainID, `
		DELETE FROM stewardship.domain_cache WHERE tenant_id = $1 AND domain_id = $2`,
		domainID)
}

func (rm *DomainCacheReadModel) DomainExists(ctx context.Context, domainID string) (bool, error) {
	domain, err := rm.getWithArchitect(ctx, domainID)
	return domain != nil, err
}

func (rm *DomainCacheReadModel) getWithArchitect(ctx context.Context, domainID string) (*cachedDomainWithArchitect, error) {
	var (
		domain        cachedDomainWithArchitect
		architectID   sql.NullString
		architectName sql.NullString
	)
	err := rm.store.read(ctx, "read cached business domain "+domainID, func(tx *sql.Tx, tenantID string) error {
		return tx.QueryRowContext(ctx, `
			SELECT d.domain_id, d.name, d.domain_architect_id, NULLIF(COALESCE(NULLIF(u.name, ''), u.email), '')
			FROM stewardship.domain_cache d
			LEFT JOIN stewardship.user_cache u ON u.tenant_id = d.tenant_id AND u.user_id = d.domain_architect_id
			WHERE d.tenant_id = $1 AND d.domain_id = $2`,
			tenantID, domainID,
		).Scan(&domain.ID, &domain.Name, &architectID, &architectName)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	domain.DomainArchitectID = nullableString(architectID)
	domain.DomainArchitectName = nullableString(architectName)
	return &domain, nil
}
