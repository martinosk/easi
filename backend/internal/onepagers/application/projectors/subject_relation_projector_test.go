package projectors_test

import (
	"context"
	"encoding/json"
	"testing"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/onepagers/application/projectors"
	"easi/backend/internal/onepagers/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type savedRelation struct {
	subject readmodels.SubjectKey
	entry   readmodels.RelationEntry
}

type replacedRelation struct {
	subject readmodels.SubjectKey
	entryID string
	entries []readmodels.RelationEntry
}

type edgeSubjectDeletion struct {
	edgeID     string
	subjectIDs []string
}

type renamedRelated struct {
	target readmodels.RelationTarget
	name   string
}

type fakeRelationStore struct {
	saved             []savedRelation
	replaced          []replacedRelation
	deletedEdges      []string
	deletedByRelated  []readmodels.RelationTarget
	deletedEdgeSubset []edgeSubjectDeletion
	deletedSubjects   []readmodels.SubjectKey
	renamed           []renamedRelated
	subjectsByEdge    map[string][]readmodels.SubjectKey
	subjectsByRelated map[readmodels.RelationTarget][]readmodels.SubjectKey
}

func (f *fakeRelationStore) Save(_ context.Context, subject readmodels.SubjectKey, entry readmodels.RelationEntry) error {
	f.saved = append(f.saved, savedRelation{subject: subject, entry: entry})
	return nil
}

func (f *fakeRelationStore) Replace(_ context.Context, subject readmodels.SubjectKey, entryID string, entries []readmodels.RelationEntry) error {
	f.replaced = append(f.replaced, replacedRelation{subject: subject, entryID: entryID, entries: entries})
	return nil
}

func (f *fakeRelationStore) DeleteByEdge(_ context.Context, edgeID string) error {
	f.deletedEdges = append(f.deletedEdges, edgeID)
	return nil
}

func (f *fakeRelationStore) DeleteEdgeForSubjects(_ context.Context, edgeID string, subjectIDs []string) error {
	f.deletedEdgeSubset = append(f.deletedEdgeSubset, edgeSubjectDeletion{edgeID: edgeID, subjectIDs: subjectIDs})
	return nil
}

func (f *fakeRelationStore) DeleteByRelated(_ context.Context, target readmodels.RelationTarget) error {
	f.deletedByRelated = append(f.deletedByRelated, target)
	return nil
}

func (f *fakeRelationStore) DeleteSubject(_ context.Context, subject readmodels.SubjectKey) error {
	f.deletedSubjects = append(f.deletedSubjects, subject)
	return nil
}

func (f *fakeRelationStore) RenameRelated(_ context.Context, target readmodels.RelationTarget, name string) error {
	f.renamed = append(f.renamed, renamedRelated{target: target, name: name})
	return nil
}

func (f *fakeRelationStore) SubjectsByEdge(_ context.Context, edgeID string) ([]readmodels.SubjectKey, error) {
	return f.subjectsByEdge[edgeID], nil
}

func (f *fakeRelationStore) SubjectsByRelated(_ context.Context, target readmodels.RelationTarget) ([]readmodels.SubjectKey, error) {
	return f.subjectsByRelated[target], nil
}

type recomputedCompleteness struct {
	subjectType string
	subjectIDs  []string
}

type fakeCompletenessRecomputer struct {
	calls []recomputedCompleteness
}

func (f *fakeCompletenessRecomputer) Recompute(_ context.Context, subjectType string, subjectIDs []string) error {
	f.calls = append(f.calls, recomputedCompleteness{subjectType: subjectType, subjectIDs: subjectIDs})
	return nil
}

func (f *fakeCompletenessRecomputer) subjectIDsFor(subjectType string) []string {
	for _, call := range f.calls {
		if call.subjectType == subjectType {
			return call.subjectIDs
		}
	}
	return nil
}

type fakeBusinessDomainNames struct {
	names   map[string]string
	deleted []string
}

func (f *fakeBusinessDomainNames) Save(_ context.Context, businessDomainID, name string) error {
	if f.names == nil {
		f.names = map[string]string{}
	}
	f.names[businessDomainID] = name
	return nil
}

func (f *fakeBusinessDomainNames) Delete(_ context.Context, businessDomainID string) error {
	f.deleted = append(f.deleted, businessDomainID)
	return nil
}

func (f *fakeBusinessDomainNames) Name(_ context.Context, businessDomainID string) (string, error) {
	return f.names[businessDomainID], nil
}

type relationHarness struct {
	t            *testing.T
	relations    *fakeRelationStore
	domains      *fakeBusinessDomainNames
	completeness *fakeCompletenessRecomputer
	projector    *projectors.SubjectRelationProjector
}

func newRelationHarness(t *testing.T) *relationHarness {
	relations := &fakeRelationStore{}
	domains := &fakeBusinessDomainNames{}
	completeness := &fakeCompletenessRecomputer{}
	return &relationHarness{
		t: t, relations: relations, domains: domains, completeness: completeness,
		projector: projectors.NewSubjectRelationProjector(relations, domains, completeness),
	}
}

func (h *relationHarness) project(eventType string, payload map[string]any) {
	h.t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(h.t, err)
	require.NoError(h.t, h.projector.ProjectEvent(context.Background(), eventType, data))
}

func TestSubjectRelationProjector_CachesSupplierRelations(t *testing.T) {
	cases := []struct {
		name      string
		eventType string
		payload   map[string]any
		want      []savedRelation
	}{
		{
			name:      "a realization is cached from both the capability and the application",
			eventType: capPL.SystemLinkedToCapability,
			payload:   map[string]any{"id": "r-1", "capabilityId": "cap-1", "componentId": "app-9"},
			want: []savedRelation{
				{
					subject: subjectKey("capability", "cap-1"),
					entry:   readmodels.RelationEntry{EntryID: "realizing-applications", RelatedType: "application", RelatedID: "app-9", EdgeID: "r-1"},
				},
				{
					subject: subjectKey("application", "app-9"),
					entry:   readmodels.RelationEntry{EntryID: "realized-capabilities", RelatedType: "capability", RelatedID: "cap-1", EdgeID: "r-1"},
				},
			},
		},
		{
			name:      "a new capability is cached under its parent",
			eventType: capPL.CapabilityCreated,
			payload:   map[string]any{"id": "cap-2", "name": "Billing", "parentId": "cap-1"},
			want: []savedRelation{
				{
					subject: subjectKey("capability", "cap-2"),
					entry:   readmodels.RelationEntry{EntryID: "parent-capability", RelatedType: "capability", RelatedID: "cap-1"},
				},
				{
					subject: subjectKey("capability", "cap-1"),
					entry:   readmodels.RelationEntry{EntryID: "child-capabilities", RelatedType: "capability", RelatedID: "cap-2"},
				},
			},
		},
		{
			name:      "a dependency is cached on the source capability only",
			eventType: capPL.CapabilityDependencyCreated,
			payload:   map[string]any{"id": "d-1", "sourceCapabilityId": "cap-1", "targetCapabilityId": "cap-2"},
			want: []savedRelation{{
				subject: subjectKey("capability", "cap-1"),
				entry:   readmodels.RelationEntry{EntryID: "depends-on", RelatedType: "capability", RelatedID: "cap-2", EdgeID: "d-1"},
			}},
		},
		{
			name:      "a component relation is cached on the source application only",
			eventType: amPL.ComponentRelationCreated,
			payload:   map[string]any{"id": "cr-1", "sourceComponentId": "app-1", "targetComponentId": "app-2", "relationType": "triggers"},
			want: []savedRelation{{
				subject: subjectKey("application", "app-1"),
				entry:   readmodels.RelationEntry{EntryID: "component-relations", RelatedType: "application", RelatedID: "app-2", EdgeID: "cr-1"},
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newRelationHarness(t)

			h.project(tc.eventType, tc.payload)

			assert.Equal(t, tc.want, h.relations.saved)
		})
	}
}

func TestSubjectRelationProjector_EdgeDeletions_RemoveTheCachedEdge(t *testing.T) {
	cases := map[string]map[string]any{
		capPL.SystemRealizationDeleted:       {"id": "edge-1"},
		capPL.CapabilityDependencyDeleted:    {"id": "edge-1"},
		capPL.CapabilityUnassignedFromDomain: {"id": "edge-1", "businessDomainId": "bd-1", "capabilityId": "cap-1"},
		amPL.ComponentRelationDeleted:        {"id": "edge-1", "sourceComponentID": "app-1", "targetComponentID": "app-2"},
	}

	for eventType, payload := range cases {
		t.Run(eventType, func(t *testing.T) {
			h := newRelationHarness(t)

			h.project(eventType, payload)

			assert.Equal(t, []string{"edge-1"}, h.relations.deletedEdges)
			assert.Empty(t, h.relations.saved)
		})
	}
}

func TestSubjectRelationProjector_InheritedRealizations_CacheUnderSourceRealization(t *testing.T) {
	h := newRelationHarness(t)

	h.project(capPL.CapabilityRealizationsInherited, map[string]any{
		"capabilityId": "cap-parent",
		"inheritedRealizations": []map[string]any{
			{"capabilityId": "cap-child", "componentId": "app-9", "sourceRealizationId": "r-1"},
		},
	})

	require.Len(t, h.relations.saved, 2)
	assert.Equal(t, subjectKey("capability", "cap-child"), h.relations.saved[0].subject)
	assert.Equal(t, "r-1", h.relations.saved[0].entry.EdgeID)
	assert.Equal(t, "realizing-applications", h.relations.saved[0].entry.EntryID)
	assert.Equal(t, subjectKey("application", "app-9"), h.relations.saved[1].subject)
	assert.Equal(t, "realized-capabilities", h.relations.saved[1].entry.EntryID)
}

func TestSubjectRelationProjector_UninheritedRealizations_RemoveOnlyNamedCapabilities(t *testing.T) {
	h := newRelationHarness(t)

	h.project(capPL.CapabilityRealizationsUninherited, map[string]any{
		"capabilityId": "cap-parent",
		"removals": []map[string]any{
			{"sourceRealizationId": "r-1", "capabilityIds": []string{"cap-child", "cap-other"}},
		},
	})

	assert.Equal(t, []edgeSubjectDeletion{{edgeID: "r-1", subjectIDs: []string{"cap-child", "cap-other"}}}, h.relations.deletedEdgeSubset)
	assert.Empty(t, h.relations.deletedEdges, "uninheriting never removes the source realization itself")
}

func TestSubjectRelationProjector_DomainAssignment_LabelsFromBusinessDomainNameCache(t *testing.T) {
	h := newRelationHarness(t)
	require.NoError(t, h.domains.Save(context.Background(), "bd-1", "Finance"))

	h.project(capPL.CapabilityAssignedToDomain, map[string]any{"id": "a-1", "businessDomainId": "bd-1", "capabilityId": "cap-1"})

	assert.Equal(t, []savedRelation{{
		subject: subjectKey("capability", "cap-1"),
		entry:   readmodels.RelationEntry{EntryID: "business-domains", RelatedID: "bd-1", RelatedName: "Finance", EdgeID: "a-1"},
	}}, h.relations.saved)
}

func TestSubjectRelationProjector_BusinessDomainRenamed_RefreshesCachedLabels(t *testing.T) {
	h := newRelationHarness(t)

	h.project(capPL.BusinessDomainUpdated, map[string]any{"id": "bd-1", "name": "Finance & Risk"})

	assert.Equal(t, "Finance & Risk", h.domains.names["bd-1"])
	assert.Equal(t, []renamedRelated{{target: readmodels.RelationTarget{EntryID: "business-domains", RelatedID: "bd-1"}, name: "Finance & Risk"}}, h.relations.renamed)
}

func TestSubjectRelationProjector_BusinessDomainDeleted_DropsAssignments(t *testing.T) {
	h := newRelationHarness(t)

	h.project(capPL.BusinessDomainDeleted, map[string]any{"id": "bd-1"})

	assert.Equal(t, []string{"bd-1"}, h.domains.deleted)
	assert.Equal(t, []readmodels.RelationTarget{{EntryID: "business-domains", RelatedID: "bd-1"}}, h.relations.deletedByRelated)
}

func TestSubjectRelationProjector_CapabilityCreatedWithoutParent_CachesNothing(t *testing.T) {
	h := newRelationHarness(t)

	h.project(capPL.CapabilityCreated, map[string]any{"id": "cap-1", "name": "Billing"})

	assert.Empty(t, h.relations.saved)
	assert.Empty(t, h.relations.replaced)
}

func TestSubjectRelationProjector_ParentChanged_MovesTheChild(t *testing.T) {
	h := newRelationHarness(t)

	h.project(capPL.CapabilityParentChanged, map[string]any{"capabilityId": "cap-2", "oldParentId": "cap-1", "newParentId": "cap-3"})

	assert.Equal(t, []replacedRelation{{
		subject: subjectKey("capability", "cap-2"),
		entryID: "parent-capability",
		entries: []readmodels.RelationEntry{{EntryID: "parent-capability", RelatedType: "capability", RelatedID: "cap-3"}},
	}}, h.relations.replaced)
	assert.Equal(t, []readmodels.RelationTarget{{EntryID: "child-capabilities", RelatedID: "cap-2"}}, h.relations.deletedByRelated)
	assert.Equal(t, []savedRelation{{
		subject: subjectKey("capability", "cap-3"),
		entry:   readmodels.RelationEntry{EntryID: "child-capabilities", RelatedType: "capability", RelatedID: "cap-2"},
	}}, h.relations.saved)
}

func TestSubjectRelationProjector_ParentCleared_LeavesNoHierarchy(t *testing.T) {
	h := newRelationHarness(t)

	h.project(capPL.CapabilityParentChanged, map[string]any{"capabilityId": "cap-2", "oldParentId": "cap-1", "newParentId": ""})

	assert.Equal(t, []replacedRelation{{subject: subjectKey("capability", "cap-2"), entryID: "parent-capability"}}, h.relations.replaced)
	assert.Equal(t, []readmodels.RelationTarget{{EntryID: "child-capabilities", RelatedID: "cap-2"}}, h.relations.deletedByRelated)
	assert.Empty(t, h.relations.saved)
}

func TestSubjectRelationProjector_OriginLinkSet_CachesBothDirections(t *testing.T) {
	cases := []struct {
		originType  string
		forward     string
		mirror      string
		relatedType string
	}{
		{originType: "built-by", forward: "built-by", mirror: "built-applications", relatedType: "internal-team"},
		{originType: "purchased-from", forward: "purchased-from", mirror: "purchased-applications", relatedType: "vendor"},
		{originType: "acquired-via", forward: "acquired-via", mirror: "acquired-applications", relatedType: "acquired-entity"},
	}

	for _, tc := range cases {
		t.Run(tc.originType, func(t *testing.T) {
			h := newRelationHarness(t)

			h.project(amPL.OriginLinkSet, map[string]any{"componentId": "app-1", "originType": tc.originType, "entityId": "e-1"})

			assert.Equal(t, []replacedRelation{{
				subject: subjectKey("application", "app-1"),
				entryID: tc.forward,
				entries: []readmodels.RelationEntry{{EntryID: tc.forward, RelatedType: tc.relatedType, RelatedID: "e-1"}},
			}}, h.relations.replaced)
			assert.Equal(t, []readmodels.RelationTarget{{EntryID: tc.mirror, RelatedID: "app-1"}}, h.relations.deletedByRelated)
			assert.Equal(t, []savedRelation{{
				subject: subjectKey(tc.relatedType, "e-1"),
				entry:   readmodels.RelationEntry{EntryID: tc.mirror, RelatedType: "application", RelatedID: "app-1"},
			}}, h.relations.saved)
		})
	}
}

func TestSubjectRelationProjector_OriginLinkReplaced_UsesTheNewEntity(t *testing.T) {
	h := newRelationHarness(t)

	h.project(amPL.OriginLinkReplaced, map[string]any{
		"componentId": "app-1", "originType": "purchased-from", "oldEntityId": "v-1", "newEntityId": "v-2",
	})

	require.Len(t, h.relations.replaced, 1)
	assert.Equal(t, []readmodels.RelationEntry{{EntryID: "purchased-from", RelatedType: "vendor", RelatedID: "v-2"}}, h.relations.replaced[0].entries)
	assert.Equal(t, []savedRelation{{
		subject: subjectKey("vendor", "v-2"),
		entry:   readmodels.RelationEntry{EntryID: "purchased-applications", RelatedType: "application", RelatedID: "app-1"},
	}}, h.relations.saved)
}

func TestSubjectRelationProjector_OriginLinkGone_ClearsBothDirections(t *testing.T) {
	for _, eventType := range []string{amPL.OriginLinkCleared, amPL.OriginLinkDeleted} {
		t.Run(eventType, func(t *testing.T) {
			h := newRelationHarness(t)

			h.project(eventType, map[string]any{"componentId": "app-1", "originType": "built-by", "entityId": "t-1"})

			assert.Equal(t, []replacedRelation{{subject: subjectKey("application", "app-1"), entryID: "built-by"}}, h.relations.replaced)
			assert.Equal(t, []readmodels.RelationTarget{{EntryID: "built-applications", RelatedID: "app-1"}}, h.relations.deletedByRelated)
			assert.Empty(t, h.relations.saved)
		})
	}
}

func TestSubjectRelationProjector_OriginLinkNotesUpdated_IsIgnored(t *testing.T) {
	h := newRelationHarness(t)

	h.project(amPL.OriginLinkNotesUpdated, map[string]any{"componentId": "app-1", "originType": "built-by", "entityId": "t-1", "newNotes": "x"})

	assert.Empty(t, h.relations.saved)
	assert.Empty(t, h.relations.replaced)
	assert.Empty(t, h.relations.deletedByRelated)
}

func TestSubjectRelationProjector_SubjectDeleted_DropsEveryRelation(t *testing.T) {
	cases := map[string]readmodels.SubjectKey{
		capPL.CapabilityDeleted:          subjectKey("capability", "x"),
		amPL.ApplicationComponentDeleted: subjectKey("application", "x"),
		amPL.VendorDeleted:               subjectKey("vendor", "x"),
		amPL.AcquiredEntityDeleted:       subjectKey("acquired-entity", "x"),
		amPL.InternalTeamDeleted:         subjectKey("internal-team", "x"),
	}

	for eventType, want := range cases {
		t.Run(eventType, func(t *testing.T) {
			h := newRelationHarness(t)

			h.project(eventType, map[string]any{"id": "x"})

			assert.Equal(t, []readmodels.SubjectKey{want}, h.relations.deletedSubjects)
		})
	}
}

func TestSubjectRelationProjector_UnknownEvent_NoOp(t *testing.T) {
	h := newRelationHarness(t)

	h.project("SomethingElse", map[string]any{"id": "x"})

	assert.Empty(t, h.relations.saved)
	assert.Empty(t, h.relations.replaced)
	assert.Empty(t, h.relations.deletedEdges)
	assert.Empty(t, h.relations.deletedSubjects)
}

func TestSubjectRelationProjector_RealizationLinked_RecomputesBothSubjects(t *testing.T) {
	h := newRelationHarness(t)

	h.project(capPL.SystemLinkedToCapability, map[string]any{"id": "r-1", "capabilityId": "cap-1", "componentId": "app-9"})

	assert.ElementsMatch(t, []string{"cap-1"}, h.completeness.subjectIDsFor("capability"))
	assert.ElementsMatch(t, []string{"app-9"}, h.completeness.subjectIDsFor("application"))
}

func TestSubjectRelationProjector_DependencyCreated_RecomputesSourceCapability(t *testing.T) {
	h := newRelationHarness(t)

	h.project(capPL.CapabilityDependencyCreated, map[string]any{"id": "d-1", "sourceCapabilityId": "cap-1", "targetCapabilityId": "cap-2"})

	assert.Equal(t, []string{"cap-1"}, h.completeness.subjectIDsFor("capability"))
}

func TestSubjectRelationProjector_EdgeDeleted_RecomputesSubjectsThatReferencedTheEdge(t *testing.T) {
	h := newRelationHarness(t)
	h.relations.subjectsByEdge = map[string][]readmodels.SubjectKey{
		"edge-1": {subjectKey("capability", "cap-1"), subjectKey("capability", "cap-2")},
	}

	h.project(capPL.CapabilityDependencyDeleted, map[string]any{"id": "edge-1"})

	assert.ElementsMatch(t, []string{"cap-1", "cap-2"}, h.completeness.subjectIDsFor("capability"))
}

func TestSubjectRelationProjector_EdgeDeleted_LooksUpAffectedSubjectsBeforeDeleting(t *testing.T) {
	h := newRelationHarness(t)
	h.relations.subjectsByEdge = map[string][]readmodels.SubjectKey{"edge-1": {subjectKey("capability", "cap-1")}}

	h.project(capPL.SystemRealizationDeleted, map[string]any{"id": "edge-1"})

	assert.Equal(t, []string{"edge-1"}, h.relations.deletedEdges)
	assert.Equal(t, []string{"cap-1"}, h.completeness.subjectIDsFor("capability"))
}

func TestSubjectRelationProjector_BusinessDomainDeleted_RecomputesFormerlyAssignedCapabilities(t *testing.T) {
	h := newRelationHarness(t)
	target := readmodels.RelationTarget{EntryID: "business-domains", RelatedID: "bd-1"}
	h.relations.subjectsByRelated = map[readmodels.RelationTarget][]readmodels.SubjectKey{
		target: {subjectKey("capability", "cap-1")},
	}

	h.project(capPL.BusinessDomainDeleted, map[string]any{"id": "bd-1"})

	assert.Equal(t, []string{"cap-1"}, h.completeness.subjectIDsFor("capability"))
}

func TestSubjectRelationProjector_BusinessDomainRenamed_DoesNotRecompute(t *testing.T) {
	h := newRelationHarness(t)

	h.project(capPL.BusinessDomainUpdated, map[string]any{"id": "bd-1", "name": "Finance & Risk"})

	assert.Empty(t, h.completeness.calls, "renaming a related entity never changes a required field's filled state")
}

func TestSubjectRelationProjector_ParentChanged_RecomputesCapabilityAndBothParents(t *testing.T) {
	h := newRelationHarness(t)
	target := readmodels.RelationTarget{EntryID: "child-capabilities", RelatedID: "cap-2"}
	h.relations.subjectsByRelated = map[readmodels.RelationTarget][]readmodels.SubjectKey{
		target: {subjectKey("capability", "cap-1")},
	}

	h.project(capPL.CapabilityParentChanged, map[string]any{"capabilityId": "cap-2", "oldParentId": "cap-1", "newParentId": "cap-3"})

	assert.ElementsMatch(t, []string{"cap-2", "cap-1", "cap-3"}, h.completeness.subjectIDsFor("capability"))
}

func TestSubjectRelationProjector_OriginLinkReplaced_RecomputesApplicationAndBothEntities(t *testing.T) {
	h := newRelationHarness(t)
	target := readmodels.RelationTarget{EntryID: "purchased-applications", RelatedID: "app-1"}
	h.relations.subjectsByRelated = map[readmodels.RelationTarget][]readmodels.SubjectKey{
		target: {subjectKey("vendor", "v-1")},
	}

	h.project(amPL.OriginLinkReplaced, map[string]any{
		"componentId": "app-1", "originType": "purchased-from", "oldEntityId": "v-1", "newEntityId": "v-2",
	})

	assert.Equal(t, []string{"app-1"}, h.completeness.subjectIDsFor("application"))
	assert.ElementsMatch(t, []string{"v-1", "v-2"}, h.completeness.subjectIDsFor("vendor"))
}

func TestSubjectRelationEventTypes_CoverEveryCachedRelation(t *testing.T) {
	types := projectors.SubjectRelationEventTypes()

	for _, eventType := range []string{
		capPL.CapabilityCreated, capPL.CapabilityDeleted, capPL.CapabilityParentChanged,
		capPL.SystemLinkedToCapability, capPL.SystemRealizationDeleted,
		capPL.CapabilityRealizationsInherited, capPL.CapabilityRealizationsUninherited,
		capPL.CapabilityDependencyCreated, capPL.CapabilityDependencyDeleted,
		capPL.CapabilityAssignedToDomain, capPL.CapabilityUnassignedFromDomain,
		capPL.BusinessDomainCreated, capPL.BusinessDomainUpdated, capPL.BusinessDomainDeleted,
		amPL.ComponentRelationCreated, amPL.ComponentRelationDeleted,
		amPL.OriginLinkSet, amPL.OriginLinkReplaced, amPL.OriginLinkCleared, amPL.OriginLinkDeleted,
		amPL.ApplicationComponentDeleted, amPL.VendorDeleted, amPL.AcquiredEntityDeleted, amPL.InternalTeamDeleted,
	} {
		assert.Containsf(t, types, eventType, "%s must be subscribed", eventType)
	}
}
