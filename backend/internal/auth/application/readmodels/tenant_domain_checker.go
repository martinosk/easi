package readmodels

import (
	"context"
	"database/sql"
	"strings"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type TenantDomainChecker struct {
	db *database.TenantAwareDB
}

func NewTenantDomainChecker(db *database.TenantAwareDB) *TenantDomainChecker {
	return &TenantDomainChecker{db: db}
}

func (c *TenantDomainChecker) IsDomainAllowed(ctx context.Context, email string) (bool, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return false, err
	}

	domain := extractDomain(email)
	if domain == "" {
		return false, nil
	}

	var exists bool
	err = c.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM tenant_domains WHERE tenant_id = $1 AND domain = $2)`,
			tenantID.Value(), domain,
		).Scan(&exists)
	})

	if err != nil {
		return false, err
	}

	return exists, nil
}

func extractDomain(email string) string {
	parts := strings.Split(strings.ToLower(email), "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
