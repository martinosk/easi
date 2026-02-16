//go:build integration

package readmodels

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getArchEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func setupArchTestDB(t *testing.T) *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getArchEnv("INTEGRATION_TEST_DB_HOST", "localhost"),
		getArchEnv("INTEGRATION_TEST_DB_PORT", "5432"),
		getArchEnv("INTEGRATION_TEST_DB_USER", "easi_app"),
		getArchEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev"),
		getArchEnv("INTEGRATION_TEST_DB_NAME", "easi"),
		getArchEnv("INTEGRATION_TEST_DB_SSLMODE", "disable"))
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() { db.Close() })
	return db
}

type archTestFixture struct {
	db       *sql.DB
	tenantDB *database.TenantAwareDB
	ctx      context.Context
	t        *testing.T
}

func newArchTestFixture(t *testing.T) *archTestFixture {
	db := setupArchTestDB(t)
	return &archTestFixture{
		db:       db,
		tenantDB: database.NewTenantAwareDB(db),
		ctx:      sharedctx.WithTenant(context.Background(), sharedvo.DefaultTenantID()),
		t:        t,
	}
}

func (f *archTestFixture) uniqueID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func (f *archTestFixture) setTenantContext() {
	_, err := f.db.Exec("SET app.current_tenant = 'default'")
	require.NoError(f.t, err)
}

type tableRef struct {
	table    string
	idColumn string
}

func (f *archTestFixture) queryRowCount(ref tableRef, id string) int {
	f.setTenantContext()
	var count int
	err := f.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1", ref.table, ref.idColumn), id).Scan(&count)
	require.NoError(f.t, err)
	return count
}

func (f *archTestFixture) cleanup(ref tableRef, id string) {
	f.t.Cleanup(func() {
		f.setTenantContext()
		f.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s = $1", ref.table, ref.idColumn), id)
	})
}

var appComponentTable = tableRef{"architecturemodeling.application_components", "id"}
var appComponentExpertsTable = tableRef{"architecturemodeling.application_component_experts", "component_id"}

func TestApplicationComponentReadModel_Insert_IdempotentReplay(t *testing.T) {
	f := newArchTestFixture(t)
	rm := NewApplicationComponentReadModel(f.tenantDB)

	id := f.uniqueID("comp-replay")
	dto := ApplicationComponentDTO{
		ID:          id,
		Name:        "Replay Component",
		Description: "Created twice by replay",
		CreatedAt:   time.Now().UTC(),
	}

	require.NoError(t, rm.Insert(f.ctx, dto))
	f.cleanup(appComponentTable, id)

	require.NoError(t, rm.Insert(f.ctx, dto))
	assert.Equal(t, 1, f.queryRowCount(appComponentTable, id))
}

func TestApplicationComponentReadModel_Insert_ReplayConvergence(t *testing.T) {
	f := newArchTestFixture(t)
	rm := NewApplicationComponentReadModel(f.tenantDB)

	id := f.uniqueID("comp-converge")
	dto := ApplicationComponentDTO{
		ID:          id,
		Name:        "Original Name",
		Description: "Original",
		CreatedAt:   time.Now().UTC(),
	}

	require.NoError(t, rm.Insert(f.ctx, dto))
	f.cleanup(appComponentTable, id)

	dto.Name = "Updated Name"
	require.NoError(t, rm.Insert(f.ctx, dto))

	f.setTenantContext()
	var name string
	err := f.db.QueryRow("SELECT name FROM architecturemodeling.application_components WHERE id = $1", id).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", name)
}

func TestApplicationComponentReadModel_AddExpert_IdempotentReplay(t *testing.T) {
	f := newArchTestFixture(t)
	rm := NewApplicationComponentReadModel(f.tenantDB)

	compID := f.uniqueID("comp-expert")
	compDTO := ApplicationComponentDTO{
		ID:        compID,
		Name:      "Expert Component",
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, rm.Insert(f.ctx, compDTO))
	f.cleanup(appComponentExpertsTable, compID)
	f.cleanup(appComponentTable, compID)

	expert := ExpertInfo{
		ComponentID: compID,
		Name:        "Jane Doe",
		Role:        "Architect",
		Contact:     "jane@example.com",
		AddedAt:     time.Now().UTC(),
	}

	require.NoError(t, rm.AddExpert(f.ctx, expert))
	require.NoError(t, rm.AddExpert(f.ctx, expert))

	f.setTenantContext()
	var count int
	err := f.db.QueryRow(
		"SELECT COUNT(*) FROM architecturemodeling.application_component_experts WHERE component_id = $1 AND expert_name = $2 AND expert_role = $3",
		compID, "Jane Doe", "Architect",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestReadModel_Insert_IdempotentReplay(t *testing.T) {
	type readModelCase struct {
		ref    tableRef
		insert func(f *archTestFixture, id string) error
	}

	cases := map[string]readModelCase{
		"AcquiredEntity": {
			ref: tableRef{"architecturemodeling.acquired_entities", "id"},
			insert: func(f *archTestFixture, id string) error {
				return NewAcquiredEntityReadModel(f.tenantDB).Insert(f.ctx, AcquiredEntityDTO{
					ID: id, Name: "Acquired Corp", IntegrationStatus: "InProgress",
					Notes: "Integration ongoing", CreatedAt: time.Now().UTC(),
				})
			},
		},
		"ComponentRelation": {
			ref: tableRef{"architecturemodeling.component_relations", "id"},
			insert: func(f *archTestFixture, id string) error {
				return NewComponentRelationReadModel(f.tenantDB).Insert(f.ctx, ComponentRelationDTO{
					ID: id, SourceComponentID: "src-comp", TargetComponentID: "tgt-comp",
					RelationType: "DependsOn", Name: "Test Relation", CreatedAt: time.Now().UTC(),
				})
			},
		},
		"InternalTeam": {
			ref: tableRef{"architecturemodeling.internal_teams", "id"},
			insert: func(f *archTestFixture, id string) error {
				return NewInternalTeamReadModel(f.tenantDB).Insert(f.ctx, InternalTeamDTO{
					ID: id, Name: "Platform Team", Department: "Engineering",
					ContactPerson: "John", Notes: "Core team", CreatedAt: time.Now().UTC(),
				})
			},
		},
		"Vendor": {
			ref: tableRef{"architecturemodeling.vendors", "id"},
			insert: func(f *archTestFixture, id string) error {
				return NewVendorReadModel(f.tenantDB).Insert(f.ctx, VendorDTO{
					ID: id, Name: "Acme Corp", ImplementationPartner: "Partner Inc",
					Notes: "Primary vendor", CreatedAt: time.Now().UTC(),
				})
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := newArchTestFixture(t)
			id := f.uniqueID(name + "-replay")

			require.NoError(t, tc.insert(f, id))
			f.cleanup(tc.ref, id)

			require.NoError(t, tc.insert(f, id))
			assert.Equal(t, 1, f.queryRowCount(tc.ref, id))
		})
	}
}
