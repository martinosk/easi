package adapters

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

type CapabilityLookupAdapter struct {
	readModel *readmodels.CapabilityReadModel
}

func NewCapabilityLookupAdapter(readModel *readmodels.CapabilityReadModel) *CapabilityLookupAdapter {
	return &CapabilityLookupAdapter{readModel: readModel}
}

func (a *CapabilityLookupAdapter) GetCapabilityInfo(ctx context.Context, id valueobjects.CapabilityID) (*services.CapabilityInfo, error) {
	dto, err := a.readModel.GetByID(ctx, id.Value())
	if err != nil {
		return nil, err
	}
	if dto == nil {
		return nil, nil
	}

	level, err := valueobjects.NewCapabilityLevel(dto.Level)
	if err != nil {
		return nil, err
	}

	var parentID valueobjects.CapabilityID
	if dto.ParentID != "" {
		parentID, err = valueobjects.NewCapabilityIDFromString(dto.ParentID)
		if err != nil {
			return nil, err
		}
	}

	return &services.CapabilityInfo{
		ID:       id,
		Level:    level,
		ParentID: parentID,
	}, nil
}

func (a *CapabilityLookupAdapter) GetChildren(ctx context.Context, parentID valueobjects.CapabilityID) ([]valueobjects.CapabilityID, error) {
	children, err := a.readModel.GetChildren(ctx, parentID.Value())
	if err != nil {
		return nil, err
	}

	result := make([]valueobjects.CapabilityID, 0, len(children))
	for _, child := range children {
		childID, err := valueobjects.NewCapabilityIDFromString(child.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, childID)
	}

	return result, nil
}
