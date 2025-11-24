package domain

import (
	"context"

	"easi/backend/internal/releases/domain/aggregates"
	"easi/backend/internal/releases/domain/valueobjects"
)

type ReleaseRepository interface {
	Save(ctx context.Context, release *aggregates.Release) error
	FindByVersion(ctx context.Context, version valueobjects.Version) (*aggregates.Release, error)
	FindLatest(ctx context.Context) (*aggregates.Release, error)
	FindAll(ctx context.Context) ([]*aggregates.Release, error)
}
