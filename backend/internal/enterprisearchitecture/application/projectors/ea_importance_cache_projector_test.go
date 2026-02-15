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

type mockImportanceCacheReadModel struct {
	upsertedEntries []readmodels.ImportanceEntry
	upsertErr       error
}

func (m *mockImportanceCacheReadModel) Upsert(ctx context.Context, entry readmodels.ImportanceEntry) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.upsertedEntries = append(m.upsertedEntries, entry)
	return nil
}

type importanceCacheWriter interface {
	Upsert(ctx context.Context, entry readmodels.ImportanceEntry) error
}

type testableImportanceCacheProjector struct {
	readModel importanceCacheWriter
}

func newTestableImportanceCacheProjector(rm importanceCacheWriter) *testableImportanceCacheProjector {
	return &testableImportanceCacheProjector{readModel: rm}
}

func (p *testableImportanceCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"EffectiveImportanceRecalculated": p.handleEffectiveImportanceRecalculated,
	}
	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *testableImportanceCacheProjector) handleEffectiveImportanceRecalculated(ctx context.Context, eventData []byte) error {
	var event effectiveImportanceRecalculatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.Upsert(ctx, readmodels.ImportanceEntry{
		CapabilityID:        event.CapabilityID,
		BusinessDomainID:    event.BusinessDomainID,
		PillarID:            event.PillarID,
		EffectiveImportance: event.Importance,
	})
}

func TestImportanceCache_Recalculated_UpsertsEntry(t *testing.T) {
	mock := &mockImportanceCacheReadModel{}
	projector := newTestableImportanceCacheProjector(mock)

	capabilityID := uuid.New().String()
	domainID := uuid.New().String()
	pillarID := uuid.New().String()
	eventData, err := json.Marshal(effectiveImportanceRecalculatedEvent{
		CapabilityID:     capabilityID,
		BusinessDomainID: domainID,
		PillarID:         pillarID,
		Importance:       85,
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "EffectiveImportanceRecalculated", eventData)
	require.NoError(t, err)

	require.Len(t, mock.upsertedEntries, 1)
	entry := mock.upsertedEntries[0]
	assert.Equal(t, capabilityID, entry.CapabilityID)
	assert.Equal(t, domainID, entry.BusinessDomainID)
	assert.Equal(t, pillarID, entry.PillarID)
	assert.Equal(t, 85, entry.EffectiveImportance)
}

func TestImportanceCache_UnknownEvent_Ignored(t *testing.T) {
	mock := &mockImportanceCacheReadModel{}
	projector := newTestableImportanceCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mock.upsertedEntries)
}

func TestImportanceCache_InvalidJSON_ReturnsError(t *testing.T) {
	mock := &mockImportanceCacheReadModel{}
	projector := newTestableImportanceCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), "EffectiveImportanceRecalculated", []byte("invalid"))
	assert.Error(t, err)
}

func TestImportanceCache_ReadModelError_ReturnsError(t *testing.T) {
	mock := &mockImportanceCacheReadModel{upsertErr: errors.New("db error")}
	projector := newTestableImportanceCacheProjector(mock)

	eventData, _ := json.Marshal(effectiveImportanceRecalculatedEvent{
		CapabilityID:     uuid.New().String(),
		BusinessDomainID: uuid.New().String(),
		PillarID:         uuid.New().String(),
		Importance:       50,
	})

	err := projector.ProjectEvent(context.Background(), "EffectiveImportanceRecalculated", eventData)
	assert.Error(t, err)
}
