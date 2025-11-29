package handlers

import (
	"context"
	"log"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/domain"
)

type AssignmentByDomainReader interface {
	GetByDomainID(ctx context.Context, domainID string) ([]readmodels.AssignmentDTO, error)
}

type OnBusinessDomainDeletedHandler struct {
	commandBus cqrs.CommandBus
	readModel  AssignmentByDomainReader
}

func NewOnBusinessDomainDeletedHandler(
	commandBus cqrs.CommandBus,
	readModel AssignmentByDomainReader,
) *OnBusinessDomainDeletedHandler {
	return &OnBusinessDomainDeletedHandler{
		commandBus: commandBus,
		readModel:  readModel,
	}
}

func (h *OnBusinessDomainDeletedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	domainID := event.AggregateID()

	log.Printf("Handling BusinessDomainDeleted for domain %s", domainID)

	assignments, err := h.readModel.GetByDomainID(ctx, domainID)
	if err != nil {
		log.Printf("Error querying assignments for domain %s: %v", domainID, err)
		return err
	}

	log.Printf("Found %d assignments for domain %s", len(assignments), domainID)

	for _, assignment := range assignments {
		cmd := &commands.UnassignCapabilityFromDomain{
			AssignmentID: assignment.AssignmentID,
		}

		if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
			log.Printf("Error unassigning capability %s from domain %s: %v",
				assignment.CapabilityID, assignment.BusinessDomainID, err)
			continue
		}

		log.Printf("Unassigned capability %s from domain %s (assignment %s)",
			assignment.CapabilityID, assignment.BusinessDomainID, assignment.AssignmentID)
	}

	return nil
}
