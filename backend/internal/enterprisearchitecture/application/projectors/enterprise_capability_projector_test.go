package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/events"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnterpriseCapabilityReadModel struct {
	insertedDTOs            []readmodels.EnterpriseCapabilityDTO
	updatedParams           []readmodels.UpdateCapabilityParams
	deletedIDs              []string
	incrementedLinkIDs      []string
	decrementedLinkIDs      []string
	recalculatedDomainIDs   []string
	updatedTargetMaturities []targetMaturityUpdate

	insertErr            error
	updateErr            error
	deleteErr            error
	incrementErr         error
	decrementErr         error
	recalculateErr       error
	updateMaturityErr    error
}

type targetMaturityUpdate struct {
	ID             string
	TargetMaturity int
}

func (m *mockEnterpriseCapabilityReadModel) Insert(ctx context.Context, dto readmodels.EnterpriseCapabilityDTO) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.insertedDTOs = append(m.insertedDTOs, dto)
	return nil
}

func (m *mockEnterpriseCapabilityReadModel) Update(ctx context.Context, params readmodels.UpdateCapabilityParams) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedParams = append(m.updatedParams, params)
	return nil
}

func (m *mockEnterpriseCapabilityReadModel) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedIDs = append(m.deletedIDs, id)
	return nil
}

func (m *mockEnterpriseCapabilityReadModel) IncrementLinkCount(ctx context.Context, id string) error {
	if m.incrementErr != nil {
		return m.incrementErr
	}
	m.incrementedLinkIDs = append(m.incrementedLinkIDs, id)
	return nil
}

func (m *mockEnterpriseCapabilityReadModel) DecrementLinkCount(ctx context.Context, id string) error {
	if m.decrementErr != nil {
		return m.decrementErr
	}
	m.decrementedLinkIDs = append(m.decrementedLinkIDs, id)
	return nil
}

func (m *mockEnterpriseCapabilityReadModel) RecalculateDomainCount(ctx context.Context, id string) error {
	if m.recalculateErr != nil {
		return m.recalculateErr
	}
	m.recalculatedDomainIDs = append(m.recalculatedDomainIDs, id)
	return nil
}

func (m *mockEnterpriseCapabilityReadModel) UpdateTargetMaturity(ctx context.Context, id string, targetMaturity int) error {
	if m.updateMaturityErr != nil {
		return m.updateMaturityErr
	}
	m.updatedTargetMaturities = append(m.updatedTargetMaturities, targetMaturityUpdate{ID: id, TargetMaturity: targetMaturity})
	return nil
}

func newProjectorWithMock(mock *mockEnterpriseCapabilityReadModel) *EnterpriseCapabilityProjector {
	return NewEnterpriseCapabilityProjector(mock)
}

type projectableEvent interface {
	EventType() string
	EventData() map[string]interface{}
}

type eventProjector interface {
	ProjectEvent(ctx context.Context, eventType string, eventData []byte) error
}

func projectEvent(t *testing.T, projector eventProjector, event projectableEvent) error {
	t.Helper()
	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)
	return projector.ProjectEvent(context.Background(), event.EventType(), eventData)
}

func TestEnterpriseCapabilityProjector_CreateUpdate_WritesReadModel(t *testing.T) {
	id := uuid.New().String()

	tests := []struct {
		name   string
		event  projectableEvent
		verify func(t *testing.T, mock *mockEnterpriseCapabilityReadModel)
	}{
		{
			name:  "created inserts active dto with all fields",
			event: events.NewEnterpriseCapabilityCreated(id, "Test Capability", "Test Description", "Test Category"),
			verify: func(t *testing.T, mock *mockEnterpriseCapabilityReadModel) {
				require.Len(t, mock.insertedDTOs, 1)
				dto := mock.insertedDTOs[0]
				assert.Equal(t, id, dto.ID)
				assert.Equal(t, "Test Capability", dto.Name)
				assert.Equal(t, "Test Description", dto.Description)
				assert.Equal(t, "Test Category", dto.Category)
				assert.True(t, dto.Active)
			},
		},
		{
			name:  "updated applies new field values",
			event: events.NewEnterpriseCapabilityUpdated(id, "Updated Name", "Updated Description", "Updated Category"),
			verify: func(t *testing.T, mock *mockEnterpriseCapabilityReadModel) {
				require.Len(t, mock.updatedParams, 1)
				params := mock.updatedParams[0]
				assert.Equal(t, id, params.ID)
				assert.Equal(t, "Updated Name", params.Name)
				assert.Equal(t, "Updated Description", params.Description)
				assert.Equal(t, "Updated Category", params.Category)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockEnterpriseCapabilityReadModel{}
			projector := newProjectorWithMock(mock)

			require.NoError(t, projectEvent(t, projector, tt.event))
			tt.verify(t, mock)
		})
	}
}

func TestEnterpriseCapabilityProjector_Created_PreservesTimestamp(t *testing.T) {
	mock := &mockEnterpriseCapabilityReadModel{}
	projector := newProjectorWithMock(mock)

	event := events.NewEnterpriseCapabilityCreated(
		uuid.New().String(),
		"Test Capability",
		"Test Description",
		"Test Category",
	)
	event.CreatedAt = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	require.NoError(t, projectEvent(t, projector, event))

	require.Len(t, mock.insertedDTOs, 1)
	assert.Equal(t, event.CreatedAt, mock.insertedDTOs[0].CreatedAt)
}

func TestEnterpriseCapabilityProjector_Deleted_MarksInactive(t *testing.T) {
	mock := &mockEnterpriseCapabilityReadModel{}
	projector := newProjectorWithMock(mock)

	event := events.NewEnterpriseCapabilityDeleted(uuid.New().String())

	require.NoError(t, projectEvent(t, projector, event))

	require.Len(t, mock.deletedIDs, 1)
	assert.Equal(t, event.ID, mock.deletedIDs[0])
}

func TestEnterpriseCapabilityProjector_LinkUnlink_UpdatesCountAndRecalculates(t *testing.T) {
	enterpriseCapabilityID := uuid.New().String()

	tests := []struct {
		name         string
		event        projectableEvent
		wantCountIDs func(*mockEnterpriseCapabilityReadModel) []string
	}{
		{
			name:         "linked increments link count",
			event:        events.NewEnterpriseCapabilityLinked(uuid.New().String(), enterpriseCapabilityID, uuid.New().String(), "user@example.com"),
			wantCountIDs: func(m *mockEnterpriseCapabilityReadModel) []string { return m.incrementedLinkIDs },
		},
		{
			name:         "unlinked decrements link count",
			event:        events.NewEnterpriseCapabilityUnlinked(uuid.New().String(), enterpriseCapabilityID, uuid.New().String()),
			wantCountIDs: func(m *mockEnterpriseCapabilityReadModel) []string { return m.decrementedLinkIDs },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockEnterpriseCapabilityReadModel{}
			projector := newProjectorWithMock(mock)

			require.NoError(t, projectEvent(t, projector, tt.event))

			require.Equal(t, []string{enterpriseCapabilityID}, tt.wantCountIDs(mock))
			require.Equal(t, []string{enterpriseCapabilityID}, mock.recalculatedDomainIDs)
		})
	}
}

func TestEnterpriseCapabilityProjector_LinkUnlink_CountUpdateError_SkipsRecalculate(t *testing.T) {
	tests := []struct {
		name      string
		event     projectableEvent
		mockSetup func(*mockEnterpriseCapabilityReadModel)
	}{
		{
			name:      "increment failure aborts before recalculate",
			event:     events.NewEnterpriseCapabilityLinked(uuid.New().String(), uuid.New().String(), uuid.New().String(), "user@example.com"),
			mockSetup: func(m *mockEnterpriseCapabilityReadModel) { m.incrementErr = errors.New("increment failed") },
		},
		{
			name:      "decrement failure aborts before recalculate",
			event:     events.NewEnterpriseCapabilityUnlinked(uuid.New().String(), uuid.New().String(), uuid.New().String()),
			mockSetup: func(m *mockEnterpriseCapabilityReadModel) { m.decrementErr = errors.New("decrement failed") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockEnterpriseCapabilityReadModel{}
			tt.mockSetup(mock)
			projector := newProjectorWithMock(mock)

			require.Error(t, projectEvent(t, projector, tt.event))
			assert.Empty(t, mock.recalculatedDomainIDs, "recalculate must not run when count update fails")
		})
	}
}

func TestEnterpriseCapabilityProjector_Linked_RecalculateError_PropagatesError(t *testing.T) {
	mock := &mockEnterpriseCapabilityReadModel{recalculateErr: errors.New("recalculate failed")}
	projector := newProjectorWithMock(mock)

	enterpriseCapabilityID := uuid.New().String()
	event := events.NewEnterpriseCapabilityLinked(
		uuid.New().String(),
		enterpriseCapabilityID,
		uuid.New().String(),
		"user@example.com",
	)

	require.Error(t, projectEvent(t, projector, event))
	assert.Equal(t, []string{enterpriseCapabilityID}, mock.incrementedLinkIDs)
}

func TestEnterpriseCapabilityProjector_TargetMaturitySet_UpdatesTargetMaturity(t *testing.T) {
	mock := &mockEnterpriseCapabilityReadModel{}
	projector := newProjectorWithMock(mock)

	id := uuid.New().String()
	event := events.NewEnterpriseCapabilityTargetMaturitySet(id, 4)

	require.NoError(t, projectEvent(t, projector, event))

	require.Len(t, mock.updatedTargetMaturities, 1)
	assert.Equal(t, targetMaturityUpdate{ID: id, TargetMaturity: 4}, mock.updatedTargetMaturities[0])
}

func TestEnterpriseCapabilityProjector_UnknownEvent_Ignored(t *testing.T) {
	mock := &mockEnterpriseCapabilityReadModel{}
	projector := newProjectorWithMock(mock)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mock.insertedDTOs)
	assert.Empty(t, mock.updatedParams)
	assert.Empty(t, mock.deletedIDs)
	assert.Empty(t, mock.incrementedLinkIDs)
	assert.Empty(t, mock.decrementedLinkIDs)
	assert.Empty(t, mock.recalculatedDomainIDs)
	assert.Empty(t, mock.updatedTargetMaturities)
}

func TestEnterpriseCapabilityProjector_ReadModelError_ReturnsError(t *testing.T) {
	mock := &mockEnterpriseCapabilityReadModel{insertErr: errors.New("database error")}
	projector := newProjectorWithMock(mock)

	event := events.NewEnterpriseCapabilityCreated(
		uuid.New().String(),
		"Test Capability",
		"Test Description",
		"Test Category",
	)

	assert.Error(t, projectEvent(t, projector, event))
}

func TestEnterpriseCapabilityProjector_InvalidJSON_ReturnsError(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
	}{
		{"created", "EnterpriseCapabilityCreated"},
		{"updated", "EnterpriseCapabilityUpdated"},
		{"deleted", "EnterpriseCapabilityDeleted"},
		{"linked", "EnterpriseCapabilityLinked"},
		{"unlinked", "EnterpriseCapabilityUnlinked"},
		{"target maturity set", "EnterpriseCapabilityTargetMaturitySet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockEnterpriseCapabilityReadModel{}
			projector := newProjectorWithMock(mock)

			err := projector.ProjectEvent(context.Background(), tt.eventType, []byte("invalid json"))
			assert.Error(t, err)
		})
	}
}
