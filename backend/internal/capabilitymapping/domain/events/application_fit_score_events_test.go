package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewApplicationFitScoreSet(t *testing.T) {
	params := ApplicationFitScoreSetParams{
		ID:          "score-123",
		ComponentID: "comp-456",
		PillarID:    "pillar-789",
		PillarName:  "Innovation",
		Score:       4,
		Rationale:   "Good alignment with innovation goals",
		ScoredBy:    "user@example.com",
	}

	event := NewApplicationFitScoreSet(params)

	assert.Equal(t, params.ID, event.ID)
	assert.Equal(t, params.ComponentID, event.ComponentID)
	assert.Equal(t, params.PillarID, event.PillarID)
	assert.Equal(t, params.PillarName, event.PillarName)
	assert.Equal(t, params.Score, event.Score)
	assert.Equal(t, params.Rationale, event.Rationale)
	assert.Equal(t, params.ScoredBy, event.ScoredBy)
	assert.NotZero(t, event.ScoredAt)
}

func TestApplicationFitScoreSet_EventType(t *testing.T) {
	event := NewApplicationFitScoreSet(ApplicationFitScoreSetParams{ID: "test"})

	assert.Equal(t, "ApplicationFitScoreSet", event.EventType())
}

func TestApplicationFitScoreSet_EventData(t *testing.T) {
	params := ApplicationFitScoreSetParams{
		ID:          "score-123",
		ComponentID: "comp-456",
		PillarID:    "pillar-789",
		PillarName:  "Innovation",
		Score:       4,
		Rationale:   "Good alignment",
		ScoredBy:    "user@example.com",
	}
	event := NewApplicationFitScoreSet(params)
	data := event.EventData()

	assert.Equal(t, "score-123", data["id"])
	assert.Equal(t, "comp-456", data["componentId"])
	assert.Equal(t, "pillar-789", data["pillarId"])
	assert.Equal(t, "Innovation", data["pillarName"])
	assert.Equal(t, 4, data["score"])
	assert.Equal(t, "Good alignment", data["rationale"])
	assert.Equal(t, "user@example.com", data["scoredBy"])
	assert.NotNil(t, data["scoredAt"])
}

func TestNewApplicationFitScoreUpdated(t *testing.T) {
	params := ApplicationFitScoreUpdatedParams{
		ID:           "score-123",
		Score:        5,
		Rationale:    "Updated rationale",
		OldScore:     3,
		OldRationale: "Original rationale",
		UpdatedBy:    "user@example.com",
	}

	event := NewApplicationFitScoreUpdated(params)

	assert.Equal(t, params.ID, event.ID)
	assert.Equal(t, params.Score, event.Score)
	assert.Equal(t, params.Rationale, event.Rationale)
	assert.Equal(t, params.OldScore, event.OldScore)
	assert.Equal(t, params.OldRationale, event.OldRationale)
	assert.Equal(t, params.UpdatedBy, event.UpdatedBy)
	assert.NotZero(t, event.UpdatedAt)
}

func TestApplicationFitScoreUpdated_EventType(t *testing.T) {
	event := NewApplicationFitScoreUpdated(ApplicationFitScoreUpdatedParams{ID: "test"})

	assert.Equal(t, "ApplicationFitScoreUpdated", event.EventType())
}

func TestApplicationFitScoreUpdated_EventData(t *testing.T) {
	params := ApplicationFitScoreUpdatedParams{
		ID:           "score-123",
		Score:        5,
		Rationale:    "Updated",
		OldScore:     3,
		OldRationale: "Original",
		UpdatedBy:    "user@example.com",
	}
	event := NewApplicationFitScoreUpdated(params)
	data := event.EventData()

	assert.Equal(t, "score-123", data["id"])
	assert.Equal(t, 5, data["score"])
	assert.Equal(t, "Updated", data["rationale"])
	assert.Equal(t, 3, data["oldScore"])
	assert.Equal(t, "Original", data["oldRationale"])
	assert.Equal(t, "user@example.com", data["updatedBy"])
	assert.NotNil(t, data["updatedAt"])
}

func TestNewApplicationFitScoreRemoved(t *testing.T) {
	id := "score-123"
	componentID := "comp-456"
	pillarID := "pillar-789"
	removedBy := "user@example.com"

	event := NewApplicationFitScoreRemoved(id, componentID, pillarID, removedBy)

	assert.Equal(t, id, event.ID)
	assert.Equal(t, componentID, event.ComponentID)
	assert.Equal(t, pillarID, event.PillarID)
	assert.Equal(t, removedBy, event.RemovedBy)
	assert.NotZero(t, event.RemovedAt)
}

func TestApplicationFitScoreRemoved_EventType(t *testing.T) {
	event := NewApplicationFitScoreRemoved("id", "comp", "pillar", "user")

	assert.Equal(t, "ApplicationFitScoreRemoved", event.EventType())
}

func TestApplicationFitScoreRemoved_EventData(t *testing.T) {
	event := NewApplicationFitScoreRemoved("score-123", "comp-456", "pillar-789", "user@example.com")
	data := event.EventData()

	assert.Equal(t, "score-123", data["id"])
	assert.Equal(t, "comp-456", data["componentId"])
	assert.Equal(t, "pillar-789", data["pillarId"])
	assert.Equal(t, "user@example.com", data["removedBy"])
	assert.NotNil(t, data["removedAt"])
}
