//go:build integration
// +build integration

package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/architecturemodeling/application/handlers"
	"easi/backend/internal/architecturemodeling/application/projectors"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	domainevents "easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type componentStack struct {
	component *ComponentHandlers
	ownership *ComponentOwnershipHandlers
	hosting   *ComponentHostingHandlers
	readModel *readmodels.ApplicationComponentReadModel
	userNames *readmodels.UserNameCacheReadModel
	teams     *readmodels.InternalTeamReadModel
	eventBus  events.EventBus
}

func setupComponentStack(db *sql.DB) *componentStack {
	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	links := NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	readModel := readmodels.NewApplicationComponentReadModel(tenantDB)
	userNames := readmodels.NewUserNameCacheReadModel(tenantDB)
	teams := readmodels.NewInternalTeamReadModel(tenantDB)

	componentProjector := projectors.NewApplicationComponentProjector(readModel)
	eventBus.Subscribe(archPL.ApplicationComponentCreated, componentProjector)
	ownershipProjector := projectors.NewApplicationOwnershipProjector(readModel)
	eventBus.Subscribe(archPL.ApplicationOwnerNominated, ownershipProjector)
	eventBus.Subscribe(archPL.ApplicationOwnershipConfirmed, ownershipProjector)
	eventBus.Subscribe(archPL.ApplicationOwnerAssigned, ownershipProjector)
	eventBus.Subscribe(archPL.ApplicationOwnershipCleared, ownershipProjector)
	eventBus.Subscribe(archPL.InternalTeamDeleted, projectors.NewTeamOwnershipDeletionReactor(readModel, commandBus))
	eventBus.Subscribe(archPL.ApplicationHostingClassified, projectors.NewApplicationHostingProjector(readModel))

	componentRepo := repositories.NewApplicationComponentRepository(eventStore)
	commandBus.Register("CreateApplicationComponent", handlers.NewCreateApplicationComponentHandler(componentRepo))
	commandBus.Register("NominateApplicationComponentOwner", handlers.NewNominateApplicationComponentOwnerHandler(componentRepo, userNames, teams))
	commandBus.Register("ConfirmApplicationComponentOwnership", handlers.NewConfirmApplicationComponentOwnershipHandler(componentRepo))
	commandBus.Register("AssignApplicationComponentOwner", handlers.NewAssignApplicationComponentOwnerHandler(componentRepo, userNames, teams))
	commandBus.Register("ClearApplicationComponentOwnership", handlers.NewClearApplicationComponentOwnershipHandler(componentRepo))
	commandBus.Register("ClassifyApplicationHosting", handlers.NewClassifyApplicationHostingHandler(componentRepo))

	return &componentStack{
		component: NewComponentHandlers(commandBus, readModel, links),
		ownership: NewComponentOwnershipHandlers(commandBus, readModel, links),
		hosting:   NewComponentHostingHandlers(commandBus, readModel, links),
		readModel: readModel,
		userNames: userNames,
		teams:     teams,
		eventBus:  eventBus,
	}
}

type ownershipCall struct {
	method string
	path   string
	body   any
}

func (s *componentStack) ownershipRequest(t *testing.T, ctx *testContext, componentID string, call ownershipCall) *httptest.ResponseRecorder {
	var payload []byte
	if call.body != nil {
		var err error
		payload, err = json.Marshal(call.body)
		require.NoError(t, err)
	}
	w, req := ctx.makeRequest(t, requestSpec{
		Method:    call.method,
		URL:       "/api/v1/components/" + componentID + call.path,
		Body:      payload,
		URLParams: map[string]string{"id": componentID},
	})
	req = withArchitectActor(req)
	switch {
	case call.method == http.MethodPost && call.path == "/ownership/nomination":
		s.ownership.NominateOwner(w, req)
	case call.method == http.MethodPost && call.path == "/ownership/confirmation":
		s.ownership.ConfirmOwnership(w, req)
	case call.method == http.MethodPut:
		s.ownership.AssignOwner(w, req)
	case call.method == http.MethodDelete:
		s.ownership.ClearOwnership(w, req)
	}
	return w
}

func decodeComponent(t *testing.T, body *bytes.Buffer) map[string]any {
	var out map[string]any
	require.NoError(t, json.Unmarshal(body.Bytes(), &out))
	return out
}

func TestOwnershipLifecycle_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	stack := setupComponentStack(testCtx.db)

	userID := fmt.Sprintf("own-user-%d", time.Now().UnixNano())
	require.NoError(t, stack.userNames.Upsert(tenantContext(), userID, "Alice Smith", "alice@example.com"))
	defer testCtx.db.Exec("DELETE FROM architecturemodeling.user_names WHERE user_id = $1", userID)

	componentID := testCtx.createComponentViaAPI(t, stack.component, "Ownership Service", "")

	w := stack.ownershipRequest(t, testCtx, componentID, ownershipCall{
		method: http.MethodPost, path: "/ownership/nomination",
		body: OwnerReferenceRequest{OwnerKind: "user", OwnerID: userID}})
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	nominated := decodeComponent(t, w.Body)
	assert.Equal(t, "nominated", nominated["ownershipState"])
	owner := nominated["owner"].(map[string]any)
	assert.Equal(t, "user", owner["kind"])
	assert.Equal(t, userID, owner["id"])
	assert.Equal(t, "Alice Smith", owner["name"])
	links := nominated["_links"].(map[string]any)
	assert.Contains(t, links, "x-confirm-owner")
	assert.Contains(t, links, "x-clear-owner")
	assert.NotContains(t, links, "x-nominate-owner")

	w = stack.ownershipRequest(t, testCtx, componentID, ownershipCall{method: http.MethodPost, path: "/ownership/confirmation"})
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	confirmed := decodeComponent(t, w.Body)
	assert.Equal(t, "owned", confirmed["ownershipState"])

	w = stack.ownershipRequest(t, testCtx, componentID, ownershipCall{method: http.MethodDelete, path: "/ownership"})
	require.Equal(t, http.StatusNoContent, w.Code)

	component, err := stack.readModel.GetByID(tenantContext(), componentID)
	require.NoError(t, err)
	assert.Equal(t, "unknown", component.OwnershipState)
	assert.Nil(t, component.Owner)
}

func TestOwnershipAssignTeamAndDeletionRevert_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	stack := setupComponentStack(testCtx.db)

	teamID := fmt.Sprintf("own-team-%d", time.Now().UnixNano())
	require.NoError(t, stack.teams.Insert(tenantContext(), readmodels.InternalTeamDTO{
		ID: teamID, Name: "Platform Ops", CreatedAt: time.Now().UTC(),
	}))
	defer testCtx.db.Exec("DELETE FROM architecturemodeling.internal_teams WHERE id = $1", teamID)

	componentID := testCtx.createComponentViaAPI(t, stack.component, "Managed Service", "")

	w := stack.ownershipRequest(t, testCtx, componentID, ownershipCall{
		method: http.MethodPut, path: "/ownership",
		body: OwnerReferenceRequest{OwnerKind: "team", OwnerID: teamID}})
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assigned := decodeComponent(t, w.Body)
	assert.Equal(t, "managed", assigned["ownershipState"])
	assert.Equal(t, "Platform Ops", assigned["owner"].(map[string]any)["name"])

	deleted := domainevents.NewInternalTeamDeleted(teamID, "Platform Ops")
	require.NoError(t, stack.eventBus.Publish(tenantContext(), []domain.DomainEvent{deleted}))

	component, err := stack.readModel.GetByID(tenantContext(), componentID)
	require.NoError(t, err)
	assert.Equal(t, "unknown", component.OwnershipState)
	assert.Nil(t, component.Owner)
}

func TestOwnershipNominateUnknownOwner_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	stack := setupComponentStack(testCtx.db)

	componentID := testCtx.createComponentViaAPI(t, stack.component, "Orphan Service", "")

	w := stack.ownershipRequest(t, testCtx, componentID, ownershipCall{
		method: http.MethodPost, path: "/ownership/nomination",
		body: OwnerReferenceRequest{OwnerKind: "user", OwnerID: "no-such-user"}})
	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}

func TestOwnershipStatistics_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	stack := setupComponentStack(testCtx.db)

	before, err := stack.readModel.Statistics(tenantContext())
	require.NoError(t, err)

	userID := fmt.Sprintf("stats-user-%d", time.Now().UnixNano())
	require.NoError(t, stack.userNames.Upsert(tenantContext(), userID, "Bob", "bob@example.com"))
	defer testCtx.db.Exec("DELETE FROM architecturemodeling.user_names WHERE user_id = $1", userID)

	firstID := testCtx.createComponentViaAPI(t, stack.component, "Stats Service A", "")
	testCtx.createComponentViaAPI(t, stack.component, "Stats Service B", "")

	w := stack.ownershipRequest(t, testCtx, firstID, ownershipCall{
		method: http.MethodPut, path: "/ownership",
		body: OwnerReferenceRequest{OwnerKind: "user", OwnerID: userID}})
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	after, err := stack.readModel.Statistics(tenantContext())
	require.NoError(t, err)
	assert.Equal(t, before.Total+2, after.Total)
	assert.Equal(t, before.Owned+1, after.Owned)
	assert.Equal(t, before.Unknown+1, after.Unknown)
}
