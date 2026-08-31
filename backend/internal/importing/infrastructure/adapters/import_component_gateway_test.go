package adapters_test

import (
	"context"
	"testing"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	"easi/backend/internal/importing/infrastructure/adapters"
)

func TestImportComponentGatewayCreateComponentDispatchesPublishedCommand(t *testing.T) {
	bus := &recordingBus{createdID: "component-1"}
	gateway := adapters.NewImportComponentGateway(bus)
	want := amPL.CreateApplicationComponent{
		Name:        "Billing",
		Description: "Handles invoices",
	}

	id, err := gateway.CreateComponent(context.Background(), want)

	assertNoError(t, err)
	assertCreatedID(t, id, "component-1")
	cmd := dispatchedCommand[*amPL.CreateApplicationComponent](t, bus, "CreateApplicationComponent")
	assertDispatched(t, *cmd, want)
}

func TestImportComponentGatewayCreateRelationDispatchesPublishedCommand(t *testing.T) {
	bus := &recordingBus{createdID: "relation-1"}
	gateway := adapters.NewImportComponentGateway(bus)
	want := amPL.CreateComponentRelation{
		SourceComponentID: "source-1",
		TargetComponentID: "target-1",
		RelationType:      "Serves",
		Name:              "serves",
		Description:       "billing serves invoicing",
	}

	id, err := gateway.CreateRelation(context.Background(), want)

	assertNoError(t, err)
	assertCreatedID(t, id, "relation-1")
	cmd := dispatchedCommand[*amPL.CreateComponentRelation](t, bus, "CreateComponentRelation")
	assertDispatched(t, *cmd, want)
}

func TestImportComponentGatewayWrapsDispatchFailure(t *testing.T) {
	gateway := adapters.NewImportComponentGateway(failingBus())

	_, err := gateway.CreateComponent(context.Background(), amPL.CreateApplicationComponent{Name: "Billing"})

	assertWrappedDispatchError(t, err, "Billing")
}
