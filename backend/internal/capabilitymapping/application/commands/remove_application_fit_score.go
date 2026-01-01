package commands

type RemoveApplicationFitScore struct {
	FitScoreID string
	RemovedBy  string
}

func (c RemoveApplicationFitScore) CommandName() string {
	return "RemoveApplicationFitScore"
}
