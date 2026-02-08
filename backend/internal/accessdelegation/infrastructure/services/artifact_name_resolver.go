package services

import (
	"context"

	appservices "easi/backend/internal/accessdelegation/application/services"
	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	viewsReadModels "easi/backend/internal/architectureviews/application/readmodels"
	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
)

type artifactNameResolver struct {
	capabilities *capReadModels.CapabilityReadModel
	components   *archReadModels.ApplicationComponentReadModel
	views        *viewsReadModels.ArchitectureViewReadModel
}

func NewArtifactNameResolver(
	capabilities *capReadModels.CapabilityReadModel,
	components *archReadModels.ApplicationComponentReadModel,
	views *viewsReadModels.ArchitectureViewReadModel,
) appservices.ArtifactNameResolver {
	return &artifactNameResolver{
		capabilities: capabilities,
		components:   components,
		views:        views,
	}
}

func (r *artifactNameResolver) ResolveName(ctx context.Context, artifactType, artifactID string) (string, error) {
	var name string
	var err error

	switch artifactType {
	case "capability":
		name, err = r.resolveCapability(ctx, artifactID)
	case "component":
		name, err = r.resolveComponent(ctx, artifactID)
	case "view":
		name, err = r.resolveView(ctx, artifactID)
	}

	if err != nil || name == "" {
		return "Deleted artifact", nil
	}
	return name, nil
}

func (r *artifactNameResolver) resolveCapability(ctx context.Context, id string) (string, error) {
	dto, err := r.capabilities.GetByID(ctx, id)
	if err != nil || dto == nil {
		return "", err
	}
	return dto.Name, nil
}

func (r *artifactNameResolver) resolveComponent(ctx context.Context, id string) (string, error) {
	dto, err := r.components.GetByID(ctx, id)
	if err != nil || dto == nil {
		return "", err
	}
	return dto.Name, nil
}

func (r *artifactNameResolver) resolveView(ctx context.Context, id string) (string, error) {
	dto, err := r.views.GetByID(ctx, id)
	if err != nil || dto == nil {
		return "", err
	}
	return dto.Name, nil
}
