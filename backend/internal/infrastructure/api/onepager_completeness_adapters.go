package api

import (
	"context"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/application/readmodels"
)

type onePagerCompletenessAdapter struct {
	indicators  *queries.CompletenessIndicators
	subjectType string
}

func (a onePagerCompletenessAdapter) CompletenessFor(ctx context.Context, subjectIDs []string) (map[string]bool, bool, error) {
	return a.indicators.ForSubjects(ctx, a.subjectType, subjectIDs)
}

func newOnePagerCompletenessIndicators(db *database.TenantAwareDB) *queries.CompletenessIndicators {
	return queries.NewCompletenessIndicators(
		readmodels.NewOnePagerConfigurationReadModel(db),
		readmodels.NewOnePagerFactsReadModel(db),
	)
}

func onePagerCompletenessFor(indicators *queries.CompletenessIndicators, subjectType string) onePagerCompletenessAdapter {
	return onePagerCompletenessAdapter{indicators: indicators, subjectType: subjectType}
}
