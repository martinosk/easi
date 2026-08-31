//go:build integration

package projectors_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/projectors"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/infrastructure/adapters"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type relationCacheFixture struct {
	t         *testing.T
	ctx       context.Context
	tenant    string
	index     *readmodels.OnePagerSubjectIndexReadModel
	relations *readmodels.SubjectRelationCacheReadModel
	configs   *readmodels.OnePagerConfigurationReadModel
	projector *projectors.SubjectRelationProjector
}

func newRelationCacheFixture(t *testing.T) *relationCacheFixture {
	t.Helper()
	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi_app password=localdev dbname=easi sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	tenant := "test-rel-" + uuid.NewString()[:8]
	tenantID, err := sharedvo.NewTenantID(tenant)
	require.NoError(t, err)

	t.Cleanup(func() {
		for _, table := range []string{
			"onepagers.subject_relation_cache",
			"onepagers.business_domain_name_cache",
			"onepagers.one_pager_subject_index",
			"onepagers.one_pager_configurations",
		} {
			_, _ = db.Exec("DELETE FROM "+table+" WHERE tenant_id = $1", tenant)
		}
		_ = db.Close()
	})

	tenantDB := database.NewTenantAwareDB(db)
	index := readmodels.NewOnePagerSubjectIndexReadModel(tenantDB)
	relations := readmodels.NewSubjectRelationCacheReadModel(tenantDB)
	configs := readmodels.NewOnePagerConfigurationReadModel(tenantDB)
	counter := queries.NewCompletenessIndicators(configs, readmodels.NewOnePagerFactsReadModel(tenantDB),
		adapters.NewOnePagerBuiltInFieldSources(tenantDB))
	indexProjector := projectors.NewSubjectIndexProjector(index, counter, adapters.NewSubjectAuditAdapter(tenantDB), configs)
	return &relationCacheFixture{
		t: t, ctx: sharedctx.WithTenant(context.Background(), tenantID), tenant: tenant,
		index: index, relations: relations, configs: configs,
		projector: projectors.NewSubjectRelationProjector(relations, readmodels.NewBusinessDomainNameCacheReadModel(tenantDB), indexProjector),
	}
}

func (f *relationCacheFixture) requireBuiltIn(subjectType, entryID string) {
	f.t.Helper()
	require.NoError(f.t, f.configs.Insert(f.ctx, readmodels.ConfigurationRecord{
		ID:          uuid.NewString(),
		SubjectType: subjectType,
		Document: readmodels.ConfigurationDocument{
			CustomFields:  []readmodels.CustomFieldRecord{},
			BuiltInFields: []readmodels.BuiltInFieldRecord{{ID: entryID, Required: true}},
			DisplayOrder:  []readmodels.FieldRefRecord{{Kind: "builtIn", ID: entryID}},
		},
		Version: 1, CreatedAt: time.Now().UTC(), ModifiedAt: time.Now().UTC(), ModifiedBy: "admin",
	}))
}

func (f *relationCacheFixture) rowFor(subject readmodels.SubjectKey) (readmodels.SubjectIndexRecord, bool) {
	f.t.Helper()
	page, _, err := f.index.Page(f.ctx, readmodels.SubjectIndexQuery{
		SubjectTypes: []string{subject.SubjectType}, Sort: readmodels.SortName, Order: readmodels.OrderAsc, Limit: 50,
	})
	require.NoError(f.t, err)
	for _, record := range page {
		if record.SubjectID == subject.SubjectID {
			return record, true
		}
	}
	return readmodels.SubjectIndexRecord{}, false
}

func (f *relationCacheFixture) seedSubject(subject readmodels.SubjectKey, name string) {
	f.t.Helper()
	require.NoError(f.t, f.index.Upsert(f.ctx, readmodels.SubjectIndexRecord{
		SubjectType: subject.SubjectType, SubjectID: subject.SubjectID, Name: name,
		CreatedAt: time.Now().UTC(), LastUpdatedAt: time.Now().UTC(),
	}))
}

func (f *relationCacheFixture) project(eventType string, payload map[string]any) {
	f.t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(f.t, err)
	require.NoError(f.t, f.projector.ProjectEvent(f.ctx, eventType, data))
}

func (f *relationCacheFixture) references(subject readmodels.SubjectKey, entryIDs ...string) []readmodels.RelationReference {
	f.t.Helper()
	references, err := f.relations.References(f.ctx, readmodels.RelationQuery{SubjectType: subject.SubjectType, SubjectIDs: []string{subject.SubjectID}, EntryIDs: entryIDs})
	require.NoError(f.t, err)
	return references[subject.SubjectID]
}

func TestSubjectRelationCache_LabelsFollowRenamesOfTheRelatedSubject(t *testing.T) {
	f := newRelationCacheFixture(t)
	f.seedSubject(subjectKey("capability", "cap-1"), "Billing")
	f.seedSubject(subjectKey("application", "app-1"), "Billing Service")

	f.project(capPL.SystemLinkedToCapability, map[string]any{"id": "rz-1", "capabilityId": "cap-1", "componentId": "app-1"})

	references := f.references(subjectKey("capability", "cap-1"), "realizing-applications")
	require.Len(t, references, 1)
	assert.Equal(t, "Billing Service", references[0].Label)
	assert.Equal(t, "application", references[0].RelatedType)

	f.seedSubject(subjectKey("application", "app-1"), "Invoicing Service")

	references = f.references(subjectKey("capability", "cap-1"), "realizing-applications")
	require.Len(t, references, 1)
	assert.Equal(t, "Invoicing Service", references[0].Label, "the label is joined live from the subject index")
}

func TestSubjectRelationCache_DeletingARealizationRemovesBothDirections(t *testing.T) {
	f := newRelationCacheFixture(t)
	f.seedSubject(subjectKey("capability", "cap-1"), "Billing")
	f.seedSubject(subjectKey("application", "app-1"), "Billing Service")
	f.project(capPL.SystemLinkedToCapability, map[string]any{"id": "rz-1", "capabilityId": "cap-1", "componentId": "app-1"})

	f.project(capPL.SystemRealizationDeleted, map[string]any{"id": "rz-1"})

	assert.Empty(t, f.references(subjectKey("capability", "cap-1"), "realizing-applications"))
	assert.Empty(t, f.references(subjectKey("application", "app-1"), "realized-capabilities"))
}

func TestSubjectRelationCache_DeletingASubjectRemovesRelationsPointingAtIt(t *testing.T) {
	f := newRelationCacheFixture(t)
	f.seedSubject(subjectKey("capability", "cap-1"), "Billing")
	f.seedSubject(subjectKey("application", "app-1"), "Billing Service")
	f.project(capPL.SystemLinkedToCapability, map[string]any{"id": "rz-1", "capabilityId": "cap-1", "componentId": "app-1"})

	f.project(amPL.ApplicationComponentDeleted, map[string]any{"id": "app-1"})

	assert.Empty(t, f.references(subjectKey("capability", "cap-1"), "realizing-applications"))
}

func TestSubjectRelationCache_OriginLinkReplacementMovesTheMirrorRelation(t *testing.T) {
	f := newRelationCacheFixture(t)
	f.seedSubject(subjectKey("application", "app-1"), "Billing Service")
	f.seedSubject(subjectKey("vendor", "v-1"), "Contoso")
	f.seedSubject(subjectKey("vendor", "v-2"), "Fabrikam")
	f.project(amPL.OriginLinkSet, map[string]any{"componentId": "app-1", "originType": "purchased-from", "entityId": "v-1"})

	f.project(amPL.OriginLinkReplaced, map[string]any{
		"componentId": "app-1", "originType": "purchased-from", "oldEntityId": "v-1", "newEntityId": "v-2",
	})

	purchasedFrom := f.references(subjectKey("application", "app-1"), "purchased-from")
	require.Len(t, purchasedFrom, 1)
	assert.Equal(t, "v-2", purchasedFrom[0].RelatedID)
	assert.Equal(t, "Fabrikam", purchasedFrom[0].Label)
	assert.Empty(t, f.references(subjectKey("vendor", "v-1"), "purchased-applications"), "the old vendor no longer lists the application")
	require.Len(t, f.references(subjectKey("vendor", "v-2"), "purchased-applications"), 1)
}

func TestSubjectRelationCache_BusinessDomainLabelsComeFromTheDomainCache(t *testing.T) {
	f := newRelationCacheFixture(t)
	f.seedSubject(subjectKey("capability", "cap-1"), "Billing")
	f.project(capPL.BusinessDomainCreated, map[string]any{"id": "bd-1", "name": "Finance"})

	f.project(capPL.CapabilityAssignedToDomain, map[string]any{"id": "asg-1", "businessDomainId": "bd-1", "capabilityId": "cap-1"})

	references := f.references(subjectKey("capability", "cap-1"), "business-domains")
	require.Len(t, references, 1)
	assert.Equal(t, "Finance", references[0].Label)
	assert.Empty(t, references[0].RelatedType, "a business domain is not a one-pager subject")

	f.project(capPL.BusinessDomainUpdated, map[string]any{"id": "bd-1", "name": "Finance & Risk"})

	references = f.references(subjectKey("capability", "cap-1"), "business-domains")
	require.Len(t, references, 1)
	assert.Equal(t, "Finance & Risk", references[0].Label)
}

func TestSubjectRelationCache_ReparentingMovesTheChild(t *testing.T) {
	f := newRelationCacheFixture(t)
	for _, id := range []string{"cap-1", "cap-2", "cap-3"} {
		f.seedSubject(subjectKey("capability", id), id)
	}
	f.project(capPL.CapabilityCreated, map[string]any{"id": "cap-2", "name": "Billing", "parentId": "cap-1"})

	f.project(capPL.CapabilityParentChanged, map[string]any{"capabilityId": "cap-2", "oldParentId": "cap-1", "newParentId": "cap-3"})

	parent := f.references(subjectKey("capability", "cap-2"), "parent-capability")
	require.Len(t, parent, 1)
	assert.Equal(t, "cap-3", parent[0].RelatedID)
	assert.Empty(t, f.references(subjectKey("capability", "cap-1"), "child-capabilities"))
	require.Len(t, f.references(subjectKey("capability", "cap-3"), "child-capabilities"), 1)
}

func TestSubjectRelationCache_ParallelEdgesSurviveIndependentDeletion(t *testing.T) {
	f := newRelationCacheFixture(t)
	f.seedSubject(subjectKey("capability", "cap-1"), "Billing")
	f.seedSubject(subjectKey("capability", "cap-2"), "Compliance")

	f.project(capPL.CapabilityDependencyCreated, map[string]any{"id": "dep-1", "sourceCapabilityId": "cap-1", "targetCapabilityId": "cap-2"})
	f.project(capPL.CapabilityDependencyCreated, map[string]any{"id": "dep-2", "sourceCapabilityId": "cap-1", "targetCapabilityId": "cap-2"})

	f.project(capPL.CapabilityDependencyDeleted, map[string]any{"id": "dep-2"})

	references := f.references(subjectKey("capability", "cap-1"), "depends-on")
	require.Len(t, references, 1, "dep-1 is a separate, still-live edge and must keep rendering")
	assert.Equal(t, "cap-2", references[0].RelatedID)

	f.project(capPL.CapabilityDependencyDeleted, map[string]any{"id": "dep-1"})
	assert.Empty(t, f.references(subjectKey("capability", "cap-1"), "depends-on"), "both parallel edges are now gone")
}

func TestSubjectRelationCache_ReferencesDeduplicateParallelEdgesToTheSameTarget(t *testing.T) {
	f := newRelationCacheFixture(t)
	f.seedSubject(subjectKey("application", "app-1"), "Billing Service")
	f.seedSubject(subjectKey("application", "app-2"), "Invoicing Service")

	require.NoError(t, f.relations.Save(f.ctx,
		subjectKey("application", "app-1"),
		readmodels.RelationEntry{EntryID: "component-relations", RelatedType: "application", RelatedID: "app-2", EdgeID: "cr-1"},
	))
	require.NoError(t, f.relations.Save(f.ctx,
		subjectKey("application", "app-1"),
		readmodels.RelationEntry{EntryID: "component-relations", RelatedType: "application", RelatedID: "app-2", EdgeID: "cr-2"},
	))

	references := f.references(subjectKey("application", "app-1"), "component-relations")
	require.Len(t, references, 1, "two parallel edges to the same target render once")
	assert.Equal(t, "app-2", references[0].RelatedID)
}

func TestSubjectRelationCache_CreatingAndDeletingARequiredRelation_FlipsStoredCompleteness(t *testing.T) {
	f := newRelationCacheFixture(t)
	f.seedSubject(subjectKey("capability", "cap-1"), "Billing")
	f.seedSubject(subjectKey("capability", "cap-2"), "Compliance")
	f.requireBuiltIn("capability", "depends-on")

	f.project(capPL.CapabilityDependencyCreated, map[string]any{"id": "dep-1", "sourceCapabilityId": "cap-1", "targetCapabilityId": "cap-2"})

	row, found := f.rowFor(subjectKey("capability", "cap-1"))
	require.True(t, found)
	assert.Equal(t, 1, row.RequiredCount)
	assert.Equal(t, 1, row.FilledCount, "the newly created required relation must flip the stored counters immediately")
	assert.Equal(t, readmodels.SignalComplete, row.Signal())

	f.project(capPL.CapabilityDependencyDeleted, map[string]any{"id": "dep-1"})

	row, found = f.rowFor(subjectKey("capability", "cap-1"))
	require.True(t, found)
	assert.Equal(t, 0, row.FilledCount, "deleting the only relation must flip the stored counters back")
	assert.Equal(t, readmodels.SignalIncomplete, row.Signal())
}

func TestSubjectRelationCache_DeletingASubjectRecomputesCompletenessOfSubjectsThatReferencedIt(t *testing.T) {
	f := newRelationCacheFixture(t)
	f.seedSubject(subjectKey("vendor", "v-1"), "Contoso")
	f.seedSubject(subjectKey("application", "app-1"), "Billing Service")
	f.requireBuiltIn("vendor", "purchased-applications")
	f.project(amPL.OriginLinkSet, map[string]any{"componentId": "app-1", "originType": "purchased-from", "entityId": "v-1"})

	row, found := f.rowFor(subjectKey("vendor", "v-1"))
	require.True(t, found)
	assert.Equal(t, 1, row.FilledCount, "the vendor's purchased-applications relation is filled once the origin link is set")
	assert.Equal(t, readmodels.SignalComplete, row.Signal())

	f.project(amPL.ApplicationComponentDeleted, map[string]any{"id": "app-1"})

	row, found = f.rowFor(subjectKey("vendor", "v-1"))
	require.True(t, found)
	assert.Equal(t, 0, row.FilledCount, "deleting the application must clear the vendor's stored completeness for the relation that referenced it")
	assert.Equal(t, readmodels.SignalIncomplete, row.Signal())
}

func TestSubjectRelationCache_UninheritingRemovesOnlyTheNamedCapabilities(t *testing.T) {
	f := newRelationCacheFixture(t)
	f.seedSubject(subjectKey("application", "app-1"), "Billing Service")
	for _, id := range []string{"cap-1", "cap-2"} {
		f.seedSubject(subjectKey("capability", id), id)
	}
	f.project(capPL.CapabilityRealizationsInherited, map[string]any{
		"capabilityId": "cap-0",
		"inheritedRealizations": []map[string]any{
			{"capabilityId": "cap-1", "componentId": "app-1", "sourceRealizationId": "rz-1"},
			{"capabilityId": "cap-2", "componentId": "app-1", "sourceRealizationId": "rz-1"},
		},
	})

	f.project(capPL.CapabilityRealizationsUninherited, map[string]any{
		"capabilityId": "cap-0",
		"removals":     []map[string]any{{"sourceRealizationId": "rz-1", "capabilityIds": []string{"cap-1"}}},
	})

	assert.Empty(t, f.references(subjectKey("capability", "cap-1"), "realizing-applications"))
	require.Len(t, f.references(subjectKey("capability", "cap-2"), "realizing-applications"), 1)
}
