package handlers

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

type DeleteImpactResult struct {
	HasDescendants                     bool
	AffectedCapabilities               []string
	RealizationsOnDeletedCapabilities  []readmodels.RealizationDTO
	RealizationsOnRetainedCapabilities []readmodels.RealizationDTO
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

	allRealizations, err := q.collectRealizations(ctx, scope)
	if err != nil {
		return nil, err
	}

	deletable, retained, err := q.classifyRealizations(ctx, allRealizations, scope)
	if err != nil {
		return nil, err
	}

	affectedIDs := make([]string, len(descendants))
	for i, d := range descendants {
		affectedIDs[i] = d.Value()
	}

	return &DeleteImpactResult{
		HasDescendants:                     len(descendants) > 0,
		AffectedCapabilities:               affectedIDs,
		RealizationsOnDeletedCapabilities:  deletable,
		RealizationsOnRetainedCapabilities: retained,
	}, nil
}

func (q *DeleteImpactQuery) collectRealizations(ctx context.Context, scope valueobjects.DeletionScope) ([]readmodels.RealizationDTO, error) {
	var all []readmodels.RealizationDTO
	for _, capID := range scope.AllIDs() {
		realizations, err := q.realizationRM.GetByCapabilityID(ctx, capID.Value())
		if err != nil {
			return nil, err
		}
		all = append(all, realizations...)
	}
	return all, nil
}

func (q *DeleteImpactQuery) classifyRealizations(ctx context.Context, realizations []readmodels.RealizationDTO, scope valueobjects.DeletionScope) (deletable, retained []readmodels.RealizationDTO, err error) {
	classified := make(map[string]bool)

	for _, r := range realizations {
		if exclusive, done := classified[r.ComponentID]; done {
			if exclusive {
				deletable = append(deletable, r)
			} else {
				retained = append(retained, r)
			}
			continue
		}

		exclusive, err := isComponentExclusiveToScope(ctx, q.realizationRM, r.ComponentID, scope)
		if err != nil {
			return nil, nil, err
		}

		classified[r.ComponentID] = exclusive
		if exclusive {
			deletable = append(deletable, r)
		} else {
			retained = append(retained, r)
		}
	}

	return deletable, retained, nil
}
