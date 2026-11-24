//go:build integration
// +build integration

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/projectors"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/adapters"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRealizationHandlers(db *database.TenantAwareDB) (*RealizationHandlers, *repositories.CapabilityRepository) {
	eventStore := eventstore.NewPostgresEventStore(db)
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")
	links := NewCapabilityMappingLinks(hateoas)

	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	realizationRM := readmodels.NewRealizationReadModel(db)
	capabilityRM := readmodels.NewCapabilityReadModel(db)
	assignmentRM := readmodels.NewDomainCapabilityAssignmentReadModel(db)
	componentCacheRM := readmodels.NewComponentCacheReadModel(db)

	capabilityProjector := projectors.NewCapabilityProjector(capabilityRM, assignmentRM)
	eventBus.Subscribe("CapabilityCreated", capabilityProjector)
	eventBus.Subscribe("CapabilityUpdated", capabilityProjector)
	eventBus.Subscribe("CapabilityMetadataUpdated", capabilityProjector)
	eventBus.Subscribe("CapabilityParentChanged", capabilityProjector)
	eventBus.Subscribe("CapabilityDeleted", capabilityProjector)

	realizationProjector := projectors.NewRealizationProjector(realizationRM, componentCacheRM)
	eventBus.Subscribe("SystemLinkedToCapability", realizationProjector)
	eventBus.Subscribe("CapabilityRealizationsInherited", realizationProjector)
	eventBus.Subscribe("CapabilityRealizationsUninherited", realizationProjector)

	realizationRepo := repositories.NewRealizationRepository(eventStore)
	capabilityRepo := repositories.NewCapabilityRepository(eventStore)
	commandBus.Register("LinkSystemToCapability", handlers.NewLinkSystemToCapabilityHandler(realizationRepo, capabilityRepo, capabilityRM, componentCacheRM))
	commandBus.Register("ChangeCapabilityParent", handlers.NewChangeCapabilityParentHandler(capabilityRepo, capabilityRM, realizationRM, services.NewCapabilityReparentingService(adapters.NewCapabilityLookupAdapter(capabilityRM))))

	return NewRealizationHandlers(commandBus, realizationRM, links), capabilityRepo
}

func newCapabilityAggregate(t *testing.T, name, level string, parentID string) *aggregates.Capability {
	t.Helper()

	capabilityName, err := valueobjects.NewCapabilityName(name)
	require.NoError(t, err)
	capabilityLevel, err := valueobjects.NewCapabilityLevel(level)
	require.NoError(t, err)

	parent := valueobjects.CapabilityID{}
	if parentID != "" {
		parent, err = valueobjects.NewCapabilityIDFromString(parentID)
		require.NoError(t, err)
	}

	capability, err := aggregates.NewCapability(capabilityName, valueobjects.MustNewDescription(""), parent, capabilityLevel)
	require.NoError(t, err)
	return capability
}

type linkedHierarchy struct {
	handlers         *RealizationHandlers
	l1ID, l2ID, l3ID string
}

func linkComponentToNewL3(t *testing.T, testCtx *testContext) linkedHierarchy {
	t.Helper()
	tenantDB := database.NewTenantAwareDB(testCtx.db)
	h, capabilityRepo := setupRealizationHandlers(tenantDB)

	l1 := newCapabilityAggregate(t, "Root", "L1", "")
	require.NoError(t, capabilityRepo.Save(tenantContext(), l1))
	l1ID := l1.ID()

	l2 := newCapabilityAggregate(t, "Parent", "L2", l1ID)
	require.NoError(t, capabilityRepo.Save(tenantContext(), l2))
	l2ID := l2.ID()

	l3 := newCapabilityAggregate(t, "Child", "L3", l2ID)
	require.NoError(t, capabilityRepo.Save(tenantContext(), l3))
	l3ID := l3.ID()

	componentID := uuid.New().String()
	testCtx.trackID(l1ID)
	testCtx.trackID(l2ID)
	testCtx.trackID(l3ID)
	time.Sleep(120 * time.Millisecond)

	var err error

	_, err = testCtx.db.Exec(
		"INSERT INTO capabilitymapping.capability_component_cache (tenant_id, id, name) VALUES ($1, $2, $3)",
		testTenantID(), componentID, "Component A",
	)
	require.NoError(t, err)

	reqBody := LinkSystemRequest{
		ComponentID:      componentID,
		RealizationLevel: "Partial",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities/"+l3ID+"/systems", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", l3ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.LinkSystemToCapability(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	time.Sleep(150 * time.Millisecond)
	testCtx.setTenantContext(t)

	return linkedHierarchy{handlers: h, l1ID: l1ID, l2ID: l2ID, l3ID: l3ID}
}

func TestLinkSystemToCapability_InheritsToAncestors_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	testCtx.setTenantContext(t)

	linked := linkComponentToNewL3(t, testCtx)
	l1ID, l2ID, l3ID := linked.l1ID, linked.l2ID, linked.l3ID
	var err error

	var directCount int
	err = testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM capabilitymapping.capability_realizations WHERE tenant_id = $1 AND capability_id = $2 AND origin = 'Direct'",
		testTenantID(), l3ID,
	).Scan(&directCount)
	require.NoError(t, err)
	assert.Equal(t, 1, directCount)

	var inheritedCount int
	err = testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM capabilitymapping.capability_realizations WHERE tenant_id = $1 AND capability_id IN ($2, $3) AND origin = 'Inherited'",
		testTenantID(), l2ID, l1ID,
	).Scan(&inheritedCount)
	require.NoError(t, err)
	assert.Equal(t, 2, inheritedCount)
}

func getSystemsAsArchitect(t *testing.T, h *RealizationHandlers, capabilityID string) []readmodels.RealizationDTO {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities/"+capabilityID+"/systems", nil)
	req = withTestTenant(req)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", capabilityID)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = sharedctx.WithActor(ctx, sharedctx.NewActor("u1", "u@example.com", sharedctx.RoleArchitect))
	w := httptest.NewRecorder()
	h.GetSystemsByCapability(w, req.WithContext(ctx))
	require.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []readmodels.RealizationDTO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	require.Len(t, response.Data, 1)
	return response.Data
}

func TestGetSystemsByCapability_OnlyDirectRealizationsCarryEditAndDelete_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()
	testCtx.setTenantContext(t)

	linked := linkComponentToNewL3(t, testCtx)

	direct := getSystemsAsArchitect(t, linked.handlers, linked.l3ID)[0]
	assert.Equal(t, readmodels.OriginDirect, direct.Origin)
	assert.Contains(t, direct.Links, "edit")
	assert.Contains(t, direct.Links, "delete")

	inherited := getSystemsAsArchitect(t, linked.handlers, linked.l2ID)[0]
	assert.Equal(t, readmodels.OriginInherited, inherited.Origin)
	assert.NotContains(t, inherited.Links, "edit")
	assert.NotContains(t, inherited.Links, "delete")
	assert.Contains(t, inherited.Links, "self")
}
