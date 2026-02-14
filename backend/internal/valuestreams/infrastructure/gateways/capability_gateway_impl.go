package gateways

import (
	"context"

	appGateways "easi/backend/internal/valuestreams/application/gateways"
	"easi/backend/internal/valuestreams/application/readmodels"
)

type CapabilityGatewayImpl struct {
	cache *readmodels.CapabilityCacheReadModel
}

func NewCapabilityGateway(cache *readmodels.CapabilityCacheReadModel) *CapabilityGatewayImpl {
	return &CapabilityGatewayImpl{cache: cache}
}

func (g *CapabilityGatewayImpl) CapabilityExists(ctx context.Context, capabilityID string) (bool, error) {
	dto, err := g.cache.GetByID(ctx, capabilityID)
	if err != nil {
		return false, err
	}
	return dto != nil, nil
}

func (g *CapabilityGatewayImpl) GetCapability(ctx context.Context, capabilityID string) (*appGateways.CapabilityInfo, error) {
	dto, err := g.cache.GetByID(ctx, capabilityID)
	if err != nil {
		return nil, err
	}
	if dto == nil {
		return nil, nil
	}
	return &appGateways.CapabilityInfo{ID: dto.ID, Name: dto.Name}, nil
}
