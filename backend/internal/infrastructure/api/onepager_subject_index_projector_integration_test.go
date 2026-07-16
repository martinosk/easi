//go:build integration
// +build integration

package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	opProjectors "easi/backend/internal/onepagers/application/projectors"
	opQueries "easi/backend/internal/onepagers/application/queries"
	opReadModels "easi/backend/internal/onepagers/application/readmodels"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type indexProjectorHarness struct {
	t         *testing.T
	db        *sql.DB
	tenantDB  *database.TenantAwareDB
	ctx       context.Context
	tenant    string
	store     *opReadModels.OnePagerSubjectIndexReadModel
	configs   *opReadModels.OnePagerConfigurationReadModel
	projector *opProjectors.SubjectIndexProjector
	configID  string
	appID     string
}

func newIndexProjectorHarness(t *testing.T) *indexProjectorHarness {
	t.Helper()
	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi_app password=localdev dbname=easi sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	tenant := "test-spi-" + uuid.NewString()[:8]
	tid, err := sharedvo.NewTenantID(tenant)
	require.NoError(t, err)
	ctx := sharedctx.WithTenant(context.Background(), tid)
	tenantDB := database.NewTenantAwareDB(db)

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM onepagers.one_pager_subject_index WHERE tenant_id = $1", tenant)
		_, _ = db.Exec("DELETE FROM onepagers.one_pager_configurations WHERE tenant_id = $1", tenant)
		_, _ = db.Exec("DELETE FROM architecturemodeling.application_components WHERE tenant_id = $1", tenant)
		_, _ = db.Exec("DELETE FROM infrastructure.events WHERE tenant_id = $1", tenant)
		_ = db.Close()
	})

	store := opReadModels.NewOnePagerSubjectIndexReadModel(tenantDB)
	configs := opReadModels.NewOnePagerConfigurationReadModel(tenantDB)
	counter := opQueries.NewCompletenessIndicators(configs, opReadModels.NewOnePagerFactsReadModel(tenantDB), newOnePagerBuiltInFieldSources(tenantDB))
	projector := opProjectors.NewSubjectIndexProjector(store, counter, newOnePagerAuditAdapter(tenantDB), configs)

	return &indexProjectorHarness{
		t: t, db: db, tenantDB: tenantDB, ctx: ctx, tenant: tenant, appID: "app-e2e",
		store: store, configs: configs, projector: projector,
	}
}

func (h *indexProjectorHarness) seedApplication(name, description string) {
	_, err := h.tenantDB.ExecContext(h.ctx,
		`INSERT INTO architecturemodeling.application_components (id, tenant_id, name, description, created_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, FALSE)`,
		h.appID, h.tenant, name, description, time.Now().UTC(),
	)
	require.NoError(h.t, err)
}

func (h *indexProjectorHarness) setDescription(description string) {
	_, err := h.tenantDB.ExecContext(h.ctx,
		`UPDATE architecturemodeling.application_components SET description = $1 WHERE tenant_id = $2 AND id = $3`,
		description, h.tenant, h.appID,
	)
	require.NoError(h.t, err)
}

func (h *indexProjectorHarness) seedCreationEvent(actor creator, at time.Time) {
	_, err := h.tenantDB.ExecContext(h.ctx,
		`INSERT INTO infrastructure.events (aggregate_id, event_type, event_data, version, occurred_at, tenant_id, actor_id, actor_email)
		VALUES ($1, $2, '{}'::jsonb, 1, $3, $4, $5, $6)`,
		h.appID, amPL.ApplicationComponentCreated, at, h.tenant, actor.id, actor.email,
	)
	require.NoError(h.t, err)
}

type creator struct {
	id    string
	email string
}

func descriptionConfigDoc(required bool) opReadModels.ConfigurationDocument {
	return opReadModels.ConfigurationDocument{
		CustomFields:  []opReadModels.CustomFieldRecord{},
		BuiltInFields: []opReadModels.BuiltInFieldRecord{{ID: "description", Required: required}},
		DisplayOrder:  []opReadModels.FieldRefRecord{{Kind: "builtIn", ID: "description"}},
	}
}

func (h *indexProjectorHarness) seedConfig(descriptionRequired bool) {
	h.configID = uuid.NewString()
	require.NoError(h.t, h.configs.Insert(h.ctx, opReadModels.ConfigurationRecord{
		ID: h.configID, SubjectType: "application", Document: descriptionConfigDoc(descriptionRequired), Version: 1,
		CreatedAt: time.Now().UTC(), ModifiedAt: time.Now().UTC(), ModifiedBy: "admin",
	}))
}

func (h *indexProjectorHarness) setDescriptionRequirement(required bool) {
	require.NoError(h.t, h.configs.Update(h.ctx, opReadModels.UpdateParams{
		ID: h.configID, Document: descriptionConfigDoc(required), Version: 2, ModifiedAt: time.Now().UTC(), ModifiedBy: "admin",
	}))
}

func (h *indexProjectorHarness) project(eventType string, at time.Time, payload map[string]any) {
	data, err := json.Marshal(payload)
	require.NoError(h.t, err)
	require.NoError(h.t, h.projector.ProjectEvent(h.ctx, eventType, at, data))
}

func (h *indexProjectorHarness) row() (opReadModels.SubjectIndexRecord, bool) {
	page, _, err := h.store.Page(h.ctx, opReadModels.SubjectIndexQuery{
		SubjectTypes: []string{"application"}, Sort: opReadModels.SortName, Order: opReadModels.OrderAsc, Limit: 50,
	})
	require.NoError(h.t, err)
	for _, r := range page {
		if r.SubjectID == h.appID {
			return r, true
		}
	}
	return opReadModels.SubjectIndexRecord{}, false
}

func TestSubjectIndexProjector_EndToEnd(t *testing.T) {
	h := newIndexProjectorHarness(t)
	created := time.Date(2026, 1, 2, 8, 0, 0, 0, time.UTC)
	h.seedApplication("Billing", "")
	h.seedCreationEvent(creator{id: "creator-1", email: "creator@dfds.com"}, created)
	h.seedConfig(true)

	h.project(amPL.ApplicationComponentCreated, created, map[string]any{"id": h.appID, "name": "Billing"})
	row, ok := h.row()
	require.True(t, ok, "created subject appears as a row")
	assert.Equal(t, "Billing", row.Name)
	assert.Equal(t, "creator-1", row.CreatorActorID)
	assert.Equal(t, "creator@dfds.com", row.CreatorEmail)
	assert.True(t, created.Equal(row.CreatedAt))
	assert.Equal(t, opReadModels.SignalIncomplete, row.Signal())

	h.setDescription("Handles payments")
	h.project(amPL.ApplicationComponentUpdated, created.Add(time.Hour), map[string]any{"id": h.appID, "name": "Billing Platform"})
	row, _ = h.row()
	assert.Equal(t, "Billing Platform", row.Name, "rename refreshes the stored name")
	assert.Equal(t, opReadModels.SignalComplete, row.Signal())

	h.setDescriptionRequirement(false)
	h.project("BuiltInFieldRequirementChanged", created.Add(2*time.Hour), map[string]any{"id": h.configID})
	row, _ = h.row()
	assert.Equal(t, opReadModels.SignalNotApplicable, row.Signal(), "no active required field is not applicable")

	h.project(amPL.ApplicationComponentDeleted, created.Add(3*time.Hour), map[string]any{"id": h.appID})
	_, ok = h.row()
	assert.False(t, ok, "deleted subject no longer appears")
}
