package handlers

import (
	"context"
	"log"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
)

type AssignmentReader interface {
	GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.AssignmentDTO, error)
	AssignmentExists(ctx context.Context, domainID, capabilityID string) (bool, error)
}

type CapabilityLookup interface {
	GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error)
}

type OnCapabilityParentChangedHandler struct {
	commandBus       cqrs.CommandBus
	assignmentReader AssignmentReader
	capabilityReader CapabilityLookup
}

func NewOnCapabilityParentChangedHandler(
	commandBus cqrs.CommandBus,
	assignmentReader AssignmentReader,
	capabilityReader CapabilityLookup,
) *OnCapabilityParentChangedHandler {
	return &OnCapabilityParentChangedHandler{
		commandBus:       commandBus,
		assignmentReader: assignmentReader,
		capabilityReader: capabilityReader,
	}
}

func (h *OnCapabilityParentChangedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	e, ok := event.(events.CapabilityParentChanged)
	if !ok {
		return nil
	}

	if e.OldLevel != "L1" || e.NewLevel == "L1" {
		return nil
	}

	log.Printf("Handling CapabilityParentChanged: capability %s changed from %s to %s", e.CapabilityID, e.OldLevel, e.NewLevel)

	assignments, err := h.assignmentReader.GetByCapabilityID(ctx, e.CapabilityID)
	if err != nil {
		log.Printf("Error querying assignments for capability %s: %v", e.CapabilityID, err)
		return err
	}

	if len(assignments) == 0 {
		log.Printf("No assignments found for capability %s", e.CapabilityID)
		return nil
	}

	l1AncestorID, err := h.findL1Ancestor(ctx, e.NewParentID)
	if err != nil {
		log.Printf("Error finding L1 ancestor for %s: %v", e.NewParentID, err)
		return err
	}

	log.Printf("Found %d assignments for capability %s, reassigning to L1 ancestor %s", len(assignments), e.CapabilityID, l1AncestorID)

	for _, assignment := range assignments {
		unassignCmd := &commands.UnassignCapabilityFromDomain{
			AssignmentID: assignment.AssignmentID,
		}

		if _, err := h.commandBus.Dispatch(ctx, unassignCmd); err != nil {
			log.Printf("Error unassigning capability %s from domain %s: %v", e.CapabilityID, assignment.BusinessDomainID, err)
			continue
		}

		log.Printf("Unassigned capability %s from domain %s", e.CapabilityID, assignment.BusinessDomainID)

		l1AlreadyAssigned, err := h.assignmentReader.AssignmentExists(ctx, assignment.BusinessDomainID, l1AncestorID)
		if err != nil {
			log.Printf("Error checking if L1 %s is assigned to domain %s: %v", l1AncestorID, assignment.BusinessDomainID, err)
			continue
		}

		if l1AlreadyAssigned {
			log.Printf("L1 ancestor %s already assigned to domain %s, skipping", l1AncestorID, assignment.BusinessDomainID)
			continue
		}

		assignCmd := &commands.AssignCapabilityToDomain{
			BusinessDomainID: assignment.BusinessDomainID,
			CapabilityID:     l1AncestorID,
		}

		if _, err := h.commandBus.Dispatch(ctx, assignCmd); err != nil {
			log.Printf("Error assigning L1 ancestor %s to domain %s: %v", l1AncestorID, assignment.BusinessDomainID, err)
			continue
		}

		log.Printf("Assigned L1 ancestor %s to domain %s", l1AncestorID, assignment.BusinessDomainID)
	}

	return nil
}

func (h *OnCapabilityParentChangedHandler) findL1Ancestor(ctx context.Context, capabilityID string) (string, error) {
	currentID := capabilityID

	for {
		capability, err := h.capabilityReader.GetByID(ctx, currentID)
		if err != nil {
			return "", err
		}

		if capability == nil {
			return currentID, nil
		}

		if capability.Level == "L1" {
			return capability.ID, nil
		}

		if capability.ParentID == "" {
			return capability.ID, nil
		}

		currentID = capability.ParentID
	}
}
