package valueobjects

type DeletionScope struct {
	rootID      CapabilityID
	descendants []CapabilityID
	idSet       map[string]bool
}

func NewDeletionScope(rootID CapabilityID, descendants []CapabilityID) DeletionScope {
	all := append([]CapabilityID{rootID}, descendants...)
	idSet := make(map[string]bool, len(all))
	for _, id := range all {
		idSet[id.Value()] = true
	}
	return DeletionScope{rootID: rootID, descendants: descendants, idSet: idSet}
}

func (s DeletionScope) Contains(id string) bool {
	return s.idSet[id]
}

func (s DeletionScope) HasDescendants() bool {
	return len(s.descendants) > 0
}

func (s DeletionScope) AllIDs() []CapabilityID {
	result := make([]CapabilityID, 0, len(s.descendants)+1)
	result = append(result, s.rootID)
	result = append(result, s.descendants...)
	return result
}

func (s DeletionScope) BottomUp() []CapabilityID {
	result := make([]CapabilityID, 0, len(s.descendants)+1)
	for i := len(s.descendants) - 1; i >= 0; i-- {
		result = append(result, s.descendants[i])
	}
	result = append(result, s.rootID)
	return result
}

func (s DeletionScope) RootID() CapabilityID {
	return s.rootID
}
