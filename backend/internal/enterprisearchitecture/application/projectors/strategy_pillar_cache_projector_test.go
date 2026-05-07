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

func TestStrategyPillarCacheProjector_ConfigurationCreated_InsertsAllPillars(t *testing.T) {
	mock := &mockStrategyPillarCacheReadModel{}
	projector := NewStrategyPillarCacheProjector(mock)

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
	projector := NewStrategyPillarCacheProjector(mock)

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

func TestStrategyPillarCacheProjector_PartialUpdates_PreserveUntouchedFields(t *testing.T) {
	existing := &readmodels.StrategyPillarCacheDTO{
		ID:                "pillar-1",
		TenantID:          "tenant-1",
		Name:              "Original Name",
		Description:       "Original Description",
		Active:            true,
		FitScoringEnabled: true,
		FitCriteria:       "original-criteria",
		FitType:           "original-type",
	}

	tests := []struct {
		name      string
		eventType string
		event     pillarEvent
		want      readmodels.StrategyPillarCacheDTO
	}{
		{
			name:      "StrategyPillarUpdated changes only name and description",
			eventType: "StrategyPillarUpdated",
			event: pillarEvent{
				ID:                "aggregate-1",
				TenantID:          "tenant-1",
				PillarID:          "pillar-1",
				NewName:           "New Name",
				NewDescription:    "New Description",
				FitScoringEnabled: false,
				FitCriteria:       "should-be-ignored",
				FitType:           "should-be-ignored",
			},
			want: readmodels.StrategyPillarCacheDTO{
				ID:                "pillar-1",
				TenantID:          "tenant-1",
				Name:              "New Name",
				Description:       "New Description",
				Active:            true,
				FitScoringEnabled: true,
				FitCriteria:       "original-criteria",
				FitType:           "original-type",
			},
		},
		{
			name:      "PillarFitConfigurationUpdated changes only fit fields",
			eventType: "PillarFitConfigurationUpdated",
			event: pillarEvent{
				ID:                "aggregate-1",
				TenantID:          "tenant-1",
				PillarID:          "pillar-1",
				NewName:           "should-be-ignored",
				NewDescription:    "should-be-ignored",
				FitScoringEnabled: true,
				FitCriteria:       "new-criteria",
				FitType:           "binary",
			},
			want: readmodels.StrategyPillarCacheDTO{
				ID:                "pillar-1",
				TenantID:          "tenant-1",
				Name:              "Original Name",
				Description:       "Original Description",
				Active:            true,
				FitScoringEnabled: true,
				FitCriteria:       "new-criteria",
				FitType:           "binary",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := *existing
			mock := &mockStrategyPillarCacheReadModel{activePillar: &seed}
			projector := NewStrategyPillarCacheProjector(mock)

			eventData, err := json.Marshal(tt.event)
			require.NoError(t, err)

			require.NoError(t, projector.ProjectEvent(context.Background(), tt.eventType, eventData))

			require.Len(t, mock.insertedDTOs, 1)
			assert.Equal(t, tt.want, mock.insertedDTOs[0])
		})
	}
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
			projector := NewStrategyPillarCacheProjector(mock)
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
	projector := NewStrategyPillarCacheProjector(mock)

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
	projector := NewStrategyPillarCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mock.insertedDTOs)
	assert.Empty(t, mock.deletedIDs)
}

func TestStrategyPillarCacheProjector_InvalidJSON_ReturnsError(t *testing.T) {
	mock := &mockStrategyPillarCacheReadModel{}
	projector := NewStrategyPillarCacheProjector(mock)

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
			projector := NewStrategyPillarCacheProjector(tt.mock)
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
	projector := NewStrategyPillarCacheProjector(mock)

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
