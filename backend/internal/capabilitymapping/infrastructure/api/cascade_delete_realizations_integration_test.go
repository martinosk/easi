//go:build integration
// +build integration

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/infrastructure/adapters"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type realizedHierarchy struct {
	testCtx        *testContext
	realizations   *RealizationHandlers
	capabilityRepo *repositories.CapabilityRepository
	capabilities   *CapabilityHandlers
}

type hierarchyNode struct {
	ID    string
	Name  string
	Depth int
}

func newRealizedHierarchy(t *testing.T) *realizedHierarchy {
	t.Helper()
	testCtx, cleanup := setupTestDB(t)
	t.Cleanup(cleanup)
	testCtx.setTenantContext(t)
	realizations, capabilityRepo := setupRealizationHandlers(database.NewTenantAwareDB(testCtx.db))
	return &realizedHierarchy{
		testCtx:        testCtx,
		realizations:   realizations,
		capabilityRepo: capabilityRepo,
		capabilities:   setupHandlers(testCtx.db),
	}
}

func (r *realizedHierarchy) root(t *testing.T, name string) hierarchyNode {
	return r.save(t, hierarchyNode{Name: name, Depth: 1}, "")
}

func (r *realizedHierarchy) child(t *testing.T, name string, parent hierarchyNode) hierarchyNode {
	return r.save(t, hierarchyNode{Name: name, Depth: parent.Depth + 1}, parent.ID)
}

func (r *realizedHierarchy) save(t *testing.T, node hierarchyNode, parentID string) hierarchyNode {
	t.Helper()
	capability := newCapabilityAggregate(t, node.Name, fmt.Sprintf("L%d", node.Depth), parentID)
	require.NoError(t, r.capabilityRepo.Save(tenantContext(), capability))
	r.testCtx.trackID(capability.ID())
	time.Sleep(50 * time.Millisecond)
	node.ID = capability.ID()
	return node
}

func (r *realizedHierarchy) realize(t *testing.T, applicationName string, node hierarchyNode) string {
	t.Helper()
	componentID := r.application(t, applicationName)
	r.link(t, componentID, node)
	return componentID
}

func (r *realizedHierarchy) application(t *testing.T, applicationName string) string {
	t.Helper()
	componentID := uuid.New().String()
	r.testCtx.setTenantContext(t)
	_, err := r.testCtx.db.Exec(
		"INSERT INTO capabilitymapping.capability_component_cache (tenant_id, id, name) VALUES ($1, $2, $3)",
		testTenantID(), componentID, applicationName,
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		r.testCtx.db.Exec("DELETE FROM capabilitymapping.capability_component_cache WHERE id = $1", componentID)
	})
	return componentID
}

func (r *realizedHierarchy) link(t *testing.T, componentID string, node hierarchyNode) {
	t.Helper()
	body, err := json.Marshal(LinkSystemRequest{ComponentID: componentID, RealizationLevel: "Full"})
	require.NoError(t, err)
	w, req := makeRequest(t, http.MethodPost, "/api/v1/capabilities/"+node.ID+"/systems", body, map[string]string{"id": node.ID})
	r.realizations.LinkSystemToCapability(w, req)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.NotEmpty(t, created.ID)
	r.testCtx.trackID(created.ID)
	time.Sleep(50 * time.Millisecond)
}

func (r *realizedHierarchy) reparent(t *testing.T, node, newParent hierarchyNode) {
	t.Helper()
	_, err := r.realizations.commandBus.Dispatch(tenantContext(), &commands.ChangeCapabilityParent{CapabilityID: node.ID, NewParentID: newParent.ID})
	require.NoError(t, err)
}

func (r *realizedHierarchy) cascadeDeleteFailingToSave(t *testing.T, node, failing hierarchyNode) error {
	t.Helper()
	tenantDB := database.NewTenantAwareDB(r.testCtx.db)
	handler := handlers.NewCascadeDeleteCapabilityHandler(handlers.CascadeDeleteDeps{
		Transactions:     tenantDB,
		Repository:       failingCapabilitySave{CascadeDeleteRepository: r.capabilityRepo, failingID: failing.ID},
		HierarchyService: services.NewCapabilityHierarchyService(adapters.NewCapabilityLookupAdapter(r.capabilities.readModel)),
		RealizationRM:    readmodels.NewRealizationReadModel(tenantDB),
		DependencyRM:     readmodels.NewDependencyReadModel(tenantDB),
		CommandBus:       r.capabilities.commandBus,
	})
	_, err := handler.Handle(tenantContext(), &commands.CascadeDeleteCapability{ID: node.ID, Cascade: true})
	return err
}

var errCapabilitySaveFailed = errors.New("capability save failed")

type failingCapabilitySave struct {
	handlers.CascadeDeleteRepository
	failingID string
}

func (f failingCapabilitySave) Save(ctx context.Context, capability *aggregates.Capability) error {
	if capability.ID() == f.failingID {
		return errCapabilitySaveFailed
	}
	return f.CascadeDeleteRepository.Save(ctx, capability)
}

func (r *realizedHierarchy) inheritedSource(t *testing.T, node hierarchyNode, componentID string) string {
	t.Helper()
	for _, realization := range r.realizationsOn(t, node) {
		if realization.ComponentID == componentID && realization.Origin == readmodels.OriginInherited {
			return realization.SourceCapabilityID
		}
	}
	return ""
}

func (r *realizedHierarchy) cascadeDelete(t *testing.T, node hierarchyNode) int {
	t.Helper()
	body := mustMarshal(DeleteCapabilityRequest{Cascade: true})
	w, req := makeRequest(t, http.MethodDelete, "/api/v1/capabilities/"+node.ID, body, map[string]string{"id": node.ID})
	r.capabilities.DeleteCapability(w, req)
	return w.Code
}

func (r *realizedHierarchy) realizationsOn(t *testing.T, node hierarchyNode) []readmodels.RealizationDTO {
	t.Helper()
	realizations, err := readmodels.NewRealizationReadModel(database.NewTenantAwareDB(r.testCtx.db)).GetByCapabilityID(tenantContext(), node.ID)
	require.NoError(t, err)
	return realizations
}

func (r *realizedHierarchy) componentOrigins(t *testing.T, node hierarchyNode) map[string]string {
	t.Helper()
	origins := map[string]string{}
	for _, realization := range r.realizationsOn(t, node) {
		origins[realization.ComponentID] = realization.Origin
	}
	return origins
}

func (r *realizedHierarchy) deletedRealizationEvents(t *testing.T, realizationIDs []string) int {
	t.Helper()
	r.testCtx.setTenantContext(t)
	var count int
	require.NoError(t, r.testCtx.db.QueryRow(
		"SELECT COUNT(*) FROM infrastructure.events WHERE tenant_id = $1 AND aggregate_id = ANY($2) AND event_type = 'SystemRealizationDeleted'",
		testTenantID(), pq.Array(realizationIDs),
	).Scan(&count))
	return count
}

func (r *realizedHierarchy) directRealizationIDs(t *testing.T, nodes ...hierarchyNode) []string {
	t.Helper()
	var ids []string
	for _, node := range nodes {
		for _, realization := range r.realizationsOn(t, node) {
			if realization.Origin == readmodels.OriginDirect {
				ids = append(ids, realization.ID)
			}
		}
	}
	return ids
}

func TestCascadeDelete_RemovesDirectAndInheritedRealizationsAcrossLevels_Integration(t *testing.T) {
	hierarchy := newRealizedHierarchy(t)

	finance := hierarchy.root(t, "Finance")
	accounting := hierarchy.child(t, "Accounting", finance)
	invoicing := hierarchy.child(t, "Invoicing", accounting)
	billing := hierarchy.realize(t, "Billing", invoicing)
	hierarchy.realize(t, "Ledger", accounting)

	require.Equal(t, readmodels.OriginInherited, hierarchy.componentOrigins(t, finance)[billing])
	directIDs := hierarchy.directRealizationIDs(t, finance, accounting, invoicing)
	require.Len(t, directIDs, 2)

	assert.Equal(t, http.StatusNoContent, hierarchy.cascadeDelete(t, finance))

	for _, node := range []hierarchyNode{finance, accounting, invoicing} {
		assert.Empty(t, hierarchy.realizationsOn(t, node), "realizations remain on %s", node.Name)
		capability, err := hierarchy.capabilities.readModel.GetByID(tenantContext(), node.ID)
		require.NoError(t, err)
		assert.Nil(t, capability)
	}
	assert.Equal(t, 2, hierarchy.deletedRealizationEvents(t, directIDs))
}

func TestCascadeDelete_OutsideAncestorKeepsOnlyRealizationsSourcedElsewhere_Integration(t *testing.T) {
	hierarchy := newRealizedHierarchy(t)

	operations := hierarchy.root(t, "Operations")
	finance := hierarchy.child(t, "Finance", operations)
	invoicing := hierarchy.child(t, "Invoicing", finance)
	logistics := hierarchy.child(t, "Logistics", operations)
	billing := hierarchy.realize(t, "Billing", invoicing)
	fleet := hierarchy.realize(t, "Fleet", logistics)

	require.Equal(t, map[string]string{billing: readmodels.OriginInherited, fleet: readmodels.OriginInherited}, hierarchy.componentOrigins(t, operations))

	assert.Equal(t, http.StatusNoContent, hierarchy.cascadeDelete(t, finance))

	assert.Equal(t, map[string]string{fleet: readmodels.OriginInherited}, hierarchy.componentOrigins(t, operations))
	assert.Equal(t, map[string]string{fleet: readmodels.OriginDirect}, hierarchy.componentOrigins(t, logistics))
}

func TestCascadeDelete_RealizedLeaf_Integration(t *testing.T) {
	hierarchy := newRealizedHierarchy(t)

	finance := hierarchy.root(t, "Finance")
	invoicing := hierarchy.child(t, "Invoicing", finance)
	hierarchy.realize(t, "Billing", invoicing)

	assert.Equal(t, http.StatusNoContent, hierarchy.cascadeDelete(t, invoicing))

	assert.Empty(t, hierarchy.realizationsOn(t, invoicing))
	assert.Empty(t, hierarchy.realizationsOn(t, finance))
}

func TestGetDeleteImpact_ListsEachDirectRealizationOnce_Integration(t *testing.T) {
	hierarchy := newRealizedHierarchy(t)

	finance := hierarchy.root(t, "Finance")
	accounting := hierarchy.child(t, "Accounting", finance)
	invoicing := hierarchy.child(t, "Invoicing", accounting)
	billing := hierarchy.realize(t, "Billing", invoicing)

	w, req := makeRequest(t, http.MethodGet, "/api/v1/capabilities/"+finance.ID+"/delete-impact", nil, map[string]string{"id": finance.ID})
	hierarchy.capabilities.GetDeleteImpact(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response DeleteImpactResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
	require.Len(t, response.Realizations, 1)
	assert.Equal(t, billing, response.Realizations[0].ComponentID)
	assert.Equal(t, invoicing.ID, response.Realizations[0].CapabilityID)
	assert.Equal(t, "Billing", response.Realizations[0].ComponentName)
	assert.Equal(t, readmodels.OriginDirect, response.Realizations[0].Origin)
}

type branchedApplication struct {
	operations, finance, invoicing, logistics hierarchyNode
	billing                                   string
}

func realizeInBothBranches(t *testing.T, hierarchy *realizedHierarchy) branchedApplication {
	t.Helper()
	operations := hierarchy.root(t, "Operations")
	finance := hierarchy.child(t, "Finance", operations)
	branched := branchedApplication{
		operations: operations,
		finance:    finance,
		invoicing:  hierarchy.child(t, "Invoicing", finance),
		logistics:  hierarchy.child(t, "Logistics", operations),
		billing:    hierarchy.application(t, "Billing"),
	}
	hierarchy.link(t, branched.billing, branched.invoicing)
	hierarchy.link(t, branched.billing, branched.logistics)
	require.Equal(t, branched.invoicing.ID, hierarchy.inheritedSource(t, operations, branched.billing))
	return branched
}

func TestCascadeDelete_SameApplicationInSurvivingBranch_KeepsInheritedRealization_Integration(t *testing.T) {
	hierarchy := newRealizedHierarchy(t)
	branched := realizeInBothBranches(t, hierarchy)

	assert.Equal(t, http.StatusNoContent, hierarchy.cascadeDelete(t, branched.finance))

	assert.Equal(t, branched.logistics.ID, hierarchy.inheritedSource(t, branched.operations, branched.billing))
}

func TestChangeParent_SameApplicationInRemainingBranch_KeepsInheritedRealization_Integration(t *testing.T) {
	hierarchy := newRealizedHierarchy(t)
	branched := realizeInBothBranches(t, hierarchy)
	archive := hierarchy.root(t, "Archive")

	hierarchy.reparent(t, branched.finance, archive)

	assert.Equal(t, branched.logistics.ID, hierarchy.inheritedSource(t, branched.operations, branched.billing))
	assert.Equal(t, branched.invoicing.ID, hierarchy.inheritedSource(t, archive, branched.billing))
}

func TestCascadeDelete_FailureAfterEarlierDeletes_RollsBackEverything_Integration(t *testing.T) {
	hierarchy := newRealizedHierarchy(t)
	finance := hierarchy.root(t, "Finance")
	invoicing := hierarchy.child(t, "Invoicing", finance)
	billing := hierarchy.realize(t, "Billing", invoicing)
	directIDs := hierarchy.directRealizationIDs(t, invoicing)

	err := hierarchy.cascadeDeleteFailingToSave(t, finance, finance)

	require.ErrorIs(t, err, errCapabilitySaveFailed)
	assert.Equal(t, map[string]string{billing: readmodels.OriginDirect}, hierarchy.componentOrigins(t, invoicing))
	assert.Equal(t, map[string]string{billing: readmodels.OriginInherited}, hierarchy.componentOrigins(t, finance))
	assert.Zero(t, hierarchy.deletedRealizationEvents(t, directIDs))
	for _, node := range []hierarchyNode{finance, invoicing} {
		capability, err := hierarchy.capabilities.readModel.GetByID(tenantContext(), node.ID)
		require.NoError(t, err)
		assert.NotNil(t, capability, "%s was deleted", node.Name)
	}
}
