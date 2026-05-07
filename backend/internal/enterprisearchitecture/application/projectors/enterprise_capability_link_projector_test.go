package projectors

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/events"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnterpriseCapabilityLinkReadModel struct {
	insertedDTOs            []readmodels.EnterpriseCapabilityLinkDTO
	deletedIDs              []string
	insertedBlocking        []readmodels.BlockingDTO
	deletedBlockingBlockers []string
	insertErr               error
	deleteErr               error
}

func (m *mockEnterpriseCapabilityLinkReadModel) Insert(ctx context.Context, dto readmodels.EnterpriseCapabilityLinkDTO) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.insertedDTOs = append(m.insertedDTOs, dto)
	return nil
}

func (m *mockEnterpriseCapabilityLinkReadModel) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedIDs = append(m.deletedIDs, id)
	return nil
}

func (m *mockEnterpriseCapabilityLinkReadModel) GetLinksForCapabilities(ctx context.Context, capabilityIDs []string) ([]readmodels.EnterpriseCapabilityLinkDTO, error) {
	return nil, nil
}

func (m *mockEnterpriseCapabilityLinkReadModel) QueryHierarchy(ctx context.Context, capabilityID string, direction readmodels.HierarchyDirection) ([]string, error) {
	return nil, nil
}

func (m *mockEnterpriseCapabilityLinkReadModel) QueryName(ctx context.Context, id string, kind readmodels.NameKind) (string, error) {
	return "", nil
}

func (m *mockEnterpriseCapabilityLinkReadModel) InsertBlocking(ctx context.Context, blocking readmodels.BlockingDTO) error {
	m.insertedBlocking = append(m.insertedBlocking, blocking)
	return nil
}

func (m *mockEnterpriseCapabilityLinkReadModel) DeleteBlockingByBlocker(ctx context.Context, blockedByCapabilityID string) error {
	m.deletedBlockingBlockers = append(m.deletedBlockingBlockers, blockedByCapabilityID)
	return nil
}

func (m *mockEnterpriseCapabilityLinkReadModel) DeleteBlockingForCapabilities(ctx context.Context, capabilityIDs []string) error {
	return nil
}

func TestEnterpriseCapabilityLinkProjector_Linked_InsertsLink(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityLinkReadModel{}
	projector := NewEnterpriseCapabilityLinkProjector(mockReadModel)

	enterpriseCapabilityID := uuid.New().String()
	domainCapabilityID := uuid.New().String()
	event := events.NewEnterpriseCapabilityLinked(
		uuid.New().String(),
		enterpriseCapabilityID,
		domainCapabilityID,
		"user@example.com",
	)

	require.NoError(t, projectEvent(t, projector, event))

	require.Len(t, mockReadModel.insertedDTOs, 1)
	dto := mockReadModel.insertedDTOs[0]
	assert.Equal(t, event.ID, dto.ID)
	assert.Equal(t, enterpriseCapabilityID, dto.EnterpriseCapabilityID)
	assert.Equal(t, domainCapabilityID, dto.DomainCapabilityID)
	assert.Equal(t, "user@example.com", dto.LinkedBy)
	assert.Equal(t, event.LinkedAt, dto.LinkedAt)
}

func TestEnterpriseCapabilityLinkProjector_Unlinked_DeletesLinkAndBlockings(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityLinkReadModel{}
	projector := NewEnterpriseCapabilityLinkProjector(mockReadModel)

	linkID := uuid.New().String()
	domainCapabilityID := uuid.New().String()
	event := events.NewEnterpriseCapabilityUnlinked(
		linkID,
		uuid.New().String(),
		domainCapabilityID,
	)

	require.NoError(t, projectEvent(t, projector, event))

	assert.Equal(t, []string{linkID}, mockReadModel.deletedIDs)
	assert.Equal(t, []string{domainCapabilityID}, mockReadModel.deletedBlockingBlockers)
}

func TestEnterpriseCapabilityLinkProjector_UnknownEvent_Ignored(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityLinkReadModel{}
	projector := NewEnterpriseCapabilityLinkProjector(mockReadModel)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mockReadModel.insertedDTOs)
	assert.Empty(t, mockReadModel.deletedIDs)
}

func TestEnterpriseCapabilityLinkProjector_StoreErrorPropagation(t *testing.T) {
	tests := []struct {
		name  string
		mock  *mockEnterpriseCapabilityLinkReadModel
		event projectableEvent
	}{
		{
			name:  "insert error during link",
			mock:  &mockEnterpriseCapabilityLinkReadModel{insertErr: errors.New("database error")},
			event: events.NewEnterpriseCapabilityLinked(uuid.New().String(), uuid.New().String(), uuid.New().String(), "user@example.com"),
		},
		{
			name:  "delete error during unlink",
			mock:  &mockEnterpriseCapabilityLinkReadModel{deleteErr: errors.New("database error")},
			event: events.NewEnterpriseCapabilityUnlinked(uuid.New().String(), uuid.New().String(), uuid.New().String()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projector := NewEnterpriseCapabilityLinkProjector(tt.mock)
			assert.Error(t, projectEvent(t, projector, tt.event))
		})
	}
}

func TestEnterpriseCapabilityLinkProjector_InvalidJSON_ReturnsError(t *testing.T) {
	mockReadModel := &mockEnterpriseCapabilityLinkReadModel{}
	projector := NewEnterpriseCapabilityLinkProjector(mockReadModel)

	err := projector.ProjectEvent(context.Background(), "EnterpriseCapabilityLinked", []byte("invalid json"))
	assert.Error(t, err)
}
