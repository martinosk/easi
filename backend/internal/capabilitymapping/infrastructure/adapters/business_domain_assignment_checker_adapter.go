package adapters

import (
	"context"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
)

type BusinessDomainAssignmentCheckerAdapter struct {
	readModel *readmodels.DomainCapabilityAssignmentReadModel
}

func NewBusinessDomainAssignmentCheckerAdapter(readModel *readmodels.DomainCapabilityAssignmentReadModel) *BusinessDomainAssignmentCheckerAdapter {
	return &BusinessDomainAssignmentCheckerAdapter{readModel: readModel}
}

func (a *BusinessDomainAssignmentCheckerAdapter) HasAssignments(ctx context.Context, domainID valueobjects.BusinessDomainID) (bool, error) {
	assignments, err := a.readModel.GetByDomainID(ctx, domainID.Value())
	if err != nil {
		return false, err
	}
	return len(assignments) > 0, nil
}
