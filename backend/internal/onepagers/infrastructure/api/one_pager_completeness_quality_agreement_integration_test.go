//go:build integration

package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/projectors"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/infrastructure/adapters"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletenessEndpointAndQualityList_AgreeAfterARelationIsAdded(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi_app password=localdev dbname=easi sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	tenant := "test-agree-" + uuid.NewString()[:8]
	tenantID, err := sharedvo.NewTenantID(tenant)
	require.NoError(t, err)
	ctx := sharedctx.WithTenant(sharedctx.WithActor(context.Background(), sharedctx.Actor{ID: "u1", Permissions: allReadPermissions()}), tenantID)

	t.Cleanup(func() {
		for _, table := range []string{
			"onepagers.subject_relation_cache",
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
	relationProjector := projectors.NewSubjectRelationProjector(relations, readmodels.NewBusinessDomainNameCacheReadModel(tenantDB), indexProjector)

	require.NoError(t, index.Upsert(ctx, readmodels.SubjectIndexRecord{
		SubjectType: "capability", SubjectID: "cap-1", Name: "Billing",
		CreatedAt: time.Now().UTC(), LastUpdatedAt: time.Now().UTC(),
	}))
	require.NoError(t, index.Upsert(ctx, readmodels.SubjectIndexRecord{
		SubjectType: "capability", SubjectID: "cap-2", Name: "Compliance",
		CreatedAt: time.Now().UTC(), LastUpdatedAt: time.Now().UTC(),
	}))
	require.NoError(t, configs.Insert(ctx, readmodels.ConfigurationRecord{
		ID: uuid.NewString(), SubjectType: "capability",
		Document: readmodels.ConfigurationDocument{
			CustomFields:  []readmodels.CustomFieldRecord{},
			BuiltInFields: []readmodels.BuiltInFieldRecord{{ID: "depends-on", Required: true}},
			DisplayOrder:  []readmodels.FieldRefRecord{{Kind: "builtIn", ID: "depends-on"}},
		},
		Version: 1, CreatedAt: time.Now().UTC(), ModifiedAt: time.Now().UTC(), ModifiedBy: "admin",
	}))

	payload, err := json.Marshal(map[string]any{"id": "dep-1", "sourceCapabilityId": "cap-1", "targetCapabilityId": "cap-2"})
	require.NoError(t, err)
	require.NoError(t, relationProjector.ProjectEvent(ctx, capPL.CapabilityDependencyCreated, payload))

	links := NewOnePagerLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	completenessHandlers := NewOnePagerCompletenessHandlers(index, links)
	qualityHandlers := NewOnePagerQualityHandlers(index, links)

	completenessRec := httptest.NewRecorder()
	completenessReq := httptest.NewRequest(http.MethodGet, "/one-pagers/capability/completeness", nil).WithContext(ctx)
	completenessHandlers.GetCompleteness("capability")(completenessRec, completenessReq)
	require.Equal(t, http.StatusOK, completenessRec.Code)
	var completenessBody OnePagerCompletenessResponse
	require.NoError(t, json.Unmarshal(completenessRec.Body.Bytes(), &completenessBody))

	qualityRec := httptest.NewRecorder()
	qualityReq := httptest.NewRequest(http.MethodGet, "/one-pager-quality", nil).WithContext(ctx)
	qualityHandlers.GetQualityList(qualityRec, qualityReq)
	require.Equal(t, http.StatusOK, qualityRec.Code)
	var qualityBody sharedAPI.PaginatedResponse
	require.NoError(t, json.Unmarshal(qualityRec.Body.Bytes(), &qualityBody))

	completeBySubjectID := map[string]bool{}
	for _, row := range completenessBody.Data {
		completeBySubjectID[row.SubjectID] = row.Complete
	}
	assert.True(t, completeBySubjectID["cap-1"], "completeness endpoint reports cap-1 complete")

	rawRows, err := json.Marshal(qualityBody.Data)
	require.NoError(t, err)
	var qualityRows []QualityRowDTO
	require.NoError(t, json.Unmarshal(rawRows, &qualityRows))
	qualitySignalBySubjectID := map[string]string{}
	for _, row := range qualityRows {
		qualitySignalBySubjectID[row.SubjectID] = row.Completeness
	}
	require.Contains(t, qualitySignalBySubjectID, "cap-1")
	assert.Equal(t, readmodels.SignalComplete, qualitySignalBySubjectID["cap-1"],
		"the quality list must agree with the completeness endpoint for the same subject")
}
