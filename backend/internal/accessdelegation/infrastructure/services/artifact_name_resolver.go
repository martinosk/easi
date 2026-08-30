package services

import (
	"context"

	appservices "easi/backend/internal/accessdelegation/application/services"
)

const deletedArtifactName = "Deleted artifact"

type ArtifactNameCache interface {
	NamesByIDs(ctx context.Context, artifactType string, artifactIDs []string) (map[string]string, error)
}

type artifactNameResolver struct {
	cache ArtifactNameCache
}

func NewArtifactNameResolver(cache ArtifactNameCache) appservices.ArtifactNameResolver {
	return &artifactNameResolver{cache: cache}
}

func (r *artifactNameResolver) ResolveName(ctx context.Context, artifactType, artifactID string) (string, error) {
	names, err := r.cache.NamesByIDs(ctx, artifactType, []string{artifactID})
	if err != nil {
		return deletedArtifactName, nil
	}
	if name := names[artifactID]; name != "" {
		return name, nil
	}
	return deletedArtifactName, nil
}
