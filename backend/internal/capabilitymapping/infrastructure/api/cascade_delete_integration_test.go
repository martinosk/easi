//go:build integration
// +build integration

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	sharedAPI "easi/backend/internal/shared/api"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capabilitySpec struct {
	ID, Name, Level, ParentID string
}

type realizationSpec struct {
	ID, ComponentID, CapabilityID string
}

type dependencySpec struct {
	ID, SourceID, TargetID string
}

type eventSpec struct {
	AggregateID, EventType, EventData string
}

func (ctx *testContext) createTestRealization(t *testing.T, spec realizationSpec) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		`INSERT INTO capabilitymapping.capability_realizations (id, component_id, capability_id, component_name, realization_level, origin, notes, tenant_id, linked_at)
		 VALUES ($1, $2, $3, 'Test Component', 'Full', 'Direct', '', $4, NOW())`,
		spec.ID, spec.ComponentID, spec.CapabilityID, testTenantID(),
	)
	require.NoError(t, err)
	ctx.trackID(spec.ID)
}

func (ctx *testContext) createTestDependency(t *testing.T, spec dependencySpec) {
	ctx.setTenantContext(t)
	_, err := ctx.db.Exec(
		`INSERT INTO capabilitymapping.capability_dependencies (id, source_capability_id, target_capability_id, dependency_type, tenant_id, created_at)
		 VALUES ($1, $2, $3, 'Requires', $4, NOW())`,
		spec.ID, spec.SourceID, spec.TargetID, testTenantID(),
	)
	require.NoError(t, err)
	ctx.trackID(spec.ID)
}

func (ctx *testContext) insertEvent(t *testing.T, spec eventSpec) {
	t.Helper()
	_, err := ctx.db.Exec(
		`INSERT INTO infrastructure.events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
		 VALUES ($1, $2, $3, $4, 1, NOW(), 'test-user', 'test@example.com')`,
		testTenantID(), spec.AggregateID, spec.EventType, spec.EventData,
	)
	require.NoError(t, err)
}

func (ctx *testContext) createCapabilityWithEvent(t *testing.T, spec capabilitySpec) {
	ctx.createTestCapability(t, spec.ID, spec.Name, spec.Level, spec.ParentID)
	ctx.setTenantContext(t)
	ctx.insertEvent(t, eventSpec{
		AggregateID: spec.ID,
		EventType:   "CapabilityCreated",
		EventData:   capabilityEventData(spec),
	})
}

func (ctx *testContext) insertRealizationEvent(t *testing.T, spec realizationSpec) {
	ctx.setTenantContext(t)
	ctx.insertEvent(t, eventSpec{
		AggregateID: spec.ID,
		EventType:   "SystemLinkedToCapability",
		EventData:   realizationEventData(spec),
	})
}

func (ctx *testContext) insertDependencyEvent(t *testing.T, spec dependencySpec) {
	ctx.setTenantContext(t)
	ctx.insertEvent(t, eventSpec{
		AggregateID: spec.ID,
		EventType:   "CapabilityDependencyCreated",
		EventData:   dependencyEventData(spec),
	})
}

func capabilityEventData(spec capabilitySpec) string {
	return fmt.Sprintf(`{"id":"%s","name":"%s","description":"","level":"%s","parentId":"%s"}`, spec.ID, spec.Name, spec.Level, spec.ParentID)
}

func realizationEventData(spec realizationSpec) string {
	return fmt.Sprintf(`{"realizationId":"%s","componentId":"%s","capabilityId":"%s","componentName":"Test Component","realizationLevel":"Full"}`, spec.ID, spec.ComponentID, spec.CapabilityID)
}

func dependencyEventData(spec dependencySpec) string {
	return fmt.Sprintf(`{"dependencyId":"%s","sourceCapabilityId":"%s","targetCapabilityId":"%s","dependencyType":"Requires"}`, spec.ID, spec.SourceID, spec.TargetID)
}

func TestGetDeleteImpact_LeafCapability_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)

	capID := uuid.New().String()
	testCtx.createCapabilityWithEvent(t, capabilitySpec{ID: capID, Name: "Leaf Capability", Level: "L1"})

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
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)

	l1ID := uuid.New().String()
	l2ID := uuid.New().String()
	l3ID := uuid.New().String()

	testCtx.createCapabilityWithEvent(t, capabilitySpec{ID: l1ID, Name: "L1 Capability", Level: "L1"})
	testCtx.createCapabilityWithEvent(t, capabilitySpec{ID: l2ID, Name: "L2 Capability", Level: "L2", ParentID: l1ID})
	testCtx.createCapabilityWithEvent(t, capabilitySpec{ID: l3ID, Name: "L3 Capability", Level: "L3", ParentID: l2ID})

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
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)

	nonExistentID := uuid.New().String()

	w, req := makeRequest(t, http.MethodGet, "/api/v1/capabilities/"+nonExistentID+"/delete-impact", nil, map[string]string{"id": nonExistentID})
	h.GetDeleteImpact(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetDeleteImpact_WithRealizations_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)

	capID := uuid.New().String()
	testCtx.createCapabilityWithEvent(t, capabilitySpec{ID: capID, Name: "Capability With Realizations", Level: "L1"})

	realizationID := uuid.New().String()
	componentID := uuid.New().String()
	rSpec := realizationSpec{ID: realizationID, ComponentID: componentID, CapabilityID: capID}
	testCtx.createTestRealization(t, rSpec)
	testCtx.insertRealizationEvent(t, rSpec)

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

func TestCascadeDelete_LeafCapability_Integration(t *testing.T) {
	tests := []struct {
		name string
		body []byte
	}{
		{"WithCascadeFalse", mustMarshal(DeleteCapabilityRequest{Cascade: false})},
		{"WithNoBody", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCtx, cleanup := setupTestDB(t)
			defer cleanup()

			h := setupHandlers(testCtx.db)

			capID := uuid.New().String()
			testCtx.createCapabilityWithEvent(t, capabilitySpec{ID: capID, Name: "Leaf To Delete", Level: "L1"})

			w, req := makeRequest(t, http.MethodDelete, "/api/v1/capabilities/"+capID, tt.body, map[string]string{"id": capID})
			h.DeleteCapability(w, req)

			assert.Equal(t, http.StatusNoContent, w.Code)

			time.Sleep(100 * time.Millisecond)

			capability, err := h.readModel.GetByID(tenantContext(), capID)
			require.NoError(t, err)
			assert.Nil(t, capability)
		})
	}
}

func TestCascadeDelete_WithChildren_NoCascade_Returns409_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)

	parentID := uuid.New().String()
	childID := uuid.New().String()

	testCtx.createCapabilityWithEvent(t, capabilitySpec{ID: parentID, Name: "Parent Capability", Level: "L1"})
	testCtx.createCapabilityWithEvent(t, capabilitySpec{ID: childID, Name: "Child Capability", Level: "L2", ParentID: parentID})

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
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)

	parentID := uuid.New().String()
	childID := uuid.New().String()
	grandchildID := uuid.New().String()

	testCtx.createCapabilityWithEvent(t, capabilitySpec{ID: parentID, Name: "Parent", Level: "L1"})
	testCtx.createCapabilityWithEvent(t, capabilitySpec{ID: childID, Name: "Child", Level: "L2", ParentID: parentID})
	testCtx.createCapabilityWithEvent(t, capabilitySpec{ID: grandchildID, Name: "Grandchild", Level: "L3", ParentID: childID})

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

func mustMarshal(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
