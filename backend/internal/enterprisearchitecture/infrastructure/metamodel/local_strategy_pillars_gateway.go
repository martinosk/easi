package metamodel

import (
	"context"

	cmMetamodel "easi/backend/internal/capabilitymapping/infrastructure/metamodel"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
)

type localStrategyPillarsGateway struct {
	cacheReadModel *readmodels.StrategyPillarCacheReadModel
}

func NewLocalStrategyPillarsGateway(cacheReadModel *readmodels.StrategyPillarCacheReadModel) cmMetamodel.StrategyPillarsGateway {
	return &localStrategyPillarsGateway{cacheReadModel: cacheReadModel}
}

func (g *localStrategyPillarsGateway) GetStrategyPillars(ctx context.Context) (*cmMetamodel.StrategyPillarsConfigDTO, error) {
	pillars, err := g.cacheReadModel.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(pillars) == 0 {
		return cmMetamodel.DefaultStrategyPillarsConfig(), nil
	}

	dtos := make([]cmMetamodel.StrategyPillarDTO, len(pillars))
	for i, pillar := range pillars {
		dtos[i] = cmMetamodel.StrategyPillarDTO{
			ID:                pillar.ID,
			Name:              pillar.Name,
			Description:       pillar.Description,
			Active:            pillar.Active,
			FitScoringEnabled: pillar.FitScoringEnabled,
			FitCriteria:       pillar.FitCriteria,
			FitType:           pillar.FitType,
		}
	}

	return &cmMetamodel.StrategyPillarsConfigDTO{Pillars: dtos}, nil
}

func (g *localStrategyPillarsGateway) GetActivePillar(ctx context.Context, pillarID string) (*cmMetamodel.StrategyPillarDTO, error) {
	pillar, err := g.cacheReadModel.GetActivePillar(ctx, pillarID)
	if err != nil {
		return nil, err
	}

	if pillar == nil {
		return nil, nil
	}

	return &cmMetamodel.StrategyPillarDTO{
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
