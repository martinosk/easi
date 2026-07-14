package readmodels

import "time"

const (
	SignalComplete      = "complete"
	SignalIncomplete    = "incomplete"
	SignalNotApplicable = "not-applicable"
)

const (
	bucketIncomplete    = 0
	bucketComplete      = 1
	bucketNotApplicable = 2
)

type SubjectIndexRecord struct {
	SubjectType    string
	SubjectID      string
	Name           string
	CreatorActorID string
	CreatorEmail   string
	CreatedAt      time.Time
	LastUpdatedAt  time.Time
	RequiredCount  int
	FilledCount    int
}

type CompletenessCounts struct {
	Required int
	Filled   int
}

type SubjectChange struct {
	Subject    SubjectKey
	Name       string
	Counts     CompletenessCounts
	OccurredAt time.Time
}

func (r SubjectIndexRecord) Signal() string {
	switch r.CompletenessBucket() {
	case bucketNotApplicable:
		return SignalNotApplicable
	case bucketComplete:
		return SignalComplete
	default:
		return SignalIncomplete
	}
}

func (r SubjectIndexRecord) CompletenessBucket() int {
	if r.RequiredCount == 0 {
		return bucketNotApplicable
	}
	if r.FilledCount >= r.RequiredCount {
		return bucketComplete
	}
	return bucketIncomplete
}

func (r SubjectIndexRecord) MissingCount() int {
	missing := r.RequiredCount - r.FilledCount
	if missing < 0 {
		return 0
	}
	return missing
}
