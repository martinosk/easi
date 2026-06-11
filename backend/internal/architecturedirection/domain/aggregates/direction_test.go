package aggregates

import (
	"testing"

	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newType(t *testing.T, v string) valueobjects.DirectionType {
	t.Helper()
	dt, err := valueobjects.NewDirectionType(v)
	require.NoError(t, err)
	return dt
}

func newHorizon(t *testing.T, v string) valueobjects.Horizon {
	t.Helper()
	h, err := valueobjects.NewHorizon(v)
	require.NoError(t, err)
	return h
}

func newPhysicalRefs(t *testing.T, n int) []valueobjects.PhysicalCapabilityRef {
	t.Helper()
	refs := make([]valueobjects.PhysicalCapabilityRef, n)
	for i := 0; i < n; i++ {
		refs[i] = newPhysicalRef(t)
	}
	return refs
}

func newPhysicalRef(t *testing.T) valueobjects.PhysicalCapabilityRef {
	t.Helper()
	ref, err := valueobjects.NewPhysicalCapabilityRef(uuid.New().String())
	require.NoError(t, err)
	return ref
}

func newECRef(t *testing.T) valueobjects.EnterpriseCapabilityRef {
	t.Helper()
	ec, err := valueobjects.NewEnterpriseCapabilityRef(uuid.New().String())
	require.NoError(t, err)
	return ec
}

func newNarrative(t *testing.T, v string) sharedvo.Description {
	t.Helper()
	n, err := sharedvo.NewDescription(v)
	require.NoError(t, err)
	return n
}

type draftOpts struct {
	directionType string
	sourceCount   int
	sources       []valueobjects.PhysicalCapabilityRef
	horizon       string
	narrative     string
}

func draftWith(t *testing.T, opts draftOpts) (*Direction, error) {
	t.Helper()
	if opts.horizon == "" {
		opts.horizon = "next"
	}
	if opts.directionType == "" {
		opts.directionType = "consolidate"
	}
	sources := opts.sources
	if sources == nil {
		sources = newPhysicalRefs(t, opts.sourceCount)
	}
	var narrative sharedvo.Description
	if opts.narrative != "" {
		narrative = newNarrative(t, opts.narrative)
	}
	return DraftDirection(DraftParams{
		EnterpriseCapabilityID: newECRef(t),
		Type:                   newType(t, opts.directionType),
		SourceCapabilityIDs:    sources,
		Horizon:                newHorizon(t, opts.horizon),
		Narrative:              narrative,
	})
}

func TestDraftDirection_Consolidate_TwoSources_Succeeds(t *testing.T) {
	d, err := draftWith(t, draftOpts{sourceCount: 2})
	require.NoError(t, err)
	assert.NotEmpty(t, d.ID())
	assert.True(t, d.Status().IsDraft())
	assert.True(t, d.Type().IsConsolidate())
	assert.Len(t, d.SourceCapabilityIDs(), 2)
	assert.Len(t, d.GetUncommittedChanges(), 1)
	assert.Equal(t, "DirectionDrafted", d.GetUncommittedChanges()[0].EventType())
}

func TestDraftDirection_SingleSourceIsAcceptedForAnyType(t *testing.T) {
	for _, directionType := range []string{"consolidate", "decompose", "stay"} {
		d, err := draftWith(t, draftOpts{directionType: directionType, sourceCount: 1, horizon: "now"})
		require.NoError(t, err, "a draft may carry a single source regardless of type (R8)")
		assert.True(t, d.Status().IsDraft())
	}
}

func TestDraftDirection_EmptySourceSetIsAccepted(t *testing.T) {
	d, err := draftWith(t, draftOpts{sourceCount: 0})
	require.NoError(t, err, "an empty source set is valid for a draft (R8)")
	assert.Empty(t, d.SourceCapabilityIDs())
}

func TestDraftDirection_DuplicateSourceIDs_Fails(t *testing.T) {
	refs := newPhysicalRefs(t, 1)
	dup := []valueobjects.PhysicalCapabilityRef{refs[0], refs[0]}
	_, err := draftWith(t, draftOpts{sources: dup})
	assert.ErrorIs(t, err, ErrDuplicateSourceCapabilities)
}

func draftConsolidate(t *testing.T) *Direction {
	t.Helper()
	d, err := draftWith(t, draftOpts{sourceCount: 2, narrative: "We consolidate."})
	require.NoError(t, err)
	d.MarkChangesAsCommitted()
	return d
}

func agreedConsolidate(t *testing.T) *Direction {
	t.Helper()
	d := draftConsolidate(t)
	require.NoError(t, d.Propose())
	require.NoError(t, d.Agree())
	d.MarkChangesAsCommitted()
	return d
}

func TestPropose_FromDraft_WithNarrative_Succeeds(t *testing.T) {
	d := draftConsolidate(t)
	err := d.Propose()
	require.NoError(t, err)
	assert.True(t, d.Status().IsProposed())
	uncommitted := d.GetUncommittedChanges()
	assert.Len(t, uncommitted, 1)
	assert.Equal(t, "DirectionProposed", uncommitted[0].EventType())
}

func TestPropose_WithoutNarrative_Fails(t *testing.T) {
	d, err := draftWith(t, draftOpts{sourceCount: 2})
	require.NoError(t, err)
	d.MarkChangesAsCommitted()

	err = d.Propose()
	assert.ErrorIs(t, err, ErrNarrativeRequiredToPropose)
}

func TestPropose_EnforcesTypeSourceCardinality(t *testing.T) {
	cases := []struct {
		name          string
		directionType string
		sourceCount   int
		wantErr       error
	}{
		{"consolidate with one source is rejected", "consolidate", 1, ErrInvalidSourceCardinality},
		{"consolidate with two sources advances", "consolidate", 2, nil},
		{"decompose with two sources is rejected", "decompose", 2, ErrInvalidSourceCardinality},
		{"decompose with one source advances", "decompose", 1, nil},
		{"stay without a source is rejected", "stay", 0, ErrInvalidSourceCardinality},
		{"stay with one source advances", "stay", 1, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := draftWith(t, draftOpts{directionType: tc.directionType, sourceCount: tc.sourceCount, horizon: "now", narrative: "Because."})
			require.NoError(t, err, "draft capture does not enforce cardinality (R8)")
			d.MarkChangesAsCommitted()

			err = d.Propose()
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.True(t, d.Status().IsProposed())
		})
	}
}

func TestPropose_FromAgreed_Fails(t *testing.T) {
	d := agreedConsolidate(t)
	err := d.Propose()
	assert.ErrorIs(t, err, ErrInvalidStatusTransition)
}

func TestAgree_FromProposed_Succeeds(t *testing.T) {
	d := draftConsolidate(t)
	require.NoError(t, d.Propose())
	d.MarkChangesAsCommitted()

	err := d.Agree()
	require.NoError(t, err)
	assert.True(t, d.Status().IsAgreed())
	uncommitted := d.GetUncommittedChanges()
	assert.Len(t, uncommitted, 1)
	assert.Equal(t, "DirectionAgreed", uncommitted[0].EventType())
}

func TestAgree_FromDraft_Fails(t *testing.T) {
	d := draftConsolidate(t)
	err := d.Agree()
	assert.ErrorIs(t, err, ErrInvalidStatusTransition)
}

func TestReject_FromDraft_Succeeds(t *testing.T) {
	d := draftConsolidate(t)
	err := d.Reject()
	require.NoError(t, err)
	assert.True(t, d.Status().IsRejected())
	assert.False(t, d.Status().IsActive())
}

func TestReject_FromProposed_Succeeds(t *testing.T) {
	d := draftConsolidate(t)
	require.NoError(t, d.Propose())
	d.MarkChangesAsCommitted()

	err := d.Reject()
	require.NoError(t, err)
	assert.True(t, d.Status().IsRejected())
}

func TestReject_FromAgreed_Succeeds(t *testing.T) {
	d := agreedConsolidate(t)
	err := d.Reject()
	require.NoError(t, err)
	assert.True(t, d.Status().IsRejected())
	assert.False(t, d.Status().IsActive())
}

func TestReject_FromRejected_Fails(t *testing.T) {
	d := draftConsolidate(t)
	require.NoError(t, d.Reject())
	d.MarkChangesAsCommitted()

	err := d.Reject()
	assert.ErrorIs(t, err, ErrInvalidStatusTransition)
}

func TestUpdateNarrative_PreAgreed_Succeeds(t *testing.T) {
	d := draftConsolidate(t)
	err := d.UpdateNarrative(newNarrative(t, "Updated narrative."))
	require.NoError(t, err)
	assert.Equal(t, "Updated narrative.", d.Narrative().Value())
}

func TestChangeHorizon_OnAgreed_Fails(t *testing.T) {
	d := agreedConsolidate(t)
	err := d.ChangeHorizon(newHorizon(t, "later"))
	assert.ErrorIs(t, err, ErrDirectionAgreedImmutable)
}

func TestChangeHorizon_OnDraft_Succeeds(t *testing.T) {
	d := draftConsolidate(t)
	err := d.ChangeHorizon(newHorizon(t, "later"))
	require.NoError(t, err)
	assert.Equal(t, "later", d.Horizon().Value())
}

func TestAddSourceCapability_AppendsAndRecordsActor(t *testing.T) {
	d := draftConsolidate(t)
	added := newPhysicalRef(t)

	err := d.AddSourceCapability(added, "architect@dfds.com")

	require.NoError(t, err)
	assert.Len(t, d.SourceCapabilityIDs(), 3)
	uncommitted := d.GetUncommittedChanges()
	require.Len(t, uncommitted, 1)
	event, ok := uncommitted[0].(events.DirectionSourceCapabilitiesChanged)
	require.True(t, ok)
	assert.Len(t, event.SourceCapabilityIDs, 3)
	assert.Contains(t, event.SourceCapabilityIDs, added.Value())
	assert.Equal(t, "architect@dfds.com", event.Actor, "source-set changes must record the acting architect (R9)")
}

func TestAddSourceCapability_AlreadyPresent_IsIdempotent(t *testing.T) {
	d := draftConsolidate(t)
	existing := d.SourceCapabilityIDs()[0]

	err := d.AddSourceCapability(existing, "architect@dfds.com")

	require.NoError(t, err)
	assert.Len(t, d.SourceCapabilityIDs(), 2)
	assert.Empty(t, d.GetUncommittedChanges(), "re-adding an existing source emits no event")
}

func TestAddSourceCapability_OnAgreed_Fails(t *testing.T) {
	d := agreedConsolidate(t)
	err := d.AddSourceCapability(newPhysicalRef(t), "architect@dfds.com")
	assert.ErrorIs(t, err, ErrDirectionAgreedImmutable)
}

func TestRemoveSourceCapability_RemovesAndRecordsActor(t *testing.T) {
	d := draftConsolidate(t)
	removed := d.SourceCapabilityIDs()[0]

	err := d.RemoveSourceCapability(removed, "architect@dfds.com")

	require.NoError(t, err)
	assert.Len(t, d.SourceCapabilityIDs(), 1)
	uncommitted := d.GetUncommittedChanges()
	require.Len(t, uncommitted, 1)
	event, ok := uncommitted[0].(events.DirectionSourceCapabilitiesChanged)
	require.True(t, ok)
	assert.NotContains(t, event.SourceCapabilityIDs, removed.Value())
	assert.Equal(t, "architect@dfds.com", event.Actor)
}

func TestRemoveSourceCapability_LeavingDraftWithSingleSourceIsAllowed(t *testing.T) {
	d := draftConsolidate(t)
	require.NoError(t, d.RemoveSourceCapability(d.SourceCapabilityIDs()[0], "architect@dfds.com"))

	err := d.RemoveSourceCapability(d.SourceCapabilityIDs()[0], "architect@dfds.com")

	require.NoError(t, err, "a draft source set may shrink below the type cardinality (R8)")
	assert.Empty(t, d.SourceCapabilityIDs())
}

func TestRemoveSourceCapability_NotInSourceSet_Fails(t *testing.T) {
	d := draftConsolidate(t)
	err := d.RemoveSourceCapability(newPhysicalRef(t), "architect@dfds.com")
	assert.ErrorIs(t, err, ErrSourceCapabilityNotInDirection)
}

func TestRemoveSourceCapability_OnAgreed_Fails(t *testing.T) {
	d := agreedConsolidate(t)
	err := d.RemoveSourceCapability(d.SourceCapabilityIDs()[0], "architect@dfds.com")
	assert.ErrorIs(t, err, ErrDirectionAgreedImmutable)
}

func TestLoadFromHistory_ReconstructsStatus(t *testing.T) {
	fresh, err := draftWith(t, draftOpts{sourceCount: 2, narrative: "Some narrative."})
	require.NoError(t, err)
	require.NoError(t, fresh.Propose())
	require.NoError(t, fresh.Agree())

	hist := fresh.GetUncommittedChanges()
	require.Len(t, hist, 3)

	loaded, err := LoadDirectionFromHistory(hist)
	require.NoError(t, err)
	assert.True(t, loaded.Status().IsAgreed())
	assert.Equal(t, fresh.ID(), loaded.ID())
	assert.Equal(t, fresh.Type().Value(), loaded.Type().Value())
	assert.Equal(t, "next", loaded.Horizon().Value())
}

func TestLoadFromHistory_ReconstructsSourceSetFromChanges(t *testing.T) {
	fresh, err := draftWith(t, draftOpts{sourceCount: 1, narrative: "x"})
	require.NoError(t, err)
	added := newPhysicalRef(t)
	require.NoError(t, fresh.AddSourceCapability(added, "architect@dfds.com"))

	loaded, err := LoadDirectionFromHistory(fresh.GetUncommittedChanges())
	require.NoError(t, err)
	require.Len(t, loaded.SourceCapabilityIDs(), 2)
}

func TestLoadFromHistory_AfterReject_IsTerminal(t *testing.T) {
	fresh, err := draftWith(t, draftOpts{sourceCount: 2, narrative: "x"})
	require.NoError(t, err)
	require.NoError(t, fresh.Reject())

	loaded, err := LoadDirectionFromHistory(fresh.GetUncommittedChanges())
	require.NoError(t, err)
	assert.True(t, loaded.Status().IsRejected())
	assert.False(t, loaded.Status().IsActive())
}
