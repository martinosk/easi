package adapters

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/readmodels"
)

type CapabilityNameAdapter struct {
	readModel *readmodels.CapabilityReadModel
}

func NewCapabilityNameAdapter(rm *readmodels.CapabilityReadModel) *CapabilityNameAdapter {
	return &CapabilityNameAdapter{readModel: rm}
}

func (a *CapabilityNameAdapter) GetArtifactName(ctx context.Context, artifactID string) (string, error) {
	dto, err := a.readModel.GetByID(ctx, artifactID)
	if err != nil || dto == nil {
		return "", err
	}
	return dto.Name, nil
}

type DomainNameAdapter struct {
	readModel *readmodels.BusinessDomainReadModel
}

func NewDomainNameAdapter(rm *readmodels.BusinessDomainReadModel) *DomainNameAdapter {
	return &DomainNameAdapter{readModel: rm}
}

func (a *DomainNameAdapter) GetArtifactName(ctx context.Context, artifactID string) (string, error) {
	dto, err := a.readModel.GetByID(ctx, artifactID)
	if err != nil || dto == nil {
		return "", err
	}
	return dto.Name, nil
}
