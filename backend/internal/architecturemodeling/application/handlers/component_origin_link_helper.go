package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
)

func getOrCreateComponentOriginLink(
	ctx context.Context,
	repo *repositories.ComponentOriginLinkRepository,
	componentID valueobjects.ComponentID,
	originType valueobjects.OriginType,
) (*aggregates.ComponentOriginLink, error) {
	aggregateID := aggregates.BuildOriginLinkAggregateID(originType.String(), componentID.String())
	link, err := repo.GetByID(ctx, aggregateID)
	if err == nil {
		return link, nil
	}
	if err != repositories.ErrComponentOriginLinkNotFound {
		return nil, err
	}
	return aggregates.NewComponentOriginLink(componentID, originType)
}
