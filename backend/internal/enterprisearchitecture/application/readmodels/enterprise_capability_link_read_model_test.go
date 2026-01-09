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

	uniqueSuffix := time.Now().UnixNano()
	enterpriseCapID := fmt.Sprintf("ec-test-%d", uniqueSuffix)
	enterpriseCapName := fmt.Sprintf("Customer Management %d", uniqueSuffix)
	domainCapID := fmt.Sprintf("dc-test-%d", uniqueSuffix)
	linkID := fmt.Sprintf("link-test-%d", uniqueSuffix)

	setLinkTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, $2, $3, $4)",
		enterpriseCapID, "default", enterpriseCapName, time.Now().UTC(),
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
	assert.Equal(t, enterpriseCapName, status.LinkedTo.Name)
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

	uniqueSuffix := time.Now().UnixNano()
	enterpriseCapID := fmt.Sprintf("ec-test-%d", uniqueSuffix)
	enterpriseCapName := fmt.Sprintf("Sales Operations %d", uniqueSuffix)
	parentCapID := fmt.Sprintf("parent-%d", uniqueSuffix)
	parentCapName := fmt.Sprintf("Parent Capability %d", uniqueSuffix)
	childCapID := fmt.Sprintf("child-%d", uniqueSuffix)
	linkID := fmt.Sprintf("link-test-%d", uniqueSuffix)

	setLinkTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, $2, $3, $4)",
		enterpriseCapID, "default", enterpriseCapName, time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		parentCapID, "default", parentCapName, nil, "L1", time.Now().UTC(),
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

	_, err = db.Exec(
		"INSERT INTO capability_link_blocking (tenant_id, domain_capability_id, blocked_by_capability_id, blocked_by_enterprise_id, blocked_by_capability_name, blocked_by_enterprise_name, is_ancestor) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		"default", childCapID, parentCapID, enterpriseCapID, parentCapName, enterpriseCapName, true,
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
	assert.Equal(t, parentCapName, status.BlockingCapability.Name)
	require.NotNil(t, status.BlockingEnterpriseCapID)
	assert.Equal(t, enterpriseCapID, *status.BlockingEnterpriseCapID)

	_, err = db.Exec("DELETE FROM capability_link_blocking WHERE tenant_id = 'default' AND domain_capability_id = $1", childCapID)
	require.NoError(t, err)
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

	uniqueSuffix := time.Now().UnixNano()
	enterpriseCapID := fmt.Sprintf("ec-test-%d", uniqueSuffix)
	enterpriseCapName := fmt.Sprintf("Marketing Automation %d", uniqueSuffix)
	parentCapID := fmt.Sprintf("parent-%d", uniqueSuffix)
	childCapID := fmt.Sprintf("child-%d", uniqueSuffix)
	childCapName := fmt.Sprintf("Child Capability %d", uniqueSuffix)
	linkID := fmt.Sprintf("link-test-%d", uniqueSuffix)

	setLinkTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, $2, $3, $4)",
		enterpriseCapID, "default", enterpriseCapName, time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		parentCapID, "default", "Parent Capability", nil, "L1", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		childCapID, "default", childCapName, parentCapID, "L2", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at) VALUES ($1, $2, $3, $4, $5, $6)",
		linkID, "default", enterpriseCapID, childCapID, "user-123", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capability_link_blocking (tenant_id, domain_capability_id, blocked_by_capability_id, blocked_by_enterprise_id, blocked_by_capability_name, blocked_by_enterprise_name, is_ancestor) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		"default", parentCapID, childCapID, enterpriseCapID, childCapName, enterpriseCapName, false,
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
	assert.Equal(t, childCapName, status.BlockingCapability.Name)
	require.NotNil(t, status.BlockingEnterpriseCapID)
	assert.Equal(t, enterpriseCapID, *status.BlockingEnterpriseCapID)

	_, err = db.Exec("DELETE FROM capability_link_blocking WHERE tenant_id = 'default' AND domain_capability_id = $1", parentCapID)
	require.NoError(t, err)
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

	uniqueSuffix := time.Now().UnixNano()
	enterpriseCapID := fmt.Sprintf("ec-test-%d", uniqueSuffix)
	enterpriseCapName := fmt.Sprintf("Product Innovation %d", uniqueSuffix)
	grandparentCapID := fmt.Sprintf("gp-%d", uniqueSuffix)
	grandparentCapName := fmt.Sprintf("Grandparent Capability %d", uniqueSuffix)
	parentCapID := fmt.Sprintf("parent-%d", uniqueSuffix)
	childCapID := fmt.Sprintf("child-%d", uniqueSuffix)
	linkID := fmt.Sprintf("link-test-%d", uniqueSuffix)

	setLinkTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, $2, $3, $4)",
		enterpriseCapID, "default", enterpriseCapName, time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		grandparentCapID, "default", grandparentCapName, nil, "L1", time.Now().UTC(),
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

	_, err = db.Exec(
		"INSERT INTO capability_link_blocking (tenant_id, domain_capability_id, blocked_by_capability_id, blocked_by_enterprise_id, blocked_by_capability_name, blocked_by_enterprise_name, is_ancestor) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		"default", childCapID, grandparentCapID, enterpriseCapID, grandparentCapName, enterpriseCapName, true,
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
	assert.Equal(t, grandparentCapName, status.BlockingCapability.Name)
	require.NotNil(t, status.BlockingEnterpriseCapID)
	assert.Equal(t, enterpriseCapID, *status.BlockingEnterpriseCapID)

	_, err = db.Exec("DELETE FROM capability_link_blocking WHERE tenant_id = 'default' AND domain_capability_id = $1", childCapID)
	require.NoError(t, err)
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

	uniqueSuffix := time.Now().UnixNano()
	enterpriseCapID := fmt.Sprintf("ec-test-%d", uniqueSuffix)
	enterpriseCapName := fmt.Sprintf("Data Analytics %d", uniqueSuffix)
	grandparentCapID := fmt.Sprintf("gp-%d", uniqueSuffix)
	parentCapID := fmt.Sprintf("parent-%d", uniqueSuffix)
	childCapID := fmt.Sprintf("child-%d", uniqueSuffix)
	childCapName := fmt.Sprintf("Child Capability %d", uniqueSuffix)
	linkID := fmt.Sprintf("link-test-%d", uniqueSuffix)

	setLinkTenantContext(t, db)
	_, err := db.Exec(
		"INSERT INTO enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, $2, $3, $4)",
		enterpriseCapID, "default", enterpriseCapName, time.Now().UTC(),
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
		childCapID, "default", childCapName, parentCapID, "L3", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at) VALUES ($1, $2, $3, $4, $5, $6)",
		linkID, "default", enterpriseCapID, childCapID, "user-123", time.Now().UTC(),
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO capability_link_blocking (tenant_id, domain_capability_id, blocked_by_capability_id, blocked_by_enterprise_id, blocked_by_capability_name, blocked_by_enterprise_name, is_ancestor) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		"default", grandparentCapID, childCapID, enterpriseCapID, childCapName, enterpriseCapName, false,
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
	assert.Equal(t, childCapName, status.BlockingCapability.Name)
	require.NotNil(t, status.BlockingEnterpriseCapID)
	assert.Equal(t, enterpriseCapID, *status.BlockingEnterpriseCapID)

	_, err = db.Exec("DELETE FROM capability_link_blocking WHERE tenant_id = 'default' AND domain_capability_id = $1", grandparentCapID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM enterprise_capability_links WHERE id = $1", linkID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM capabilities WHERE id IN ($1, $2, $3)", childCapID, parentCapID, grandparentCapID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM enterprise_capabilities WHERE id = $1", enterpriseCapID)
	require.NoError(t, err)
}
