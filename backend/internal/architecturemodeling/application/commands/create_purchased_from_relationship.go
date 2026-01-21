package commands

type CreatePurchasedFromRelationship struct {
	VendorID        string
	ComponentID     string
	Notes           string
	ReplaceExisting bool
}

func (c CreatePurchasedFromRelationship) CommandName() string {
	return "CreatePurchasedFromRelationship"
}
