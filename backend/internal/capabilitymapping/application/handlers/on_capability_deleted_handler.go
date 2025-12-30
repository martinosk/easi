package handlers

import (
	"context"
	"log"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
)

type AssignmentByCapabilityReader interface {
	GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.AssignmentDTO, error)
}

type OnCapabilityDeletedHandler struct {
	commandBus cqrs.CommandBus
	readModel  AssignmentByCapabilityReader
}

func NewOnCapabilityDeletedHandler(
	commandBus cqrs.CommandBus,
	readModel AssignmentByCapabilityReader,
) *OnCapabilityDeletedHandler {
	return &OnCapabilityDeletedHandler{
		commandBus: commandBus,
		readModel:  readModel,
	}
}

func (h *OnCapabilityDeletedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	capabilityID := event.AggregateID()

	log.Printf("Handling CapabilityDeleted for capability %s", capabilityID)

	assignments, err := h.readModel.GetByCapabilityID(ctx, capabilityID)
	if err != nil {
		log.Printf("Error querying assignments for capability %s: %v", capabilityID, err)
		return err
	}

	log.Printf("Found %d assignments for capability %s", len(assignments), capabilityID)

	for _, assignment := range assignments {
		cmd := &commands.UnassignCapabilityFromDomain{
			AssignmentID: assignment.AssignmentID,
		}

		if _, err := h.commandBus.Dispatch(ctx, cmd); err != nil {
			log.Printf("Error unassigning capability %s from domain %s: %v",
				assignment.CapabilityID, assignment.BusinessDomainID, err)
			continue
		}

		log.Printf("Unassigned capability %s from domain %s (assignment %s)",
			assignment.CapabilityID, assignment.BusinessDomainID, assignment.AssignmentID)
	}

	return nil
}
