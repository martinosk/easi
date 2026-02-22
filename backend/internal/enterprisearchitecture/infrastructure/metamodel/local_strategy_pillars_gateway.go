package metamodel

import (
	"context"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
)

type localStrategyPillarsGateway struct {
	cacheReadModel *readmodels.StrategyPillarCacheReadModel
}

func NewLocalStrategyPillarsGateway(cacheReadModel *readmodels.StrategyPillarCacheReadModel) mmPL.StrategyPillarsGateway {
	return &localStrategyPillarsGateway{cacheReadModel: cacheReadModel}
}

func (g *localStrategyPillarsGateway) GetStrategyPillars(ctx context.Context) (*mmPL.StrategyPillarsConfigDTO, error) {
	pillars, err := g.cacheReadModel.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(pillars) == 0 {
		return mmPL.DefaultStrategyPillarsConfig(), nil
	}

	dtos := make([]mmPL.StrategyPillarDTO, len(pillars))
	for i, pillar := range pillars {
		dtos[i] = mmPL.StrategyPillarDTO{
			ID:                pillar.ID,
			Name:              pillar.Name,
			Description:       pillar.Description,
			Active:            pillar.Active,
			FitScoringEnabled: pillar.FitScoringEnabled,
			FitCriteria:       pillar.FitCriteria,
			FitType:           pillar.FitType,
		}
	}

	return &mmPL.StrategyPillarsConfigDTO{Pillars: dtos}, nil
}

func (g *localStrategyPillarsGateway) GetActivePillar(ctx context.Context, pillarID string) (*mmPL.StrategyPillarDTO, error) {
	pillar, err := g.cacheReadModel.GetActivePillar(ctx, pillarID)
	if err != nil {
		return nil, err
	}

	if pillar == nil {
		return nil, nil
	}

	return &mmPL.StrategyPillarDTO{
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
	// No-op: cache is event-sourced
}
