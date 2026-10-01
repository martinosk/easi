package events

import (
	opPL "easi/backend/internal/onepagers/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type SubjectCompleteness struct {
	SubjectType   string
	SubjectID     string
	Completeness  string
	RequiredCount int
	MissingCount  int
}

type SubjectCompletenessRecalculated struct {
	domain.BaseEvent
	SubjectCompleteness
}

func NewSubjectCompletenessRecalculated(completeness SubjectCompleteness) SubjectCompletenessRecalculated {
	return SubjectCompletenessRecalculated{
		BaseEvent:           domain.NewBaseEvent(completeness.SubjectID),
		SubjectCompleteness: completeness,
	}
}

func (e SubjectCompletenessRecalculated) EventType() string {
	return opPL.SubjectCompletenessRecalculated
}

func (e SubjectCompletenessRecalculated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"subjectType":    e.SubjectType,
		"subjectId":      e.SubjectID,
		"completeness":   e.Completeness,
		"requiredCount":  e.RequiredCount,
		"missingCount":   e.MissingCount,
		"recalculatedAt": e.OccurredAt(),
	}
}
