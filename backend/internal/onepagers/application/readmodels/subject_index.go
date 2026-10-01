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
	Attributes     SubjectAttributes
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

type CompletenessTransition struct {
	SubjectID string
	Previous  CompletenessCounts
	Current   CompletenessCounts
}

func (t CompletenessTransition) Changed() bool {
	return t.Previous.Signal() != t.Current.Signal() || t.Previous.Missing() != t.Current.Missing()
}

func (c CompletenessCounts) Signal() string {
	switch c.bucket() {
	case bucketNotApplicable:
		return SignalNotApplicable
	case bucketComplete:
		return SignalComplete
	default:
		return SignalIncomplete
	}
}

func (c CompletenessCounts) bucket() int {
	if c.Required == 0 {
		return bucketNotApplicable
	}
	if c.Filled >= c.Required {
		return bucketComplete
	}
	return bucketIncomplete
}

func (c CompletenessCounts) Missing() int {
	return max(c.Required-c.Filled, 0)
}

func (r SubjectIndexRecord) counts() CompletenessCounts {
	return CompletenessCounts{Required: r.RequiredCount, Filled: r.FilledCount}
}

func (r SubjectIndexRecord) Signal() string {
	return r.counts().Signal()
}

func (r SubjectIndexRecord) CompletenessBucket() int {
	return r.counts().bucket()
}

func (r SubjectIndexRecord) MissingCount() int {
	return r.counts().Missing()
}
