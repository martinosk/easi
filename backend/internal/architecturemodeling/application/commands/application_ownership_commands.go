package commands

type NominateApplicationComponentOwner struct {
	ComponentID string
	OwnerKind   string
	OwnerID     string
}

func (c NominateApplicationComponentOwner) CommandName() string {
	return "NominateApplicationComponentOwner"
}

type ConfirmApplicationComponentOwnership struct {
	ComponentID string
}

func (c ConfirmApplicationComponentOwnership) CommandName() string {
	return "ConfirmApplicationComponentOwnership"
}

type AssignApplicationComponentOwner struct {
	ComponentID string
	OwnerKind   string
	OwnerID     string
}

func (c AssignApplicationComponentOwner) CommandName() string {
	return "AssignApplicationComponentOwner"
}

type ClearApplicationComponentOwnership struct {
	ComponentID string
}

func (c ClearApplicationComponentOwnership) CommandName() string {
	return "ClearApplicationComponentOwnership"
}
