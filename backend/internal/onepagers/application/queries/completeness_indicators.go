package queries

import (
	"context"
	"fmt"

	"easi/backend/internal/onepagers/application/readmodels"
)

type FilledCountsSource interface {
	FilledFieldCounts(ctx context.Context, subjectType string, subjectIDs, fieldIDs []string) (map[string]int, error)
}

type CompletenessIndicators struct {
	configs ConfigurationSource
	facts   FilledCountsSource
}

func NewCompletenessIndicators(configs ConfigurationSource, facts FilledCountsSource) *CompletenessIndicators {
	return &CompletenessIndicators{configs: configs, facts: facts}
}

func (s *CompletenessIndicators) ForSubjects(ctx context.Context, subjectType string, subjectIDs []string) (map[string]bool, bool, error) {
	config, err := s.configs.GetBySubjectType(ctx, subjectType)
	if err != nil {
		return nil, false, fmt.Errorf("get one-pager configuration for subject type %s: %w", subjectType, err)
	}

	fieldIDs := activeRequiredFieldIDs(config)
	if len(fieldIDs) == 0 {
		return nil, false, nil
	}
	if len(subjectIDs) == 0 {
		return map[string]bool{}, true, nil
	}

	counts, err := s.facts.FilledFieldCounts(ctx, subjectType, subjectIDs, fieldIDs)
	if err != nil {
		return nil, false, fmt.Errorf("count filled required fields for subject type %s: %w", subjectType, err)
	}

	indicators := make(map[string]bool, len(subjectIDs))
	for _, subjectID := range subjectIDs {
		indicators[subjectID] = counts[subjectID] == len(fieldIDs)
	}
	return indicators, true, nil
}

func activeRequiredFieldIDs(config *readmodels.ConfigurationRecord) []string {
	if config == nil {
		return nil
	}
	var fieldIDs []string
	for _, field := range config.Document.CustomFields {
		if field.Active && field.Required {
			fieldIDs = append(fieldIDs, field.ID)
		}
	}
	return fieldIDs
}
