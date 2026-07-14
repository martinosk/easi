package queries

import (
	"context"
	"fmt"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
)

type FilledCountsSource interface {
	FilledFieldCounts(ctx context.Context, subjectType string, subjectIDs, fieldIDs []string) (map[string]int, error)
}

type CompletenessIndicators struct {
	configs  ConfigurationSource
	facts    FilledCountsSource
	builtIns map[string]ports.BuiltInFieldSource
}

func NewCompletenessIndicators(configs ConfigurationSource, facts FilledCountsSource, builtIns map[string]ports.BuiltInFieldSource) *CompletenessIndicators {
	return &CompletenessIndicators{configs: configs, facts: facts, builtIns: builtIns}
}

func (s *CompletenessIndicators) ForSubjects(ctx context.Context, subjectType string, subjectIDs []string) (map[string]bool, bool, error) {
	requiredCount, filled, err := s.CountsForSubjects(ctx, subjectType, subjectIDs)
	if err != nil {
		return nil, false, err
	}
	if requiredCount == 0 {
		return nil, false, nil
	}

	indicators := make(map[string]bool, len(subjectIDs))
	for _, subjectID := range subjectIDs {
		indicators[subjectID] = filled[subjectID] == requiredCount
	}
	return indicators, true, nil
}

func (s *CompletenessIndicators) CountsForSubjects(ctx context.Context, subjectType string, subjectIDs []string) (int, map[string]int, error) {
	config, err := s.configs.GetBySubjectType(ctx, subjectType)
	if err != nil {
		return 0, nil, fmt.Errorf("get one-pager configuration for subject type %s: %w", subjectType, err)
	}

	customFieldIDs, builtInEntryIDs := activeRequiredFieldRefs(config)
	requiredCount := len(customFieldIDs) + len(builtInEntryIDs)

	filled := make(map[string]int, len(subjectIDs))
	for _, subjectID := range subjectIDs {
		filled[subjectID] = 0
	}
	if requiredCount == 0 || len(subjectIDs) == 0 {
		return requiredCount, filled, nil
	}

	customCounts, err := s.customFilledCounts(ctx, subjectType, subjectIDs, customFieldIDs)
	if err != nil {
		return 0, nil, err
	}
	builtInFilled, err := s.builtInFilled(ctx, subjectType, subjectIDs, builtInEntryIDs)
	if err != nil {
		return 0, nil, err
	}

	for _, subjectID := range subjectIDs {
		filled[subjectID] = customCounts[subjectID] + filledBuiltInCount(builtInFilled[subjectID])
	}
	return requiredCount, filled, nil
}

func (s *CompletenessIndicators) customFilledCounts(ctx context.Context, subjectType string, subjectIDs, fieldIDs []string) (map[string]int, error) {
	if len(fieldIDs) == 0 {
		return map[string]int{}, nil
	}
	counts, err := s.facts.FilledFieldCounts(ctx, subjectType, subjectIDs, fieldIDs)
	if err != nil {
		return nil, fmt.Errorf("count filled required fields for subject type %s: %w", subjectType, err)
	}
	return counts, nil
}

func (s *CompletenessIndicators) builtInFilled(ctx context.Context, subjectType string, subjectIDs, entryIDs []string) (map[string]map[string]bool, error) {
	if len(entryIDs) == 0 {
		return map[string]map[string]bool{}, nil
	}
	source, found := s.builtIns[subjectType]
	if !found {
		return nil, fmt.Errorf("no built-in field source configured for subject type %s", subjectType)
	}
	filled, err := source.FilledBuiltInFields(ctx, subjectIDs, entryIDs)
	if err != nil {
		return nil, fmt.Errorf("evaluate required built-in fields for subject type %s: %w", subjectType, err)
	}
	return filled, nil
}

func filledBuiltInCount(filled map[string]bool) int {
	count := 0
	for _, isFilled := range filled {
		if isFilled {
			count++
		}
	}
	return count
}

func activeRequiredFieldRefs(config *readmodels.ConfigurationRecord) (customFieldIDs, builtInEntryIDs []string) {
	if config == nil {
		return nil, nil
	}
	return config.Document.ActiveRequiredCustomFieldIDs(), config.Document.RequiredBuiltInEntryIDs()
}
