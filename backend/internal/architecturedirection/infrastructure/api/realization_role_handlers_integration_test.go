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
	"strings"
	"testing"
	"time"

	"easi/backend/internal/architecturedirection/application/handlers"
	"easi/backend/internal/architecturedirection/application/projectors"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/architecturedirection/infrastructure/repositories"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
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

type realizationRoleTestContext struct {
	db                   *sql.DB
	tenantDB             *database.TenantAwareDB
	handlers             *RealizationRoleHandlers
	timeHandlers         *TimeAssessmentHandlers
	readModel            *readmodels.RealizationRoleReadModel
	eventBus             events.EventBus
	directExists         map[string]string
	cleanupCapabilityIDs []string
}

func (tc *realizationRoleTestContext) trackCapability(capabilityID string) {
	tc.cleanupCapabilityIDs = append(tc.cleanupCapabilityIDs, capabilityID)
}

func (tc *realizationRoleTestContext) allowDirect(pair rolePair, realizationID string) {
	tc.directExists[pair.capID+"|"+pair.compID] = realizationID
}

func realizationRoleEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func setupRealizationRoleTestDB(t *testing.T) (*realizationRoleTestContext, func()) {
	t.Helper()
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		realizationRoleEnv("INTEGRATION_TEST_DB_HOST", "localhost"),
		realizationRoleEnv("INTEGRATION_TEST_DB_PORT", "5432"),
		realizationRoleEnv("INTEGRATION_TEST_DB_USER", "easi_app"),
		realizationRoleEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev"),
		realizationRoleEnv("INTEGRATION_TEST_DB_NAME", "easi"),
		realizationRoleEnv("INTEGRATION_TEST_DB_SSLMODE", "disable"),
	)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	readModel := readmodels.NewRealizationRoleReadModel(tenantDB)
	repo := repositories.NewRealizationRolesRepository(eventStore)

	timeReadModel := readmodels.NewTimeAssessmentReadModel(tenantDB)
	timeRepo := repositories.NewTimeAssessmentRepository(eventStore)

	projector := projectors.NewRealizationRoleProjector(readModel)
	referenceProjector := projectors.NewRealizationRoleReferenceProjector(readModel)
	reactor := projectors.NewRealizationRoleDeletionReactor(readModel, commandBus)
	eventBus.Subscribe(pl.RealizationRoleAssigned, projector)
	eventBus.Subscribe(pl.RealizationRoleCleared, projector)
	eventBus.Subscribe(cmPL.CapabilityDeleted, referenceProjector)
	eventBus.Subscribe(cmPL.SystemRealizationDeleted, reactor)

	timeProjector := projectors.NewTimeAssessmentProjector(timeReadModel)
	eventBus.Subscribe(pl.TimeAssessmentRecorded, timeProjector)
	eventBus.Subscribe(pl.TimeAssessmentRemoved, timeProjector)

	directExists := map[string]string{}
	directLookup := services.DirectRealizationLookup(func(_ context.Context, capID, compID string) (string, bool, error) {
		realizationID, exists := directExists[capID+"|"+compID]
		return realizationID, exists, nil
	})
	commandBus.Register("AssignRealizationRole", handlers.NewAssignRealizationRoleHandler(repo, readModel, directLookup))
	commandBus.Register("ClearRealizationRole", handlers.NewClearRealizationRoleHandler(repo, readModel))
	commandBus.Register("AssessRealization", handlers.NewAssessRealizationHandler(timeRepo, timeReadModel, directLookup))

	links := NewRealizationRoleLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	httpHandlers := NewRealizationRoleHandlers(commandBus, readModel, links)

	timeLinks := NewTimeAssessmentLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	timeHTTPHandlers := NewTimeAssessmentHandlers(commandBus, timeReadModel, timeLinks)

	ctx := &realizationRoleTestContext{
		db:           db,
		tenantDB:     tenantDB,
		handlers:     httpHandlers,
		timeHandlers: timeHTTPHandlers,
		readModel:    readModel,
		eventBus:     eventBus,
		directExists: directExists,
	}

	cleanup := func() {
		_, _ = db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", sharedvo.DefaultTenantID().Value()))
		for _, capID := range ctx.cleanupCapabilityIDs {
			var aggID string
			err := db.QueryRow(
				"SELECT aggregate_id FROM architecturedirection.realization_role_aggregates WHERE tenant_id = $1 AND capability_id = $2",
				sharedvo.DefaultTenantID().Value(), capID,
			).Scan(&aggID)
			if err == nil {
				_, _ = db.Exec("DELETE FROM infrastructure.events WHERE aggregate_id = $1", aggID)
			}
			_, _ = db.Exec("DELETE FROM architecturedirection.realization_roles WHERE tenant_id = $1 AND capability_id = $2",
				sharedvo.DefaultTenantID().Value(), capID)
			_, _ = db.Exec("DELETE FROM architecturedirection.realization_role_aggregates WHERE tenant_id = $1 AND capability_id = $2",
				sharedvo.DefaultTenantID().Value(), capID)
			_, _ = db.Exec("DELETE FROM architecturedirection.time_assessments WHERE tenant_id = $1 AND capability_id = $2",
				sharedvo.DefaultTenantID().Value(), capID)
		}
		db.Close()
	}
	return ctx, cleanup
}

func runRealizationRoleRequest(req *http.Request, pattern string, handler http.HandlerFunc, actor sharedcontext.Actor) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	r.Method(req.Method, pattern, handler)
	ctx := sharedcontext.WithTenant(req.Context(), sharedvo.DefaultTenantID())
	ctx = sharedcontext.WithActor(ctx, actor)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

const realizationRoleItemPattern = "/api/v1/capabilities/{id}/components/{componentId}/realization-role"
const realizationRolesCollectionPattern = "/api/v1/realization-roles"

type rolePair struct {
	capID  string
	compID string
}

type putRealizationRoleRequest struct {
	handlers *RealizationRoleHandlers
	capID    string
	compID   string
	role     string
	actor    sharedcontext.Actor
}

func (p putRealizationRoleRequest) execute() *httptest.ResponseRecorder {
	body, _ := json.Marshal(AssignRealizationRoleRequest{Role: p.role})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/capabilities/"+p.capID+"/components/"+p.compID+"/realization-role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	actor := p.actor
	if actor.Email == "" {
		actor = architectActor()
	}
	return runRealizationRoleRequest(req, realizationRoleItemPattern, p.handlers.PutRealizationRole, actor)
}

func getRealizationRoleReq(h *RealizationRoleHandlers, pair rolePair, actor sharedcontext.Actor) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/"+pair.capID+"/components/"+pair.compID+"/realization-role", nil)
	return runRealizationRoleRequest(req, realizationRoleItemPattern, h.GetRealizationRole, actor)
}

func deleteRealizationRoleReq(h *RealizationRoleHandlers, pair rolePair) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/capabilities/"+pair.capID+"/components/"+pair.compID+"/realization-role", nil)
	return runRealizationRoleRequest(req, realizationRoleItemPattern, h.DeleteRealizationRole, architectActor())
}

func listRealizationRolesReq(h *RealizationRoleHandlers, capabilityIDs []string, actor sharedcontext.Actor) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, realizationRolesCollectionPattern+"?capabilityIds="+strings.Join(capabilityIDs, ","), nil)
	return runRealizationRoleRequest(req, "/api/v1/realization-roles", h.GetRealizationRoles, actor)
}

func TestRealizationRoleIntegration_FirstAssign_Returns201(t *testing.T) {
	tc, cleanup := setupRealizationRoleTestDB(t)
	defer cleanup()

	capID, compID := uuid.New().String(), uuid.New().String()
	tc.trackCapability(capID)
	tc.allowDirect(rolePair{capID, compID}, uuid.New().String())

	rec := putRealizationRoleRequest{handlers: tc.handlers, capID: capID, compID: compID, role: valueobjects.RealizationRoleStandard}.execute()

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.NotEmpty(t, rec.Header().Get("Location"))
	var dto readmodels.RealizationRoleDTO
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&dto))
	assert.Equal(t, valueobjects.RealizationRoleStandard, dto.Role)
}

func TestRealizationRoleIntegration_ReAssign_Returns200(t *testing.T) {
	tc, cleanup := setupRealizationRoleTestDB(t)
	defer cleanup()

	capID, compID := uuid.New().String(), uuid.New().String()
	tc.trackCapability(capID)
	tc.allowDirect(rolePair{capID, compID}, uuid.New().String())

	first := putRealizationRoleRequest{handlers: tc.handlers, capID: capID, compID: compID, role: valueobjects.RealizationRoleStandard}.execute()
	require.Equal(t, http.StatusCreated, first.Code, first.Body.String())

	second := putRealizationRoleRequest{handlers: tc.handlers, capID: capID, compID: compID, role: valueobjects.RealizationRoleLegacy}.execute()
	require.Equal(t, http.StatusOK, second.Code, second.Body.String())

	get := getRealizationRoleReq(tc.handlers, rolePair{capID, compID}, architectActor())
	require.Equal(t, http.StatusOK, get.Code)
	var dto readmodels.RealizationRoleDTO
	require.NoError(t, json.NewDecoder(get.Body).Decode(&dto))
	assert.Equal(t, valueobjects.RealizationRoleLegacy, dto.Role)
}

func TestRealizationRoleIntegration_AssignStandard_DisplacesPreviousHolder(t *testing.T) {
	tc, cleanup := setupRealizationRoleTestDB(t)
	defer cleanup()

	capID := uuid.New().String()
	seabook, phoenix := uuid.New().String(), uuid.New().String()
	tc.trackCapability(capID)
	tc.allowDirect(rolePair{capID, seabook}, uuid.New().String())
	tc.allowDirect(rolePair{capID, phoenix}, uuid.New().String())

	require.Equal(t, http.StatusCreated,
		putRealizationRoleRequest{handlers: tc.handlers, capID: capID, compID: seabook, role: valueobjects.RealizationRoleStandard}.execute().Code)

	require.Equal(t, http.StatusCreated,
		putRealizationRoleRequest{handlers: tc.handlers, capID: capID, compID: phoenix, role: valueobjects.RealizationRoleStandard}.execute().Code)

	seabookGet := getRealizationRoleReq(tc.handlers, rolePair{capID, seabook}, architectActor())
	assert.Equal(t, http.StatusNotFound, seabookGet.Code, "the previous standard holder becomes unclassified")

	phoenixGet := getRealizationRoleReq(tc.handlers, rolePair{capID, phoenix}, architectActor())
	require.Equal(t, http.StatusOK, phoenixGet.Code)
	var dto readmodels.RealizationRoleDTO
	require.NoError(t, json.NewDecoder(phoenixGet.Body).Decode(&dto))
	assert.Equal(t, valueobjects.RealizationRoleStandard, dto.Role)
}

func TestRealizationRoleIntegration_Clear_Returns204(t *testing.T) {
	tc, cleanup := setupRealizationRoleTestDB(t)
	defer cleanup()

	capID, compID := uuid.New().String(), uuid.New().String()
	tc.trackCapability(capID)
	tc.allowDirect(rolePair{capID, compID}, uuid.New().String())
	require.Equal(t, http.StatusCreated,
		putRealizationRoleRequest{handlers: tc.handlers, capID: capID, compID: compID, role: valueobjects.RealizationRoleLegacy}.execute().Code)

	del := deleteRealizationRoleReq(tc.handlers, rolePair{capID, compID})
	require.Equal(t, http.StatusNoContent, del.Code)

	get := getRealizationRoleReq(tc.handlers, rolePair{capID, compID}, architectActor())
	assert.Equal(t, http.StatusNotFound, get.Code)
}

func TestRealizationRoleIntegration_NoDirectRealization_Fails(t *testing.T) {
	tc, cleanup := setupRealizationRoleTestDB(t)
	defer cleanup()

	capID, compID := uuid.New().String(), uuid.New().String()

	rec := putRealizationRoleRequest{handlers: tc.handlers, capID: capID, compID: compID, role: valueobjects.RealizationRoleStandard}.execute()
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRealizationRoleIntegration_RoleAndTimeAssessment_CoexistOnSamePair(t *testing.T) {
	tc, cleanup := setupRealizationRoleTestDB(t)
	defer cleanup()

	capID, compID := uuid.New().String(), uuid.New().String()
	tc.trackCapability(capID)
	tc.allowDirect(rolePair{capID, compID}, uuid.New().String())

	roleRec := putRealizationRoleRequest{handlers: tc.handlers, capID: capID, compID: compID, role: valueobjects.RealizationRoleStandard}.execute()
	require.Equal(t, http.StatusCreated, roleRec.Code, roleRec.Body.String())

	body, _ := json.Marshal(AssessRealizationRequest{Grade: "Tolerate", Rationale: "still viable"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/capabilities/"+capID+"/components/"+compID+"/time-assessment", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	timeRec := runRealizationRoleRequest(req, timeAssessmentItemPattern, tc.timeHandlers.PutTimeAssessment, architectActor())
	require.Equal(t, http.StatusCreated, timeRec.Code, timeRec.Body.String(), "TIME assessment write must not be rejected because a role already exists on the pair")

	roleGet := getRealizationRoleReq(tc.handlers, rolePair{capID, compID}, architectActor())
	require.Equal(t, http.StatusOK, roleGet.Code)
	var roleDTO readmodels.RealizationRoleDTO
	require.NoError(t, json.NewDecoder(roleGet.Body).Decode(&roleDTO))
	assert.Equal(t, valueobjects.RealizationRoleStandard, roleDTO.Role)

	timeGet := getTimeAssessmentReq(t, tc.timeHandlers, timeAssessmentPairID{CapabilityID: capID, ComponentID: compID})
	require.Equal(t, http.StatusOK, timeGet.Code)
	var timeDTO readmodels.TimeAssessmentDTO
	require.NoError(t, json.NewDecoder(timeGet.Body).Decode(&timeDTO))
	require.NotNil(t, timeDTO.Grade)
	assert.Equal(t, "Tolerate", *timeDTO.Grade)
}

func TestRealizationRoleIntegration_ReadOnlyActor_NoWriteAffordances(t *testing.T) {
	tc, cleanup := setupRealizationRoleTestDB(t)
	defer cleanup()

	capID, compID := uuid.New().String(), uuid.New().String()
	tc.trackCapability(capID)
	tc.allowDirect(rolePair{capID, compID}, uuid.New().String())
	require.Equal(t, http.StatusCreated,
		putRealizationRoleRequest{handlers: tc.handlers, capID: capID, compID: compID, role: valueobjects.RealizationRoleLegacy}.execute().Code)

	architectGet := getRealizationRoleReq(tc.handlers, rolePair{capID, compID}, architectActor())
	require.Equal(t, http.StatusOK, architectGet.Code)
	var architectDTO readmodels.RealizationRoleDTO
	require.NoError(t, json.NewDecoder(architectGet.Body).Decode(&architectDTO))
	assert.Contains(t, architectDTO.Links, "edit")
	assert.Contains(t, architectDTO.Links, "delete")

	readOnlyGet := getRealizationRoleReq(tc.handlers, rolePair{capID, compID}, stakeholderActor())
	require.Equal(t, http.StatusOK, readOnlyGet.Code)
	var readOnlyDTO readmodels.RealizationRoleDTO
	require.NoError(t, json.NewDecoder(readOnlyGet.Body).Decode(&readOnlyDTO))
	assert.Equal(t, valueobjects.RealizationRoleLegacy, readOnlyDTO.Role, "read-only users still see the role")
	assert.NotContains(t, readOnlyDTO.Links, "edit")
	assert.NotContains(t, readOnlyDTO.Links, "delete")
}

func TestRealizationRoleIntegration_BulkGet_ReturnsRolesAcrossCapabilities(t *testing.T) {
	tc, cleanup := setupRealizationRoleTestDB(t)
	defer cleanup()

	capA, capB := uuid.New().String(), uuid.New().String()
	compA, compB := uuid.New().String(), uuid.New().String()
	tc.trackCapability(capA)
	tc.trackCapability(capB)
	tc.allowDirect(rolePair{capA, compA}, uuid.New().String())
	tc.allowDirect(rolePair{capB, compB}, uuid.New().String())

	require.Equal(t, http.StatusCreated,
		putRealizationRoleRequest{handlers: tc.handlers, capID: capA, compID: compA, role: valueobjects.RealizationRoleStandard}.execute().Code)
	require.Equal(t, http.StatusCreated,
		putRealizationRoleRequest{handlers: tc.handlers, capID: capB, compID: compB, role: valueobjects.RealizationRoleLegacy}.execute().Code)

	rec := listRealizationRolesReq(tc.handlers, []string{capA, capB}, architectActor())
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var collection struct {
		Data  []readmodels.RealizationRoleDTO `json:"data"`
		Links sharedAPI.Links                 `json:"_links"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&collection))
	require.Len(t, collection.Data, 2)
	assert.Contains(t, collection.Links, "x-assign")
}

type systemRealizationDeletedTestEvent struct {
	realizationID string
}

func (e systemRealizationDeletedTestEvent) AggregateID() string { return e.realizationID }
func (e systemRealizationDeletedTestEvent) EventType() string   { return cmPL.SystemRealizationDeleted }
func (e systemRealizationDeletedTestEvent) EventData() map[string]interface{} {
	return map[string]interface{}{"id": e.realizationID, "deletedAt": time.Now().UTC()}
}
func (e systemRealizationDeletedTestEvent) OccurredAt() time.Time { return time.Now().UTC() }

func TestRealizationRoleIntegration_SystemRealizationDeleted_ClearsRoleViaReactor(t *testing.T) {
	tc, cleanup := setupRealizationRoleTestDB(t)
	defer cleanup()

	capID, compID := uuid.New().String(), uuid.New().String()
	realizationID := uuid.New().String()
	tc.trackCapability(capID)
	tc.allowDirect(rolePair{capID, compID}, realizationID)
	require.Equal(t, http.StatusCreated,
		putRealizationRoleRequest{handlers: tc.handlers, capID: capID, compID: compID, role: valueobjects.RealizationRoleStandard}.execute().Code)

	ctx := sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
	require.NoError(t, tc.eventBus.Publish(ctx, []domain.DomainEvent{systemRealizationDeletedTestEvent{realizationID: realizationID}}))

	get := getRealizationRoleReq(tc.handlers, rolePair{capID, compID}, architectActor())
	assert.Equal(t, http.StatusNotFound, get.Code, "R6: deleting the realisation clears the role via a recorded reaction")
}
