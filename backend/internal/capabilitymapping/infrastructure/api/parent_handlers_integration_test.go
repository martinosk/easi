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

	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/projectors"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/infrastructure/adapters"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type parentTestFixture struct {
	t       *testing.T
	testCtx *testContext
	h       *CapabilityHandlers
}

func newParentTestFixture(t *testing.T) (*parentTestFixture, func()) {
	testCtx, cleanup := setupTestDB(t)
	h := setupParentHandlers(testCtx.db)
	return &parentTestFixture{t: t, testCtx: testCtx, h: h}, cleanup
}

func setupParentHandlers(db *sql.DB) *CapabilityHandlers {
	tenantDB := database.NewTenantAwareDB(db)

	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")
	links := NewCapabilityMappingLinks(hateoas)

	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	readModel := readmodels.NewCapabilityReadModel(tenantDB)
	realizationReadModel := readmodels.NewRealizationReadModel(tenantDB)
	assignmentReadModel := readmodels.NewDomainCapabilityAssignmentReadModel(tenantDB)

	projector := projectors.NewCapabilityProjector(readModel, assignmentReadModel)
	eventBus.Subscribe("CapabilityCreated", projector)
	eventBus.Subscribe("CapabilityUpdated", projector)
	eventBus.Subscribe("CapabilityMetadataUpdated", projector)
	eventBus.Subscribe("CapabilityExpertAdded", projector)
	eventBus.Subscribe("CapabilityTagAdded", projector)
	eventBus.Subscribe("CapabilityParentChanged", projector)

	capabilityRepo := repositories.NewCapabilityRepository(eventStore)
	createHandler := handlers.NewCreateCapabilityHandler(capabilityRepo)
	updateHandler := handlers.NewUpdateCapabilityHandler(capabilityRepo)
	updateMetadataHandler := handlers.NewUpdateCapabilityMetadataHandler(capabilityRepo)
	addExpertHandler := handlers.NewAddCapabilityExpertHandler(capabilityRepo)
	addTagHandler := handlers.NewAddCapabilityTagHandler(capabilityRepo)
	reparentingService := services.NewCapabilityReparentingService(adapters.NewCapabilityLookupAdapter(readModel))
	changeParentHandler := handlers.NewChangeCapabilityParentHandler(capabilityRepo, readModel, realizationReadModel, reparentingService)

	commandBus.Register("CreateCapability", createHandler)
	commandBus.Register("UpdateCapability", updateHandler)
	commandBus.Register("UpdateCapabilityMetadata", updateMetadataHandler)
	commandBus.Register("AddCapabilityExpert", addExpertHandler)
	commandBus.Register("AddCapabilityTag", addTagHandler)
	commandBus.Register("ChangeCapabilityParent", changeParentHandler)

	return NewCapabilityHandlers(commandBus, readModel, links)
}

func (f *parentTestFixture) createCapability(name, level, parentID string) string {
	reqBody := CreateCapabilityRequest{
		Name:     name,
		Level:    level,
		ParentID: parentID,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	w := httptest.NewRecorder()

	f.h.CreateCapability(w, req)
	require.Equal(f.t, http.StatusCreated, w.Code, "Failed to create capability: %s", w.Body.String())

	var response readmodels.CapabilityDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(f.t, err)

	f.testCtx.trackID(response.ID)
	return response.ID
}

func (f *parentTestFixture) changeParent(capabilityID, newParentID string) *httptest.ResponseRecorder {
	reqBody := ChangeCapabilityParentRequest{
		ParentID: newParentID,
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(f.t, http.MethodPatch, "/api/v1/capabilities/"+capabilityID+"/parent", body, map[string]string{
		"id": capabilityID,
	})
	f.h.ChangeCapabilityParent(w, req)
	return w
}

func (f *parentTestFixture) waitForProjection() {
	time.Sleep(100 * time.Millisecond)
}

func (f *parentTestFixture) getCapability(id string) *readmodels.CapabilityDTO {
	capability, err := f.h.readModel.GetByID(tenantContext(), id)
	require.NoError(f.t, err)
	return capability
}

func TestChangeCapabilityParent_SuccessfullyChangeParent_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	parentID := f.createCapability("Parent Capability", "L1", "")
	newParentID := f.createCapability("New Parent Capability", "L1", "")
	childID := f.createCapability("Child Capability", "L2", parentID)
	f.waitForProjection()

	w := f.changeParent(childID, newParentID)
	assert.Equal(t, http.StatusNoContent, w.Code)

	f.waitForProjection()

	capability := f.getCapability(childID)
	assert.Equal(t, newParentID, capability.ParentID)
	assert.Equal(t, "L2", capability.Level)
}

func TestChangeCapabilityParent_MakeCapabilityRoot_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	parentID := f.createCapability("Parent Capability", "L1", "")
	childID := f.createCapability("Child Capability", "L2", parentID)
	f.waitForProjection()

	w := f.changeParent(childID, "")
	assert.Equal(t, http.StatusNoContent, w.Code)

	f.waitForProjection()

	capability := f.getCapability(childID)
	assert.Equal(t, "", capability.ParentID)
	assert.Equal(t, "L1", capability.Level)
}

func TestChangeCapabilityParent_LevelAutoCalculation_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	l1ID := f.createCapability("L1 Capability", "L1", "")
	anotherL1ID := f.createCapability("Another L1", "L1", "")
	anotherL2ID := f.createCapability("Another L2", "L2", anotherL1ID)
	anotherL3ID := f.createCapability("Another L3", "L3", anotherL2ID)
	f.waitForProjection()

	w := f.changeParent(l1ID, anotherL3ID)
	assert.Equal(t, http.StatusNoContent, w.Code)

	f.waitForProjection()

	capability := f.getCapability(l1ID)
	assert.Equal(t, anotherL3ID, capability.ParentID)
	assert.Equal(t, "L4", capability.Level)
}

func TestChangeCapabilityParent_RejectSelfReference_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	capabilityID := f.createCapability("Test Capability", "L1", "")
	f.waitForProjection()

	w := f.changeParent(capabilityID, capabilityID)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestChangeCapabilityParent_RejectCircularReference_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	l1ID := f.createCapability("L1 Capability", "L1", "")
	l2ID := f.createCapability("L2 Capability", "L2", l1ID)
	l3ID := f.createCapability("L3 Capability", "L3", l2ID)
	f.waitForProjection()

	w := f.changeParent(l1ID, l3ID)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestChangeCapabilityParent_RejectL5PlusHierarchy_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	l1ID := f.createCapability("L1 Capability", "L1", "")
	l2ID := f.createCapability("L2 Capability", "L2", l1ID)
	l3ID := f.createCapability("L3 Capability", "L3", l2ID)
	l4ID := f.createCapability("L4 Capability", "L4", l3ID)
	anotherL1ID := f.createCapability("Another L1", "L1", "")
	f.waitForProjection()

	w := f.changeParent(anotherL1ID, l4ID)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestChangeCapabilityParent_RejectL5PlusHierarchyWithSubtree_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	l1ID := f.createCapability("L1 Capability", "L1", "")
	l2ID := f.createCapability("L2 Capability", "L2", l1ID)
	l3ID := f.createCapability("L3 Capability", "L3", l2ID)
	targetL1ID := f.createCapability("Target L1", "L1", "")
	_ = f.createCapability("Target L2", "L2", targetL1ID)
	f.waitForProjection()

	w := f.changeParent(targetL1ID, l3ID)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestChangeCapabilityParent_NonExistentCapability_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	parentID := f.createCapability("Parent Capability", "L1", "")
	f.waitForProjection()

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())
	w := f.changeParent(nonExistentID, parentID)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChangeCapabilityParent_NonExistentParent_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	capabilityID := f.createCapability("Test Capability", "L1", "")
	f.waitForProjection()

	w := f.changeParent(capabilityID, "00000000-0000-0000-0000-000000000000")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChangeCapabilityParent_DescendantLevelsUpdated_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	l1ID := f.createCapability("L1 Capability", "L1", "")
	l2ID := f.createCapability("L2 Capability", "L2", l1ID)
	l3ID := f.createCapability("L3 Capability", "L3", l2ID)
	f.waitForProjection()

	w := f.changeParent(l2ID, "")
	assert.Equal(t, http.StatusNoContent, w.Code)

	time.Sleep(150 * time.Millisecond)

	l2Capability := f.getCapability(l2ID)
	assert.Equal(t, "L1", l2Capability.Level)
	assert.Equal(t, "", l2Capability.ParentID)

	l3Capability := f.getCapability(l3ID)
	assert.Equal(t, "L2", l3Capability.Level)
	assert.Equal(t, l2ID, l3Capability.ParentID)
}
