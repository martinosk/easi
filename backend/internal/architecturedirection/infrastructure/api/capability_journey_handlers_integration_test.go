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
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/architecturedirection/infrastructure/repositories"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	authPL "easi/backend/internal/auth/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	sharedcontext "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capabilityJourneyTestContext struct {
	db                   *sql.DB
	tenantDB             *database.TenantAwareDB
	handlers             *CapabilityJourneyHandlers
	readModel            *readmodels.CapabilityJourneyReadModel
	eventBus             events.EventBus
	cleanupCapabilityIDs []string
}

func (tc *capabilityJourneyTestContext) trackCapability(capabilityID string) {
	tc.cleanupCapabilityIDs = append(tc.cleanupCapabilityIDs, capabilityID)
}

func capabilityJourneyEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func setupCapabilityJourneyTestDB(t *testing.T) (*capabilityJourneyTestContext, func()) {
	t.Helper()
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		capabilityJourneyEnv("INTEGRATION_TEST_DB_HOST", "localhost"),
		capabilityJourneyEnv("INTEGRATION_TEST_DB_PORT", "5432"),
		capabilityJourneyEnv("INTEGRATION_TEST_DB_USER", "easi_app"),
		capabilityJourneyEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev"),
		capabilityJourneyEnv("INTEGRATION_TEST_DB_NAME", "easi"),
		capabilityJourneyEnv("INTEGRATION_TEST_DB_SSLMODE", "disable"),
	)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	readModel := readmodels.NewCapabilityJourneyReadModel(tenantDB)
	repo := repositories.NewCapabilityJourneyRepository(eventStore)

	subscribeCapabilityJourneyTestEvents(eventBus, readModel)
	registerCapabilityJourneyTestCommands(commandBus, repo, readModel)

	links := NewCapabilityJourneyLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	httpHandlers := NewCapabilityJourneyHandlers(commandBus, readModel, links)

	ctx := &capabilityJourneyTestContext{
		db:        db,
		tenantDB:  tenantDB,
		handlers:  httpHandlers,
		readModel: readModel,
		eventBus:  eventBus,
	}

	cleanup := func() {
		cleanupCapabilityJourneyTestData(db, ctx.cleanupCapabilityIDs)
		_ = db.Close()
	}
	return ctx, cleanup
}

func subscribeCapabilityJourneyTestEvents(eventBus events.EventBus, readModel *readmodels.CapabilityJourneyReadModel) {
	projector := projectors.NewCapabilityJourneyProjector(readModel)
	eventBus.Subscribe(pl.JourneyPlanned, projector)
	eventBus.Subscribe(pl.JourneyStarted, projector)
	eventBus.Subscribe(pl.JourneyCompleted, projector)
	eventBus.Subscribe(pl.JourneyAbandoned, projector)
	eventBus.Subscribe(pl.JourneyProgressUpdated, projector)
	eventBus.Subscribe(pl.JourneyDetailsUpdated, projector)
	eventBus.Subscribe(pl.JourneySourceApplicationsChanged, projector)
	eventBus.Subscribe(pl.JourneyMilestoneAdded, projector)
	eventBus.Subscribe(pl.JourneyMilestoneUpdated, projector)
	eventBus.Subscribe(pl.JourneyMilestoneRemoved, projector)

	referenceProjector := projectors.NewCapabilityJourneyReferenceProjector(readModel)
	eventBus.Subscribe(cmPL.CapabilityCreated, referenceProjector)
	eventBus.Subscribe(cmPL.CapabilityUpdated, referenceProjector)
	eventBus.Subscribe(cmPL.CapabilityDeleted, referenceProjector)
	eventBus.Subscribe(cmPL.BusinessDomainCreated, referenceProjector)
	eventBus.Subscribe(cmPL.BusinessDomainUpdated, referenceProjector)
	eventBus.Subscribe(cmPL.BusinessDomainDeleted, referenceProjector)
	eventBus.Subscribe(amPL.ApplicationComponentCreated, referenceProjector)
	eventBus.Subscribe(amPL.ApplicationComponentUpdated, referenceProjector)
	eventBus.Subscribe(amPL.ApplicationComponentDeleted, referenceProjector)
	eventBus.Subscribe(authPL.UserCreated, referenceProjector)
}

func registerCapabilityJourneyTestCommands(commandBus cqrs.CommandBus, repo *repositories.CapabilityJourneyRepository, readModel *readmodels.CapabilityJourneyReadModel) {
	refs := handlers.JourneyReferenceChecks{
		CapabilityExists:              services.CapabilityExists(func(context.Context, string) (bool, error) { return true, nil }),
		ComponentExists:               services.ComponentExists(func(context.Context, string) (bool, error) { return true, nil }),
		DomainExists:                  services.DomainExists(func(context.Context, string) (bool, error) { return true, nil }),
		CapabilityEffectivelyInDomain: services.CapabilityEffectivelyInDomain(func(context.Context, string, string) (bool, error) { return true, nil }),
	}
	commandBus.Register("PlanJourney", handlers.NewPlanJourneyHandler(repo, readModel, refs))
	commandBus.Register("StartJourney", handlers.NewStartJourneyHandler(repo))
	commandBus.Register("CompleteJourney", handlers.NewCompleteJourneyHandler(repo))
	commandBus.Register("AbandonJourney", handlers.NewAbandonJourneyHandler(repo))
	commandBus.Register("UpdateJourneyProgress", handlers.NewUpdateJourneyProgressHandler(repo))
	commandBus.Register("UpdateJourneyDetails", handlers.NewUpdateJourneyDetailsHandler(repo))
	commandBus.Register("ChangeJourneySourceApplications", handlers.NewChangeJourneySourceApplicationsHandler(repo, refs.ComponentExists))
	commandBus.Register("AddJourneyMilestone", handlers.NewAddJourneyMilestoneHandler(repo))
	commandBus.Register("UpdateJourneyMilestone", handlers.NewUpdateJourneyMilestoneHandler(repo))
	commandBus.Register("RemoveJourneyMilestone", handlers.NewRemoveJourneyMilestoneHandler(repo))
}

func cleanupCapabilityJourneyTestData(db *sql.DB, capabilityIDs []string) {
	_, _ = db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", sharedvo.DefaultTenantID().Value()))
	for _, capID := range capabilityIDs {
		cleanupCapabilityJourneysForCapability(db, capID)
	}
}

func cleanupCapabilityJourneysForCapability(db *sql.DB, capID string) {
	tenantID := sharedvo.DefaultTenantID().Value()
	for _, jid := range journeyIDsForCapability(db, tenantID, capID) {
		_, _ = db.Exec("DELETE FROM infrastructure.events WHERE aggregate_id = $1", jid)
		_, _ = db.Exec("DELETE FROM architecturedirection.capability_journey_sources WHERE tenant_id = $1 AND journey_id = $2", tenantID, jid)
		_, _ = db.Exec("DELETE FROM architecturedirection.capability_journey_milestones WHERE tenant_id = $1 AND journey_id = $2", tenantID, jid)
	}
	_, _ = db.Exec("DELETE FROM architecturedirection.capability_journeys WHERE tenant_id = $1 AND capability_id = $2", tenantID, capID)
}

func journeyIDsForCapability(db *sql.DB, tenantID, capID string) []string {
	rows, err := db.Query(
		"SELECT id FROM architecturedirection.capability_journeys WHERE tenant_id = $1 AND capability_id = $2",
		tenantID, capID,
	)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()
	var journeyIDs []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			journeyIDs = append(journeyIDs, id)
		}
	}
	return journeyIDs
}

func runCapabilityJourneyRequest(req *http.Request, pattern string, handler http.HandlerFunc, actor sharedcontext.Actor) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	r.Method(req.Method, pattern, handler)
	ctx := sharedcontext.WithTenant(req.Context(), sharedvo.DefaultTenantID())
	ctx = sharedcontext.WithActor(ctx, actor)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

const journeyItemPattern = "/api/v1/capabilities/{id}/journey"
const journeyTransitionPattern = "/api/v1/capability-journeys/{journeyId}/start"
const journeyCompletePattern = "/api/v1/capability-journeys/{journeyId}/complete"
const journeyAbandonPattern = "/api/v1/capability-journeys/{journeyId}/abandon"
const journeyProgressPattern = "/api/v1/capability-journeys/{journeyId}/progress"
const journeyMilestonesPattern = "/api/v1/capability-journeys/{journeyId}/milestones"

func captureJourneyReq(h *CapabilityJourneyHandlers, capID string, req CaptureJourneyRequest, actor sharedcontext.Actor) *httptest.ResponseRecorder {
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities/"+capID+"/journey", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	return runCapabilityJourneyRequest(httpReq, journeyItemPattern, h.CaptureJourney, actor)
}

func getJourneyForCapabilityReq(h *CapabilityJourneyHandlers, capID string, actor sharedcontext.Actor) *httptest.ResponseRecorder {
	httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/"+capID+"/journey", nil)
	return runCapabilityJourneyRequest(httpReq, journeyItemPattern, h.GetJourneyForCapability, actor)
}

func decodeJourneyEnvelope(t *testing.T, rec *httptest.ResponseRecorder) CapabilityJourneyResponse {
	t.Helper()
	var envelope CapabilityJourneyResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&envelope))
	return envelope
}

func TestCapabilityJourneyIntegration_FullLifecycle_CaptureStartProgressMilestoneComplete(t *testing.T) {
	tc, cleanup := setupCapabilityJourneyTestDB(t)
	defer cleanup()

	capID := uuid.New().String()
	tc.trackCapability(capID)
	fromApp := uuid.New().String()
	toApp := uuid.New().String()

	captureRec := captureJourneyReq(tc.handlers, capID, CaptureJourneyRequest{
		Kind:             valueobjects.JourneyKindMigration,
		FromComponentIDs: []string{fromApp},
		ToComponentID:    toApp,
		Note:             "moving to a modern platform",
		TargetPeriod:     &TargetPeriodRequest{Year: 2027, Quarter: 2},
	}, architectActor())
	require.Equal(t, http.StatusCreated, captureRec.Code, captureRec.Body.String())

	var captured readmodels.CapabilityJourneyDTO
	require.NoError(t, json.NewDecoder(captureRec.Body).Decode(&captured))
	assert.Equal(t, valueobjects.JourneyStatusPlanned, captured.Status)
	journeyID := captured.ID

	startRec := runCapabilityJourneyRequest(
		httptest.NewRequest(http.MethodPost, "/api/v1/capability-journeys/"+journeyID+"/start", nil),
		journeyTransitionPattern, tc.handlers.StartJourney, architectActor())
	require.Equal(t, http.StatusOK, startRec.Code, startRec.Body.String())

	progressBody, _ := json.Marshal(UpdateJourneyProgressRequest{Progress: 60})
	progressRec := runCapabilityJourneyRequest(
		httptest.NewRequest(http.MethodPut, "/api/v1/capability-journeys/"+journeyID+"/progress", bytes.NewReader(progressBody)),
		journeyProgressPattern, tc.handlers.PutJourneyProgress, architectActor())
	require.Equal(t, http.StatusOK, progressRec.Code, progressRec.Body.String())

	milestoneBody, _ := json.Marshal(AddJourneyMilestoneRequest{Label: "Cut over region A", Status: valueobjects.MilestoneStatusDone})
	milestoneRec := runCapabilityJourneyRequest(
		httptest.NewRequest(http.MethodPost, "/api/v1/capability-journeys/"+journeyID+"/milestones", bytes.NewReader(milestoneBody)),
		journeyMilestonesPattern, tc.handlers.PostJourneyMilestone, architectActor())
	require.Equal(t, http.StatusCreated, milestoneRec.Code, milestoneRec.Body.String())

	completeRec := runCapabilityJourneyRequest(
		httptest.NewRequest(http.MethodPost, "/api/v1/capability-journeys/"+journeyID+"/complete", nil),
		journeyCompletePattern, tc.handlers.CompleteJourney, architectActor())
	require.Equal(t, http.StatusOK, completeRec.Code, completeRec.Body.String())

	var final readmodels.CapabilityJourneyDTO
	require.NoError(t, json.NewDecoder(completeRec.Body).Decode(&final))
	assert.Equal(t, valueobjects.JourneyStatusDone, final.Status)
	require.NotNil(t, final.Progress)
	assert.Equal(t, 60, *final.Progress)
	require.Len(t, final.Milestones, 1)
	assert.Equal(t, "Cut over region A", final.Milestones[0].Label)
	assert.Equal(t, valueobjects.MilestoneStatusDone, final.Milestones[0].Status)
	assert.NotContains(t, final.Links, "edit", "terminal journeys must not carry write affordances")
}

func TestCapabilityJourneyIntegration_OneActivePerCapability_RejectedWithExistingID(t *testing.T) {
	tc, cleanup := setupCapabilityJourneyTestDB(t)
	defer cleanup()

	capID := uuid.New().String()
	tc.trackCapability(capID)

	first := captureJourneyReq(tc.handlers, capID, CaptureJourneyRequest{
		Kind: valueobjects.JourneyKindMigration, FromComponentIDs: []string{uuid.New().String()}, ToComponentID: uuid.New().String(),
	}, architectActor())
	require.Equal(t, http.StatusCreated, first.Code, first.Body.String())
	var firstJourney readmodels.CapabilityJourneyDTO
	require.NoError(t, json.NewDecoder(first.Body).Decode(&firstJourney))

	second := captureJourneyReq(tc.handlers, capID, CaptureJourneyRequest{
		Kind: valueobjects.JourneyKindMigration, FromComponentIDs: []string{uuid.New().String()}, ToComponentID: uuid.New().String(),
	}, architectActor())
	require.Equal(t, http.StatusConflict, second.Code)
	assert.Contains(t, second.Body.String(), firstJourney.ID)

	abandonRec := runCapabilityJourneyRequest(
		httptest.NewRequest(http.MethodPost, "/api/v1/capability-journeys/"+firstJourney.ID+"/abandon", nil),
		journeyAbandonPattern, tc.handlers.AbandonJourney, architectActor())
	require.Equal(t, http.StatusOK, abandonRec.Code, abandonRec.Body.String())

	third := captureJourneyReq(tc.handlers, capID, CaptureJourneyRequest{
		Kind: valueobjects.JourneyKindMigration, FromComponentIDs: []string{uuid.New().String()}, ToComponentID: uuid.New().String(),
	}, architectActor())
	require.Equal(t, http.StatusCreated, third.Code, third.Body.String(), "recapture after abandonment must succeed")
}

func TestCapabilityJourneyIntegration_DBPartialUniqueIndex_BlocksConcurrentActiveInsert(t *testing.T) {
	tc, cleanup := setupCapabilityJourneyTestDB(t)
	defer cleanup()

	capID := uuid.New().String()
	tc.trackCapability(capID)
	ctx := sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())

	err := tc.readModel.InsertJourney(ctx, readmodels.InsertJourneyParams{
		ID: uuid.New().String(), CapabilityID: capID, Kind: valueobjects.JourneyKindMigration,
		FromComponentIDs: []string{uuid.New().String()}, ToComponentID: uuid.New().String(), PlannedBy: "a@example.com",
	})
	require.NoError(t, err)

	err = tc.readModel.InsertJourney(ctx, readmodels.InsertJourneyParams{
		ID: uuid.New().String(), CapabilityID: capID, Kind: valueobjects.JourneyKindMigration,
		FromComponentIDs: []string{uuid.New().String()}, ToComponentID: uuid.New().String(), PlannedBy: "a@example.com",
	})
	require.ErrorIs(t, err, readmodels.ErrActiveCapabilityJourneyExists, "the partial unique index must backstop the handler-level check")
}

func TestCapabilityJourneyIntegration_MoveJourney_CapturesDestination(t *testing.T) {
	tc, cleanup := setupCapabilityJourneyTestDB(t)
	defer cleanup()

	capID := uuid.New().String()
	tc.trackCapability(capID)
	domainID := uuid.New().String()
	parentID := uuid.New().String()

	rec := captureJourneyReq(tc.handlers, capID, CaptureJourneyRequest{
		Kind:           valueobjects.JourneyKindMove,
		ToComponentID:  uuid.New().String(),
		Note:           "relocating to group functions",
		TargetDomainID: domainID,
		TargetParentID: parentID,
		ResultingName:  "Freight invoicing",
	}, architectActor())
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var dto readmodels.CapabilityJourneyDTO
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&dto))
	require.NotNil(t, dto.Move)
	assert.Equal(t, domainID, dto.Move.TargetDomainID)
	assert.Equal(t, parentID, dto.Move.TargetParentID)
	assert.Equal(t, "Freight invoicing", dto.Move.ResultingName)

	startRec := runCapabilityJourneyRequest(
		httptest.NewRequest(http.MethodPost, "/api/v1/capability-journeys/"+dto.ID+"/start", nil),
		journeyTransitionPattern, tc.handlers.StartJourney, architectActor())
	require.Equal(t, http.StatusOK, startRec.Code, startRec.Body.String())

	completeRec := runCapabilityJourneyRequest(
		httptest.NewRequest(http.MethodPost, "/api/v1/capability-journeys/"+dto.ID+"/complete", nil),
		journeyCompletePattern, tc.handlers.CompleteJourney, architectActor())
	require.Equal(t, http.StatusOK, completeRec.Code, completeRec.Body.String())

	var noRows int
	require.NoError(t, tc.db.QueryRow("SELECT COUNT(*) FROM capabilitymapping.capabilities WHERE id = $1", capID).Scan(&noRows))
	assert.Equal(t, 0, noRows, "completing a move journey must not touch capabilitymapping")
}

func TestCapabilityJourneyIntegration_CompletingJourney_TouchesNothingOutsideOwnAggregate(t *testing.T) {
	tc, cleanup := setupCapabilityJourneyTestDB(t)
	defer cleanup()

	capID := uuid.New().String()
	tc.trackCapability(capID)

	captureRec := captureJourneyReq(tc.handlers, capID, CaptureJourneyRequest{
		Kind: valueobjects.JourneyKindMigration, FromComponentIDs: []string{uuid.New().String()}, ToComponentID: uuid.New().String(),
	}, architectActor())
	require.Equal(t, http.StatusCreated, captureRec.Code)
	var captured readmodels.CapabilityJourneyDTO
	require.NoError(t, json.NewDecoder(captureRec.Body).Decode(&captured))

	startRec := runCapabilityJourneyRequest(
		httptest.NewRequest(http.MethodPost, "/api/v1/capability-journeys/"+captured.ID+"/start", nil),
		journeyTransitionPattern, tc.handlers.StartJourney, architectActor())
	require.Equal(t, http.StatusOK, startRec.Code)

	_, _ = tc.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", sharedvo.DefaultTenantID().Value()))
	var eventCountBefore int
	require.NoError(t, tc.db.QueryRow("SELECT COUNT(*) FROM infrastructure.events WHERE aggregate_id = $1", captured.ID).Scan(&eventCountBefore))
	assert.Equal(t, 2, eventCountBefore, "planned + started")

	completeRec := runCapabilityJourneyRequest(
		httptest.NewRequest(http.MethodPost, "/api/v1/capability-journeys/"+captured.ID+"/complete", nil),
		journeyCompletePattern, tc.handlers.CompleteJourney, architectActor())
	require.Equal(t, http.StatusOK, completeRec.Code)

	var eventCountAfter int
	require.NoError(t, tc.db.QueryRow("SELECT COUNT(*) FROM infrastructure.events WHERE aggregate_id = $1", captured.ID).Scan(&eventCountAfter))
	assert.Equal(t, 3, eventCountAfter, "planned + started + completed — exactly one new event for the aggregate")

	var eventsForCapability int
	require.NoError(t, tc.db.QueryRow("SELECT COUNT(*) FROM infrastructure.events WHERE aggregate_id = $1", capID).Scan(&eventsForCapability))
	assert.Equal(t, 0, eventsForCapability, "the capability itself must never receive an event stream from journey completion")

	var capabilityRows int
	require.NoError(t, tc.db.QueryRow("SELECT COUNT(*) FROM capabilitymapping.capabilities WHERE id = $1", capID).Scan(&capabilityRows))
	assert.Equal(t, 0, capabilityRows, "completion is plan-only and must not create or touch capabilitymapping rows")
}

func TestCapabilityJourneyIntegration_ReadOnlyActor_NoWriteAffordances(t *testing.T) {
	tc, cleanup := setupCapabilityJourneyTestDB(t)
	defer cleanup()

	capID := uuid.New().String()
	tc.trackCapability(capID)
	require.Equal(t, http.StatusCreated, captureJourneyReq(tc.handlers, capID, CaptureJourneyRequest{
		Kind: valueobjects.JourneyKindMigration, FromComponentIDs: []string{uuid.New().String()}, ToComponentID: uuid.New().String(),
	}, architectActor()).Code)

	architectGet := getJourneyForCapabilityReq(tc.handlers, capID, architectActor())
	require.Equal(t, http.StatusOK, architectGet.Code)
	architectEnvelope := decodeJourneyEnvelope(t, architectGet)
	require.NotNil(t, architectEnvelope.Journey)
	assert.Contains(t, architectEnvelope.Journey.Links, "x-start")

	readOnlyGet := getJourneyForCapabilityReq(tc.handlers, capID, stakeholderActor())
	require.Equal(t, http.StatusOK, readOnlyGet.Code)
	readOnlyEnvelope := decodeJourneyEnvelope(t, readOnlyGet)
	require.NotNil(t, readOnlyEnvelope.Journey)
	assert.NotContains(t, readOnlyEnvelope.Journey.Links, "x-start")
	assert.NotContains(t, readOnlyEnvelope.Links, "x-capture")
}

type applicationComponentDeletedTestEvent struct {
	componentID string
}

func (e applicationComponentDeletedTestEvent) AggregateID() string { return e.componentID }
func (e applicationComponentDeletedTestEvent) EventType() string {
	return amPL.ApplicationComponentDeleted
}
func (e applicationComponentDeletedTestEvent) EventData() map[string]interface{} {
	return map[string]interface{}{"id": e.componentID, "name": "Phoenix", "deletedAt": time.Now().UTC()}
}
func (e applicationComponentDeletedTestEvent) OccurredAt() time.Time { return time.Now().UTC() }

func TestCapabilityJourneyIntegration_StaleReference_MarksStaleButJourneyRenders(t *testing.T) {
	tc, cleanup := setupCapabilityJourneyTestDB(t)
	defer cleanup()

	capID := uuid.New().String()
	tc.trackCapability(capID)
	toApp := uuid.New().String()
	require.Equal(t, http.StatusCreated, captureJourneyReq(tc.handlers, capID, CaptureJourneyRequest{
		Kind: valueobjects.JourneyKindMigration, FromComponentIDs: []string{uuid.New().String()}, ToComponentID: toApp,
	}, architectActor()).Code)

	ctx := sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
	require.NoError(t, tc.eventBus.Publish(ctx, []domain.DomainEvent{applicationComponentDeletedTestEvent{componentID: toApp}}))

	get := getJourneyForCapabilityReq(tc.handlers, capID, architectActor())
	require.Equal(t, http.StatusOK, get.Code)
	envelope := decodeJourneyEnvelope(t, get)
	require.NotNil(t, envelope.Journey)
	assert.True(t, envelope.Journey.ToApplication.Stale)
	assert.Equal(t, toApp, envelope.Journey.ToApplication.ComponentID, "the journey renders normally despite the stale reference")
}
