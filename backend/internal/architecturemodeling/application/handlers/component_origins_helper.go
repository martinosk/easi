package handlers

import (
	"context"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
)

func getOrCreateComponentOrigins(
	ctx context.Context,
	repo *repositories.ComponentOriginsRepository,
	componentID valueobjects.ComponentID,
) (*aggregates.ComponentOrigins, error) {
	aggregateID := "component-origins:" + componentID.String()
	origins, err := repo.GetByID(ctx, aggregateID)
	if err == nil {
		return origins, nil
	}
	if err != repositories.ErrComponentOriginsNotFound {
		return nil, err
	}
	return aggregates.NewComponentOrigins(componentID)
}
