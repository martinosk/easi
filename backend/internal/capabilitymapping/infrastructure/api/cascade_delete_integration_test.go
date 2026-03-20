//go:build integration
// +build integration

package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/projectors"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/infrastructure/adapters"
	"easi/backend/internal/capabilitymapping/infrastructure/architecturemodeling"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCascadeTestDB(t *testing.T) (*testContext, func()) {
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

	ctx := &testContext{
		db:         db,
		testID:     testID,
		createdIDs: make([]string, 0),
	}

	cleanup := func() {
		for _, id := range ctx.createdIDs {
			db.Exec("DELETE FROM capabilitymapping.capability_realizations WHERE id = $1", id)
			db.Exec("DELETE FROM capabilitymapping.capability_dependencies WHERE id = $1", id)
			db.Exec("DELETE FROM capabilitymapping.capabilities WHERE id = $1", id)
			db.Exec("DELETE FROM infrastructure.events WHERE aggregate_id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

func setupCascadeHandlers(db *sql.DB) *CapabilityHandlers {
	tenantDB := database.NewTenantAwareDB(db)

	es := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")
	links := NewCapabilityMappingLinks(hateoas)

	eventBus := events.NewInMemoryEventBus()
	es.SetEventBus(eventBus)

	capabilityRM := readmodels.NewCapabilityReadModel(tenantDB)
	realizationRM := readmodels.NewRealizationReadModel(tenantDB)
	dependencyRM := readmodels.NewDependencyReadModel(tenantDB)
	assignmentRM := readmodels.NewDomainCapabilityAssignmentReadModel(tenantDB)

	capabilityProjector := projectors.NewCapabilityProjector(capabilityRM, assignmentRM)
	for _, event := range []string{
		"CapabilityCreated", "CapabilityUpdated", "CapabilityMetadataUpdated",
		"CapabilityExpertAdded", "CapabilityTagAdded", "CapabilityDeleted",
	} {
		eventBus.Subscribe(event, capabilityProjector)
	}

	dependencyProjector := projectors.NewDependencyProjector(dependencyRM)
	eventBus.Subscribe(cmPL.CapabilityDependencyCreated, dependencyProjector)
	eventBus.Subscribe(cmPL.CapabilityDependencyDeleted, dependencyProjector)

	realizationProjector := projectors.NewRealizationProjector(realizationRM, &noOpComponentGateway{})
	eventBus.Subscribe("SystemLinkedToCapability", realizationProjector)
	eventBus.Subscribe("SystemRealizationDeleted", realizationProjector)

	capabilityRepo := repositories.NewCapabilityRepository(es)
	realizationRepo := repositories.NewRealizationRepository(es)
	dependencyRepo := repositories.NewDependencyRepository(es)

	lookupAdapter := adapters.NewCapabilityLookupAdapter(capabilityRM)
	hierarchyService := services.NewCapabilityHierarchyService(lookupAdapter)
	childrenChecker := adapters.NewCapabilityChildrenCheckerAdapter(capabilityRM)
	deletionService := services.NewCapabilityDeletionService(childrenChecker)

	commandBus.Register("CreateCapability", handlers.NewCreateCapabilityHandler(capabilityRepo))
	commandBus.Register("UpdateCapability", handlers.NewUpdateCapabilityHandler(capabilityRepo))
	commandBus.Register("DeleteCapability", handlers.NewDeleteCapabilityHandler(capabilityRepo, deletionService, realizationRM, capabilityRM))
	commandBus.Register("DeleteSystemRealization", handlers.NewDeleteSystemRealizationHandler(realizationRepo))
	commandBus.Register("DeleteCapabilityDependency", handlers.NewDeleteCapabilityDependencyHandler(dependencyRepo))
	commandBus.Register("CascadeDeleteCapability", handlers.NewCascadeDeleteCapabilityHandler(handlers.CascadeDeleteDeps{
		Repository:       capabilityRepo,
		HierarchyService: hierarchyService,
		RealizationRM:    realizationRM,
		DependencyRM:     dependencyRM,
		CommandBus:       commandBus,
		CapabilityLookup: capabilityRM,
		ComponentDeleter: adapters.NewNoOpComponentDeleter(),
	}))

	impactQuery := handlers.NewDeleteImpactQuery(hierarchyService, realizationRM)

	return NewCapabilityHandlers(commandBus, capabilityRM, links, impactQuery)
}

func (ctx *testContext) createTestCapabilityWithParent(t *testing.T, id, name, level, parentID string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		"INSERT INTO capabilitymapping.capabilities (id, name, description, level, parent_id, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())",
		id, name, "", level, parentID, testTenantID(), "Genesis", "Active",
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

func (ctx *testContext) insertCapabilityCreatedEvent(t *testing.T, id, name, level, parentID string) {
	ctx.setTenantContext(t)
	eventData := fmt.Sprintf(`{"id":"%s","name":"%s","description":"","level":"%s","parentId":"%s"}`, id, name, level, parentID)
	_, err := ctx.db.Exec(
		`INSERT INTO infrastructure.events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
		 VALUES ($1, $2, 'CapabilityCreated', $3, 1, NOW(), 'test-user', 'test@example.com')`,
		testTenantID(), id, eventData,
	)
	require.NoError(t, err)
}

func (ctx *testContext) createTestRealization(t *testing.T, id, componentID, capabilityID string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		`INSERT INTO capabilitymapping.capability_realizations (id, component_id, capability_id, component_name, realization_level, origin, notes, tenant_id, linked_at)
		 VALUES ($1, $2, $3, 'Test Component', 'Full', 'Direct', '', $4, NOW())`,
		id, componentID, capabilityID, testTenantID(),
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

func (ctx *testContext) insertRealizationCreatedEvent(t *testing.T, id, componentID, capabilityID string) {
	ctx.setTenantContext(t)
	eventData := fmt.Sprintf(`{"realizationId":"%s","componentId":"%s","capabilityId":"%s","componentName":"Test Component","realizationLevel":"Full"}`, id, componentID, capabilityID)
	_, err := ctx.db.Exec(
		`INSERT INTO infrastructure.events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
		 VALUES ($1, $2, 'SystemLinkedToCapability', $3, 1, NOW(), 'test-user', 'test@example.com')`,
		testTenantID(), id, eventData,
	)
	require.NoError(t, err)
}

func (ctx *testContext) createTestDependency(t *testing.T, id, sourceID, targetID string) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		`INSERT INTO capabilitymapping.capability_dependencies (id, source_capability_id, target_capability_id, dependency_type, tenant_id, created_at)
		 VALUES ($1, $2, $3, 'Requires', $4, NOW())`,
		id, sourceID, targetID, testTenantID(),
	)
	require.NoError(t, err)
	ctx.trackID(id)
}

func (ctx *testContext) insertDependencyCreatedEvent(t *testing.T, id, sourceID, targetID string) {
	ctx.setTenantContext(t)
	eventData := fmt.Sprintf(`{"dependencyId":"%s","sourceCapabilityId":"%s","targetCapabilityId":"%s","dependencyType":"Requires"}`, id, sourceID, targetID)
	_, err := ctx.db.Exec(
		`INSERT INTO infrastructure.events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
		 VALUES ($1, $2, 'CapabilityDependencyCreated', $3, 1, NOW(), 'test-user', 'test@example.com')`,
		testTenantID(), id, eventData,
	)
	require.NoError(t, err)
}

type noOpComponentGateway struct{}

func (g *noOpComponentGateway) GetByID(_ context.Context, _ string) (*architecturemodeling.ComponentDTO, error) {
	return nil, nil
}

func TestGetDeleteImpact_LeafCapability_Integration(t *testing.T) {
	testCtx, cleanup := setupCascadeTestDB(t)
	defer cleanup()

	h := setupCascadeHandlers(testCtx.db)

	capID := uuid.New().String()
	testCtx.createTestCapability(t, capID, "Leaf Capability", "L1")
	testCtx.insertCapabilityCreatedEvent(t, capID, "Leaf Capability", "L1", "")

	w, req := makeRequest(t, http.MethodGet, "/api/v1/capabilities/"+capID+"/delete-impact", nil, map[string]string{"id": capID})
	h.GetDeleteImpact(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DeleteImpactResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, capID, response.CapabilityID)
	assert.False(t, response.HasDescendants)
	assert.Empty(t, response.AffectedCapabilities)
	assert.Empty(t, response.RealizationsOnDeletedCapabilities)
	assert.Empty(t, response.RealizationsOnRetainedCapabilities)
}

func TestGetDeleteImpact_WithDescendants_Integration(t *testing.T) {
	testCtx, cleanup := setupCascadeTestDB(t)
	defer cleanup()

	h := setupCascadeHandlers(testCtx.db)

	l1ID := uuid.New().String()
	l2ID := uuid.New().String()
	l3ID := uuid.New().String()

	testCtx.createTestCapability(t, l1ID, "L1 Capability", "L1")
	testCtx.insertCapabilityCreatedEvent(t, l1ID, "L1 Capability", "L1", "")

	testCtx.createTestCapabilityWithParent(t, l2ID, "L2 Capability", "L2", l1ID)
	testCtx.insertCapabilityCreatedEvent(t, l2ID, "L2 Capability", "L2", l1ID)

	testCtx.createTestCapabilityWithParent(t, l3ID, "L3 Capability", "L3", l2ID)
	testCtx.insertCapabilityCreatedEvent(t, l3ID, "L3 Capability", "L3", l2ID)

	w, req := makeRequest(t, http.MethodGet, "/api/v1/capabilities/"+l1ID+"/delete-impact", nil, map[string]string{"id": l1ID})
	h.GetDeleteImpact(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DeleteImpactResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, l1ID, response.CapabilityID)
	assert.True(t, response.HasDescendants)
	assert.Len(t, response.AffectedCapabilities, 2)

	affectedIDs := make(map[string]bool)
	for _, cap := range response.AffectedCapabilities {
		affectedIDs[cap.ID] = true
	}
	assert.True(t, affectedIDs[l2ID])
	assert.True(t, affectedIDs[l3ID])
}

func TestGetDeleteImpact_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupCascadeTestDB(t)
	defer cleanup()

	h := setupCascadeHandlers(testCtx.db)

	nonExistentID := uuid.New().String()

	w, req := makeRequest(t, http.MethodGet, "/api/v1/capabilities/"+nonExistentID+"/delete-impact", nil, map[string]string{"id": nonExistentID})
	h.GetDeleteImpact(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetDeleteImpact_WithRealizations_Integration(t *testing.T) {
	testCtx, cleanup := setupCascadeTestDB(t)
	defer cleanup()

	h := setupCascadeHandlers(testCtx.db)

	capID := uuid.New().String()
	testCtx.createTestCapability(t, capID, "Capability With Realizations", "L1")
	testCtx.insertCapabilityCreatedEvent(t, capID, "Capability With Realizations", "L1", "")

	realizationID := uuid.New().String()
	componentID := uuid.New().String()
	testCtx.createTestRealization(t, realizationID, componentID, capID)
	testCtx.insertRealizationCreatedEvent(t, realizationID, componentID, capID)

	w, req := makeRequest(t, http.MethodGet, "/api/v1/capabilities/"+capID+"/delete-impact", nil, map[string]string{"id": capID})
	h.GetDeleteImpact(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DeleteImpactResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, capID, response.CapabilityID)
	assert.False(t, response.HasDescendants)

	totalRealizations := len(response.RealizationsOnDeletedCapabilities) + len(response.RealizationsOnRetainedCapabilities)
	assert.Equal(t, 1, totalRealizations)
}

func TestCascadeDelete_LeafCapability_NoCascade_Integration(t *testing.T) {
	testCtx, cleanup := setupCascadeTestDB(t)
	defer cleanup()

	h := setupCascadeHandlers(testCtx.db)

	capID := uuid.New().String()
	testCtx.createTestCapability(t, capID, "Leaf To Delete", "L1")
	testCtx.insertCapabilityCreatedEvent(t, capID, "Leaf To Delete", "L1", "")

	body, _ := json.Marshal(DeleteCapabilityRequest{Cascade: false})
	w, req := makeRequest(t, http.MethodDelete, "/api/v1/capabilities/"+capID, body, map[string]string{"id": capID})
	h.DeleteCapability(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	time.Sleep(100 * time.Millisecond)

	capability, err := h.readModel.GetByID(tenantContext(), capID)
	require.NoError(t, err)
	assert.Nil(t, capability)
}

func TestCascadeDelete_WithChildren_NoCascade_Returns409_Integration(t *testing.T) {
	testCtx, cleanup := setupCascadeTestDB(t)
	defer cleanup()

	h := setupCascadeHandlers(testCtx.db)

	parentID := uuid.New().String()
	childID := uuid.New().String()

	testCtx.createTestCapability(t, parentID, "Parent Capability", "L1")
	testCtx.insertCapabilityCreatedEvent(t, parentID, "Parent Capability", "L1", "")

	testCtx.createTestCapabilityWithParent(t, childID, "Child Capability", "L2", parentID)
	testCtx.insertCapabilityCreatedEvent(t, childID, "Child Capability", "L2", parentID)

	body, _ := json.Marshal(DeleteCapabilityRequest{Cascade: false})
	w, req := makeRequest(t, http.MethodDelete, "/api/v1/capabilities/"+parentID, body, map[string]string{"id": parentID})
	h.DeleteCapability(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response struct {
		Error   string                    `json:"error"`
		Message string                    `json:"message"`
		Links   map[string]sharedAPI.Link `json:"_links"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response.Message, "descendants")
	assert.Contains(t, response.Links, "x-delete-impact")

	parentCapability, err := h.readModel.GetByID(tenantContext(), parentID)
	require.NoError(t, err)
	assert.NotNil(t, parentCapability)
}

func TestCascadeDelete_WithChildren_CascadeTrue_Integration(t *testing.T) {
	testCtx, cleanup := setupCascadeTestDB(t)
	defer cleanup()

	h := setupCascadeHandlers(testCtx.db)

	parentID := uuid.New().String()
	childID := uuid.New().String()
	grandchildID := uuid.New().String()

	testCtx.createTestCapability(t, parentID, "Parent", "L1")
	testCtx.insertCapabilityCreatedEvent(t, parentID, "Parent", "L1", "")

	testCtx.createTestCapabilityWithParent(t, childID, "Child", "L2", parentID)
	testCtx.insertCapabilityCreatedEvent(t, childID, "Child", "L2", parentID)

	testCtx.createTestCapabilityWithParent(t, grandchildID, "Grandchild", "L3", childID)
	testCtx.insertCapabilityCreatedEvent(t, grandchildID, "Grandchild", "L3", childID)

	body, _ := json.Marshal(DeleteCapabilityRequest{Cascade: true})
	w, req := makeRequest(t, http.MethodDelete, "/api/v1/capabilities/"+parentID, body, map[string]string{"id": parentID})
	h.DeleteCapability(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	time.Sleep(200 * time.Millisecond)

	for _, id := range []string{parentID, childID, grandchildID} {
		cap, err := h.readModel.GetByID(tenantContext(), id)
		require.NoError(t, err)
		assert.Nil(t, cap, "capability %s should have been deleted", id)
	}
}

func TestCascadeDelete_NoBody_LeafCapability_Integration(t *testing.T) {
	testCtx, cleanup := setupCascadeTestDB(t)
	defer cleanup()

	h := setupCascadeHandlers(testCtx.db)

	capID := uuid.New().String()
	testCtx.createTestCapability(t, capID, "Leaf No Body", "L1")
	testCtx.insertCapabilityCreatedEvent(t, capID, "Leaf No Body", "L1", "")

	w, req := makeRequest(t, http.MethodDelete, "/api/v1/capabilities/"+capID, nil, map[string]string{"id": capID})
	h.DeleteCapability(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	time.Sleep(100 * time.Millisecond)

	capability, err := h.readModel.GetByID(tenantContext(), capID)
	require.NoError(t, err)
	assert.Nil(t, capability)
}
