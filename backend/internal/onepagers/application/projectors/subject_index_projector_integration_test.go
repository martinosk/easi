//go:build integration

package projectors_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/projectors"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/infrastructure/adapters"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type indexProjectorFixture struct {
	t         *testing.T
	ctx       context.Context
	tenant    string
	tenantDB  *database.TenantAwareDB
	index     *readmodels.OnePagerSubjectIndexReadModel
	configs   *readmodels.OnePagerConfigurationReadModel
	projector *projectors.SubjectIndexProjector
	configID  string
	appID     string
}

func newIndexProjectorFixture(t *testing.T) *indexProjectorFixture {
	t.Helper()
	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi_app password=localdev dbname=easi sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	tenant := "test-spi-" + uuid.NewString()[:8]
	tenantID, err := sharedvo.NewTenantID(tenant)
	require.NoError(t, err)

	t.Cleanup(func() {
		for _, table := range []string{
			"onepagers.subject_relation_cache",
			"onepagers.one_pager_subject_index",
			"onepagers.one_pager_configurations",
			"infrastructure.events",
		} {
			_, _ = db.Exec("DELETE FROM "+table+" WHERE tenant_id = $1", tenant)
		}
		_ = db.Close()
	})

	tenantDB := database.NewTenantAwareDB(db)
	index := readmodels.NewOnePagerSubjectIndexReadModel(tenantDB)
	configs := readmodels.NewOnePagerConfigurationReadModel(tenantDB)
	counter := queries.NewCompletenessIndicators(configs, readmodels.NewOnePagerFactsReadModel(tenantDB),
		adapters.NewOnePagerBuiltInFieldSources(tenantDB))

	return &indexProjectorFixture{
		t: t, ctx: sharedctx.WithTenant(context.Background(), tenantID), tenant: tenant, tenantDB: tenantDB,
		index: index, configs: configs, appID: "app-e2e",
		projector: projectors.NewSubjectIndexProjector(index, counter, adapters.NewSubjectAuditAdapter(tenantDB), configs),
	}
}

func (f *indexProjectorFixture) seedCreationEvent(actorID, actorEmail string, at time.Time) {
	f.t.Helper()
	_, err := f.tenantDB.ExecContext(f.ctx,
		`INSERT INTO infrastructure.events (aggregate_id, event_type, event_data, version, occurred_at, tenant_id, actor_id, actor_email)
		VALUES ($1, $2, '{}'::jsonb, 1, $3, $4, $5, $6)`,
		f.appID, amPL.ApplicationComponentCreated, at, f.tenant, actorID, actorEmail,
	)
	require.NoError(f.t, err)
}

func descriptionConfigDocument(required bool) readmodels.ConfigurationDocument {
	return readmodels.ConfigurationDocument{
		CustomFields:  []readmodels.CustomFieldRecord{},
		BuiltInFields: []readmodels.BuiltInFieldRecord{{ID: "description", Required: required}},
		DisplayOrder:  []readmodels.FieldRefRecord{{Kind: "builtIn", ID: "description"}},
	}
}

func (f *indexProjectorFixture) seedConfig(descriptionRequired bool) {
	f.t.Helper()
	f.configID = uuid.NewString()
	require.NoError(f.t, f.configs.Insert(f.ctx, readmodels.ConfigurationRecord{
		ID: f.configID, SubjectType: "application", Document: descriptionConfigDocument(descriptionRequired), Version: 1,
		CreatedAt: time.Now().UTC(), ModifiedAt: time.Now().UTC(), ModifiedBy: "admin",
	}))
}

func (f *indexProjectorFixture) requireDescription(required bool) {
	f.t.Helper()
	require.NoError(f.t, f.configs.Update(f.ctx, readmodels.UpdateParams{
		ID: f.configID, Document: descriptionConfigDocument(required), Version: 2,
		ModifiedAt: time.Now().UTC(), ModifiedBy: "admin",
	}))
}

func (f *indexProjectorFixture) project(eventType string, at time.Time, payload map[string]any) {
	f.t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(f.t, err)
	require.NoError(f.t, f.projector.ProjectEvent(f.ctx, eventType, at, data))
}

func (f *indexProjectorFixture) row() (readmodels.SubjectIndexRecord, bool) {
	f.t.Helper()
	page, _, err := f.index.Page(f.ctx, readmodels.SubjectIndexQuery{
		SubjectTypes: []string{"application"}, Sort: readmodels.SortName, Order: readmodels.OrderAsc, Limit: 50,
	})
	require.NoError(f.t, err)
	for _, record := range page {
		if record.SubjectID == f.appID {
			return record, true
		}
	}
	return readmodels.SubjectIndexRecord{}, false
}

func TestSubjectIndexProjector_EndToEnd_OverOwnedCaches(t *testing.T) {
	f := newIndexProjectorFixture(t)
	created := time.Date(2026, 1, 2, 8, 0, 0, 0, time.UTC)
	f.seedCreationEvent("creator-1", "creator@dfds.com", created)
	f.seedConfig(true)

	f.project(amPL.ApplicationComponentCreated, created, map[string]any{"id": f.appID, "name": "Billing", "description": ""})
	row, found := f.row()
	require.True(t, found, "created subject appears as a row")
	assert.Equal(t, "Billing", row.Name)
	assert.Equal(t, "creator-1", row.CreatorActorID)
	assert.Equal(t, "creator@dfds.com", row.CreatorEmail)
	assert.True(t, created.Equal(row.CreatedAt))
	assert.Equal(t, readmodels.SignalIncomplete, row.Signal())

	f.project(amPL.ApplicationComponentUpdated, created.Add(time.Hour),
		map[string]any{"id": f.appID, "name": "Billing Platform", "description": "Handles payments"})
	row, _ = f.row()
	assert.Equal(t, "Billing Platform", row.Name, "rename refreshes the stored name")
	assert.Equal(t, readmodels.SignalComplete, row.Signal(), "the cached description makes the subject complete")

	f.requireDescription(false)
	f.project("BuiltInFieldRequirementChanged", created.Add(2*time.Hour), map[string]any{"id": f.configID})
	row, _ = f.row()
	assert.Equal(t, readmodels.SignalNotApplicable, row.Signal(), "no active required field is not applicable")

	f.project(amPL.ApplicationComponentDeleted, created.Add(3*time.Hour), map[string]any{"id": f.appID})
	_, found = f.row()
	assert.False(t, found, "deleted subject no longer appears")
}

func TestSubjectIndexProjector_ExpertsAccumulateInTheAttributeCache(t *testing.T) {
	f := newIndexProjectorFixture(t)
	at := time.Date(2026, 1, 2, 8, 0, 0, 0, time.UTC)
	f.seedConfig(false)
	f.project(amPL.ApplicationComponentCreated, at, map[string]any{"id": f.appID, "name": "Billing"})

	f.project(amPL.ApplicationComponentExpertAdded, at, map[string]any{
		"componentId": f.appID, "expertName": "Alice", "expertRole": "Owner", "contactInfo": "alice@dfds.com",
	})
	f.project(amPL.ApplicationComponentExpertAdded, at, map[string]any{
		"componentId": f.appID, "expertName": "Bob", "expertRole": "Lead", "contactInfo": "bob@dfds.com",
	})

	sources := adapters.NewOnePagerBuiltInFieldSources(f.tenantDB)
	snapshot, err := sources["application"].FetchSubject(f.ctx, f.appID, nil)
	require.NoError(t, err)
	assert.Equal(t, ports.ExpertsValue{Experts: []ports.Expert{
		{Name: "Alice", Role: "Owner", Contact: "alice@dfds.com"},
		{Name: "Bob", Role: "Lead", Contact: "bob@dfds.com"},
	}}, snapshot.Fields["experts"])

	f.project(amPL.ApplicationComponentExpertRemoved, at, map[string]any{
		"componentId": f.appID, "expertName": "Alice", "expertRole": "Owner", "contactInfo": "alice@dfds.com",
	})

	snapshot, err = sources["application"].FetchSubject(f.ctx, f.appID, nil)
	require.NoError(t, err)
	assert.Equal(t, ports.ExpertsValue{Experts: []ports.Expert{{Name: "Bob", Role: "Lead", Contact: "bob@dfds.com"}}}, snapshot.Fields["experts"])
}
