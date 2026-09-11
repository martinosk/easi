package queries

import (
	"context"
	"fmt"

	"easi/backend/internal/onepagers/application/ports"
)

type FilledCountsSource interface {
	FilledFieldCounts(ctx context.Context, subjectType string, subjectIDs, fieldIDs []string) (map[string]int, error)
}

type CompletenessIndicators struct {
	configs     ConfigurationSource
	definitions DefinitionSource
	facts       FilledCountsSource
	builtIns    map[string]ports.BuiltInFieldSource
}

func NewCompletenessIndicators(configs ConfigurationSource, definitions DefinitionSource, facts FilledCountsSource, builtIns map[string]ports.BuiltInFieldSource) *CompletenessIndicators {
	return &CompletenessIndicators{configs: configs, definitions: definitions, facts: facts, builtIns: builtIns}
}

type subjectScope struct {
	subjectType string
	subjectIDs  []string
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

type requiredFields struct {
	customFieldIDs  []string
	builtInEntryIDs []string
}

func (r requiredFields) count() int {
	return len(r.customFieldIDs) + len(r.builtInEntryIDs)
}

func (s *CompletenessIndicators) CountsForSubjects(ctx context.Context, subjectType string, subjectIDs []string) (int, map[string]int, error) {
	scope := subjectScope{subjectType: subjectType, subjectIDs: subjectIDs}
	required, err := s.requiredFields(ctx, scope)
	if err != nil {
		return 0, nil, err
	}
	filled := zeroCounts(scope)
	if required.count() == 0 || len(scope.subjectIDs) == 0 {
		return required.count(), filled, nil
	}

	customCounts, err := s.customFilledCounts(ctx, scope, required)
	if err != nil {
		return 0, nil, err
	}
	builtInFilled, err := s.builtInFilled(ctx, scope, required)
	if err != nil {
		return 0, nil, err
	}

	for _, subjectID := range scope.subjectIDs {
		filled[subjectID] = customCounts[subjectID] + filledBuiltInCount(builtInFilled[subjectID])
	}
	return required.count(), filled, nil
}

func zeroCounts(scope subjectScope) map[string]int {
	filled := make(map[string]int, len(scope.subjectIDs))
	for _, subjectID := range scope.subjectIDs {
		filled[subjectID] = 0
	}
	return filled
}

func (s *CompletenessIndicators) customFilledCounts(ctx context.Context, scope subjectScope, required requiredFields) (map[string]int, error) {
	if len(required.customFieldIDs) == 0 {
		return map[string]int{}, nil
	}
	counts, err := s.facts.FilledFieldCounts(ctx, scope.subjectType, scope.subjectIDs, required.customFieldIDs)
	if err != nil {
		return nil, fmt.Errorf("count filled required fields for subject type %s: %w", scope.subjectType, err)
	}
	return counts, nil
}

func (s *CompletenessIndicators) builtInFilled(ctx context.Context, scope subjectScope, required requiredFields) (map[string]map[string]bool, error) {
	if len(required.builtInEntryIDs) == 0 {
		return map[string]map[string]bool{}, nil
	}
	source, found := s.builtIns[scope.subjectType]
	if !found {
		return nil, fmt.Errorf("no built-in field source configured for subject type %s", scope.subjectType)
	}
	filled, err := source.FilledBuiltInFields(ctx, scope.subjectIDs, required.builtInEntryIDs)
	if err != nil {
		return nil, fmt.Errorf("evaluate required built-in fields for subject type %s: %w", scope.subjectType, err)
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

func (s *CompletenessIndicators) requiredFields(ctx context.Context, scope subjectScope) (requiredFields, error) {
	config, err := s.configs.GetBySubjectType(ctx, scope.subjectType)
	if err != nil {
		return requiredFields{}, fmt.Errorf("get one-pager configuration for subject type %s: %w", scope.subjectType, err)
	}
	if config == nil {
		return requiredFields{}, nil
	}
	required := requiredFields{builtInEntryIDs: config.Document.RequiredBuiltInEntryIDs()}
	if len(config.Document.RequiredCustomFieldIDs()) == 0 {
		return required, nil
	}
	definitions, err := s.definitions.ForSubjectType(ctx, scope.subjectType)
	if err != nil {
		return requiredFields{}, fmt.Errorf("get custom field definitions for subject type %s: %w", scope.subjectType, err)
	}
	required.customFieldIDs = activeRequiredCustomFieldIDs(config.Document, definitions)
	return required, nil
}
