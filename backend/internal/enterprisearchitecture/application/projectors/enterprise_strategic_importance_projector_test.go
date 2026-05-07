package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/events"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnterpriseStrategicImportanceReadModel struct {
	insertedDTOs []readmodels.EnterpriseStrategicImportanceDTO
	updates      []struct {
		ID         string
		Importance int
		Rationale  string
	}
	deletedIDs []string
	insertErr  error
	updateErr  error
	deleteErr  error
}

func (m *mockEnterpriseStrategicImportanceReadModel) Insert(ctx context.Context, dto readmodels.EnterpriseStrategicImportanceDTO) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.insertedDTOs = append(m.insertedDTOs, dto)
	return nil
}

func (m *mockEnterpriseStrategicImportanceReadModel) Update(ctx context.Context, id string, importance int, rationale string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updates = append(m.updates, struct {
		ID         string
		Importance int
		Rationale  string
	}{id, importance, rationale})
	return nil
}

func (m *mockEnterpriseStrategicImportanceReadModel) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedIDs = append(m.deletedIDs, id)
	return nil
}

func TestEnterpriseStrategicImportanceProjector_Set_InsertsRating(t *testing.T) {
	mockReadModel := &mockEnterpriseStrategicImportanceReadModel{}
	projector := NewEnterpriseStrategicImportanceProjector(mockReadModel)

	enterpriseCapabilityID := uuid.New().String()
	pillarID := uuid.New().String()
	event := events.NewEnterpriseStrategicImportanceSet(events.EnterpriseStrategicImportanceSetParams{
		ID:                     uuid.New().String(),
		EnterpriseCapabilityID: enterpriseCapabilityID,
		PillarID:               pillarID,
		PillarName:             "Strategic Pillar 1",
		Importance:             4,
		Rationale:              "Critical for operations",
	})

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseStrategicImportanceSet", eventData)
	require.NoError(t, err)

	require.Len(t, mockReadModel.insertedDTOs, 1)
	dto := mockReadModel.insertedDTOs[0]
	assert.Equal(t, event.ID, dto.ID)
	assert.Equal(t, enterpriseCapabilityID, dto.EnterpriseCapabilityID)
	assert.Equal(t, pillarID, dto.PillarID)
	assert.Equal(t, "Strategic Pillar 1", dto.PillarName)
	assert.Equal(t, 4, dto.Importance)
	assert.Equal(t, "Critical for operations", dto.Rationale)
	assert.Equal(t, event.SetAt, dto.SetAt)
}

func TestEnterpriseStrategicImportanceProjector_Updated_UpdatesRating(t *testing.T) {
	mockReadModel := &mockEnterpriseStrategicImportanceReadModel{}
	projector := NewEnterpriseStrategicImportanceProjector(mockReadModel)

	ratingID := uuid.New().String()
	event := events.NewEnterpriseStrategicImportanceUpdated(events.EnterpriseStrategicImportanceUpdatedParams{
		ID:            ratingID,
		Importance:    5,
		Rationale:     "Updated rationale",
		OldImportance: 3,
		OldRationale:  "Old rationale",
	})

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseStrategicImportanceUpdated", eventData)
	require.NoError(t, err)

	require.Len(t, mockReadModel.updates, 1)
	update := mockReadModel.updates[0]
	assert.Equal(t, ratingID, update.ID)
	assert.Equal(t, 5, update.Importance)
	assert.Equal(t, "Updated rationale", update.Rationale)
}

func TestEnterpriseStrategicImportanceProjector_Removed_DeletesRating(t *testing.T) {
	mockReadModel := &mockEnterpriseStrategicImportanceReadModel{}
	projector := NewEnterpriseStrategicImportanceProjector(mockReadModel)

	ratingID := uuid.New().String()
	event := events.NewEnterpriseStrategicImportanceRemoved(
		ratingID,
		uuid.New().String(),
		uuid.New().String(),
	)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseStrategicImportanceRemoved", eventData)
	require.NoError(t, err)

	require.Len(t, mockReadModel.deletedIDs, 1)
	assert.Equal(t, ratingID, mockReadModel.deletedIDs[0])
}

func TestEnterpriseStrategicImportanceProjector_UnknownEvent_Ignored(t *testing.T) {
	mockReadModel := &mockEnterpriseStrategicImportanceReadModel{}
	projector := NewEnterpriseStrategicImportanceProjector(mockReadModel)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mockReadModel.insertedDTOs)
	assert.Empty(t, mockReadModel.updates)
	assert.Empty(t, mockReadModel.deletedIDs)
}

func TestEnterpriseStrategicImportanceProjector_StoreErrorPropagation(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockEnterpriseStrategicImportanceReadModel
		eventType string
		event     projectableEvent
	}{
		{
			name:      "insert error during set",
			mock:      &mockEnterpriseStrategicImportanceReadModel{insertErr: errors.New("database error")},
			eventType: "EnterpriseStrategicImportanceSet",
			event: events.NewEnterpriseStrategicImportanceSet(events.EnterpriseStrategicImportanceSetParams{
				ID: uuid.New().String(), EnterpriseCapabilityID: uuid.New().String(), PillarID: uuid.New().String(),
				PillarName: "Test", Importance: 3,
			}),
		},
		{
			name:      "update error during update",
			mock:      &mockEnterpriseStrategicImportanceReadModel{updateErr: errors.New("database error")},
			eventType: "EnterpriseStrategicImportanceUpdated",
			event: events.NewEnterpriseStrategicImportanceUpdated(events.EnterpriseStrategicImportanceUpdatedParams{
				ID: uuid.New().String(), Importance: 4, Rationale: "Test", OldImportance: 3, OldRationale: "Old",
			}),
		},
		{
			name:      "delete error during remove",
			mock:      &mockEnterpriseStrategicImportanceReadModel{deleteErr: errors.New("database error")},
			eventType: "EnterpriseStrategicImportanceRemoved",
			event:     events.NewEnterpriseStrategicImportanceRemoved(uuid.New().String(), uuid.New().String(), uuid.New().String()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projector := NewEnterpriseStrategicImportanceProjector(tt.mock)
			eventData, err := json.Marshal(tt.event.EventData())
			require.NoError(t, err)
			assert.Error(t, projector.ProjectEvent(context.Background(), tt.eventType, eventData))
		})
	}
}

func TestEnterpriseStrategicImportanceProjector_InvalidJSON_ReturnsError(t *testing.T) {
	mockReadModel := &mockEnterpriseStrategicImportanceReadModel{}
	projector := NewEnterpriseStrategicImportanceProjector(mockReadModel)

	err := projector.ProjectEvent(context.Background(), "EnterpriseStrategicImportanceSet", []byte("invalid json"))
	assert.Error(t, err)
}
