package commands

type CreatePurchasedFromRelationship struct {
	VendorID    string
	ComponentID string
	Notes       string
}

func (c CreatePurchasedFromRelationship) CommandName() string {
	return "CreatePurchasedFromRelationship"
}
