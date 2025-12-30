package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/metamodel"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

var (
	ErrPillarNotFound          = errors.New("pillar not found or inactive")
	ErrImportanceAlreadyExists = errors.New("importance rating already exists for this domain, capability, and pillar combination")
	ErrImportanceNotFound      = errors.New("strategy importance not found")
	ErrInvalidImportanceValue  = errors.New("importance must be between 1 and 5")
)

type StrategyImportanceDeps struct {
	ImportanceRepo   *repositories.StrategyImportanceRepository
	DomainReader     *readmodels.BusinessDomainReadModel
	CapabilityReader *readmodels.CapabilityReadModel
	ImportanceReader *readmodels.StrategyImportanceReadModel
	PillarsGateway   metamodel.StrategyPillarsGateway
}

type SetStrategyImportanceHandler struct {
	deps StrategyImportanceDeps
}

func NewSetStrategyImportanceHandler(deps StrategyImportanceDeps) *SetStrategyImportanceHandler {
	return &SetStrategyImportanceHandler{deps: deps}
}

func (h *SetStrategyImportanceHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.SetStrategyImportance)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	pillarName, err := h.validateAndGetPillarName(ctx, command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	aggregate, err := h.createAggregate(command, pillarName)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.deps.ImportanceRepo.Save(ctx, aggregate); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(aggregate.ID()), nil
}

func (h *SetStrategyImportanceHandler) createAggregate(cmd *commands.SetStrategyImportance, pillarName string) (*aggregates.StrategyImportance, error) {
	businessDomainID, err := valueobjects.NewBusinessDomainIDFromString(cmd.BusinessDomainID)
	if err != nil {
		return nil, err
	}

	capabilityID, err := valueobjects.NewCapabilityIDFromString(cmd.CapabilityID)
	if err != nil {
		return nil, err
	}

	pillarID, err := valueobjects.NewPillarIDFromString(cmd.PillarID)
	if err != nil {
		return nil, err
	}

	importance, err := valueobjects.NewImportance(cmd.Importance)
	if err != nil {
		return nil, ErrInvalidImportanceValue
	}

	rationale, err := valueobjects.NewRationale(cmd.Rationale)
	if err != nil {
		return nil, err
	}

	return aggregates.SetStrategyImportance(aggregates.NewImportanceParams{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capabilityID,
		PillarID:         pillarID,
		PillarName:       pillarName,
		Importance:       importance,
		Rationale:        rationale,
	})
}

func (h *SetStrategyImportanceHandler) validateAndGetPillarName(ctx context.Context, command *commands.SetStrategyImportance) (string, error) {
	if err := h.validateBusinessDomainExists(ctx, command.BusinessDomainID); err != nil {
		return "", err
	}

	if err := h.validateCapabilityExists(ctx, command.CapabilityID); err != nil {
		return "", err
	}

	pillarName, err := h.getPillarName(ctx, command.PillarID)
	if err != nil {
		return "", err
	}

	if err := h.validateImportanceDoesNotExist(ctx, command.BusinessDomainID, command.CapabilityID, command.PillarID); err != nil {
		return "", err
	}

	return pillarName, nil
}

func (h *SetStrategyImportanceHandler) validateBusinessDomainExists(ctx context.Context, domainID string) error {
	return validateEntityExists(ctx, domainID, h.deps.DomainReader.GetByID, ErrBusinessDomainNotFound)
}

func (h *SetStrategyImportanceHandler) validateCapabilityExists(ctx context.Context, capabilityID string) error {
	return validateEntityExists(ctx, capabilityID, h.deps.CapabilityReader.GetByID, ErrCapabilityNotFound)
}

func (h *SetStrategyImportanceHandler) getPillarName(ctx context.Context, pillarID string) (string, error) {
	if h.deps.PillarsGateway == nil {
		return "", nil
	}
	pillar, err := h.deps.PillarsGateway.GetActivePillar(ctx, pillarID)
	if err != nil {
		return "", err
	}
	if pillar == nil {
		return "", ErrPillarNotFound
	}
	return pillar.Name, nil
}

func (h *SetStrategyImportanceHandler) validateImportanceDoesNotExist(ctx context.Context, domainID, capabilityID, pillarID string) error {
	exists, err := h.deps.ImportanceReader.Exists(ctx, domainID, capabilityID, pillarID)
	if err != nil {
		return err
	}
	if exists {
		return ErrImportanceAlreadyExists
	}
	return nil
}

func validateEntityExists[T any](ctx context.Context, id string, getter func(context.Context, string) (*T, error), notFoundErr error) error {
	entity, err := getter(ctx, id)
	if err != nil {
		return err
	}
	if entity == nil {
		return notFoundErr
	}
	return nil
}
