package gateways

import (
	"context"

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
