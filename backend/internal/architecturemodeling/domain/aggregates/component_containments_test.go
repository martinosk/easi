package aggregates

import (
	"testing"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func containmentKind(t *testing.T, value string) valueobjects.ContainmentKind {
	t.Helper()
	kind, err := valueobjects.NewContainmentKind(value)
	require.NoError(t, err)
	return kind
}

func attachedContainments(t *testing.T, part, parent valueobjects.ComponentID, kind string) *ComponentContainments {
	t.Helper()
	containments := NewComponentContainments()
	require.NoError(t, containments.Attach(part, parent, containmentKind(t, kind)))
	containments.MarkChangesAsCommitted()
	return containments
}

func TestComponentContainments_AttachRecordsPartParentAndKind(t *testing.T) {
	for _, kind := range []string{valueobjects.ContainmentComposition, valueobjects.ContainmentAggregation} {
		t.Run(kind, func(t *testing.T) {
			part, parent := valueobjects.NewComponentID(), valueobjects.NewComponentID()
			containments := NewComponentContainments()

			require.NoError(t, containments.Attach(part, parent, containmentKind(t, kind)))

			containment, ok := containments.ContainmentOf(part)
			require.True(t, ok)
			assert.True(t, containment.Parent.Equals(parent))
			assert.Equal(t, kind, containment.Kind.String())
			assert.True(t, containments.HasParts(parent))
			changes := containments.GetUncommittedChanges()
			require.Len(t, changes, 1)
			attached := changes[0].(events.ComponentAttached)
			assert.Equal(t, containments.ID(), attached.ContainmentsID)
			assert.Equal(t, part.Value(), attached.PartID)
			assert.Equal(t, parent.Value(), attached.ParentID)
			assert.Equal(t, kind, attached.Kind)
		})
	}
}

func TestComponentContainments_AttachRejectsRuleViolations(t *testing.T) {
	quoting, crm, erp, billing := valueobjects.NewComponentID(), valueobjects.NewComponentID(), valueobjects.NewComponentID(), valueobjects.NewComponentID()
	cases := []struct {
		name         string
		part, parent valueobjects.ComponentID
		expected     error
	}{
		{"a part has exactly one parent", quoting, erp, ErrPartAlreadyAttached},
		{"a part cannot accept parts", billing, quoting, ErrParentIsPart},
		{"a parent cannot become a part", crm, erp, ErrPartHasParts},
		{"a component cannot contain itself", erp, erp, ErrSelfContainment},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			containments := attachedContainments(t, quoting, crm, valueobjects.ContainmentComposition)

			err := containments.Attach(tc.part, tc.parent, containmentKind(t, valueobjects.ContainmentAggregation))

			assert.ErrorIs(t, err, tc.expected)
			assert.Empty(t, containments.GetUncommittedChanges())
		})
	}
}

func TestComponentContainments_ParentMayAcceptSeveralParts(t *testing.T) {
	quoting, billing, crm := valueobjects.NewComponentID(), valueobjects.NewComponentID(), valueobjects.NewComponentID()
	containments := attachedContainments(t, quoting, crm, valueobjects.ContainmentComposition)

	require.NoError(t, containments.Attach(billing, crm, containmentKind(t, valueobjects.ContainmentAggregation)))

	assert.True(t, containments.IsPart(quoting))
	assert.True(t, containments.IsPart(billing))
}

func TestComponentContainments_DetachReleasesPart(t *testing.T) {
	quoting, crm := valueobjects.NewComponentID(), valueobjects.NewComponentID()
	containments := attachedContainments(t, quoting, crm, valueobjects.ContainmentComposition)

	require.NoError(t, containments.Detach(quoting))

	assert.False(t, containments.IsPart(quoting))
	assert.False(t, containments.HasParts(crm))
	changes := containments.GetUncommittedChanges()
	require.Len(t, changes, 1)
	detached := changes[0].(events.ComponentDetached)
	assert.Equal(t, quoting.Value(), detached.PartID)
	assert.Equal(t, crm.Value(), detached.ParentID)
}

func TestComponentContainments_DetachedPartCanAttachElsewhere(t *testing.T) {
	quoting, crm, erp := valueobjects.NewComponentID(), valueobjects.NewComponentID(), valueobjects.NewComponentID()
	containments := attachedContainments(t, quoting, crm, valueobjects.ContainmentComposition)
	require.NoError(t, containments.Detach(quoting))

	require.NoError(t, containments.Attach(quoting, erp, containmentKind(t, valueobjects.ContainmentAggregation)))

	containment, ok := containments.ContainmentOf(quoting)
	require.True(t, ok)
	assert.True(t, containment.Parent.Equals(erp))
}

func TestComponentContainments_DetachRejectsStandaloneComponent(t *testing.T) {
	containments := NewComponentContainments()

	err := containments.Detach(valueobjects.NewComponentID())

	assert.ErrorIs(t, err, ErrNotAPart)
	assert.Empty(t, containments.GetUncommittedChanges())
}

func TestLoadComponentContainmentsFromHistory_ReplaysAttachAndDetach(t *testing.T) {
	quoting, billing, crm := valueobjects.NewComponentID(), valueobjects.NewComponentID(), valueobjects.NewComponentID()
	containmentsID := valueobjects.NewComponentID().Value()
	history := []domain.DomainEvent{
		events.NewComponentAttached(containmentsID, quoting.Value(), crm.Value(), valueobjects.ContainmentComposition),
		events.NewComponentAttached(containmentsID, billing.Value(), crm.Value(), valueobjects.ContainmentAggregation),
		events.NewComponentDetached(containmentsID, quoting.Value(), crm.Value()),
	}

	loaded, err := LoadComponentContainmentsFromHistory(history)

	require.NoError(t, err)
	assert.Equal(t, containmentsID, loaded.ID())
	assert.Equal(t, 3, loaded.Version())
	assert.False(t, loaded.IsPart(quoting))
	containment, ok := loaded.ContainmentOf(billing)
	require.True(t, ok)
	assert.Equal(t, valueobjects.ContainmentAggregation, containment.Kind.String())
}

func TestLoadComponentContainmentsFromHistory_RejectsCorruptedAndUnknownEvents(t *testing.T) {
	containmentsID := valueobjects.NewComponentID().Value()
	cases := []struct {
		name     string
		event    domain.DomainEvent
		expected error
	}{
		{"unknown event", events.NewApplicationComponentCreated("c1", "Billing", ""), ErrUnknownComponentContainmentsEvent},
		{"invalid kind", events.NewComponentAttached(containmentsID, valueobjects.NewComponentID().Value(), valueobjects.NewComponentID().Value(), "nesting"), domain.ErrCorruptedEvent},
		{"invalid part id", events.NewComponentAttached(containmentsID, "not-a-uuid", valueobjects.NewComponentID().Value(), valueobjects.ContainmentComposition), domain.ErrCorruptedEvent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadComponentContainmentsFromHistory([]domain.DomainEvent{tc.event})

			assert.ErrorIs(t, err, tc.expected)
		})
	}
}
