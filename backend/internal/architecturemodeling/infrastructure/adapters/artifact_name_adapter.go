package adapters

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/readmodels"
)

type ComponentNameAdapter struct {
	readModel *readmodels.ApplicationComponentReadModel
}

func NewComponentNameAdapter(rm *readmodels.ApplicationComponentReadModel) *ComponentNameAdapter {
	return &ComponentNameAdapter{readModel: rm}
}

func (a *ComponentNameAdapter) GetArtifactName(ctx context.Context, artifactID string) (string, error) {
	dto, err := a.readModel.GetByID(ctx, artifactID)
	if err != nil || dto == nil {
		return "", err
	}
	return dto.Name, nil
}

type VendorNameAdapter struct {
	readModel *readmodels.VendorReadModel
}

func NewVendorNameAdapter(rm *readmodels.VendorReadModel) *VendorNameAdapter {
	return &VendorNameAdapter{readModel: rm}
}

func (a *VendorNameAdapter) GetArtifactName(ctx context.Context, artifactID string) (string, error) {
	dto, err := a.readModel.GetByID(ctx, artifactID)
	if err != nil || dto == nil {
		return "", err
	}
	return dto.Name, nil
}

type AcquiredEntityNameAdapter struct {
	readModel *readmodels.AcquiredEntityReadModel
}

func NewAcquiredEntityNameAdapter(rm *readmodels.AcquiredEntityReadModel) *AcquiredEntityNameAdapter {
	return &AcquiredEntityNameAdapter{readModel: rm}
}

func (a *AcquiredEntityNameAdapter) GetArtifactName(ctx context.Context, artifactID string) (string, error) {
	dto, err := a.readModel.GetByID(ctx, artifactID)
	if err != nil || dto == nil {
		return "", err
	}
	return dto.Name, nil
}

type InternalTeamNameAdapter struct {
	readModel *readmodels.InternalTeamReadModel
}

func NewInternalTeamNameAdapter(rm *readmodels.InternalTeamReadModel) *InternalTeamNameAdapter {
	return &InternalTeamNameAdapter{readModel: rm}
}

func (a *InternalTeamNameAdapter) GetArtifactName(ctx context.Context, artifactID string) (string, error) {
	dto, err := a.readModel.GetByID(ctx, artifactID)
	if err != nil || dto == nil {
		return "", err
	}
	return dto.Name, nil
}
