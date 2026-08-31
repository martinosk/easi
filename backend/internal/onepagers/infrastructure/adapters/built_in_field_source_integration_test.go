//go:build integration

package adapters_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/catalog"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/onepagers/infrastructure/adapters"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type cacheFixture struct {
	t         *testing.T
	ctx       context.Context
	index     *readmodels.OnePagerSubjectIndexReadModel
	relations *readmodels.SubjectRelationCacheReadModel
	sources   map[string]ports.BuiltInFieldSource
	subjects  ports.SubjectExistenceChecker
}

func newCacheFixture(t *testing.T) *cacheFixture {
	t.Helper()
	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi_app password=localdev dbname=easi sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	tenant := "test-bif-" + uuid.NewString()[:8]
	tenantID, err := sharedvo.NewTenantID(tenant)
	require.NoError(t, err)

	t.Cleanup(func() {
		for _, table := range []string{"onepagers.subject_relation_cache", "onepagers.one_pager_subject_index"} {
			_, _ = db.Exec("DELETE FROM "+table+" WHERE tenant_id = $1", tenant)
		}
		_ = db.Close()
	})

	tenantDB := database.NewTenantAwareDB(db)
	return &cacheFixture{
		t:         t,
		ctx:       sharedctx.WithTenant(context.Background(), tenantID),
		index:     readmodels.NewOnePagerSubjectIndexReadModel(tenantDB),
		relations: readmodels.NewSubjectRelationCacheReadModel(tenantDB),
		sources:   adapters.NewOnePagerBuiltInFieldSources(tenantDB),
		subjects:  adapters.NewOnePagerSubjectExistenceChecker(tenantDB),
	}
}

func (f *cacheFixture) seed(subjectType, subjectID, name string, values map[string]any) {
	f.t.Helper()
	attributes := readmodels.SubjectAttributes{}
	for key, value := range values {
		require.NoError(f.t, attributes.Set(key, value))
	}
	require.NoError(f.t, f.index.Upsert(f.ctx, readmodels.SubjectIndexRecord{
		SubjectType: subjectType, SubjectID: subjectID, Name: name,
		CreatedAt: time.Now().UTC(), LastUpdatedAt: time.Now().UTC(), Attributes: attributes,
	}))
}

func (f *cacheFixture) relationEntryIDs(subjectType string) []string {
	f.t.Helper()
	st, err := valueobjects.NewSubjectType(subjectType)
	require.NoError(f.t, err)
	ids := []string{}
	for _, entry := range catalog.EntriesFor(st) {
		if entry.Relation {
			ids = append(ids, entry.ID)
		}
	}
	return ids
}

func TestBuiltInFieldSources_CatalogContract_Integration(t *testing.T) {
	f := newCacheFixture(t)

	for _, st := range valueobjects.AllSubjectTypes() {
		subjectType := st.Value()
		t.Run(subjectType, func(t *testing.T) {
			f.seed(subjectType, "s-"+subjectType, "Subject "+subjectType, nil)

			snapshot, err := f.sources[subjectType].FetchSubject(f.ctx, "s-"+subjectType, nil)

			require.NoError(t, err)
			require.NotNil(t, snapshot)
			assert.Equal(t, "Subject "+subjectType, snapshot.Name)
			for _, entry := range catalog.DefaultEntriesFor(st) {
				_, present := snapshot.Fields[entry.ID]
				assert.Truef(t, present, "catalog entry %q missing from snapshot fields", entry.ID)
			}
		})
	}
}

func TestBuiltInFieldSources_RelationCatalogContract_Integration(t *testing.T) {
	f := newCacheFixture(t)

	for _, st := range valueobjects.AllSubjectTypes() {
		subjectType := st.Value()
		t.Run(subjectType, func(t *testing.T) {
			f.seed(subjectType, "r-"+subjectType, "Subject", nil)

			for _, entryID := range f.relationEntryIDs(subjectType) {
				t.Run(entryID, func(t *testing.T) {
					snapshot, err := f.sources[subjectType].FetchSubject(f.ctx, "r-"+subjectType, []string{entryID})

					require.NoError(t, err)
					require.NotNil(t, snapshot)
					value, present := snapshot.Fields[entryID]
					require.Truef(t, present, "relation %q missing from snapshot fields", entryID)
					assert.IsTypef(t, ports.ReferenceListValue{}, value, "relation %q must resolve to a ReferenceListValue", entryID)
				})
			}
		})
	}
}

func TestBuiltInFieldSources_ResolvesRelationsAndAttributes_Integration(t *testing.T) {
	f := newCacheFixture(t)
	f.seed("capability", "cap-1", "Billing", map[string]any{"description": "Bills customers", "maturityValue": 62})
	f.seed("application", "app-1", "Billing Service", nil)
	require.NoError(t, f.relations.Save(f.ctx,
		readmodels.SubjectKey{SubjectType: "capability", SubjectID: "cap-1"},
		readmodels.RelationEntry{EntryID: "realizing-applications", RelatedType: "application", RelatedID: "app-1", EdgeID: "rz-1"},
	))

	snapshot, err := f.sources["capability"].FetchSubject(f.ctx, "cap-1", []string{"realizing-applications"})

	require.NoError(t, err)
	assert.Equal(t, ports.TextValue{Text: "Billing"}, snapshot.Fields["name"])
	assert.Equal(t, ports.TextValue{Text: "Bills customers"}, snapshot.Fields["description"])
	assert.Equal(t, ports.MaturityValue{Value: 62}, snapshot.Fields["maturity"])
	assert.Equal(t, ports.ReferenceListValue{References: []ports.Reference{
		{ID: "app-1", Label: "Billing Service", SubjectType: "application"},
	}}, snapshot.Fields["realizing-applications"])
}

func TestBuiltInFieldSources_FilledBuiltInFields_Integration(t *testing.T) {
	f := newCacheFixture(t)
	f.seed("application", "app-1", "Billing", map[string]any{
		"description": "Handles invoicing",
		"experts":     []readmodels.SubjectExpert{{Name: "Alice", Role: "Owner", Contact: "alice@dfds.com"}},
	})
	f.seed("application", "app-2", "Payments", nil)

	filled, err := f.sources["application"].FilledBuiltInFields(f.ctx, []string{"app-1", "app-2"}, []string{"description", "experts"})

	require.NoError(t, err)
	assert.Equal(t, map[string]map[string]bool{
		"app-1": {"description": true, "experts": true},
		"app-2": {"description": false, "experts": false},
	}, filled)
}

func TestBuiltInFieldSources_CountSubjectsWithBuiltInValue_Integration(t *testing.T) {
	f := newCacheFixture(t)
	f.seed("application", "app-1", "Billing", map[string]any{
		"experts": []readmodels.SubjectExpert{{Name: "Alice"}},
	})
	f.seed("application", "app-2", "Payments", nil)
	require.NoError(t, f.relations.Save(f.ctx,
		readmodels.SubjectKey{SubjectType: "application", SubjectID: "app-1"},
		readmodels.RelationEntry{EntryID: "built-by", RelatedType: "internal-team", RelatedID: "team-1", EdgeID: "app-1"},
	))

	population, err := f.sources["application"].CountSubjects(f.ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, population)

	withExperts, err := f.sources["application"].CountSubjectsWithBuiltInValue(f.ctx, "experts")
	require.NoError(t, err)
	assert.Equal(t, 1, withExperts)

	withTeam, err := f.sources["application"].CountSubjectsWithBuiltInValue(f.ctx, "built-by")
	require.NoError(t, err)
	assert.Equal(t, 1, withTeam)
}

func TestSubjectExistenceChecker_Integration(t *testing.T) {
	f := newCacheFixture(t)
	f.seed("vendor", "v-1", "Contoso", nil)

	found, err := f.subjects.SubjectExists(f.ctx, "vendor", "v-1")
	require.NoError(t, err)
	assert.True(t, found)

	missing, err := f.subjects.SubjectExists(f.ctx, "vendor", "v-missing")
	require.NoError(t, err)
	assert.False(t, missing)
}
