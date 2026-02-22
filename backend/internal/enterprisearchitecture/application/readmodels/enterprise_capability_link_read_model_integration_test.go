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
	db                     *sql.DB
	tenantDB               *database.TenantAwareDB
	readModel              *EnterpriseCapabilityLinkReadModel
	enterpriseCapabilityRM *EnterpriseCapabilityReadModel
	ctx                    context.Context
	t                      *testing.T
}

func newTestFixture(t *testing.T) *testFixture {
	db := setupTestDB(t)
	tenantDB := database.NewTenantAwareDB(db)

	_, err := db.Exec("SET app.current_tenant = 'default'")
	require.NoError(t, err)

	return &testFixture{
		db:                     db,
		tenantDB:               tenantDB,
		readModel:              NewEnterpriseCapabilityLinkReadModel(tenantDB),
		enterpriseCapabilityRM: NewEnterpriseCapabilityReadModel(tenantDB),
		ctx:                    sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID()),
		t:                      t,
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
		_, err = f.db.Exec(`INSERT INTO capabilitymapping.capabilities (id, tenant_id, name, level, parent_id, status, created_at)
			VALUES ($1, 'default', $2, 'L1', NULL, 'Active', NOW())`, id, name)
	} else {
		_, err = f.db.Exec(`INSERT INTO capabilitymapping.capabilities (id, tenant_id, name, level, parent_id, status, created_at)
			VALUES ($1, 'default', $2, 'L1', $3, 'Active', NOW())`, id, name, *parentID)
	}
	require.NoError(f.t, err)
	f.t.Cleanup(func() { f.db.Exec("DELETE FROM capabilitymapping.capabilities WHERE id = $1", id) })
}

func (f *testFixture) createEnterpriseCapability(id, name string) {
	_, err := f.db.Exec(`INSERT INTO enterprisearchitecture.enterprise_capabilities (id, tenant_id, name, description, category, active, link_count, domain_count, created_at)
		VALUES ($1, 'default', $2, $2, 'Test', true, 0, 0, NOW())`, id, name)
	require.NoError(f.t, err)
	f.t.Cleanup(func() { f.db.Exec("DELETE FROM enterprisearchitecture.enterprise_capabilities WHERE id = $1", id) })
}

func (f *testFixture) createLink(id, enterpriseCapID, domainCapID string) {
	_, err := f.db.Exec(`INSERT INTO enterprisearchitecture.enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at)
		VALUES ($1, 'default', $2, $3, 'test@example.com', NOW())`, id, enterpriseCapID, domainCapID)
	require.NoError(f.t, err)
	f.t.Cleanup(func() { f.db.Exec("DELETE FROM enterprisearchitecture.enterprise_capability_links WHERE id = $1", id) })
}

type blockingInfo struct {
	capID          string
	enterpriseID   string
	capName        string
	enterpriseName string
}

func (f *testFixture) createBlocking(domainCapID string, blockedBy blockingInfo, isAncestor bool) {
	_, err := f.db.Exec(`INSERT INTO enterprisearchitecture.capability_link_blocking (tenant_id, domain_capability_id, blocked_by_capability_id, blocked_by_enterprise_id, blocked_by_capability_name, blocked_by_enterprise_name, is_ancestor)
		VALUES ('default', $1, $2, $3, $4, $5, $6)`, domainCapID, blockedBy.capID, blockedBy.enterpriseID, blockedBy.capName, blockedBy.enterpriseName, isAncestor)
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.db.Exec("DELETE FROM enterprisearchitecture.capability_link_blocking WHERE tenant_id = 'default' AND blocked_by_capability_id = $1", blockedBy.capID)
	})
}

func (f *testFixture) setLinkCount(enterpriseCapID string, count int) {
	_, err := f.db.Exec("UPDATE enterprisearchitecture.enterprise_capabilities SET link_count = $1 WHERE id = $2", count, enterpriseCapID)
	require.NoError(f.t, err)
}

func TestCheckHierarchyConflict_RelativeLinkedToDifferent(t *testing.T) {
	tests := []struct {
		name               string
		setupHierarchy     func(f *testFixture) (checkCapID, conflictCapID, ecID1, ecID2 string)
		expectedCapName    string
		expectedIsAncestor bool
	}{
		{
			name: "ancestor linked to different",
			setupHierarchy: func(f *testFixture) (string, string, string, string) {
				grandparentID := f.uniqueID("cap-gp")
				parentID := f.uniqueID("cap-p")
				childID := f.uniqueID("cap-c")
				ecID1 := f.uniqueID("ec-1")
				ecID2 := f.uniqueID("ec-2")

				f.createCapability(grandparentID, "Grandparent", nil)
				f.createCapability(parentID, "Parent", &grandparentID)
				f.createCapability(childID, "Child", &parentID)
				f.createEnterpriseCapability(ecID1, "EC1")
				f.createEnterpriseCapability(ecID2, "EC2")
				f.createLink(f.uniqueID("link"), ecID1, parentID)

				blocking := blockingInfo{capID: parentID, enterpriseID: ecID1, capName: "Parent", enterpriseName: "EC1"}
				f.createBlocking(grandparentID, blocking, false)
				f.createBlocking(childID, blocking, true)

				return childID, parentID, ecID1, ecID2
			},
			expectedCapName:    "Parent",
			expectedIsAncestor: true,
		},
		{
			name: "descendant linked to different",
			setupHierarchy: func(f *testFixture) (string, string, string, string) {
				parentID := f.uniqueID("cap-p")
				childID := f.uniqueID("cap-c")
				ecID1 := f.uniqueID("ec-1")
				ecID2 := f.uniqueID("ec-2")

				f.createCapability(parentID, "Parent", nil)
				f.createCapability(childID, "Child", &parentID)
				f.createEnterpriseCapability(ecID1, "EC1")
				f.createEnterpriseCapability(ecID2, "EC2")
				f.createLink(f.uniqueID("link"), ecID1, childID)

				blocking := blockingInfo{capID: childID, enterpriseID: ecID1, capName: "Child", enterpriseName: "EC1"}
				f.createBlocking(parentID, blocking, false)

				return parentID, childID, ecID1, ecID2
			},
			expectedCapName:    "Child",
			expectedIsAncestor: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newTestFixture(t)
			checkCapID, conflictCapID, ecID1, ecID2 := tc.setupHierarchy(f)

			conflict, err := f.readModel.CheckHierarchyConflict(f.ctx, checkCapID, ecID2)
			require.NoError(t, err)
			require.NotNil(t, conflict)

			assert.Equal(t, conflictCapID, conflict.ConflictingCapabilityID)
			assert.Equal(t, tc.expectedCapName, conflict.ConflictingCapabilityName)
			assert.Equal(t, ecID1, conflict.LinkedToCapabilityID)
			assert.Equal(t, "EC1", conflict.LinkedToCapabilityName)
			assert.Equal(t, tc.expectedIsAncestor, conflict.IsAncestor)
		})
	}
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
	f.createBlocking(parentID, blockingInfo{capID: childID, enterpriseID: enterpriseCapID, capName: "Child", enterpriseName: "EC"}, false)

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

	blocking := blockingInfo{capID: capabilityIDs[8], enterpriseID: enterpriseCapID1, capName: "Level 8", enterpriseName: "EC1"}
	for i := 0; i < 8; i++ {
		f.createBlocking(capabilityIDs[i], blocking, false)
	}
	for i := 9; i < 12; i++ {
		f.createBlocking(capabilityIDs[i], blocking, true)
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
	err := f.db.QueryRow("SELECT link_count FROM enterprisearchitecture.enterprise_capabilities WHERE tenant_id = 'default' AND id = $1", enterpriseCapID).Scan(&count)
	require.NoError(f.t, err)
	return count
}

func (f *testFixture) getDomainCount(enterpriseCapID string) int {
	var count int
	err := f.db.QueryRow("SELECT domain_count FROM enterprisearchitecture.enterprise_capabilities WHERE tenant_id = 'default' AND id = $1", enterpriseCapID).Scan(&count)
	require.NoError(f.t, err)
	return count
}

func (f *testFixture) createCapabilityMetadata(capabilityID, capabilityName string, domainID, domainName *string) {
	var err error
	if domainID != nil {
		_, err = f.db.Exec(`INSERT INTO enterprisearchitecture.domain_capability_metadata (tenant_id, capability_id, capability_name, capability_level, l1_capability_id, business_domain_id, business_domain_name)
			VALUES ('default', $1, $2, 'L1', $1, $3, $4) ON CONFLICT DO NOTHING`, capabilityID, capabilityName, *domainID, *domainName)
	} else {
		_, err = f.db.Exec(`INSERT INTO enterprisearchitecture.domain_capability_metadata (tenant_id, capability_id, capability_name, capability_level, l1_capability_id)
			VALUES ('default', $1, $2, 'L1', $1) ON CONFLICT DO NOTHING`, capabilityID, capabilityName)
	}
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.db.Exec("DELETE FROM enterprisearchitecture.domain_capability_metadata WHERE capability_id = $1", capabilityID)
	})
}

func TestLinkCount_IncrementAndDecrement(t *testing.T) {
	tests := []struct {
		name          string
		initialCount  int
		operation     func(*EnterpriseCapabilityReadModel, context.Context, string) error
		expectedCount int
	}{
		{
			name:         "increment from zero",
			initialCount: 0,
			operation: func(rm *EnterpriseCapabilityReadModel, ctx context.Context, id string) error {
				return rm.IncrementLinkCount(ctx, id)
			},
			expectedCount: 1,
		},
		{
			name:         "increment from nonzero",
			initialCount: 1,
			operation: func(rm *EnterpriseCapabilityReadModel, ctx context.Context, id string) error {
				return rm.IncrementLinkCount(ctx, id)
			},
			expectedCount: 2,
		},
		{
			name:         "decrement from nonzero",
			initialCount: 3,
			operation: func(rm *EnterpriseCapabilityReadModel, ctx context.Context, id string) error {
				return rm.DecrementLinkCount(ctx, id)
			},
			expectedCount: 2,
		},
		{
			name:         "decrement does not go below zero",
			initialCount: 0,
			operation: func(rm *EnterpriseCapabilityReadModel, ctx context.Context, id string) error {
				return rm.DecrementLinkCount(ctx, id)
			},
			expectedCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newTestFixture(t)

			enterpriseCapID := f.uniqueID("ec")
			f.createEnterpriseCapability(enterpriseCapID, "EC")
			f.setLinkCount(enterpriseCapID, tc.initialCount)

			assert.Equal(t, tc.initialCount, f.getLinkCount(enterpriseCapID))

			err := tc.operation(f.enterpriseCapabilityRM, f.ctx, enterpriseCapID)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedCount, f.getLinkCount(enterpriseCapID))
		})
	}
}

func TestIncrementAndDecrement_CounterConsistency(t *testing.T) {
	f := newTestFixture(t)

	enterpriseCapID := f.uniqueID("ec")
	f.createEnterpriseCapability(enterpriseCapID, "EC")

	for i := 0; i < 5; i++ {
		err := f.enterpriseCapabilityRM.IncrementLinkCount(f.ctx, enterpriseCapID)
		require.NoError(t, err)
	}
	assert.Equal(t, 5, f.getLinkCount(enterpriseCapID))

	for i := 0; i < 3; i++ {
		err := f.enterpriseCapabilityRM.DecrementLinkCount(f.ctx, enterpriseCapID)
		require.NoError(t, err)
	}
	assert.Equal(t, 2, f.getLinkCount(enterpriseCapID))

	for i := 0; i < 5; i++ {
		err := f.enterpriseCapabilityRM.DecrementLinkCount(f.ctx, enterpriseCapID)
		require.NoError(t, err)
	}
	assert.Equal(t, 0, f.getLinkCount(enterpriseCapID))
}

func TestRecalculateDomainCount_AfterIncrementLinkCount(t *testing.T) {
	f := newTestFixture(t)

	enterpriseCapID := f.uniqueID("ec")
	capID1 := f.uniqueID("cap-1")
	capID2 := f.uniqueID("cap-2")
	domainID1 := f.uniqueID("domain-1")
	domainID2 := f.uniqueID("domain-2")
	domainName1 := "Domain 1"
	domainName2 := "Domain 2"

	f.createEnterpriseCapability(enterpriseCapID, "EC")
	f.createCapability(capID1, "Cap1", nil)
	f.createCapability(capID2, "Cap2", nil)

	f.createCapabilityMetadata(capID1, "Cap1", &domainID1, &domainName1)
	f.createCapabilityMetadata(capID2, "Cap2", &domainID2, &domainName2)

	f.createLink(f.uniqueID("link1"), enterpriseCapID, capID1)

	err := f.enterpriseCapabilityRM.IncrementLinkCount(f.ctx, enterpriseCapID)
	require.NoError(t, err)
	err = f.enterpriseCapabilityRM.RecalculateDomainCount(f.ctx, enterpriseCapID)
	require.NoError(t, err)

	domainCount := f.getDomainCount(enterpriseCapID)
	assert.Equal(t, 1, domainCount)

	f.createLink(f.uniqueID("link2"), enterpriseCapID, capID2)

	err = f.enterpriseCapabilityRM.IncrementLinkCount(f.ctx, enterpriseCapID)
	require.NoError(t, err)
	err = f.enterpriseCapabilityRM.RecalculateDomainCount(f.ctx, enterpriseCapID)
	require.NoError(t, err)

	domainCount = f.getDomainCount(enterpriseCapID)
	assert.Equal(t, 2, domainCount)
}

func TestRecalculateDomainCount_AfterDecrementLinkCount(t *testing.T) {
	f := newTestFixture(t)

	enterpriseCapID := f.uniqueID("ec")
	capID1 := f.uniqueID("cap-1")
	domainID1 := f.uniqueID("domain-1")
	domainName1 := "Domain 1"

	f.createEnterpriseCapability(enterpriseCapID, "EC")
	f.createCapability(capID1, "Cap1", nil)
	f.createCapabilityMetadata(capID1, "Cap1", &domainID1, &domainName1)
	f.createLink(f.uniqueID("link1"), enterpriseCapID, capID1)

	_, err := f.db.Exec("UPDATE enterprisearchitecture.enterprise_capabilities SET link_count = 1, domain_count = 1 WHERE id = $1", enterpriseCapID)
	require.NoError(t, err)

	_, err = f.db.Exec("DELETE FROM enterprisearchitecture.enterprise_capability_links WHERE enterprise_capability_id = $1", enterpriseCapID)
	require.NoError(t, err)

	err = f.enterpriseCapabilityRM.DecrementLinkCount(f.ctx, enterpriseCapID)
	require.NoError(t, err)
	err = f.enterpriseCapabilityRM.RecalculateDomainCount(f.ctx, enterpriseCapID)
	require.NoError(t, err)

	domainCount := f.getDomainCount(enterpriseCapID)
	assert.Equal(t, 0, domainCount)
	assert.Equal(t, 0, f.getLinkCount(enterpriseCapID))
}
