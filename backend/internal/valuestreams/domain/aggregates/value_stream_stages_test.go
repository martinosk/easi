package aggregates

import (
	"testing"

	"easi/backend/internal/valuestreams/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueStream_AddStage(t *testing.T) {
	vs := createValueStream(t, "Customer Onboarding")
	vs.MarkChangesAsCommitted()

	name, _ := valueobjects.NewStageName("Discovery")
	desc := valueobjects.MustNewDescription("Initial discovery phase")

	stageID, err := vs.AddStage(name, desc, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, stageID.Value())
	assert.Equal(t, 1, vs.StageCount())

	stages := vs.Stages()
	assert.Equal(t, "Discovery", stages[0].Name().Value())
	assert.Equal(t, 1, stages[0].Position().Value())

	uncommitted := vs.GetUncommittedChanges()
	require.Len(t, uncommitted, 1)
	assert.Equal(t, "ValueStreamStageAdded", uncommitted[0].EventType())
}

func TestValueStream_AddStage_DuplicateName(t *testing.T) {
	vs := createValueStreamWithStage(t, "Test VS", "Discovery")

	name, _ := valueobjects.NewStageName("Discovery")
	desc := valueobjects.MustNewDescription("")

	_, err := vs.AddStage(name, desc, nil)
	assert.ErrorIs(t, err, ErrStageNameExists)
}

func TestValueStream_AddStage_WithInsertPosition(t *testing.T) {
	vs := createValueStream(t, "Test VS")
	vs.MarkChangesAsCommitted()

	name1, _ := valueobjects.NewStageName("First")
	name2, _ := valueobjects.NewStageName("Second")
	name3, _ := valueobjects.NewStageName("Inserted")
	desc := valueobjects.MustNewDescription("")

	vs.AddStage(name1, desc, nil)
	vs.AddStage(name2, desc, nil)
	vs.MarkChangesAsCommitted()

	insertPos, _ := valueobjects.NewStagePosition(1)
	_, err := vs.AddStage(name3, desc, &insertPos)
	require.NoError(t, err)

	assert.Equal(t, 3, vs.StageCount())
	stages := vs.Stages()

	positions := make(map[string]int)
	for _, s := range stages {
		positions[s.Name().Value()] = s.Position().Value()
	}
	assert.Equal(t, 1, positions["Inserted"])
	assert.Equal(t, 2, positions["First"])
	assert.Equal(t, 3, positions["Second"])
}

func TestValueStream_UpdateStage(t *testing.T) {
	vs := createValueStreamWithStage(t, "Test VS", "Discovery")
	stageID := vs.Stages()[0].ID()
	vs.MarkChangesAsCommitted()

	newName, _ := valueobjects.NewStageName("Research")
	newDesc := valueobjects.MustNewDescription("Updated description")

	err := vs.UpdateStage(stageID, newName, newDesc)
	require.NoError(t, err)

	assert.Equal(t, "Research", vs.Stages()[0].Name().Value())
	assert.Equal(t, "Updated description", vs.Stages()[0].Description().Value())

	uncommitted := vs.GetUncommittedChanges()
	require.Len(t, uncommitted, 1)
	assert.Equal(t, "ValueStreamStageUpdated", uncommitted[0].EventType())
}

func TestValueStream_UpdateStage_NotFound(t *testing.T) {
	vs := createValueStream(t, "Test VS")
	vs.MarkChangesAsCommitted()

	fakeID := valueobjects.NewStageID()
	name, _ := valueobjects.NewStageName("Test")
	desc := valueobjects.MustNewDescription("")

	err := vs.UpdateStage(fakeID, name, desc)
	assert.ErrorIs(t, err, ErrStageNotFound)
}

func TestValueStream_UpdateStage_DuplicateName(t *testing.T) {
	vs := createValueStream(t, "Test VS")
	vs.MarkChangesAsCommitted()

	name1, _ := valueobjects.NewStageName("First")
	name2, _ := valueobjects.NewStageName("Second")
	desc := valueobjects.MustNewDescription("")

	vs.AddStage(name1, desc, nil)
	vs.AddStage(name2, desc, nil)
	vs.MarkChangesAsCommitted()

	secondID := findStageByName(vs, "Second")
	duplicateName, _ := valueobjects.NewStageName("First")
	err := vs.UpdateStage(secondID, duplicateName, desc)
	assert.ErrorIs(t, err, ErrStageNameExists)
}

func TestValueStream_RemoveStage(t *testing.T) {
	vs := createValueStream(t, "Test VS")
	vs.MarkChangesAsCommitted()

	name1, _ := valueobjects.NewStageName("First")
	name2, _ := valueobjects.NewStageName("Second")
	name3, _ := valueobjects.NewStageName("Third")
	desc := valueobjects.MustNewDescription("")

	vs.AddStage(name1, desc, nil)
	vs.AddStage(name2, desc, nil)
	vs.AddStage(name3, desc, nil)
	vs.MarkChangesAsCommitted()

	secondID := findStageByName(vs, "Second")
	err := vs.RemoveStage(secondID)
	require.NoError(t, err)

	assert.Equal(t, 2, vs.StageCount())

	positions := make(map[string]int)
	for _, s := range vs.Stages() {
		positions[s.Name().Value()] = s.Position().Value()
	}
	assert.Equal(t, 1, positions["First"])
	assert.Equal(t, 2, positions["Third"])

	uncommitted := vs.GetUncommittedChanges()
	require.Len(t, uncommitted, 1)
	assert.Equal(t, "ValueStreamStageRemoved", uncommitted[0].EventType())
}

func TestValueStream_RemoveStage_NotFound(t *testing.T) {
	vs := createValueStream(t, "Test VS")
	vs.MarkChangesAsCommitted()

	fakeID := valueobjects.NewStageID()
	err := vs.RemoveStage(fakeID)
	assert.ErrorIs(t, err, ErrStageNotFound)
}

func TestValueStream_ReorderStages(t *testing.T) {
	vs := createValueStream(t, "Test VS")
	vs.MarkChangesAsCommitted()

	name1, _ := valueobjects.NewStageName("First")
	name2, _ := valueobjects.NewStageName("Second")
	name3, _ := valueobjects.NewStageName("Third")
	desc := valueobjects.MustNewDescription("")

	vs.AddStage(name1, desc, nil)
	vs.AddStage(name2, desc, nil)
	vs.AddStage(name3, desc, nil)
	vs.MarkChangesAsCommitted()

	firstID := findStageByName(vs, "First")
	secondID := findStageByName(vs, "Second")
	thirdID := findStageByName(vs, "Third")

	err := vs.ReorderStages([]StagePositionUpdate{
		{StageID: thirdID.Value(), Position: 1},
		{StageID: firstID.Value(), Position: 2},
		{StageID: secondID.Value(), Position: 3},
	})
	require.NoError(t, err)

	positions := make(map[string]int)
	for _, s := range vs.Stages() {
		positions[s.Name().Value()] = s.Position().Value()
	}
	assert.Equal(t, 1, positions["Third"])
	assert.Equal(t, 2, positions["First"])
	assert.Equal(t, 3, positions["Second"])
}

func TestValueStream_ReorderStages_InvalidPositions(t *testing.T) {
	vs := createValueStream(t, "Test VS")
	vs.MarkChangesAsCommitted()

	name1, _ := valueobjects.NewStageName("First")
	name2, _ := valueobjects.NewStageName("Second")
	desc := valueobjects.MustNewDescription("")

	vs.AddStage(name1, desc, nil)
	vs.AddStage(name2, desc, nil)
	vs.MarkChangesAsCommitted()

	firstID := findStageByName(vs, "First")

	err := vs.ReorderStages([]StagePositionUpdate{
		{StageID: firstID.Value(), Position: 1},
	})
	assert.ErrorIs(t, err, ErrInvalidStagePositions)
}

func TestValueStream_ReorderStages_MissingStage(t *testing.T) {
	vs := createValueStream(t, "Test VS")
	vs.MarkChangesAsCommitted()

	name1, _ := valueobjects.NewStageName("First")
	name2, _ := valueobjects.NewStageName("Second")
	desc := valueobjects.MustNewDescription("")

	vs.AddStage(name1, desc, nil)
	vs.AddStage(name2, desc, nil)
	vs.MarkChangesAsCommitted()

	fakeID := valueobjects.NewStageID()
	firstID := findStageByName(vs, "First")

	err := vs.ReorderStages([]StagePositionUpdate{
		{StageID: firstID.Value(), Position: 1},
		{StageID: fakeID.Value(), Position: 2},
	})
	assert.ErrorIs(t, err, ErrInvalidStagePositions)
}

func TestValueStream_AddCapabilityToStage(t *testing.T) {
	vs := createValueStreamWithStage(t, "Test VS", "Discovery")
	stageID := vs.Stages()[0].ID()
	vs.MarkChangesAsCommitted()

	capRef, _ := valueobjects.NewCapabilityRef("cap-123")
	err := vs.AddCapabilityToStage(stageID, capRef)
	require.NoError(t, err)

	stage := vs.Stages()[0]
	assert.Len(t, stage.CapabilityRefs(), 1)
	assert.True(t, stage.HasCapability(capRef))

	uncommitted := vs.GetUncommittedChanges()
	require.Len(t, uncommitted, 1)
	assert.Equal(t, "ValueStreamStageCapabilityAdded", uncommitted[0].EventType())
}

func TestValueStream_AddCapabilityToStage_AlreadyMapped(t *testing.T) {
	vs := createValueStreamWithStage(t, "Test VS", "Discovery")
	stageID := vs.Stages()[0].ID()

	capRef, _ := valueobjects.NewCapabilityRef("cap-123")
	vs.AddCapabilityToStage(stageID, capRef)
	vs.MarkChangesAsCommitted()

	err := vs.AddCapabilityToStage(stageID, capRef)
	assert.ErrorIs(t, err, ErrCapabilityAlreadyMapped)
}

func TestValueStream_AddCapabilityToStage_StageNotFound(t *testing.T) {
	vs := createValueStream(t, "Test VS")
	vs.MarkChangesAsCommitted()

	fakeID := valueobjects.NewStageID()
	capRef, _ := valueobjects.NewCapabilityRef("cap-123")

	err := vs.AddCapabilityToStage(fakeID, capRef)
	assert.ErrorIs(t, err, ErrStageNotFound)
}

func TestValueStream_RemoveCapabilityFromStage(t *testing.T) {
	vs := createValueStreamWithStage(t, "Test VS", "Discovery")
	stageID := vs.Stages()[0].ID()

	capRef, _ := valueobjects.NewCapabilityRef("cap-123")
	vs.AddCapabilityToStage(stageID, capRef)
	vs.MarkChangesAsCommitted()

	err := vs.RemoveCapabilityFromStage(stageID, capRef)
	require.NoError(t, err)

	assert.Empty(t, vs.Stages()[0].CapabilityRefs())

	uncommitted := vs.GetUncommittedChanges()
	require.Len(t, uncommitted, 1)
	assert.Equal(t, "ValueStreamStageCapabilityRemoved", uncommitted[0].EventType())
}

func TestValueStream_RemoveCapabilityFromStage_NotMapped(t *testing.T) {
	vs := createValueStreamWithStage(t, "Test VS", "Discovery")
	stageID := vs.Stages()[0].ID()
	vs.MarkChangesAsCommitted()

	capRef, _ := valueobjects.NewCapabilityRef("cap-123")
	err := vs.RemoveCapabilityFromStage(stageID, capRef)
	assert.ErrorIs(t, err, ErrCapabilityNotMapped)
}

func TestValueStream_LoadFromHistory_WithStageEvents(t *testing.T) {
	vs := createValueStream(t, "Test VS")

	name1, _ := valueobjects.NewStageName("Discovery")
	desc := valueobjects.MustNewDescription("Phase 1")
	vs.AddStage(name1, desc, nil)

	name2, _ := valueobjects.NewStageName("Design")
	vs.AddStage(name2, desc, nil)

	stageID := vs.Stages()[0].ID()
	capRef, _ := valueobjects.NewCapabilityRef("cap-1")
	vs.AddCapabilityToStage(stageID, capRef)

	allEvents := vs.GetUncommittedChanges()

	loaded, err := LoadValueStreamFromHistory(allEvents)
	require.NoError(t, err)

	assert.Equal(t, 2, loaded.StageCount())
	assert.Equal(t, "Discovery", loaded.Stages()[0].Name().Value())
	assert.Len(t, loaded.Stages()[0].CapabilityRefs(), 1)
}

func createValueStreamWithStage(t *testing.T, vsName, stageName string) *ValueStream {
	t.Helper()
	vs := createValueStream(t, vsName)
	vs.MarkChangesAsCommitted()

	name, err := valueobjects.NewStageName(stageName)
	require.NoError(t, err)
	desc := valueobjects.MustNewDescription("")

	_, err = vs.AddStage(name, desc, nil)
	require.NoError(t, err)

	return vs
}

func findStageByName(vs *ValueStream, name string) valueobjects.StageID {
	for _, s := range vs.Stages() {
		if s.Name().Value() == name {
			return s.ID()
		}
	}
	return valueobjects.StageID{}
}
