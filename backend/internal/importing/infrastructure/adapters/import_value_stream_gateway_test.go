package adapters_test

import (
	"context"
	"testing"

	"easi/backend/internal/importing/infrastructure/adapters"
	vsPL "easi/backend/internal/valuestreams/publishedlanguage"
)

func TestImportValueStreamGatewayCreateValueStreamDispatchesPublishedCommand(t *testing.T) {
	bus := &recordingBus{createdID: "value-stream-1"}
	gateway := adapters.NewImportValueStreamGateway(bus)
	want := vsPL.CreateValueStream{
		Name:        "Order to Cash",
		Description: "End to end",
	}

	id, err := gateway.CreateValueStream(context.Background(), want)

	assertNoError(t, err)
	assertCreatedID(t, id, "value-stream-1")
	cmd := dispatchedCommand[*vsPL.CreateValueStream](t, bus, "CreateValueStream")
	assertDispatched(t, *cmd, want)
}

func TestImportValueStreamGatewayAddStageDispatchesPublishedCommand(t *testing.T) {
	bus := &recordingBus{createdID: "stage-1"}
	gateway := adapters.NewImportValueStreamGateway(bus)
	want := vsPL.AddStage{
		ValueStreamID: "value-stream-1",
		Name:          "Main Flow",
		Description:   "default",
	}

	id, err := gateway.AddStage(context.Background(), want)

	assertNoError(t, err)
	assertCreatedID(t, id, "stage-1")
	cmd := dispatchedCommand[*vsPL.AddStage](t, bus, "AddStage")
	assertDispatched(t, *cmd, want)
}

func TestImportValueStreamGatewayMapCapabilityToStageDispatchesPublishedCommand(t *testing.T) {
	bus := &recordingBus{}
	gateway := adapters.NewImportValueStreamGateway(bus)
	want := vsPL.AddStageCapability{
		ValueStreamID: "value-stream-1",
		StageID:       "stage-1",
		CapabilityID:  "capability-1",
	}

	err := gateway.MapCapabilityToStage(context.Background(), want)

	assertNoError(t, err)
	cmd := dispatchedCommand[*vsPL.AddStageCapability](t, bus, "AddStageCapability")
	assertDispatched(t, *cmd, want)
}

func TestImportValueStreamGatewayWrapsDispatchFailure(t *testing.T) {
	gateway := adapters.NewImportValueStreamGateway(failingBus())

	_, err := gateway.AddStage(context.Background(), vsPL.AddStage{ValueStreamID: "value-stream-1"})

	assertWrappedDispatchError(t, err, "value-stream-1")
}
