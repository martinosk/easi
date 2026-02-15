package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRealizationCacheReadModel struct {
	upsertedEntries      []readmodels.RealizationEntry
	deletedIDs           []string
	deletedCapabilityIDs []string
	updatedNames         []componentNameUpdate
	upsertErr            error
	deleteErr            error
	deleteByCapErr       error
	updateNameErr        error
}

type componentNameUpdate struct {
	ComponentID   string
	ComponentName string
}

func (m *mockRealizationCacheReadModel) Upsert(ctx context.Context, entry readmodels.RealizationEntry) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.upsertedEntries = append(m.upsertedEntries, entry)
	return nil
}

func (m *mockRealizationCacheReadModel) Delete(ctx context.Context, realizationID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedIDs = append(m.deletedIDs, realizationID)
	return nil
}

func (m *mockRealizationCacheReadModel) DeleteByCapabilityID(ctx context.Context, capabilityID string) error {
	if m.deleteByCapErr != nil {
		return m.deleteByCapErr
	}
	m.deletedCapabilityIDs = append(m.deletedCapabilityIDs, capabilityID)
	return nil
}

func (m *mockRealizationCacheReadModel) UpdateComponentName(ctx context.Context, componentID, componentName string) error {
	if m.updateNameErr != nil {
		return m.updateNameErr
	}
	m.updatedNames = append(m.updatedNames, componentNameUpdate{ComponentID: componentID, ComponentName: componentName})
	return nil
}

func TestRealizationCache_SystemLinked_UpsertsEntry(t *testing.T) {
	mock := &mockRealizationCacheReadModel{}
	projector := NewEARealizationCacheProjector(mock)

	realizationID := uuid.New().String()
	capabilityID := uuid.New().String()
	componentID := uuid.New().String()
	eventData, err := json.Marshal(systemLinkedToCapabilityEvent{
		ID:               realizationID,
		CapabilityID:     capabilityID,
		ComponentID:      componentID,
		ComponentName:    "My Component",
		RealizationLevel: "Direct",
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), cmPL.SystemLinkedToCapability, eventData)
	require.NoError(t, err)

	require.Len(t, mock.upsertedEntries, 1)
	entry := mock.upsertedEntries[0]
	assert.Equal(t, realizationID, entry.RealizationID)
	assert.Equal(t, capabilityID, entry.CapabilityID)
	assert.Equal(t, componentID, entry.ComponentID)
	assert.Equal(t, "My Component", entry.ComponentName)
	assert.Equal(t, "Direct", entry.Origin)
}

func TestRealizationCache_DeleteEvents(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		event     any
		assertFn  func(*testing.T, *mockRealizationCacheReadModel, string)
	}{
		{
			"realization deleted removes by ID",
			cmPL.SystemRealizationDeleted,
			systemRealizationDeletedEvent{},
			func(t *testing.T, m *mockRealizationCacheReadModel, id string) {
				require.Len(t, m.deletedIDs, 1)
				assert.Equal(t, id, m.deletedIDs[0])
			},
		},
		{
			"capability deleted removes by capability ID",
			cmPL.CapabilityDeleted,
			realizationCapabilityDeletedEvent{},
			func(t *testing.T, m *mockRealizationCacheReadModel, id string) {
				require.Len(t, m.deletedCapabilityIDs, 1)
				assert.Equal(t, id, m.deletedCapabilityIDs[0])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRealizationCacheReadModel{}
			projector := NewEARealizationCacheProjector(mock)

			id := uuid.New().String()
			eventData, err := json.Marshal(map[string]string{"id": id})
			require.NoError(t, err)

			err = projector.ProjectEvent(context.Background(), tt.eventType, eventData)
			require.NoError(t, err)

			tt.assertFn(t, mock, id)
		})
	}
}

func TestRealizationCache_ComponentUpdated_UpdatesName(t *testing.T) {
	mock := &mockRealizationCacheReadModel{}
	projector := NewEARealizationCacheProjector(mock)

	componentID := uuid.New().String()
	eventData, err := json.Marshal(applicationComponentUpdatedEvent{
		ID:   componentID,
		Name: "Renamed Component",
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), amPL.ApplicationComponentUpdated, eventData)
	require.NoError(t, err)

	require.Len(t, mock.updatedNames, 1)
	assert.Equal(t, componentID, mock.updatedNames[0].ComponentID)
	assert.Equal(t, "Renamed Component", mock.updatedNames[0].ComponentName)
}

func TestRealizationCache_UnknownEvent_Ignored(t *testing.T) {
	mock := &mockRealizationCacheReadModel{}
	projector := NewEARealizationCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mock.upsertedEntries)
	assert.Empty(t, mock.deletedIDs)
	assert.Empty(t, mock.deletedCapabilityIDs)
	assert.Empty(t, mock.updatedNames)
}

func TestRealizationCache_InvalidJSON_ReturnsError(t *testing.T) {
	mock := &mockRealizationCacheReadModel{}
	projector := NewEARealizationCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), cmPL.SystemLinkedToCapability, []byte("invalid"))
	assert.Error(t, err)
}

func TestRealizationCache_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockRealizationCacheReadModel
		eventType string
		eventData any
	}{
		{
			"upsert error",
			&mockRealizationCacheReadModel{upsertErr: errors.New("db error")},
			cmPL.SystemLinkedToCapability,
			systemLinkedToCapabilityEvent{ID: uuid.New().String(), CapabilityID: uuid.New().String(), ComponentID: uuid.New().String(), ComponentName: "C", RealizationLevel: "Direct"},
		},
		{
			"delete error",
			&mockRealizationCacheReadModel{deleteErr: errors.New("db error")},
			cmPL.SystemRealizationDeleted,
			systemRealizationDeletedEvent{ID: uuid.New().String()},
		},
		{
			"delete by capability error",
			&mockRealizationCacheReadModel{deleteByCapErr: errors.New("db error")},
			cmPL.CapabilityDeleted,
			realizationCapabilityDeletedEvent{ID: uuid.New().String()},
		},
		{
			"update name error",
			&mockRealizationCacheReadModel{updateNameErr: errors.New("db error")},
			amPL.ApplicationComponentUpdated,
			applicationComponentUpdatedEvent{ID: uuid.New().String(), Name: "N"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projector := NewEARealizationCacheProjector(tt.mock)
			eventData, _ := json.Marshal(tt.eventData)
			err := projector.ProjectEvent(context.Background(), tt.eventType, eventData)
			assert.Error(t, err)
		})
	}
}
