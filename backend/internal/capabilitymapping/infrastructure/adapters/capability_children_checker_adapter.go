package adapters

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

type CapabilityChildrenCheckerAdapter struct {
	readModel *readmodels.CapabilityReadModel
}

func NewCapabilityChildrenCheckerAdapter(readModel *readmodels.CapabilityReadModel) *CapabilityChildrenCheckerAdapter {
	return &CapabilityChildrenCheckerAdapter{readModel: readModel}
}

func (a *CapabilityChildrenCheckerAdapter) HasChildren(ctx context.Context, capabilityID valueobjects.CapabilityID) (bool, error) {
	children, err := a.readModel.GetChildren(ctx, capabilityID.Value())
	if err != nil {
		return false, err
	}
	return len(children) > 0, nil
}
