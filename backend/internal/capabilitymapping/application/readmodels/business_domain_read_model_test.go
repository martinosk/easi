//go:build integration

package readmodels

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"easi/backend/internal/infrastructure/database"
	sharedcontext "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		"localhost", "5432", "easi_app", "localdev", "easi", "disable")
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	return db, func() { db.Close() }
}

func tenantContext() context.Context {
	return sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
}

func setTenantContext(t *testing.T, db *sql.DB) {
	_, err := db.Exec("SET app.current_tenant = 'default'")
	require.NoError(t, err)
}

type businessDomainTestFixture struct {
	db        *sql.DB
	readModel *BusinessDomainReadModel
	ctx       context.Context
	t         *testing.T
}

func newBusinessDomainTestFixture(t *testing.T) *businessDomainTestFixture {
	db, cleanup := setupTestDB(t)
	t.Cleanup(cleanup)

	tenantDB := database.NewTenantAwareDB(db)

	return &businessDomainTestFixture{
		db:        db,
		readModel: NewBusinessDomainReadModel(tenantDB),
		ctx:       tenantContext(),
		t:         t,
	}
}

func (f *businessDomainTestFixture) uniqueID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func (f *businessDomainTestFixture) setTenantContext() {
	_, err := f.db.Exec("SET app.current_tenant = 'default'")
	require.NoError(f.t, err)
}

func (f *businessDomainTestFixture) insertDomain(id, name, description string, capabilityCount int) {
	f.setTenantContext()
	_, err := f.db.Exec(
		"INSERT INTO business_domains (id, tenant_id, name, description, capability_count, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		id, "default", name, description, capabilityCount, time.Now().UTC(),
	)
	require.NoError(f.t, err)
	f.t.Cleanup(func() { f.db.Exec("DELETE FROM business_domains WHERE id = $1", id) })
}

func (f *businessDomainTestFixture) queryName(id string) string {
	f.setTenantContext()
	var name string
	err := f.db.QueryRow("SELECT name FROM business_domains WHERE id = $1", id).Scan(&name)
	require.NoError(f.t, err)
	return name
}

func (f *businessDomainTestFixture) queryCapabilityCount(id string) int {
	f.setTenantContext()
	var count int
	err := f.db.QueryRow("SELECT capability_count FROM business_domains WHERE id = $1", id).Scan(&count)
	require.NoError(f.t, err)
	return count
}

func (f *businessDomainTestFixture) queryRowCount(id string) int {
	f.setTenantContext()
	var count int
	err := f.db.QueryRow("SELECT COUNT(*) FROM business_domains WHERE id = $1", id).Scan(&count)
	require.NoError(f.t, err)
	return count
}

func TestBusinessDomainReadModel_Insert(t *testing.T) {
	f := newBusinessDomainTestFixture(t)
	domainID := f.uniqueID("bd-test")
	now := time.Now().UTC()
	dto := BusinessDomainDTO{
		ID:          domainID,
		Name:        "Finance",
		Description: "Financial operations and planning",
		CreatedAt:   now,
	}

	err := f.readModel.Insert(f.ctx, dto)
	require.NoError(t, err)
	t.Cleanup(func() { f.db.Exec("DELETE FROM business_domains WHERE id = $1", domainID) })

	f.setTenantContext()
	var name, description string
	var createdAt time.Time
	err = f.db.QueryRow(
		"SELECT name, description, created_at FROM business_domains WHERE id = $1", domainID,
	).Scan(&name, &description, &createdAt)
	require.NoError(t, err)

	assert.Equal(t, "Finance", name)
	assert.Equal(t, "Financial operations and planning", description)
	assert.WithinDuration(t, now, createdAt, time.Second)
}

func TestBusinessDomainReadModel_Insert_IdempotentReplay(t *testing.T) {
	f := newBusinessDomainTestFixture(t)
	domainID := f.uniqueID("bd-replay")
	now := time.Now().UTC()
	dto := BusinessDomainDTO{
		ID:          domainID,
		Name:        "Finance-" + domainID,
		Description: "Financial operations and planning",
		CreatedAt:   now,
	}

	err := f.readModel.Insert(f.ctx, dto)
	require.NoError(t, err)
	t.Cleanup(func() { f.db.Exec("DELETE FROM business_domains WHERE id = $1", domainID) })

	err = f.readModel.Insert(f.ctx, dto)
	require.NoError(t, err)

	assert.Equal(t, 1, f.queryRowCount(domainID))
}

func TestBusinessDomainReadModel_Insert_ReplayConvergence(t *testing.T) {
	f := newBusinessDomainTestFixture(t)
	domainID := f.uniqueID("bd-converge")
	dto := BusinessDomainDTO{
		ID:          domainID,
		Name:        "Original Name",
		Description: "Original description",
		CreatedAt:   time.Now().UTC(),
	}

	require.NoError(t, f.readModel.Insert(f.ctx, dto))
	t.Cleanup(func() { f.db.Exec("DELETE FROM business_domains WHERE id = $1", domainID) })

	dto.Name = "Updated Name"
	require.NoError(t, f.readModel.Insert(f.ctx, dto))

	assert.Equal(t, "Updated Name", f.queryName(domainID))
}

func TestBusinessDomainReadModel_Update(t *testing.T) {
	f := newBusinessDomainTestFixture(t)
	domainID := f.uniqueID("bd-test")
	f.insertDomain(domainID, "Original Name", "Original Description", 0)

	err := f.readModel.Update(f.ctx, domainID, BusinessDomainUpdate{Name: "Updated Name", Description: "Updated Description"})
	require.NoError(t, err)

	f.setTenantContext()
	var name, description string
	err = f.db.QueryRow("SELECT name, description FROM business_domains WHERE id = $1", domainID).Scan(&name, &description)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", name)
	assert.Equal(t, "Updated Description", description)
}

func TestBusinessDomainReadModel_Delete(t *testing.T) {
	f := newBusinessDomainTestFixture(t)
	domainID := f.uniqueID("bd-test")
	f.insertDomain(domainID, "To Delete", "Will be removed", 0)

	err := f.readModel.Delete(f.ctx, domainID)
	require.NoError(t, err)

	assert.Equal(t, 0, f.queryRowCount(domainID))
}

func TestBusinessDomainReadModel_IncrementCapabilityCount(t *testing.T) {
	f := newBusinessDomainTestFixture(t)
	domainID := f.uniqueID("bd-test")
	f.insertDomain(domainID, "Test Domain", "Description", 0)

	require.NoError(t, f.readModel.IncrementCapabilityCount(f.ctx, domainID))
	assert.Equal(t, 1, f.queryCapabilityCount(domainID))

	require.NoError(t, f.readModel.IncrementCapabilityCount(f.ctx, domainID))
	assert.Equal(t, 2, f.queryCapabilityCount(domainID))
}

func TestBusinessDomainReadModel_DecrementCapabilityCount(t *testing.T) {
	f := newBusinessDomainTestFixture(t)
	domainID := f.uniqueID("bd-test")
	f.insertDomain(domainID, "Test Domain", "Description", 2)

	require.NoError(t, f.readModel.DecrementCapabilityCount(f.ctx, domainID))
	assert.Equal(t, 1, f.queryCapabilityCount(domainID))
}

func TestBusinessDomainReadModel_GetAll(t *testing.T) {
	f := newBusinessDomainTestFixture(t)
	domain1ID := f.uniqueID("bd-test-1")
	domain2ID := f.uniqueID("bd-test-2")
	f.insertDomain(domain1ID, "Finance", "Finance domain", 3)
	f.insertDomain(domain2ID, "Operations", "Operations domain", 5)

	domains, err := f.readModel.GetAll(f.ctx)
	require.NoError(t, err)

	var found1, found2 bool
	for _, d := range domains {
		if d.ID == domain1ID {
			found1 = true
			assert.Equal(t, "Finance", d.Name)
			assert.Equal(t, 3, d.CapabilityCount)
		}
		if d.ID == domain2ID {
			found2 = true
			assert.Equal(t, "Operations", d.Name)
			assert.Equal(t, 5, d.CapabilityCount)
		}
	}
	assert.True(t, found1, "Domain 1 not found in results")
	assert.True(t, found2, "Domain 2 not found in results")
}

func TestBusinessDomainReadModel_GetByID(t *testing.T) {
	f := newBusinessDomainTestFixture(t)
	domainID := f.uniqueID("bd-test")
	f.insertDomain(domainID, "Customer Experience", "CX domain", 7)

	domain, err := f.readModel.GetByID(f.ctx, domainID)
	require.NoError(t, err)
	require.NotNil(t, domain)

	assert.Equal(t, domainID, domain.ID)
	assert.Equal(t, "Customer Experience", domain.Name)
	assert.Equal(t, "CX domain", domain.Description)
	assert.Equal(t, 7, domain.CapabilityCount)
}

func TestBusinessDomainReadModel_GetByID_NotFound(t *testing.T) {
	f := newBusinessDomainTestFixture(t)

	domain, err := f.readModel.GetByID(f.ctx, "bd-nonexistent")
	require.NoError(t, err)
	assert.Nil(t, domain)
}

func TestBusinessDomainReadModel_GetByName(t *testing.T) {
	f := newBusinessDomainTestFixture(t)
	domainID := f.uniqueID("bd-test")
	f.insertDomain(domainID, "Technology", "Tech domain", 4)

	domain, err := f.readModel.GetByName(f.ctx, "Technology")
	require.NoError(t, err)
	require.NotNil(t, domain)

	assert.Equal(t, domainID, domain.ID)
	assert.Equal(t, "Technology", domain.Name)
}

func TestBusinessDomainReadModel_GetByName_NotFound(t *testing.T) {
	f := newBusinessDomainTestFixture(t)

	domain, err := f.readModel.GetByName(f.ctx, "NonExistentDomain")
	require.NoError(t, err)
	assert.Nil(t, domain)
}

func TestBusinessDomainReadModel_NameExists(t *testing.T) {
	f := newBusinessDomainTestFixture(t)
	domainID := f.uniqueID("bd-test")
	f.insertDomain(domainID, "Marketing", "Marketing domain", 0)

	exists, err := f.readModel.NameExists(f.ctx, "Marketing", "")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = f.readModel.NameExists(f.ctx, "NonExistent", "")
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = f.readModel.NameExists(f.ctx, "Marketing", domainID)
	require.NoError(t, err)
	assert.False(t, exists, "Should exclude self when checking name uniqueness")
}
