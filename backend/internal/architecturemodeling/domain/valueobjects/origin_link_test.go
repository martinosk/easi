package valueobjects

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewOriginLink_ValidData(t *testing.T) {
	entityID := "entity-123"
	notes, _ := NewNotes("Test notes")
	linkedAt := time.Now()

	link := NewOriginLink(entityID, notes, linkedAt)

	assert.Equal(t, entityID, link.EntityID())
	assert.Equal(t, notes, link.Notes())
	assert.Equal(t, linkedAt, link.LinkedAt())
	assert.False(t, link.IsEmpty())
}

func TestNewOriginLink_EmptyEntityID(t *testing.T) {
	notes, _ := NewNotes("Test notes")
	linkedAt := time.Now()

	link := NewOriginLink("", notes, linkedAt)

	assert.Equal(t, "", link.EntityID())
	assert.Equal(t, notes, link.Notes())
	assert.Equal(t, linkedAt, link.LinkedAt())
	assert.True(t, link.IsEmpty())
}

func TestEmptyOriginLink(t *testing.T) {
	link := EmptyOriginLink()

	assert.Equal(t, "", link.EntityID())
	assert.True(t, link.Notes().IsEmpty())
	assert.True(t, link.LinkedAt().IsZero())
	assert.True(t, link.IsEmpty())
}

func TestOriginLink_IsEmpty_WhenEntityIDIsEmpty(t *testing.T) {
	notes, _ := NewNotes("Test notes")
	linkedAt := time.Now()

	link := NewOriginLink("", notes, linkedAt)

	assert.True(t, link.IsEmpty())
}

func TestOriginLink_Equals_SameValues(t *testing.T) {
	entityID := "entity-123"
	notes, _ := NewNotes("Test notes")
	linkedAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	link1 := NewOriginLink(entityID, notes, linkedAt)
	link2 := NewOriginLink(entityID, notes, linkedAt)

	assert.True(t, link1.Equals(link2))
}

func TestOriginLink_Equals_DifferentEntityID(t *testing.T) {
	notes, _ := NewNotes("Test notes")
	linkedAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	link1 := NewOriginLink("entity-123", notes, linkedAt)
	link2 := NewOriginLink("entity-456", notes, linkedAt)

	assert.False(t, link1.Equals(link2))
}

func TestOriginLink_Equals_DifferentNotes(t *testing.T) {
	entityID := "entity-123"
	notes1, _ := NewNotes("Test notes 1")
	notes2, _ := NewNotes("Test notes 2")
	linkedAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	link1 := NewOriginLink(entityID, notes1, linkedAt)
	link2 := NewOriginLink(entityID, notes2, linkedAt)

	assert.False(t, link1.Equals(link2))
}

func TestOriginLink_Equals_DifferentLinkedAt(t *testing.T) {
	entityID := "entity-123"
	notes, _ := NewNotes("Test notes")
	linkedAt1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	linkedAt2 := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	link1 := NewOriginLink(entityID, notes, linkedAt1)
	link2 := NewOriginLink(entityID, notes, linkedAt2)

	assert.False(t, link1.Equals(link2))
}

func TestOriginLink_Equals_BothEmpty(t *testing.T) {
	link1 := EmptyOriginLink()
	link2 := EmptyOriginLink()

	assert.True(t, link1.Equals(link2))
}
