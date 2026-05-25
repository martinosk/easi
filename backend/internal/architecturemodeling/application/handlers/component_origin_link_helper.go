package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
)

type OriginLinkRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ComponentOriginLink, error)
}

func getOrCreateComponentOriginLink(
	ctx context.Context,
	repo OriginLinkRepository,
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
