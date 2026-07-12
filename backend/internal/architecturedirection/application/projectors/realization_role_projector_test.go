package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRealizationRoleStore struct {
	registered  []string
	upserts     []readmodels.UpsertRealizationRoleParams
	deleted     []string
	registerErr error
	upsertErr   error
	deleteErr   error
}

func (m *mockRealizationRoleStore) RegisterAggregate(_ context.Context, capabilityID, aggregateID string) error {
	if m.registerErr != nil {
		return m.registerErr
	}
	m.registered = append(m.registered, capabilityID+"|"+aggregateID)
	return nil
}

func (m *mockRealizationRoleStore) UpsertRole(_ context.Context, p readmodels.UpsertRealizationRoleParams) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.upserts = append(m.upserts, p)
	return nil
}

func (m *mockRealizationRoleStore) DeleteRole(_ context.Context, capabilityID, componentID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deleted = append(m.deleted, capabilityID+"|"+componentID)
	return nil
}

func projectRealizationRoleEvent(t *testing.T, projector *RealizationRoleProjector, eventType string, payload map[string]interface{}) error {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	return projector.ProjectEvent(context.Background(), eventType, data)
}

func TestRealizationRoleProjector_Assigned_RegistersAggregateThenUpserts(t *testing.T) {
	store := &mockRealizationRoleStore{}
	projector := NewRealizationRoleProjector(store)

	aggID := uuid.New().String()
	capID := uuid.New().String()
	compID := uuid.New().String()
	realizationID := uuid.New().String()
	evt := events.NewRealizationRoleAssigned(events.RealizationRoleAssignedFields{
		ID:            aggID,
		CapabilityID:  capID,
		ComponentID:   compID,
		RealizationID: realizationID,
		Role:          valueobjects.RealizationRoleStandard,
		AssignedBy:    "a@example.com",
	})

	require.NoError(t, projectRealizationRoleEvent(t, projector, evt.EventType(), evt.EventData()))

	require.Equal(t, []string{capID + "|" + aggID}, store.registered)
	require.Len(t, store.upserts, 1)
	assert.Equal(t, capID, store.upserts[0].CapabilityID)
	assert.Equal(t, compID, store.upserts[0].ComponentID)
	assert.Equal(t, realizationID, store.upserts[0].RealizationID)
	assert.Equal(t, valueobjects.RealizationRoleStandard, store.upserts[0].Role)
	assert.Equal(t, "a@example.com", store.upserts[0].AssignedBy)
	assert.Equal(t, aggID, store.upserts[0].AggregateID)
	assert.Empty(t, store.upserts[0].DisplacedComponentID)
	assert.False(t, store.upserts[0].AssignedAt.IsZero())
}

func TestRealizationRoleProjector_Assigned_WithDisplacement_PassesDisplacedComponentID(t *testing.T) {
	store := &mockRealizationRoleStore{}
	projector := NewRealizationRoleProjector(store)

	evt := events.NewRealizationRoleAssigned(events.RealizationRoleAssignedFields{
		ID:                   uuid.New().String(),
		CapabilityID:         uuid.New().String(),
		ComponentID:          uuid.New().String(),
		RealizationID:        uuid.New().String(),
		Role:                 valueobjects.RealizationRoleStandard,
		DisplacedComponentID: uuid.New().String(),
		AssignedBy:           "a@example.com",
	})

	require.NoError(t, projectRealizationRoleEvent(t, projector, evt.EventType(), evt.EventData()))

	require.Len(t, store.upserts, 1)
	assert.Equal(t, evt.DisplacedComponentID, store.upserts[0].DisplacedComponentID,
		"the displaced component id must flow through so the read model can delete that row before upserting (partial unique index ordering)")
}

func TestRealizationRoleProjector_Cleared_DeletesRow(t *testing.T) {
	store := &mockRealizationRoleStore{}
	projector := NewRealizationRoleProjector(store)

	capID := uuid.New().String()
	compID := uuid.New().String()
	evt := events.NewRealizationRoleCleared(events.RealizationRoleClearedFields{
		ID:           uuid.New().String(),
		CapabilityID: capID,
		ComponentID:  compID,
		ClearedBy:    "a@example.com",
	})

	require.NoError(t, projectRealizationRoleEvent(t, projector, evt.EventType(), evt.EventData()))

	assert.Equal(t, []string{capID + "|" + compID}, store.deleted)
}

func TestRealizationRoleProjector_UnknownEvent_NoOp(t *testing.T) {
	store := &mockRealizationRoleStore{}
	projector := NewRealizationRoleProjector(store)

	err := projector.ProjectEvent(context.Background(), "SomeUnrelatedEvent", []byte(`{}`))

	assert.NoError(t, err)
	assert.Empty(t, store.upserts)
	assert.Empty(t, store.deleted)
	assert.Empty(t, store.registered)
}

func TestRealizationRoleProjector_StoreErrors_Propagate(t *testing.T) {
	assigned := events.NewRealizationRoleAssigned(events.RealizationRoleAssignedFields{
		ID:           uuid.New().String(),
		CapabilityID: uuid.New().String(),
		ComponentID:  uuid.New().String(),
		Role:         valueobjects.RealizationRoleLegacy,
		AssignedBy:   "a@example.com",
	})
	cleared := events.NewRealizationRoleCleared(events.RealizationRoleClearedFields{
		ID:           uuid.New().String(),
		CapabilityID: uuid.New().String(),
		ComponentID:  uuid.New().String(),
		ClearedBy:    "a@example.com",
	})

	cases := []struct {
		name      string
		store     *mockRealizationRoleStore
		eventType string
		payload   map[string]interface{}
	}{
		{"register error on assigned", &mockRealizationRoleStore{registerErr: errors.New("db")}, assigned.EventType(), assigned.EventData()},
		{"upsert error on assigned", &mockRealizationRoleStore{upsertErr: errors.New("db")}, assigned.EventType(), assigned.EventData()},
		{"delete error on cleared", &mockRealizationRoleStore{deleteErr: errors.New("db")}, cleared.EventType(), cleared.EventData()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			projector := NewRealizationRoleProjector(tc.store)
			err := projectRealizationRoleEvent(t, projector, tc.eventType, tc.payload)
			assert.Error(t, err)
		})
	}
}
