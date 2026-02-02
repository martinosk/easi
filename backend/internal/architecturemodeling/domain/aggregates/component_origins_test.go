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

type originTypeTestCase struct {
	name     string
	newID    func() string
	set      func(*ComponentOrigins, string, valueobjects.Notes) error
	clear    func(*ComponentOrigins) error
	getLink  func(*ComponentOrigins) valueobjects.OriginLink
	events   [4]string
	clearErr error
}

func (tc originTypeTestCase) setEvent() string     { return tc.events[0] }
func (tc originTypeTestCase) replaceEvent() string  { return tc.events[1] }
func (tc originTypeTestCase) notesEvent() string    { return tc.events[2] }
func (tc originTypeTestCase) clearEvent() string    { return tc.events[3] }

func allOriginTypeCases() []originTypeTestCase {
	return []originTypeTestCase{
		{
			name:    "AcquiredVia",
			newID:   func() string { return valueobjects.NewAcquiredEntityID().String() },
			set: func(o *ComponentOrigins, id string, n valueobjects.Notes) error {
				eid, _ := valueobjects.NewAcquiredEntityIDFromString(id)
				return o.SetAcquiredVia(eid, n)
			},
			clear:    func(o *ComponentOrigins) error { return o.ClearAcquiredVia() },
			getLink:  func(o *ComponentOrigins) valueobjects.OriginLink { return o.AcquiredVia() },
			events:   [4]string{"AcquiredViaRelationshipSet", "AcquiredViaRelationshipReplaced", "AcquiredViaNotesUpdated", "AcquiredViaRelationshipCleared"},
			clearErr: ErrNoAcquiredViaRelationship,
		},
		{
			name:    "PurchasedFrom",
			newID:   func() string { return valueobjects.NewVendorID().String() },
			set: func(o *ComponentOrigins, id string, n valueobjects.Notes) error {
				vid, _ := valueobjects.NewVendorIDFromString(id)
				return o.SetPurchasedFrom(vid, n)
			},
			clear:    func(o *ComponentOrigins) error { return o.ClearPurchasedFrom() },
			getLink:  func(o *ComponentOrigins) valueobjects.OriginLink { return o.PurchasedFrom() },
			events:   [4]string{"PurchasedFromRelationshipSet", "PurchasedFromRelationshipReplaced", "PurchasedFromNotesUpdated", "PurchasedFromRelationshipCleared"},
			clearErr: ErrNoPurchasedFromRelationship,
		},
		{
			name:    "BuiltBy",
			newID:   func() string { return valueobjects.NewInternalTeamID().String() },
			set: func(o *ComponentOrigins, id string, n valueobjects.Notes) error {
				tid, _ := valueobjects.NewInternalTeamIDFromString(id)
				return o.SetBuiltBy(tid, n)
			},
			clear:    func(o *ComponentOrigins) error { return o.ClearBuiltBy() },
			getLink:  func(o *ComponentOrigins) valueobjects.OriginLink { return o.BuiltBy() },
			events:   [4]string{"BuiltByRelationshipSet", "BuiltByRelationshipReplaced", "BuiltByNotesUpdated", "BuiltByRelationshipCleared"},
			clearErr: ErrNoBuiltByRelationship,
		},
	}
}

func newOriginsWithRelationship(tc originTypeTestCase) (*ComponentOrigins, string) {
	origins, _ := NewComponentOrigins(valueobjects.NewComponentID())
	entityID := tc.newID()
	notes, _ := valueobjects.NewNotes("Initial")
	tc.set(origins, entityID, notes)
	origins.MarkChangesAsCommitted()
	return origins, entityID
}

func TestComponentOrigins_SetFirstTime(t *testing.T) {
	for _, tc := range allOriginTypeCases() {
		t.Run(tc.name, func(t *testing.T) {
			origins, _ := NewComponentOrigins(valueobjects.NewComponentID())
			entityID := tc.newID()
			notes, _ := valueobjects.NewNotes("First time notes")

			err := tc.set(origins, entityID, notes)

			require.NoError(t, err)
			link := tc.getLink(origins)
			assert.False(t, link.IsEmpty())
			assert.Equal(t, entityID, link.EntityID())
			assert.Equal(t, notes, link.Notes())

			evts := origins.GetUncommittedChanges()
			assert.Len(t, evts, 2)
			assert.Equal(t, "ComponentOriginsCreated", evts[0].EventType())
			assert.Equal(t, tc.setEvent(), evts[1].EventType())
		})
	}
}

func TestComponentOrigins_ExistingRelationshipOperations(t *testing.T) {
	for _, tc := range allOriginTypeCases() {
		t.Run(tc.name+"/Replace", func(t *testing.T) {
			origins, _ := newOriginsWithRelationship(tc)
			newID := tc.newID()
			notes, _ := valueobjects.NewNotes("Replacement")

			err := tc.set(origins, newID, notes)

			require.NoError(t, err)
			assert.Equal(t, newID, tc.getLink(origins).EntityID())
			evts := origins.GetUncommittedChanges()
			assert.Len(t, evts, 1)
			assert.Equal(t, tc.replaceEvent(), evts[0].EventType())
		})

		t.Run(tc.name+"/UpdateNotesOnly", func(t *testing.T) {
			origins, entityID := newOriginsWithRelationship(tc)
			notes2, _ := valueobjects.NewNotes("Updated notes")

			err := tc.set(origins, entityID, notes2)

			require.NoError(t, err)
			assert.Equal(t, notes2, tc.getLink(origins).Notes())
			evts := origins.GetUncommittedChanges()
			assert.Len(t, evts, 1)
			assert.Equal(t, tc.notesEvent(), evts[0].EventType())
		})

		t.Run(tc.name+"/Clear", func(t *testing.T) {
			origins, _ := newOriginsWithRelationship(tc)

			err := tc.clear(origins)

			require.NoError(t, err)
			assert.True(t, tc.getLink(origins).IsEmpty())
			evts := origins.GetUncommittedChanges()
			assert.Len(t, evts, 1)
			assert.Equal(t, tc.clearEvent(), evts[0].EventType())
		})

		t.Run(tc.name+"/ClearWhenNotExists", func(t *testing.T) {
			origins, _ := NewComponentOrigins(valueobjects.NewComponentID())

			err := tc.clear(origins)

			assert.Error(t, err)
			assert.Equal(t, tc.clearErr, err)
		})
	}
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

func TestComponentOrigins_MutationOperationsUseNamespacedID(t *testing.T) {
	tc := allOriginTypeCases()[0]

	tests := []struct {
		name          string
		expectedEvent string
		act           func(*ComponentOrigins, string)
	}{
		{
			name:          "Clear",
			expectedEvent: tc.clearEvent(),
			act: func(o *ComponentOrigins, _ string) {
				tc.clear(o)
			},
		},
		{
			name:          "Replace",
			expectedEvent: tc.replaceEvent(),
			act: func(o *ComponentOrigins, _ string) {
				notes2, _ := valueobjects.NewNotes("Second")
				tc.set(o, tc.newID(), notes2)
			},
		},
		{
			name:          "NotesUpdate",
			expectedEvent: tc.notesEvent(),
			act: func(o *ComponentOrigins, entityID string) {
				notes2, _ := valueobjects.NewNotes("Updated notes")
				tc.set(o, entityID, notes2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentID := valueobjects.NewComponentID()
			expectedAggregateID := "component-origins:" + componentID.String()

			origins, _ := NewComponentOrigins(componentID)
			entityID := tc.newID()
			notes, _ := valueobjects.NewNotes("Original")
			tc.set(origins, entityID, notes)
			origins.MarkChangesAsCommitted()

			tt.act(origins, entityID)

			evts := origins.GetUncommittedChanges()
			require.Len(t, evts, 1)
			assert.Equal(t, tt.expectedEvent, evts[0].EventType())
			assert.Equal(t, expectedAggregateID, evts[0].AggregateID())
		})
	}
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
