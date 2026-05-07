package projectors

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAssignmentReadModel struct {
	insertedAssignments []readmodels.AssignmentDTO
	deletedIDs          []string
	insertErr           error
	deleteErr           error
}

func (m *mockAssignmentReadModel) Insert(ctx context.Context, dto readmodels.AssignmentDTO) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.insertedAssignments = append(m.insertedAssignments, dto)
	return nil
}

func (m *mockAssignmentReadModel) Delete(ctx context.Context, assignmentID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedIDs = append(m.deletedIDs, assignmentID)
	return nil
}

type mockBusinessDomainReadModelForProjector struct {
	domains map[string]*readmodels.BusinessDomainDTO
	getErr  error
}

func (m *mockBusinessDomainReadModelForProjector) GetByID(ctx context.Context, id string) (*readmodels.BusinessDomainDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	domain, ok := m.domains[id]
	if !ok {
		return nil, nil
	}
	return domain, nil
}

type mockCapabilityReadModelForAssignmentProjector struct {
	capabilities map[string]*readmodels.CapabilityDTO
	getErr       error
}

func (m *mockCapabilityReadModelForAssignmentProjector) GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	cap, ok := m.capabilities[id]
	if !ok {
		return nil, nil
	}
	return cap, nil
}

func TestBusinessDomainAssignmentProjector_HandleCapabilityAssignedToDomain(t *testing.T) {
	assignmentMock := &mockAssignmentReadModel{}
	domainMock := &mockBusinessDomainReadModelForProjector{
		domains: map[string]*readmodels.BusinessDomainDTO{
			"bd-456": {
				ID:   "bd-456",
				Name: "Finance",
			},
		},
	}
	capabilityMock := &mockCapabilityReadModelForAssignmentProjector{
		capabilities: map[string]*readmodels.CapabilityDTO{
			"cap-789": {
				ID:    "cap-789",
				Name:  "Financial Reporting",
				Level: "L1",
			},
		},
	}

	projector := NewBusinessDomainAssignmentProjector(assignmentMock, domainMock, capabilityMock)

	event := events.NewCapabilityAssignedToDomain(
		"assign-123",
		"bd-456",
		"cap-789",
	)

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	ctx := context.Background()
	err = projector.ProjectEvent(ctx, "CapabilityAssignedToDomain", eventData)
	require.NoError(t, err)

	assert.Len(t, assignmentMock.insertedAssignments, 1)
	assert.Equal(t, "assign-123", assignmentMock.insertedAssignments[0].AssignmentID)
	assert.Equal(t, "bd-456", assignmentMock.insertedAssignments[0].BusinessDomainID)
	assert.Equal(t, "Finance", assignmentMock.insertedAssignments[0].BusinessDomainName)
	assert.Equal(t, "cap-789", assignmentMock.insertedAssignments[0].CapabilityID)
	assert.Equal(t, "Financial Reporting", assignmentMock.insertedAssignments[0].CapabilityName)
	assert.Equal(t, "L1", assignmentMock.insertedAssignments[0].CapabilityLevel)
	assert.WithinDuration(t, time.Now().UTC(), assignmentMock.insertedAssignments[0].AssignedAt, time.Second)
}

func TestBusinessDomainAssignmentProjector_HandleCapabilityUnassignedFromDomain(t *testing.T) {
	assignmentMock := &mockAssignmentReadModel{}
	domainMock := &mockBusinessDomainReadModelForProjector{}
	capabilityMock := &mockCapabilityReadModelForAssignmentProjector{}

	projector := NewBusinessDomainAssignmentProjector(assignmentMock, domainMock, capabilityMock)

	event := events.NewCapabilityUnassignedFromDomain(
		"assign-123",
		"bd-456",
		"cap-789",
	)

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	ctx := context.Background()
	err = projector.ProjectEvent(ctx, "CapabilityUnassignedFromDomain", eventData)
	require.NoError(t, err)

	assert.Len(t, assignmentMock.deletedIDs, 1)
	assert.Equal(t, "assign-123", assignmentMock.deletedIDs[0])
}
