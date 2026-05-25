package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/events"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStandardApplicationStore struct {
	upserts        []readmodels.UpsertStandardApplicationParams
	historyAppends []readmodels.AppendStandardApplicationHistoryParams
	upsertErr      error
	appendErr      error
}

func (m *mockStandardApplicationStore) UpsertCurrent(_ context.Context, p readmodels.UpsertStandardApplicationParams) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.upserts = append(m.upserts, p)
	return nil
}

func (m *mockStandardApplicationStore) AppendHistory(_ context.Context, p readmodels.AppendStandardApplicationHistoryParams) error {
	if m.appendErr != nil {
		return m.appendErr
	}
	m.historyAppends = append(m.historyAppends, p)
	return nil
}

func projectStandardApplicationEvent(t *testing.T, projector *StandardApplicationProjector, eventType string, payload map[string]interface{}) error {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	return projector.ProjectEvent(context.Background(), eventType, data)
}

func TestStandardApplicationProjector_FirstSet_UpsertsAndAppendsHistory(t *testing.T) {
	store := &mockStandardApplicationStore{}
	projector := NewStandardApplicationProjector(store)

	id := uuid.New().String()
	ec := uuid.New().String()
	app := uuid.New().String()
	evt := events.NewStandardApplicationSet(events.StandardApplicationSetFields{
		ID:                     id,
		EnterpriseCapabilityID: ec,
		ApplicationID:          app,
		Narrative:              "first",
	})

	require.NoError(t, projectStandardApplicationEvent(t, projector, evt.EventType(), evt.EventData()))

	require.Len(t, store.upserts, 1)
	assert.Equal(t, id, store.upserts[0].ID)
	assert.Equal(t, ec, store.upserts[0].EnterpriseCapabilityID)
	assert.Equal(t, app, store.upserts[0].ApplicationID)
	assert.Equal(t, "first", store.upserts[0].Narrative)
	assert.False(t, store.upserts[0].SetAt.IsZero())

	require.Len(t, store.historyAppends, 1)
	assert.Equal(t, id, store.historyAppends[0].StandardApplicationID)
	assert.Equal(t, app, store.historyAppends[0].ApplicationID)
	assert.Empty(t, store.historyAppends[0].PreviousApplicationID)
	assert.Equal(t, "first", store.historyAppends[0].Narrative)
}

func TestStandardApplicationProjector_Replacement_CarriesPreviousApplicationID(t *testing.T) {
	store := &mockStandardApplicationStore{}
	projector := NewStandardApplicationProjector(store)

	id := uuid.New().String()
	previousApp := uuid.New().String()
	newApp := uuid.New().String()
	evt := events.NewStandardApplicationSet(events.StandardApplicationSetFields{
		ID:                     id,
		EnterpriseCapabilityID: uuid.New().String(),
		ApplicationID:          newApp,
		PreviousApplicationID:  previousApp,
		Narrative:              "second",
	})

	require.NoError(t, projectStandardApplicationEvent(t, projector, evt.EventType(), evt.EventData()))

	require.Len(t, store.upserts, 1)
	assert.Equal(t, newApp, store.upserts[0].ApplicationID)

	require.Len(t, store.historyAppends, 1)
	assert.Equal(t, newApp, store.historyAppends[0].ApplicationID)
	assert.Equal(t, previousApp, store.historyAppends[0].PreviousApplicationID)
}

func TestStandardApplicationProjector_UnknownEvent_NoOp(t *testing.T) {
	store := &mockStandardApplicationStore{}
	projector := NewStandardApplicationProjector(store)

	err := projector.ProjectEvent(context.Background(), "SomeUnrelatedEvent", []byte(`{}`))

	assert.NoError(t, err)
	assert.Empty(t, store.upserts)
	assert.Empty(t, store.historyAppends)
}

func TestStandardApplicationProjector_UpsertError_Propagates(t *testing.T) {
	store := &mockStandardApplicationStore{upsertErr: errors.New("db")}
	projector := NewStandardApplicationProjector(store)

	evt := events.NewStandardApplicationSet(events.StandardApplicationSetFields{
		ID:                     uuid.New().String(),
		EnterpriseCapabilityID: uuid.New().String(),
		ApplicationID:          uuid.New().String(),
		Narrative:              "x",
	})
	err := projectStandardApplicationEvent(t, projector, evt.EventType(), evt.EventData())

	assert.Error(t, err)
	assert.Empty(t, store.historyAppends, "history must not be written when upsert failed")
}
