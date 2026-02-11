package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReorderStagesHandler_Success(t *testing.T) {
	vs := newTestValueStream(t)
	name1, _ := valueobjects.NewStageName("First")
	name2, _ := valueobjects.NewStageName("Second")
	desc := valueobjects.MustNewDescription("")
	id1, _ := vs.AddStage(name1, desc, nil)
	id2, _ := vs.AddStage(name2, desc, nil)
	vs.MarkChangesAsCommitted()

	repo := &mockStageRepository{stream: vs}
	handler := NewReorderStagesHandler(repo)

	cmd := &commands.ReorderStages{
		ValueStreamID: vs.ID(),
		Positions: []commands.StagePositionEntry{
			{StageID: id2.Value(), Position: 1},
			{StageID: id1.Value(), Position: 2},
		},
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
}

func TestReorderStagesHandler_InvalidPositions(t *testing.T) {
	vs := newTestValueStream(t)
	name1, _ := valueobjects.NewStageName("First")
	name2, _ := valueobjects.NewStageName("Second")
	desc := valueobjects.MustNewDescription("")
	id1, _ := vs.AddStage(name1, desc, nil)
	vs.AddStage(name2, desc, nil)
	vs.MarkChangesAsCommitted()

	repo := &mockStageRepository{stream: vs}
	handler := NewReorderStagesHandler(repo)

	cmd := &commands.ReorderStages{
		ValueStreamID: vs.ID(),
		Positions: []commands.StagePositionEntry{
			{StageID: id1.Value(), Position: 1},
		},
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
