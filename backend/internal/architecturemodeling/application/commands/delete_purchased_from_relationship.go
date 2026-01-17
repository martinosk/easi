package commands

type DeletePurchasedFromRelationship struct {
	ID string
}

func (c DeletePurchasedFromRelationship) CommandName() string {
	return "DeletePurchasedFromRelationship"
}
