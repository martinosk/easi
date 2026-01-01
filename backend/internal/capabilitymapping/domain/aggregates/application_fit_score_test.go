package aggregates

import (
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetApplicationFitScore(t *testing.T) {
	componentID, _ := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440000")
	pillarID, _ := valueobjects.NewPillarIDFromString("550e8400-e29b-41d4-a716-446655440001")
	pillarName, _ := valueobjects.NewPillarName("Digital Transformation")
	score, _ := valueobjects.NewFitScore(4)
	rationale, _ := valueobjects.NewFitRationale("Strong alignment with digital strategy")
	scoredBy, _ := valueobjects.NewUserIdentifier("user@example.com")

	aggregate, err := SetApplicationFitScore(NewFitScoreParams{
		ComponentID: componentID,
		PillarID:    pillarID,
		PillarName:  pillarName,
		Score:       score,
		Rationale:   rationale,
		ScoredBy:    scoredBy,
	})
	require.NoError(t, err)
	assert.NotNil(t, aggregate)
	assert.NotEmpty(t, aggregate.ID())
	assert.Equal(t, componentID, aggregate.ComponentID())
	assert.Equal(t, pillarID, aggregate.PillarID())
	assert.Equal(t, score, aggregate.Score())
	assert.Equal(t, rationale, aggregate.Rationale())
	assert.Equal(t, scoredBy, aggregate.ScoredBy())
	assert.Len(t, aggregate.GetUncommittedChanges(), 1)
}

func TestSetApplicationFitScore_WithEmptyRationale(t *testing.T) {
	componentID, _ := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440000")
	pillarID, _ := valueobjects.NewPillarIDFromString("550e8400-e29b-41d4-a716-446655440001")
	pillarName, _ := valueobjects.NewPillarName("Customer Focus")
	score, _ := valueobjects.NewFitScore(5)
	rationale, _ := valueobjects.NewFitRationale("")
	scoredBy, _ := valueobjects.NewUserIdentifier("user@example.com")

	aggregate, err := SetApplicationFitScore(NewFitScoreParams{
		ComponentID: componentID,
		PillarID:    pillarID,
		PillarName:  pillarName,
		Score:       score,
		Rationale:   rationale,
		ScoredBy:    scoredBy,
	})
	require.NoError(t, err)
	assert.True(t, aggregate.Rationale().IsEmpty())
}

func TestSetApplicationFitScore_RaisesEvent(t *testing.T) {
	componentID, _ := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440000")
	pillarID, _ := valueobjects.NewPillarIDFromString("550e8400-e29b-41d4-a716-446655440001")
	pillarName, _ := valueobjects.NewPillarName("Innovation")
	score, _ := valueobjects.NewFitScore(3)
	rationale, _ := valueobjects.NewFitRationale("Supports innovation goals")
	scoredBy, _ := valueobjects.NewUserIdentifier("user@example.com")

	aggregate, err := SetApplicationFitScore(NewFitScoreParams{
		ComponentID: componentID,
		PillarID:    pillarID,
		PillarName:  pillarName,
		Score:       score,
		Rationale:   rationale,
		ScoredBy:    scoredBy,
	})
	require.NoError(t, err)

	events := aggregate.GetUncommittedChanges()
	require.Len(t, events, 1)
	assert.Equal(t, "ApplicationFitScoreSet", events[0].EventType())

	eventData := events[0].EventData()
	assert.Equal(t, aggregate.ID(), eventData["id"])
	assert.Equal(t, componentID.Value(), eventData["componentId"])
	assert.Equal(t, pillarID.Value(), eventData["pillarId"])
	assert.Equal(t, "Innovation", eventData["pillarName"])
	assert.Equal(t, 3, eventData["score"])
	assert.Equal(t, "Supports innovation goals", eventData["rationale"])
	assert.Equal(t, "user@example.com", eventData["scoredBy"])
}

func TestUpdateApplicationFitScore(t *testing.T) {
	aggregate := createApplicationFitScore(t)
	aggregate.MarkChangesAsCommitted()

	newScore, _ := valueobjects.NewFitScore(5)
	newRationale, _ := valueobjects.NewFitRationale("Now excellent alignment")
	updatedBy, _ := valueobjects.NewUserIdentifier("admin@example.com")

	err := aggregate.Update(newScore, newRationale, updatedBy)
	require.NoError(t, err)
	assert.Equal(t, newScore, aggregate.Score())
	assert.Equal(t, newRationale, aggregate.Rationale())

	events := aggregate.GetUncommittedChanges()
	require.Len(t, events, 1)
	assert.Equal(t, "ApplicationFitScoreUpdated", events[0].EventType())
}

func TestUpdateApplicationFitScore_EventContainsOldValues(t *testing.T) {
	aggregate := createApplicationFitScore(t)
	oldScore := aggregate.Score().Value()
	oldRationale := aggregate.Rationale().Value()
	aggregate.MarkChangesAsCommitted()

	newScore, _ := valueobjects.NewFitScore(5)
	newRationale, _ := valueobjects.NewFitRationale("Updated rationale")
	updatedBy, _ := valueobjects.NewUserIdentifier("admin@example.com")

	err := aggregate.Update(newScore, newRationale, updatedBy)
	require.NoError(t, err)

	events := aggregate.GetUncommittedChanges()
	eventData := events[0].EventData()
	assert.Equal(t, oldScore, eventData["oldScore"])
	assert.Equal(t, oldRationale, eventData["oldRationale"])
	assert.Equal(t, "admin@example.com", eventData["updatedBy"])
}

func TestRemoveApplicationFitScore(t *testing.T) {
	aggregate := createApplicationFitScore(t)
	aggregate.MarkChangesAsCommitted()

	removedBy, _ := valueobjects.NewUserIdentifier("admin@example.com")
	err := aggregate.Remove(removedBy)
	require.NoError(t, err)

	events := aggregate.GetUncommittedChanges()
	require.Len(t, events, 1)
	assert.Equal(t, "ApplicationFitScoreRemoved", events[0].EventType())

	eventData := events[0].EventData()
	assert.Equal(t, aggregate.ID(), eventData["id"])
	assert.Equal(t, aggregate.ComponentID().Value(), eventData["componentId"])
	assert.Equal(t, aggregate.PillarID().Value(), eventData["pillarId"])
	assert.Equal(t, "admin@example.com", eventData["removedBy"])
}

func TestLoadApplicationFitScoreFromHistory(t *testing.T) {
	original := createApplicationFitScore(t)
	events := original.GetUncommittedChanges()

	loaded, err := LoadApplicationFitScoreFromHistory(events)
	require.NoError(t, err)
	assert.Equal(t, original.ID(), loaded.ID())
	assert.Equal(t, original.ComponentID().Value(), loaded.ComponentID().Value())
	assert.Equal(t, original.PillarID().Value(), loaded.PillarID().Value())
	assert.Equal(t, original.Score().Value(), loaded.Score().Value())
	assert.Equal(t, original.Rationale().Value(), loaded.Rationale().Value())
	assert.Equal(t, original.ScoredBy().Value(), loaded.ScoredBy().Value())
}

func TestLoadApplicationFitScoreFromHistory_WithUpdate(t *testing.T) {
	aggregate := createApplicationFitScore(t)

	newScore, _ := valueobjects.NewFitScore(5)
	newRationale, _ := valueobjects.NewFitRationale("Updated")
	updatedBy, _ := valueobjects.NewUserIdentifier("admin@example.com")
	aggregate.Update(newScore, newRationale, updatedBy)

	events := aggregate.GetUncommittedChanges()

	loaded, err := LoadApplicationFitScoreFromHistory(events)
	require.NoError(t, err)
	assert.Equal(t, 5, loaded.Score().Value())
	assert.Equal(t, "Updated", loaded.Rationale().Value())
}

func createApplicationFitScore(t *testing.T) *ApplicationFitScore {
	t.Helper()

	componentID, _ := valueobjects.NewComponentIDFromString("550e8400-e29b-41d4-a716-446655440000")
	pillarID, _ := valueobjects.NewPillarIDFromString("550e8400-e29b-41d4-a716-446655440001")
	pillarName, _ := valueobjects.NewPillarName("Test Pillar")
	score, _ := valueobjects.NewFitScore(3)
	rationale, _ := valueobjects.NewFitRationale("Initial rationale")
	scoredBy, _ := valueobjects.NewUserIdentifier("user@example.com")

	aggregate, err := SetApplicationFitScore(NewFitScoreParams{
		ComponentID: componentID,
		PillarID:    pillarID,
		PillarName:  pillarName,
		Score:       score,
		Rationale:   rationale,
		ScoredBy:    scoredBy,
	})
	require.NoError(t, err)

	return aggregate
}
