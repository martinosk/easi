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

type StandardApplicationRepository interface {
	Save(ctx context.Context, sa *aggregates.StandardApplication) error
	GetByID(ctx context.Context, id string) (*aggregates.StandardApplication, error)
}

type ExistingStandardApplicationLookup interface {
	FindAggregateIDForEnterpriseCapability(ctx context.Context, ecID string) (string, bool, error)
}

type SetStandardApplicationHandler struct {
	repo   StandardApplicationRepository
	lookup ExistingStandardApplicationLookup
	refs   *services.ReferenceChecker
}

func NewSetStandardApplicationHandler(
	repo StandardApplicationRepository,
	lookup ExistingStandardApplicationLookup,
	refs *services.ReferenceChecker,
) *SetStandardApplicationHandler {
	return &SetStandardApplicationHandler{repo: repo, lookup: lookup, refs: refs}
}

type setStandardApplicationInputs struct {
	ec        valueobjects.EnterpriseCapabilityRef
	app       valueobjects.ApplicationRef
	narrative sharedvo.Description
}

func (h *SetStandardApplicationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.SetStandardApplication)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	inputs, err := parseSetStandardApplicationInputs(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.verifyEnterpriseCapabilityExists(ctx, inputs.ec); err != nil {
		return cqrs.EmptyResult(), err
	}
	existingID, exists, err := h.lookup.FindAggregateIDForEnterpriseCapability(ctx, inputs.ec.Value())
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if exists {
		return h.changeExisting(ctx, existingID, inputs)
	}
	return h.createNew(ctx, inputs)
}

func (h *SetStandardApplicationHandler) verifyEnterpriseCapabilityExists(ctx context.Context, ec valueobjects.EnterpriseCapabilityRef) error {
	exists, err := h.refs.EnterpriseCapabilityExists(ctx, ec.Value())
	if err != nil {
		return err
	}
	if !exists {
		return services.ErrReferencedEntityNotFound
	}
	return nil
}

func (h *SetStandardApplicationHandler) changeExisting(
	ctx context.Context,
	existingID string,
	inputs setStandardApplicationInputs,
) (cqrs.CommandResult, error) {
	existing, err := h.repo.GetByID(ctx, existingID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := existing.Change(inputs.app, inputs.narrative); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, existing); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(existing.ID()), nil
}

func (h *SetStandardApplicationHandler) createNew(
	ctx context.Context,
	inputs setStandardApplicationInputs,
) (cqrs.CommandResult, error) {
	sa, err := aggregates.NewStandardApplication(inputs.ec, inputs.app, inputs.narrative)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, sa); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(sa.ID()), nil
}

func parseSetStandardApplicationInputs(command *commands.SetStandardApplication) (setStandardApplicationInputs, error) {
	ec, err := valueobjects.NewEnterpriseCapabilityRef(command.EnterpriseCapabilityID)
	if err != nil {
		return setStandardApplicationInputs{}, err
	}
	app, err := valueobjects.NewApplicationRef(command.ApplicationID)
	if err != nil {
		return setStandardApplicationInputs{}, err
	}
	narrative, err := sharedvo.NewDescription(command.Narrative)
	if err != nil {
		return setStandardApplicationInputs{}, err
	}
	return setStandardApplicationInputs{ec: ec, app: app, narrative: narrative}, nil
}
