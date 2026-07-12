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

func newCapabilityRef(t *testing.T) valueobjects.PhysicalCapabilityRef {
	t.Helper()
	ref, err := valueobjects.NewPhysicalCapabilityRef(uuid.New().String())
	require.NoError(t, err)
	return ref
}

func newComponentRef(t *testing.T) valueobjects.ApplicationRef {
	t.Helper()
	ref, err := valueobjects.NewApplicationRef(uuid.New().String())
	require.NoError(t, err)
	return ref
}

func newRationale(t *testing.T, v string) sharedvo.Description {
	t.Helper()
	d, err := sharedvo.NewDescription(v)
	require.NoError(t, err)
	return d
}

func newGrade(t *testing.T, v string) valueobjects.TimeGrade {
	t.Helper()
	g, err := valueobjects.NewTimeGrade(v)
	require.NoError(t, err)
	return g
}

func newRealizationID(t *testing.T) string {
	t.Helper()
	return uuid.New().String()
}

func TestNewTimeAssessment_Succeeds(t *testing.T) {
	cap := newCapabilityRef(t)
	comp := newComponentRef(t)
	realizationID := newRealizationID(t)
	grade := newGrade(t, valueobjects.TimeGradeMigrate)
	rationale := newRationale(t, "carve-out candidate")

	ta, err := NewTimeAssessment(TimeAssessmentFacts{
		CapabilityID:  cap,
		ComponentID:   comp,
		RealizationID: realizationID,
		Grade:         grade,
		Rationale:     rationale,
		AssessedBy:    "architect@example.com",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, ta.ID())
	assert.Equal(t, cap.Value(), ta.CapabilityID().Value())
	assert.Equal(t, comp.Value(), ta.ComponentID().Value())
	assert.Equal(t, valueobjects.TimeGradeMigrate, ta.Grade().Value())
	assert.Equal(t, "carve-out candidate", ta.Rationale().Value())
	assert.Equal(t, "architect@example.com", ta.AssessedBy())
	assert.False(t, ta.AssessedAt().IsZero())
	assert.False(t, ta.IsRemoved())

	changes := ta.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.TimeAssessmentRecorded)
	require.True(t, ok)
	assert.Equal(t, cap.Value(), evt.CapabilityID)
	assert.Equal(t, comp.Value(), evt.ComponentID)
	assert.Equal(t, realizationID, evt.RealizationID)
	assert.Equal(t, valueobjects.TimeGradeMigrate, evt.Grade)
	assert.Empty(t, evt.PreviousGrade, "first assessment carries no previous grade")
	assert.Equal(t, "architect@example.com", evt.AssessedBy)
	assert.False(t, evt.OccurredOn.IsZero(), "BR3: server timestamp recorded on the event")
}

func TestNewTimeAssessment_EmptyRationale_Succeeds(t *testing.T) {
	ta, err := NewTimeAssessment(TimeAssessmentFacts{
		CapabilityID:  newCapabilityRef(t),
		ComponentID:   newComponentRef(t),
		RealizationID: newRealizationID(t),
		Grade:         newGrade(t, valueobjects.TimeGradeInvest),
		Rationale:     sharedvo.Description{},
		AssessedBy:    "a@example.com",
	})
	require.NoError(t, err)
	assert.True(t, ta.Rationale().IsEmpty())
}

func TestTimeAssessment_Reassess_ReplacesGradeAndCarriesPrevious(t *testing.T) {
	ta, err := NewTimeAssessment(TimeAssessmentFacts{
		CapabilityID:  newCapabilityRef(t),
		ComponentID:   newComponentRef(t),
		RealizationID: newRealizationID(t),
		Grade:         newGrade(t, valueobjects.TimeGradeTolerate),
		Rationale:     newRationale(t, "first"),
		AssessedBy:    "a@example.com",
	})
	require.NoError(t, err)
	ta.MarkChangesAsCommitted()

	err = ta.Reassess(newRealizationID(t), newGrade(t, valueobjects.TimeGradeEliminate), newRationale(t, "reconsidered"), "b@example.com")

	require.NoError(t, err)
	assert.Equal(t, valueobjects.TimeGradeEliminate, ta.Grade().Value())
	assert.Equal(t, "reconsidered", ta.Rationale().Value())
	assert.Equal(t, "b@example.com", ta.AssessedBy())

	changes := ta.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.TimeAssessmentRecorded)
	require.True(t, ok)
	assert.Equal(t, valueobjects.TimeGradeEliminate, evt.Grade)
	assert.Equal(t, valueobjects.TimeGradeTolerate, evt.PreviousGrade, "the replaced grade must be preserved on the event for history reconstruction")
}

func TestTimeAssessment_Reassess_SameGrade_IsValidReaffirmation(t *testing.T) {
	ta, err := NewTimeAssessment(TimeAssessmentFacts{
		CapabilityID:  newCapabilityRef(t),
		ComponentID:   newComponentRef(t),
		RealizationID: newRealizationID(t),
		Grade:         newGrade(t, valueobjects.TimeGradeMigrate),
		Rationale:     newRationale(t, "first"),
		AssessedBy:    "a@example.com",
	})
	require.NoError(t, err)
	ta.MarkChangesAsCommitted()
	firstAssessedAt := ta.AssessedAt()

	time.Sleep(time.Millisecond)
	err = ta.Reassess(newRealizationID(t), newGrade(t, valueobjects.TimeGradeMigrate), newRationale(t, "still true"), "b@example.com")

	require.NoError(t, err)
	assert.Equal(t, valueobjects.TimeGradeMigrate, ta.Grade().Value())
	assert.Equal(t, "b@example.com", ta.AssessedBy(), "re-affirmation refreshes the assessor")
	assert.True(t, ta.AssessedAt().After(firstAssessedAt), "re-affirmation refreshes the date")

	changes := ta.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt := changes[0].(events.TimeAssessmentRecorded)
	assert.Equal(t, valueobjects.TimeGradeMigrate, evt.PreviousGrade)
}

func TestTimeAssessment_Remove_RaisesRemovedEvent(t *testing.T) {
	ta, err := NewTimeAssessment(TimeAssessmentFacts{
		CapabilityID:  newCapabilityRef(t),
		ComponentID:   newComponentRef(t),
		RealizationID: newRealizationID(t),
		Grade:         newGrade(t, valueobjects.TimeGradeInvest),
		Rationale:     newRationale(t, "x"),
		AssessedBy:    "a@example.com",
	})
	require.NoError(t, err)
	ta.MarkChangesAsCommitted()

	err = ta.Remove("remover@example.com")

	require.NoError(t, err)
	assert.True(t, ta.IsRemoved())
	changes := ta.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.TimeAssessmentRemoved)
	require.True(t, ok)
	assert.Equal(t, "remover@example.com", evt.RemovedBy)
	assert.False(t, evt.OccurredOn.IsZero())
}

func newRemovedAssessment(t *testing.T) *TimeAssessment {
	t.Helper()
	ta, err := NewTimeAssessment(TimeAssessmentFacts{
		CapabilityID:  newCapabilityRef(t),
		ComponentID:   newComponentRef(t),
		RealizationID: newRealizationID(t),
		Grade:         newGrade(t, valueobjects.TimeGradeInvest),
		Rationale:     newRationale(t, "x"),
		AssessedBy:    "a@example.com",
	})
	require.NoError(t, err)
	require.NoError(t, ta.Remove("a@example.com"))
	return ta
}

func TestTimeAssessment_Reassess_AfterRemoval_Fails(t *testing.T) {
	ta := newRemovedAssessment(t)

	err := ta.Reassess(newRealizationID(t), newGrade(t, valueobjects.TimeGradeMigrate), newRationale(t, "y"), "a@example.com")
	assert.ErrorIs(t, err, ErrTimeAssessmentAlreadyRemoved)
}

func TestTimeAssessment_Remove_AfterRemoval_Fails(t *testing.T) {
	ta := newRemovedAssessment(t)

	err := ta.Remove("a@example.com")
	assert.ErrorIs(t, err, ErrTimeAssessmentAlreadyRemoved)
}

func TestLoadTimeAssessmentFromHistory_RehydratesCurrent(t *testing.T) {
	ta, err := NewTimeAssessment(TimeAssessmentFacts{
		CapabilityID:  newCapabilityRef(t),
		ComponentID:   newComponentRef(t),
		RealizationID: newRealizationID(t),
		Grade:         newGrade(t, valueobjects.TimeGradeTolerate),
		Rationale:     newRationale(t, "first"),
		AssessedBy:    "a@example.com",
	})
	require.NoError(t, err)
	require.NoError(t, ta.Reassess(newRealizationID(t), newGrade(t, valueobjects.TimeGradeEliminate), newRationale(t, "second"), "b@example.com"))
	history := ta.GetUncommittedChanges()

	loaded, err := LoadTimeAssessmentFromHistory(history)

	require.NoError(t, err)
	assert.Equal(t, ta.ID(), loaded.ID())
	assert.Equal(t, valueobjects.TimeGradeEliminate, loaded.Grade().Value())
	assert.Equal(t, "second", loaded.Rationale().Value())
	assert.Equal(t, "b@example.com", loaded.AssessedBy())
	assert.False(t, loaded.IsRemoved())
	assert.Empty(t, loaded.GetUncommittedChanges())
}

func TestLoadTimeAssessmentFromHistory_RehydratesRemoved(t *testing.T) {
	ta := newRemovedAssessment(t)
	history := ta.GetUncommittedChanges()

	loaded, err := LoadTimeAssessmentFromHistory(history)

	require.NoError(t, err)
	assert.True(t, loaded.IsRemoved())
}

func TestApplyTimeAssessment_UnknownEvent_Fails(t *testing.T) {
	_, err := LoadTimeAssessmentFromHistory([]domain.DomainEvent{unknownTimeAssessmentEventForTest{}})
	assert.ErrorIs(t, err, ErrUnknownTimeAssessmentEvent)
}

type unknownTimeAssessmentEventForTest struct{}

func (unknownTimeAssessmentEventForTest) AggregateID() string               { return "" }
func (unknownTimeAssessmentEventForTest) EventType() string                 { return "UnknownEvent" }
func (unknownTimeAssessmentEventForTest) EventData() map[string]interface{} { return nil }
func (unknownTimeAssessmentEventForTest) OccurredAt() time.Time             { return time.Time{} }
