package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCapabilityGateway struct {
	exists bool
	err    error
}

func (m *mockCapabilityGateway) CapabilityExists(ctx context.Context, capabilityID string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.exists, nil
}

func TestAddStageCapabilityHandler_Success(t *testing.T) {
	vs := newTestValueStream(t)
	name, _ := valueobjects.NewStageName("Discovery")
	desc := valueobjects.MustNewDescription("")
	stageID, _ := vs.AddStage(name, desc, nil)
	vs.MarkChangesAsCommitted()

	repo := &mockStageRepository{stream: vs}
	gateway := &mockCapabilityGateway{exists: true}
	handler := NewAddStageCapabilityHandler(repo, gateway)

	cmd := &commands.AddStageCapability{
		ValueStreamID: vs.ID(),
		StageID:       stageID.Value(),
		CapabilityID:  "cap-123",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
}

func TestAddStageCapabilityHandler_CapabilityNotFound(t *testing.T) {
	vs := newTestValueStream(t)
	name, _ := valueobjects.NewStageName("Discovery")
	desc := valueobjects.MustNewDescription("")
	stageID, _ := vs.AddStage(name, desc, nil)
	vs.MarkChangesAsCommitted()

	repo := &mockStageRepository{stream: vs}
	gateway := &mockCapabilityGateway{exists: false}
	handler := NewAddStageCapabilityHandler(repo, gateway)

	cmd := &commands.AddStageCapability{
		ValueStreamID: vs.ID(),
		StageID:       stageID.Value(),
		CapabilityID:  "cap-nonexistent",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrCapabilityNotFound)
}

func TestAddStageCapabilityHandler_AlreadyMapped(t *testing.T) {
	vs := newTestValueStream(t)
	name, _ := valueobjects.NewStageName("Discovery")
	desc := valueobjects.MustNewDescription("")
	stageID, _ := vs.AddStage(name, desc, nil)

	capRef, _ := valueobjects.NewCapabilityRef("cap-123")
	vs.AddCapabilityToStage(stageID, capRef)
	vs.MarkChangesAsCommitted()

	repo := &mockStageRepository{stream: vs}
	gateway := &mockCapabilityGateway{exists: true}
	handler := NewAddStageCapabilityHandler(repo, gateway)

	cmd := &commands.AddStageCapability{
		ValueStreamID: vs.ID(),
		StageID:       stageID.Value(),
		CapabilityID:  "cap-123",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestAddStageCapabilityHandler_StageNotFound(t *testing.T) {
	vs := newTestValueStream(t)
	repo := &mockStageRepository{stream: vs}
	gateway := &mockCapabilityGateway{exists: true}
	handler := NewAddStageCapabilityHandler(repo, gateway)

	cmd := &commands.AddStageCapability{
		ValueStreamID: vs.ID(),
		StageID:       valueobjects.NewStageID().Value(),
		CapabilityID:  "cap-123",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrStageNotFound)
}
