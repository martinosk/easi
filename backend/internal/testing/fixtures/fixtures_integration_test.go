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
