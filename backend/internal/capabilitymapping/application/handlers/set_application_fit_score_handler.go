package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/shared/cqrs"
)

var (
	ErrFitScoreAlreadyExists    = errors.New("fit score already exists for this component and pillar combination")
	ErrFitScoreNotFound         = errors.New("application fit score not found")
	ErrInvalidFitScoreValue     = errors.New("fit score must be between 1 and 5")
	ErrPillarFitScoringDisabled = errors.New("fit scoring is not enabled for this pillar")
)

type ApplicationFitScoreDeps struct {
	FitScoreRepo   *repositories.ApplicationFitScoreRepository
	FitScoreReader *readmodels.ApplicationFitScoreReadModel
	PillarsGateway mmPL.StrategyPillarsGateway
}

type SetApplicationFitScoreHandler struct {
	deps ApplicationFitScoreDeps
}

func NewSetApplicationFitScoreHandler(deps ApplicationFitScoreDeps) *SetApplicationFitScoreHandler {
	return &SetApplicationFitScoreHandler{deps: deps}
}

func (h *SetApplicationFitScoreHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.SetApplicationFitScore)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	pillarNameStr, err := h.validateAndGetPillarName(ctx, command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	pillarName, err := valueobjects.NewPillarName(pillarNameStr)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	aggregate, err := h.createAggregate(command, pillarName)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.deps.FitScoreRepo.Save(ctx, aggregate); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(aggregate.ID()), nil
}

func (h *SetApplicationFitScoreHandler) createAggregate(cmd *commands.SetApplicationFitScore, pillarName valueobjects.PillarName) (*aggregates.ApplicationFitScore, error) {
	componentID, err := valueobjects.NewComponentIDFromString(cmd.ComponentID)
	if err != nil {
		return nil, err
	}

	pillarID, err := valueobjects.NewPillarIDFromString(cmd.PillarID)
	if err != nil {
		return nil, err
	}

	score, err := valueobjects.NewFitScore(cmd.Score)
	if err != nil {
		return nil, ErrInvalidFitScoreValue
	}

	rationale, err := valueobjects.NewFitRationale(cmd.Rationale)
	if err != nil {
		return nil, err
	}

	scoredBy, err := valueobjects.NewUserIdentifier(cmd.ScoredBy)
	if err != nil {
		return nil, err
	}

	return aggregates.SetApplicationFitScore(aggregates.NewFitScoreParams{
		ComponentID: componentID,
		PillarID:    pillarID,
		PillarName:  pillarName,
		Score:       score,
		Rationale:   rationale,
		ScoredBy:    scoredBy,
	})
}

func (h *SetApplicationFitScoreHandler) validateAndGetPillarName(ctx context.Context, command *commands.SetApplicationFitScore) (string, error) {
	pillarName, fitEnabled, err := h.getPillarNameAndFitEnabled(ctx, command.PillarID)
	if err != nil {
		return "", err
	}

	if !fitEnabled {
		return "", ErrPillarFitScoringDisabled
	}

	if err := h.validateFitScoreDoesNotExist(ctx, command.ComponentID, command.PillarID); err != nil {
		return "", err
	}

	return pillarName, nil
}

func (h *SetApplicationFitScoreHandler) getPillarNameAndFitEnabled(ctx context.Context, pillarID string) (string, bool, error) {
	if h.deps.PillarsGateway == nil {
		return "", true, nil
	}
	pillar, err := h.deps.PillarsGateway.GetActivePillar(ctx, pillarID)
	if err != nil {
		return "", false, err
	}
	if pillar == nil {
		return "", false, ErrPillarNotFound
	}
	return pillar.Name, pillar.FitScoringEnabled, nil
}

func (h *SetApplicationFitScoreHandler) validateFitScoreDoesNotExist(ctx context.Context, componentID, pillarID string) error {
	exists, err := h.deps.FitScoreReader.Exists(ctx, componentID, pillarID)
	if err != nil {
		return err
	}
	if exists {
		return ErrFitScoreAlreadyExists
	}
	return nil
}
