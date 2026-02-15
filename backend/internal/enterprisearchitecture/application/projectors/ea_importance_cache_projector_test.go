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

func TestImportanceCache_Recalculated_UpsertsEntry(t *testing.T) {
	mock := &mockImportanceCacheReadModel{}
	projector := NewEAImportanceCacheProjector(mock)

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

	err = projector.ProjectEvent(context.Background(), cmPL.EffectiveImportanceRecalculated, eventData)
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
	projector := NewEAImportanceCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mock.upsertedEntries)
}

func TestImportanceCache_InvalidJSON_ReturnsError(t *testing.T) {
	mock := &mockImportanceCacheReadModel{}
	projector := NewEAImportanceCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), cmPL.EffectiveImportanceRecalculated, []byte("invalid"))
	assert.Error(t, err)
}

func TestImportanceCache_ReadModelError_ReturnsError(t *testing.T) {
	mock := &mockImportanceCacheReadModel{upsertErr: errors.New("db error")}
	projector := NewEAImportanceCacheProjector(mock)

	eventData, _ := json.Marshal(effectiveImportanceRecalculatedEvent{
		CapabilityID:     uuid.New().String(),
		BusinessDomainID: uuid.New().String(),
		PillarID:         uuid.New().String(),
		Importance:       50,
	})

	err := projector.ProjectEvent(context.Background(), cmPL.EffectiveImportanceRecalculated, eventData)
	assert.Error(t, err)
}
