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

type enterpriseStrategicImportanceReadModelForProjector interface {
	Insert(ctx context.Context, dto readmodels.EnterpriseStrategicImportanceDTO) error
	Update(ctx context.Context, id string, importance int, rationale string) error
	Delete(ctx context.Context, id string) error
}

type testableEnterpriseStrategicImportanceProjector struct {
	readModel enterpriseStrategicImportanceReadModelForProjector
}

func newTestableEnterpriseStrategicImportanceProjector(readModel enterpriseStrategicImportanceReadModelForProjector) *testableEnterpriseStrategicImportanceProjector {
	return &testableEnterpriseStrategicImportanceProjector{readModel: readModel}
}

func (p *testableEnterpriseStrategicImportanceProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"EnterpriseStrategicImportanceSet":     p.handleSet,
		"EnterpriseStrategicImportanceUpdated": p.handleUpdated,
		"EnterpriseStrategicImportanceRemoved": p.handleRemoved,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *testableEnterpriseStrategicImportanceProjector) handleSet(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseStrategicImportanceSet
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	dto := readmodels.EnterpriseStrategicImportanceDTO{
		ID:                     event.ID,
		EnterpriseCapabilityID: event.EnterpriseCapabilityID,
		PillarID:               event.PillarID,
		PillarName:             event.PillarName,
		Importance:             event.Importance,
		Rationale:              event.Rationale,
		SetAt:                  event.SetAt,
	}
	return p.readModel.Insert(ctx, dto)
}

func (p *testableEnterpriseStrategicImportanceProjector) handleUpdated(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseStrategicImportanceUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.Update(ctx, event.ID, event.Importance, event.Rationale)
}

func (p *testableEnterpriseStrategicImportanceProjector) handleRemoved(ctx context.Context, eventData []byte) error {
	var event events.EnterpriseStrategicImportanceRemoved
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}

func TestEnterpriseStrategicImportanceProjector_Set_InsertsRating(t *testing.T) {
	mockReadModel := &mockEnterpriseStrategicImportanceReadModel{}
	projector := newTestableEnterpriseStrategicImportanceProjector(mockReadModel)

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
	projector := newTestableEnterpriseStrategicImportanceProjector(mockReadModel)

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
	projector := newTestableEnterpriseStrategicImportanceProjector(mockReadModel)

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
	projector := newTestableEnterpriseStrategicImportanceProjector(mockReadModel)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mockReadModel.insertedDTOs)
	assert.Empty(t, mockReadModel.updates)
	assert.Empty(t, mockReadModel.deletedIDs)
}

func TestEnterpriseStrategicImportanceProjector_InsertError_ReturnsError(t *testing.T) {
	mockReadModel := &mockEnterpriseStrategicImportanceReadModel{
		insertErr: errors.New("database error"),
	}
	projector := newTestableEnterpriseStrategicImportanceProjector(mockReadModel)

	event := events.NewEnterpriseStrategicImportanceSet(events.EnterpriseStrategicImportanceSetParams{
		ID:                     uuid.New().String(),
		EnterpriseCapabilityID: uuid.New().String(),
		PillarID:               uuid.New().String(),
		PillarName:             "Test",
		Importance:             3,
		Rationale:              "",
	})

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseStrategicImportanceSet", eventData)
	assert.Error(t, err)
}

func TestEnterpriseStrategicImportanceProjector_UpdateError_ReturnsError(t *testing.T) {
	mockReadModel := &mockEnterpriseStrategicImportanceReadModel{
		updateErr: errors.New("database error"),
	}
	projector := newTestableEnterpriseStrategicImportanceProjector(mockReadModel)

	event := events.NewEnterpriseStrategicImportanceUpdated(events.EnterpriseStrategicImportanceUpdatedParams{
		ID:            uuid.New().String(),
		Importance:    4,
		Rationale:     "Test",
		OldImportance: 3,
		OldRationale:  "Old",
	})

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseStrategicImportanceUpdated", eventData)
	assert.Error(t, err)
}

func TestEnterpriseStrategicImportanceProjector_DeleteError_ReturnsError(t *testing.T) {
	mockReadModel := &mockEnterpriseStrategicImportanceReadModel{
		deleteErr: errors.New("database error"),
	}
	projector := newTestableEnterpriseStrategicImportanceProjector(mockReadModel)

	event := events.NewEnterpriseStrategicImportanceRemoved(
		uuid.New().String(),
		uuid.New().String(),
		uuid.New().String(),
	)

	eventData, err := json.Marshal(event.EventData())
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EnterpriseStrategicImportanceRemoved", eventData)
	assert.Error(t, err)
}

func TestEnterpriseStrategicImportanceProjector_InvalidJSON_ReturnsError(t *testing.T) {
	mockReadModel := &mockEnterpriseStrategicImportanceReadModel{}
	projector := newTestableEnterpriseStrategicImportanceProjector(mockReadModel)

	err := projector.ProjectEvent(context.Background(), "EnterpriseStrategicImportanceSet", []byte("invalid json"))
	assert.Error(t, err)
}
