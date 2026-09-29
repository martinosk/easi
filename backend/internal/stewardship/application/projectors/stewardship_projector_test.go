package projectors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/stewardship/application/readmodels"
	"easi/backend/internal/stewardship/domain/events"
)

type fakeStewardshipStore struct {
	rows    map[string]readmodels.StewardshipRow
	deleted []string
}

func (f *fakeStewardshipStore) Upsert(_ context.Context, row readmodels.StewardshipRow) error {
	f.rows[row.ID] = row
	return nil
}

func (f *fakeStewardshipStore) Delete(_ context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	delete(f.rows, id)
	return nil
}

func TestStewardshipProjector_RecordsTheCurrentSteward(t *testing.T) {
	store := &fakeStewardshipStore{rows: map[string]readmodels.StewardshipRow{}}
	event := events.NewStewardAssigned(events.StewardAssignedFields{ID: "s-1", DomainID: "ce", Concern: "assessment", StewardID: "mette", AssignedBy: "alice@example.com"})

	require.NoError(t, NewStewardshipProjector(store).Handle(ctx, event))

	row := store.rows["s-1"]
	assert.Equal(t, "ce", row.DomainID)
	assert.Equal(t, "assessment", row.Concern)
	assert.Equal(t, "mette", row.StewardID)
	assert.Equal(t, "alice@example.com", row.AssignedBy)
	assert.WithinDuration(t, event.AssignedAt, row.AssignedAt, time.Millisecond)
}

func TestStewardshipProjector_ReleaseRemovesTheRow(t *testing.T) {
	store := &fakeStewardshipStore{rows: map[string]readmodels.StewardshipRow{"s-1": {ID: "s-1"}}}

	require.NoError(t, NewStewardshipProjector(store).Handle(ctx, events.NewStewardReleased("s-1", "ce", "assessment", "bob@example.com")))

	assert.Equal(t, []string{"s-1"}, store.deleted)
	assert.Empty(t, store.rows)
}

func TestStewardshipProjector_IgnoresOtherEvents(t *testing.T) {
	store := &fakeStewardshipStore{rows: map[string]readmodels.StewardshipRow{}}

	require.NoError(t, NewStewardshipProjector(store).Handle(ctx, storedEvent(t, "x", "SomethingElse", map[string]any{"id": "x"})))

	assert.Empty(t, store.rows)
	assert.Empty(t, store.deleted)
}
