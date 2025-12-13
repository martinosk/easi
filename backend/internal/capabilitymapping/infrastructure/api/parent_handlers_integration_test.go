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

func setupParentHandlers(db *sql.DB) *CapabilityHandlers {
	tenantDB := database.NewTenantAwareDB(db)

	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")

	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	readModel := readmodels.NewCapabilityReadModel(tenantDB)

	projector := projectors.NewCapabilityProjector(readModel)
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
	changeParentHandler := handlers.NewChangeCapabilityParentHandler(capabilityRepo, readModel)

	commandBus.Register("CreateCapability", createHandler)
	commandBus.Register("UpdateCapability", updateHandler)
	commandBus.Register("UpdateCapabilityMetadata", updateMetadataHandler)
	commandBus.Register("AddCapabilityExpert", addExpertHandler)
	commandBus.Register("AddCapabilityTag", addTagHandler)
	commandBus.Register("ChangeCapabilityParent", changeParentHandler)

	return NewCapabilityHandlers(commandBus, readModel, hateoas)
}

func createCapabilityViaAPI(t *testing.T, h *CapabilityHandlers, name, level, parentID string) string {
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

	h.CreateCapability(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "Failed to create capability: %s", w.Body.String())

	var response readmodels.CapabilityDTO
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	return response.ID
}

func TestChangeCapabilityParent_SuccessfullyChangeParent_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupParentHandlers(testCtx.db)

	parentID := createCapabilityViaAPI(t, h, "Parent Capability", "L1", "")
	testCtx.trackID(parentID)

	newParentID := createCapabilityViaAPI(t, h, "New Parent Capability", "L1", "")
	testCtx.trackID(newParentID)

	childID := createCapabilityViaAPI(t, h, "Child Capability", "L2", parentID)
	testCtx.trackID(childID)

	time.Sleep(100 * time.Millisecond)

	reqBody := ChangeCapabilityParentRequest{
		ParentID: newParentID,
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(t, http.MethodPatch, "/api/v1/capabilities/"+childID+"/parent", body, map[string]string{
		"id": childID,
	})
	h.ChangeCapabilityParent(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	time.Sleep(100 * time.Millisecond)

	capability, err := h.readModel.GetByID(tenantContext(), childID)
	require.NoError(t, err)
	assert.Equal(t, newParentID, capability.ParentID)
	assert.Equal(t, "L2", capability.Level)
}

func TestChangeCapabilityParent_MakeCapabilityRoot_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupParentHandlers(testCtx.db)

	parentID := createCapabilityViaAPI(t, h, "Parent Capability", "L1", "")
	testCtx.trackID(parentID)

	childID := createCapabilityViaAPI(t, h, "Child Capability", "L2", parentID)
	testCtx.trackID(childID)

	time.Sleep(100 * time.Millisecond)

	reqBody := ChangeCapabilityParentRequest{
		ParentID: "",
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(t, http.MethodPatch, "/api/v1/capabilities/"+childID+"/parent", body, map[string]string{
		"id": childID,
	})
	h.ChangeCapabilityParent(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	time.Sleep(100 * time.Millisecond)

	capability, err := h.readModel.GetByID(tenantContext(), childID)
	require.NoError(t, err)
	assert.Equal(t, "", capability.ParentID)
	assert.Equal(t, "L1", capability.Level)
}

func TestChangeCapabilityParent_LevelAutoCalculation_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupParentHandlers(testCtx.db)

	l1ID := createCapabilityViaAPI(t, h, "L1 Capability", "L1", "")
	testCtx.trackID(l1ID)

	l2ID := createCapabilityViaAPI(t, h, "L2 Capability", "L2", l1ID)
	testCtx.trackID(l2ID)

	anotherL1ID := createCapabilityViaAPI(t, h, "Another L1", "L1", "")
	testCtx.trackID(anotherL1ID)

	anotherL2ID := createCapabilityViaAPI(t, h, "Another L2", "L2", anotherL1ID)
	testCtx.trackID(anotherL2ID)

	anotherL3ID := createCapabilityViaAPI(t, h, "Another L3", "L3", anotherL2ID)
	testCtx.trackID(anotherL3ID)

	time.Sleep(100 * time.Millisecond)

	reqBody := ChangeCapabilityParentRequest{
		ParentID: anotherL3ID,
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(t, http.MethodPatch, "/api/v1/capabilities/"+l1ID+"/parent", body, map[string]string{
		"id": l1ID,
	})
	h.ChangeCapabilityParent(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	time.Sleep(100 * time.Millisecond)

	capability, err := h.readModel.GetByID(tenantContext(), l1ID)
	require.NoError(t, err)
	assert.Equal(t, anotherL3ID, capability.ParentID)
	assert.Equal(t, "L4", capability.Level)
}

func TestChangeCapabilityParent_RejectSelfReference_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupParentHandlers(testCtx.db)

	capabilityID := createCapabilityViaAPI(t, h, "Test Capability", "L1", "")
	testCtx.trackID(capabilityID)

	time.Sleep(100 * time.Millisecond)

	reqBody := ChangeCapabilityParentRequest{
		ParentID: capabilityID,
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(t, http.MethodPatch, "/api/v1/capabilities/"+capabilityID+"/parent", body, map[string]string{
		"id": capabilityID,
	})
	h.ChangeCapabilityParent(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChangeCapabilityParent_RejectCircularReference_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupParentHandlers(testCtx.db)

	l1ID := createCapabilityViaAPI(t, h, "L1 Capability", "L1", "")
	testCtx.trackID(l1ID)

	l2ID := createCapabilityViaAPI(t, h, "L2 Capability", "L2", l1ID)
	testCtx.trackID(l2ID)

	l3ID := createCapabilityViaAPI(t, h, "L3 Capability", "L3", l2ID)
	testCtx.trackID(l3ID)

	time.Sleep(100 * time.Millisecond)

	reqBody := ChangeCapabilityParentRequest{
		ParentID: l3ID,
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(t, http.MethodPatch, "/api/v1/capabilities/"+l1ID+"/parent", body, map[string]string{
		"id": l1ID,
	})
	h.ChangeCapabilityParent(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChangeCapabilityParent_RejectL5PlusHierarchy_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupParentHandlers(testCtx.db)

	l1ID := createCapabilityViaAPI(t, h, "L1 Capability", "L1", "")
	testCtx.trackID(l1ID)

	l2ID := createCapabilityViaAPI(t, h, "L2 Capability", "L2", l1ID)
	testCtx.trackID(l2ID)

	l3ID := createCapabilityViaAPI(t, h, "L3 Capability", "L3", l2ID)
	testCtx.trackID(l3ID)

	l4ID := createCapabilityViaAPI(t, h, "L4 Capability", "L4", l3ID)
	testCtx.trackID(l4ID)

	anotherL1ID := createCapabilityViaAPI(t, h, "Another L1", "L1", "")
	testCtx.trackID(anotherL1ID)

	time.Sleep(100 * time.Millisecond)

	reqBody := ChangeCapabilityParentRequest{
		ParentID: l4ID,
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(t, http.MethodPatch, "/api/v1/capabilities/"+anotherL1ID+"/parent", body, map[string]string{
		"id": anotherL1ID,
	})
	h.ChangeCapabilityParent(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChangeCapabilityParent_RejectL5PlusHierarchyWithSubtree_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupParentHandlers(testCtx.db)

	l1ID := createCapabilityViaAPI(t, h, "L1 Capability", "L1", "")
	testCtx.trackID(l1ID)

	l2ID := createCapabilityViaAPI(t, h, "L2 Capability", "L2", l1ID)
	testCtx.trackID(l2ID)

	l3ID := createCapabilityViaAPI(t, h, "L3 Capability", "L3", l2ID)
	testCtx.trackID(l3ID)

	targetL1ID := createCapabilityViaAPI(t, h, "Target L1", "L1", "")
	testCtx.trackID(targetL1ID)

	targetL2ID := createCapabilityViaAPI(t, h, "Target L2", "L2", targetL1ID)
	testCtx.trackID(targetL2ID)

	time.Sleep(100 * time.Millisecond)

	reqBody := ChangeCapabilityParentRequest{
		ParentID: l3ID,
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(t, http.MethodPatch, "/api/v1/capabilities/"+targetL1ID+"/parent", body, map[string]string{
		"id": targetL1ID,
	})
	h.ChangeCapabilityParent(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChangeCapabilityParent_NonExistentCapability_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupParentHandlers(testCtx.db)

	parentID := createCapabilityViaAPI(t, h, "Parent Capability", "L1", "")
	testCtx.trackID(parentID)

	time.Sleep(100 * time.Millisecond)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())

	reqBody := ChangeCapabilityParentRequest{
		ParentID: parentID,
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(t, http.MethodPatch, "/api/v1/capabilities/"+nonExistentID+"/parent", body, map[string]string{
		"id": nonExistentID,
	})
	h.ChangeCapabilityParent(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChangeCapabilityParent_NonExistentParent_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupParentHandlers(testCtx.db)

	capabilityID := createCapabilityViaAPI(t, h, "Test Capability", "L1", "")
	testCtx.trackID(capabilityID)

	time.Sleep(100 * time.Millisecond)

	nonExistentParentID := fmt.Sprintf("non-existent-parent-%d", time.Now().UnixNano())

	reqBody := ChangeCapabilityParentRequest{
		ParentID: nonExistentParentID,
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(t, http.MethodPatch, "/api/v1/capabilities/"+capabilityID+"/parent", body, map[string]string{
		"id": capabilityID,
	})
	h.ChangeCapabilityParent(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChangeCapabilityParent_DescendantLevelsUpdated_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupParentHandlers(testCtx.db)

	l1ID := createCapabilityViaAPI(t, h, "L1 Capability", "L1", "")
	testCtx.trackID(l1ID)

	l2ID := createCapabilityViaAPI(t, h, "L2 Capability", "L2", l1ID)
	testCtx.trackID(l2ID)

	l3ID := createCapabilityViaAPI(t, h, "L3 Capability", "L3", l2ID)
	testCtx.trackID(l3ID)

	time.Sleep(100 * time.Millisecond)

	reqBody := ChangeCapabilityParentRequest{
		ParentID: "",
	}
	body, _ := json.Marshal(reqBody)

	w, req := makeRequest(t, http.MethodPatch, "/api/v1/capabilities/"+l2ID+"/parent", body, map[string]string{
		"id": l2ID,
	})
	h.ChangeCapabilityParent(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	time.Sleep(150 * time.Millisecond)

	l2Capability, err := h.readModel.GetByID(tenantContext(), l2ID)
	require.NoError(t, err)
	assert.Equal(t, "L1", l2Capability.Level)
	assert.Equal(t, "", l2Capability.ParentID)

	l3Capability, err := h.readModel.GetByID(tenantContext(), l3ID)
	require.NoError(t, err)
	assert.Equal(t, "L2", l3Capability.Level)
	assert.Equal(t, l2ID, l3Capability.ParentID)
}
