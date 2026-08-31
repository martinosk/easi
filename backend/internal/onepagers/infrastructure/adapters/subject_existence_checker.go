package adapters

import (
	"context"
	"fmt"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/valueobjects"
)

type subjectExistenceChecker struct {
	subjects SubjectAttributeStore
}

func NewOnePagerSubjectExistenceChecker(db *database.TenantAwareDB) ports.SubjectExistenceChecker {
	return NewSubjectExistenceChecker(readmodels.NewOnePagerSubjectIndexReadModel(db))
}

func NewSubjectExistenceChecker(subjects SubjectAttributeStore) ports.SubjectExistenceChecker {
	return subjectExistenceChecker{subjects: subjects}
}

func (c subjectExistenceChecker) SubjectExists(ctx context.Context, subjectType, subjectID string) (bool, error) {
	if _, err := valueobjects.NewSubjectType(subjectType); err != nil {
		return false, fmt.Errorf("unknown one-pager subject type %q: %w", subjectType, err)
	}
	exists, err := c.subjects.Exists(ctx, readmodels.SubjectKey{SubjectType: subjectType, SubjectID: subjectID})
	if err != nil {
		return false, fmt.Errorf("check whether %s %s exists: %w", subjectType, subjectID, err)
	}
	return exists, nil
}
