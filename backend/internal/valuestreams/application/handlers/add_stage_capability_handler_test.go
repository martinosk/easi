package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/application/gateways"
	"easi/backend/internal/valuestreams/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCapabilityGateway struct {
	info *gateways.CapabilityInfo
	err  error
}

func (m *mockCapabilityGateway) CapabilityExists(ctx context.Context, capabilityID string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.info != nil, nil
}

func (m *mockCapabilityGateway) GetCapability(ctx context.Context, capabilityID string) (*gateways.CapabilityInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.info, nil
}

func TestAddStageCapabilityHandler_Success(t *testing.T) {
	vs := newTestValueStream(t)
	name, _ := valueobjects.NewStageName("Discovery")
	desc := valueobjects.MustNewDescription("")
	stageID, _ := vs.AddStage(name, desc, nil)
	vs.MarkChangesAsCommitted()

	repo := &mockStageRepository{stream: vs}
	gateway := &mockCapabilityGateway{info: &gateways.CapabilityInfo{ID: "00000000-0000-0000-0000-000000000123", Name: "Test Cap"}}
	handler := NewAddStageCapabilityHandler(repo, gateway)

	cmd := &commands.AddStageCapability{
		ValueStreamID: vs.ID(),
		StageID:       stageID.Value(),
		CapabilityID:  "00000000-0000-0000-0000-000000000123",
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
	gateway := &mockCapabilityGateway{info: nil}
	handler := NewAddStageCapabilityHandler(repo, gateway)

	cmd := &commands.AddStageCapability{
		ValueStreamID: vs.ID(),
		StageID:       stageID.Value(),
		CapabilityID:  "00000000-0000-0000-0000-000000000999",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrCapabilityNotFound)
}

func TestAddStageCapabilityHandler_AlreadyMapped(t *testing.T) {
	vs := newTestValueStream(t)
	name, _ := valueobjects.NewStageName("Discovery")
	desc := valueobjects.MustNewDescription("")
	stageID, _ := vs.AddStage(name, desc, nil)

	capRef, _ := valueobjects.NewCapabilityRef("00000000-0000-0000-0000-000000000123")
	vs.AddCapabilityToStage(stageID, capRef, "Test Cap")
	vs.MarkChangesAsCommitted()

	repo := &mockStageRepository{stream: vs}
	gateway := &mockCapabilityGateway{info: &gateways.CapabilityInfo{ID: "00000000-0000-0000-0000-000000000123", Name: "Test Cap"}}
	handler := NewAddStageCapabilityHandler(repo, gateway)

	cmd := &commands.AddStageCapability{
		ValueStreamID: vs.ID(),
		StageID:       stageID.Value(),
		CapabilityID:  "00000000-0000-0000-0000-000000000123",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestAddStageCapabilityHandler_StageNotFound(t *testing.T) {
	vs := newTestValueStream(t)
	repo := &mockStageRepository{stream: vs}
	gateway := &mockCapabilityGateway{info: &gateways.CapabilityInfo{ID: "00000000-0000-0000-0000-000000000123", Name: "Test Cap"}}
	handler := NewAddStageCapabilityHandler(repo, gateway)

	cmd := &commands.AddStageCapability{
		ValueStreamID: vs.ID(),
		StageID:       valueobjects.NewStageID().Value(),
		CapabilityID:  "00000000-0000-0000-0000-000000000123",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrStageNotFound)
}
