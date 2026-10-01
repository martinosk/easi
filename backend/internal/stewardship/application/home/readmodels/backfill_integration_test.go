//go:build integration

package readmodels

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	adPL "easi/backend/internal/accessdelegation/publishedlanguage"
	directionPL "easi/backend/internal/architecturedirection/publishedlanguage"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/stewardship/application/home"
	"easi/backend/internal/testing/testdb"
)

const homeCacheBackfillMigration = "169_backfill_stewardship_home_caches.sql"

var homeCacheRows = map[string]string{
	"capability_cache":        "SELECT capability_id, name, level, COALESCE(parent_id, ''), status, COALESCE(ea_owner, '') FROM stewardship.capability_cache WHERE tenant_id = $1 ORDER BY 1",
	"domain_assignment_cache": "SELECT capability_id, domain_id FROM stewardship.domain_assignment_cache WHERE tenant_id = $1 ORDER BY 1, 2",
	"application_cache":       "SELECT component_id, name, ownership_state, COALESCE(owner_kind, ''), COALESCE(owner_id, '') FROM stewardship.application_cache WHERE tenant_id = $1 ORDER BY 1",
	"realization_cache":       "SELECT realization_id, capability_id, component_id FROM stewardship.realization_cache WHERE tenant_id = $1 ORDER BY 1",
	"time_assessment_cache":   "SELECT capability_id, component_id, grade, to_char(assessed_at, 'YYYY-MM-DD HH24:MI:SS') FROM stewardship.time_assessment_cache WHERE tenant_id = $1 ORDER BY 1, 2",
	"edit_grant_cache":        "SELECT grant_id, artifact_type, artifact_id, grantee_email, to_char(expires_at, 'YYYY-MM-DD HH24:MI:SS') FROM stewardship.edit_grant_cache WHERE tenant_id = $1 ORDER BY 1",
}

type backfillSource struct {
	l1, l2, domain, owned, deleted, direct, inherited, grant, revokedGrant string
	ownerID                                                                string
	assessedAt, expiresAt                                                  time.Time
}

func newBackfillSource() backfillSource {
	return backfillSource{
		l1: uuid.NewString(), l2: uuid.NewString(), domain: uuid.NewString(),
		owned: uuid.NewString(), deleted: uuid.NewString(),
		direct: uuid.NewString(), inherited: uuid.NewString(),
		grant: uuid.NewString(), revokedGrant: uuid.NewString(), ownerID: uuid.NewString(),
		assessedAt: time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC),
		expiresAt:  time.Date(2099, 10, 21, 8, 0, 0, 0, time.UTC),
	}
}

func seedSupplierTables(t *testing.T, admin *sql.DB, tenant string, s backfillSource) {
	t.Helper()
	t.Cleanup(func() {
		for _, table := range []string{"capabilitymapping.capabilities", "capabilitymapping.domain_capability_assignments",
			"capabilitymapping.capability_realizations", "architecturemodeling.application_components",
			"architecturedirection.time_assessments", "accessdelegation.edit_grants"} {
			_, _ = admin.Exec("DELETE FROM "+table+" WHERE tenant_id = $1", tenant)
		}
	})
	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO capabilitymapping.capabilities (id, tenant_id, name, level, parent_id, created_at, status, ea_owner) VALUES
			($2, $1, 'Accounting', 'L1', NULL, NOW(), 'Active', NULL),
			($3, $1, 'Invoicing', 'L2', $2, NOW(), 'Planned', $4)`, []any{s.l1, s.l2, s.ownerID}},
		{`INSERT INTO capabilitymapping.domain_capability_assignments
			(assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_name, capability_level, assigned_at)
			VALUES ($2, $1, $3, 'Finance', $4, 'Accounting', 'L1', NOW())`, []any{uuid.NewString(), s.domain, s.l1}},
		{`INSERT INTO architecturemodeling.application_components (id, tenant_id, name, created_at, is_deleted, ownership_state, owner_kind, owner_id) VALUES
			($2, $1, 'CRM', NOW(), FALSE, 'owned', 'user', $4),
			($3, $1, 'Legacy', NOW(), TRUE, 'unknown', NULL, NULL)`, []any{s.owned, s.deleted, s.ownerID}},
		{`INSERT INTO capabilitymapping.capability_realizations (id, tenant_id, capability_id, component_id, realization_level, linked_at, origin) VALUES
			($2, $1, $4, $5, 'Full', NOW(), 'Direct'),
			($3, $1, $6, $5, 'Full', NOW(), 'Inherited')`, []any{s.direct, s.inherited, s.l2, s.owned, s.l1}},
		{`INSERT INTO architecturedirection.time_assessments (id, tenant_id, capability_id, component_id, realization_id, grade, assessed_by, assessed_by_name, assessed_at)
			VALUES ($2, $1, $3, $4, $5, 'Invest', 'u', 'U', $6)`, []any{uuid.NewString(), s.l2, s.owned, s.direct, s.assessedAt}},
		{`INSERT INTO accessdelegation.edit_grants (id, tenant_id, grantor_id, grantor_email, grantee_email, artifact_type, artifact_id, status, expires_at) VALUES
			($2, $1, 'g', 'g@example.com', 'ole@example.com', 'component', $4, 'active', $5),
			($3, $1, 'g', 'g@example.com', 'ole@example.com', 'capability', $6, 'revoked', $5)`, []any{s.grant, s.revokedGrant, s.owned, s.expiresAt, s.l2}},
	}
	for _, stmt := range statements {
		_, err := admin.Exec(stmt.query, append([]any{tenant}, stmt.args...)...)
		require.NoError(t, err)
	}
}

func projectSameFacts(t *testing.T, l *landscape, s backfillSource) {
	t.Helper()
	caches := l.caches
	steps := []struct {
		handler   eventHandler
		eventType string
		data      map[string]any
	}{
		{home.NewCapabilityCacheProjector(caches), capPL.CapabilityCreated, map[string]any{"id": s.l1, "name": "Accounting", "level": "L1"}},
		{home.NewCapabilityCacheProjector(caches), capPL.CapabilityCreated, map[string]any{"id": s.l2, "name": "Invoicing", "level": "L2", "parentId": s.l1}},
		{home.NewCapabilityCacheProjector(caches), capPL.CapabilityMetadataUpdated, map[string]any{"id": s.l2, "status": "Planned", "eaOwner": s.ownerID}},
		{home.NewDomainAssignmentCacheProjector(caches), capPL.CapabilityAssignedToDomain, map[string]any{"capabilityId": s.l1, "businessDomainId": s.domain}},
		{home.NewApplicationCacheProjector(caches), amPL.ApplicationComponentCreated, map[string]any{"id": s.owned, "name": "CRM"}},
		{home.NewApplicationCacheProjector(caches), amPL.ApplicationOwnershipConfirmed, map[string]any{"componentId": s.owned, "ownerKind": "user", "ownerId": s.ownerID, "ownershipState": "owned"}},
		{home.NewApplicationCacheProjector(caches), amPL.ApplicationComponentCreated, map[string]any{"id": s.deleted, "name": "Legacy"}},
		{home.NewApplicationCacheProjector(caches), amPL.ApplicationComponentDeleted, map[string]any{"id": s.deleted}},
		{home.NewRealizationCacheProjector(caches), capPL.SystemLinkedToCapability, map[string]any{"id": s.direct, "capabilityId": s.l2, "componentId": s.owned}},
		{home.NewTimeAssessmentCacheProjector(caches), directionPL.TimeAssessmentRecorded, map[string]any{"capabilityId": s.l2, "componentId": s.owned, "grade": "Invest", "occurredOn": s.assessedAt}},
		{home.NewEditGrantCacheProjector(caches), adPL.EditGrantActivated, map[string]any{"id": s.grant, "artifactType": "component", "artifactId": s.owned, "granteeEmail": "ole@example.com", "expiresAt": s.expiresAt.Format(time.RFC3339)}},
		{home.NewEditGrantCacheProjector(caches), adPL.EditGrantActivated, map[string]any{"id": s.revokedGrant, "artifactType": "capability", "artifactId": s.l2, "granteeEmail": "ole@example.com", "expiresAt": s.expiresAt.Format(time.RFC3339)}},
		{home.NewEditGrantCacheProjector(caches), adPL.EditGrantRevoked, map[string]any{"id": s.revokedGrant}},
	}
	for _, step := range steps {
		require.NoError(t, step.handler.Handle(l.ctx, storedEvent(t, step.eventType, step.data)))
	}
}

func cacheRows(t *testing.T, admin *sql.DB, query, tenant string) [][]string {
	t.Helper()
	rows, err := admin.QueryContext(context.Background(), query, tenant)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	columns, err := rows.Columns()
	require.NoError(t, err)
	var result [][]string
	for rows.Next() {
		values := make([]string, len(columns))
		targets := make([]any, len(columns))
		for i := range values {
			targets[i] = &values[i]
		}
		require.NoError(t, rows.Scan(targets...))
		result = append(result, values)
	}
	require.NoError(t, rows.Err())
	return result
}

func runHomeCacheBackfill(t *testing.T, admin *sql.DB) {
	t.Helper()
	sqlBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", "deploy-scripts", "migrations", homeCacheBackfillMigration))
	require.NoError(t, err)
	_, err = admin.Exec(string(sqlBytes))
	require.NoError(t, err)
}

func TestHomeCacheBackfill_SeedsWhatTheProjectorsWouldHold(t *testing.T) {
	l := newLandscape(t)
	admin := testdb.OpenAdmin(t)
	source := newBackfillSource()
	backfilled := l.tenant
	seedSupplierTables(t, admin, backfilled, source)

	runHomeCacheBackfill(t, admin)
	runHomeCacheBackfill(t, admin)

	l.useTenant("home-it-" + uuid.NewString()[:8])
	projectSameFacts(t, l, source)

	for cache, query := range homeCacheRows {
		t.Run(cache, func(t *testing.T) {
			got := cacheRows(t, admin, query, backfilled)
			assert.NotEmpty(t, got)
			assert.Equal(t, cacheRows(t, admin, query, l.tenant), got)
		})
	}
}
