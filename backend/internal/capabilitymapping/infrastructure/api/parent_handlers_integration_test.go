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
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
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
	for _, event := range []string{
		cmPL.CapabilityCreated, cmPL.CapabilityUpdated, cmPL.CapabilityMetadataUpdated,
		cmPL.CapabilityExpertAdded, cmPL.CapabilityTagAdded,
		cmPL.CapabilityParentChanged, cmPL.CapabilityLevelChanged,
	} {
		eventBus.Subscribe(event, projector)
	}

	capabilityRepo := repositories.NewCapabilityRepository(eventStore)
	reparentingService := services.NewCapabilityReparentingService(adapters.NewCapabilityLookupAdapter(readModel))

	commandBus.Register("CreateCapability", handlers.NewCreateCapabilityHandler(capabilityRepo))
	commandBus.Register("UpdateCapability", handlers.NewUpdateCapabilityHandler(capabilityRepo))
	commandBus.Register("UpdateCapabilityMetadata", handlers.NewUpdateCapabilityMetadataHandler(capabilityRepo))
	commandBus.Register("AddCapabilityExpert", handlers.NewAddCapabilityExpertHandler(capabilityRepo))
	commandBus.Register("AddCapabilityTag", handlers.NewAddCapabilityTagHandler(capabilityRepo))
	commandBus.Register("ChangeCapabilityParent", handlers.NewChangeCapabilityParentHandler(capabilityRepo, readModel, realizationReadModel, reparentingService))

	return NewCapabilityHandlers(commandBus, readModel, links)
}

func (f *parentTestFixture) createCapability(capReq CreateCapabilityRequest) string {
	body, _ := json.Marshal(capReq)

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

func (f *parentTestFixture) changeParent(capabilityID string, parentReq ChangeCapabilityParentRequest) *httptest.ResponseRecorder {
	body, _ := json.Marshal(parentReq)

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

type parentChangeExpectation struct {
	parentID string
	level    string
}

func (f *parentTestFixture) changeParentAndVerify(capabilityID string, parentReq ChangeCapabilityParentRequest, expect parentChangeExpectation) {
	f.t.Helper()
	w := f.changeParent(capabilityID, parentReq)
	assert.Equal(f.t, http.StatusNoContent, w.Code)

	f.waitForProjection()

	cap := f.getCapability(capabilityID)
	assert.Equal(f.t, expect.parentID, cap.ParentID)
	assert.Equal(f.t, expect.level, cap.Level)
}

func (f *parentTestFixture) changeParentAndExpectConflict(capabilityID string, parentReq ChangeCapabilityParentRequest) {
	f.t.Helper()
	w := f.changeParent(capabilityID, parentReq)
	assert.Equal(f.t, http.StatusConflict, w.Code)
}

func (f *parentTestFixture) createHierarchy(levels ...string) []string {
	ids := make([]string, len(levels))
	parentID := ""
	for i, level := range levels {
		ids[i] = f.createCapability(CreateCapabilityRequest{
			Name:     fmt.Sprintf("%s Capability", level),
			Level:    level,
			ParentID: parentID,
		})
		parentID = ids[i]
	}
	return ids
}

func TestChangeCapabilityParent_SuccessfullyChangeParent_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	parentID := f.createCapability(CreateCapabilityRequest{Name: "Parent Capability", Level: "L1"})
	newParentID := f.createCapability(CreateCapabilityRequest{Name: "New Parent Capability", Level: "L1"})
	childID := f.createCapability(CreateCapabilityRequest{Name: "Child Capability", Level: "L2", ParentID: parentID})
	f.waitForProjection()

	f.changeParentAndVerify(childID, ChangeCapabilityParentRequest{ParentID: newParentID}, parentChangeExpectation{parentID: newParentID, level: "L2"})
}

func TestChangeCapabilityParent_MakeCapabilityRoot_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	parentID := f.createCapability(CreateCapabilityRequest{Name: "Parent Capability", Level: "L1"})
	childID := f.createCapability(CreateCapabilityRequest{Name: "Child Capability", Level: "L2", ParentID: parentID})
	f.waitForProjection()

	f.changeParentAndVerify(childID, ChangeCapabilityParentRequest{}, parentChangeExpectation{parentID: "", level: "L1"})
}

func TestChangeCapabilityParent_LevelAutoCalculation_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	l1ID := f.createCapability(CreateCapabilityRequest{Name: "L1 Capability", Level: "L1"})
	hierarchy := f.createHierarchy("L1", "L2", "L3")
	f.waitForProjection()

	f.changeParentAndVerify(l1ID, ChangeCapabilityParentRequest{ParentID: hierarchy[2]}, parentChangeExpectation{parentID: hierarchy[2], level: "L4"})
}

func TestChangeCapabilityParent_RejectSelfReference_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	capabilityID := f.createCapability(CreateCapabilityRequest{Name: "Test Capability", Level: "L1"})
	f.waitForProjection()

	f.changeParentAndExpectConflict(capabilityID, ChangeCapabilityParentRequest{ParentID: capabilityID})
}

func TestChangeCapabilityParent_RejectCircularReference_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	hierarchy := f.createHierarchy("L1", "L2", "L3")
	f.waitForProjection()

	f.changeParentAndExpectConflict(hierarchy[0], ChangeCapabilityParentRequest{ParentID: hierarchy[2]})
}

func TestChangeCapabilityParent_RejectL5PlusHierarchy_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	hierarchy := f.createHierarchy("L1", "L2", "L3", "L4")
	anotherL1ID := f.createCapability(CreateCapabilityRequest{Name: "Another L1", Level: "L1"})
	f.waitForProjection()

	f.changeParentAndExpectConflict(anotherL1ID, ChangeCapabilityParentRequest{ParentID: hierarchy[3]})
}

func TestChangeCapabilityParent_RejectL5PlusHierarchyWithSubtree_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	hierarchy := f.createHierarchy("L1", "L2", "L3")
	targetL1ID := f.createCapability(CreateCapabilityRequest{Name: "Target L1", Level: "L1"})
	_ = f.createCapability(CreateCapabilityRequest{Name: "Target L2", Level: "L2", ParentID: targetL1ID})
	f.waitForProjection()

	f.changeParentAndExpectConflict(targetL1ID, ChangeCapabilityParentRequest{ParentID: hierarchy[2]})
}

func TestChangeCapabilityParent_NonExistentCapability_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	parentID := f.createCapability(CreateCapabilityRequest{Name: "Parent Capability", Level: "L1"})
	f.waitForProjection()

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())
	w := f.changeParent(nonExistentID, ChangeCapabilityParentRequest{ParentID: parentID})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChangeCapabilityParent_NonExistentParent_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	capabilityID := f.createCapability(CreateCapabilityRequest{Name: "Test Capability", Level: "L1"})
	f.waitForProjection()

	w := f.changeParent(capabilityID, ChangeCapabilityParentRequest{ParentID: "00000000-0000-0000-0000-000000000000"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChangeCapabilityParent_DescendantLevelsUpdated_Integration(t *testing.T) {
	f, cleanup := newParentTestFixture(t)
	defer cleanup()

	l1ID := f.createCapability(CreateCapabilityRequest{Name: "L1 Capability", Level: "L1"})
	l2ID := f.createCapability(CreateCapabilityRequest{Name: "L2 Capability", Level: "L2", ParentID: l1ID})
	l3ID := f.createCapability(CreateCapabilityRequest{Name: "L3 Capability", Level: "L3", ParentID: l2ID})
	f.waitForProjection()

	w := f.changeParent(l2ID, ChangeCapabilityParentRequest{})
	assert.Equal(t, http.StatusNoContent, w.Code)

	time.Sleep(150 * time.Millisecond)

	l2Capability := f.getCapability(l2ID)
	assert.Equal(t, "L1", l2Capability.Level)
	assert.Equal(t, "", l2Capability.ParentID)

	l3Capability := f.getCapability(l3ID)
	assert.Equal(t, "L2", l3Capability.Level)
	assert.Equal(t, l2ID, l3Capability.ParentID)
}
