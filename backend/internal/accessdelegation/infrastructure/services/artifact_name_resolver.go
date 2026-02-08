package services

import (
	"context"

	appservices "easi/backend/internal/accessdelegation/application/services"
	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	viewsReadModels "easi/backend/internal/architectureviews/application/readmodels"
	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
)

type nameResolveFn func(ctx context.Context, id string) (string, error)

type artifactNameResolver struct {
	resolvers map[string]nameResolveFn
}

type ArtifactNameResolverDeps struct {
	Capabilities     *capReadModels.CapabilityReadModel
	Components       *archReadModels.ApplicationComponentReadModel
	Views            *viewsReadModels.ArchitectureViewReadModel
	Domains          *capReadModels.BusinessDomainReadModel
	Vendors          *archReadModels.VendorReadModel
	AcquiredEntities *archReadModels.AcquiredEntityReadModel
	InternalTeams    *archReadModels.InternalTeamReadModel
}

func NewArtifactNameResolver(deps ArtifactNameResolverDeps) appservices.ArtifactNameResolver {
	return &artifactNameResolver{
		resolvers: buildResolverMap(deps),
	}
}

func buildResolverMap(deps ArtifactNameResolverDeps) map[string]nameResolveFn {
	return map[string]nameResolveFn{
		"capability":      capabilityResolver(deps.Capabilities),
		"component":       componentResolver(deps.Components),
		"view":            viewResolver(deps.Views),
		"domain":          domainResolver(deps.Domains),
		"vendor":          vendorResolver(deps.Vendors),
		"acquired_entity": acquiredEntityResolver(deps.AcquiredEntities),
		"internal_team":   internalTeamResolver(deps.InternalTeams),
	}
}

func capabilityResolver(rm *capReadModels.CapabilityReadModel) nameResolveFn {
	return func(ctx context.Context, id string) (string, error) {
		dto, err := rm.GetByID(ctx, id)
		if err != nil || dto == nil {
			return "", err
		}
		return dto.Name, nil
	}
}

func componentResolver(rm *archReadModels.ApplicationComponentReadModel) nameResolveFn {
	return func(ctx context.Context, id string) (string, error) {
		dto, err := rm.GetByID(ctx, id)
		if err != nil || dto == nil {
			return "", err
		}
		return dto.Name, nil
	}
}

func viewResolver(rm *viewsReadModels.ArchitectureViewReadModel) nameResolveFn {
	return func(ctx context.Context, id string) (string, error) {
		dto, err := rm.GetByID(ctx, id)
		if err != nil || dto == nil {
			return "", err
		}
		return dto.Name, nil
	}
}

func domainResolver(rm *capReadModels.BusinessDomainReadModel) nameResolveFn {
	return func(ctx context.Context, id string) (string, error) {
		dto, err := rm.GetByID(ctx, id)
		if err != nil || dto == nil {
			return "", err
		}
		return dto.Name, nil
	}
}

func vendorResolver(rm *archReadModels.VendorReadModel) nameResolveFn {
	return func(ctx context.Context, id string) (string, error) {
		dto, err := rm.GetByID(ctx, id)
		if err != nil || dto == nil {
			return "", err
		}
		return dto.Name, nil
	}
}

func acquiredEntityResolver(rm *archReadModels.AcquiredEntityReadModel) nameResolveFn {
	return func(ctx context.Context, id string) (string, error) {
		dto, err := rm.GetByID(ctx, id)
		if err != nil || dto == nil {
			return "", err
		}
		return dto.Name, nil
	}
}

func internalTeamResolver(rm *archReadModels.InternalTeamReadModel) nameResolveFn {
	return func(ctx context.Context, id string) (string, error) {
		dto, err := rm.GetByID(ctx, id)
		if err != nil || dto == nil {
			return "", err
		}
		return dto.Name, nil
	}
}

func (r *artifactNameResolver) ResolveName(ctx context.Context, artifactType, artifactID string) (string, error) {
	resolve, ok := r.resolvers[artifactType]
	if !ok {
		return "Deleted artifact", nil
	}

	name, err := resolve(ctx, artifactID)
	if err != nil || name == "" {
		return "Deleted artifact", nil
	}
	return name, nil
}
