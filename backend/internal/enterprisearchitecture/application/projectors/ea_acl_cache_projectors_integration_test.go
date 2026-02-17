//go:build integration

package projectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const aclTestTenant = "acl-projector-test"

type aclCacheTestFixture struct {
	t                    *testing.T
	db                   *sql.DB
	ctx                  context.Context
	realizationRM        *readmodels.EARealizationCacheReadModel
	importanceRM         *readmodels.EAImportanceCacheReadModel
	fitScoreRM           *readmodels.EAFitScoreCacheReadModel
	metadataRM           *readmodels.DomainCapabilityMetadataReadModel
	realizationProjector *EARealizationCacheProjector
	importanceProjector  *EAImportanceCacheProjector
	fitScoreProjector    *EAFitScoreCacheProjector
	metadataProjector    *DomainCapabilityMetadataProjector
}

func setupACLCacheTest(t *testing.T) *aclCacheTestFixture {
	t.Helper()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		"localhost", "5432", "easi_app", "localdev", "easi", "disable")
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	_, err = db.Exec("SELECT set_config('app.current_tenant', $1, false)", aclTestTenant)
	require.NoError(t, err)

	tenantDB := database.NewTenantAwareDB(db)
	ctx := sharedctx.WithTenant(context.Background(), valueobjects.MustNewTenantID(aclTestTenant))

	realizationRM := readmodels.NewEARealizationCacheReadModel(tenantDB)
	importanceRM := readmodels.NewEAImportanceCacheReadModel(tenantDB)
	fitScoreRM := readmodels.NewEAFitScoreCacheReadModel(tenantDB)
	metadataRM := readmodels.NewDomainCapabilityMetadataReadModel(tenantDB)
	capabilityRM := readmodels.NewEnterpriseCapabilityReadModel(tenantDB)
	linkRM := readmodels.NewEnterpriseCapabilityLinkReadModel(tenantDB)

	t.Cleanup(func() {
		db.Exec("DELETE FROM enterprisearchitecture.ea_realization_cache WHERE tenant_id = $1", aclTestTenant)
		db.Exec("DELETE FROM enterprisearchitecture.ea_importance_cache WHERE tenant_id = $1", aclTestTenant)
		db.Exec("DELETE FROM enterprisearchitecture.ea_fit_score_cache WHERE tenant_id = $1", aclTestTenant)
		db.Exec("DELETE FROM enterprisearchitecture.domain_capability_metadata WHERE tenant_id = $1", aclTestTenant)
		db.Close()
	})

	return &aclCacheTestFixture{
		t:                    t,
		db:                   db,
		ctx:                  ctx,
		realizationRM:        realizationRM,
		importanceRM:         importanceRM,
		fitScoreRM:           fitScoreRM,
		metadataRM:           metadataRM,
		realizationProjector: NewEARealizationCacheProjector(realizationRM),
		importanceProjector:  NewEAImportanceCacheProjector(importanceRM),
		fitScoreProjector:    NewEAFitScoreCacheProjector(fitScoreRM),
		metadataProjector:    NewDomainCapabilityMetadataProjector(metadataRM, capabilityRM, linkRM),
	}
}

func (f *aclCacheTestFixture) seedRealization(capabilityID, componentID, componentName string) string {
	f.t.Helper()
	realizationID := uuid.New().String()
	err := f.realizationRM.Upsert(f.ctx, readmodels.RealizationEntry{
		RealizationID: realizationID,
		CapabilityID:  capabilityID,
		ComponentID:   componentID,
		ComponentName: componentName,
		Origin:        "Direct",
	})
	require.NoError(f.t, err)
	return realizationID
}

func (f *aclCacheTestFixture) countRealizationsByColumn(column, value string) int {
	f.t.Helper()
	var count int
	err := f.db.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM enterprisearchitecture.ea_realization_cache WHERE tenant_id = $1 AND %s = $2", column),
		aclTestTenant, value,
	).Scan(&count)
	require.NoError(f.t, err)
	return count
}

func TestRealizationCacheProjector_SystemLinked_InsertsCache(t *testing.T) {
	f := setupACLCacheTest(t)

	realizationID := uuid.New().String()
	capabilityID := uuid.New().String()
	componentID := uuid.New().String()

	eventData, err := json.Marshal(map[string]interface{}{
		"id":               realizationID,
		"capabilityId":     capabilityID,
		"componentId":      componentID,
		"componentName":    "Test Component",
		"realizationLevel": "Direct",
	})
	require.NoError(t, err)

	err = f.realizationProjector.ProjectEvent(f.ctx, "SystemLinkedToCapability", eventData)
	require.NoError(t, err)

	assert.Equal(t, 1, f.countRealizationsByColumn("realization_id", realizationID))
}

func TestRealizationCacheProjector_SystemDeleted_RemovesCache(t *testing.T) {
	f := setupACLCacheTest(t)

	realizationID := f.seedRealization(uuid.New().String(), uuid.New().String(), "Test")

	eventData, _ := json.Marshal(map[string]interface{}{"id": realizationID})
	err := f.realizationProjector.ProjectEvent(f.ctx, "SystemRealizationDeleted", eventData)
	require.NoError(t, err)

	assert.Equal(t, 0, f.countRealizationsByColumn("realization_id", realizationID))
}

func TestRealizationCacheProjector_CapabilityDeleted_RemovesAllForCapability(t *testing.T) {
	f := setupACLCacheTest(t)

	capabilityID := uuid.New().String()
	for i := 0; i < 3; i++ {
		f.seedRealization(capabilityID, uuid.New().String(), "Comp")
	}

	eventData, _ := json.Marshal(map[string]interface{}{"id": capabilityID})
	err := f.realizationProjector.ProjectEvent(f.ctx, "CapabilityDeleted", eventData)
	require.NoError(t, err)

	assert.Equal(t, 0, f.countRealizationsByColumn("capability_id", capabilityID))
}

func TestRealizationCacheProjector_ComponentUpdated_UpdatesCachedName(t *testing.T) {
	f := setupACLCacheTest(t)

	componentID := uuid.New().String()
	f.seedRealization(uuid.New().String(), componentID, "Old Name")

	eventData, _ := json.Marshal(map[string]interface{}{"id": componentID, "name": "New Name"})
	err := f.realizationProjector.ProjectEvent(f.ctx, "ApplicationComponentUpdated", eventData)
	require.NoError(t, err)

	var name string
	err = f.db.QueryRow(
		"SELECT component_name FROM enterprisearchitecture.ea_realization_cache WHERE tenant_id = $1 AND component_id = $2",
		aclTestTenant, componentID,
	).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "New Name", name)
}

func TestImportanceCacheProjector_Recalculated_InsertsCache(t *testing.T) {
	f := setupACLCacheTest(t)

	capabilityID := uuid.New().String()
	domainID := uuid.New().String()
	pillarID := uuid.New().String()

	eventData, err := json.Marshal(map[string]interface{}{
		"capabilityId":     capabilityID,
		"businessDomainId": domainID,
		"pillarId":         pillarID,
		"importance":       85,
	})
	require.NoError(t, err)

	err = f.importanceProjector.ProjectEvent(f.ctx, "EffectiveImportanceRecalculated", eventData)
	require.NoError(t, err)

	var importance int
	err = f.db.QueryRow(
		"SELECT effective_importance FROM enterprisearchitecture.ea_importance_cache WHERE tenant_id = $1 AND capability_id = $2 AND pillar_id = $3",
		aclTestTenant, capabilityID, pillarID,
	).Scan(&importance)
	require.NoError(t, err)
	assert.Equal(t, 85, importance)
}

func TestImportanceCacheProjector_Recalculated_UpdatesExistingEntry(t *testing.T) {
	f := setupACLCacheTest(t)

	capabilityID := uuid.New().String()
	domainID := uuid.New().String()
	pillarID := uuid.New().String()

	err := f.importanceRM.Upsert(f.ctx, readmodels.ImportanceEntry{
		CapabilityID:        capabilityID,
		BusinessDomainID:    domainID,
		PillarID:            pillarID,
		EffectiveImportance: 50,
	})
	require.NoError(t, err)

	eventData, _ := json.Marshal(map[string]interface{}{
		"capabilityId":     capabilityID,
		"businessDomainId": domainID,
		"pillarId":         pillarID,
		"importance":       90,
	})
	err = f.importanceProjector.ProjectEvent(f.ctx, "EffectiveImportanceRecalculated", eventData)
	require.NoError(t, err)

	var importance int
	err = f.db.QueryRow(
		"SELECT effective_importance FROM enterprisearchitecture.ea_importance_cache WHERE tenant_id = $1 AND capability_id = $2 AND pillar_id = $3",
		aclTestTenant, capabilityID, pillarID,
	).Scan(&importance)
	require.NoError(t, err)
	assert.Equal(t, 90, importance)
}

func TestFitScoreCacheProjector_ScoreSet_InsertsCache(t *testing.T) {
	f := setupACLCacheTest(t)

	componentID := uuid.New().String()
	pillarID := uuid.New().String()

	eventData, err := json.Marshal(map[string]interface{}{
		"componentId": componentID,
		"pillarId":    pillarID,
		"score":       75,
		"rationale":   "Good technical fit",
	})
	require.NoError(t, err)

	err = f.fitScoreProjector.ProjectEvent(f.ctx, "ApplicationFitScoreSet", eventData)
	require.NoError(t, err)

	var score int
	var rationale string
	err = f.db.QueryRow(
		"SELECT score, rationale FROM enterprisearchitecture.ea_fit_score_cache WHERE tenant_id = $1 AND component_id = $2 AND pillar_id = $3",
		aclTestTenant, componentID, pillarID,
	).Scan(&score, &rationale)
	require.NoError(t, err)
	assert.Equal(t, 75, score)
	assert.Equal(t, "Good technical fit", rationale)
}

func TestFitScoreCacheProjector_ScoreRemoved_DeletesCache(t *testing.T) {
	f := setupACLCacheTest(t)

	componentID := uuid.New().String()
	pillarID := uuid.New().String()

	err := f.fitScoreRM.Upsert(f.ctx, readmodels.FitScoreEntry{
		ComponentID: componentID,
		PillarID:    pillarID,
		Score:       75,
		Rationale:   "test",
	})
	require.NoError(t, err)

	eventData, _ := json.Marshal(map[string]interface{}{
		"componentId": componentID,
		"pillarId":    pillarID,
	})
	err = f.fitScoreProjector.ProjectEvent(f.ctx, "ApplicationFitScoreRemoved", eventData)
	require.NoError(t, err)

	var count int
	err = f.db.QueryRow(
		"SELECT COUNT(*) FROM enterprisearchitecture.ea_fit_score_cache WHERE tenant_id = $1 AND component_id = $2 AND pillar_id = $3",
		aclTestTenant, componentID, pillarID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestMetadataProjector_CapabilityMetadataUpdated_UpdatesMaturity(t *testing.T) {
	f := setupACLCacheTest(t)

	capabilityID := uuid.New().String()
	err := f.metadataRM.Insert(f.ctx, readmodels.DomainCapabilityMetadataDTO{
		CapabilityID:    capabilityID,
		CapabilityName:  "Test Capability",
		CapabilityLevel: "L1",
		L1CapabilityID:  capabilityID,
	})
	require.NoError(t, err)

	eventData, err := json.Marshal(map[string]interface{}{
		"id":            capabilityID,
		"maturityValue": 65,
	})
	require.NoError(t, err)

	err = f.metadataProjector.ProjectEvent(f.ctx, "CapabilityMetadataUpdated", eventData)
	require.NoError(t, err)

	var maturityValue int
	err = f.db.QueryRow(
		"SELECT maturity_value FROM enterprisearchitecture.domain_capability_metadata WHERE tenant_id = $1 AND capability_id = $2",
		aclTestTenant, capabilityID,
	).Scan(&maturityValue)
	require.NoError(t, err)
	assert.Equal(t, 65, maturityValue)
}
