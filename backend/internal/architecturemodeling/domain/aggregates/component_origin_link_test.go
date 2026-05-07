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

type originLinkTestCase struct {
	name       string
	originType string
	newID      func() string
}

func allOriginLinkCases() []originLinkTestCase {
	return []originLinkTestCase{
		{
			name:       "AcquiredVia",
			originType: valueobjects.OriginTypeAcquiredVia,
			newID:      func() string { return valueobjects.NewAcquiredEntityID().String() },
		},
		{
			name:       "PurchasedFrom",
			originType: valueobjects.OriginTypePurchasedFrom,
			newID:      func() string { return valueobjects.NewVendorID().String() },
		},
		{
			name:       "BuiltBy",
			originType: valueobjects.OriginTypeBuiltBy,
			newID:      func() string { return valueobjects.NewInternalTeamID().String() },
		},
	}
}

func forAllOriginTypes(t *testing.T, run func(t *testing.T, tc originLinkTestCase)) {
	t.Helper()
	for _, tc := range allOriginLinkCases() {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func createOriginLink(t *testing.T, tc originLinkTestCase) *ComponentOriginLink {
	t.Helper()
	componentID := valueobjects.NewComponentID()
	originType, err := valueobjects.NewOriginType(tc.originType)
	require.NoError(t, err)
	link, err := NewComponentOriginLink(componentID, originType)
	require.NoError(t, err)
	return link
}

func createOriginLinkWithEntity(t *testing.T, tc originLinkTestCase) (*ComponentOriginLink, string) {
	t.Helper()
	link := createOriginLink(t, tc)
	entityID := tc.newID()
	notes, _ := valueobjects.NewNotes("Initial")
	require.NoError(t, link.Set(entityID, notes))
	link.MarkChangesAsCommitted()
	return link, entityID
}

func TestNewComponentOriginLink(t *testing.T) {
	forAllOriginTypes(t, func(t *testing.T, tc originLinkTestCase) {
		componentID := valueobjects.NewComponentID()
		originType, err := valueobjects.NewOriginType(tc.originType)
		require.NoError(t, err)

		link, err := NewComponentOriginLink(componentID, originType)

		require.NoError(t, err)
		assert.NotNil(t, link)
		expectedID := "origin-link:" + tc.originType + ":" + componentID.String()
		assert.Equal(t, expectedID, link.ID())
		assert.True(t, link.Link().IsEmpty())
		assert.NotZero(t, link.CreatedAt())
		assert.False(t, link.IsDeleted())
	})
}

func TestComponentOriginLink_AggregateIDFormat(t *testing.T) {
	componentID, err := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440001")
	require.NoError(t, err)
	originType, _ := valueobjects.NewOriginType(valueobjects.OriginTypeAcquiredVia)

	link, err := NewComponentOriginLink(componentID, originType)

	require.NoError(t, err)
	assert.Equal(t, "origin-link:acquired-via:550e8400-e29b-41d4-a716-446655440001", link.ID())
}

func TestComponentOriginLink_SetFirstTime(t *testing.T) {
	forAllOriginTypes(t, func(t *testing.T, tc originLinkTestCase) {
		link := createOriginLink(t, tc)
		entityID := tc.newID()
		notes, _ := valueobjects.NewNotes("First time notes")

		err := link.Set(entityID, notes)

		require.NoError(t, err)
		assert.False(t, link.Link().IsEmpty())
		assert.Equal(t, entityID, link.Link().EntityID())
		assert.Equal(t, notes, link.Link().Notes())

		evts := link.GetUncommittedChanges()
		assert.Len(t, evts, 2)
		assert.Equal(t, "OriginLinkCreated", evts[0].EventType())
		assert.Equal(t, "OriginLinkSet", evts[1].EventType())
	})
}

func TestComponentOriginLink_SetAfterInitial(t *testing.T) {
	cases := []struct {
		name      string
		useNewID  bool
		notesText string
		check     func(t *testing.T, link *ComponentOriginLink, entityID string, notes valueobjects.Notes)
	}{
		{
			name:      "Replace with different entity emits replaced event",
			useNewID:  true,
			notesText: "Replacement",
			check: func(t *testing.T, link *ComponentOriginLink, entityID string, _ valueobjects.Notes) {
				assert.Equal(t, entityID, link.Link().EntityID())
				assertSingleEvent(t, link, "OriginLinkReplaced")
			},
		},
		{
			name:      "Update notes on same entity emits notes-updated event",
			useNewID:  false,
			notesText: "Updated notes",
			check: func(t *testing.T, link *ComponentOriginLink, _ string, notes valueobjects.Notes) {
				assert.Equal(t, notes, link.Link().Notes())
				assertSingleEvent(t, link, "OriginLinkNotesUpdated")
			},
		},
		{
			name:      "Same entity and notes is idempotent",
			useNewID:  false,
			notesText: "Initial",
			check: func(t *testing.T, link *ComponentOriginLink, _ string, _ valueobjects.Notes) {
				assert.Empty(t, link.GetUncommittedChanges())
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			forAllOriginTypes(t, func(t *testing.T, tc originLinkTestCase) {
				link, entityID := createOriginLinkWithEntity(t, tc)
				if c.useNewID {
					entityID = tc.newID()
				}
				notes, _ := valueobjects.NewNotes(c.notesText)

				require.NoError(t, link.Set(entityID, notes))

				c.check(t, link, entityID, notes)
			})
		})
	}
}

func TestComponentOriginLink_Clear(t *testing.T) {
	forAllOriginTypes(t, func(t *testing.T, tc originLinkTestCase) {
		link, _ := createOriginLinkWithEntity(t, tc)

		err := link.Clear()

		require.NoError(t, err)
		assert.True(t, link.Link().IsEmpty())
		assertSingleEvent(t, link, "OriginLinkCleared")
	})
}

func TestComponentOriginLink_ClearWhenEmpty(t *testing.T) {
	forAllOriginTypes(t, func(t *testing.T, tc originLinkTestCase) {
		link := createOriginLink(t, tc)

		err := link.Clear()

		assert.ErrorIs(t, err, ErrNoOriginLink)
	})
}

func TestComponentOriginLink_Delete(t *testing.T) {
	forAllOriginTypes(t, func(t *testing.T, tc originLinkTestCase) {
		link, _ := createOriginLinkWithEntity(t, tc)

		err := link.Delete()

		require.NoError(t, err)
		assert.True(t, link.IsDeleted())
		assertSingleEvent(t, link, "OriginLinkDeleted")
	})
}

func assertSingleEvent(t *testing.T, link *ComponentOriginLink, expectedEventType string) {
	t.Helper()
	evts := link.GetUncommittedChanges()
	require.Len(t, evts, 1)
	assert.Equal(t, expectedEventType, evts[0].EventType())
}

func TestComponentOriginLink_LoadFromHistory(t *testing.T) {
	forAllOriginTypes(t, func(t *testing.T, tc originLinkTestCase) {
		componentID := valueobjects.NewComponentID()
		aggregateID := BuildOriginLinkAggregateID(tc.originType, componentID.String())
		entityID := tc.newID()
		notes, _ := valueobjects.NewNotes("Test notes")
		now := time.Now().UTC()

		base := events.NewOriginLinkBase(aggregateID, componentID.String(), tc.originType)
		createdEvent := events.NewOriginLinkCreatedEvent(base, now)
		setEvent := events.NewOriginLinkSetEvent(base, entityID, notes.String(), now)

		link, err := LoadComponentOriginLinkFromHistory([]domain.DomainEvent{createdEvent, setEvent})

		require.NoError(t, err)
		assert.Equal(t, aggregateID, link.ID())
		assert.False(t, link.Link().IsEmpty())
		assert.Equal(t, entityID, link.Link().EntityID())
		assert.Equal(t, 2, link.Version())
	})
}

func TestComponentOriginLink_AllEventsContainCorrectAggregateID(t *testing.T) {
	forAllOriginTypes(t, func(t *testing.T, tc originLinkTestCase) {
		link := createOriginLink(t, tc)
		expectedID := link.ID()
		entityID := tc.newID()
		notes, _ := valueobjects.NewNotes("Test")

		_ = link.Set(entityID, notes)

		evts := link.GetUncommittedChanges()
		for _, event := range evts {
			assert.Equal(t, expectedID, event.AggregateID(),
				"Event %s must have correct aggregate ID", event.EventType())
		}
	})
}

func TestBuildOriginLinkAggregateID(t *testing.T) {
	assert.Equal(t, "origin-link:acquired-via:abc-123", BuildOriginLinkAggregateID("acquired-via", "abc-123"))
	assert.Equal(t, "origin-link:purchased-from:def-456", BuildOriginLinkAggregateID("purchased-from", "def-456"))
	assert.Equal(t, "origin-link:built-by:ghi-789", BuildOriginLinkAggregateID("built-by", "ghi-789"))
}
