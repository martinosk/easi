//go:build integration
// +build integration

package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"easi/backend/internal/architecturedirection/application/handlers"
	"easi/backend/internal/architecturedirection/application/projectors"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/infrastructure/repositories"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	sharedcontext "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type standardApplicationTestContext struct {
	db         *sql.DB
	tenantDB   *database.TenantAwareDB
	handlers   *StandardApplicationHandlers
	readModel  *readmodels.StandardApplicationReadModel
	cleanupECs []string
}

func (tc *standardApplicationTestContext) trackEC(ecID string) {
	tc.cleanupECs = append(tc.cleanupECs, ecID)
}

func standardAppEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func setupStandardApplicationTestDB(t *testing.T) (*standardApplicationTestContext, func()) {
	t.Helper()
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		standardAppEnv("INTEGRATION_TEST_DB_HOST", "localhost"),
		standardAppEnv("INTEGRATION_TEST_DB_PORT", "5432"),
		standardAppEnv("INTEGRATION_TEST_DB_USER", "easi_app"),
		standardAppEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev"),
		standardAppEnv("INTEGRATION_TEST_DB_NAME", "easi"),
		standardAppEnv("INTEGRATION_TEST_DB_SSLMODE", "disable"),
	)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	readModel := readmodels.NewStandardApplicationReadModel(tenantDB)
	repo := repositories.NewStandardApplicationRepository(eventStore)

	projector := projectors.NewStandardApplicationProjector(readModel)
	eventBus.Subscribe(pl.StandardApplicationSet, projector)

	alwaysExists := func(_ context.Context, _ string) (bool, error) { return true, nil }
	refs := &services.ReferenceChecker{
		EnterpriseCapabilityExists: alwaysExists,
		PhysicalCapabilityExists:   alwaysExists,
		BusinessDomainExists:       alwaysExists,
	}
	commandBus.Register("SetStandardApplication", handlers.NewSetStandardApplicationHandler(repo, readModel, refs))

	links := NewStandardApplicationLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	httpHandlers := NewStandardApplicationHandlers(commandBus, readModel, links)

	ctx := &standardApplicationTestContext{
		db:        db,
		tenantDB:  tenantDB,
		handlers:  httpHandlers,
		readModel: readModel,
	}

	cleanup := func() {
		_, _ = db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", sharedvo.DefaultTenantID().Value()))
		for _, ecID := range ctx.cleanupECs {
			var aggID string
			err := db.QueryRow(
				"SELECT id FROM architecturedirection.standard_applications WHERE tenant_id = $1 AND enterprise_capability_id = $2",
				sharedvo.DefaultTenantID().Value(), ecID,
			).Scan(&aggID)
			if err != nil {
				continue
			}
			_, _ = db.Exec("DELETE FROM architecturedirection.standard_application_history WHERE standard_application_id = $1", aggID)
			_, _ = db.Exec("DELETE FROM architecturedirection.standard_applications WHERE id = $1", aggID)
			_, _ = db.Exec("DELETE FROM infrastructure.events WHERE aggregate_id = $1", aggID)
		}
		db.Close()
	}
	return ctx, cleanup
}

type putStandardRequest struct {
	handlers      *StandardApplicationHandlers
	ecID          string
	applicationID string
	narrative     string
}

func (p putStandardRequest) execute(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(SetStandardApplicationRequest{ApplicationID: p.applicationID, Narrative: p.narrative})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/enterprise-capabilities/"+p.ecID+"/standard-application", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return runStandardRequest(req, "/api/v1/enterprise-capabilities/{id}/standard-application", p.handlers.SetStandardApplication, http.MethodPut)
}

func getStandardEnvelopeIntegration(t *testing.T, h *StandardApplicationHandlers, ecID string) *httptest.ResponseRecorder {
	t.Helper()
	return runStandardRequest(
		httptest.NewRequest(http.MethodGet, "/api/v1/enterprise-capabilities/"+ecID+"/standard-application", nil),
		"/api/v1/enterprise-capabilities/{id}/standard-application",
		h.GetStandardApplicationForEnterpriseCapability,
		http.MethodGet,
	)
}

func getStandardHistoryIntegration(t *testing.T, h *StandardApplicationHandlers, ecID string) *httptest.ResponseRecorder {
	t.Helper()
	return runStandardRequest(
		httptest.NewRequest(http.MethodGet, "/api/v1/enterprise-capabilities/"+ecID+"/standard-application/history", nil),
		"/api/v1/enterprise-capabilities/{id}/standard-application/history",
		h.GetStandardApplicationHistory,
		http.MethodGet,
	)
}

func runStandardRequest(req *http.Request, routePattern string, handler http.HandlerFunc, method string) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	r.Method(method, routePattern, handler)
	req = withArchitectTestActor(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func withArchitectTestActor(req *http.Request) *http.Request {
	ctx := sharedcontext.WithTenant(req.Context(), sharedvo.DefaultTenantID())
	actor := sharedcontext.NewActor("test-architect", "architect@example.com", sharedcontext.RoleArchitect)
	ctx = sharedcontext.WithActor(ctx, actor)
	return req.WithContext(ctx)
}

func TestSetStandardApplication_FirstSet_PersistsAndReturns201(t *testing.T) {
	tc, cleanup := setupStandardApplicationTestDB(t)
	defer cleanup()

	ecID := uuid.New().String()
	appID := uuid.New().String()
	tc.trackEC(ecID)

	rec := putStandardRequest{handlers: tc.handlers, ecID: ecID, applicationID: appID, narrative: "covers operational and reporting layers"}.execute(t)

	require.Equal(t, http.StatusCreated, rec.Code, "first set must return 201; body=%s", rec.Body.String())
	assert.Equal(t, "/api/v1/enterprise-capabilities/"+ecID+"/standard-application", rec.Header().Get("Location"))

	var body readmodels.StandardApplicationDTO
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, ecID, body.EnterpriseCapabilityID)
	assert.Equal(t, appID, body.ApplicationID)
	assert.Equal(t, "covers operational and reporting layers", body.Narrative)
	require.NotNil(t, body.Links["self"])
	assert.Equal(t, http.MethodGet, body.Links["self"].Method)
	require.NotNil(t, body.Links["edit"])
	assert.Equal(t, http.MethodPut, body.Links["edit"].Method)
}

func TestSetStandardApplication_AggregateIDIsIndependentOfEnterpriseCapabilityID(t *testing.T) {
	tc, cleanup := setupStandardApplicationTestDB(t)
	defer cleanup()

	ecID := uuid.New().String()
	appID := uuid.New().String()
	tc.trackEC(ecID)

	rec := putStandardRequest{handlers: tc.handlers, ecID: ecID, applicationID: appID, narrative: "first set"}.execute(t)
	require.Equal(t, http.StatusCreated, rec.Code)

	_, _ = tc.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", sharedvo.DefaultTenantID().Value()))
	var aggregateID string
	require.NoError(t, tc.db.QueryRow(
		"SELECT id FROM architecturedirection.standard_applications WHERE tenant_id = $1 AND enterprise_capability_id = $2",
		sharedvo.DefaultTenantID().Value(), ecID,
	).Scan(&aggregateID))
	assert.NotEqual(t, ecID, aggregateID, "aggregate identity must be its own UUID, not the EC's")

	var ecStreamHits int
	require.NoError(t, tc.db.QueryRow(
		"SELECT COUNT(*) FROM infrastructure.events WHERE tenant_id = $1 AND aggregate_id = $2",
		sharedvo.DefaultTenantID().Value(), ecID,
	).Scan(&ecStreamHits))
	assert.Equal(t, 0, ecStreamHits, "no standard-application event must ever land in the EC's stream")
}

func TestSetStandardApplication_TwoConcurrentFirstSetsForSameEC_OnlyOneSucceeds(t *testing.T) {
	tc, cleanup := setupStandardApplicationTestDB(t)
	defer cleanup()

	ecID := uuid.New().String()
	tc.trackEC(ecID)

	first := putStandardRequest{handlers: tc.handlers, ecID: ecID, applicationID: uuid.New().String(), narrative: "first"}.execute(t)
	require.Equal(t, http.StatusCreated, first.Code)

	second := putStandardRequest{handlers: tc.handlers, ecID: ecID, applicationID: uuid.New().String(), narrative: "second"}.execute(t)
	require.Equal(t, http.StatusOK, second.Code,
		"a subsequent set for the same EC must route through the read-model lookup and load the existing aggregate (not create a second one)")

	_, _ = tc.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", sharedvo.DefaultTenantID().Value()))
	var aggCount int
	require.NoError(t, tc.db.QueryRow(
		"SELECT COUNT(*) FROM architecturedirection.standard_applications WHERE tenant_id = $1 AND enterprise_capability_id = $2",
		sharedvo.DefaultTenantID().Value(), ecID,
	).Scan(&aggCount))
	assert.Equal(t, 1, aggCount, "the per-EC uniqueness invariant must hold across multiple set commands")
}

func TestSetStandardApplication_Replacement_Returns200WithNewApplication(t *testing.T) {
	tc, cleanup := setupStandardApplicationTestDB(t)
	defer cleanup()

	ecID := uuid.New().String()
	appA := uuid.New().String()
	appB := uuid.New().String()
	tc.trackEC(ecID)

	require.Equal(t, http.StatusCreated, putStandardRequest{handlers: tc.handlers, ecID: ecID, applicationID: appA, narrative: "first"}.execute(t).Code)

	rec := putStandardRequest{handlers: tc.handlers, ecID: ecID, applicationID: appB, narrative: "replacement covers reporting layer only"}.execute(t)
	require.Equal(t, http.StatusOK, rec.Code, "replacement must be 200, not 201")
	assert.Empty(t, rec.Header().Get("Location"))

	getRec := getStandardEnvelopeIntegration(t, tc.handlers, ecID)
	require.Equal(t, http.StatusOK, getRec.Code)
	var env ECStandardApplicationResponse
	require.NoError(t, json.NewDecoder(getRec.Body).Decode(&env))
	require.NotNil(t, env.Standard)
	assert.Equal(t, appB, env.Standard.ApplicationID)
	assert.Equal(t, "replacement covers reporting layer only", env.Standard.Narrative)
}

func TestGetStandardApplicationHistory_ReturnsBothEntriesInOrder(t *testing.T) {
	tc, cleanup := setupStandardApplicationTestDB(t)
	defer cleanup()

	ecID := uuid.New().String()
	appA := uuid.New().String()
	appB := uuid.New().String()
	tc.trackEC(ecID)

	require.Equal(t, http.StatusCreated, putStandardRequest{handlers: tc.handlers, ecID: ecID, applicationID: appA, narrative: "first"}.execute(t).Code)
	time.Sleep(10 * time.Millisecond) // ensure distinct set_at timestamps
	require.Equal(t, http.StatusOK, putStandardRequest{handlers: tc.handlers, ecID: ecID, applicationID: appB, narrative: "second"}.execute(t).Code)

	rec := getStandardHistoryIntegration(t, tc.handlers, ecID)
	require.Equal(t, http.StatusOK, rec.Code)

	var body readmodels.StandardApplicationHistoryDTO
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Len(t, body.Entries, 2)
	assert.Equal(t, appB, body.Entries[0].ApplicationID, "history must be reverse-chronological")
	assert.Equal(t, appA, body.Entries[0].PreviousApplicationID)
	assert.Equal(t, "second", body.Entries[0].Narrative)
	assert.Equal(t, appA, body.Entries[1].ApplicationID)
	assert.Empty(t, body.Entries[1].PreviousApplicationID, "first entry has no previous application")
	assert.Equal(t, "first", body.Entries[1].Narrative)
}

func TestGetStandardApplicationHistory_NoStandard_Returns200WithEmptyEntries(t *testing.T) {
	tc, cleanup := setupStandardApplicationTestDB(t)
	defer cleanup()

	ecID := uuid.New().String()

	rec := getStandardHistoryIntegration(t, tc.handlers, ecID)
	require.Equal(t, http.StatusOK, rec.Code)

	var body readmodels.StandardApplicationHistoryDTO
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Empty(t, body.Entries)
	assert.Contains(t, body.Links, "self")
}

