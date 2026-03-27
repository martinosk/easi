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

func (ctx *testContext) insertEvent(t *testing.T, aggregateID, eventType, eventData string) {
	t.Helper()
	_, err := ctx.db.Exec(
		`INSERT INTO infrastructure.events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
		 VALUES ($1, $2, $3, $4, 1, NOW(), 'test-user', 'test@example.com')`,
		testTenantID(), aggregateID, eventType, eventData,
	)
	require.NoError(t, err)
}

func (ctx *testContext) createCapabilityWithEvent(t *testing.T, id, name, level string, parentID ...string) {
	parent := ""
	if len(parentID) > 0 {
		parent = parentID[0]
	}
	ctx.createTestCapability(t, id, name, level, parent)
	ctx.setTenantContext(t)
	ctx.insertEvent(t, id, "CapabilityCreated", capabilityEventData(id, name, level, parent))
}

func (ctx *testContext) insertRealizationEvent(t *testing.T, id, componentID, capabilityID string) {
	ctx.setTenantContext(t)
	ctx.insertEvent(t, id, "SystemLinkedToCapability", realizationEventData(id, componentID, capabilityID))
}

func (ctx *testContext) insertDependencyEvent(t *testing.T, id, sourceID, targetID string) {
	ctx.setTenantContext(t)
	ctx.insertEvent(t, id, "CapabilityDependencyCreated", dependencyEventData(id, sourceID, targetID))
}

func capabilityEventData(id, name, level, parentID string) string {
	return fmt.Sprintf(`{"id":"%s","name":"%s","description":"","level":"%s","parentId":"%s"}`, id, name, level, parentID)
}

func realizationEventData(id, componentID, capabilityID string) string {
	return fmt.Sprintf(`{"realizationId":"%s","componentId":"%s","capabilityId":"%s","componentName":"Test Component","realizationLevel":"Full"}`, id, componentID, capabilityID)
}

func dependencyEventData(id, sourceID, targetID string) string {
	return fmt.Sprintf(`{"dependencyId":"%s","sourceCapabilityId":"%s","targetCapabilityId":"%s","dependencyType":"Requires"}`, id, sourceID, targetID)
}

func TestGetDeleteImpact_LeafCapability_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	h := setupHandlers(testCtx.db)

	capID := uuid.New().String()
	testCtx.createCapabilityWithEvent(t, capID, "Leaf Capability", "L1")

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

	testCtx.createCapabilityWithEvent(t, l1ID, "L1 Capability", "L1")
	testCtx.createCapabilityWithEvent(t, l2ID, "L2 Capability", "L2", l1ID)
	testCtx.createCapabilityWithEvent(t, l3ID, "L3 Capability", "L3", l2ID)

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
	testCtx.createCapabilityWithEvent(t, capID, "Capability With Realizations", "L1")

	realizationID := uuid.New().String()
	componentID := uuid.New().String()
	testCtx.createTestRealization(t, realizationID, componentID, capID)
	testCtx.insertRealizationEvent(t, realizationID, componentID, capID)

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
			testCtx.createCapabilityWithEvent(t, capID, "Leaf To Delete", "L1")

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

	testCtx.createCapabilityWithEvent(t, parentID, "Parent Capability", "L1")
	testCtx.createCapabilityWithEvent(t, childID, "Child Capability", "L2", parentID)

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

	testCtx.createCapabilityWithEvent(t, parentID, "Parent", "L1")
	testCtx.createCapabilityWithEvent(t, childID, "Child", "L2", parentID)
	testCtx.createCapabilityWithEvent(t, grandchildID, "Grandchild", "L3", childID)

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
