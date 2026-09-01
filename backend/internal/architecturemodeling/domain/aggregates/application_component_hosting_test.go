package aggregates

import (
	"testing"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func hosting(t *testing.T, value string) valueobjects.HostingClassification {
	t.Helper()
	classification, err := valueobjects.NewHostingClassification(value)
	require.NoError(t, err)
	return classification
}

func TestApplicationComponent_StartsWithUnknownHosting(t *testing.T) {
	component := newTestComponent(t)

	assert.Equal(t, valueobjects.HostingUnknown, component.Hosting().String())
}

func TestApplicationComponent_ClassifyHosting(t *testing.T) {
	component := newTestComponent(t)

	require.NoError(t, component.ClassifyHosting(hosting(t, valueobjects.HostingSaaS)))

	assert.Equal(t, valueobjects.HostingSaaS, component.Hosting().String())
	changes := component.GetUncommittedChanges()
	require.Len(t, changes, 1)
	assert.Equal(t, "ApplicationHostingClassified", changes[0].EventType())
}

func TestApplicationComponent_ReclassifyHostingIsUnrestricted(t *testing.T) {
	component := newTestComponent(t)
	require.NoError(t, component.ClassifyHosting(hosting(t, valueobjects.HostingSaaS)))
	component.MarkChangesAsCommitted()

	require.NoError(t, component.ClassifyHosting(hosting(t, valueobjects.HostingOnPremises)))

	assert.Equal(t, valueobjects.HostingOnPremises, component.Hosting().String())
	changes := component.GetUncommittedChanges()
	require.Len(t, changes, 1)
	assert.Equal(t, "ApplicationHostingClassified", changes[0].EventType())
}

func TestLoadApplicationComponentFromHistory_AppliesHosting(t *testing.T) {
	component := newTestComponent(t)
	require.NoError(t, component.ClassifyHosting(hosting(t, valueobjects.HostingCloud)))

	history := []domain.DomainEvent{
		events.NewApplicationComponentCreated(component.ID(), "Billing Service", "Handles invoicing"),
		events.NewApplicationHostingClassified(component.ID(), valueobjects.HostingCloud),
	}

	loaded, err := LoadApplicationComponentFromHistory(history)
	require.NoError(t, err)
	assert.Equal(t, valueobjects.HostingCloud, loaded.Hosting().String())
}

func TestLoadApplicationComponentFromHistory_RejectsCorruptedHosting(t *testing.T) {
	corrupted := events.ApplicationHostingClassified{
		ComponentID: "component-1",
		Hosting:     "mainframe",
	}
	history := []domain.DomainEvent{
		events.NewApplicationComponentCreated("component-1", "Billing Service", "Handles invoicing"),
		corrupted,
	}

	_, err := LoadApplicationComponentFromHistory(history)
	assert.ErrorIs(t, err, domain.ErrCorruptedEvent)
}
