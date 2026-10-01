package events_test

import (
	"testing"

	opevents "easi/backend/internal/onepagers/domain/events"
	opPL "easi/backend/internal/onepagers/publishedlanguage"

	"github.com/stretchr/testify/assert"
)

func TestSubjectCompletenessRecalculated_Payload(t *testing.T) {
	event := opevents.NewSubjectCompletenessRecalculated(opevents.SubjectCompleteness{
		SubjectType:   "application",
		SubjectID:     "app-1",
		Completeness:  "incomplete",
		RequiredCount: 3,
		MissingCount:  2,
	})

	assert.Equal(t, opPL.SubjectCompletenessRecalculated, event.EventType())
	assert.Equal(t, "app-1", event.AggregateID())
	data := event.EventData()
	assert.ElementsMatch(t, []string{"subjectType", "subjectId", "completeness", "requiredCount", "missingCount", "recalculatedAt"}, keys(data))
	assert.Equal(t, "application", data["subjectType"])
	assert.Equal(t, "app-1", data["subjectId"])
	assert.Equal(t, "incomplete", data["completeness"])
	assert.Equal(t, 3, data["requiredCount"])
	assert.Equal(t, 2, data["missingCount"])
	assert.Equal(t, event.OccurredAt(), data["recalculatedAt"])
}

func keys(data map[string]interface{}) []string {
	result := make([]string, 0, len(data))
	for key := range data {
		result = append(result, key)
	}
	return result
}
