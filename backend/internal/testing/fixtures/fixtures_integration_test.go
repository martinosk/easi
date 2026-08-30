//go:build integration

package fixtures

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilityFixtures_CreateCapability(t *testing.T) {
	tc := NewTestContext(t)
	cf := NewCapabilityFixtures(tc)

	id := cf.CreateL1Capability("Customer Management")

	capability := cf.GetCapability(id)
	require.NotNil(t, capability)
	assert.Equal(t, "Customer Management", capability.Name)
	assert.Equal(t, "L1", capability.Level)
}

func TestCapabilityFixtures_CreateHierarchy(t *testing.T) {
	tc := NewTestContext(t)
	cf := NewCapabilityFixtures(tc)

	parentID := cf.CreateL1Capability("Customer Management")
	childID := cf.CreateChildCapability("Customer Onboarding", parentID, "L2")

	parent := cf.GetCapability(parentID)
	child := cf.GetCapability(childID)

	require.NotNil(t, parent)
	require.NotNil(t, child)
	assert.Equal(t, "Customer Management", parent.Name)
	assert.Equal(t, "Customer Onboarding", child.Name)
	assert.Equal(t, parentID, child.ParentID)
}

func TestBusinessDomainFixtures_CreateAndAssign(t *testing.T) {
	tc := NewTestContext(t)
	cf := NewCapabilityFixtures(tc)
	bf := NewBusinessDomainFixtures(tc)

	capabilityID := cf.CreateL1Capability("Order Processing")
	domainID := bf.CreateDomain("Sales")

	domain := bf.GetDomain(domainID)
	require.NotNil(t, domain)
	assert.Equal(t, "Sales", domain.Name)

	assignmentID := bf.AssignCapabilityToDomain(capabilityID, domainID)
	require.NotEmpty(t, assignmentID)
}

func TestApplicationFixtures_CreateComponent(t *testing.T) {
	tc := NewTestContext(t)
	af := NewApplicationFixtures(tc)

	componentID := af.CreateApplication("CRM System")

	component := af.GetComponent(componentID)
	require.NotNil(t, component)
	assert.Equal(t, "CRM System", component.Name)
}

func TestTestContext_Cleanup_RemovesArchitectureDirectionCacheRows(t *testing.T) {
	tc := NewTestContext(t)
	id := "ad-cache-cleanup-" + tc.TenantID.Value()
	tc.TrackID(id)
	tc.setTenantContext()

	_, err := tc.DB.Exec(
		`INSERT INTO architecturedirection.capability_node_cache (tenant_id, capability_id, capability_name, capability_level, l1_capability_id, maturity_value) VALUES ($1, $2, 'Cleanup Test', 'L1', $2, 12)`,
		tc.TenantID.Value(), id,
	)
	require.NoError(t, err)
	_, err = tc.DB.Exec(
		`INSERT INTO architecturedirection.enterprise_capability_cache (tenant_id, id, name, active) VALUES ($1, $2, 'Cleanup Test', true)`,
		tc.TenantID.Value(), id,
	)
	require.NoError(t, err)
	_, err = tc.DB.Exec(
		`INSERT INTO architecturedirection.realization_cache (tenant_id, realization_id, capability_id, component_id) VALUES ($1, $2, $2, $2)`,
		tc.TenantID.Value(), id,
	)
	require.NoError(t, err)
	_, err = tc.DB.Exec(
		`INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name) VALUES ($1, 'application', $2, 'Cleanup Test')`,
		tc.TenantID.Value(), id,
	)
	require.NoError(t, err)

	tc.cleanup()

	var count int
	require.NoError(t, tc.DB.QueryRow(`SELECT COUNT(*) FROM architecturedirection.capability_node_cache WHERE capability_id = $1`, id).Scan(&count))
	assert.Zero(t, count, "capability_node_cache row must be cleaned up")
	require.NoError(t, tc.DB.QueryRow(`SELECT COUNT(*) FROM architecturedirection.enterprise_capability_cache WHERE id = $1`, id).Scan(&count))
	assert.Zero(t, count, "enterprise_capability_cache row must be cleaned up")
	require.NoError(t, tc.DB.QueryRow(`SELECT COUNT(*) FROM architecturedirection.realization_cache WHERE realization_id = $1`, id).Scan(&count))
	assert.Zero(t, count, "realization_cache row must be cleaned up")
	require.NoError(t, tc.DB.QueryRow(`SELECT COUNT(*) FROM architecturedirection.reference_name_cache WHERE entity_id = $1`, id).Scan(&count))
	assert.Zero(t, count, "reference_name_cache row must be cleaned up")
}

func TestCombinedFixtures_FullScenario(t *testing.T) {
	tc := NewTestContext(t)
	cf := NewCapabilityFixtures(tc)
	bf := NewBusinessDomainFixtures(tc)
	af := NewApplicationFixtures(tc)

	capabilityID := cf.CreateL1Capability("Payment Processing")
	domainID := bf.CreateDomain("Finance")
	componentID := af.CreateApplication("Payment Gateway")

	bf.AssignCapabilityToDomain(capabilityID, domainID)

	capability := cf.GetCapability(capabilityID)
	domain := bf.GetDomain(domainID)
	component := af.GetComponent(componentID)

	assert.NotNil(t, capability)
	assert.NotNil(t, domain)
	assert.NotNil(t, component)
	assert.Equal(t, "Payment Processing", capability.Name)
	assert.Equal(t, "Finance", domain.Name)
	assert.Equal(t, "Payment Gateway", component.Name)
}
