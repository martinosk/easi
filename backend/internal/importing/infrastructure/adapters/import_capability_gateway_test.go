package adapters_test

import (
	"context"
	"testing"

	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/importing/infrastructure/adapters"
)

func TestImportCapabilityGatewayCreateCapabilityDispatchesPublishedCommand(t *testing.T) {
	bus := &recordingBus{createdID: "capability-1"}
	gateway := adapters.NewImportCapabilityGateway(bus)
	want := cmPL.CreateCapability{
		Name:        "Invoicing",
		Description: "Bill customers",
		ParentID:    "parent-1",
		Level:       "L2",
	}

	id, err := gateway.CreateCapability(context.Background(), want)

	assertNoError(t, err)
	assertCreatedID(t, id, "capability-1")
	cmd := dispatchedCommand[*cmPL.CreateCapability](t, bus, "CreateCapability")
	assertDispatched(t, *cmd, want)
}

func TestImportCapabilityGatewayUpdateMetadataDispatchesPublishedCommand(t *testing.T) {
	bus := &recordingBus{}
	gateway := adapters.NewImportCapabilityGateway(bus)
	want := cmPL.UpdateCapabilityMetadata{
		ID:      "capability-1",
		EAOwner: "owner@example.com",
		Status:  "Active",
	}

	err := gateway.UpdateMetadata(context.Background(), want)

	assertNoError(t, err)
	cmd := dispatchedCommand[*cmPL.UpdateCapabilityMetadata](t, bus, "UpdateCapabilityMetadata")
	assertDispatched(t, *cmd, want)
}

func TestImportCapabilityGatewayLinkSystemDispatchesPublishedCommand(t *testing.T) {
	bus := &recordingBus{createdID: "realization-1"}
	gateway := adapters.NewImportCapabilityGateway(bus)
	want := cmPL.LinkSystemToCapability{
		CapabilityID:     "capability-1",
		ComponentID:      "component-1",
		RealizationLevel: "full",
		Notes:            "imported",
	}

	id, err := gateway.LinkSystem(context.Background(), want)

	assertNoError(t, err)
	assertCreatedID(t, id, "realization-1")
	cmd := dispatchedCommand[*cmPL.LinkSystemToCapability](t, bus, "LinkSystemToCapability")
	assertDispatched(t, *cmd, want)
}

func TestImportCapabilityGatewayAssignToDomainDispatchesPublishedCommand(t *testing.T) {
	bus := &recordingBus{}
	gateway := adapters.NewImportCapabilityGateway(bus)
	want := cmPL.AssignCapabilityToDomain{
		CapabilityID:     "capability-1",
		BusinessDomainID: "domain-1",
	}

	err := gateway.AssignToDomain(context.Background(), want)

	assertNoError(t, err)
	cmd := dispatchedCommand[*cmPL.AssignCapabilityToDomain](t, bus, "AssignCapabilityToDomain")
	assertDispatched(t, *cmd, want)
}

func TestImportCapabilityGatewayWrapsDispatchFailure(t *testing.T) {
	gateway := adapters.NewImportCapabilityGateway(failingBus())

	err := gateway.AssignToDomain(context.Background(), cmPL.AssignCapabilityToDomain{
		CapabilityID:     "capability-1",
		BusinessDomainID: "domain-1",
	})

	assertWrappedDispatchError(t, err, "capability-1")
}
