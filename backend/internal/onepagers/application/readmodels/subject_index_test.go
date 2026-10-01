package readmodels_test

import (
	"testing"

	"easi/backend/internal/onepagers/application/readmodels"

	"github.com/stretchr/testify/assert"
)

func TestSubjectIndexRecordSignal(t *testing.T) {
	cases := []struct {
		name     string
		required int
		filled   int
		signal   string
		bucket   int
		missing  int
	}{
		{"no required fields is not applicable", 0, 0, readmodels.SignalNotApplicable, 2, 0},
		{"all required fields filled is complete", 3, 3, readmodels.SignalComplete, 1, 0},
		{"over-filled is complete", 2, 3, readmodels.SignalComplete, 1, 0},
		{"some required fields missing is incomplete", 3, 1, readmodels.SignalIncomplete, 0, 2},
		{"none filled is incomplete", 2, 0, readmodels.SignalIncomplete, 0, 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			record := readmodels.SubjectIndexRecord{RequiredCount: tc.required, FilledCount: tc.filled}
			assert.Equal(t, tc.signal, record.Signal())
			assert.Equal(t, tc.bucket, record.CompletenessBucket())
			assert.Equal(t, tc.missing, record.MissingCount())
		})
	}
}

func TestCompletenessTransitionChanged(t *testing.T) {
	cases := []struct {
		name     string
		previous readmodels.CompletenessCounts
		current  readmodels.CompletenessCounts
		changed  bool
	}{
		{"filling the last missing field", readmodels.CompletenessCounts{Required: 3, Filled: 2}, readmodels.CompletenessCounts{Required: 3, Filled: 3}, true},
		{"clearing a required field", readmodels.CompletenessCounts{Required: 4, Filled: 4}, readmodels.CompletenessCounts{Required: 4, Filled: 3}, true},
		{"missing count alone changes", readmodels.CompletenessCounts{Required: 5, Filled: 2}, readmodels.CompletenessCounts{Required: 5, Filled: 3}, true},
		{"nothing changes", readmodels.CompletenessCounts{Required: 3, Filled: 1}, readmodels.CompletenessCounts{Required: 3, Filled: 1}, false},
		{"required and filled grow together", readmodels.CompletenessCounts{Required: 2, Filled: 2}, readmodels.CompletenessCounts{Required: 3, Filled: 3}, false},
		{"over-filled stays complete", readmodels.CompletenessCounts{Required: 2, Filled: 3}, readmodels.CompletenessCounts{Required: 2, Filled: 2}, false},
		{"requirements removed", readmodels.CompletenessCounts{Required: 1, Filled: 1}, readmodels.CompletenessCounts{Required: 0, Filled: 0}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			transition := readmodels.CompletenessTransition{SubjectID: "s-1", Previous: tc.previous, Current: tc.current}
			assert.Equal(t, tc.changed, transition.Changed())
		})
	}
}

func TestCompletenessCountsSignalAndMissing(t *testing.T) {
	counts := readmodels.CompletenessCounts{Required: 3, Filled: 1}
	assert.Equal(t, readmodels.SignalIncomplete, counts.Signal())
	assert.Equal(t, 2, counts.Missing())
	assert.Equal(t, readmodels.SignalNotApplicable, readmodels.CompletenessCounts{}.Signal())
}
