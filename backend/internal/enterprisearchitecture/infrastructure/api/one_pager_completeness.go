package api

import (
	"context"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
)

type OnePagerCompletenessSource interface {
	CompletenessFor(ctx context.Context, subjectIDs []string) (map[string]bool, bool, error)
}

func decorateEnterpriseCapabilitiesOnePagerCompleteness(ctx context.Context, source OnePagerCompletenessSource, rows []readmodels.EnterpriseCapabilityDTO) error {
	if source == nil || len(rows) == 0 {
		return nil
	}
	ids := make([]string, len(rows))
	for i := range rows {
		ids[i] = rows[i].ID
	}
	indicators, present, err := source.CompletenessFor(ctx, ids)
	if err != nil {
		return err
	}
	if !present {
		return nil
	}
	for i := range rows {
		complete := indicators[ids[i]]
		rows[i].OnePagerComplete = &complete
	}
	return nil
}
