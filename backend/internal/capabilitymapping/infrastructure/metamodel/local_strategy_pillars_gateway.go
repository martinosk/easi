package metamodel

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/readmodels"
)

type localStrategyPillarsGateway struct {
	cacheReadModel *readmodels.StrategyPillarCacheReadModel
}

func NewLocalStrategyPillarsGateway(cacheReadModel *readmodels.StrategyPillarCacheReadModel) StrategyPillarsGateway {
	return &localStrategyPillarsGateway{cacheReadModel: cacheReadModel}
}

func (g *localStrategyPillarsGateway) GetStrategyPillars(ctx context.Context) (*StrategyPillarsConfigDTO, error) {
	pillars, err := g.cacheReadModel.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(pillars) == 0 {
		return DefaultStrategyPillarsConfig(), nil
	}

	dtos := make([]StrategyPillarDTO, len(pillars))
	for i, pillar := range pillars {
		dtos[i] = StrategyPillarDTO{
			ID:                pillar.ID,
			Name:              pillar.Name,
			Description:       pillar.Description,
			Active:            pillar.Active,
			FitScoringEnabled: pillar.FitScoringEnabled,
			FitCriteria:       pillar.FitCriteria,
			FitType:           pillar.FitType,
		}
	}

	return &StrategyPillarsConfigDTO{Pillars: dtos}, nil
}

func (g *localStrategyPillarsGateway) GetActivePillar(ctx context.Context, pillarID string) (*StrategyPillarDTO, error) {
	pillar, err := g.cacheReadModel.GetActivePillar(ctx, pillarID)
	if err != nil {
		return nil, err
	}

	if pillar == nil {
		return nil, nil
	}

	return &StrategyPillarDTO{
		ID:                pillar.ID,
		Name:              pillar.Name,
		Description:       pillar.Description,
		Active:            pillar.Active,
		FitScoringEnabled: pillar.FitScoringEnabled,
		FitCriteria:       pillar.FitCriteria,
		FitType:           pillar.FitType,
	}, nil
}

func (g *localStrategyPillarsGateway) InvalidateCache(tenantID string) {
}
