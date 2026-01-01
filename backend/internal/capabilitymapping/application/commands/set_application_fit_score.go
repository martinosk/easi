package commands

type SetApplicationFitScore struct {
	ComponentID string
	PillarID    string
	Score       int
	Rationale   string
	ScoredBy    string
}

func (c SetApplicationFitScore) CommandName() string {
	return "SetApplicationFitScore"
}
