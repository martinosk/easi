package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCapabilityJourneyRepository struct {
	saved   []*aggregates.CapabilityJourney
	loaded  *aggregates.CapabilityJourney
	getErr  error
	saveErr error
}

func (m *mockCapabilityJourneyRepository) Save(_ context.Context, j *aggregates.CapabilityJourney) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, j)
	return nil
}

func (m *mockCapabilityJourneyRepository) GetByID(_ context.Context, _ string) (*aggregates.CapabilityJourney, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.loaded, nil
}

type mockActiveJourneyLookup struct {
	id          string
	exists      bool
	err         error
	askedKinds  []string
	activeKinds []string
}

func (m *mockActiveJourneyLookup) FindActiveJourneyIDForCapability(_ context.Context, _ string, kinds []string) (string, bool, error) {
	if m.err != nil {
		return "", false, m.err
	}
	m.askedKinds = kinds
	if m.activeKinds != nil {
		return m.id, intersects(kinds, m.activeKinds), nil
	}
	return m.id, m.exists, nil
}

func intersects(a, b []string) bool {
	for _, left := range a {
		for _, right := range b {
			if left == right {
				return true
			}
		}
	}
	return false
}

func maturityOf(value int) func(context.Context, string) (int, error) {
	return func(context.Context, string) (int, error) { return value, nil }
}

func alwaysExists(_ context.Context, _ string) (bool, error) { return true, nil }
func neverExists(_ context.Context, _ string) (bool, error)  { return false, nil }

func allExistingRefs() JourneyReferenceChecks {
	return JourneyReferenceChecks{
		CapabilityExists:              alwaysExists,
		ComponentExists:               alwaysExists,
		DomainExists:                  alwaysExists,
		CapabilityEffectivelyInDomain: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
	}
}

func validPlanJourneyCmd() *commands.PlanJourney {
	return &commands.PlanJourney{
		CapabilityID:     uuid.New().String(),
		Kind:             valueobjects.JourneyKindMigration,
		FromComponentIDs: []string{uuid.New().String()},
		ToComponentID:    uuid.New().String(),
		Note:             "moving to a modern platform",
		PlannedBy:        "architect@example.com",
	}
}

func TestPlanJourneyHandler_ValidMigration_CreatesJourney(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{exists: false}
	h := NewPlanJourneyHandler(repo, lookup, allExistingRefs(), nil)

	cmd := validPlanJourneyCmd()
	result, err := h.Handle(context.Background(), cmd)

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, repo.saved[0].ID(), result.CreatedID)
	assert.Equal(t, valueobjects.JourneyStatusPlanned, repo.saved[0].Status().Value())
}

func TestPlanJourneyHandler_ActiveJourneyExists_ReturnsErrorCarryingExistingID(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	existingID := uuid.New().String()
	lookup := &mockActiveJourneyLookup{id: existingID, exists: true}
	h := NewPlanJourneyHandler(repo, lookup, allExistingRefs(), nil)

	_, err := h.Handle(context.Background(), validPlanJourneyCmd())

	require.Error(t, err)
	var conflict *services.ActiveJourneyError
	require.True(t, errors.As(err, &conflict))
	assert.Equal(t, existingID, conflict.ExistingJourneyID)
	assert.Empty(t, repo.saved)
}

func TestPlanJourneyHandler_CapabilityDoesNotExist_Fails(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{exists: false}
	refs := allExistingRefs()
	refs.CapabilityExists = neverExists
	h := NewPlanJourneyHandler(repo, lookup, refs, nil)

	_, err := h.Handle(context.Background(), validPlanJourneyCmd())

	assert.ErrorIs(t, err, services.ErrReferencedEntityNotFound)
	assert.Empty(t, repo.saved)
}

func TestPlanJourneyHandler_ComponentDoesNotExist_Fails(t *testing.T) {
	cases := []struct {
		name          string
		missingCompFn func(cmd *commands.PlanJourney) string
	}{
		{"to-component missing", func(cmd *commands.PlanJourney) string { return cmd.ToComponentID }},
		{"from-component missing", func(cmd *commands.PlanJourney) string { return cmd.FromComponentIDs[0] }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCapabilityJourneyRepository{}
			lookup := &mockActiveJourneyLookup{exists: false}
			cmd := validPlanJourneyCmd()
			missingComponentID := tc.missingCompFn(cmd)
			refs := allExistingRefs()
			refs.ComponentExists = func(_ context.Context, id string) (bool, error) { return id != missingComponentID, nil }
			h := NewPlanJourneyHandler(repo, lookup, refs, nil)

			_, err := h.Handle(context.Background(), cmd)

			assert.ErrorIs(t, err, services.ErrReferencedEntityNotFound)
			assert.Empty(t, repo.saved)
		})
	}
}

func TestPlanJourneyHandler_KindCardinalityViolation_Fails(t *testing.T) {
	cases := []struct {
		name    string
		kind    string
		sources int
	}{
		{"migration with zero from-apps rejected (rule 3)", valueobjects.JourneyKindMigration, 0},
		{"consolidation with zero from-apps rejected (rule 3)", valueobjects.JourneyKindConsolidation, 0},
		{"carve-out with two from-apps rejected (rule 3)", valueobjects.JourneyKindCarveOut, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCapabilityJourneyRepository{}
			lookup := &mockActiveJourneyLookup{exists: false}
			h := NewPlanJourneyHandler(repo, lookup, allExistingRefs(), nil)
			cmd := validPlanJourneyCmd()
			cmd.Kind = tc.kind
			ids := make([]string, tc.sources)
			for i := range ids {
				ids[i] = uuid.New().String()
			}
			cmd.FromComponentIDs = ids

			_, err := h.Handle(context.Background(), cmd)

			assert.ErrorIs(t, err, valueobjects.ErrInvalidSourceApplicationCount)
			assert.Empty(t, repo.saved)
		})
	}
}

func TestPlanJourneyHandler_AggregateInvariantViolations_Fail(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(cmd *commands.PlanJourney)
		wantErr error
	}{
		{
			name:    "target among sources",
			mutate:  func(cmd *commands.PlanJourney) { cmd.FromComponentIDs = []string{cmd.ToComponentID} },
			wantErr: aggregates.ErrJourneyTargetAmongSources,
		},
		{
			name: "move without target domain",
			mutate: func(cmd *commands.PlanJourney) {
				cmd.Kind = valueobjects.JourneyKindMove
				cmd.FromComponentIDs = []string{}
			},
			wantErr: aggregates.ErrJourneyMoveRequiresTargetDomain,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCapabilityJourneyRepository{}
			lookup := &mockActiveJourneyLookup{exists: false}
			h := NewPlanJourneyHandler(repo, lookup, allExistingRefs(), nil)
			cmd := validPlanJourneyCmd()
			tc.mutate(cmd)

			_, err := h.Handle(context.Background(), cmd)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Empty(t, repo.saved)
		})
	}
}

func validMovePlanJourneyCmd() *commands.PlanJourney {
	cmd := validPlanJourneyCmd()
	cmd.Kind = valueobjects.JourneyKindMove
	cmd.FromComponentIDs = []string{}
	cmd.TargetDomainID = uuid.New().String()
	cmd.TargetParentID = uuid.New().String()
	cmd.ResultingName = "Freight invoicing"
	return cmd
}

func TestPlanJourneyHandler_ValidMove_VerifiesDomainAndParent(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{exists: false}
	cmd := validMovePlanJourneyCmd()

	var checkedCapability, checkedDomain string
	refs := allExistingRefs()
	refs.CapabilityEffectivelyInDomain = func(_ context.Context, capabilityID, domainID string) (bool, error) {
		checkedCapability, checkedDomain = capabilityID, domainID
		return true, nil
	}
	h := NewPlanJourneyHandler(repo, lookup, refs, nil)

	_, err := h.Handle(context.Background(), cmd)

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, cmd.TargetParentID, checkedCapability)
	assert.Equal(t, cmd.TargetDomainID, checkedDomain)
}

func TestPlanJourneyHandler_MoveParentReferenceViolations_Fail(t *testing.T) {
	cases := []struct {
		name    string
		refs    func(cmd *commands.PlanJourney) JourneyReferenceChecks
		wantErr error
	}{
		{
			name: "parent not effectively in target domain",
			refs: func(_ *commands.PlanJourney) JourneyReferenceChecks {
				refs := allExistingRefs()
				refs.CapabilityEffectivelyInDomain = func(_ context.Context, _, _ string) (bool, error) { return false, nil }
				return refs
			},
			wantErr: services.ErrTargetParentNotInTargetDomain,
		},
		{
			name: "parent does not exist",
			refs: func(cmd *commands.PlanJourney) JourneyReferenceChecks {
				refs := allExistingRefs()
				refs.CapabilityExists = func(_ context.Context, id string) (bool, error) { return id != cmd.TargetParentID, nil }
				return refs
			},
			wantErr: services.ErrReferencedEntityNotFound,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCapabilityJourneyRepository{}
			lookup := &mockActiveJourneyLookup{exists: false}
			cmd := validMovePlanJourneyCmd()
			h := NewPlanJourneyHandler(repo, lookup, tc.refs(cmd), nil)

			_, err := h.Handle(context.Background(), cmd)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Empty(t, repo.saved)
		})
	}
}

func validMaturityPlanJourneyCmd(target int) *commands.PlanJourney {
	cmd := validPlanJourneyCmd()
	cmd.Kind = valueobjects.JourneyKindMaturity
	cmd.FromComponentIDs = nil
	cmd.ToComponentID = ""
	cmd.TargetMaturity = &target
	return cmd
}

func TestPlanJourneyHandler_ValidMaturity_ReadsCurrentMaturityAndPlans(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{exists: false}
	h := NewPlanJourneyHandler(repo, lookup, allExistingRefs(), maturityOf(30))

	_, err := h.Handle(context.Background(), validMaturityPlanJourneyCmd(65))

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	require.NotNil(t, repo.saved[0].TargetMaturity())
	assert.Equal(t, 65, repo.saved[0].TargetMaturity().Value())
	assert.Equal(t, []string{valueobjects.JourneyKindMaturity}, lookup.askedKinds)
}

func TestPlanJourneyHandler_MaturityTargetNotAboveCurrent_Fails_Spec211Rule3(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{exists: false}
	h := NewPlanJourneyHandler(repo, lookup, allExistingRefs(), maturityOf(65))

	_, err := h.Handle(context.Background(), validMaturityPlanJourneyCmd(65))

	assert.ErrorIs(t, err, aggregates.ErrJourneyMaturityTargetNotAboveCurrent)
	assert.Empty(t, repo.saved)
}

func TestPlanJourneyHandler_MaturityAlongsideApplicationJourney_Spec211Rule6(t *testing.T) {
	cases := []struct {
		name        string
		activeKinds []string
		cmd         func() *commands.PlanJourney
		wantBlocked bool
	}{
		{"maturity is allowed beside an active migration", []string{valueobjects.JourneyKindMigration}, func() *commands.PlanJourney { return validMaturityPlanJourneyCmd(65) }, false},
		{"a second maturity is rejected", []string{valueobjects.JourneyKindMaturity}, func() *commands.PlanJourney { return validMaturityPlanJourneyCmd(65) }, true},
		{"a migration is allowed beside an active maturity", []string{valueobjects.JourneyKindMaturity}, validPlanJourneyCmd, false},
		{"a second migration is rejected", []string{valueobjects.JourneyKindMigration}, validPlanJourneyCmd, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCapabilityJourneyRepository{}
			lookup := &mockActiveJourneyLookup{id: uuid.New().String(), activeKinds: tc.activeKinds}
			h := NewPlanJourneyHandler(repo, lookup, allExistingRefs(), maturityOf(30))

			_, err := h.Handle(context.Background(), tc.cmd())

			if tc.wantBlocked {
				var conflict *services.ActiveJourneyError
				require.True(t, errors.As(err, &conflict))
				assert.Empty(t, repo.saved)
				return
			}
			require.NoError(t, err)
			assert.Len(t, repo.saved, 1)
		})
	}
}

func TestPlanJourneyHandler_MaturityTargetOutOfRange_Fails(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{exists: false}
	h := NewPlanJourneyHandler(repo, lookup, allExistingRefs(), maturityOf(30))

	_, err := h.Handle(context.Background(), validMaturityPlanJourneyCmd(120))

	assert.ErrorIs(t, err, valueobjects.ErrTargetMaturityOutOfRange)
	assert.Empty(t, repo.saved)
}

func TestPlanJourneyHandler_LookupError_Fails(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{err: errors.New("db down")}
	h := NewPlanJourneyHandler(repo, lookup, allExistingRefs(), nil)

	_, err := h.Handle(context.Background(), validPlanJourneyCmd())

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func TestPlanJourneyHandler_InvalidCommandType_Fails(t *testing.T) {
	repo := &mockCapabilityJourneyRepository{}
	lookup := &mockActiveJourneyLookup{}
	h := NewPlanJourneyHandler(repo, lookup, allExistingRefs(), nil)

	_, err := h.Handle(context.Background(), &commands.StartJourney{})

	assert.Error(t, err)
}
