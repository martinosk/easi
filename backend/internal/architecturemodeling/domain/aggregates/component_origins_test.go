package aggregates

import (
	"testing"
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewComponentOrigins(t *testing.T) {
	componentID := valueobjects.NewComponentID()

	origins, err := NewComponentOrigins(componentID)

	require.NoError(t, err)
	assert.NotNil(t, origins)
	assert.Equal(t, "component-origins:"+componentID.String(), origins.ID())
	assert.True(t, origins.AcquiredVia().IsEmpty())
	assert.True(t, origins.PurchasedFrom().IsEmpty())
	assert.True(t, origins.BuiltBy().IsEmpty())
	assert.NotZero(t, origins.CreatedAt())
	assert.False(t, origins.IsDeleted())
}

func TestComponentOrigins_SetAcquiredVia_FirstTime(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	entityID := valueobjects.NewAcquiredEntityID()
	notes, _ := valueobjects.NewNotes("Acquired in 2021 merger")

	err := origins.SetAcquiredVia(entityID, notes)

	require.NoError(t, err)
	assert.False(t, origins.AcquiredVia().IsEmpty())
	assert.Equal(t, entityID.String(), origins.AcquiredVia().EntityID())
	assert.Equal(t, notes, origins.AcquiredVia().Notes())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 2)
	assert.Equal(t, "ComponentOriginsCreated", events[0].EventType())
	assert.Equal(t, "AcquiredViaRelationshipSet", events[1].EventType())
}

func TestComponentOrigins_SetAcquiredVia_Replace(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	entityID1 := valueobjects.NewAcquiredEntityID()
	notes1, _ := valueobjects.NewNotes("Original entity")
	origins.SetAcquiredVia(entityID1, notes1)
	origins.MarkChangesAsCommitted()

	entityID2 := valueobjects.NewAcquiredEntityID()
	notes2, _ := valueobjects.NewNotes("Corrected origin")
	err := origins.SetAcquiredVia(entityID2, notes2)

	require.NoError(t, err)
	assert.Equal(t, entityID2.String(), origins.AcquiredVia().EntityID())
	assert.Equal(t, notes2, origins.AcquiredVia().Notes())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 1)
	assert.Equal(t, "AcquiredViaRelationshipReplaced", events[0].EventType())
}

func TestComponentOrigins_SetAcquiredVia_Idempotent(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	entityID := valueobjects.NewAcquiredEntityID()
	notes, _ := valueobjects.NewNotes("Acquired 2021")
	origins.SetAcquiredVia(entityID, notes)
	origins.MarkChangesAsCommitted()

	err := origins.SetAcquiredVia(entityID, notes)

	require.NoError(t, err)
	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 0)
}

func TestComponentOrigins_SetAcquiredVia_UpdateNotesOnly(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	entityID := valueobjects.NewAcquiredEntityID()
	notes1, _ := valueobjects.NewNotes("Acquired 2021")
	origins.SetAcquiredVia(entityID, notes1)
	origins.MarkChangesAsCommitted()

	notes2, _ := valueobjects.NewNotes("Acquired in Q1 2021 merger")
	err := origins.SetAcquiredVia(entityID, notes2)

	require.NoError(t, err)
	assert.Equal(t, notes2, origins.AcquiredVia().Notes())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 1)
	assert.Equal(t, "AcquiredViaNotesUpdated", events[0].EventType())
}

func TestComponentOrigins_ClearAcquiredVia(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	entityID := valueobjects.NewAcquiredEntityID()
	notes, _ := valueobjects.NewNotes("Test notes")
	origins.SetAcquiredVia(entityID, notes)
	origins.MarkChangesAsCommitted()

	err := origins.ClearAcquiredVia()

	require.NoError(t, err)
	assert.True(t, origins.AcquiredVia().IsEmpty())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 1)
	assert.Equal(t, "AcquiredViaRelationshipCleared", events[0].EventType())
}

func TestComponentOrigins_ClearAcquiredVia_WhenNotExists(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	err := origins.ClearAcquiredVia()

	assert.Error(t, err)
	assert.Equal(t, ErrNoAcquiredViaRelationship, err)
}

func TestComponentOrigins_SetPurchasedFrom_FirstTime(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	vendorID := valueobjects.NewVendorID()
	notes, _ := valueobjects.NewNotes("Purchased in 2022")

	err := origins.SetPurchasedFrom(vendorID, notes)

	require.NoError(t, err)
	assert.False(t, origins.PurchasedFrom().IsEmpty())
	assert.Equal(t, vendorID.String(), origins.PurchasedFrom().EntityID())
	assert.Equal(t, notes, origins.PurchasedFrom().Notes())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 2)
	assert.Equal(t, "ComponentOriginsCreated", events[0].EventType())
	assert.Equal(t, "PurchasedFromRelationshipSet", events[1].EventType())
}

func TestComponentOrigins_SetPurchasedFrom_Replace(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	vendorID1 := valueobjects.NewVendorID()
	notes1, _ := valueobjects.NewNotes("Original vendor")
	origins.SetPurchasedFrom(vendorID1, notes1)
	origins.MarkChangesAsCommitted()

	vendorID2 := valueobjects.NewVendorID()
	notes2, _ := valueobjects.NewNotes("Corrected vendor")
	err := origins.SetPurchasedFrom(vendorID2, notes2)

	require.NoError(t, err)
	assert.Equal(t, vendorID2.String(), origins.PurchasedFrom().EntityID())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 1)
	assert.Equal(t, "PurchasedFromRelationshipReplaced", events[0].EventType())
}

func TestComponentOrigins_SetPurchasedFrom_UpdateNotesOnly(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	vendorID := valueobjects.NewVendorID()
	notes1, _ := valueobjects.NewNotes("Original notes")
	origins.SetPurchasedFrom(vendorID, notes1)
	origins.MarkChangesAsCommitted()

	notes2, _ := valueobjects.NewNotes("Updated notes")
	err := origins.SetPurchasedFrom(vendorID, notes2)

	require.NoError(t, err)
	assert.Equal(t, notes2, origins.PurchasedFrom().Notes())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 1)
	assert.Equal(t, "PurchasedFromNotesUpdated", events[0].EventType())
}

func TestComponentOrigins_ClearPurchasedFrom(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	vendorID := valueobjects.NewVendorID()
	notes, _ := valueobjects.NewNotes("Test notes")
	origins.SetPurchasedFrom(vendorID, notes)
	origins.MarkChangesAsCommitted()

	err := origins.ClearPurchasedFrom()

	require.NoError(t, err)
	assert.True(t, origins.PurchasedFrom().IsEmpty())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 1)
	assert.Equal(t, "PurchasedFromRelationshipCleared", events[0].EventType())
}

func TestComponentOrigins_ClearPurchasedFrom_WhenNotExists(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	err := origins.ClearPurchasedFrom()

	assert.Error(t, err)
	assert.Equal(t, ErrNoPurchasedFromRelationship, err)
}

func TestComponentOrigins_SetBuiltBy_FirstTime(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	teamID := valueobjects.NewInternalTeamID()
	notes, _ := valueobjects.NewNotes("Built by Platform Team")

	err := origins.SetBuiltBy(teamID, notes)

	require.NoError(t, err)
	assert.False(t, origins.BuiltBy().IsEmpty())
	assert.Equal(t, teamID.String(), origins.BuiltBy().EntityID())
	assert.Equal(t, notes, origins.BuiltBy().Notes())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 2)
	assert.Equal(t, "ComponentOriginsCreated", events[0].EventType())
	assert.Equal(t, "BuiltByRelationshipSet", events[1].EventType())
}

func TestComponentOrigins_SetBuiltBy_Replace(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	teamID1 := valueobjects.NewInternalTeamID()
	notes1, _ := valueobjects.NewNotes("Original team")
	origins.SetBuiltBy(teamID1, notes1)
	origins.MarkChangesAsCommitted()

	teamID2 := valueobjects.NewInternalTeamID()
	notes2, _ := valueobjects.NewNotes("Corrected team")
	err := origins.SetBuiltBy(teamID2, notes2)

	require.NoError(t, err)
	assert.Equal(t, teamID2.String(), origins.BuiltBy().EntityID())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 1)
	assert.Equal(t, "BuiltByRelationshipReplaced", events[0].EventType())
}

func TestComponentOrigins_SetBuiltBy_UpdateNotesOnly(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	teamID := valueobjects.NewInternalTeamID()
	notes1, _ := valueobjects.NewNotes("Original notes")
	origins.SetBuiltBy(teamID, notes1)
	origins.MarkChangesAsCommitted()

	notes2, _ := valueobjects.NewNotes("Updated notes")
	err := origins.SetBuiltBy(teamID, notes2)

	require.NoError(t, err)
	assert.Equal(t, notes2, origins.BuiltBy().Notes())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 1)
	assert.Equal(t, "BuiltByNotesUpdated", events[0].EventType())
}

func TestComponentOrigins_ClearBuiltBy(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	teamID := valueobjects.NewInternalTeamID()
	notes, _ := valueobjects.NewNotes("Test notes")
	origins.SetBuiltBy(teamID, notes)
	origins.MarkChangesAsCommitted()

	err := origins.ClearBuiltBy()

	require.NoError(t, err)
	assert.True(t, origins.BuiltBy().IsEmpty())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 1)
	assert.Equal(t, "BuiltByRelationshipCleared", events[0].EventType())
}

func TestComponentOrigins_ClearBuiltBy_WhenNotExists(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	err := origins.ClearBuiltBy()

	assert.Error(t, err)
	assert.Equal(t, ErrNoBuiltByRelationship, err)
}

func TestComponentOrigins_MultipleOriginTypes(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	entityID := valueobjects.NewAcquiredEntityID()
	notes1, _ := valueobjects.NewNotes("Acquired")
	origins.SetAcquiredVia(entityID, notes1)

	teamID := valueobjects.NewInternalTeamID()
	notes2, _ := valueobjects.NewNotes("Built by team")
	origins.SetBuiltBy(teamID, notes2)

	assert.False(t, origins.AcquiredVia().IsEmpty())
	assert.True(t, origins.PurchasedFrom().IsEmpty())
	assert.False(t, origins.BuiltBy().IsEmpty())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 3)
	assert.Equal(t, "ComponentOriginsCreated", events[0].EventType())
	assert.Equal(t, "AcquiredViaRelationshipSet", events[1].EventType())
	assert.Equal(t, "BuiltByRelationshipSet", events[2].EventType())
}

func TestComponentOrigins_Delete(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	entityID := valueobjects.NewAcquiredEntityID()
	notes1, _ := valueobjects.NewNotes("Acquired")
	origins.SetAcquiredVia(entityID, notes1)

	teamID := valueobjects.NewInternalTeamID()
	notes2, _ := valueobjects.NewNotes("Built by team")
	origins.SetBuiltBy(teamID, notes2)

	origins.MarkChangesAsCommitted()

	err := origins.Delete()

	require.NoError(t, err)
	assert.True(t, origins.IsDeleted())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 1)
	assert.Equal(t, "ComponentOriginsDeleted", events[0].EventType())
}

func TestComponentOrigins_ReplacingWithPreviouslyLinkedEntity(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)

	entityID1 := valueobjects.NewAcquiredEntityID()
	notes1, _ := valueobjects.NewNotes("First link")
	origins.SetAcquiredVia(entityID1, notes1)
	origins.ClearAcquiredVia()

	entityID2 := valueobjects.NewAcquiredEntityID()
	notes2, _ := valueobjects.NewNotes("Second link")
	origins.SetAcquiredVia(entityID2, notes2)
	origins.MarkChangesAsCommitted()

	notes3, _ := valueobjects.NewNotes("Back to first")
	err := origins.SetAcquiredVia(entityID1, notes3)

	require.NoError(t, err)
	assert.Equal(t, entityID1.String(), origins.AcquiredVia().EntityID())

	events := origins.GetUncommittedChanges()
	assert.Len(t, events, 1)
	assert.Equal(t, "AcquiredViaRelationshipReplaced", events[0].EventType())
}

func TestComponentOrigins_LoadFromHistory(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	aggregateID := "component-origins:" + componentID.String()
	entityID := valueobjects.NewAcquiredEntityID()
	notes, _ := valueobjects.NewNotes("Test notes")
	linkedAt := time.Now()

	createdEvent := events.NewComponentOriginsCreatedEvent(aggregateID, componentID, linkedAt)
	setEvent := events.NewAcquiredViaRelationshipSetEvent(aggregateID, componentID, entityID, notes, linkedAt)

	origins := &ComponentOrigins{}
	origins.LoadFromHistory([]domain.DomainEvent{createdEvent, setEvent}, func(event domain.DomainEvent) {
		origins.apply(event)
	})

	assert.Equal(t, aggregateID, origins.ID())
	assert.False(t, origins.AcquiredVia().IsEmpty())
	assert.Equal(t, entityID.String(), origins.AcquiredVia().EntityID())
	assert.Equal(t, 2, origins.Version())
}

// Regression tests for the duplicate key violation bug fix
// These tests ensure ComponentOrigins uses a namespaced aggregate ID
// to prevent collisions with ApplicationComponent events in the event store

func TestComponentOrigins_AggregateIDIsNamespaced(t *testing.T) {
	componentID, err := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440001")
	require.NoError(t, err)

	origins, err := NewComponentOrigins(componentID)

	require.NoError(t, err)
	expectedAggregateID := "component-origins:550e8400-e29b-41d4-a716-446655440001"
	assert.Equal(t, expectedAggregateID, origins.ID(),
		"ComponentOrigins must use namespaced aggregate ID to avoid event store collisions")
}

func TestComponentOrigins_AllEventsContainNamespacedAggregateID(t *testing.T) {
	componentID, err := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440002")
	require.NoError(t, err)
	entityID := valueobjects.NewAcquiredEntityID()
	notes, err := valueobjects.NewNotes("Test notes")
	require.NoError(t, err)

	origins, err := NewComponentOrigins(componentID)
	require.NoError(t, err)
	err = origins.SetAcquiredVia(entityID, notes)
	require.NoError(t, err)

	expectedAggregateID := "component-origins:550e8400-e29b-41d4-a716-446655440002"
	events := origins.GetUncommittedChanges()

	for i, event := range events {
		assert.Equal(t, expectedAggregateID, event.AggregateID(),
			"Event %d must have namespaced aggregate ID to prevent event store collisions", i)
	}
}

func TestComponentOrigins_DoesNotCollideWithComponentAggregate(t *testing.T) {
	// This test ensures ComponentOrigins and ApplicationComponent
	// can coexist without event store conflicts
	componentID, err := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440003")
	require.NoError(t, err)

	origins, err := NewComponentOrigins(componentID)
	require.NoError(t, err)

	// ComponentOrigins should use "component-origins:550e8400-e29b-41d4-a716-446655440003"
	// NOT just "550e8400-e29b-41d4-a716-446655440003" which would collide with ApplicationComponent
	assert.NotEqual(t, componentID.String(), origins.ID(),
		"ComponentOrigins MUST NOT use bare component ID - will cause event store duplicate key violations!")

	assert.Contains(t, origins.ID(), "component-origins:",
		"ComponentOrigins aggregate ID must have namespace prefix")
}

func TestComponentOrigins_NamespaceAppliedToAllEventTypes(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	origins, _ := NewComponentOrigins(componentID)
	expectedAggregateID := "component-origins:" + componentID.String()

	// Test SetAcquiredVia
	entityID := valueobjects.NewAcquiredEntityID()
	notes1, _ := valueobjects.NewNotes("Acquired")
	origins.SetAcquiredVia(entityID, notes1)

	// Test SetPurchasedFrom
	vendorID := valueobjects.NewVendorID()
	notes2, _ := valueobjects.NewNotes("Purchased")
	origins.SetPurchasedFrom(vendorID, notes2)

	// Test SetBuiltBy
	teamID := valueobjects.NewInternalTeamID()
	notes3, _ := valueobjects.NewNotes("Built")
	origins.SetBuiltBy(teamID, notes3)

	// Verify all events have namespaced aggregate ID
	events := origins.GetUncommittedChanges()
	assert.Greater(t, len(events), 0, "Should have generated events")

	for _, event := range events {
		assert.Equal(t, expectedAggregateID, event.AggregateID(),
			"Event %s must have namespaced aggregate ID", event.EventType())
	}
}

func TestComponentOrigins_ClearOperationsUseNamespacedID(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	expectedAggregateID := "component-origins:" + componentID.String()

	// Setup
	origins, _ := NewComponentOrigins(componentID)
	entityID := valueobjects.NewAcquiredEntityID()
	notes, _ := valueobjects.NewNotes("Test")
	origins.SetAcquiredVia(entityID, notes)
	origins.MarkChangesAsCommitted()

	// Act
	origins.ClearAcquiredVia()

	// Assert
	events := origins.GetUncommittedChanges()
	require.Len(t, events, 1)
	assert.Equal(t, expectedAggregateID, events[0].AggregateID())
}

func TestComponentOrigins_ReplaceOperationsUseNamespacedID(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	expectedAggregateID := "component-origins:" + componentID.String()

	// Setup
	origins, _ := NewComponentOrigins(componentID)
	entityID1 := valueobjects.NewAcquiredEntityID()
	notes1, _ := valueobjects.NewNotes("First")
	origins.SetAcquiredVia(entityID1, notes1)
	origins.MarkChangesAsCommitted()

	// Act - replace with different entity
	entityID2 := valueobjects.NewAcquiredEntityID()
	notes2, _ := valueobjects.NewNotes("Second")
	origins.SetAcquiredVia(entityID2, notes2)

	// Assert
	events := origins.GetUncommittedChanges()
	require.Len(t, events, 1)
	assert.Equal(t, "AcquiredViaRelationshipReplaced", events[0].EventType())
	assert.Equal(t, expectedAggregateID, events[0].AggregateID())
}

func TestComponentOrigins_NotesUpdateOperationsUseNamespacedID(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	expectedAggregateID := "component-origins:" + componentID.String()

	// Setup
	origins, _ := NewComponentOrigins(componentID)
	entityID := valueobjects.NewAcquiredEntityID()
	notes1, _ := valueobjects.NewNotes("Original notes")
	origins.SetAcquiredVia(entityID, notes1)
	origins.MarkChangesAsCommitted()

	// Act - update notes only
	notes2, _ := valueobjects.NewNotes("Updated notes")
	origins.SetAcquiredVia(entityID, notes2)

	// Assert
	events := origins.GetUncommittedChanges()
	require.Len(t, events, 1)
	assert.Equal(t, "AcquiredViaNotesUpdated", events[0].EventType())
	assert.Equal(t, expectedAggregateID, events[0].AggregateID())
}

func TestComponentOrigins_DeleteOperationUsesNamespacedID(t *testing.T) {
	componentID := valueobjects.NewComponentID()
	expectedAggregateID := "component-origins:" + componentID.String()

	// Setup
	origins, _ := NewComponentOrigins(componentID)
	origins.MarkChangesAsCommitted()

	// Act
	origins.Delete()

	// Assert
	events := origins.GetUncommittedChanges()
	require.Len(t, events, 1)
	assert.Equal(t, "ComponentOriginsDeleted", events[0].EventType())
	assert.Equal(t, expectedAggregateID, events[0].AggregateID())
}
