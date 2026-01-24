package commands

type SetAcquiredVia struct {
	ComponentID string
	EntityID    string
	Notes       string
}

func (c SetAcquiredVia) CommandName() string {
	return "SetAcquiredVia"
}

type ClearAcquiredVia struct {
	ComponentID string
}

func (c ClearAcquiredVia) CommandName() string {
	return "ClearAcquiredVia"
}

type SetPurchasedFrom struct {
	ComponentID string
	VendorID    string
	Notes       string
}

func (c SetPurchasedFrom) CommandName() string {
	return "SetPurchasedFrom"
}

type ClearPurchasedFrom struct {
	ComponentID string
}

func (c ClearPurchasedFrom) CommandName() string {
	return "ClearPurchasedFrom"
}

type SetBuiltBy struct {
	ComponentID string
	TeamID      string
	Notes       string
}

func (c SetBuiltBy) CommandName() string {
	return "SetBuiltBy"
}

type ClearBuiltBy struct {
	ComponentID string
}

func (c ClearBuiltBy) CommandName() string {
	return "ClearBuiltBy"
}
