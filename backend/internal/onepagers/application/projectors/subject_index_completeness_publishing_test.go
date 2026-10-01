package projectors_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/onepagers/application/projectors"
	"easi/backend/internal/onepagers/application/readmodels"
	opevents "easi/backend/internal/onepagers/domain/events"
	opPL "easi/backend/internal/onepagers/publishedlanguage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type publishedCompleteness struct {
	subjectType  string
	subjectID    string
	completeness string
	required     int
	missing      int
}

func (h *projectorHarness) publishedCompleteness() []publishedCompleteness {
	h.t.Helper()
	result := make([]publishedCompleteness, 0, len(h.publisher.published))
	for _, event := range h.publisher.published {
		require.Equal(h.t, opPL.SubjectCompletenessRecalculated, event.EventType())
		data := event.EventData()
		result = append(result, publishedCompleteness{
			subjectType:  data["subjectType"].(string),
			subjectID:    data["subjectId"].(string),
			completeness: data["completeness"].(string),
			required:     data["requiredCount"].(int),
			missing:      data["missingCount"].(int),
		})
	}
	return result
}

func storeWith(subjectType string, counts map[string]readmodels.CompletenessCounts) *fakeIndexStore {
	store := &fakeIndexStore{counts: map[readmodels.SubjectKey]readmodels.CompletenessCounts{}, idsByType: map[string][]string{}}
	for subjectID, current := range counts {
		store.counts[subjectKey(subjectType, subjectID)] = current
		store.idsByType[subjectType] = append(store.idsByType[subjectType], subjectID)
	}
	return store
}

func factsEvent(subjectType, subjectID string) map[string]any {
	return map[string]any{"subjectType": subjectType, "subjectId": subjectID, "fieldId": "f1"}
}

func TestCompletenessPublished_FactsEvents(t *testing.T) {
	cases := []struct {
		name      string
		eventType string
		previous  readmodels.CompletenessCounts
		filled    int
		want      []publishedCompleteness
	}{
		{
			name: "filling the last missing field", eventType: opevents.TypeFieldValueRecorded,
			previous: readmodels.CompletenessCounts{Required: 3, Filled: 2}, filled: 3,
			want: []publishedCompleteness{{"application", "app-1", "complete", 3, 0}},
		},
		{
			name: "clearing a required field", eventType: opevents.TypeFieldValueCleared,
			previous: readmodels.CompletenessCounts{Required: 3, Filled: 3}, filled: 2,
			want: []publishedCompleteness{{"application", "app-1", "incomplete", 3, 1}},
		},
		{
			name: "a change in missing count alone", eventType: opevents.TypeFieldValueRecorded,
			previous: readmodels.CompletenessCounts{Required: 5, Filled: 2}, filled: 3,
			want: []publishedCompleteness{{"application", "app-1", "incomplete", 5, 2}},
		},
		{
			name: "archiving the facts", eventType: opevents.TypeOnePagerFactsArchived,
			previous: readmodels.CompletenessCounts{Required: 3, Filled: 3}, filled: 0,
			want: []publishedCompleteness{{"application", "app-1", "incomplete", 3, 3}},
		},
		{
			name: "an optional field changes nothing", eventType: opevents.TypeFieldValueRecorded,
			previous: readmodels.CompletenessCounts{Required: 4, Filled: 2}, filled: 2,
			want: []publishedCompleteness{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := storeWith("application", map[string]readmodels.CompletenessCounts{"app-1": tc.previous})
			counter := fakeCounter{required: tc.previous.Required, filled: map[string]int{"app-1": tc.filled}}
			h := newHarness(t, projectorFakes{store: store, counter: counter})

			h.project(tc.eventType, time.Now(), factsEvent("application", "app-1"))

			assert.Equal(t, tc.want, h.publishedCompleteness())
		})
	}
}

func TestCompletenessPublished_NewSubjectPublishesItsFirstCompleteness(t *testing.T) {
	cases := []struct {
		name      string
		eventType string
		required  int
		want      publishedCompleteness
	}{
		{"application with requirements", amPL.ApplicationComponentCreated, 3, publishedCompleteness{"application", "new-1", "incomplete", 3, 3}},
		{"vendor with nothing required", amPL.VendorCreated, 0, publishedCompleteness{"vendor", "new-1", "not-applicable", 0, 0}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t, projectorFakes{store: &fakeIndexStore{}, counter: fakeCounter{required: tc.required}})

			h.project(tc.eventType, time.Now(), map[string]any{"id": "new-1", "name": "Portal"})

			assert.Equal(t, []publishedCompleteness{tc.want}, h.publishedCompleteness())
		})
	}
}

func TestCompletenessPublished_SubjectUpdate(t *testing.T) {
	cases := []struct {
		name   string
		filled int
		want   []publishedCompleteness
	}{
		{"completeness changed", 3, []publishedCompleteness{{"capability", "cap-1", "complete", 3, 0}}},
		{"completeness unchanged", 1, []publishedCompleteness{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := storeWith("capability", map[string]readmodels.CompletenessCounts{"cap-1": {Required: 3, Filled: 1}})
			counter := fakeCounter{required: 3, filled: map[string]int{"cap-1": tc.filled}}
			h := newHarness(t, projectorFakes{store: store, counter: counter})

			h.project(capPL.CapabilityUpdated, time.Now(), map[string]any{"id": "cap-1", "name": "Billing"})

			assert.Equal(t, tc.want, h.publishedCompleteness())
		})
	}
}

func TestCompletenessPublished_RequiringAFieldPublishesOncePerAffectedSubjectInOneCall(t *testing.T) {
	previous := map[string]readmodels.CompletenessCounts{}
	filled := map[string]int{}
	for i := range 10 {
		subjectID := fmt.Sprintf("app-%d", i)
		previous[subjectID] = readmodels.CompletenessCounts{Required: 0, Filled: 0}
		if i < 4 {
			previous[subjectID] = readmodels.CompletenessCounts{Required: 1, Filled: 1}
			filled[subjectID] = 1
		}
	}
	store := storeWith("application", previous)
	lookup := fakeConfigLookup{bySubjectType: map[string]string{"cfg-app": "application"}}
	counter := fakeCounter{required: 1, filled: filled}
	h := newHarness(t, projectorFakes{store: store, counter: counter, lookup: lookup})

	h.project(opevents.TypeBuiltInFieldRequirementChanged, time.Now(), map[string]any{"id": "cfg-app"})

	published := h.publishedCompleteness()
	assert.Len(t, published, 6)
	for _, event := range published {
		assert.Equal(t, publishedCompleteness{"application", event.subjectID, "incomplete", 1, 1}, event)
		assert.NotContains(t, []string{"app-0", "app-1", "app-2", "app-3"}, event.subjectID)
	}
	assert.Equal(t, 1, h.publisher.calls, "a bulk recompute publishes in one call")
}

func TestCompletenessPublished_ConfigurationChangeThatChangesNothingPublishesNothing(t *testing.T) {
	store := storeWith("application", map[string]readmodels.CompletenessCounts{"app-1": {Required: 2, Filled: 1}})
	lookup := fakeConfigLookup{bySubjectType: map[string]string{"cfg-app": "application"}}
	h := newHarness(t, projectorFakes{store: store, counter: fakeCounter{required: 2, filled: map[string]int{"app-1": 1}}, lookup: lookup})

	h.project(opevents.TypeOnePagerFieldsReordered, time.Now(), map[string]any{"id": "cfg-app"})

	assert.Zero(t, h.publisher.calls)
}

func TestCompletenessPublished_RelationFillingARequiredBuiltIn(t *testing.T) {
	store := storeWith("capability", map[string]readmodels.CompletenessCounts{"cap-1": {Required: 1, Filled: 0}})
	publisher := &fakePublisher{}
	index := newSubjectIndexProjector(projectorFakes{store: store, counter: fakeCounter{required: 1, filled: map[string]int{"cap-1": 1}}, publisher: publisher})
	relations := projectors.NewSubjectRelationProjector(&fakeRelationStore{}, &fakeBusinessDomainNames{}, index)
	h := &projectorHarness{t: t, store: store, publisher: publisher}

	require.NoError(t, relations.ProjectEvent(context.Background(), capPL.CapabilityAssignedToDomain,
		[]byte(`{"id":"a-1","businessDomainId":"bd-1","capabilityId":"cap-1"}`)))

	assert.Equal(t, []publishedCompleteness{{"capability", "cap-1", "complete", 1, 0}}, h.publishedCompleteness())
}

func TestCompletenessPublished_DeletionAndAttributeOnlyEventsPublishNothing(t *testing.T) {
	cases := []struct {
		eventType string
		payload   map[string]any
	}{
		{amPL.ApplicationComponentDeleted, map[string]any{"id": "app-1"}},
		{capPL.CapabilityLevelChanged, map[string]any{"capabilityId": "cap-1", "newLevel": "L2"}},
	}

	for _, tc := range cases {
		t.Run(tc.eventType, func(t *testing.T) {
			store := storeWith("application", map[string]readmodels.CompletenessCounts{"app-1": {Required: 2, Filled: 1}})
			h := newHarness(t, projectorFakes{store: store})

			h.project(tc.eventType, time.Now(), tc.payload)

			assert.Zero(t, h.publisher.calls)
		})
	}
}
