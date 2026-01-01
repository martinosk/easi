package commands

type UpdateApplicationFitScore struct {
	FitScoreID string
	Score      int
	Rationale  string
	UpdatedBy  string
}

func (c UpdateApplicationFitScore) CommandName() string {
	return "UpdateApplicationFitScore"
}
