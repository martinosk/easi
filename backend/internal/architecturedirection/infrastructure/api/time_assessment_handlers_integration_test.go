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
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	sharedcontext "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type timeAssessmentTestContext struct {
	db           *sql.DB
	tenantDB     *database.TenantAwareDB
	handlers     *TimeAssessmentHandlers
	readModel    *readmodels.TimeAssessmentReadModel
	eventBus     events.EventBus
	directExists map[string]string
	cleanupIDs   []string
}

func (tc *timeAssessmentTestContext) trackAssessment(pair timeAssessmentPairID) {
	tc.cleanupIDs = append(tc.cleanupIDs, pair.CapabilityID+"|"+pair.ComponentID)
}

func (tc *timeAssessmentTestContext) allowDirect(pair timeAssessmentPairID, realizationID string) {
	tc.directExists[pair.CapabilityID+"|"+pair.ComponentID] = realizationID
}

func timeAssessmentEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func setupTimeAssessmentTestDB(t *testing.T) (*timeAssessmentTestContext, func()) {
	t.Helper()
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		timeAssessmentEnv("INTEGRATION_TEST_DB_HOST", "localhost"),
		timeAssessmentEnv("INTEGRATION_TEST_DB_PORT", "5432"),
		timeAssessmentEnv("INTEGRATION_TEST_DB_USER", "easi_app"),
		timeAssessmentEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev"),
		timeAssessmentEnv("INTEGRATION_TEST_DB_NAME", "easi"),
		timeAssessmentEnv("INTEGRATION_TEST_DB_SSLMODE", "disable"),
	)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	readModel := readmodels.NewTimeAssessmentReadModel(tenantDB)
	repo := repositories.NewTimeAssessmentRepository(eventStore)

	projector := projectors.NewTimeAssessmentProjector(readModel)
	referenceProjector := projectors.NewTimeAssessmentReferenceProjector(readModel)
	reactor := projectors.NewTimeAssessmentDeletionReactor(readModel, commandBus)
	eventBus.Subscribe(pl.TimeAssessmentRecorded, projector)
	eventBus.Subscribe(pl.TimeAssessmentRemoved, projector)
	eventBus.Subscribe(cmPL.CapabilityDeleted, referenceProjector)
	eventBus.Subscribe(amPL.ApplicationComponentDeleted, referenceProjector)
	eventBus.Subscribe(cmPL.SystemRealizationDeleted, reactor)

	directExists := map[string]string{}
	directLookup := services.DirectRealizationLookup(func(_ context.Context, capID, compID string) (string, bool, error) {
		realizationID, exists := directExists[capID+"|"+compID]
		return realizationID, exists, nil
	})
	commandBus.Register("AssessRealization", handlers.NewAssessRealizationHandler(repo, readModel, directLookup))
	commandBus.Register("RemoveTimeAssessment", handlers.NewRemoveTimeAssessmentHandler(repo, readModel))

	links := NewTimeAssessmentLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	httpHandlers := NewTimeAssessmentHandlers(commandBus, readModel, links)

	ctx := &timeAssessmentTestContext{
		db:           db,
		tenantDB:     tenantDB,
		handlers:     httpHandlers,
		readModel:    readModel,
		eventBus:     eventBus,
		directExists: directExists,
	}

	cleanup := func() {
		_, _ = db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", sharedvo.DefaultTenantID().Value()))
		for _, pair := range ctx.cleanupIDs {
			var aggID string
			capID, compID := splitPair(pair)
			err := db.QueryRow(
				"SELECT id FROM architecturedirection.time_assessments WHERE tenant_id = $1 AND capability_id = $2 AND component_id = $3",
				sharedvo.DefaultTenantID().Value(), capID, compID,
			).Scan(&aggID)
			if err == nil {
				_, _ = db.Exec("DELETE FROM infrastructure.events WHERE aggregate_id = $1", aggID)
			}
			_, _ = db.Exec("DELETE FROM architecturedirection.time_assessments WHERE tenant_id = $1 AND capability_id = $2 AND component_id = $3",
				sharedvo.DefaultTenantID().Value(), capID, compID)
		}
		db.Close()
	}
	return ctx, cleanup
}

func splitPair(pair string) (string, string) {
	for i := range pair {
		if pair[i] == '|' {
			return pair[:i], pair[i+1:]
		}
	}
	return pair, ""
}

func withArchitectActor(req *http.Request) *http.Request {
	ctx := sharedcontext.WithTenant(req.Context(), sharedvo.DefaultTenantID())
	actor := sharedcontext.NewActor("test-architect", "architect@example.com", sharedcontext.RoleArchitect)
	ctx = sharedcontext.WithActor(ctx, actor)
	return req.WithContext(ctx)
}

func runTimeAssessmentRequest(req *http.Request, pattern string, handler http.HandlerFunc, method string) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	r.Method(method, pattern, handler)
	req = withArchitectActor(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

const timeAssessmentItemPattern = "/api/v1/capabilities/{id}/components/{componentId}/time-assessment"

type putTimeAssessmentRequest struct {
	handlers  *TimeAssessmentHandlers
	capID     string
	compID    string
	grade     string
	rationale string
}

func (p putTimeAssessmentRequest) execute(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(AssessRealizationRequest{Grade: p.grade, Rationale: p.rationale})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/capabilities/"+p.capID+"/components/"+p.compID+"/time-assessment", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return runTimeAssessmentRequest(req, timeAssessmentItemPattern, p.handlers.PutTimeAssessment, http.MethodPut)
}

func getTimeAssessmentReq(t *testing.T, h *TimeAssessmentHandlers, pair timeAssessmentPairID) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/"+pair.CapabilityID+"/components/"+pair.ComponentID+"/time-assessment", nil)
	return runTimeAssessmentRequest(req, timeAssessmentItemPattern, h.GetTimeAssessment, http.MethodGet)
}

func deleteTimeAssessmentReq(t *testing.T, h *TimeAssessmentHandlers, pair timeAssessmentPairID) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/capabilities/"+pair.CapabilityID+"/components/"+pair.ComponentID+"/time-assessment", nil)
	return runTimeAssessmentRequest(req, timeAssessmentItemPattern, h.DeleteTimeAssessment, http.MethodDelete)
}

func TestTimeAssessmentIntegration_AssessThenReassess_ReplacesGrade(t *testing.T) {
	tc, cleanup := setupTimeAssessmentTestDB(t)
	defer cleanup()

	capID, compID := uuid.New().String(), uuid.New().String()
	pair := timeAssessmentPairID{CapabilityID: capID, ComponentID: compID}
	tc.trackAssessment(pair)
	tc.allowDirect(pair, uuid.New().String())

	first := putTimeAssessmentRequest{handlers: tc.handlers, capID: capID, compID: compID, grade: "Tolerate", rationale: "first cut"}.execute(t)
	require.Equal(t, http.StatusCreated, first.Code, first.Body.String())

	second := putTimeAssessmentRequest{handlers: tc.handlers, capID: capID, compID: compID, grade: "Eliminate", rationale: "reconsidered"}.execute(t)
	require.Equal(t, http.StatusOK, second.Code, second.Body.String())

	get := getTimeAssessmentReq(t, tc.handlers, pair)
	require.Equal(t, http.StatusOK, get.Code)
	var dto readmodels.TimeAssessmentDTO
	require.NoError(t, json.NewDecoder(get.Body).Decode(&dto))
	assert.Equal(t, "Eliminate", dto.Grade)
	assert.Equal(t, "reconsidered", dto.Rationale)
	assert.False(t, dto.Stale)
}

func TestTimeAssessmentIntegration_NoDirectRealization_Fails(t *testing.T) {
	tc, cleanup := setupTimeAssessmentTestDB(t)
	defer cleanup()

	capID, compID := uuid.New().String(), uuid.New().String()

	rec := putTimeAssessmentRequest{handlers: tc.handlers, capID: capID, compID: compID, grade: "Migrate"}.execute(t)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestTimeAssessmentIntegration_HidesAssessment(t *testing.T) {
	tests := []struct {
		name string
		hide func(t *testing.T, tc *timeAssessmentTestContext, pair timeAssessmentPairID, realizationID string)
	}{
		{
			name: "explicit removal returns to unassessed",
			hide: func(t *testing.T, tc *timeAssessmentTestContext, pair timeAssessmentPairID, _ string) {
				del := deleteTimeAssessmentReq(t, tc.handlers, pair)
				require.Equal(t, http.StatusNoContent, del.Code)
			},
		},
		{
			name: "system realization deleted",
			hide: func(t *testing.T, tc *timeAssessmentTestContext, _ timeAssessmentPairID, realizationID string) {
				ctx := sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
				require.NoError(t, tc.eventBus.Publish(ctx, []domain.DomainEvent{systemRealizationDeletedTestEvent{realizationID: realizationID}}))
			},
		},
		{
			name: "capability deleted",
			hide: func(t *testing.T, tc *timeAssessmentTestContext, pair timeAssessmentPairID, _ string) {
				ctx := sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
				require.NoError(t, tc.readModel.DeleteByCapabilityID(ctx, pair.CapabilityID))
			},
		},
		{
			name: "component deleted",
			hide: func(t *testing.T, tc *timeAssessmentTestContext, pair timeAssessmentPairID, _ string) {
				ctx := sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
				require.NoError(t, tc.readModel.DeleteByComponentID(ctx, pair.ComponentID))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc, cleanup := setupTimeAssessmentTestDB(t)
			defer cleanup()

			capID, compID := uuid.New().String(), uuid.New().String()
			pair := timeAssessmentPairID{CapabilityID: capID, ComponentID: compID}
			realizationID := uuid.New().String()
			tc.trackAssessment(pair)
			tc.allowDirect(pair, realizationID)
			require.Equal(t, http.StatusCreated, putTimeAssessmentRequest{handlers: tc.handlers, capID: capID, compID: compID, grade: "Migrate"}.execute(t).Code)

			tt.hide(t, tc, pair, realizationID)

			get := getTimeAssessmentReq(t, tc.handlers, pair)
			assert.Equal(t, http.StatusNotFound, get.Code)
		})
	}
}

func TestTimeAssessmentIntegration_SystemRealizationDeleted_RemovesViaReactorWithAudit(t *testing.T) {
	tc, cleanup := setupTimeAssessmentTestDB(t)
	defer cleanup()

	capID, compID := uuid.New().String(), uuid.New().String()
	pair := timeAssessmentPairID{CapabilityID: capID, ComponentID: compID}
	realizationID := uuid.New().String()
	tc.trackAssessment(pair)
	tc.allowDirect(pair, realizationID)
	require.Equal(t, http.StatusCreated, putTimeAssessmentRequest{handlers: tc.handlers, capID: capID, compID: compID, grade: "Migrate"}.execute(t).Code)

	get := getTimeAssessmentReq(t, tc.handlers, pair)
	require.Equal(t, http.StatusOK, get.Code)
	var dto readmodels.TimeAssessmentDTO
	require.NoError(t, json.NewDecoder(get.Body).Decode(&dto))
	assessmentID := dto.ID

	ctx := sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
	require.NoError(t, tc.eventBus.Publish(ctx, []domain.DomainEvent{systemRealizationDeletedTestEvent{realizationID: realizationID}}))

	afterGet := getTimeAssessmentReq(t, tc.handlers, pair)
	assert.Equal(t, http.StatusNotFound, afterGet.Code, "the read-model row must be gone after the reactor removes the assessment")

	_, _ = tc.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", sharedvo.DefaultTenantID().Value()))
	var eventType, removedBy string
	err := tc.db.QueryRow(
		`SELECT event_type, event_data->>'removedBy' FROM infrastructure.events
		 WHERE aggregate_id = $1 AND event_type = $2`,
		assessmentID, pl.TimeAssessmentRemoved,
	).Scan(&eventType, &removedBy)
	require.NoError(t, err, "a TimeAssessmentRemoved event must be recorded for the audit trail")
	assert.Equal(t, pl.TimeAssessmentRemoved, eventType)
	assert.Equal(t, "system:realization-deleted", removedBy)
}

func TestTimeAssessmentIntegration_StaleWhenOlderThanTwelveMonths(t *testing.T) {
	tc, cleanup := setupTimeAssessmentTestDB(t)
	defer cleanup()

	capID, compID := uuid.New().String(), uuid.New().String()
	pair := timeAssessmentPairID{CapabilityID: capID, ComponentID: compID}
	tc.trackAssessment(pair)
	tc.allowDirect(pair, uuid.New().String())
	require.Equal(t, http.StatusCreated, putTimeAssessmentRequest{handlers: tc.handlers, capID: capID, compID: compID, grade: "Migrate"}.execute(t).Code)

	_, _ = tc.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", sharedvo.DefaultTenantID().Value()))
	thirteenMonthsAgo := time.Now().AddDate(0, -13, 0)
	_, err := tc.db.Exec(
		"UPDATE architecturedirection.time_assessments SET assessed_at = $1 WHERE tenant_id = $2 AND capability_id = $3 AND component_id = $4",
		thirteenMonthsAgo, sharedvo.DefaultTenantID().Value(), capID, compID,
	)
	require.NoError(t, err)

	get := getTimeAssessmentReq(t, tc.handlers, pair)
	require.Equal(t, http.StatusOK, get.Code)
	var dto readmodels.TimeAssessmentDTO
	require.NoError(t, json.NewDecoder(get.Body).Decode(&dto))
	assert.True(t, dto.Stale)
}

func TestTimeAssessmentIntegration_RollupCountsAcrossCapabilities(t *testing.T) {
	tc, cleanup := setupTimeAssessmentTestDB(t)
	defer cleanup()

	compID := uuid.New().String()
	grades := []string{"Invest", "Tolerate", "Migrate", "Eliminate"}
	for _, grade := range grades {
		capID := uuid.New().String()
		pair := timeAssessmentPairID{CapabilityID: capID, ComponentID: compID}
		tc.trackAssessment(pair)
		tc.allowDirect(pair, uuid.New().String())
		require.Equal(t, http.StatusCreated, putTimeAssessmentRequest{handlers: tc.handlers, capID: capID, compID: compID, grade: grade}.execute(t).Code)
	}

	ctx := sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
	rollups, err := tc.readModel.GetRollupsByComponentIDs(ctx, []string{compID})
	require.NoError(t, err)
	require.Len(t, rollups, 1)
	assert.Equal(t, 1, rollups[0].Counts.Invest)
	assert.Equal(t, 1, rollups[0].Counts.Tolerate)
	assert.Equal(t, 1, rollups[0].Counts.Migrate)
	assert.Equal(t, 1, rollups[0].Counts.Eliminate)
}

func TestTimeAssessmentIntegration_UniqueConstraintBackstop_SecondFirstAssessRejected(t *testing.T) {
	tc, cleanup := setupTimeAssessmentTestDB(t)
	defer cleanup()

	capID, compID := uuid.New().String(), uuid.New().String()
	tc.trackAssessment(timeAssessmentPairID{CapabilityID: capID, ComponentID: compID})

	ctx := sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
	first := readmodels.UpsertTimeAssessmentParams{
		ID: uuid.New().String(), CapabilityID: capID, ComponentID: compID,
		RealizationID: uuid.New().String(), Grade: "Invest", AssessedBy: "a@example.com", AssessedAt: time.Now(),
	}
	require.NoError(t, tc.readModel.UpsertCurrent(ctx, first))

	second := first
	second.ID = uuid.New().String()
	err := tc.readModel.UpsertCurrent(ctx, second)

	assert.ErrorIs(t, err, readmodels.ErrTimeAssessmentAlreadyExists,
		"a second aggregate racing to claim the same (capability, component) pair must be rejected by the DB unique constraint")
}
