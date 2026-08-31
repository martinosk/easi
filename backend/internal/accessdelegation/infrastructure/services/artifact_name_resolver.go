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
	names, err := r.ResolveNames(ctx, artifactType, []string{artifactID})
	if err != nil {
		return deletedArtifactName, nil
	}
	return names[artifactID], nil
}

func (r *artifactNameResolver) ResolveNames(ctx context.Context, artifactType string, artifactIDs []string) (map[string]string, error) {
	cached, err := r.cache.NamesByIDs(ctx, artifactType, artifactIDs)
	if err != nil {
		cached = nil
	}
	resolved := make(map[string]string, len(artifactIDs))
	for _, artifactID := range artifactIDs {
		if name := cached[artifactID]; name != "" {
			resolved[artifactID] = name
			continue
		}
		resolved[artifactID] = deletedArtifactName
	}
	return resolved, nil
}
