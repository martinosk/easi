package services

import (
	"context"

	appservices "easi/backend/internal/accessdelegation/application/services"
	"easi/backend/internal/accessdelegation/application/ports"
)

type artifactNameResolver struct {
	resolvers map[string]ports.ArtifactNameLookup
}

type ArtifactNameResolverDeps struct {
	Capabilities     ports.ArtifactNameLookup
	Components       ports.ArtifactNameLookup
	Views            ports.ArtifactNameLookup
	Domains          ports.ArtifactNameLookup
	Vendors          ports.ArtifactNameLookup
	AcquiredEntities ports.ArtifactNameLookup
	InternalTeams    ports.ArtifactNameLookup
}

func NewArtifactNameResolver(deps ArtifactNameResolverDeps) appservices.ArtifactNameResolver {
	return &artifactNameResolver{
		resolvers: map[string]ports.ArtifactNameLookup{
			"capability":      deps.Capabilities,
			"component":       deps.Components,
			"view":            deps.Views,
			"domain":          deps.Domains,
			"vendor":          deps.Vendors,
			"acquired_entity": deps.AcquiredEntities,
			"internal_team":   deps.InternalTeams,
		},
	}
}

func (r *artifactNameResolver) ResolveName(ctx context.Context, artifactType, artifactID string) (string, error) {
	lookup, ok := r.resolvers[artifactType]
	if !ok {
		return "Deleted artifact", nil
	}

	name, err := lookup.GetArtifactName(ctx, artifactID)
	if err != nil || name == "" {
		return "Deleted artifact", nil
	}
	return name, nil
}
