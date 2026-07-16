//go:build integration
// +build integration

package api

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
	eaReadModels "easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/domain/catalog"
	"easi/backend/internal/onepagers/domain/valueobjects"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type builtInFieldIntegrationContext struct {
	sqlDB    *sql.DB
	tenantDB *database.TenantAwareDB
	ctx      context.Context
	tenantID string
}

func setupBuiltInFieldIntegration(t *testing.T) *builtInFieldIntegrationContext {
	t.Helper()

	dbHost := getEnv("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getEnv("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getEnv("INTEGRATION_TEST_DB_USER", "easi_app")
	dbPassword := getEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev")
	dbName := getEnv("INTEGRATION_TEST_DB_NAME", "easi")
	dbSSLMode := getEnv("INTEGRATION_TEST_DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	sqlDB, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, sqlDB.Ping())
	t.Cleanup(func() { _ = sqlDB.Close() })

	tenantIDValue := fmt.Sprintf("test-%s", uuid.New().String())
	tenantID, err := sharedvo.NewTenantID(tenantIDValue)
	require.NoError(t, err)

	return &builtInFieldIntegrationContext{
		sqlDB:    sqlDB,
		tenantDB: database.NewTenantAwareDB(sqlDB),
		ctx:      sharedctx.WithTenant(context.Background(), tenantID),
		tenantID: tenantIDValue,
	}
}

func (ic *builtInFieldIntegrationContext) cleanupRow(t *testing.T, table, id string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = ic.sqlDB.Exec(fmt.Sprintf("DELETE FROM %s WHERE tenant_id = $1 AND id = $2", table), ic.tenantID, id)
	})
}

type catalogContractCase struct {
	subjectType string
	table       string
	name        string
	insert      func(ctx context.Context, id string) error
}

func (ic *builtInFieldIntegrationContext) catalogContractCases() []catalogContractCase {
	return []catalogContractCase{
		{
			subjectType: "capability",
			table:       "capabilitymapping.capabilities",
			name:        "Order Management",
			insert: func(ctx context.Context, id string) error {
				return capReadModels.NewCapabilityReadModel(ic.tenantDB).Insert(ctx, capReadModels.CapabilityDTO{
					ID: id, Name: "Order Management", Description: "Handles customer orders", Level: "L1", CreatedAt: time.Now(),
				})
			},
		},
		{
			subjectType: "enterprise-capability",
			table:       "enterprisearchitecture.enterprise_capabilities",
			name:        "Customer Experience",
			insert: func(ctx context.Context, id string) error {
				return eaReadModels.NewEnterpriseCapabilityReadModel(ic.tenantDB).Insert(ctx, eaReadModels.EnterpriseCapabilityDTO{
					ID: id, Name: "Customer Experience", Active: true, CreatedAt: time.Now(),
				})
			},
		},
		{
			subjectType: "application",
			table:       "architecturemodeling.application_components",
			name:        "Billing Service",
			insert: func(ctx context.Context, id string) error {
				return archReadModels.NewApplicationComponentReadModel(ic.tenantDB).Insert(ctx, archReadModels.ApplicationComponentDTO{
					ID: id, Name: "Billing Service", CreatedAt: time.Now(),
				})
			},
		},
		{
			subjectType: "acquired-entity",
			table:       "architecturemodeling.acquired_entities",
			name:        "Acme Corp",
			insert: func(ctx context.Context, id string) error {
				return archReadModels.NewAcquiredEntityReadModel(ic.tenantDB).Insert(ctx, archReadModels.AcquiredEntityDTO{
					ID: id, Name: "Acme Corp", CreatedAt: time.Now(),
				})
			},
		},
		{
			subjectType: "vendor",
			table:       "architecturemodeling.vendors",
			name:        "Contoso",
			insert: func(ctx context.Context, id string) error {
				return archReadModels.NewVendorReadModel(ic.tenantDB).Insert(ctx, archReadModels.VendorDTO{
					ID: id, Name: "Contoso", CreatedAt: time.Now(),
				})
			},
		},
		{
			subjectType: "internal-team",
			table:       "architecturemodeling.internal_teams",
			name:        "Platform Team",
			insert: func(ctx context.Context, id string) error {
				return archReadModels.NewInternalTeamReadModel(ic.tenantDB).Insert(ctx, archReadModels.InternalTeamDTO{
					ID: id, Name: "Platform Team", CreatedAt: time.Now(),
				})
			},
		},
	}
}

func assertCatalogContract(t *testing.T, ic *builtInFieldIntegrationContext, sources map[string]ports.BuiltInFieldSource, c catalogContractCase) {
	t.Helper()

	id := uuid.New().String()
	require.NoError(t, c.insert(ic.ctx, id))
	ic.cleanupRow(t, c.table, id)

	snapshot, err := sources[c.subjectType].FetchSubject(ic.ctx, id, nil)

	require.NoError(t, err)
	require.NotNil(t, snapshot)
	assert.Equal(t, c.name, snapshot.Name)
	assertCatalogFieldsPresent(t, c.subjectType, snapshot)
}

func TestOnePagerBuiltInFieldSources_CatalogContract_Integration(t *testing.T) {
	ic := setupBuiltInFieldIntegration(t)
	sources := newOnePagerBuiltInFieldSources(ic.tenantDB)

	for _, c := range ic.catalogContractCases() {
		t.Run(c.subjectType, func(t *testing.T) {
			assertCatalogContract(t, ic, sources, c)
		})
	}
}

func relationEntryIDs(t *testing.T, subjectType string) []string {
	t.Helper()
	st, err := valueobjects.NewSubjectType(subjectType)
	require.NoError(t, err)
	ids := []string{}
	for _, entry := range catalog.EntriesFor(st) {
		if entry.Relation {
			ids = append(ids, entry.ID)
		}
	}
	return ids
}

func assertRelationCatalogContract(t *testing.T, ic *builtInFieldIntegrationContext, sources map[string]ports.BuiltInFieldSource, c catalogContractCase) {
	id := uuid.New().String()
	require.NoError(t, c.insert(ic.ctx, id))
	ic.cleanupRow(t, c.table, id)

	for _, entryID := range relationEntryIDs(t, c.subjectType) {
		t.Run(entryID, func(t *testing.T) {
			snapshot, err := sources[c.subjectType].FetchSubject(ic.ctx, id, []string{entryID})

			require.NoError(t, err, "relation %q must resolve against its supplier read model", entryID)
			require.NotNil(t, snapshot)
			value, present := snapshot.Fields[entryID]
			require.Truef(t, present, "relation %q missing from snapshot fields", entryID)
			_, isReferenceList := value.(ports.ReferenceListValue)
			assert.Truef(t, isReferenceList, "relation %q must resolve to a ReferenceListValue", entryID)
		})
	}
}

func TestOnePagerBuiltInFieldSources_RelationCatalogContract_Integration(t *testing.T) {
	ic := setupBuiltInFieldIntegration(t)
	sources := newOnePagerBuiltInFieldSources(ic.tenantDB)

	for _, c := range ic.catalogContractCases() {
		t.Run(c.subjectType, func(t *testing.T) {
			assertRelationCatalogContract(t, ic, sources, c)
		})
	}
}

func TestOnePagerBuiltInFieldSources_FilledBuiltInFields_Integration(t *testing.T) {
	ic := setupBuiltInFieldIntegration(t)
	sources := newOnePagerBuiltInFieldSources(ic.tenantDB)
	apps := archReadModels.NewApplicationComponentReadModel(ic.tenantDB)

	withExperts := uuid.New().String()
	require.NoError(t, apps.Insert(ic.ctx, archReadModels.ApplicationComponentDTO{ID: withExperts, Name: "Billing", Description: "Handles invoicing", CreatedAt: time.Now()}))
	ic.cleanupRow(t, "architecturemodeling.application_components", withExperts)
	require.NoError(t, apps.AddExpert(ic.ctx, archReadModels.ExpertInfo{ComponentID: withExperts, Name: "Alice", Role: "Owner", Contact: "alice@example.com", AddedAt: time.Now()}))

	noExperts := uuid.New().String()
	require.NoError(t, apps.Insert(ic.ctx, archReadModels.ApplicationComponentDTO{ID: noExperts, Name: "Payments", CreatedAt: time.Now()}))
	ic.cleanupRow(t, "architecturemodeling.application_components", noExperts)

	filled, err := sources["application"].FilledBuiltInFields(ic.ctx, []string{withExperts, noExperts}, []string{"description", "experts"})

	require.NoError(t, err)
	assert.True(t, filled[withExperts]["experts"], "app with an expert has experts filled")
	assert.True(t, filled[withExperts]["description"])
	assert.False(t, filled[noExperts]["experts"], "app with no expert has experts unfilled")
	assert.False(t, filled[noExperts]["description"])
}

func TestOnePagerBuiltInFieldSources_CountSubjectsWithBuiltInValue_Integration(t *testing.T) {
	ic := setupBuiltInFieldIntegration(t)
	sources := newOnePagerBuiltInFieldSources(ic.tenantDB)
	apps := archReadModels.NewApplicationComponentReadModel(ic.tenantDB)

	withExperts := uuid.New().String()
	require.NoError(t, apps.Insert(ic.ctx, archReadModels.ApplicationComponentDTO{ID: withExperts, Name: "Billing", CreatedAt: time.Now()}))
	ic.cleanupRow(t, "architecturemodeling.application_components", withExperts)
	require.NoError(t, apps.AddExpert(ic.ctx, archReadModels.ExpertInfo{ComponentID: withExperts, Name: "Alice", Role: "Owner", Contact: "alice@example.com", AddedAt: time.Now()}))

	noExperts := uuid.New().String()
	require.NoError(t, apps.Insert(ic.ctx, archReadModels.ApplicationComponentDTO{ID: noExperts, Name: "Payments", CreatedAt: time.Now()}))
	ic.cleanupRow(t, "architecturemodeling.application_components", noExperts)

	count, err := sources["application"].CountSubjectsWithBuiltInValue(ic.ctx, "experts")

	require.NoError(t, err)
	assert.Equal(t, 1, count, "one application has a value for experts")
}
