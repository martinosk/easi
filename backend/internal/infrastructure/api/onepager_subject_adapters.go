package api

import (
	"context"
	"fmt"

	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
	eaReadModels "easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/infrastructure/database"
)

type subjectExistenceCheck func(ctx context.Context, id string) (bool, error)

type onePagerSubjectExistenceAdapter struct {
	checks map[string]subjectExistenceCheck
}

func newOnePagerSubjectExistenceAdapter(db *database.TenantAwareDB) onePagerSubjectExistenceAdapter {
	return onePagerSubjectExistenceAdapter{checks: map[string]subjectExistenceCheck{
		"capability":            subjectExists(capReadModels.NewCapabilityReadModel(db).GetByID),
		"enterprise-capability": subjectExists(eaReadModels.NewEnterpriseCapabilityReadModel(db).GetByID),
		"application":           subjectExists(archReadModels.NewApplicationComponentReadModel(db).GetByID),
		"acquired-entity":       subjectExists(archReadModels.NewAcquiredEntityReadModel(db).GetByID),
		"vendor":                subjectExists(archReadModels.NewVendorReadModel(db).GetByID),
		"internal-team":         subjectExists(archReadModels.NewInternalTeamReadModel(db).GetByID),
	}}
}

func (a onePagerSubjectExistenceAdapter) SubjectExists(ctx context.Context, subjectType, subjectID string) (bool, error) {
	check, found := a.checks[subjectType]
	if !found {
		return false, fmt.Errorf("unknown one-pager subject type %q", subjectType)
	}
	return check(ctx, subjectID)
}

func subjectExists[T any](getByID func(context.Context, string) (*T, error)) subjectExistenceCheck {
	return func(ctx context.Context, id string) (bool, error) {
		dto, err := getByID(ctx, id)
		if err != nil {
			return false, err
		}
		return dto != nil, nil
	}
}
