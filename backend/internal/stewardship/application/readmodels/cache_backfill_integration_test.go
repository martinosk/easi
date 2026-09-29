//go:build integration

package readmodels

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
	"easi/backend/internal/stewardship/domain/valueobjects"
	"easi/backend/internal/testing/testdb"
)

const (
	cacheBackfillMigration = "167_backfill_stewardship_caches.sql"
	backfillTestTenant     = "stewardship-backfill-test"
)

type backfillFixture struct {
	admin     *sql.DB
	tenant    string
	activeID  string
	disabled  string
	architect string
	domainID  string
}

func seedBackfillSources(t *testing.T) backfillFixture {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	f := backfillFixture{
		admin:     testdb.OpenAdmin(t),
		tenant:    backfillTestTenant,
		activeID:  uuid.NewString(),
		disabled:  uuid.NewString(),
		architect: uuid.NewString(),
		domainID:  uuid.NewString(),
	}
	t.Cleanup(func() {
		_, _ = f.admin.Exec("DELETE FROM stewardship.user_cache WHERE tenant_id = $1 AND user_id = ANY($2::text[])", f.tenant, "{"+f.activeID+","+f.disabled+","+f.architect+"}")
		_, _ = f.admin.Exec("DELETE FROM stewardship.domain_cache WHERE tenant_id = $1 AND domain_id = $2", f.tenant, f.domainID)
		_, _ = f.admin.Exec("DELETE FROM capabilitymapping.business_domains WHERE tenant_id = $1 AND id = $2", f.tenant, f.domainID)
		_, _ = f.admin.Exec("DELETE FROM auth.users WHERE tenant_id = $1", f.tenant)
		_, _ = f.admin.Exec("DELETE FROM auth.tenants WHERE id = $1", f.tenant)
	})
	exec := func(query string, args ...any) {
		t.Helper()
		_, err := f.admin.Exec(query, args...)
		require.NoError(t, err)
	}
	exec(`INSERT INTO auth.tenants (id, name, status, created_at, updated_at) VALUES ($1, 'Stewardship Backfill', 'active', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING`, f.tenant)
	exec(`INSERT INTO auth.users (id, tenant_id, email, name, role, status) VALUES
		($2, $1, $5, 'Mette Gram', 'stakeholder', 'active'),
		($3, $1, $6, 'Jonas Holm', 'stakeholder', 'disabled'),
		($4, $1, $7, 'Alice Smith', 'architect', 'active')`,
		f.tenant, f.activeID, f.disabled, f.architect,
		f.activeID+"@example.com", f.disabled+"@example.com", f.architect+"@example.com")
	exec(`INSERT INTO capabilitymapping.business_domains (id, tenant_id, name, created_at, domain_architect_id)
		VALUES ($2, $1, $3, NOW(), $4)`, f.tenant, f.domainID, "Backfill "+f.domainID[:8], f.architect)
	return f
}

func runCacheBackfill(t *testing.T, db *sql.DB) {
	t.Helper()
	sqlBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "deploy-scripts", "migrations", cacheBackfillMigration))
	require.NoError(t, err)
	_, err = db.Exec(string(sqlBytes))
	require.NoError(t, err)
}

func TestBackfillMigration_SeedsUsersWithStatusAndDomainsWithArchitect(t *testing.T) {
	f := seedBackfillSources(t)

	runCacheBackfill(t, f.admin)
	runCacheBackfill(t, f.admin)

	db := database.NewTenantAwareDB(f.admin)
	ctx := sharedctx.WithTenant(context.Background(), sharedvo.MustNewTenantID(backfillTestTenant))
	users := NewUserCacheReadModel(db)
	standing, err := users.UserStanding(ctx, f.activeID)
	require.NoError(t, err)
	assert.Equal(t, valueobjects.UserActive, standing)
	standing, err = users.UserStanding(ctx, f.disabled)
	require.NoError(t, err)
	assert.Equal(t, valueobjects.UserDisabled, standing)

	domain, err := NewDomainCacheReadModel(db).getWithArchitect(ctx, f.domainID)
	require.NoError(t, err)
	require.NotNil(t, domain)
	assert.Equal(t, f.architect, *domain.DomainArchitectID)
	assert.Equal(t, "Alice Smith", *domain.DomainArchitectName)
}
