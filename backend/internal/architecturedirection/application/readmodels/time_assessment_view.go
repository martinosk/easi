package readmodels

import "context"

type TimeAssessmentSource interface {
	GetByPair(ctx context.Context, capabilityID, componentID string) (*TimeAssessmentDTO, error)
	GetByCapabilityIDs(ctx context.Context, capabilityIDs []string) ([]TimeAssessmentDTO, error)
	GetAll(ctx context.Context) ([]TimeAssessmentDTO, error)
	GetRollupsByComponentIDs(ctx context.Context, componentIDs []string) ([]TimeAssessmentRollupDTO, error)
}

type TimeSuggestionSource interface {
	All(ctx context.Context) ([]SuggestedRealization, error)
	ForCapabilities(ctx context.Context, capabilityIDs []string) ([]SuggestedRealization, error)
	ForPair(ctx context.Context, capabilityID, componentID string) (*TimeSuggestionDTO, error)
}

type TimeAssessmentView struct {
	assessments TimeAssessmentSource
	suggestions TimeSuggestionSource
}

func NewTimeAssessmentView(assessments TimeAssessmentSource, suggestions TimeSuggestionSource) *TimeAssessmentView {
	return &TimeAssessmentView{assessments: assessments, suggestions: suggestions}
}

func (v *TimeAssessmentView) GetByPair(ctx context.Context, capabilityID, componentID string) (*TimeAssessmentDTO, error) {
	assessment, err := v.assessments.GetByPair(ctx, capabilityID, componentID)
	if err != nil || assessment == nil {
		return nil, err
	}
	suggestion, err := v.suggestions.ForPair(ctx, capabilityID, componentID)
	if err != nil {
		return nil, err
	}
	assessment.Suggestion = suggestion
	return assessment, nil
}

func (v *TimeAssessmentView) GetByCapabilityIDs(ctx context.Context, capabilityIDs []string) ([]TimeAssessmentDTO, error) {
	return composeTimeRealizations(timeRealizationSources{
		recorded:  func() ([]TimeAssessmentDTO, error) { return v.assessments.GetByCapabilityIDs(ctx, capabilityIDs) },
		suggested: func() ([]SuggestedRealization, error) { return v.suggestions.ForCapabilities(ctx, capabilityIDs) },
	})
}

func (v *TimeAssessmentView) GetAll(ctx context.Context) ([]TimeAssessmentDTO, error) {
	return composeTimeRealizations(timeRealizationSources{
		recorded:  func() ([]TimeAssessmentDTO, error) { return v.assessments.GetAll(ctx) },
		suggested: func() ([]SuggestedRealization, error) { return v.suggestions.All(ctx) },
	})
}

func (v *TimeAssessmentView) GetRollupsByComponentIDs(ctx context.Context, componentIDs []string) ([]TimeAssessmentRollupDTO, error) {
	return v.assessments.GetRollupsByComponentIDs(ctx, componentIDs)
}

type timeRealizationSources struct {
	recorded  func() ([]TimeAssessmentDTO, error)
	suggested func() ([]SuggestedRealization, error)
}

func composeTimeRealizations(sources timeRealizationSources) ([]TimeAssessmentDTO, error) {
	recorded, err := sources.recorded()
	if err != nil {
		return nil, err
	}
	suggested, err := sources.suggested()
	if err != nil {
		return nil, err
	}
	return mergeSuggestionsIntoAssessments(recorded, suggested), nil
}

func mergeSuggestionsIntoAssessments(assessments []TimeAssessmentDTO, suggested []SuggestedRealization) []TimeAssessmentDTO {
	merged := append([]TimeAssessmentDTO{}, assessments...)
	indexByPair := make(map[RealizationPair]int, len(merged))
	for i := range merged {
		indexByPair[RealizationPair{CapabilityID: merged[i].CapabilityID, ComponentID: merged[i].ComponentID}] = i
	}
	for _, realization := range suggested {
		suggestion := realization.Suggestion
		if index, assessed := indexByPair[realization.Pair]; assessed {
			merged[index].Suggestion = &suggestion
			continue
		}
		if suggestion.Grade == nil {
			continue
		}
		merged = append(merged, unassessedRealization(realization))
	}
	return merged
}

func unassessedRealization(realization SuggestedRealization) TimeAssessmentDTO {
	suggestion := realization.Suggestion
	return TimeAssessmentDTO{
		CapabilityID:   realization.Pair.CapabilityID,
		CapabilityName: realization.CapabilityName,
		ComponentID:    realization.Pair.ComponentID,
		ComponentName:  realization.ComponentName,
		Suggestion:     &suggestion,
	}
}
