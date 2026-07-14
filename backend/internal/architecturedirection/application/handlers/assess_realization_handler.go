package handlers

import (
	"context"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type TimeAssessmentRepository interface {
	Save(ctx context.Context, ta *aggregates.TimeAssessment) error
	GetByID(ctx context.Context, id string) (*aggregates.TimeAssessment, error)
}

type ExistingTimeAssessmentLookup interface {
	FindAggregateIDForPair(ctx context.Context, capabilityID, componentID string) (string, bool, error)
}

type AssessRealizationHandler struct {
	repo              TimeAssessmentRepository
	lookup            ExistingTimeAssessmentLookup
	directRealization services.DirectRealizationLookup
}

func NewAssessRealizationHandler(
	repo TimeAssessmentRepository,
	lookup ExistingTimeAssessmentLookup,
	directRealization services.DirectRealizationLookup,
) *AssessRealizationHandler {
	return &AssessRealizationHandler{
		repo:              repo,
		lookup:            lookup,
		directRealization: directRealization,
	}
}

type assessRealizationInputs struct {
	capability valueobjects.PhysicalCapabilityRef
	component  valueobjects.ApplicationRef
	grade      valueobjects.TimeGrade
	rationale  sharedvo.Description
	assessedBy string
}

func (h *AssessRealizationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.AssessRealization)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	inputs, err := parseAssessRealizationInputs(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	realizationID, err := h.verifyDirectRealizationExists(ctx, inputs)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	existingID, exists, err := h.lookup.FindAggregateIDForPair(ctx, inputs.capability.Value(), inputs.component.Value())
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if exists {
		return h.reassessExisting(ctx, existingID, realizationID, inputs)
	}
	return h.createNew(ctx, realizationID, inputs)
}

func (h *AssessRealizationHandler) verifyDirectRealizationExists(ctx context.Context, inputs assessRealizationInputs) (string, error) {
	realizationID, exists, err := h.directRealization(ctx, inputs.capability.Value(), inputs.component.Value())
	if err != nil {
		return "", err
	}
	if !exists {
		return "", services.ErrReferencedEntityNotFound
	}
	return realizationID, nil
}

func (h *AssessRealizationHandler) reassessExisting(
	ctx context.Context,
	existingID string,
	realizationID string,
	inputs assessRealizationInputs,
) (cqrs.CommandResult, error) {
	existing, err := h.repo.GetByID(ctx, existingID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := existing.Reassess(realizationID, inputs.grade, inputs.rationale, inputs.assessedBy); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, existing); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(existing.ID()), nil
}

func (h *AssessRealizationHandler) createNew(
	ctx context.Context,
	realizationID string,
	inputs assessRealizationInputs,
) (cqrs.CommandResult, error) {
	ta, err := aggregates.NewTimeAssessment(aggregates.TimeAssessmentFacts{
		CapabilityID:  inputs.capability,
		ComponentID:   inputs.component,
		RealizationID: realizationID,
		Grade:         inputs.grade,
		Rationale:     inputs.rationale,
		AssessedBy:    inputs.assessedBy,
	})
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, ta); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(ta.ID()), nil
}

func parseAssessRealizationInputs(command *commands.AssessRealization) (assessRealizationInputs, error) {
	capability, err := valueobjects.NewPhysicalCapabilityRef(command.CapabilityID)
	if err != nil {
		return assessRealizationInputs{}, err
	}
	component, err := valueobjects.NewApplicationRef(command.ComponentID)
	if err != nil {
		return assessRealizationInputs{}, err
	}
	grade, err := valueobjects.NewTimeGrade(command.Grade)
	if err != nil {
		return assessRealizationInputs{}, err
	}
	rationale, err := sharedvo.NewDescription(command.Rationale)
	if err != nil {
		return assessRealizationInputs{}, err
	}
	return assessRealizationInputs{
		capability: capability,
		component:  component,
		grade:      grade,
		rationale:  rationale,
		assessedBy: command.AssessedBy,
	}, nil
}
