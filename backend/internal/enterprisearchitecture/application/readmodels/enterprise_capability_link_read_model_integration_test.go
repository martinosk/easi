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

type testFixture struct {
	db                         *sql.DB
	tenantDB                   *database.TenantAwareDB
	readModel                  *EnterpriseCapabilityLinkReadModel
	enterpriseCapabilityRM     *EnterpriseCapabilityReadModel
	ctx                        context.Context
	t                          *testing.T
}

func newTestFixture(t *testing.T) *testFixture {
	db := setupTestDB(t)
	tenantDB := database.NewTenantAwareDB(db)

	_, err := db.Exec("SET app.current_tenant = 'default'")
	require.NoError(t, err)

	return &testFixture{
		db:                         db,
		tenantDB:                   tenantDB,
		readModel:                  NewEnterpriseCapabilityLinkReadModel(tenantDB),
		enterpriseCapabilityRM:     NewEnterpriseCapabilityReadModel(tenantDB),
		ctx:                        sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID()),
		t:                          t,
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		"localhost", "5432", "easi_app", "localdev", "easi", "disable")
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() { db.Close() })
	return db
}

func (f *testFixture) uniqueID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func (f *testFixture) createCapability(id, name string, parentID *string) {
	var err error
	if parentID == nil {
		_, err = f.db.Exec(`INSERT INTO capabilities (id, tenant_id, name, level, parent_id, status, created_at)
			VALUES ($1, 'default', $2, 'L1', NULL, 'active', NOW())`, id, name)
	} else {
		_, err = f.db.Exec(`INSERT INTO capabilities (id, tenant_id, name, level, parent_id, status, created_at)
			VALUES ($1, 'default', $2, 'L1', $3, 'active', NOW())`, id, name, *parentID)
	}
	require.NoError(f.t, err)
	f.t.Cleanup(func() { f.db.Exec("DELETE FROM capabilities WHERE id = $1", id) })
}

func (f *testFixture) createEnterpriseCapability(id, name string) {
	_, err := f.db.Exec(`INSERT INTO enterprise_capabilities (id, tenant_id, name, description, category, active, link_count, domain_count, created_at)
		VALUES ($1, 'default', $2, $2, 'Test', true, 0, 0, NOW())`, id, name)
	require.NoError(f.t, err)
	f.t.Cleanup(func() { f.db.Exec("DELETE FROM enterprise_capabilities WHERE id = $1", id) })
}

func (f *testFixture) createLink(id, enterpriseCapID, domainCapID string) {
	_, err := f.db.Exec(`INSERT INTO enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at)
		VALUES ($1, 'default', $2, $3, 'test@example.com', NOW())`, id, enterpriseCapID, domainCapID)
	require.NoError(f.t, err)
	f.t.Cleanup(func() { f.db.Exec("DELETE FROM enterprise_capability_links WHERE id = $1", id) })
}

func (f *testFixture) createBlocking(domainCapID, blockedByCapID, blockedByEnterpriseID, blockedByCapName, blockedByEnterpriseName string, isAncestor bool) {
	_, err := f.db.Exec(`INSERT INTO capability_link_blocking (tenant_id, domain_capability_id, blocked_by_capability_id, blocked_by_enterprise_id, blocked_by_capability_name, blocked_by_enterprise_name, is_ancestor)
		VALUES ('default', $1, $2, $3, $4, $5, $6)`, domainCapID, blockedByCapID, blockedByEnterpriseID, blockedByCapName, blockedByEnterpriseName, isAncestor)
	require.NoError(f.t, err)
	f.t.Cleanup(func() { f.db.Exec("DELETE FROM capability_link_blocking WHERE tenant_id = 'default' AND blocked_by_capability_id = $1", blockedByCapID) })
}

func TestCheckHierarchyConflict_AncestorLinkedToDifferent(t *testing.T) {
	f := newTestFixture(t)

	grandparentID := f.uniqueID("cap-gp")
	parentID := f.uniqueID("cap-p")
	childID := f.uniqueID("cap-c")
	enterpriseCapID1 := f.uniqueID("ec-1")
	enterpriseCapID2 := f.uniqueID("ec-2")

	f.createCapability(grandparentID, "Grandparent", nil)
	f.createCapability(parentID, "Parent", &grandparentID)
	f.createCapability(childID, "Child", &parentID)
	f.createEnterpriseCapability(enterpriseCapID1, "EC1")
	f.createEnterpriseCapability(enterpriseCapID2, "EC2")
	f.createLink(f.uniqueID("link"), enterpriseCapID1, parentID)
	f.createBlocking(grandparentID, parentID, enterpriseCapID1, "Parent", "EC1", false)
	f.createBlocking(childID, parentID, enterpriseCapID1, "Parent", "EC1", true)

	conflict, err := f.readModel.CheckHierarchyConflict(f.ctx, childID, enterpriseCapID2)
	require.NoError(t, err)
	require.NotNil(t, conflict)

	assert.Equal(t, parentID, conflict.ConflictingCapabilityID)
	assert.Equal(t, "Parent", conflict.ConflictingCapabilityName)
	assert.Equal(t, enterpriseCapID1, conflict.LinkedToCapabilityID)
	assert.Equal(t, "EC1", conflict.LinkedToCapabilityName)
	assert.True(t, conflict.IsAncestor)
}

func TestCheckHierarchyConflict_DescendantLinkedToDifferent(t *testing.T) {
	f := newTestFixture(t)

	parentID := f.uniqueID("cap-p")
	childID := f.uniqueID("cap-c")
	enterpriseCapID1 := f.uniqueID("ec-1")
	enterpriseCapID2 := f.uniqueID("ec-2")

	f.createCapability(parentID, "Parent", nil)
	f.createCapability(childID, "Child", &parentID)
	f.createEnterpriseCapability(enterpriseCapID1, "EC1")
	f.createEnterpriseCapability(enterpriseCapID2, "EC2")
	f.createLink(f.uniqueID("link"), enterpriseCapID1, childID)
	f.createBlocking(parentID, childID, enterpriseCapID1, "Child", "EC1", false)

	conflict, err := f.readModel.CheckHierarchyConflict(f.ctx, parentID, enterpriseCapID2)
	require.NoError(t, err)
	require.NotNil(t, conflict)

	assert.Equal(t, childID, conflict.ConflictingCapabilityID)
	assert.Equal(t, "Child", conflict.ConflictingCapabilityName)
	assert.Equal(t, enterpriseCapID1, conflict.LinkedToCapabilityID)
	assert.Equal(t, "EC1", conflict.LinkedToCapabilityName)
	assert.False(t, conflict.IsAncestor)
}

func TestCheckHierarchyConflict_SameEnterpriseCapability_NoConflict(t *testing.T) {
	f := newTestFixture(t)

	parentID := f.uniqueID("cap-p")
	childID := f.uniqueID("cap-c")
	enterpriseCapID := f.uniqueID("ec")

	f.createCapability(parentID, "Parent", nil)
	f.createCapability(childID, "Child", &parentID)
	f.createEnterpriseCapability(enterpriseCapID, "EC")
	f.createLink(f.uniqueID("link"), enterpriseCapID, childID)
	f.createBlocking(parentID, childID, enterpriseCapID, "Child", "EC", false)

	conflict, err := f.readModel.CheckHierarchyConflict(f.ctx, parentID, enterpriseCapID)
	require.NoError(t, err)
	assert.Nil(t, conflict)
}

func TestCheckHierarchyConflict_DeepHierarchy_HandlesDepthLimit(t *testing.T) {
	f := newTestFixture(t)

	capabilityIDs := make([]string, 12)
	for i := 0; i < 12; i++ {
		capabilityIDs[i] = f.uniqueID(fmt.Sprintf("cap-deep-%d", i))
	}

	enterpriseCapID1 := f.uniqueID("ec-1")
	enterpriseCapID2 := f.uniqueID("ec-2")

	f.createEnterpriseCapability(enterpriseCapID1, "EC1")
	f.createEnterpriseCapability(enterpriseCapID2, "EC2")

	var parentID *string
	for i := 0; i < 12; i++ {
		f.createCapability(capabilityIDs[i], fmt.Sprintf("Level %d", i), parentID)
		parentID = &capabilityIDs[i]
	}

	f.createLink(f.uniqueID("link"), enterpriseCapID1, capabilityIDs[8])

	for i := 0; i < 8; i++ {
		f.createBlocking(capabilityIDs[i], capabilityIDs[8], enterpriseCapID1, "Level 8", "EC1", false)
	}
	for i := 9; i < 12; i++ {
		f.createBlocking(capabilityIDs[i], capabilityIDs[8], enterpriseCapID1, "Level 8", "EC1", true)
	}

	conflict, err := f.readModel.CheckHierarchyConflict(f.ctx, capabilityIDs[11], enterpriseCapID2)
	require.NoError(t, err)
	require.NotNil(t, conflict)
	assert.True(t, conflict.IsAncestor)

	conflict, err = f.readModel.CheckHierarchyConflict(f.ctx, capabilityIDs[0], enterpriseCapID2)
	require.NoError(t, err)
	require.NotNil(t, conflict)
	assert.False(t, conflict.IsAncestor)
}

func TestCheckHierarchyConflict_NoRelationship_NoConflict(t *testing.T) {
	f := newTestFixture(t)

	cap1ID := f.uniqueID("cap-1")
	cap2ID := f.uniqueID("cap-2")
	enterpriseCapID1 := f.uniqueID("ec-1")
	enterpriseCapID2 := f.uniqueID("ec-2")

	f.createCapability(cap1ID, "Cap1", nil)
	f.createCapability(cap2ID, "Cap2", nil)
	f.createEnterpriseCapability(enterpriseCapID1, "EC1")
	f.createEnterpriseCapability(enterpriseCapID2, "EC2")
	f.createLink(f.uniqueID("link"), enterpriseCapID1, cap2ID)

	conflict, err := f.readModel.CheckHierarchyConflict(f.ctx, cap1ID, enterpriseCapID2)
	require.NoError(t, err)
	assert.Nil(t, conflict)
}

func (f *testFixture) getLinkCount(enterpriseCapID string) int {
	var count int
	err := f.db.QueryRow("SELECT link_count FROM enterprise_capabilities WHERE tenant_id = 'default' AND id = $1", enterpriseCapID).Scan(&count)
	require.NoError(f.t, err)
	return count
}

func (f *testFixture) getDomainCount(enterpriseCapID string) int {
	var count int
	err := f.db.QueryRow("SELECT domain_count FROM enterprise_capabilities WHERE tenant_id = 'default' AND id = $1", enterpriseCapID).Scan(&count)
	require.NoError(f.t, err)
	return count
}

func (f *testFixture) createBusinessDomainAssignment(capabilityID, domainID, domainName string) {
	_, err := f.db.Exec(`INSERT INTO domain_capability_assignments (tenant_id, capability_id, business_domain_id, business_domain_name)
		VALUES ('default', $1, $2, $3) ON CONFLICT DO NOTHING`, capabilityID, domainID, domainName)
	require.NoError(f.t, err)
	f.t.Cleanup(func() { f.db.Exec("DELETE FROM domain_capability_assignments WHERE capability_id = $1", capabilityID) })
}

func TestIncrementLinkCountAndRecalculateDomainCount_IncrementsLinkCount(t *testing.T) {
	f := newTestFixture(t)

	enterpriseCapID := f.uniqueID("ec")
	f.createEnterpriseCapability(enterpriseCapID, "EC")

	initialCount := f.getLinkCount(enterpriseCapID)
	assert.Equal(t, 0, initialCount)

	err := f.enterpriseCapabilityRM.IncrementLinkCountAndRecalculateDomainCount(f.ctx, enterpriseCapID)
	require.NoError(t, err)

	afterIncrement := f.getLinkCount(enterpriseCapID)
	assert.Equal(t, 1, afterIncrement)

	err = f.enterpriseCapabilityRM.IncrementLinkCountAndRecalculateDomainCount(f.ctx, enterpriseCapID)
	require.NoError(t, err)

	afterSecondIncrement := f.getLinkCount(enterpriseCapID)
	assert.Equal(t, 2, afterSecondIncrement)
}

func TestDecrementLinkCountAndRecalculateDomainCount_DecrementsLinkCount(t *testing.T) {
	f := newTestFixture(t)

	enterpriseCapID := f.uniqueID("ec")
	f.createEnterpriseCapability(enterpriseCapID, "EC")

	_, err := f.db.Exec("UPDATE enterprise_capabilities SET link_count = 3 WHERE id = $1", enterpriseCapID)
	require.NoError(t, err)

	initialCount := f.getLinkCount(enterpriseCapID)
	assert.Equal(t, 3, initialCount)

	err = f.enterpriseCapabilityRM.DecrementLinkCountAndRecalculateDomainCount(f.ctx, enterpriseCapID)
	require.NoError(t, err)

	afterDecrement := f.getLinkCount(enterpriseCapID)
	assert.Equal(t, 2, afterDecrement)
}

func TestDecrementLinkCountAndRecalculateDomainCount_DoesNotGoBelowZero(t *testing.T) {
	f := newTestFixture(t)

	enterpriseCapID := f.uniqueID("ec")
	f.createEnterpriseCapability(enterpriseCapID, "EC")

	initialCount := f.getLinkCount(enterpriseCapID)
	assert.Equal(t, 0, initialCount)

	err := f.enterpriseCapabilityRM.DecrementLinkCountAndRecalculateDomainCount(f.ctx, enterpriseCapID)
	require.NoError(t, err)

	afterDecrement := f.getLinkCount(enterpriseCapID)
	assert.Equal(t, 0, afterDecrement)
}

func TestIncrementAndDecrement_CounterConsistency(t *testing.T) {
	f := newTestFixture(t)

	enterpriseCapID := f.uniqueID("ec")
	f.createEnterpriseCapability(enterpriseCapID, "EC")

	for i := 0; i < 5; i++ {
		err := f.enterpriseCapabilityRM.IncrementLinkCountAndRecalculateDomainCount(f.ctx, enterpriseCapID)
		require.NoError(t, err)
	}
	assert.Equal(t, 5, f.getLinkCount(enterpriseCapID))

	for i := 0; i < 3; i++ {
		err := f.enterpriseCapabilityRM.DecrementLinkCountAndRecalculateDomainCount(f.ctx, enterpriseCapID)
		require.NoError(t, err)
	}
	assert.Equal(t, 2, f.getLinkCount(enterpriseCapID))

	for i := 0; i < 5; i++ {
		err := f.enterpriseCapabilityRM.DecrementLinkCountAndRecalculateDomainCount(f.ctx, enterpriseCapID)
		require.NoError(t, err)
	}
	assert.Equal(t, 0, f.getLinkCount(enterpriseCapID))
}

func TestIncrementLinkCountAndRecalculateDomainCount_RecalculatesDomainCount(t *testing.T) {
	f := newTestFixture(t)

	enterpriseCapID := f.uniqueID("ec")
	capID1 := f.uniqueID("cap-1")
	capID2 := f.uniqueID("cap-2")
	domainID1 := f.uniqueID("domain-1")
	domainID2 := f.uniqueID("domain-2")

	f.createEnterpriseCapability(enterpriseCapID, "EC")
	f.createCapability(capID1, "Cap1", nil)
	f.createCapability(capID2, "Cap2", nil)

	f.createBusinessDomainAssignment(capID1, domainID1, "Domain 1")
	f.createBusinessDomainAssignment(capID2, domainID2, "Domain 2")

	f.createLink(f.uniqueID("link1"), enterpriseCapID, capID1)

	err := f.enterpriseCapabilityRM.IncrementLinkCountAndRecalculateDomainCount(f.ctx, enterpriseCapID)
	require.NoError(t, err)

	domainCount := f.getDomainCount(enterpriseCapID)
	assert.Equal(t, 1, domainCount)

	f.createLink(f.uniqueID("link2"), enterpriseCapID, capID2)

	err = f.enterpriseCapabilityRM.IncrementLinkCountAndRecalculateDomainCount(f.ctx, enterpriseCapID)
	require.NoError(t, err)

	domainCount = f.getDomainCount(enterpriseCapID)
	assert.Equal(t, 2, domainCount)
}

func TestDecrementLinkCountAndRecalculateDomainCount_RecalculatesDomainCount(t *testing.T) {
	f := newTestFixture(t)

	enterpriseCapID := f.uniqueID("ec")
	capID1 := f.uniqueID("cap-1")
	domainID1 := f.uniqueID("domain-1")

	f.createEnterpriseCapability(enterpriseCapID, "EC")
	f.createCapability(capID1, "Cap1", nil)
	f.createBusinessDomainAssignment(capID1, domainID1, "Domain 1")
	f.createLink(f.uniqueID("link1"), enterpriseCapID, capID1)

	_, err := f.db.Exec("UPDATE enterprise_capabilities SET link_count = 1, domain_count = 1 WHERE id = $1", enterpriseCapID)
	require.NoError(t, err)

	_, err = f.db.Exec("DELETE FROM enterprise_capability_links WHERE enterprise_capability_id = $1", enterpriseCapID)
	require.NoError(t, err)

	err = f.enterpriseCapabilityRM.DecrementLinkCountAndRecalculateDomainCount(f.ctx, enterpriseCapID)
	require.NoError(t, err)

	domainCount := f.getDomainCount(enterpriseCapID)
	assert.Equal(t, 0, domainCount)
	assert.Equal(t, 0, f.getLinkCount(enterpriseCapID))
}
