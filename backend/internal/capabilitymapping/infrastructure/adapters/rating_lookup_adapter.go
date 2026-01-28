package adapters

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

type RatingLookupAdapter struct {
	importanceReadModel *readmodels.StrategyImportanceReadModel
}

func NewRatingLookupAdapter(importanceReadModel *readmodels.StrategyImportanceReadModel) *RatingLookupAdapter {
	return &RatingLookupAdapter{importanceReadModel: importanceReadModel}
}

func (a *RatingLookupAdapter) GetRating(ctx context.Context, capabilityID valueobjects.CapabilityID, pillarID valueobjects.PillarID, businessDomainID valueobjects.BusinessDomainID) (*services.RatingInfo, error) {
	ratings, err := a.importanceReadModel.GetByDomainAndCapability(ctx, businessDomainID.Value(), capabilityID.Value())
	if err != nil {
		return nil, err
	}

	for _, rating := range ratings {
		if rating.PillarID == pillarID.Value() {
			importance, err := valueobjects.NewImportance(rating.Importance)
			if err != nil {
				return nil, err
			}

			capID, err := valueobjects.NewCapabilityIDFromString(rating.CapabilityID)
			if err != nil {
				return nil, err
			}

			return &services.RatingInfo{
				Importance:     importance,
				CapabilityID:   capID,
				CapabilityName: rating.CapabilityName,
				Rationale:      rating.Rationale,
			}, nil
		}
	}

	return nil, nil
}
