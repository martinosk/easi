package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockFitScoreCacheReadModel struct {
	upsertedEntries []readmodels.FitScoreEntry
	deletedKeys     []fitScoreDeleteKey
	upsertErr       error
	deleteErr       error
}

type fitScoreDeleteKey struct {
	ComponentID string
	PillarID    string
}

func (m *mockFitScoreCacheReadModel) Upsert(ctx context.Context, entry readmodels.FitScoreEntry) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.upsertedEntries = append(m.upsertedEntries, entry)
	return nil
}

func (m *mockFitScoreCacheReadModel) Delete(ctx context.Context, componentID, pillarID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedKeys = append(m.deletedKeys, fitScoreDeleteKey{ComponentID: componentID, PillarID: pillarID})
	return nil
}

type fitScoreCacheWriter interface {
	Upsert(ctx context.Context, entry readmodels.FitScoreEntry) error
	Delete(ctx context.Context, componentID, pillarID string) error
}

type testableFitScoreCacheProjector struct {
	readModel fitScoreCacheWriter
}

func newTestableFitScoreCacheProjector(rm fitScoreCacheWriter) *testableFitScoreCacheProjector {
	return &testableFitScoreCacheProjector{readModel: rm}
}

func (p *testableFitScoreCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"ApplicationFitScoreSet":     p.handleApplicationFitScoreSet,
		"ApplicationFitScoreRemoved": p.handleApplicationFitScoreRemoved,
	}
	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *testableFitScoreCacheProjector) handleApplicationFitScoreSet(ctx context.Context, eventData []byte) error {
	var event applicationFitScoreSetEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.Upsert(ctx, readmodels.FitScoreEntry{
		ComponentID: event.ComponentID,
		PillarID:    event.PillarID,
		Score:       event.Score,
		Rationale:   event.Rationale,
	})
}

func (p *testableFitScoreCacheProjector) handleApplicationFitScoreRemoved(ctx context.Context, eventData []byte) error {
	var event applicationFitScoreRemovedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.Delete(ctx, event.ComponentID, event.PillarID)
}

func TestFitScoreCache_FitScoreSet_UpsertsEntry(t *testing.T) {
	mock := &mockFitScoreCacheReadModel{}
	projector := newTestableFitScoreCacheProjector(mock)

	componentID := uuid.New().String()
	pillarID := uuid.New().String()
	eventData, err := json.Marshal(applicationFitScoreSetEvent{
		ComponentID: componentID,
		PillarID:    pillarID,
		Score:       75,
		Rationale:   "Good fit",
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ApplicationFitScoreSet", eventData)
	require.NoError(t, err)

	require.Len(t, mock.upsertedEntries, 1)
	entry := mock.upsertedEntries[0]
	assert.Equal(t, componentID, entry.ComponentID)
	assert.Equal(t, pillarID, entry.PillarID)
	assert.Equal(t, 75, entry.Score)
	assert.Equal(t, "Good fit", entry.Rationale)
}

func TestFitScoreCache_FitScoreRemoved_DeletesEntry(t *testing.T) {
	mock := &mockFitScoreCacheReadModel{}
	projector := newTestableFitScoreCacheProjector(mock)

	componentID := uuid.New().String()
	pillarID := uuid.New().String()
	eventData, err := json.Marshal(applicationFitScoreRemovedEvent{
		ComponentID: componentID,
		PillarID:    pillarID,
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ApplicationFitScoreRemoved", eventData)
	require.NoError(t, err)

	require.Len(t, mock.deletedKeys, 1)
	assert.Equal(t, componentID, mock.deletedKeys[0].ComponentID)
	assert.Equal(t, pillarID, mock.deletedKeys[0].PillarID)
}

func TestFitScoreCache_UnknownEvent_Ignored(t *testing.T) {
	mock := &mockFitScoreCacheReadModel{}
	projector := newTestableFitScoreCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mock.upsertedEntries)
	assert.Empty(t, mock.deletedKeys)
}

func TestFitScoreCache_InvalidJSON_ReturnsError(t *testing.T) {
	mock := &mockFitScoreCacheReadModel{}
	projector := newTestableFitScoreCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), "ApplicationFitScoreSet", []byte("invalid"))
	assert.Error(t, err)
}

func TestFitScoreCache_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockFitScoreCacheReadModel
		eventType string
		eventData any
	}{
		{
			"upsert error",
			&mockFitScoreCacheReadModel{upsertErr: errors.New("db error")},
			"ApplicationFitScoreSet",
			applicationFitScoreSetEvent{ComponentID: uuid.New().String(), PillarID: uuid.New().String(), Score: 50},
		},
		{
			"delete error",
			&mockFitScoreCacheReadModel{deleteErr: errors.New("db error")},
			"ApplicationFitScoreRemoved",
			applicationFitScoreRemovedEvent{ComponentID: uuid.New().String(), PillarID: uuid.New().String()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projector := newTestableFitScoreCacheProjector(tt.mock)
			eventData, _ := json.Marshal(tt.eventData)
			err := projector.ProjectEvent(context.Background(), tt.eventType, eventData)
			assert.Error(t, err)
		})
	}
}
