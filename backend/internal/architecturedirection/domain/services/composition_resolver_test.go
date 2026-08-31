package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func cap(id, parentID string) CapabilityNode {
	return CapabilityNode{ID: id, ParentID: parentID, Name: "Cap " + id, Level: "L2"}
}

func capInDomain(id, parentID, domainID, domainName string) CapabilityNode {
	node := cap(id, parentID)
	node.BusinessDomainID = domainID
	node.BusinessDomainName = domainName
	return node
}

func direction(ecID, ecName string, sourceIDs ...string) ActiveDirectionSources {
	return ActiveDirectionSources{EnterpriseCapabilityID: ecID, EnterpriseCapabilityName: ecName, SourceCapabilityIDs: sourceIDs}
}

func roleByID(resolved []ResolvedCapability) map[string]CompositionRole {
	out := make(map[string]CompositionRole, len(resolved))
	for _, r := range resolved {
		out[r.Node.ID] = r.Role
	}
	return out
}

func TestResolveComposition_NoActiveDirectionYieldsEmptyComposition(t *testing.T) {
	resolved := ResolveComposition("ec-1", nil, []CapabilityNode{cap("cap-1", "")})

	assert.Empty(t, resolved)
}

func TestResolveComposition_SingleSourceWithoutDescendants(t *testing.T) {
	resolved := ResolveComposition("ec-1",
		[]ActiveDirectionSources{direction("ec-1", "Customer Identity", "cap-1")},
		[]CapabilityNode{cap("cap-1", "")},
	)

	assert.Len(t, resolved, 1)
	assert.Equal(t, "cap-1", resolved[0].Node.ID)
	assert.Equal(t, RoleSource, resolved[0].Role)
	assert.Nil(t, resolved[0].CarvedOutBy)
}

func TestResolveComposition_SourceSubtreeIsIncludedAsImplicit(t *testing.T) {
	resolved := ResolveComposition("ec-1",
		[]ActiveDirectionSources{direction("ec-1", "CRM", "cap-root")},
		[]CapabilityNode{cap("cap-root", ""), cap("cap-child", "cap-root"), cap("cap-grandchild", "cap-child")},
	)

	roles := roleByID(resolved)
	assert.Equal(t, RoleSource, roles["cap-root"])
	assert.Equal(t, RoleImplicit, roles["cap-child"])
	assert.Equal(t, RoleImplicit, roles["cap-grandchild"])
}

func TestResolveComposition_SourcesAtAnyLevelAreAllIncluded(t *testing.T) {
	resolved := ResolveComposition("ec-1",
		[]ActiveDirectionSources{direction("ec-1", "Order Management", "l1", "l2", "l3", "l4")},
		[]CapabilityNode{cap("l1", ""), cap("l2", ""), cap("l3", ""), cap("l4", "")},
	)

	assert.Len(t, resolved, 4)
	for _, r := range resolved {
		assert.Equal(t, RoleSource, r.Role)
	}
}

func TestResolveComposition_DescendantSourcedElsewhereIsCarvedOut(t *testing.T) {
	resolved := ResolveComposition("ec-crm",
		[]ActiveDirectionSources{
			direction("ec-crm", "CRM", "cap-identity"),
			direction("ec-pay", "Take Payment", "cap-fraud"),
		},
		[]CapabilityNode{cap("cap-identity", ""), cap("cap-consent", "cap-identity"), cap("cap-fraud", "cap-identity")},
	)

	roles := roleByID(resolved)
	assert.Equal(t, RoleSource, roles["cap-identity"])
	assert.Equal(t, RoleImplicit, roles["cap-consent"])
	assert.Equal(t, RoleCarvedOut, roles["cap-fraud"])

	for _, r := range resolved {
		if r.Node.ID == "cap-fraud" {
			assert.Equal(t, "ec-pay", r.CarvedOutBy.EnterpriseCapabilityID)
			assert.Equal(t, "Take Payment", r.CarvedOutBy.EnterpriseCapabilityName)
		}
	}
}

func TestResolveComposition_CarveOutCarriesItsEntireSubtree(t *testing.T) {
	resolved := ResolveComposition("ec-crm",
		[]ActiveDirectionSources{
			direction("ec-crm", "CRM", "cap-identity"),
			direction("ec-pay", "Take Payment", "cap-fraud"),
		},
		[]CapabilityNode{
			cap("cap-identity", ""),
			cap("cap-fraud", "cap-identity"),
			cap("cap-chargeback", "cap-fraud"),
		},
	)

	roles := roleByID(resolved)
	assert.Equal(t, RoleCarvedOut, roles["cap-fraud"])
	assert.NotContains(t, roles, "cap-chargeback", "carved-out subtrees are not descended into")
}

func TestResolveComposition_TargetSourceClaimedElsewhereIsKeptByTarget(t *testing.T) {
	resolved := ResolveComposition("ec-disputes",
		[]ActiveDirectionSources{
			direction("ec-disputes", "Disputes", "cap-chargeback"),
			direction("ec-pay", "Take Payment", "cap-fraud"),
		},
		[]CapabilityNode{cap("cap-fraud", ""), cap("cap-chargeback", "cap-fraud"), cap("cap-evidence", "cap-chargeback")},
	)

	roles := roleByID(resolved)
	assert.Equal(t, RoleSource, roles["cap-chargeback"], "most-specific source wins over another EC's claim")
	assert.Equal(t, RoleImplicit, roles["cap-evidence"])
}

func TestResolveComposition_StaleSourceIsSkipped(t *testing.T) {
	resolved := ResolveComposition("ec-1",
		[]ActiveDirectionSources{direction("ec-1", "CRM", "cap-gone", "cap-1")},
		[]CapabilityNode{cap("cap-1", "")},
	)

	assert.Len(t, resolved, 1)
	assert.Equal(t, "cap-1", resolved[0].Node.ID)
}

func TestResolveComposition_OverlappingSourcesAreResolvedOnce(t *testing.T) {
	resolved := ResolveComposition("ec-1",
		[]ActiveDirectionSources{direction("ec-1", "CRM", "cap-root", "cap-child")},
		[]CapabilityNode{cap("cap-root", ""), cap("cap-child", "cap-root")},
	)

	assert.Len(t, resolved, 2)
	roles := roleByID(resolved)
	assert.Equal(t, RoleSource, roles["cap-root"])
	assert.Equal(t, RoleSource, roles["cap-child"], "explicit source keeps role source even when reached via subtree")
}

func TestResolveComposition_PreservesDepthFirstInsertionOrder(t *testing.T) {
	resolved := ResolveComposition("ec-1",
		[]ActiveDirectionSources{direction("ec-1", "CRM", "cap-b", "cap-a")},
		[]CapabilityNode{cap("cap-a", ""), cap("cap-b", ""), cap("cap-b1", "cap-b")},
	)

	ids := make([]string, len(resolved))
	for i, r := range resolved {
		ids[i] = r.Node.ID
	}
	assert.Equal(t, []string{"cap-b", "cap-b1", "cap-a"}, ids)
}

func TestCompositionCounts(t *testing.T) {
	resolved := ResolveComposition("ec-crm",
		[]ActiveDirectionSources{
			direction("ec-crm", "CRM", "cap-identity"),
			direction("ec-pay", "Take Payment", "cap-fraud"),
		},
		[]CapabilityNode{
			capInDomain("cap-identity", "", "dom-1", "Customer"),
			capInDomain("cap-consent", "cap-identity", "dom-1", "Customer"),
			capInDomain("cap-fraud", "cap-identity", "dom-2", "Payments"),
		},
	)

	counts := CountComposition(1, resolved)
	assert.Equal(t, 1, counts.SourceCount)
	assert.Equal(t, 2, counts.IncludedCount)
	assert.Equal(t, 1, counts.CarvedOutCount)
	assert.Equal(t, 1, counts.DomainCount, "carved-out capabilities do not contribute to domain count")
}

func TestCompositionCounts_NilDomainExcludedFromDomainCount(t *testing.T) {
	resolved := ResolveComposition("ec-1",
		[]ActiveDirectionSources{direction("ec-1", "CRM", "cap-1", "cap-2")},
		[]CapabilityNode{capInDomain("cap-1", "", "dom-1", "Customer"), cap("cap-2", "")},
	)

	counts := CountComposition(2, resolved)
	assert.Equal(t, 1, counts.DomainCount)
}

func TestResolveCompositionWithIndex_MatchesResolveComposition(t *testing.T) {
	directions := []ActiveDirectionSources{
		direction("ec-crm", "CRM", "cap-identity"),
		direction("ec-pay", "Take Payment", "cap-fraud"),
	}
	capabilities := []CapabilityNode{
		capInDomain("cap-identity", "", "dom-1", "Customer"),
		capInDomain("cap-consent", "cap-identity", "dom-1", "Customer"),
		capInDomain("cap-fraud", "cap-identity", "dom-2", "Payments"),
	}

	index := BuildCapabilityIndex(capabilities)

	assert.Equal(t,
		ResolveComposition("ec-crm", directions, capabilities),
		ResolveCompositionWithIndex("ec-crm", directions, index),
	)
}

func TestBuildCapabilityIndex_IsReusableAcrossMultipleTargets(t *testing.T) {
	directions := []ActiveDirectionSources{
		direction("ec-1", "One", "cap-1"),
		direction("ec-2", "Two", "cap-2"),
	}
	capabilities := []CapabilityNode{cap("cap-1", ""), cap("cap-2", "")}
	index := BuildCapabilityIndex(capabilities)

	firstResolved := ResolveCompositionWithIndex("ec-1", directions, index)
	secondResolved := ResolveCompositionWithIndex("ec-2", directions, index)

	assert.Equal(t, []string{"cap-1"}, idsOf(firstResolved))
	assert.Equal(t, []string{"cap-2"}, idsOf(secondResolved))
}

func idsOf(resolved []ResolvedCapability) []string {
	ids := make([]string, len(resolved))
	for i, r := range resolved {
		ids[i] = r.Node.ID
	}
	return ids
}
