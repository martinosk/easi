package aggregates

import (
	"testing"
	"time"

	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAppRef(t *testing.T) valueobjects.ApplicationRef {
	t.Helper()
	ref, err := valueobjects.NewApplicationRef(uuid.New().String())
	require.NoError(t, err)
	return ref
}

func TestNewStandardApplication_Succeeds(t *testing.T) {
	ec := newECRef(t)
	app := newAppRef(t)
	narrative := newNarrative(t, "covers operational and reporting layers")

	sa, err := NewStandardApplication(ec, app, narrative)

	require.NoError(t, err)
	assert.NotEmpty(t, sa.ID())
	assert.Equal(t, ec.Value(), sa.EnterpriseCapabilityID().Value())
	assert.Equal(t, app.Value(), sa.CurrentApplication().Value())
	assert.Equal(t, narrative.Value(), sa.CurrentNarrative().Value())

	changes := sa.GetUncommittedChanges()
	require.Len(t, changes, 1)
	assert.Equal(t, "StandardApplicationSet", changes[0].EventType())
	evt, ok := changes[0].(events.StandardApplicationSet)
	require.True(t, ok)
	assert.Equal(t, ec.Value(), evt.EnterpriseCapabilityID)
	assert.Equal(t, app.Value(), evt.ApplicationID)
	assert.Empty(t, evt.PreviousApplicationID)
}

func TestNewStandardApplication_EmptyNarrative_Fails(t *testing.T) {
	_, err := NewStandardApplication(newECRef(t), newAppRef(t), sharedvo.Description{})
	assert.ErrorIs(t, err, ErrNarrativeRequiredForStandardApplication)
}

func TestStandardApplication_Change_RaisesEventCarryingPreviousApplication(t *testing.T) {
	ec := newECRef(t)
	app1 := newAppRef(t)
	sa, err := NewStandardApplication(ec, app1, newNarrative(t, "first"))
	require.NoError(t, err)
	sa.MarkChangesAsCommitted()

	app2 := newAppRef(t)
	err = sa.Change(app2, newNarrative(t, "second"))

	require.NoError(t, err)
	assert.Equal(t, app2.Value(), sa.CurrentApplication().Value())
	assert.Equal(t, "second", sa.CurrentNarrative().Value())

	changes := sa.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.StandardApplicationSet)
	require.True(t, ok)
	assert.Equal(t, app2.Value(), evt.ApplicationID)
	assert.Equal(t, app1.Value(), evt.PreviousApplicationID)
}

func TestStandardApplication_Change_EmptyNarrative_Fails(t *testing.T) {
	sa, _ := NewStandardApplication(newECRef(t), newAppRef(t), newNarrative(t, "first"))
	err := sa.Change(newAppRef(t), sharedvo.Description{})
	assert.ErrorIs(t, err, ErrNarrativeRequiredForStandardApplication)
}

func TestLoadStandardApplicationFromHistory_RehydratesCurrent(t *testing.T) {
	ec := newECRef(t)
	app1 := newAppRef(t)
	app2 := newAppRef(t)
	sa, _ := NewStandardApplication(ec, app1, newNarrative(t, "first"))
	require.NoError(t, sa.Change(app2, newNarrative(t, "second")))
	history := sa.GetUncommittedChanges()

	loaded, err := LoadStandardApplicationFromHistory(history)

	require.NoError(t, err)
	assert.Equal(t, sa.ID(), loaded.ID())
	assert.Equal(t, ec.Value(), loaded.EnterpriseCapabilityID().Value())
	assert.Equal(t, app2.Value(), loaded.CurrentApplication().Value())
	assert.Equal(t, "second", loaded.CurrentNarrative().Value())
}

func TestLoadStandardApplicationFromHistory_ThreeEventsPreservesOrder(t *testing.T) {
	ec := newECRef(t)
	appA := newAppRef(t)
	appB := newAppRef(t)
	appC := newAppRef(t)
	sa, _ := NewStandardApplication(ec, appA, newNarrative(t, "A"))
	require.NoError(t, sa.Change(appB, newNarrative(t, "B")))
	require.NoError(t, sa.Change(appC, newNarrative(t, "C")))
	history := sa.GetUncommittedChanges()
	require.Len(t, history, 3)

	loaded, err := LoadStandardApplicationFromHistory(history)

	require.NoError(t, err)
	assert.Equal(t, appC.Value(), loaded.CurrentApplication().Value(),
		"replay must reconstruct the current application from the last event in order")
	assert.Equal(t, "C", loaded.CurrentNarrative().Value())
	assert.Empty(t, loaded.GetUncommittedChanges(),
		"rehydration must not produce any uncommitted changes")
}

func TestApplyStandardApplication_UnknownEvent_Fails(t *testing.T) {
	_, err := LoadStandardApplicationFromHistory([]domain.DomainEvent{unknownEventForTest{}})
	assert.ErrorIs(t, err, ErrUnknownStandardApplicationEvent)
}

type unknownEventForTest struct{}

func (unknownEventForTest) AggregateID() string               { return "" }
func (unknownEventForTest) EventType() string                 { return "UnknownEvent" }
func (unknownEventForTest) EventData() map[string]interface{} { return nil }
func (unknownEventForTest) OccurredAt() time.Time             { return time.Time{} }
