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
}

type maturityUpdate struct {
	CapabilityID  string
	MaturityValue int
}

func (m *mockMetadataStore) GetByID(ctx context.Context, capabilityID string) (*readmodels.DomainCapabilityMetadataDTO, error) {
	return nil, nil
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
	return nil
}
func (m *mockMetadataStore) RecalculateL1ForSubtree(ctx context.Context, capabilityID string) error {
	return nil
}
func (m *mockMetadataStore) GetSubtreeCapabilityIDs(ctx context.Context, rootID string) ([]string, error) {
	return nil, nil
}
func (m *mockMetadataStore) GetEnterpriseCapabilitiesLinkedToCapabilities(ctx context.Context, capabilityIDs []string) ([]string, error) {
	return nil, nil
}
func (m *mockMetadataStore) LookupBusinessDomainName(ctx context.Context, businessDomainID string) (string, error) {
	return "", nil
}
func (m *mockMetadataStore) UpdateMaturityValue(ctx context.Context, capabilityID string, maturityValue int) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedMaturityValues = append(m.updatedMaturityValues, maturityUpdate{CapabilityID: capabilityID, MaturityValue: maturityValue})
	return nil
}

type mockCapabilityCountUpdater struct{}

func (m *mockCapabilityCountUpdater) DecrementLinkCount(ctx context.Context, id string) error {
	return nil
}
func (m *mockCapabilityCountUpdater) RecalculateDomainCount(ctx context.Context, enterpriseCapabilityID string) error {
	return nil
}

type mockCapabilityLinkStore struct{}

func (m *mockCapabilityLinkStore) GetByDomainCapabilityID(ctx context.Context, domainCapabilityID string) (*readmodels.EnterpriseCapabilityLinkDTO, error) {
	return nil, nil
}
func (m *mockCapabilityLinkStore) Delete(ctx context.Context, id string) error { return nil }
func (m *mockCapabilityLinkStore) DeleteBlockingByBlocker(ctx context.Context, blockedByCapabilityID string) error {
	return nil
}

func newMetadataProjectorWithMock(mock *mockMetadataStore) *DomainCapabilityMetadataProjector {
	return NewDomainCapabilityMetadataProjector(mock, &mockCapabilityCountUpdater{}, &mockCapabilityLinkStore{})
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
