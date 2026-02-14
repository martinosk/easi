package publishedlanguage

import "context"

type StrategyPillarDTO struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Active            bool   `json:"active"`
	FitScoringEnabled bool   `json:"fitScoringEnabled"`
	FitCriteria       string `json:"fitCriteria"`
	FitType           string `json:"fitType"`
}

type StrategyPillarsConfigDTO struct {
	Pillars []StrategyPillarDTO `json:"data"`
}

type StrategyPillarsGateway interface {
	GetStrategyPillars(ctx context.Context) (*StrategyPillarsConfigDTO, error)
	GetActivePillar(ctx context.Context, pillarID string) (*StrategyPillarDTO, error)
	InvalidateCache(tenantID string)
}

func DefaultStrategyPillarsConfig() *StrategyPillarsConfigDTO {
	return &StrategyPillarsConfigDTO{
		Pillars: []StrategyPillarDTO{
			{ID: "default-always-on", Name: "Always On", Description: "Core capabilities that must always be operational", Active: true},
			{ID: "default-grow", Name: "Grow", Description: "Capabilities driving business growth", Active: true},
			{ID: "default-transform", Name: "Transform", Description: "Capabilities enabling digital transformation", Active: true},
		},
	}
}
