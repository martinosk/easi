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
