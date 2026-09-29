package aggregates

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/stewardship/domain/events"
	"easi/backend/internal/stewardship/domain/valueobjects"
)

func concern(t *testing.T, value string) valueobjects.Concern {
	t.Helper()
	c, err := valueobjects.NewConcern(value)
	require.NoError(t, err)
	return c
}

func steward(t *testing.T, id string) valueobjects.StewardRef {
	t.Helper()
	ref, err := valueobjects.NewStewardRef(id, valueobjects.UserActive)
	require.NoError(t, err)
	return ref
}

func newAssessmentStewardship(t *testing.T) *Stewardship {
	t.Helper()
	s, err := NewStewardship(Assignment{
		DomainID:   "domain-1",
		Concern:    concern(t, "assessment"),
		Steward:    steward(t, "mette"),
		AssignedBy: "alice@example.com",
	})
	require.NoError(t, err)
	return s
}

func TestNewStewardship_RaisesStewardAssignedWithAttribution(t *testing.T) {
	s := newAssessmentStewardship(t)

	raised := s.GetUncommittedChanges()
	require.Len(t, raised, 1)
	assigned, ok := raised[0].(events.StewardAssigned)
	require.True(t, ok)
	assert.Equal(t, s.ID(), assigned.ID)
	assert.Equal(t, "domain-1", assigned.DomainID)
	assert.Equal(t, "assessment", assigned.Concern)
	assert.Equal(t, "mette", assigned.StewardID)
	assert.Equal(t, "alice@example.com", assigned.AssignedBy)
	assert.False(t, assigned.AssignedAt.IsZero())
	assert.NotEmpty(t, s.ID())
	assert.Equal(t, "mette", s.Steward().Value())
}

func TestNewStewardship_RequiresTheDomain(t *testing.T) {
	_, err := NewStewardship(Assignment{Concern: concern(t, "assessment"), Steward: steward(t, "mette"), AssignedBy: "alice@example.com"})
	assert.ErrorIs(t, err, ErrDomainRequired)
}

func TestNewStewardship_RequiresTheActor(t *testing.T) {
	_, err := NewStewardship(Assignment{DomainID: "domain-1", Concern: concern(t, "assessment"), Steward: steward(t, "mette")})
	assert.ErrorIs(t, err, ErrActorRequired)
}

func TestAssign_ReplacesTheStewardInOneEvent(t *testing.T) {
	s := newAssessmentStewardship(t)
	s.MarkChangesAsCommitted()

	require.NoError(t, s.Assign(steward(t, "jonas"), "bob@example.com"))

	raised := s.GetUncommittedChanges()
	require.Len(t, raised, 1)
	assigned := raised[0].(events.StewardAssigned)
	assert.Equal(t, "jonas", assigned.StewardID)
	assert.Equal(t, "bob@example.com", assigned.AssignedBy)
	assert.Equal(t, "jonas", s.Steward().Value())
}

func TestAssign_TheCurrentStewardAgainChangesNothing(t *testing.T) {
	s := newAssessmentStewardship(t)
	s.MarkChangesAsCommitted()
	originalAt := s.AssignedAt()

	require.NoError(t, s.Assign(steward(t, "mette"), "bob@example.com"))

	assert.Empty(t, s.GetUncommittedChanges())
	assert.Equal(t, "alice@example.com", s.AssignedBy())
	assert.Equal(t, originalAt, s.AssignedAt())
}

func TestAssign_RequiresTheActor(t *testing.T) {
	s := newAssessmentStewardship(t)
	assert.ErrorIs(t, s.Assign(steward(t, "jonas"), ""), ErrActorRequired)
}

func TestAssign_AReleasedStewardshipIsRefused(t *testing.T) {
	s := newAssessmentStewardship(t)
	require.NoError(t, s.Release("bob@example.com"))

	assert.ErrorIs(t, s.Assign(steward(t, "jonas"), "bob@example.com"), ErrStewardshipReleased)
}

func TestRelease_RaisesStewardReleased(t *testing.T) {
	s := newAssessmentStewardship(t)
	s.MarkChangesAsCommitted()

	require.NoError(t, s.Release("system:domain-deleted"))

	raised := s.GetUncommittedChanges()
	require.Len(t, raised, 1)
	released := raised[0].(events.StewardReleased)
	assert.Equal(t, s.ID(), released.ID)
	assert.Equal(t, "domain-1", released.DomainID)
	assert.Equal(t, "assessment", released.Concern)
	assert.Equal(t, "system:domain-deleted", released.ReleasedBy)
	assert.False(t, released.ReleasedAt.IsZero())
	assert.True(t, s.IsReleased())
}

func TestRelease_TwiceChangesNothing(t *testing.T) {
	s := newAssessmentStewardship(t)
	require.NoError(t, s.Release("bob@example.com"))
	s.MarkChangesAsCommitted()

	require.NoError(t, s.Release("bob@example.com"))

	assert.Empty(t, s.GetUncommittedChanges())
}

func TestRelease_RequiresTheActor(t *testing.T) {
	s := newAssessmentStewardship(t)
	assert.ErrorIs(t, s.Release(""), ErrActorRequired)
}

func TestLoadStewardshipFromHistory_RestoresTheCurrentSteward(t *testing.T) {
	original := newAssessmentStewardship(t)
	require.NoError(t, original.Assign(steward(t, "jonas"), "bob@example.com"))
	history := append([]domain.DomainEvent{}, original.GetUncommittedChanges()...)

	loaded, err := LoadStewardshipFromHistory(history)

	require.NoError(t, err)
	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, "domain-1", loaded.DomainID())
	assert.Equal(t, "assessment", loaded.Concern().Value())
	assert.Equal(t, "jonas", loaded.Steward().Value())
	assert.Equal(t, "bob@example.com", loaded.AssignedBy())
	assert.False(t, loaded.IsReleased())
	assert.Equal(t, 2, loaded.Version())
}

func TestLoadStewardshipFromHistory_RejectsACorruptConcern(t *testing.T) {
	_, err := LoadStewardshipFromHistory([]domain.DomainEvent{
		events.NewStewardAssigned(events.StewardAssignedFields{ID: "s-1", DomainID: "domain-1", Concern: "budget", StewardID: "mette", AssignedBy: "alice@example.com"}),
	})
	assert.ErrorIs(t, err, ErrCorruptedStewardshipEvent)
}
