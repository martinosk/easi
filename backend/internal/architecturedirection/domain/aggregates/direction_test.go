package aggregates

import (
	"testing"

	"easi/backend/internal/architecturedirection/domain/valueobjects"

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
		ref, err := valueobjects.NewPhysicalCapabilityRef(uuid.New().String())
		require.NoError(t, err)
		refs[i] = ref
	}
	return refs
}

func newPlacement(t *testing.T) valueobjects.Placement {
	t.Helper()
	p, err := valueobjects.NewPlacement(uuid.New().String(), "")
	require.NoError(t, err)
	return p
}

func newECRef(t *testing.T) valueobjects.EnterpriseCapabilityRef {
	t.Helper()
	ec, err := valueobjects.NewEnterpriseCapabilityRef(uuid.New().String())
	require.NoError(t, err)
	return ec
}

func newNarrative(t *testing.T, v string) valueobjects.Narrative {
	t.Helper()
	n, err := valueobjects.NewNarrative(v)
	require.NoError(t, err)
	return n
}

type draftOpts struct {
	directionType   string
	sourceCount     int
	sources         []valueobjects.PhysicalCapabilityRef
	includePlacement bool
	horizon         string
	narrative       string
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
	var placements []valueobjects.Placement
	if opts.includePlacement {
		placements = []valueobjects.Placement{newPlacement(t)}
	}
	var narrative valueobjects.Narrative
	if opts.narrative == "" {
		narrative = valueobjects.EmptyNarrative()
	} else {
		narrative = newNarrative(t, opts.narrative)
	}
	return DraftDirection(DraftParams{
		EnterpriseCapabilityID: newECRef(t),
		Type:                   newType(t, opts.directionType),
		SourceCapabilityIDs:    sources,
		Placements:             placements,
		Horizon:                newHorizon(t, opts.horizon),
		Narrative:              narrative,
	})
}


func TestDraftDirection_Consolidate_TwoSources_Succeeds(t *testing.T) {
	d, err := draftWith(t, draftOpts{sourceCount: 2, includePlacement: true})
	require.NoError(t, err)
	assert.NotEmpty(t, d.ID())
	assert.True(t, d.Status().IsDraft())
	assert.True(t, d.Type().IsConsolidate())
	assert.Len(t, d.SourceCapabilityIDs(), 2)
	assert.Len(t, d.GetUncommittedChanges(), 1)
	assert.Equal(t, "DirectionDrafted", d.GetUncommittedChanges()[0].EventType())
}

func TestDraftDirection_Consolidate_OneSource_Fails(t *testing.T) {
	_, err := draftWith(t, draftOpts{sourceCount: 1, includePlacement: true})
	assert.ErrorIs(t, err, ErrInvalidSourceCardinality)
}

func TestDraftDirection_Decompose_OneSource_Succeeds(t *testing.T) {
	d, err := draftWith(t, draftOpts{directionType: "decompose", sourceCount: 1, includePlacement: true, horizon: "now"})
	require.NoError(t, err)
	assert.True(t, d.Type().IsDecompose())
}

func TestDraftDirection_Decompose_TwoSources_Fails(t *testing.T) {
	_, err := draftWith(t, draftOpts{directionType: "decompose", sourceCount: 2, includePlacement: true, horizon: "now"})
	assert.ErrorIs(t, err, ErrInvalidSourceCardinality)
}

func TestDraftDirection_Stay_OneSource_NoPlacements_Succeeds(t *testing.T) {
	d, err := draftWith(t, draftOpts{directionType: "stay", sourceCount: 1, horizon: "now"})
	require.NoError(t, err)
	assert.True(t, d.Type().IsStay())
	assert.Empty(t, d.Placements())
}

func TestDraftDirection_Stay_WithPlacements_Fails(t *testing.T) {
	_, err := draftWith(t, draftOpts{directionType: "stay", sourceCount: 1, includePlacement: true, horizon: "now"})
	assert.ErrorIs(t, err, ErrInvalidPlacementCardinality)
}

func TestDraftDirection_Consolidate_NoPlacements_Fails(t *testing.T) {
	_, err := draftWith(t, draftOpts{sourceCount: 2, horizon: "now"})
	assert.ErrorIs(t, err, ErrInvalidPlacementCardinality)
}

func TestDraftDirection_Consolidate_MultiplePlacements_Fails(t *testing.T) {
	_, err := DraftDirection(DraftParams{
		EnterpriseCapabilityID: newECRef(t),
		Type:                   newType(t, "consolidate"),
		SourceCapabilityIDs:    newPhysicalRefs(t, 2),
		Placements:             []valueobjects.Placement{newPlacement(t), newPlacement(t)},
		Horizon:                newHorizon(t, "now"),
		Narrative:              valueobjects.EmptyNarrative(),
	})
	assert.ErrorIs(t, err, ErrInvalidPlacementCardinality)
}

func TestDraftDirection_Decompose_MultiplePlacements_Succeeds(t *testing.T) {
	d, err := DraftDirection(DraftParams{
		EnterpriseCapabilityID: newECRef(t),
		Type:                   newType(t, "decompose"),
		SourceCapabilityIDs:    newPhysicalRefs(t, 1),
		Placements:             []valueobjects.Placement{newPlacement(t), newPlacement(t)},
		Horizon:                newHorizon(t, "now"),
		Narrative:              valueobjects.EmptyNarrative(),
	})
	require.NoError(t, err)
	assert.Len(t, d.Placements(), 2)
}

func TestDraftDirection_DuplicateSourceIDs_Fails(t *testing.T) {
	refs := newPhysicalRefs(t, 1)
	dup := []valueobjects.PhysicalCapabilityRef{refs[0], refs[0]}
	_, err := draftWith(t, draftOpts{sources: dup, includePlacement: true})
	assert.ErrorIs(t, err, ErrDuplicateSourceCapabilities)
}


func draftConsolidate(t *testing.T) *Direction {
	t.Helper()
	d, err := draftWith(t, draftOpts{sourceCount: 2, includePlacement: true, narrative: "We consolidate."})
	require.NoError(t, err)
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
	d, err := draftWith(t, draftOpts{sourceCount: 2, includePlacement: true})
	require.NoError(t, err)
	d.MarkChangesAsCommitted()

	err = d.Propose()
	assert.ErrorIs(t, err, ErrNarrativeRequiredToPropose)
}

func TestPropose_FromAgreed_Fails(t *testing.T) {
	d := draftConsolidate(t)
	require.NoError(t, d.Propose())
	require.NoError(t, d.Agree())
	d.MarkChangesAsCommitted()
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
	d := draftConsolidate(t)
	require.NoError(t, d.Propose())
	require.NoError(t, d.Agree())
	d.MarkChangesAsCommitted()

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
	d := draftConsolidate(t)
	require.NoError(t, d.Propose())
	require.NoError(t, d.Agree())
	d.MarkChangesAsCommitted()

	err := d.ChangeHorizon(newHorizon(t, "later"))
	assert.ErrorIs(t, err, ErrDirectionAgreedImmutable)
}

func TestChangeHorizon_OnDraft_Succeeds(t *testing.T) {
	d := draftConsolidate(t)
	err := d.ChangeHorizon(newHorizon(t, "later"))
	require.NoError(t, err)
	assert.Equal(t, "later", d.Horizon().Value())
}

func TestChangeSourceCapabilities_OnAgreed_Fails(t *testing.T) {
	d := draftConsolidate(t)
	require.NoError(t, d.Propose())
	require.NoError(t, d.Agree())
	d.MarkChangesAsCommitted()

	err := d.ChangeSourceCapabilities(newPhysicalRefs(t, 2))
	assert.ErrorIs(t, err, ErrDirectionAgreedImmutable)
}

func TestChangeSourceCapabilities_RespectsTypeCardinality(t *testing.T) {
	d := draftConsolidate(t)
	err := d.ChangeSourceCapabilities(newPhysicalRefs(t, 1))
	assert.ErrorIs(t, err, ErrInvalidSourceCardinality)
}

func TestChangePlacements_OnAgreed_Fails(t *testing.T) {
	d := draftConsolidate(t)
	require.NoError(t, d.Propose())
	require.NoError(t, d.Agree())
	d.MarkChangesAsCommitted()

	err := d.ChangePlacements([]valueobjects.Placement{newPlacement(t)})
	assert.ErrorIs(t, err, ErrDirectionAgreedImmutable)
}


func TestLoadFromHistory_ReconstructsStatus(t *testing.T) {
	fresh, err := draftWith(t, draftOpts{sourceCount: 2, includePlacement: true, narrative: "Some narrative."})
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

func TestLoadFromHistory_AfterReject_IsTerminal(t *testing.T) {
	fresh, err := draftWith(t, draftOpts{sourceCount: 2, includePlacement: true, narrative: "x"})
	require.NoError(t, err)
	require.NoError(t, fresh.Reject())

	loaded, err := LoadDirectionFromHistory(fresh.GetUncommittedChanges())
	require.NoError(t, err)
	assert.True(t, loaded.Status().IsRejected())
	assert.False(t, loaded.Status().IsActive())
}
