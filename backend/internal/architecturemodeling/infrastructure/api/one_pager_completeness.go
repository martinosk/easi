package api

import (
	"context"

	"easi/backend/internal/architecturemodeling/application/readmodels"
)

type OnePagerCompletenessSource interface {
	CompletenessFor(ctx context.Context, subjectIDs []string) (map[string]bool, bool, error)
}

type OnePagerCompletenessSources struct {
	Components       OnePagerCompletenessSource
	AcquiredEntities OnePagerCompletenessSource
	Vendors          OnePagerCompletenessSource
	InternalTeams    OnePagerCompletenessSource
}

func decorateOnePagerCompleteness[T any](ctx context.Context, source OnePagerCompletenessSource, rows []T, subjectID func(*T) string, apply func(*T, bool)) error {
	if source == nil || len(rows) == 0 {
		return nil
	}
	ids := make([]string, len(rows))
	for i := range rows {
		ids[i] = subjectID(&rows[i])
	}
	indicators, present, err := source.CompletenessFor(ctx, ids)
	if err != nil {
		return err
	}
	if !present {
		return nil
	}
	for i := range rows {
		apply(&rows[i], indicators[ids[i]])
	}
	return nil
}

func decorateComponentsOnePagerCompleteness(ctx context.Context, source OnePagerCompletenessSource, rows []readmodels.ApplicationComponentDTO) error {
	return decorateOnePagerCompleteness(ctx, source, rows,
		func(row *readmodels.ApplicationComponentDTO) string { return row.ID },
		func(row *readmodels.ApplicationComponentDTO, complete bool) { row.OnePagerComplete = &complete })
}

func decorateAcquiredEntitiesOnePagerCompleteness(ctx context.Context, source OnePagerCompletenessSource, rows []readmodels.AcquiredEntityDTO) error {
	return decorateOnePagerCompleteness(ctx, source, rows,
		func(row *readmodels.AcquiredEntityDTO) string { return row.ID },
		func(row *readmodels.AcquiredEntityDTO, complete bool) { row.OnePagerComplete = &complete })
}

func decorateVendorsOnePagerCompleteness(ctx context.Context, source OnePagerCompletenessSource, rows []readmodels.VendorDTO) error {
	return decorateOnePagerCompleteness(ctx, source, rows,
		func(row *readmodels.VendorDTO) string { return row.ID },
		func(row *readmodels.VendorDTO, complete bool) { row.OnePagerComplete = &complete })
}

func decorateInternalTeamsOnePagerCompleteness(ctx context.Context, source OnePagerCompletenessSource, rows []readmodels.InternalTeamDTO) error {
	return decorateOnePagerCompleteness(ctx, source, rows,
		func(row *readmodels.InternalTeamDTO) string { return row.ID },
		func(row *readmodels.InternalTeamDTO, complete bool) { row.OnePagerComplete = &complete })
}
