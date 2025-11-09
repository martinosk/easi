// +build integration

package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/architecturemodeling/application/handlers"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Connect to test database
	connStr := "host=localhost port=5432 user=easi password=easi dbname=easi sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	// Initialize event store schema
	eventStore := eventstore.NewPostgresEventStore(db)
	err = eventStore.InitializeSchema()
	require.NoError(t, err)

	// Initialize read model schema
	readModel := readmodels.NewApplicationComponentReadModel(db)
	err = readModel.InitializeSchema()
	require.NoError(t, err)

	cleanup := func() {
		// Clean up test data
		db.Exec("TRUNCATE TABLE application_components CASCADE")
		db.Exec("TRUNCATE TABLE events CASCADE")
		db.Close()
	}

	return db, cleanup
}

func setupHandlers(db *sql.DB) (*ComponentHandlers, *readmodels.ApplicationComponentReadModel) {
	eventStore := eventstore.NewPostgresEventStore(db)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")

	// Setup repository and handlers
	componentRepo := repositories.NewApplicationComponentRepository(eventStore)
	createHandler := handlers.NewCreateApplicationComponentHandler(componentRepo)
	commandBus.Register("CreateApplicationComponent", createHandler)

	// Setup read model
	readModel := readmodels.NewApplicationComponentReadModel(db)

	// Setup HTTP handlers
	componentHandlers := NewComponentHandlers(commandBus, readModel, hateoas)

	return componentHandlers, readModel
}

func TestCreateComponent_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, readModel := setupHandlers(db)

	// Create component via API
	reqBody := CreateApplicationComponentRequest{
		Name:        "User Service",
		Description: "Handles user authentication and authorization",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/components", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateApplicationComponent(w, req)

	// Assert HTTP response
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify event was saved to event store
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM events WHERE event_type = 'ApplicationComponentCreated'").Scan(&count)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1, "At least one event should be created")

	// Verify event data
	var eventData string
	err = db.QueryRow("SELECT event_data FROM events WHERE event_type = 'ApplicationComponentCreated'").Scan(&eventData)
	require.NoError(t, err)
	assert.Contains(t, eventData, "User Service")
	assert.Contains(t, eventData, "Handles user authentication and authorization")

	// Note: Read model would be populated by event projector in production
	// For now, we manually project the event to test the full flow
	var aggregateID string
	err = db.QueryRow("SELECT aggregate_id FROM events WHERE event_type = 'ApplicationComponentCreated'").Scan(&aggregateID)
	require.NoError(t, err)

	// Manually insert into read model for testing
	_, err = db.Exec(
		"INSERT INTO application_components (id, name, description, created_at) VALUES ($1, $2, $3, NOW())",
		aggregateID, "User Service", "Handles user authentication and authorization",
	)
	require.NoError(t, err)

	// Now test GET endpoint
	components, err := readModel.GetAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, components, 1)
	assert.Equal(t, "User Service", components[0].Name)
	assert.Equal(t, "Handles user authentication and authorization", components[0].Description)
}

func TestGetAllComponents_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, _ := setupHandlers(db)

	// Create test data directly in read model
	db.Exec("INSERT INTO application_components (id, name, description, created_at) VALUES ($1, $2, $3, NOW())",
		"test-id-1", "Service A", "Description A")
	db.Exec("INSERT INTO application_components (id, name, description, created_at) VALUES ($1, $2, $3, NOW())",
		"test-id-2", "Service B", "Description B")

	// Test GET all
	req := httptest.NewRequest(http.MethodGet, "/api/v1/components", nil)
	w := httptest.NewRecorder()

	handlers.GetAllComponents(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []readmodels.ApplicationComponentDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response.Data, 2)

	// Verify HATEOAS links are present
	for _, comp := range response.Data {
		assert.NotNil(t, comp.Links)
		assert.Contains(t, comp.Links, "self")
		assert.Contains(t, comp.Links, "archimate")
	}
}

func TestGetComponentByID_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, _ := setupHandlers(db)

	// Create test data
	componentID := "test-component-123"
	db.Exec("INSERT INTO application_components (id, name, description, created_at) VALUES ($1, $2, $3, NOW())",
		componentID, "Test Service", "Test Description")

	// Test GET by ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/components/"+componentID, nil)
	w := httptest.NewRecorder()

	// Add URL param using chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", componentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetComponentByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data readmodels.ApplicationComponentDTO `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, componentID, response.Data.ID)
	assert.Equal(t, "Test Service", response.Data.Name)
	assert.Equal(t, "Test Description", response.Data.Description)
	assert.NotNil(t, response.Data.Links)
}

func TestGetComponentByID_NotFound_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, _ := setupHandlers(db)

	// Test GET non-existent component
	req := httptest.NewRequest(http.MethodGet, "/api/v1/components/non-existent", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "non-existent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetComponentByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateComponent_ValidationError_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, _ := setupHandlers(db)

	// Create component with empty name (should fail validation)
	reqBody := CreateApplicationComponentRequest{
		Name:        "",
		Description: "Some description",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/components", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateApplicationComponent(w, req)

	// Assert validation error
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify no event was created
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
