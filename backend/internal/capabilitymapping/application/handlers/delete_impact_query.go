package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

type DeleteImpactResult struct {
	HasDescendants       bool
	AffectedCapabilities []string
	Realizations         []readmodels.RealizationDTO
}

type DeleteImpactQuery struct {
	hierarchyService CascadeHierarchyService
	realizationRM    CascadeRealizationReadModel
}

func NewDeleteImpactQuery(
	hierarchyService CascadeHierarchyService,
	realizationRM CascadeRealizationReadModel,
) *DeleteImpactQuery {
	return &DeleteImpactQuery{
		hierarchyService: hierarchyService,
		realizationRM:    realizationRM,
	}
}

func (q *DeleteImpactQuery) Execute(ctx context.Context, capabilityID string) (*DeleteImpactResult, error) {
	capID, err := valueobjects.NewCapabilityIDFromString(capabilityID)
	if err != nil {
		return nil, err
	}

	descendants, err := q.hierarchyService.GetDescendants(ctx, capID)
	if err != nil {
		return nil, err
	}

	scope := valueobjects.NewDeletionScope(capID, descendants)

	realizations, err := collectDirectRealizations(ctx, q.realizationRM, scope)
	if err != nil {
		return nil, err
	}

	affectedIDs := make([]string, len(descendants))
	for i, d := range descendants {
		affectedIDs[i] = d.Value()
	}

	return &DeleteImpactResult{
		HasDescendants:       len(descendants) > 0,
		AffectedCapabilities: affectedIDs,
		Realizations:         realizations,
	}, nil
}
