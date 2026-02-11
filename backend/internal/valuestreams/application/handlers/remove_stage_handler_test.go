package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveStageHandler_Success(t *testing.T) {
	vs := newTestValueStream(t)
	name, _ := valueobjects.NewStageName("Discovery")
	desc := valueobjects.MustNewDescription("")
	stageID, _ := vs.AddStage(name, desc, nil)
	vs.MarkChangesAsCommitted()

	repo := &mockStageRepository{stream: vs}
	handler := NewRemoveStageHandler(repo)

	cmd := &commands.RemoveStage{
		ValueStreamID: vs.ID(),
		StageID:       stageID.Value(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, 0, repo.saved[0].StageCount())
}

func TestRemoveStageHandler_NotFound(t *testing.T) {
	vs := newTestValueStream(t)
	repo := &mockStageRepository{stream: vs}
	handler := NewRemoveStageHandler(repo)

	cmd := &commands.RemoveStage{
		ValueStreamID: vs.ID(),
		StageID:       valueobjects.NewStageID().Value(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrStageNotFound)
}
