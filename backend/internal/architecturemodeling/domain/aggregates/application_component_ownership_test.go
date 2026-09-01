package aggregates

import (
	"testing"

	"easi/backend/internal/architecturemodeling/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestComponent(t *testing.T) *ApplicationComponent {
	t.Helper()
	name, err := valueobjects.NewComponentName("Billing Service")
	require.NoError(t, err)
	component, err := NewApplicationComponent(name, valueobjects.MustNewDescription("Handles invoicing"))
	require.NoError(t, err)
	component.MarkChangesAsCommitted()
	return component
}

func userRef(t *testing.T) valueobjects.OwnerReference {
	t.Helper()
	ref, err := valueobjects.NewOwnerReference(valueobjects.OwnerKindUser, "user-123")
	require.NoError(t, err)
	return ref
}

func teamRef(t *testing.T) valueobjects.OwnerReference {
	t.Helper()
	ref, err := valueobjects.NewOwnerReference(valueobjects.OwnerKindTeam, "team-456")
	require.NoError(t, err)
	return ref
}

func TestApplicationComponent_StartsWithUnknownOwnership(t *testing.T) {
	component := newTestComponent(t)

	assert.True(t, component.OwnershipState().IsUnknown())
	_, hasOwner := component.Owner()
	assert.False(t, hasOwner)
}

type ownershipTransitionCase struct {
	name        string
	arrange     func(t *testing.T, c *ApplicationComponent)
	act         func(t *testing.T, c *ApplicationComponent) error
	wantState   string
	wantEvent   string
	wantOwnerID string
}

func runOwnershipTransitionCase(t *testing.T, tc ownershipTransitionCase) {
	component := newTestComponent(t)
	if tc.arrange != nil {
		tc.arrange(t, component)
		component.MarkChangesAsCommitted()
	}

	require.NoError(t, tc.act(t, component))

	assert.Equal(t, tc.wantState, component.OwnershipState().String())
	owner, hasOwner := component.Owner()
	if tc.wantOwnerID == "" {
		assert.False(t, hasOwner)
	} else {
		require.True(t, hasOwner)
		assert.Equal(t, tc.wantOwnerID, owner.ID())
	}

	events := component.GetUncommittedChanges()
	require.Len(t, events, 1)
	assert.Equal(t, tc.wantEvent, events[0].EventType())
}

func TestApplicationComponent_OwnershipTransitions(t *testing.T) {
	cases := []ownershipTransitionCase{
		{
			name:        "nominate user",
			act:         func(t *testing.T, c *ApplicationComponent) error { return c.NominateOwner(userRef(t)) },
			wantState:   valueobjects.OwnershipStateNominated,
			wantEvent:   "ApplicationOwnerNominated",
			wantOwnerID: "user-123",
		},
		{
			name: "confirm nominated user resolves to owned",
			arrange: func(t *testing.T, c *ApplicationComponent) {
				require.NoError(t, c.NominateOwner(userRef(t)))
			},
			act:         func(_ *testing.T, c *ApplicationComponent) error { return c.ConfirmOwnership() },
			wantState:   valueobjects.OwnershipStateOwned,
			wantEvent:   "ApplicationOwnershipConfirmed",
			wantOwnerID: "user-123",
		},
		{
			name: "confirm nominated team resolves to managed",
			arrange: func(t *testing.T, c *ApplicationComponent) {
				require.NoError(t, c.NominateOwner(teamRef(t)))
			},
			act:         func(_ *testing.T, c *ApplicationComponent) error { return c.ConfirmOwnership() },
			wantState:   valueobjects.OwnershipStateManaged,
			wantEvent:   "ApplicationOwnershipConfirmed",
			wantOwnerID: "team-456",
		},
		{
			name:        "assign user resolves to owned",
			act:         func(t *testing.T, c *ApplicationComponent) error { return c.AssignOwner(userRef(t)) },
			wantState:   valueobjects.OwnershipStateOwned,
			wantEvent:   "ApplicationOwnerAssigned",
			wantOwnerID: "user-123",
		},
		{
			name:        "assign team resolves to managed",
			act:         func(t *testing.T, c *ApplicationComponent) error { return c.AssignOwner(teamRef(t)) },
			wantState:   valueobjects.OwnershipStateManaged,
			wantEvent:   "ApplicationOwnerAssigned",
			wantOwnerID: "team-456",
		},
		{
			name: "clear assigned owner",
			arrange: func(t *testing.T, c *ApplicationComponent) {
				require.NoError(t, c.AssignOwner(userRef(t)))
			},
			act:       func(_ *testing.T, c *ApplicationComponent) error { return c.ClearOwnership() },
			wantState: valueobjects.OwnershipStateUnknown,
			wantEvent: "ApplicationOwnershipCleared",
		},
		{
			name: "clear nomination",
			arrange: func(t *testing.T, c *ApplicationComponent) {
				require.NoError(t, c.NominateOwner(teamRef(t)))
			},
			act:       func(_ *testing.T, c *ApplicationComponent) error { return c.ClearOwnership() },
			wantState: valueobjects.OwnershipStateUnknown,
			wantEvent: "ApplicationOwnershipCleared",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runOwnershipTransitionCase(t, tc)
		})
	}
}

func TestApplicationComponent_OwnershipTransitionRejections(t *testing.T) {
	cases := []struct {
		name    string
		arrange func(t *testing.T, c *ApplicationComponent)
		act     func(t *testing.T, c *ApplicationComponent) error
		wantErr error
	}{
		{
			name: "nominate while nominated",
			arrange: func(t *testing.T, c *ApplicationComponent) {
				require.NoError(t, c.NominateOwner(userRef(t)))
			},
			act:     func(t *testing.T, c *ApplicationComponent) error { return c.NominateOwner(teamRef(t)) },
			wantErr: ErrOwnershipNotUnknown,
		},
		{
			name: "assign while owned",
			arrange: func(t *testing.T, c *ApplicationComponent) {
				require.NoError(t, c.AssignOwner(userRef(t)))
			},
			act:     func(t *testing.T, c *ApplicationComponent) error { return c.AssignOwner(teamRef(t)) },
			wantErr: ErrOwnershipNotUnknown,
		},
		{
			name:    "confirm without nomination",
			act:     func(_ *testing.T, c *ApplicationComponent) error { return c.ConfirmOwnership() },
			wantErr: ErrNoNominationToConfirm,
		},
		{
			name:    "clear while unknown",
			act:     func(_ *testing.T, c *ApplicationComponent) error { return c.ClearOwnership() },
			wantErr: ErrNoOwnershipToClear,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			component := newTestComponent(t)
			if tc.arrange != nil {
				tc.arrange(t, component)
				component.MarkChangesAsCommitted()
			}

			err := tc.act(t, component)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Empty(t, component.GetUncommittedChanges())
		})
	}
}

func TestLoadApplicationComponentFromHistory_ReplaysOwnership(t *testing.T) {
	name, _ := valueobjects.NewComponentName("Billing Service")
	component, err := NewApplicationComponent(name, valueobjects.MustNewDescription(""))
	require.NoError(t, err)
	require.NoError(t, component.NominateOwner(teamRef(t)))
	require.NoError(t, component.ConfirmOwnership())

	reconstructed, err := LoadApplicationComponentFromHistory(component.GetUncommittedChanges())
	require.NoError(t, err)

	assert.Equal(t, valueobjects.OwnershipStateManaged, reconstructed.OwnershipState().String())
	owner, hasOwner := reconstructed.Owner()
	assert.True(t, hasOwner)
	assert.Equal(t, "team-456", owner.ID())
	assert.True(t, owner.IsTeam())
}

func TestLoadApplicationComponentFromHistory_NoOwnershipEventsReadsUnknown(t *testing.T) {
	name, _ := valueobjects.NewComponentName("Legacy Service")
	component, err := NewApplicationComponent(name, valueobjects.MustNewDescription(""))
	require.NoError(t, err)

	reconstructed, err := LoadApplicationComponentFromHistory(component.GetUncommittedChanges())
	require.NoError(t, err)

	assert.True(t, reconstructed.OwnershipState().IsUnknown())
	_, hasOwner := reconstructed.Owner()
	assert.False(t, hasOwner)
}
