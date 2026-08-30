//go:build integration

package readmodels_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/readmodels"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const backfillMigration = "148_backfill_onepagers_subject_caches.sql"

var backfilledTables = []string{
	"onepagers.subject_relation_cache",
	"onepagers.business_domain_name_cache",
	"onepagers.maturity_scale_cache",
	"onepagers.one_pager_subject_index",
	"capabilitymapping.capabilities",
	"capabilitymapping.capability_experts",
	"capabilitymapping.capability_realizations",
	"capabilitymapping.capability_dependencies",
	"capabilitymapping.domain_capability_assignments",
	"capabilitymapping.business_domains",
	"architecturemodeling.application_components",
	"architecturemodeling.built_by_relationships",
	"architecturemodeling.internal_teams",
	"metamodel.meta_model_configurations",
}

type backfillFixture struct {
	t      *testing.T
	db     *sql.DB
	ctx    context.Context
	tenant string
}

func newBackfillFixture(t *testing.T) *backfillFixture {
	t.Helper()
	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi password=easi dbname=easi sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	tenant := "test-bf-" + uuid.NewString()[:8]
	tenantID, err := sharedvo.NewTenantID(tenant)
	require.NoError(t, err)

	t.Cleanup(func() {
		for _, table := range backfilledTables {
			_, _ = db.Exec("DELETE FROM "+table+" WHERE tenant_id = $1", tenant)
		}
		_ = db.Close()
	})

	return &backfillFixture{t: t, db: db, ctx: sharedctx.WithTenant(context.Background(), tenantID), tenant: tenant}
}

func (f *backfillFixture) exec(query string, args ...any) {
	f.t.Helper()
	_, err := f.db.Exec(query, args...)
	require.NoError(f.t, err)
}

func (f *backfillFixture) seedSuppliers() {
	f.exec(`INSERT INTO capabilitymapping.business_domains (id, tenant_id, name, created_at) VALUES ('bd-1', $1, 'Finance', NOW())`, f.tenant)
	f.exec(`INSERT INTO capabilitymapping.capabilities (id, tenant_id, name, description, level, parent_id, maturity_value, status, created_at) VALUES
		('cap-1', $1, 'Finance Ops', 'Runs finance', 'L1', NULL, 40, 'Active', NOW()),
		('cap-2', $1, 'Billing', 'Bills customers', 'L2', 'cap-1', 55, 'Active', NOW())`, f.tenant)
	f.exec(`INSERT INTO capabilitymapping.capability_experts (capability_id, tenant_id, expert_name, expert_role, contact_info, added_at)
		VALUES ('cap-2', $1, 'Alice', 'Owner', 'alice@dfds.com', NOW())`, f.tenant)
	f.exec(`INSERT INTO capabilitymapping.capability_dependencies (id, tenant_id, source_capability_id, target_capability_id, dependency_type, created_at)
		VALUES ('dep-1', $1, 'cap-2', 'cap-1', 'Requires', NOW())`, f.tenant)
	f.exec(`INSERT INTO capabilitymapping.domain_capability_assignments
		(assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_name, capability_level, assigned_at)
		VALUES ('asg-1', $1, 'bd-1', 'Stale Name', 'cap-2', 'Billing', 'L2', NOW())`, f.tenant)
	f.exec(`INSERT INTO architecturemodeling.application_components (id, tenant_id, name, description, created_at, is_deleted)
		VALUES ('app-1', $1, 'Billing Service', 'Handles invoicing', NOW(), FALSE)`, f.tenant)
	f.exec(`INSERT INTO capabilitymapping.capability_realizations (id, tenant_id, capability_id, component_id, component_name, realization_level, linked_at)
		VALUES ('rz-1', $1, 'cap-2', 'app-1', 'Billing Service', 'Full', NOW())`, f.tenant)
	f.exec(`INSERT INTO architecturemodeling.internal_teams (id, tenant_id, name, department, contact_person, created_at, is_deleted)
		VALUES ('team-1', $1, 'Platform', 'Engineering', 'Carol', NOW(), FALSE)`, f.tenant)
	f.exec(`INSERT INTO architecturemodeling.built_by_relationships (id, tenant_id, internal_team_id, component_id, created_at)
		VALUES ('app-1', $1, 'team-1', 'app-1', NOW())`, f.tenant)
	f.exec(`INSERT INTO metamodel.meta_model_configurations (id, tenant_id, sections, strategy_pillars, version, is_default, created_at, modified_at, modified_by)
		VALUES ('cfg-1', $1, '[{"order":1,"name":"Exploring","minValue":0,"maxValue":39}]'::jsonb, '[]'::jsonb, 1, false, NOW(), NOW(), 'admin')`, f.tenant)

	for _, subject := range []struct{ subjectType, id, name string }{
		{"capability", "cap-1", "Finance Ops"},
		{"capability", "cap-2", "Billing"},
		{"application", "app-1", "Billing Service"},
		{"internal-team", "team-1", "Platform"},
	} {
		f.exec(`INSERT INTO onepagers.one_pager_subject_index
			(tenant_id, subject_type, subject_id, name, creator_actor_id, creator_email, created_at, last_updated_at)
			VALUES ($1, $2, $3, $4, '', '', NOW(), NOW())`, f.tenant, subject.subjectType, subject.id, subject.name)
	}
}

func (f *backfillFixture) runBackfill() {
	f.t.Helper()
	sqlBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "deploy-scripts", "migrations", backfillMigration))
	require.NoError(f.t, err)
	f.exec(string(sqlBytes))
}

func subjectKey(subjectType, subjectID string) readmodels.SubjectKey {
	return readmodels.SubjectKey{SubjectType: subjectType, SubjectID: subjectID}
}

func (f *backfillFixture) attributes(subject readmodels.SubjectKey) readmodels.SubjectAttributes {
	f.t.Helper()
	index := readmodels.NewOnePagerSubjectIndexReadModel(database.NewTenantAwareDB(f.db))
	row, err := index.AttributeRow(f.ctx, subject)
	require.NoError(f.t, err)
	require.NotNil(f.t, row)
	return row.Attributes
}

func (f *backfillFixture) references(subject readmodels.SubjectKey, entryIDs ...string) []readmodels.RelationReference {
	f.t.Helper()
	relations := readmodels.NewSubjectRelationCacheReadModel(database.NewTenantAwareDB(f.db))
	references, err := relations.References(f.ctx, readmodels.RelationQuery{SubjectType: subject.SubjectType, SubjectIDs: []string{subject.SubjectID}, EntryIDs: entryIDs})
	require.NoError(f.t, err)
	return references[subject.SubjectID]
}

func TestBackfillMigration_SeedsPublishedAttributesFromSupplierTables(t *testing.T) {
	f := newBackfillFixture(t)
	f.seedSuppliers()

	f.runBackfill()

	capability := f.attributes(subjectKey("capability", "cap-2"))
	assert.JSONEq(t, `"Bills customers"`, string(capability["description"]))
	assert.JSONEq(t, `55`, string(capability["maturityValue"]))
	assert.JSONEq(t, `"L2"`, string(capability["level"]), "attributes beyond the rendered catalogue are cached too")
	assert.JSONEq(t, `"cap-1"`, string(capability["parentId"]))
	assert.JSONEq(t, `[{"expertName":"Alice","expertRole":"Owner","contactInfo":"alice@dfds.com"}]`, string(capability["experts"]))

	team := f.attributes(subjectKey("internal-team", "team-1"))
	assert.JSONEq(t, `"Engineering"`, string(team["department"]))
	assert.JSONEq(t, `"Carol"`, string(team["contactPerson"]))
}

func TestBackfillMigration_SeedsRelationsInBothDirections(t *testing.T) {
	f := newBackfillFixture(t)
	f.seedSuppliers()

	f.runBackfill()

	capability := f.references(subjectKey("capability", "cap-2"), "realizing-applications", "depends-on", "parent-capability", "business-domains")
	byEntry := map[string]readmodels.RelationReference{}
	for _, reference := range capability {
		byEntry[reference.EntryID] = reference
	}
	assert.Equal(t, "app-1", byEntry["realizing-applications"].RelatedID)
	assert.Equal(t, "Billing Service", byEntry["realizing-applications"].Label, "labels come from the subject index")
	assert.Equal(t, "cap-1", byEntry["depends-on"].RelatedID)
	assert.Equal(t, "cap-1", byEntry["parent-capability"].RelatedID)
	assert.Equal(t, "bd-1", byEntry["business-domains"].RelatedID)
	assert.Equal(t, "Finance", byEntry["business-domains"].Label, "business domain labels come from the live domain table")

	parent := f.references(subjectKey("capability", "cap-1"), "child-capabilities")
	require.Len(t, parent, 1)
	assert.Equal(t, "cap-2", parent[0].RelatedID)

	application := f.references(subjectKey("application", "app-1"), "realized-capabilities", "built-by")
	assert.Len(t, application, 2)

	team := f.references(subjectKey("internal-team", "team-1"), "built-applications")
	require.Len(t, team, 1)
	assert.Equal(t, "app-1", team[0].RelatedID)
	assert.Equal(t, "Billing Service", team[0].Label)
}

func TestBackfillMigration_SeedsBusinessDomainNamesAndMaturityScale(t *testing.T) {
	f := newBackfillFixture(t)
	f.seedSuppliers()

	f.runBackfill()

	tenantDB := database.NewTenantAwareDB(f.db)
	name, err := readmodels.NewBusinessDomainNameCacheReadModel(tenantDB).Name(f.ctx, "bd-1")
	require.NoError(t, err)
	assert.Equal(t, "Finance", name)

	sections, err := readmodels.NewMaturityScaleCacheReadModel(tenantDB).Sections(f.ctx)
	require.NoError(t, err)
	assert.Equal(t, []readmodels.MaturityScaleSection{{Name: "Exploring", MinValue: 0, MaxValue: 39}}, sections)
}

func TestBackfillMigration_IsIdempotent(t *testing.T) {
	f := newBackfillFixture(t)
	f.seedSuppliers()

	f.runBackfill()
	f.runBackfill()

	var rows int
	require.NoError(t, f.db.QueryRow(
		`SELECT COUNT(*) FROM onepagers.subject_relation_cache WHERE tenant_id = $1`, f.tenant).Scan(&rows))
	assert.Equal(t, 8, rows, "re-running the backfill neither duplicates nor drops relation rows")
}
