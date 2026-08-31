package services

type CompositionRole string

const (
	RoleSource    CompositionRole = "source"
	RoleImplicit  CompositionRole = "implicit"
	RoleCarvedOut CompositionRole = "carved-out"
)

type CapabilityNode struct {
	ID                 string
	ParentID           string
	Name               string
	Level              string
	BusinessDomainID   string
	BusinessDomainName string
}

type ActiveDirectionSources struct {
	EnterpriseCapabilityID   string
	EnterpriseCapabilityName string
	SourceCapabilityIDs      []string
}

type CarvedOutBy struct {
	EnterpriseCapabilityID   string
	EnterpriseCapabilityName string
}

type ResolvedCapability struct {
	Node        CapabilityNode
	Role        CompositionRole
	CarvedOutBy *CarvedOutBy
}

type CompositionCounts struct {
	SourceCount    int
	IncludedCount  int
	CarvedOutCount int
	DomainCount    int
}

type traversal struct {
	capByID       map[string]CapabilityNode
	children      map[string][]string
	targetSources map[string]struct{}
	ownedByOther  map[string]CarvedOutBy
	visited       map[string]struct{}
	resolved      []ResolvedCapability
}

type CapabilityIndex struct {
	capByID  map[string]CapabilityNode
	children map[string][]string
}

func BuildCapabilityIndex(capabilities []CapabilityNode) CapabilityIndex {
	return CapabilityIndex{capByID: indexByID(capabilities), children: indexChildren(capabilities)}
}

func ResolveComposition(targetECID string, directions []ActiveDirectionSources, capabilities []CapabilityNode) []ResolvedCapability {
	return ResolveCompositionWithIndex(targetECID, directions, BuildCapabilityIndex(capabilities))
}

func ResolveCompositionWithIndex(targetECID string, directions []ActiveDirectionSources, index CapabilityIndex) []ResolvedCapability {
	target, ok := findDirection(directions, targetECID)
	if !ok {
		return nil
	}

	t := &traversal{
		capByID:       index.capByID,
		children:      index.children,
		targetSources: toSet(target.SourceCapabilityIDs),
		ownedByOther:  ownershipByOtherEC(directions, targetECID),
		visited:       make(map[string]struct{}),
	}
	for _, sourceID := range target.SourceCapabilityIDs {
		t.visit(sourceID)
	}
	return t.resolved
}

func EvaluateSourceEligibility(capabilityID, targetECID string, directions []ActiveDirectionSources) *CarvedOutBy {
	for _, d := range directions {
		if d.EnterpriseCapabilityID != targetECID && d.sources(capabilityID) {
			return &CarvedOutBy{
				EnterpriseCapabilityID:   d.EnterpriseCapabilityID,
				EnterpriseCapabilityName: d.EnterpriseCapabilityName,
			}
		}
	}
	return nil
}

func (d ActiveDirectionSources) sources(capabilityID string) bool {
	for _, sourceID := range d.SourceCapabilityIDs {
		if sourceID == capabilityID {
			return true
		}
	}
	return false
}

func CountComposition(sourceCount int, resolved []ResolvedCapability) CompositionCounts {
	counts := CompositionCounts{SourceCount: sourceCount}
	domains := make(map[string]struct{})
	for _, r := range resolved {
		if r.Role == RoleCarvedOut {
			counts.CarvedOutCount++
			continue
		}
		counts.IncludedCount++
		if r.Node.BusinessDomainID != "" {
			domains[r.Node.BusinessDomainID] = struct{}{}
		}
	}
	counts.DomainCount = len(domains)
	return counts
}

func (t *traversal) visit(capabilityID string) {
	if _, done := t.visited[capabilityID]; done {
		return
	}
	capability, exists := t.capByID[capabilityID]
	if !exists {
		return
	}
	t.visited[capabilityID] = struct{}{}

	_, isTargetSource := t.targetSources[capabilityID]
	if owner, claimed := t.ownedByOther[capabilityID]; claimed && !isTargetSource {
		t.resolved = append(t.resolved, ResolvedCapability{Node: capability, Role: RoleCarvedOut, CarvedOutBy: &owner})
		return
	}

	role := RoleImplicit
	if isTargetSource {
		role = RoleSource
	}
	t.resolved = append(t.resolved, ResolvedCapability{Node: capability, Role: role})
	for _, childID := range t.children[capabilityID] {
		t.visit(childID)
	}
}

func findDirection(directions []ActiveDirectionSources, ecID string) (ActiveDirectionSources, bool) {
	for _, d := range directions {
		if d.EnterpriseCapabilityID == ecID {
			return d, true
		}
	}
	return ActiveDirectionSources{}, false
}

func ownershipByOtherEC(directions []ActiveDirectionSources, targetECID string) map[string]CarvedOutBy {
	owners := make(map[string]CarvedOutBy)
	for _, d := range directions {
		if d.EnterpriseCapabilityID == targetECID {
			continue
		}
		for _, sourceID := range d.SourceCapabilityIDs {
			owners[sourceID] = CarvedOutBy{
				EnterpriseCapabilityID:   d.EnterpriseCapabilityID,
				EnterpriseCapabilityName: d.EnterpriseCapabilityName,
			}
		}
	}
	return owners
}

func indexByID(capabilities []CapabilityNode) map[string]CapabilityNode {
	index := make(map[string]CapabilityNode, len(capabilities))
	for _, c := range capabilities {
		index[c.ID] = c
	}
	return index
}

func indexChildren(capabilities []CapabilityNode) map[string][]string {
	index := make(map[string][]string)
	for _, c := range capabilities {
		if c.ParentID == "" {
			continue
		}
		index[c.ParentID] = append(index[c.ParentID], c.ID)
	}
	return index
}

func toSet(ids []string) map[string]struct{} {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}
