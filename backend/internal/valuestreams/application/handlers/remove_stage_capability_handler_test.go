package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveStageCapabilityHandler_Success(t *testing.T) {
	vs := newTestValueStream(t)
	name, _ := valueobjects.NewStageName("Discovery")
	desc := valueobjects.MustNewDescription("")
	stageID, _ := vs.AddStage(name, desc, nil)

	capRef, _ := valueobjects.NewCapabilityRef("cap-123")
	vs.AddCapabilityToStage(stageID, capRef)
	vs.MarkChangesAsCommitted()

	repo := &mockStageRepository{stream: vs}
	handler := NewRemoveStageCapabilityHandler(repo)

	cmd := &commands.RemoveStageCapability{
		ValueStreamID: vs.ID(),
		StageID:       stageID.Value(),
		CapabilityID:  "cap-123",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
}

func TestRemoveStageCapabilityHandler_NotMapped(t *testing.T) {
	vs := newTestValueStream(t)
	name, _ := valueobjects.NewStageName("Discovery")
	desc := valueobjects.MustNewDescription("")
	stageID, _ := vs.AddStage(name, desc, nil)
	vs.MarkChangesAsCommitted()

	repo := &mockStageRepository{stream: vs}
	handler := NewRemoveStageCapabilityHandler(repo)

	cmd := &commands.RemoveStageCapability{
		ValueStreamID: vs.ID(),
		StageID:       stageID.Value(),
		CapabilityID:  "cap-123",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
