package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMetadataStore struct {
	updatedMaturityValues []maturityUpdate
	updateErr             error
	metaByID              map[string]*readmodels.DomainCapabilityMetadataDTO
	domainSubtreeUpdates  []domainSubtreeUpdate
	renamedDomains        []renamedDomain
}

type domainSubtreeUpdate struct {
	L1CapabilityID string
	BusinessDomain readmodels.BusinessDomainRef
}

type renamedDomain struct {
	BusinessDomainID string
	Name             string
}

type maturityUpdate struct {
	CapabilityID  string
	MaturityValue int
}

func (m *mockMetadataStore) GetByID(ctx context.Context, capabilityID string) (*readmodels.DomainCapabilityMetadataDTO, error) {
	return m.metaByID[capabilityID], nil
}
func (m *mockMetadataStore) Insert(ctx context.Context, dto readmodels.DomainCapabilityMetadataDTO) error {
	return nil
}
func (m *mockMetadataStore) Delete(ctx context.Context, capabilityID string) error { return nil }
func (m *mockMetadataStore) UpdateParentAndL1(ctx context.Context, update readmodels.ParentL1Update) error {
	return nil
}
func (m *mockMetadataStore) UpdateLevel(ctx context.Context, capabilityID string, newLevel string) error {
	return nil
}
func (m *mockMetadataStore) UpdateBusinessDomainForL1Subtree(ctx context.Context, l1CapabilityID string, bd readmodels.BusinessDomainRef) error {
	m.domainSubtreeUpdates = append(m.domainSubtreeUpdates, domainSubtreeUpdate{L1CapabilityID: l1CapabilityID, BusinessDomain: bd})
	return nil
}

func (m *mockMetadataStore) UpdateBusinessDomainNameForDomain(ctx context.Context, businessDomainID, name string) error {
	m.renamedDomains = append(m.renamedDomains, renamedDomain{BusinessDomainID: businessDomainID, Name: name})
	return nil
}
func (m *mockMetadataStore) RecalculateL1ForSubtree(ctx context.Context, capabilityID string) error {
	return nil
}
func (m *mockMetadataStore) UpdateMaturityValue(ctx context.Context, capabilityID string, maturityValue int) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedMaturityValues = append(m.updatedMaturityValues, maturityUpdate{CapabilityID: capabilityID, MaturityValue: maturityValue})
	return nil
}

func newMetadataProjectorWithMock(mock *mockMetadataStore) *DomainCapabilityMetadataProjector {
	domainNames := func(_ context.Context, businessDomainID string) (string, error) {
		return "Domain " + businessDomainID, nil
	}
	return NewDomainCapabilityMetadataProjector(mock, domainNames)
}

func TestMetadataProjector_AssignToDomain_ResolvesNameFromBusinessDomainLookup(t *testing.T) {
	capabilityID := uuid.New().String()
	domainID := uuid.New().String()
	mock := &mockMetadataStore{metaByID: map[string]*readmodels.DomainCapabilityMetadataDTO{
		capabilityID: {CapabilityID: capabilityID, CapabilityLevel: "L1", L1CapabilityID: capabilityID},
	}}
	projector := newMetadataProjectorWithMock(mock)

	eventData, _ := json.Marshal(capabilityAssignedToDomainEvent{
		ID:               uuid.New().String(),
		BusinessDomainID: domainID,
		CapabilityID:     capabilityID,
	})
	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.CapabilityAssignedToDomain, eventData))

	require.Len(t, mock.domainSubtreeUpdates, 1)
	update := mock.domainSubtreeUpdates[0]
	assert.Equal(t, capabilityID, update.L1CapabilityID)
	assert.Equal(t, domainID, update.BusinessDomain.ID)
	assert.Equal(t, "Domain "+domainID, update.BusinessDomain.Name,
		"the domain name must come from the business-domain read model, not from metadata rows that may not exist yet")
}

func TestMetadataProjector_BusinessDomainRenamed_UpdatesDenormalizedNames(t *testing.T) {
	domainID := uuid.New().String()
	mock := &mockMetadataStore{}
	projector := newMetadataProjectorWithMock(mock)

	eventData, _ := json.Marshal(map[string]string{"id": domainID, "name": "Marketing & Growth"})
	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.BusinessDomainUpdated, eventData))

	require.Len(t, mock.renamedDomains, 1)
	assert.Equal(t, domainID, mock.renamedDomains[0].BusinessDomainID)
	assert.Equal(t, "Marketing & Growth", mock.renamedDomains[0].Name)
}

func TestMetadataProjector_MetadataUpdated_UpdatesMaturityValue(t *testing.T) {
	tests := []struct {
		name          string
		maturityValue int
	}{
		{"positive value", 3},
		{"zero value", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockMetadataStore{}
			projector := newMetadataProjectorWithMock(mock)

			capabilityID := uuid.New().String()
			eventData, err := json.Marshal(capabilityMetadataUpdatedEvent{
				ID:            capabilityID,
				MaturityValue: tt.maturityValue,
			})
			require.NoError(t, err)

			err = projector.ProjectEvent(context.Background(), cmPL.CapabilityMetadataUpdated, eventData)
			require.NoError(t, err)

			require.Len(t, mock.updatedMaturityValues, 1)
			assert.Equal(t, capabilityID, mock.updatedMaturityValues[0].CapabilityID)
			assert.Equal(t, tt.maturityValue, mock.updatedMaturityValues[0].MaturityValue)
		})
	}
}

func TestMetadataProjector_MetadataUpdated_UnknownEvent_Ignored(t *testing.T) {
	mock := &mockMetadataStore{}
	projector := newMetadataProjectorWithMock(mock)

	err := projector.ProjectEvent(context.Background(), "SomeOtherEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mock.updatedMaturityValues)
}

func TestMetadataProjector_MetadataUpdated_InvalidJSON_ReturnsError(t *testing.T) {
	mock := &mockMetadataStore{}
	projector := newMetadataProjectorWithMock(mock)

	err := projector.ProjectEvent(context.Background(), cmPL.CapabilityMetadataUpdated, []byte("invalid"))
	assert.Error(t, err)
}

func TestMetadataProjector_MetadataUpdated_ReadModelError_ReturnsError(t *testing.T) {
	mock := &mockMetadataStore{updateErr: errors.New("db error")}
	projector := newMetadataProjectorWithMock(mock)

	eventData, _ := json.Marshal(capabilityMetadataUpdatedEvent{
		ID:            uuid.New().String(),
		MaturityValue: 5,
	})

	err := projector.ProjectEvent(context.Background(), cmPL.CapabilityMetadataUpdated, eventData)
	assert.Error(t, err)
}
