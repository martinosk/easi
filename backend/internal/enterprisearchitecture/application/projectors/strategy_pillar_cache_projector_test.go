package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStrategyPillarCacheReadModel struct {
	insertedDTOs []readmodels.StrategyPillarCacheDTO
	deletedIDs   []string
	activePillar *readmodels.StrategyPillarCacheDTO
	insertErr    error
	deleteErr    error
	getErr       error
}

func (m *mockStrategyPillarCacheReadModel) Insert(ctx context.Context, dto readmodels.StrategyPillarCacheDTO) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.insertedDTOs = append(m.insertedDTOs, dto)
	return nil
}

func (m *mockStrategyPillarCacheReadModel) Delete(ctx context.Context, pillarID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedIDs = append(m.deletedIDs, pillarID)
	return nil
}

func (m *mockStrategyPillarCacheReadModel) GetActivePillar(ctx context.Context, pillarID string) (*readmodels.StrategyPillarCacheDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.activePillar, nil
}

type strategyPillarCacheReadModelForProjector interface {
	Insert(ctx context.Context, dto readmodels.StrategyPillarCacheDTO) error
	Delete(ctx context.Context, pillarID string) error
	GetActivePillar(ctx context.Context, pillarID string) (*readmodels.StrategyPillarCacheDTO, error)
}

type testableStrategyPillarCacheProjector struct {
	readModel strategyPillarCacheReadModelForProjector
}

func newTestableStrategyPillarCacheProjector(readModel strategyPillarCacheReadModelForProjector) *testableStrategyPillarCacheProjector {
	return &testableStrategyPillarCacheProjector{readModel: readModel}
}

func (p *testableStrategyPillarCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"MetaModelConfigurationCreated": p.handleConfigurationCreated,
		"StrategyPillarAdded":           p.handlePillarAdded,
		"StrategyPillarUpdated":         p.handlePillarUpdated,
		"StrategyPillarRemoved":         p.handlePillarRemoved,
		"PillarFitConfigurationUpdated": p.handleFitConfigurationUpdated,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *testableStrategyPillarCacheProjector) handleConfigurationCreated(ctx context.Context, eventData []byte) error {
	var event configurationCreatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	for _, pillar := range event.Pillars {
		dto := readmodels.StrategyPillarCacheDTO{
			ID:                pillar.ID,
			TenantID:          event.TenantID,
			Name:              pillar.Name,
			Description:       pillar.Description,
			Active:            pillar.Active,
			FitScoringEnabled: pillar.FitScoringEnabled,
			FitCriteria:       pillar.FitCriteria,
			FitType:           pillar.FitType,
		}
		if err := p.readModel.Insert(ctx, dto); err != nil {
			return err
		}
	}

	return nil
}

func (p *testableStrategyPillarCacheProjector) handlePillarAdded(ctx context.Context, eventData []byte) error {
	var event pillarAddedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	dto := readmodels.StrategyPillarCacheDTO{
		ID:                event.PillarID,
		TenantID:          event.TenantID,
		Name:              event.Name,
		Description:       event.Description,
		Active:            true,
		FitScoringEnabled: false,
		FitCriteria:       "",
		FitType:           "",
	}

	return p.readModel.Insert(ctx, dto)
}

func (p *testableStrategyPillarCacheProjector) handlePillarUpdated(ctx context.Context, eventData []byte) error {
	return p.unmarshalAndUpdate(ctx, eventData, func(event pillarEvent, existing *readmodels.StrategyPillarCacheDTO) {
		existing.Name = event.NewName
		existing.Description = event.NewDescription
	})
}

func (p *testableStrategyPillarCacheProjector) handlePillarRemoved(ctx context.Context, eventData []byte) error {
	var event pillarEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	return p.readModel.Delete(ctx, event.PillarID)
}

func (p *testableStrategyPillarCacheProjector) handleFitConfigurationUpdated(ctx context.Context, eventData []byte) error {
	return p.unmarshalAndUpdate(ctx, eventData, func(event pillarEvent, existing *readmodels.StrategyPillarCacheDTO) {
		existing.FitScoringEnabled = event.FitScoringEnabled
		existing.FitCriteria = event.FitCriteria
		existing.FitType = event.FitType
	})
}

func (p *testableStrategyPillarCacheProjector) unmarshalAndUpdate(ctx context.Context, eventData []byte, mutate func(pillarEvent, *readmodels.StrategyPillarCacheDTO)) error {
	var event pillarEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	existing, err := p.readModel.GetActivePillar(ctx, event.PillarID)
	if err != nil {
		return err
	}

	if existing == nil {
		existing = &readmodels.StrategyPillarCacheDTO{
			ID:       event.PillarID,
			TenantID: event.TenantID,
			Active:   true,
		}
	}

	mutate(event, existing)
	return p.readModel.Insert(ctx, *existing)
}

func TestStrategyPillarCacheProjector_ConfigurationCreated_InsertsAllPillars(t *testing.T) {
	mock := &mockStrategyPillarCacheReadModel{}
	projector := newTestableStrategyPillarCacheProjector(mock)

	eventData, err := json.Marshal(configurationCreatedEvent{
		TenantID: "tenant-1",
		Pillars: []configurationCreatedPillar{
			{
				ID:                "pillar-1",
				Name:              "Security",
				Description:       "Security pillar",
				Active:            true,
				FitScoringEnabled: true,
				FitCriteria:       "compliance",
				FitType:           "scale",
			},
			{
				ID:                "pillar-2",
				Name:              "Performance",
				Description:       "Performance pillar",
				Active:            false,
				FitScoringEnabled: false,
				FitCriteria:       "",
				FitType:           "",
			},
		},
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "MetaModelConfigurationCreated", eventData)
	require.NoError(t, err)

	require.Len(t, mock.insertedDTOs, 2)

	assert.Equal(t, "pillar-1", mock.insertedDTOs[0].ID)
	assert.Equal(t, "tenant-1", mock.insertedDTOs[0].TenantID)
	assert.Equal(t, "Security", mock.insertedDTOs[0].Name)
	assert.Equal(t, "Security pillar", mock.insertedDTOs[0].Description)
	assert.True(t, mock.insertedDTOs[0].Active)
	assert.True(t, mock.insertedDTOs[0].FitScoringEnabled)
	assert.Equal(t, "compliance", mock.insertedDTOs[0].FitCriteria)
	assert.Equal(t, "scale", mock.insertedDTOs[0].FitType)

	assert.Equal(t, "pillar-2", mock.insertedDTOs[1].ID)
	assert.Equal(t, "Performance", mock.insertedDTOs[1].Name)
	assert.False(t, mock.insertedDTOs[1].Active)
	assert.False(t, mock.insertedDTOs[1].FitScoringEnabled)
}

func TestStrategyPillarCacheProjector_PillarAdded_InsertsWithDefaults(t *testing.T) {
	mock := &mockStrategyPillarCacheReadModel{}
	projector := newTestableStrategyPillarCacheProjector(mock)

	eventData, err := json.Marshal(pillarAddedEvent{
		ID:          "aggregate-1",
		TenantID:    "tenant-1",
		PillarID:    "pillar-1",
		Name:        "Innovation",
		Description: "Innovation pillar",
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "StrategyPillarAdded", eventData)
	require.NoError(t, err)

	require.Len(t, mock.insertedDTOs, 1)
	dto := mock.insertedDTOs[0]
	assert.Equal(t, "pillar-1", dto.ID)
	assert.Equal(t, "tenant-1", dto.TenantID)
	assert.Equal(t, "Innovation", dto.Name)
	assert.Equal(t, "Innovation pillar", dto.Description)
	assert.True(t, dto.Active)
	assert.False(t, dto.FitScoringEnabled)
	assert.Empty(t, dto.FitCriteria)
	assert.Empty(t, dto.FitType)
}

func TestStrategyPillarCacheProjector_PillarUpdated_UpdatesOnlyNameAndDescription(t *testing.T) {
	mock := &mockStrategyPillarCacheReadModel{
		activePillar: &readmodels.StrategyPillarCacheDTO{
			ID:                "pillar-1",
			TenantID:          "tenant-1",
			Name:              "Old Name",
			Description:       "Old Description",
			Active:            true,
			FitScoringEnabled: true,
			FitCriteria:       "original-criteria",
			FitType:           "original-type",
		},
	}
	projector := newTestableStrategyPillarCacheProjector(mock)

	eventData, err := json.Marshal(pillarEvent{
		ID:                "aggregate-1",
		TenantID:          "tenant-1",
		PillarID:          "pillar-1",
		NewName:           "New Name",
		NewDescription:    "New Description",
		FitScoringEnabled: false,
		FitCriteria:       "should-be-ignored",
		FitType:           "should-be-ignored",
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "StrategyPillarUpdated", eventData)
	require.NoError(t, err)

	require.Len(t, mock.insertedDTOs, 1)
	dto := mock.insertedDTOs[0]
	assert.Equal(t, "New Name", dto.Name)
	assert.Equal(t, "New Description", dto.Description)
	assert.True(t, dto.FitScoringEnabled, "FitScoringEnabled must be preserved from existing pillar")
	assert.Equal(t, "original-criteria", dto.FitCriteria, "FitCriteria must be preserved from existing pillar")
	assert.Equal(t, "original-type", dto.FitType, "FitType must be preserved from existing pillar")
}

func TestStrategyPillarCacheProjector_FitConfigurationUpdated_UpdatesOnlyFitFields(t *testing.T) {
	mock := &mockStrategyPillarCacheReadModel{
		activePillar: &readmodels.StrategyPillarCacheDTO{
			ID:                "pillar-1",
			TenantID:          "tenant-1",
			Name:              "Original Name",
			Description:       "Original Description",
			Active:            true,
			FitScoringEnabled: false,
			FitCriteria:       "",
			FitType:           "",
		},
	}
	projector := newTestableStrategyPillarCacheProjector(mock)

	eventData, err := json.Marshal(pillarEvent{
		ID:                "aggregate-1",
		TenantID:          "tenant-1",
		PillarID:          "pillar-1",
		NewName:           "should-be-ignored",
		NewDescription:    "should-be-ignored",
		FitScoringEnabled: true,
		FitCriteria:       "new-criteria",
		FitType:           "binary",
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "PillarFitConfigurationUpdated", eventData)
	require.NoError(t, err)

	require.Len(t, mock.insertedDTOs, 1)
	dto := mock.insertedDTOs[0]
	assert.Equal(t, "Original Name", dto.Name, "Name must be preserved from existing pillar")
	assert.Equal(t, "Original Description", dto.Description, "Description must be preserved from existing pillar")
	assert.True(t, dto.FitScoringEnabled)
	assert.Equal(t, "new-criteria", dto.FitCriteria)
	assert.Equal(t, "binary", dto.FitType)
}

func TestStrategyPillarCacheProjector_CreatesNewDTOWhenPillarNotFound(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		event     pillarEvent
		verify    func(t *testing.T, dto readmodels.StrategyPillarCacheDTO)
	}{
		{
			name:      "pillar updated",
			eventType: "StrategyPillarUpdated",
			event:     pillarEvent{ID: "aggregate-1", TenantID: "tenant-1", PillarID: "pillar-1", NewName: "New Name", NewDescription: "New Description"},
			verify: func(t *testing.T, dto readmodels.StrategyPillarCacheDTO) {
				assert.Equal(t, "New Name", dto.Name)
				assert.Equal(t, "New Description", dto.Description)
				assert.False(t, dto.FitScoringEnabled)
				assert.Empty(t, dto.FitCriteria)
				assert.Empty(t, dto.FitType)
			},
		},
		{
			name:      "fit configuration updated",
			eventType: "PillarFitConfigurationUpdated",
			event:     pillarEvent{ID: "aggregate-1", TenantID: "tenant-1", PillarID: "pillar-1", FitScoringEnabled: true, FitCriteria: "criteria", FitType: "scale"},
			verify: func(t *testing.T, dto readmodels.StrategyPillarCacheDTO) {
				assert.Empty(t, dto.Name)
				assert.Empty(t, dto.Description)
				assert.True(t, dto.FitScoringEnabled)
				assert.Equal(t, "criteria", dto.FitCriteria)
				assert.Equal(t, "scale", dto.FitType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStrategyPillarCacheReadModel{activePillar: nil}
			projector := newTestableStrategyPillarCacheProjector(mock)
			eventData, err := json.Marshal(tt.event)
			require.NoError(t, err)
			err = projector.ProjectEvent(context.Background(), tt.eventType, eventData)
			require.NoError(t, err)
			require.Len(t, mock.insertedDTOs, 1)
			dto := mock.insertedDTOs[0]
			assert.Equal(t, "pillar-1", dto.ID)
			assert.Equal(t, "tenant-1", dto.TenantID)
			assert.True(t, dto.Active)
			tt.verify(t, dto)
		})
	}
}

func TestStrategyPillarCacheProjector_PillarRemoved_DeletesPillar(t *testing.T) {
	mock := &mockStrategyPillarCacheReadModel{}
	projector := newTestableStrategyPillarCacheProjector(mock)

	eventData, err := json.Marshal(pillarEvent{
		ID:       "aggregate-1",
		TenantID: "tenant-1",
		PillarID: "pillar-1",
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "StrategyPillarRemoved", eventData)
	require.NoError(t, err)

	require.Len(t, mock.deletedIDs, 1)
	assert.Equal(t, "pillar-1", mock.deletedIDs[0])
}

func TestStrategyPillarCacheProjector_UnknownEvent_Ignored(t *testing.T) {
	mock := &mockStrategyPillarCacheReadModel{}
	projector := newTestableStrategyPillarCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mock.insertedDTOs)
	assert.Empty(t, mock.deletedIDs)
}

func TestStrategyPillarCacheProjector_InvalidJSON_ReturnsError(t *testing.T) {
	mock := &mockStrategyPillarCacheReadModel{}
	projector := newTestableStrategyPillarCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), "StrategyPillarUpdated", []byte("invalid json"))
	assert.Error(t, err)
}

func TestStrategyPillarCacheProjector_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockStrategyPillarCacheReadModel
		eventType string
		eventData interface{}
	}{
		{
			name:      "get active pillar error",
			mock:      &mockStrategyPillarCacheReadModel{getErr: errors.New("database error")},
			eventType: "StrategyPillarUpdated",
			eventData: pillarEvent{ID: "aggregate-1", TenantID: "tenant-1", PillarID: "pillar-1", NewName: "Name"},
		},
		{
			name:      "insert error",
			mock:      &mockStrategyPillarCacheReadModel{insertErr: errors.New("database error")},
			eventType: "StrategyPillarAdded",
			eventData: pillarAddedEvent{ID: "aggregate-1", TenantID: "tenant-1", PillarID: "pillar-1", Name: "Test"},
		},
		{
			name:      "delete error",
			mock:      &mockStrategyPillarCacheReadModel{deleteErr: errors.New("database error")},
			eventType: "StrategyPillarRemoved",
			eventData: pillarEvent{ID: "aggregate-1", TenantID: "tenant-1", PillarID: "pillar-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projector := newTestableStrategyPillarCacheProjector(tt.mock)
			eventData, err := json.Marshal(tt.eventData)
			require.NoError(t, err)
			err = projector.ProjectEvent(context.Background(), tt.eventType, eventData)
			assert.Error(t, err)
		})
	}
}

func TestStrategyPillarCacheProjector_ConfigurationCreated_StopsOnInsertError(t *testing.T) {
	mock := &mockStrategyPillarCacheReadModel{
		insertErr: errors.New("database error"),
	}
	projector := newTestableStrategyPillarCacheProjector(mock)

	eventData, err := json.Marshal(configurationCreatedEvent{
		TenantID: "tenant-1",
		Pillars: []configurationCreatedPillar{
			{ID: "pillar-1", Name: "First"},
			{ID: "pillar-2", Name: "Second"},
		},
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "MetaModelConfigurationCreated", eventData)
	assert.Error(t, err)
}
