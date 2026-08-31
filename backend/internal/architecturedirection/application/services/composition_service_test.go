package services

import (
	"context"
	"testing"

	domainservices "easi/backend/internal/architecturedirection/domain/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDirectionSources struct {
	items []DirectionSources
	err   error
	calls int
}

func (f *fakeDirectionSources) ActiveDirectionSources(_ context.Context) ([]DirectionSources, error) {
	f.calls++
	return f.items, f.err
}

type fakeCapabilityNodes struct {
	nodes []domainservices.CapabilityNode
	calls int
}

func (f *fakeCapabilityNodes) AllCapabilityNodes(_ context.Context) ([]domainservices.CapabilityNode, error) {
	f.calls++
	return f.nodes, nil
}

type fakeECNames struct {
	active map[string]string
	calls  int
}

func (f *fakeECNames) ActiveEnterpriseCapabilityNames(_ context.Context) (map[string]string, error) {
	f.calls++
	return f.active, nil
}

func node(id, parentID, domainID, domainName string) domainservices.CapabilityNode {
	return domainservices.CapabilityNode{
		ID: id, ParentID: parentID, Name: "Cap " + id, Level: "L2",
		BusinessDomainID: domainID, BusinessDomainName: domainName,
	}
}

func newService(directions []DirectionSources, nodes []domainservices.CapabilityNode, activeECs map[string]string) *CompositionService {
	return NewCompositionService(
		&fakeDirectionSources{items: directions},
		&fakeCapabilityNodes{nodes: nodes},
		&fakeECNames{active: activeECs},
	)
}

func TestCompositionForEC_NoActiveDirection(t *testing.T) {
	svc := newService(nil, []domainservices.CapabilityNode{node("cap-1", "", "", "")}, map[string]string{"ec-1": "CRM"})

	result, err := svc.CompositionForEC(context.Background(), "ec-1")

	require.NoError(t, err)
	assert.False(t, result.HasActiveDirection)
	assert.Empty(t, result.Resolved)
	assert.Equal(t, 0, result.Counts.SourceCount)
}

func TestCompositionForEC_ResolvesSourcesAndSubtrees(t *testing.T) {
	svc := newService(
		[]DirectionSources{{EnterpriseCapabilityID: "ec-1", Status: "draft", SourceCapabilityIDs: []string{"cap-1"}}},
		[]domainservices.CapabilityNode{node("cap-1", "", "dom-1", "Customer"), node("cap-2", "cap-1", "dom-1", "Customer")},
		map[string]string{"ec-1": "CRM"},
	)

	result, err := svc.CompositionForEC(context.Background(), "ec-1")

	require.NoError(t, err)
	assert.True(t, result.HasActiveDirection)
	assert.Equal(t, "draft", result.DirectionStatus)
	require.Len(t, result.Resolved, 2)
	assert.Equal(t, 1, result.Counts.SourceCount)
	assert.Equal(t, 2, result.Counts.IncludedCount)
	assert.Equal(t, 1, result.Counts.DomainCount)
}

func TestCompositionForEC_DirectionOnInactiveECIsIgnored(t *testing.T) {
	svc := newService(
		[]DirectionSources{
			{EnterpriseCapabilityID: "ec-1", Status: "draft", SourceCapabilityIDs: []string{"cap-1"}},
			{EnterpriseCapabilityID: "ec-gone", Status: "agreed", SourceCapabilityIDs: []string{"cap-2"}},
		},
		[]domainservices.CapabilityNode{node("cap-1", "", "", ""), node("cap-2", "cap-1", "", "")},
		map[string]string{"ec-1": "CRM"},
	)

	result, err := svc.CompositionForEC(context.Background(), "ec-1")

	require.NoError(t, err)
	roles := map[string]domainservices.CompositionRole{}
	for _, r := range result.Resolved {
		roles[r.Node.ID] = r.Role
	}
	assert.Equal(t, domainservices.RoleImplicit, roles["cap-2"], "a deleted EC's direction must not carve out")
}

func twoECsWithCarveOut() *CompositionService {
	return newService(
		[]DirectionSources{
			{EnterpriseCapabilityID: "ec-1", Status: "agreed", SourceCapabilityIDs: []string{"cap-1"}},
			{EnterpriseCapabilityID: "ec-2", Status: "draft", SourceCapabilityIDs: []string{"cap-2"}},
		},
		[]domainservices.CapabilityNode{
			node("cap-1", "", "dom-1", "Customer"),
			node("cap-2", "cap-1", "dom-2", "Payments"),
			node("cap-3", "cap-2", "dom-2", "Payments"),
		},
		map[string]string{"ec-1": "CRM", "ec-2": "Take Payment"},
	)
}

func TestSummariesForAll_SinglePassAcrossECs(t *testing.T) {
	summaries, err := twoECsWithCarveOut().SummariesForAll(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, summaries.Counts["ec-1"].IncludedCount, "cap-2 is carved out of ec-1 by ec-2")
	assert.Equal(t, 1, summaries.Counts["ec-1"].DomainCount)
	assert.Equal(t, 2, summaries.Counts["ec-2"].IncludedCount)
	assert.ElementsMatch(t, []string{"ec-1", "ec-2"}, summaries.EnterpriseCapabilityIDs)
	assert.Equal(t, "agreed", summaries.Statuses["ec-1"])
}

func TestSummariesForAll_IncludesActiveECsWithoutADirection(t *testing.T) {
	svc := newService(nil, nil, map[string]string{"ec-1": "CRM", "ec-2": "Identity"})

	summaries, err := svc.SummariesForAll(context.Background())

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"ec-1", "ec-2"}, summaries.EnterpriseCapabilityIDs)
	assert.Empty(t, summaries.Counts["ec-1"])
	assert.Empty(t, summaries.Statuses["ec-1"])
}

func TestSummariesForAll_LoadsEachInputExactlyOnce(t *testing.T) {
	directions := &fakeDirectionSources{items: []DirectionSources{
		{EnterpriseCapabilityID: "ec-1", Status: "draft", SourceCapabilityIDs: []string{"cap-1"}},
	}}
	nodes := &fakeCapabilityNodes{nodes: []domainservices.CapabilityNode{node("cap-1", "", "", "")}}
	ecNames := &fakeECNames{active: map[string]string{"ec-1": "CRM"}}
	svc := NewCompositionService(directions, nodes, ecNames)

	_, err := svc.SummariesForAll(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, directions.calls, "composition summaries must load direction sources exactly once")
	assert.Equal(t, 1, nodes.calls, "composition summaries must load capability nodes exactly once")
	assert.Equal(t, 1, ecNames.calls, "composition summaries must load enterprise capability names exactly once")
}

func TestIncludedCapabilityIDsByEC_ExcludesCarvedOutNodes(t *testing.T) {
	included, err := twoECsWithCarveOut().IncludedCapabilityIDsByEC(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []string{"cap-1"}, included["ec-1"], "cap-2 and its subtree are carved out by ec-2")
	assert.Equal(t, []string{"cap-2", "cap-3"}, included["ec-2"])
}

func TestPreview_UsesRequestedSourcesInsteadOfPersistedDirection(t *testing.T) {
	svc := newService(
		[]DirectionSources{{EnterpriseCapabilityID: "ec-1", Status: "draft", SourceCapabilityIDs: []string{"cap-old"}}},
		[]domainservices.CapabilityNode{node("cap-old", "", "", ""), node("cap-new", "", "", "")},
		map[string]string{"ec-1": "CRM"},
	)

	preview, err := svc.Preview(context.Background(), "ec-1", []string{"cap-new"})

	require.NoError(t, err)
	require.Len(t, preview.Resolved, 1)
	assert.Equal(t, "cap-new", preview.Resolved[0].Node.ID)
	assert.Equal(t, 1, preview.Counts.SourceCount)
}

func TestPreview_ReportsPerSourceEligibility(t *testing.T) {
	svc := newService(
		[]DirectionSources{{EnterpriseCapabilityID: "ec-other", Status: "proposed", SourceCapabilityIDs: []string{"cap-taken"}}},
		[]domainservices.CapabilityNode{node("cap-taken", "", "", ""), node("cap-free", "", "", "")},
		map[string]string{"ec-1": "CRM", "ec-other": "Take Payment"},
	)

	preview, err := svc.Preview(context.Background(), "ec-1", []string{"cap-taken", "cap-free"})

	require.NoError(t, err)
	require.Len(t, preview.SourceEligibility, 2)

	taken := preview.SourceEligibility[0]
	assert.Equal(t, "cap-taken", taken.CapabilityID)
	assert.False(t, taken.Eligible)
	assert.Equal(t, "Already an explicit source of an active direction on 'Take Payment'", *taken.IneligibilityReason)
	require.NotNil(t, taken.ConflictingEnterpriseCapability)
	assert.Equal(t, "ec-other", taken.ConflictingEnterpriseCapability.EnterpriseCapabilityID)

	free := preview.SourceEligibility[1]
	assert.True(t, free.Eligible)
	assert.Nil(t, free.IneligibilityReason)
	assert.Nil(t, free.ConflictingEnterpriseCapability)
}

func TestFirstSourceConflict_ReturnsFirstViolation(t *testing.T) {
	svc := newService(
		[]DirectionSources{{EnterpriseCapabilityID: "ec-other", Status: "draft", SourceCapabilityIDs: []string{"cap-taken"}}},
		[]domainservices.CapabilityNode{node("cap-taken", "", "", "")},
		map[string]string{"ec-1": "CRM", "ec-other": "Customer Identity"},
	)

	conflict, err := svc.FirstSourceConflict(context.Background(), "ec-1", []string{"cap-free", "cap-taken"})

	require.NoError(t, err)
	require.NotNil(t, conflict)
	assert.Equal(t, "cap-taken", conflict.CapabilityID)
	assert.Equal(t, "Cap cap-taken", conflict.CapabilityName)
	assert.Equal(t, "ec-other", conflict.ConflictingEnterpriseCapabilityID)
	assert.Equal(t, "Customer Identity", conflict.ConflictingEnterpriseCapabilityName)
}

func TestFirstSourceConflict_NilWhenAllEligible(t *testing.T) {
	svc := newService(nil, []domainservices.CapabilityNode{node("cap-1", "", "", "")}, map[string]string{"ec-1": "CRM"})

	conflict, err := svc.FirstSourceConflict(context.Background(), "ec-1", []string{"cap-1"})

	require.NoError(t, err)
	assert.Nil(t, conflict)
}

func TestSourceCandidates_FiltersByNameAndDomainWithEligibility(t *testing.T) {
	customer := node("cap-1", "", "dom-1", "Customer")
	customer.Name = "Customer Account Creation"
	fraud := node("cap-2", "", "dom-1", "Customer")
	fraud.Name = "Customer Fraud Prevention"
	payments := node("cap-3", "", "dom-2", "Payments")
	payments.Name = "Customer Payments"
	unrelated := node("cap-4", "", "dom-1", "Customer")
	unrelated.Name = "Order Handling"

	svc := newService(
		[]DirectionSources{{EnterpriseCapabilityID: "ec-other", Status: "draft", SourceCapabilityIDs: []string{"cap-2"}}},
		[]domainservices.CapabilityNode{customer, fraud, payments, unrelated},
		map[string]string{"ec-1": "CRM", "ec-other": "Take Payment"},
	)

	result, err := svc.SourceCandidates(context.Background(), SourceCandidatesQuery{
		EnterpriseCapabilityID: "ec-1",
		Search:                 "customer",
		BusinessDomainID:       "dom-1",
		Limit:                  10,
	})

	require.NoError(t, err)
	require.Len(t, result.Candidates, 2)
	assert.False(t, result.HasMore)
	assert.True(t, result.Candidates[0].Eligible)
	assert.False(t, result.Candidates[1].Eligible)
	assert.Equal(t, "ec-other", result.Candidates[1].ConflictingEnterpriseCapability.EnterpriseCapabilityID)
}

func TestSourceCandidates_LimitAndHasMore(t *testing.T) {
	a := node("cap-a", "", "", "")
	a.Name = "Customer A"
	b := node("cap-b", "", "", "")
	b.Name = "Customer B"

	svc := newService(nil, []domainservices.CapabilityNode{a, b}, map[string]string{"ec-1": "CRM"})

	result, err := svc.SourceCandidates(context.Background(), SourceCandidatesQuery{
		EnterpriseCapabilityID: "ec-1",
		Search:                 "customer",
		Limit:                  1,
	})

	require.NoError(t, err)
	assert.Len(t, result.Candidates, 1)
	assert.True(t, result.HasMore)
}
