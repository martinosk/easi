package aggregates

import (
	"testing"
	"time"

	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRole(t *testing.T, v string) valueobjects.RealizationRole {
	t.Helper()
	role, err := valueobjects.NewRealizationRole(v)
	require.NoError(t, err)
	return role
}

func TestNewRealizationRoles_Succeeds(t *testing.T) {
	cap := newCapabilityRef(t)
	component := newComponentRef(t)
	realizationID := newRealizationID(t)
	role := newRole(t, valueobjects.RealizationRoleStandard)

	rr, err := NewRealizationRoles(RealizationRolesFacts{
		CapabilityID:  cap,
		ComponentID:   component,
		RealizationID: realizationID,
		Role:          role,
		AssignedBy:    "architect@example.com",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, rr.ID())
	assert.Equal(t, cap.Value(), rr.CapabilityID().Value())

	current, ok := rr.RoleFor(component)
	require.True(t, ok)
	assert.Equal(t, valueobjects.RealizationRoleStandard, current.Value())

	changes := rr.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.RealizationRoleAssigned)
	require.True(t, ok)
	assert.Equal(t, cap.Value(), evt.CapabilityID)
	assert.Equal(t, component.Value(), evt.ComponentID)
	assert.Equal(t, realizationID, evt.RealizationID)
	assert.Equal(t, valueobjects.RealizationRoleStandard, evt.Role)
	assert.Empty(t, evt.DisplacedComponentID, "first assignment displaces nothing")
	assert.Equal(t, "architect@example.com", evt.AssignedBy)
	assert.False(t, evt.OccurredOn.IsZero(), "BR7: server timestamp recorded on the event")
}

func TestRealizationRoles_Assign_Standard_DisplacesPreviousHolder(t *testing.T) {
	cap := newCapabilityRef(t)
	seabook := newComponentRef(t)
	phoenix := newComponentRef(t)
	rr, err := NewRealizationRoles(RealizationRolesFacts{
		CapabilityID:  cap,
		ComponentID:   seabook,
		RealizationID: newRealizationID(t),
		Role:          newRole(t, valueobjects.RealizationRoleStandard),
		AssignedBy:    "a@example.com",
	})
	require.NoError(t, err)
	rr.MarkChangesAsCommitted()

	err = rr.Assign(phoenix, newRealizationID(t), newRole(t, valueobjects.RealizationRoleStandard), "b@example.com")

	require.NoError(t, err)
	_, seabookHasRole := rr.RoleFor(seabook)
	assert.False(t, seabookHasRole, "the previous standard holder becomes unclassified")
	phoenixRole, ok := rr.RoleFor(phoenix)
	require.True(t, ok)
	assert.Equal(t, valueobjects.RealizationRoleStandard, phoenixRole.Value())

	changes := rr.GetUncommittedChanges()
	require.Len(t, changes, 1, "displacement is atomic: a single event")
	evt, ok := changes[0].(events.RealizationRoleAssigned)
	require.True(t, ok)
	assert.Equal(t, seabook.Value(), evt.DisplacedComponentID, "the displacement is explicit in the event payload for history reconstruction")
	assert.Equal(t, phoenix.Value(), evt.ComponentID)
}

func TestRealizationRoles_Assign_MultipleLegacy_NoError(t *testing.T) {
	cap := newCapabilityRef(t)
	first := newComponentRef(t)
	second := newComponentRef(t)
	third := newComponentRef(t)
	rr, err := NewRealizationRoles(RealizationRolesFacts{
		CapabilityID:  cap,
		ComponentID:   first,
		RealizationID: newRealizationID(t),
		Role:          newRole(t, valueobjects.RealizationRoleLegacy),
		AssignedBy:    "a@example.com",
	})
	require.NoError(t, err)

	require.NoError(t, rr.Assign(second, newRealizationID(t), newRole(t, valueobjects.RealizationRoleLegacy), "a@example.com"))
	require.NoError(t, rr.Assign(third, newRealizationID(t), newRole(t, valueobjects.RealizationRoleLegacy), "a@example.com"))

	for _, component := range []valueobjects.ApplicationRef{first, second, third} {
		role, ok := rr.RoleFor(component)
		require.True(t, ok)
		assert.Equal(t, valueobjects.RealizationRoleLegacy, role.Value())
	}
}

func TestRealizationRoles_Assign_SameComponentAgain_ReplacesInPlaceWithoutDisplacement(t *testing.T) {
	cases := []struct {
		name        string
		initialRole string
		newRole     string
	}{
		{
			name:        "legacy over own standard demotes in place",
			initialRole: valueobjects.RealizationRoleStandard,
			newRole:     valueobjects.RealizationRoleLegacy,
		},
		{
			name:        "same standard again is a valid re-affirmation",
			initialRole: valueobjects.RealizationRoleStandard,
			newRole:     valueobjects.RealizationRoleStandard,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cap := newCapabilityRef(t)
			component := newComponentRef(t)
			rr, err := NewRealizationRoles(RealizationRolesFacts{
				CapabilityID:  cap,
				ComponentID:   component,
				RealizationID: newRealizationID(t),
				Role:          newRole(t, tc.initialRole),
				AssignedBy:    "a@example.com",
			})
			require.NoError(t, err)
			rr.MarkChangesAsCommitted()

			err = rr.Assign(component, newRealizationID(t), newRole(t, tc.newRole), "b@example.com")

			require.NoError(t, err)
			role, ok := rr.RoleFor(component)
			require.True(t, ok)
			assert.Equal(t, tc.newRole, role.Value())

			changes := rr.GetUncommittedChanges()
			require.Len(t, changes, 1)
			evt, ok := changes[0].(events.RealizationRoleAssigned)
			require.True(t, ok)
			assert.Empty(t, evt.DisplacedComponentID, "re-assigning the same component displaces nobody")
			assert.Equal(t, "b@example.com", evt.AssignedBy)
		})
	}
}

func TestRealizationRoles_Clear_RemovesRole(t *testing.T) {
	cap := newCapabilityRef(t)
	component := newComponentRef(t)
	rr, err := NewRealizationRoles(RealizationRolesFacts{
		CapabilityID:  cap,
		ComponentID:   component,
		RealizationID: newRealizationID(t),
		Role:          newRole(t, valueobjects.RealizationRoleLegacy),
		AssignedBy:    "a@example.com",
	})
	require.NoError(t, err)
	rr.MarkChangesAsCommitted()

	err = rr.Clear(component, "a@example.com")

	require.NoError(t, err)
	_, ok := rr.RoleFor(component)
	assert.False(t, ok)

	changes := rr.GetUncommittedChanges()
	require.Len(t, changes, 1)
	evt, ok := changes[0].(events.RealizationRoleCleared)
	require.True(t, ok)
	assert.Equal(t, component.Value(), evt.ComponentID)
	assert.Equal(t, "a@example.com", evt.ClearedBy)
	assert.False(t, evt.OccurredOn.IsZero())
}

func TestRealizationRoles_Clear_NoRoleForComponent_Fails(t *testing.T) {
	cap := newCapabilityRef(t)
	holder := newComponentRef(t)
	rr, err := NewRealizationRoles(RealizationRolesFacts{
		CapabilityID:  cap,
		ComponentID:   holder,
		RealizationID: newRealizationID(t),
		Role:          newRole(t, valueobjects.RealizationRoleLegacy),
		AssignedBy:    "a@example.com",
	})
	require.NoError(t, err)

	err = rr.Clear(newComponentRef(t), "a@example.com")

	assert.ErrorIs(t, err, ErrNoRoleToClear)
}

func TestLoadRealizationRolesFromHistory_RehydratesCurrentState(t *testing.T) {
	cap := newCapabilityRef(t)
	seabook := newComponentRef(t)
	phoenix := newComponentRef(t)
	rr, err := NewRealizationRoles(RealizationRolesFacts{
		CapabilityID:  cap,
		ComponentID:   seabook,
		RealizationID: newRealizationID(t),
		Role:          newRole(t, valueobjects.RealizationRoleStandard),
		AssignedBy:    "a@example.com",
	})
	require.NoError(t, err)
	require.NoError(t, rr.Assign(phoenix, newRealizationID(t), newRole(t, valueobjects.RealizationRoleStandard), "b@example.com"))
	history := rr.GetUncommittedChanges()

	loaded, err := LoadRealizationRolesFromHistory(history)

	require.NoError(t, err)
	assert.Equal(t, rr.ID(), loaded.ID())
	_, seabookHasRole := loaded.RoleFor(seabook)
	assert.False(t, seabookHasRole)
	phoenixRole, ok := loaded.RoleFor(phoenix)
	require.True(t, ok)
	assert.Equal(t, valueobjects.RealizationRoleStandard, phoenixRole.Value())
	assert.Empty(t, loaded.GetUncommittedChanges())
}

func TestLoadRealizationRolesFromHistory_RehydratesClearedRole(t *testing.T) {
	cap := newCapabilityRef(t)
	component := newComponentRef(t)
	rr, err := NewRealizationRoles(RealizationRolesFacts{
		CapabilityID:  cap,
		ComponentID:   component,
		RealizationID: newRealizationID(t),
		Role:          newRole(t, valueobjects.RealizationRoleLegacy),
		AssignedBy:    "a@example.com",
	})
	require.NoError(t, err)
	require.NoError(t, rr.Clear(component, "a@example.com"))
	history := rr.GetUncommittedChanges()

	loaded, err := LoadRealizationRolesFromHistory(history)

	require.NoError(t, err)
	_, ok := loaded.RoleFor(component)
	assert.False(t, ok)
}

func TestApplyRealizationRoles_UnknownEvent_Fails(t *testing.T) {
	_, err := LoadRealizationRolesFromHistory([]domain.DomainEvent{unknownRealizationRolesEventForTest{}})
	assert.ErrorIs(t, err, ErrUnknownRealizationRolesEvent)
}

type unknownRealizationRolesEventForTest struct{}

func (unknownRealizationRolesEventForTest) AggregateID() string               { return "" }
func (unknownRealizationRolesEventForTest) EventType() string                 { return "UnknownEvent" }
func (unknownRealizationRolesEventForTest) EventData() map[string]interface{} { return nil }
func (unknownRealizationRolesEventForTest) OccurredAt() time.Time             { return time.Time{} }
