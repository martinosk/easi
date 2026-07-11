package ports

import "context"

type SubjectExistenceChecker interface {
	SubjectExists(ctx context.Context, subjectType, subjectID string) (bool, error)
}
