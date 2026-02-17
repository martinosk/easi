//go:build integration
// +build integration

package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/architectureviews/application/handlers"
	"easi/backend/internal/architectureviews/application/projectors"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type viewTestHarness struct {
	viewHandlers      *ViewHandlers
	componentHandlers *ViewComponentHandlers
	colorHandlers     *ViewColorHandlers
	elementHandlers   *ViewElementHandlers
}

func setupViewHandlers(db *sql.DB) *viewTestHarness {
	tenantDB := database.NewTenantAwareDB(db)

	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	links := NewViewLinks(sharedAPI.NewHATEOASLinks("/api/v1"))

	readModel := readmodels.NewArchitectureViewReadModel(tenantDB)
	projector := projectors.NewArchitectureViewProjector(readModel)

	eventBus := events.NewInMemoryEventBus()
	eventBus.SubscribeAll(projector)
	eventStore.SetEventBus(eventBus)

	viewRepo := repositories.NewArchitectureViewRepository(eventStore)
	layoutRepo := repositories.NewViewLayoutRepository(tenantDB)

	commandBus.Register("CreateView", handlers.NewCreateViewHandler(viewRepo, readModel))
	commandBus.Register("AddComponentToView", handlers.NewAddComponentToViewHandler(viewRepo, layoutRepo))
	commandBus.Register("UpdateComponentPosition", handlers.NewUpdateComponentPositionHandler(layoutRepo))
	commandBus.Register("RenameView", handlers.NewRenameViewHandler(viewRepo))
	commandBus.Register("DeleteView", handlers.NewDeleteViewHandler(viewRepo))
	commandBus.Register("RemoveComponentFromView", handlers.NewRemoveComponentFromViewHandler(viewRepo))
	commandBus.Register("SetDefaultView", handlers.NewSetDefaultViewHandler(viewRepo, readModel))
	commandBus.Register("UpdateViewEdgeType", handlers.NewUpdateViewEdgeTypeHandler(layoutRepo))
	commandBus.Register("UpdateViewLayoutDirection", handlers.NewUpdateViewLayoutDirectionHandler(layoutRepo))
	commandBus.Register("UpdateViewColorScheme", handlers.NewUpdateViewColorSchemeHandler(layoutRepo))
	commandBus.Register("UpdateElementColor", handlers.NewUpdateElementColorHandler(layoutRepo))
	commandBus.Register("ClearElementColor", handlers.NewClearElementColorHandler(layoutRepo))

	return &viewTestHarness{
		viewHandlers:      NewViewHandlers(commandBus, readModel, links),
		componentHandlers: NewViewComponentHandlers(commandBus, readModel),
		colorHandlers:     NewViewColorHandlers(commandBus, readModel, links),
		elementHandlers:   NewViewElementHandlers(layoutRepo, readModel),
	}
}

type viewTestContext struct {
	db         *sql.DB
	testID     string
	createdIDs []string
}

func setupViewTestDB(t *testing.T) (*viewTestContext, func()) {
	dbHost := getEnv("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getEnv("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getEnv("INTEGRATION_TEST_DB_USER", "easi_app")
	dbPassword := getEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev")
	dbName := getEnv("INTEGRATION_TEST_DB_NAME", "easi")
	dbSSLMode := getEnv("INTEGRATION_TEST_DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	testID := fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())

	ctx := &viewTestContext{
		db:         db,
		testID:     testID,
		createdIDs: make([]string, 0),
	}

	cleanup := func() {
		db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
		for _, id := range ctx.createdIDs {
			db.Exec("DELETE FROM architectureviews.view_element_positions WHERE view_id = $1", id)
			db.Exec("DELETE FROM architectureviews.view_preferences WHERE view_id = $1", id)
			db.Exec("DELETE FROM architectureviews.architecture_views WHERE id = $1", id)
			db.Exec("DELETE FROM infrastructure.events WHERE aggregate_id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

func (ctx *viewTestContext) trackID(id string) {
	ctx.createdIDs = append(ctx.createdIDs, id)
}

type position struct {
	x float64
	y float64
}

func (ctx *viewTestContext) makeRequest(t *testing.T, method, url string, body []byte, urlParams map[string]string) (*httptest.ResponseRecorder, *http.Request) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, url, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req = withTestTenantAndWritePermission(req, "views")

	if len(urlParams) > 0 {
		rctx := chi.NewRouteContext()
		for key, value := range urlParams {
			rctx.URLParams.Add(key, value)
		}
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}

	return httptest.NewRecorder(), req
}

func (ctx *viewTestContext) createViewViaAPI(t *testing.T, h *viewTestHarness, name, description string) string {
	body, _ := json.Marshal(CreateViewRequest{Name: name, Description: description})
	w, req := ctx.makeRequest(t, http.MethodPost, "/api/v1/views", body, nil)
	h.viewHandlers.CreateView(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var response struct {
		ID string `json:"id"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	require.NotEmpty(t, response.ID)
	ctx.trackID(response.ID)
	return response.ID
}

func (ctx *viewTestContext) addComponentViaAPI(t *testing.T, h *viewTestHarness, viewID, componentID string, pos position) {
	body, _ := json.Marshal(AddComponentRequest{ComponentID: componentID, X: pos.x, Y: pos.y})
	w, req := ctx.makeRequest(t, http.MethodPost, "/api/v1/views/"+viewID+"/components", body, map[string]string{"id": viewID})
	h.componentHandlers.AddComponentToView(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func (ctx *viewTestContext) addCapabilityViaAPI(t *testing.T, h *viewTestHarness, viewID, capabilityID string, pos position) {
	body, _ := json.Marshal(AddCapabilityRequest{CapabilityID: capabilityID, X: pos.x, Y: pos.y})
	w, req := ctx.makeRequest(t, http.MethodPost, "/api/v1/views/"+viewID+"/capabilities", body, map[string]string{"id": viewID})
	h.elementHandlers.AddCapabilityToView(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func (ctx *viewTestContext) getViewViaAPI(t *testing.T, h *viewTestHarness, viewID string) readmodels.ArchitectureViewDTO {
	w, req := ctx.makeRequest(t, http.MethodGet, "/api/v1/views/"+viewID, nil, map[string]string{"id": viewID})
	h.viewHandlers.GetViewByID(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response readmodels.ArchitectureViewDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	return response
}

func (ctx *viewTestContext) setElementColorViaAPI(t *testing.T, h *viewTestHarness, viewID, elementID, elementType, color string) {
	body, _ := json.Marshal(UpdateElementColorRequest{Color: color})

	var urlPath, urlParam string
	var handler func(http.ResponseWriter, *http.Request)
	if elementType == "component" {
		urlPath, urlParam = "components", "componentId"
		handler = h.colorHandlers.UpdateComponentColor
	} else {
		urlPath, urlParam = "capabilities", "capabilityId"
		handler = h.colorHandlers.UpdateCapabilityColor
	}

	w, req := ctx.makeRequest(t, http.MethodPatch, "/api/v1/views/"+viewID+"/"+urlPath+"/"+elementID+"/color", body, map[string]string{
		"id":     viewID,
		urlParam: elementID,
	})
	handler(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)
}
