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

func setupLinkTestDB(t *testing.T) (*sql.DB, func()) {
	dbHost := getEnvOrDefault("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getEnvOrDefault("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getEnvOrDefault("INTEGRATION_TEST_DB_USER", "easi_app")
	dbPassword := getEnvOrDefault("INTEGRATION_TEST_DB_PASSWORD", "localdev")
	dbName := getEnvOrDefault("INTEGRATION_TEST_DB_NAME", "easi")
	dbSSLMode := getEnvOrDefault("INTEGRATION_TEST_DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func getEnvOrDefault(key, defaultValue string) string {
	return defaultValue
}

func linkTenantContext() context.Context {
	return sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
}

func setLinkTenantContext(t *testing.T, db *sql.DB) {
	_, err := db.Exec("SET app.current_tenant = 'default'")
	require.NoError(t, err)
}

func TestGetLinkStatus_DirectlyLinked(t *testing.T) {
	db, cleanup := setupLinkTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewEnterpriseCapabilityLinkReadModel(tenantDB)

	enterpriseCapID := fmt.Sprintf("ec-test-%d", time.Now().UnixNano())
	domainCapID := fmt.Sprintf("dc-test-%d", time.Now().UnixNano())
	linkID := fmt.Sprintf("link-test-%d", time.Now().UnixNano())

	setLinkTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, $2, $3, $4)",
		enterpriseCapID, "default", "Customer Management", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, level, created_at) VALUES ($1, $2, $3, $4, $5)",
		domainCapID, "default", "Order Processing", "L2", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at) VALUES ($1, $2, $3, $4, $5, $6)",
		linkID, "default", enterpriseCapID, domainCapID, "user-123", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := linkTenantContext()
	status, err := readModel.GetLinkStatus(ctx, domainCapID)
	require.NoError(t, err)
	require.NotNil(t, status)

	assert.Equal(t, domainCapID, status.CapabilityID)
	assert.Equal(t, LinkStatusLinked, status.Status)
	require.NotNil(t, status.LinkedTo)
	assert.Equal(t, enterpriseCapID, status.LinkedTo.ID)
	assert.Equal(t, "Customer Management", status.LinkedTo.Name)
	assert.Nil(t, status.BlockingCapability)
	assert.Nil(t, status.BlockingEnterpriseCapID)

	_, err = db.Exec("DELETE FROM enterprise_capability_links WHERE id = $1", linkID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM capabilities WHERE id = $1", domainCapID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM enterprise_capabilities WHERE id = $1", enterpriseCapID)
	require.NoError(t, err)
}

func TestGetLinkStatus_Available_NoLinksNoConflicts(t *testing.T) {
	db, cleanup := setupLinkTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewEnterpriseCapabilityLinkReadModel(tenantDB)

	domainCapID := fmt.Sprintf("dc-test-%d", time.Now().UnixNano())

	setLinkTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, level, created_at) VALUES ($1, $2, $3, $4, $5)",
		domainCapID, "default", "Standalone Capability", "L2", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := linkTenantContext()
	status, err := readModel.GetLinkStatus(ctx, domainCapID)
	require.NoError(t, err)
	require.NotNil(t, status)

	assert.Equal(t, domainCapID, status.CapabilityID)
	assert.Equal(t, LinkStatusAvailable, status.Status)
	assert.Nil(t, status.LinkedTo)
	assert.Nil(t, status.BlockingCapability)
	assert.Nil(t, status.BlockingEnterpriseCapID)

	_, err = db.Exec("DELETE FROM capabilities WHERE id = $1", domainCapID)
	require.NoError(t, err)
}

func TestGetLinkStatus_BlockedByParent(t *testing.T) {
	db, cleanup := setupLinkTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewEnterpriseCapabilityLinkReadModel(tenantDB)

	enterpriseCapID := fmt.Sprintf("ec-test-%d", time.Now().UnixNano())
	parentCapID := fmt.Sprintf("parent-%d", time.Now().UnixNano())
	childCapID := fmt.Sprintf("child-%d", time.Now().UnixNano())
	linkID := fmt.Sprintf("link-test-%d", time.Now().UnixNano())

	setLinkTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, $2, $3, $4)",
		enterpriseCapID, "default", "Sales Operations", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		parentCapID, "default", "Parent Capability", nil, "L1", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		childCapID, "default", "Child Capability", parentCapID, "L2", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at) VALUES ($1, $2, $3, $4, $5, $6)",
		linkID, "default", enterpriseCapID, parentCapID, "user-123", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := linkTenantContext()
	status, err := readModel.GetLinkStatus(ctx, childCapID)
	require.NoError(t, err)
	require.NotNil(t, status)

	assert.Equal(t, childCapID, status.CapabilityID)
	assert.Equal(t, LinkStatusBlockedByParent, status.Status)
	assert.Nil(t, status.LinkedTo)
	require.NotNil(t, status.BlockingCapability)
	assert.Equal(t, parentCapID, status.BlockingCapability.ID)
	assert.Equal(t, "Parent Capability", status.BlockingCapability.Name)
	require.NotNil(t, status.BlockingEnterpriseCapID)
	assert.Equal(t, enterpriseCapID, *status.BlockingEnterpriseCapID)

	_, err = db.Exec("DELETE FROM enterprise_capability_links WHERE id = $1", linkID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM capabilities WHERE id IN ($1, $2)", childCapID, parentCapID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM enterprise_capabilities WHERE id = $1", enterpriseCapID)
	require.NoError(t, err)
}

func TestGetLinkStatus_BlockedByChild(t *testing.T) {
	db, cleanup := setupLinkTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewEnterpriseCapabilityLinkReadModel(tenantDB)

	enterpriseCapID := fmt.Sprintf("ec-test-%d", time.Now().UnixNano())
	parentCapID := fmt.Sprintf("parent-%d", time.Now().UnixNano())
	childCapID := fmt.Sprintf("child-%d", time.Now().UnixNano())
	linkID := fmt.Sprintf("link-test-%d", time.Now().UnixNano())

	setLinkTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, $2, $3, $4)",
		enterpriseCapID, "default", "Marketing Automation", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		parentCapID, "default", "Parent Capability", nil, "L1", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		childCapID, "default", "Child Capability", parentCapID, "L2", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at) VALUES ($1, $2, $3, $4, $5, $6)",
		linkID, "default", enterpriseCapID, childCapID, "user-123", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := linkTenantContext()
	status, err := readModel.GetLinkStatus(ctx, parentCapID)
	require.NoError(t, err)
	require.NotNil(t, status)

	assert.Equal(t, parentCapID, status.CapabilityID)
	assert.Equal(t, LinkStatusBlockedByChild, status.Status)
	assert.Nil(t, status.LinkedTo)
	require.NotNil(t, status.BlockingCapability)
	assert.Equal(t, childCapID, status.BlockingCapability.ID)
	assert.Equal(t, "Child Capability", status.BlockingCapability.Name)
	require.NotNil(t, status.BlockingEnterpriseCapID)
	assert.Equal(t, enterpriseCapID, *status.BlockingEnterpriseCapID)

	_, err = db.Exec("DELETE FROM enterprise_capability_links WHERE id = $1", linkID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM capabilities WHERE id IN ($1, $2)", childCapID, parentCapID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM enterprise_capabilities WHERE id = $1", enterpriseCapID)
	require.NoError(t, err)
}

func TestGetLinkStatus_MultiLevelHierarchy_BlockedByGrandparent(t *testing.T) {
	db, cleanup := setupLinkTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewEnterpriseCapabilityLinkReadModel(tenantDB)

	enterpriseCapID := fmt.Sprintf("ec-test-%d", time.Now().UnixNano())
	grandparentCapID := fmt.Sprintf("gp-%d", time.Now().UnixNano())
	parentCapID := fmt.Sprintf("parent-%d", time.Now().UnixNano())
	childCapID := fmt.Sprintf("child-%d", time.Now().UnixNano())
	linkID := fmt.Sprintf("link-test-%d", time.Now().UnixNano())

	setLinkTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, $2, $3, $4)",
		enterpriseCapID, "default", "Product Innovation", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		grandparentCapID, "default", "Grandparent Capability", nil, "L1", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		parentCapID, "default", "Parent Capability", grandparentCapID, "L2", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		childCapID, "default", "Child Capability", parentCapID, "L3", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at) VALUES ($1, $2, $3, $4, $5, $6)",
		linkID, "default", enterpriseCapID, grandparentCapID, "user-123", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := linkTenantContext()
	status, err := readModel.GetLinkStatus(ctx, childCapID)
	require.NoError(t, err)
	require.NotNil(t, status)

	assert.Equal(t, childCapID, status.CapabilityID)
	assert.Equal(t, LinkStatusBlockedByParent, status.Status)
	assert.Nil(t, status.LinkedTo)
	require.NotNil(t, status.BlockingCapability)
	assert.Equal(t, grandparentCapID, status.BlockingCapability.ID)
	assert.Equal(t, "Grandparent Capability", status.BlockingCapability.Name)
	require.NotNil(t, status.BlockingEnterpriseCapID)
	assert.Equal(t, enterpriseCapID, *status.BlockingEnterpriseCapID)

	_, err = db.Exec("DELETE FROM enterprise_capability_links WHERE id = $1", linkID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM capabilities WHERE id IN ($1, $2, $3)", childCapID, parentCapID, grandparentCapID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM enterprise_capabilities WHERE id = $1", enterpriseCapID)
	require.NoError(t, err)
}

func TestGetLinkStatus_MultiLevelHierarchy_BlockedByGrandchild(t *testing.T) {
	db, cleanup := setupLinkTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(db)
	readModel := NewEnterpriseCapabilityLinkReadModel(tenantDB)

	enterpriseCapID := fmt.Sprintf("ec-test-%d", time.Now().UnixNano())
	grandparentCapID := fmt.Sprintf("gp-%d", time.Now().UnixNano())
	parentCapID := fmt.Sprintf("parent-%d", time.Now().UnixNano())
	childCapID := fmt.Sprintf("child-%d", time.Now().UnixNano())
	linkID := fmt.Sprintf("link-test-%d", time.Now().UnixNano())

	setLinkTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, $2, $3, $4)",
		enterpriseCapID, "default", "Data Analytics", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		grandparentCapID, "default", "Grandparent Capability", nil, "L1", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		parentCapID, "default", "Parent Capability", grandparentCapID, "L2", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		childCapID, "default", "Child Capability", parentCapID, "L3", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at) VALUES ($1, $2, $3, $4, $5, $6)",
		linkID, "default", enterpriseCapID, childCapID, "user-123", time.Now().UTC(),
	)
	require.NoError(t, err)

	ctx := linkTenantContext()
	status, err := readModel.GetLinkStatus(ctx, grandparentCapID)
	require.NoError(t, err)
	require.NotNil(t, status)

	assert.Equal(t, grandparentCapID, status.CapabilityID)
	assert.Equal(t, LinkStatusBlockedByChild, status.Status)
	assert.Nil(t, status.LinkedTo)
	require.NotNil(t, status.BlockingCapability)
	assert.Equal(t, childCapID, status.BlockingCapability.ID)
	assert.Equal(t, "Child Capability", status.BlockingCapability.Name)
	require.NotNil(t, status.BlockingEnterpriseCapID)
	assert.Equal(t, enterpriseCapID, *status.BlockingEnterpriseCapID)

	_, err = db.Exec("DELETE FROM enterprise_capability_links WHERE id = $1", linkID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM capabilities WHERE id IN ($1, $2, $3)", childCapID, parentCapID, grandparentCapID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM enterprise_capabilities WHERE id = $1", enterpriseCapID)
	require.NoError(t, err)
}
