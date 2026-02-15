//go:build integration
// +build integration

package projectors

import (
	"encoding/json"
	"fmt"
	"testing"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/database"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type effectiveBDTestHarness struct {
	testCtx     *realizationProjectorIntegrationContext
	projector   *EffectiveBusinessDomainProjector
	effectiveBD *readmodels.CMEffectiveBusinessDomainReadModel
}

func setupEffectiveBDTestHarness(t *testing.T) (*effectiveBDTestHarness, func()) {
	testCtx, cleanup := setupRealizationProjectorIntegrationDB(t)
	testCtx.setTenantContext(t)
	tenantDB := database.NewTenantAwareDB(testCtx.db)
	effectiveBDRM := readmodels.NewCMEffectiveBusinessDomainReadModel(tenantDB)
	return &effectiveBDTestHarness{
		testCtx:     testCtx,
		projector:   NewEffectiveBusinessDomainProjector(effectiveBDRM, readmodels.NewBusinessDomainReadModel(tenantDB), readmodels.NewCapabilityReadModel(tenantDB)),
		effectiveBD: effectiveBDRM,
	}, cleanup
}

func TestEffectiveBDProjectorIntegration_L1AssignedToDomain_PropagatesToChildren(t *testing.T) {
	h, cleanup := setupEffectiveBDTestHarness(t)
	defer cleanup()

	l1ID := uuid.New().String()
	l2ID := uuid.New().String()
	l3ID := uuid.New().String()
	domainID := uuid.New().String()
	suffix := l1ID[:8]
	domainName := fmt.Sprintf("Finance-%s", suffix)

	h.testCtx.setTenantContext(t)
	for _, q := range []struct {
		sql  string
		args []any
	}{
		{"INSERT INTO capabilities (id, name, description, level, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, '', 'L1', 'default', 'Genesis', 'Active', NOW())", []any{l1ID, fmt.Sprintf("Payment Processing-%s", suffix)}},
		{"INSERT INTO capabilities (id, name, description, level, parent_id, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, '', 'L2', $3, 'default', 'Genesis', 'Active', NOW())", []any{l2ID, fmt.Sprintf("Card Payments-%s", suffix), l1ID}},
		{"INSERT INTO capabilities (id, name, description, level, parent_id, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, '', 'L3', $3, 'default', 'Genesis', 'Active', NOW())", []any{l3ID, fmt.Sprintf("Visa Processing-%s", suffix), l2ID}},
		{"INSERT INTO business_domains (id, tenant_id, name, description, capability_count, created_at) VALUES ($1, 'default', $2, '', 0, NOW())", []any{domainID, domainName}},
	} {
		_, err := h.testCtx.db.Exec(q.sql, q.args...)
		require.NoError(t, err)
	}

	defer func() {
		h.testCtx.setTenantContext(t)
		h.testCtx.db.Exec("DELETE FROM cm_effective_business_domain WHERE tenant_id = 'default' AND capability_id IN ($1, $2, $3)", l1ID, l2ID, l3ID)
		h.testCtx.db.Exec("DELETE FROM capabilities WHERE id IN ($1, $2, $3)", l1ID, l2ID, l3ID)
		h.testCtx.db.Exec("DELETE FROM business_domains WHERE id = $1", domainID)
	}()

	ctx := tenantContext()
	for _, ev := range []struct {
		eventType string
		event     any
	}{
		{"CapabilityCreated", events.NewCapabilityCreated(l1ID, fmt.Sprintf("Payment Processing-%s", suffix), "", "", "L1")},
		{"CapabilityCreated", events.NewCapabilityCreated(l2ID, fmt.Sprintf("Card Payments-%s", suffix), "", l1ID, "L2")},
		{"CapabilityCreated", events.NewCapabilityCreated(l3ID, fmt.Sprintf("Visa Processing-%s", suffix), "", l2ID, "L3")},
	} {
		eventData, err := json.Marshal(ev.event)
		require.NoError(t, err)
		require.NoError(t, h.projector.ProjectEvent(ctx, ev.eventType, eventData))
	}

	for _, tc := range []struct {
		capID  string
		wantL1 string
	}{
		{l1ID, l1ID}, {l2ID, l1ID}, {l3ID, l1ID},
	} {
		row, err := h.effectiveBD.GetByCapabilityID(ctx, tc.capID)
		require.NoError(t, err)
		require.NotNil(t, row)
		assert.Equal(t, tc.wantL1, row.L1CapabilityID)
	}

	assignEvent := events.NewCapabilityAssignedToDomain(uuid.New().String(), domainID, l1ID)
	eventData, err := json.Marshal(assignEvent)
	require.NoError(t, err)
	require.NoError(t, h.projector.ProjectEvent(ctx, "CapabilityAssignedToDomain", eventData))

	for _, capID := range []string{l1ID, l2ID, l3ID} {
		row, err := h.effectiveBD.GetByCapabilityID(ctx, capID)
		require.NoError(t, err)
		require.NotNil(t, row, "Expected row for %s", capID)
		assert.Equal(t, domainID, row.BusinessDomainID, "BD should propagate to %s", capID)
		assert.Equal(t, domainName, row.BusinessDomainName, "BD name should propagate to %s", capID)
	}
}

func TestEffectiveBDProjectorIntegration_MoveSubtreeToNewParent(t *testing.T) {
	h, cleanup := setupEffectiveBDTestHarness(t)
	defer cleanup()

	l1A := uuid.New().String()
	l1B := uuid.New().String()
	l2Child := uuid.New().String()
	domainA := uuid.New().String()
	domainB := uuid.New().String()
	suffix := l1A[:8]
	domainAName := fmt.Sprintf("Domain A-%s", suffix)
	domainBName := fmt.Sprintf("Domain B-%s", suffix)

	h.testCtx.setTenantContext(t)
	for _, q := range []struct {
		sql  string
		args []any
	}{
		{"INSERT INTO capabilities (id, name, description, level, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, '', 'L1', 'default', 'Genesis', 'Active', NOW())", []any{l1A, fmt.Sprintf("L1-A-%s", suffix)}},
		{"INSERT INTO capabilities (id, name, description, level, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, '', 'L1', 'default', 'Genesis', 'Active', NOW())", []any{l1B, fmt.Sprintf("L1-B-%s", suffix)}},
		{"INSERT INTO capabilities (id, name, description, level, parent_id, tenant_id, maturity_level, status, created_at) VALUES ($1, $2, '', 'L2', $3, 'default', 'Genesis', 'Active', NOW())", []any{l2Child, fmt.Sprintf("L2-Child-%s", suffix), l1A}},
		{"INSERT INTO business_domains (id, tenant_id, name, description, capability_count, created_at) VALUES ($1, 'default', $2, '', 0, NOW())", []any{domainA, domainAName}},
		{"INSERT INTO business_domains (id, tenant_id, name, description, capability_count, created_at) VALUES ($1, 'default', $2, '', 0, NOW())", []any{domainB, domainBName}},
	} {
		_, err := h.testCtx.db.Exec(q.sql, q.args...)
		require.NoError(t, err)
	}

	defer func() {
		h.testCtx.setTenantContext(t)
		h.testCtx.db.Exec("DELETE FROM cm_effective_business_domain WHERE tenant_id = 'default' AND capability_id IN ($1, $2, $3)", l1A, l1B, l2Child)
		h.testCtx.db.Exec("DELETE FROM capabilities WHERE id IN ($1, $2, $3)", l1A, l1B, l2Child)
		h.testCtx.db.Exec("DELETE FROM business_domains WHERE id IN ($1, $2)", domainA, domainB)
	}()

	ctx := tenantContext()
	for _, ev := range []struct {
		eventType string
		event     any
	}{
		{"CapabilityCreated", events.NewCapabilityCreated(l1A, fmt.Sprintf("L1-A-%s", suffix), "", "", "L1")},
		{"CapabilityCreated", events.NewCapabilityCreated(l1B, fmt.Sprintf("L1-B-%s", suffix), "", "", "L1")},
		{"CapabilityCreated", events.NewCapabilityCreated(l2Child, fmt.Sprintf("L2-Child-%s", suffix), "", l1A, "L2")},
		{"CapabilityAssignedToDomain", events.NewCapabilityAssignedToDomain(uuid.New().String(), domainA, l1A)},
		{"CapabilityAssignedToDomain", events.NewCapabilityAssignedToDomain(uuid.New().String(), domainB, l1B)},
	} {
		eventData, err := json.Marshal(ev.event)
		require.NoError(t, err)
		require.NoError(t, h.projector.ProjectEvent(ctx, ev.eventType, eventData))
	}

	child, err := h.effectiveBD.GetByCapabilityID(ctx, l2Child)
	require.NoError(t, err)
	assert.Equal(t, domainA, child.BusinessDomainID, "Before move: should belong to Domain A")
	assert.Equal(t, l1A, child.L1CapabilityID)

	moveEvent := events.NewCapabilityParentChanged(l2Child, l1A, l1B, "L2", "L2")
	eventData, err := json.Marshal(moveEvent)
	require.NoError(t, err)
	require.NoError(t, h.projector.ProjectEvent(ctx, "CapabilityParentChanged", eventData))

	child, err = h.effectiveBD.GetByCapabilityID(ctx, l2Child)
	require.NoError(t, err)
	assert.Equal(t, domainB, child.BusinessDomainID, "After move: should belong to Domain B")
	assert.Equal(t, l1B, child.L1CapabilityID, "After move: L1 should be L1-B")
}
