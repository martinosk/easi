package adapters

import (
	"context"

	"easi/backend/internal/architectureviews/application/readmodels"
)

type ViewNameAdapter struct {
	readModel *readmodels.ArchitectureViewReadModel
}

func NewViewNameAdapter(rm *readmodels.ArchitectureViewReadModel) *ViewNameAdapter {
	return &ViewNameAdapter{readModel: rm}
}

func (a *ViewNameAdapter) GetArtifactName(ctx context.Context, artifactID string) (string, error) {
	dto, err := a.readModel.GetByID(ctx, artifactID)
	if err != nil || dto == nil {
		return "", err
	}
	return dto.Name, nil
}
