//go:build integration

package projectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"easi/backend/internal/architecturedirection/application/readmodels"
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
	t                   *testing.T
	db                  *sql.DB
	ctx                 context.Context
	importanceRM        *readmodels.EAImportanceCacheReadModel
	fitScoreRM          *readmodels.EAFitScoreCacheReadModel
	importanceProjector *EAImportanceCacheProjector
	fitScoreProjector   *EAFitScoreCacheProjector
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

	importanceRM := readmodels.NewEAImportanceCacheReadModel(tenantDB)
	fitScoreRM := readmodels.NewEAFitScoreCacheReadModel(tenantDB)

	t.Cleanup(func() {
		db.Exec("DELETE FROM architecturedirection.ea_importance_cache WHERE tenant_id = $1", aclTestTenant)
		db.Exec("DELETE FROM architecturedirection.ea_fit_score_cache WHERE tenant_id = $1", aclTestTenant)
		db.Close()
	})

	return &aclCacheTestFixture{
		t:                   t,
		db:                  db,
		ctx:                 ctx,
		importanceRM:        importanceRM,
		fitScoreRM:          fitScoreRM,
		importanceProjector: NewEAImportanceCacheProjector(importanceRM),
		fitScoreProjector:   NewEAFitScoreCacheProjector(fitScoreRM),
	}
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
		"SELECT effective_importance FROM architecturedirection.ea_importance_cache WHERE tenant_id = $1 AND capability_id = $2 AND pillar_id = $3",
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
		"SELECT effective_importance FROM architecturedirection.ea_importance_cache WHERE tenant_id = $1 AND capability_id = $2 AND pillar_id = $3",
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
		"SELECT score, rationale FROM architecturedirection.ea_fit_score_cache WHERE tenant_id = $1 AND component_id = $2 AND pillar_id = $3",
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
		"SELECT COUNT(*) FROM architecturedirection.ea_fit_score_cache WHERE tenant_id = $1 AND component_id = $2 AND pillar_id = $3",
		aclTestTenant, componentID, pillarID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
