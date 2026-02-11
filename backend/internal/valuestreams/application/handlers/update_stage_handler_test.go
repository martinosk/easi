package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateStageHandler_Success(t *testing.T) {
	vs := newTestValueStream(t)
	name, _ := valueobjects.NewStageName("Discovery")
	desc := valueobjects.MustNewDescription("")
	stageID, _ := vs.AddStage(name, desc, nil)
	vs.MarkChangesAsCommitted()

	repo := &mockStageRepository{stream: vs}
	handler := NewUpdateStageHandler(repo)

	cmd := &commands.UpdateStage{
		ValueStreamID: vs.ID(),
		StageID:       stageID.Value(),
		Name:          "Research",
		Description:   "Updated description",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, "Research", repo.saved[0].Stages()[0].Name().Value())
}

func TestUpdateStageHandler_NotFound(t *testing.T) {
	vs := newTestValueStream(t)
	repo := &mockStageRepository{stream: vs}
	handler := NewUpdateStageHandler(repo)

	cmd := &commands.UpdateStage{
		ValueStreamID: vs.ID(),
		StageID:       valueobjects.NewStageID().Value(),
		Name:          "Test",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrStageNotFound)
}

func TestUpdateStageHandler_DuplicateName(t *testing.T) {
	vs := newTestValueStream(t)
	name1, _ := valueobjects.NewStageName("First")
	name2, _ := valueobjects.NewStageName("Second")
	desc := valueobjects.MustNewDescription("")
	vs.AddStage(name1, desc, nil)
	stageID2, _ := vs.AddStage(name2, desc, nil)
	vs.MarkChangesAsCommitted()

	repo := &mockStageRepository{stream: vs}
	handler := NewUpdateStageHandler(repo)

	cmd := &commands.UpdateStage{
		ValueStreamID: vs.ID(),
		StageID:       stageID2.Value(),
		Name:          "First",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrStageNameExists)
}
