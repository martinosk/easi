package gateways

import (
	"context"

	appGateways "easi/backend/internal/valuestreams/application/gateways"
	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
)

type CapabilityGatewayImpl struct {
	capabilityReadModel *capReadModels.CapabilityReadModel
}

func NewCapabilityGateway(capabilityReadModel *capReadModels.CapabilityReadModel) *CapabilityGatewayImpl {
	return &CapabilityGatewayImpl{capabilityReadModel: capabilityReadModel}
}

func (g *CapabilityGatewayImpl) CapabilityExists(ctx context.Context, capabilityID string) (bool, error) {
	dto, err := g.capabilityReadModel.GetByID(ctx, capabilityID)
	if err != nil {
		return false, err
	}
	return dto != nil, nil
}

func (g *CapabilityGatewayImpl) GetCapability(ctx context.Context, capabilityID string) (*appGateways.CapabilityInfo, error) {
	dto, err := g.capabilityReadModel.GetByID(ctx, capabilityID)
	if err != nil {
		return nil, err
	}
	if dto == nil {
		return nil, nil
	}
	return &appGateways.CapabilityInfo{ID: dto.ID, Name: dto.Name}, nil
}
