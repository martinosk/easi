package services

import (
	"context"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

type RatingInfo struct {
	Importance     valueobjects.Importance
	CapabilityID   valueobjects.CapabilityID
	CapabilityName string
	Rationale      string
}

type RatingLookup interface {
	GetRating(ctx context.Context, capabilityID valueobjects.CapabilityID, pillarID valueobjects.PillarID, businessDomainID valueobjects.BusinessDomainID) (*RatingInfo, error)
}

type ResolvedRating struct {
	EffectiveImportance  valueobjects.EffectiveImportance
	SourceCapabilityName string
	Rationale            string
}

type HierarchicalRatingResolver interface {
	ResolveEffectiveImportance(
		ctx context.Context,
		capabilityID valueobjects.CapabilityID,
		pillarID valueobjects.PillarID,
		businessDomainID valueobjects.BusinessDomainID,
	) (*ResolvedRating, error)
}

type hierarchicalRatingResolver struct {
	hierarchyService CapabilityHierarchyService
	ratingLookup     RatingLookup
	capabilityLookup CapabilityLookup
}

func NewHierarchicalRatingResolver(
	hierarchyService CapabilityHierarchyService,
	ratingLookup RatingLookup,
	capabilityLookup CapabilityLookup,
) HierarchicalRatingResolver {
	return &hierarchicalRatingResolver{
		hierarchyService: hierarchyService,
		ratingLookup:     ratingLookup,
		capabilityLookup: capabilityLookup,
	}
}

func (r *hierarchicalRatingResolver) ResolveEffectiveImportance(
	ctx context.Context,
	capabilityID valueobjects.CapabilityID,
	pillarID valueobjects.PillarID,
	businessDomainID valueobjects.BusinessDomainID,
) (*ResolvedRating, error) {
	currentID := capabilityID
	isInherited := false
	visited := make(map[string]bool)

	for {
		if visited[currentID.Value()] {
			return nil, ErrWouldCreateCircularHierarchy
		}
		visited[currentID.Value()] = true

		rating, err := r.ratingLookup.GetRating(ctx, currentID, pillarID, businessDomainID)
		if err != nil {
			return nil, err
		}

		if rating != nil {
			effectiveImportance := valueobjects.NewEffectiveImportance(
				rating.Importance,
				rating.CapabilityID,
				isInherited,
			)
			return &ResolvedRating{
				EffectiveImportance:  effectiveImportance,
				SourceCapabilityName: rating.CapabilityName,
				Rationale:            rating.Rationale,
			}, nil
		}

		info, err := r.capabilityLookup.GetCapabilityInfo(ctx, currentID)
		if err != nil {
			return nil, err
		}
		if info == nil {
			return nil, ErrCapabilityNotFound
		}

		if info.ParentID.Value() == "" {
			return nil, nil
		}

		currentID = info.ParentID
		isInherited = true
	}
}
