package commands

type SetTargetMaturity struct {
	ID             string
	TargetMaturity int
}

func (c SetTargetMaturity) CommandName() string {
	return "SetTargetMaturity"
}
