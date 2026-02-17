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

type linkTestFixture struct {
	t      *testing.T
	rawDB  *sql.DB
	linkRM *EnterpriseCapabilityLinkReadModel
	ecRM   *EnterpriseCapabilityReadModel
	siRM   *EnterpriseStrategicImportanceReadModel
}

func newLinkFixture(t *testing.T) *linkTestFixture {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		"localhost", "5432", "easi_app", "localdev", "easi", "disable")
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() { db.Close() })

	tenantDB := database.NewTenantAwareDB(db)
	return &linkTestFixture{
		t: t, rawDB: db,
		linkRM: NewEnterpriseCapabilityLinkReadModel(tenantDB),
		ecRM:   NewEnterpriseCapabilityReadModel(tenantDB),
		siRM:   NewEnterpriseStrategicImportanceReadModel(tenantDB),
	}
}

func (f *linkTestFixture) ctx() context.Context {
	return sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
}

func (f *linkTestFixture) exec(query string, args ...any) {
	_, err := f.rawDB.Exec("SET app.current_tenant = 'default'")
	require.NoError(f.t, err)
	_, err = f.rawDB.Exec(query, args...)
	require.NoError(f.t, err)
}

func (f *linkTestFixture) queryScalar(dest any, query string, args ...any) {
	_, err := f.rawDB.Exec("SET app.current_tenant = 'default'")
	require.NoError(f.t, err)
	require.NoError(f.t, f.rawDB.QueryRow(query, args...).Scan(dest))
}

type blockingTestResult struct {
	blockedCapID    string
	expectedStatus  LinkStatus
	blockerID       string
	blockerName     string
	enterpriseCapID string
}

func TestGetLinkStatus_DirectlyLinked(t *testing.T) {
	f := newLinkFixture(t)
	s := time.Now().UnixNano()
	ecID, ecName := fmt.Sprintf("ec-%d", s), fmt.Sprintf("Customer Management %d", s)
	dcID, linkID := fmt.Sprintf("dc-%d", s), fmt.Sprintf("link-%d", s)

	f.exec("INSERT INTO enterprisearchitecture.enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, $2, $3, $4)", ecID, "default", ecName, time.Now().UTC())
	t.Cleanup(func() { f.rawDB.Exec("DELETE FROM enterprisearchitecture.enterprise_capabilities WHERE id = $1", ecID) })
	f.exec("INSERT INTO capabilitymapping.capabilities (id, tenant_id, name, level, created_at) VALUES ($1, $2, $3, $4, $5)", dcID, "default", "Order Processing", "L2", time.Now().UTC())
	t.Cleanup(func() { f.rawDB.Exec("DELETE FROM capabilitymapping.capabilities WHERE id = $1", dcID) })
	f.exec("INSERT INTO enterprisearchitecture.enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at) VALUES ($1, $2, $3, $4, $5, $6)", linkID, "default", ecID, dcID, "user-123", time.Now().UTC())
	t.Cleanup(func() { f.rawDB.Exec("DELETE FROM enterprisearchitecture.enterprise_capability_links WHERE id = $1", linkID) })

	status, err := f.linkRM.GetLinkStatus(f.ctx(), dcID)
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, LinkStatusLinked, status.Status)
	require.NotNil(t, status.LinkedTo)
	assert.Equal(t, ecID, status.LinkedTo.ID)
	assert.Equal(t, ecName, status.LinkedTo.Name)
}

func TestGetLinkStatus_Available(t *testing.T) {
	f := newLinkFixture(t)
	dcID := fmt.Sprintf("dc-%d", time.Now().UnixNano())
	f.exec("INSERT INTO capabilitymapping.capabilities (id, tenant_id, name, level, created_at) VALUES ($1, $2, $3, $4, $5)", dcID, "default", "Standalone", "L2", time.Now().UTC())
	t.Cleanup(func() { f.rawDB.Exec("DELETE FROM capabilitymapping.capabilities WHERE id = $1", dcID) })

	status, err := f.linkRM.GetLinkStatus(f.ctx(), dcID)
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, LinkStatusAvailable, status.Status)
	assert.Nil(t, status.LinkedTo)
	assert.Nil(t, status.BlockingCapability)
}

type blockingScenario struct {
	ecName         string
	capNames       []string
	linkedCapIdx   int
	blockedCapIdx  int
	isAncestor     bool
	expectedStatus LinkStatus
}

func setupBlockingScenario(f *linkTestFixture, s int64, sc blockingScenario) blockingTestResult {
	ecID := fmt.Sprintf("ec-%d", s)
	ecName := fmt.Sprintf("%s %d", sc.ecName, s)
	f.exec("INSERT INTO enterprisearchitecture.enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, 'default', $2, $3)", ecID, ecName, time.Now().UTC())
	f.t.Cleanup(func() { f.rawDB.Exec("DELETE FROM enterprisearchitecture.enterprise_capabilities WHERE id = $1", ecID) })

	capIDs := make([]string, len(sc.capNames))
	capNames := make([]string, len(sc.capNames))
	for i, base := range sc.capNames {
		id := fmt.Sprintf("cap%d-%d", i, s)
		capIDs[i] = id
		capNames[i] = fmt.Sprintf("%s %d", base, s)
		var parentID any
		if i > 0 {
			parentID = capIDs[i-1]
		}
		f.exec("INSERT INTO capabilitymapping.capabilities (id, tenant_id, name, parent_id, level, created_at) VALUES ($1, 'default', $2, $3, $4, $5)",
			id, capNames[i], parentID, fmt.Sprintf("L%d", i+1), time.Now().UTC())
		f.t.Cleanup(func() { f.rawDB.Exec("DELETE FROM capabilitymapping.capabilities WHERE id = $1", id) })
	}

	linkID := fmt.Sprintf("lnk-%d", s)
	f.exec("INSERT INTO enterprisearchitecture.enterprise_capability_links (id, tenant_id, enterprise_capability_id, domain_capability_id, linked_by, linked_at) VALUES ($1, 'default', $2, $3, 'user-123', $4)",
		linkID, ecID, capIDs[sc.linkedCapIdx], time.Now().UTC())
	f.t.Cleanup(func() { f.rawDB.Exec("DELETE FROM enterprisearchitecture.enterprise_capability_links WHERE id = $1", linkID) })

	blockedID, blockerID := capIDs[sc.blockedCapIdx], capIDs[sc.linkedCapIdx]
	blockerName := capNames[sc.linkedCapIdx]
	f.exec("INSERT INTO enterprisearchitecture.capability_link_blocking (tenant_id, domain_capability_id, blocked_by_capability_id, blocked_by_enterprise_id, blocked_by_capability_name, blocked_by_enterprise_name, is_ancestor) VALUES ('default', $1, $2, $3, $4, $5, $6)",
		blockedID, blockerID, ecID, blockerName, ecName, sc.isAncestor)
	f.t.Cleanup(func() { f.rawDB.Exec("DELETE FROM enterprisearchitecture.capability_link_blocking WHERE tenant_id = 'default' AND domain_capability_id = $1", blockedID) })

	return blockingTestResult{blockedID, sc.expectedStatus, blockerID, blockerName, ecID}
}

func TestGetLinkStatus_Blocked(t *testing.T) {
	cases := map[string]blockingScenario{
		"by parent":      {ecName: "Sales Ops", capNames: []string{"Parent", "Child"}, linkedCapIdx: 0, blockedCapIdx: 1, isAncestor: true, expectedStatus: LinkStatusBlockedByParent},
		"by child":       {ecName: "Marketing", capNames: []string{"Parent", "Child"}, linkedCapIdx: 1, blockedCapIdx: 0, isAncestor: false, expectedStatus: LinkStatusBlockedByChild},
		"by grandparent": {ecName: "Innovation", capNames: []string{"Grandparent", "Parent", "Child"}, linkedCapIdx: 0, blockedCapIdx: 2, isAncestor: true, expectedStatus: LinkStatusBlockedByParent},
		"by grandchild":  {ecName: "Analytics", capNames: []string{"Grandparent", "Parent", "Child"}, linkedCapIdx: 2, blockedCapIdx: 0, isAncestor: false, expectedStatus: LinkStatusBlockedByChild},
	}

	for name, sc := range cases {
		t.Run(name, func(t *testing.T) {
			f := newLinkFixture(t)
			result := setupBlockingScenario(f, time.Now().UnixNano(), sc)

			status, err := f.linkRM.GetLinkStatus(f.ctx(), result.blockedCapID)
			require.NoError(t, err)
			require.NotNil(t, status)
			assert.Equal(t, result.expectedStatus, status.Status)
			assert.Nil(t, status.LinkedTo)
			require.NotNil(t, status.BlockingCapability)
			assert.Equal(t, result.blockerID, status.BlockingCapability.ID)
			assert.Equal(t, result.blockerName, status.BlockingCapability.Name)
			require.NotNil(t, status.BlockingEnterpriseCapID)
			assert.Equal(t, result.enterpriseCapID, *status.BlockingEnterpriseCapID)
		})
	}
}

func TestEnterpriseCapabilityLinkReadModel_Insert_IdempotentReplay(t *testing.T) {
	f := newLinkFixture(t)
	s := time.Now().UnixNano()
	ecID, dcID := fmt.Sprintf("ec-replay-%d", s), fmt.Sprintf("dc-replay-%d", s)
	linkID := fmt.Sprintf("link-replay-%d", s)

	f.exec("INSERT INTO enterprisearchitecture.enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, 'default', $2, $3)", ecID, "Replay EC", time.Now().UTC())
	t.Cleanup(func() { f.rawDB.Exec("DELETE FROM enterprisearchitecture.enterprise_capabilities WHERE id = $1", ecID) })
	f.exec("INSERT INTO capabilitymapping.capabilities (id, tenant_id, name, level, created_at) VALUES ($1, 'default', $2, $3, $4)", dcID, "Replay DC", "L2", time.Now().UTC())
	t.Cleanup(func() { f.rawDB.Exec("DELETE FROM capabilitymapping.capabilities WHERE id = $1", dcID) })

	dto := EnterpriseCapabilityLinkDTO{
		ID: linkID, EnterpriseCapabilityID: ecID, DomainCapabilityID: dcID,
		LinkedBy: "user-123", LinkedAt: time.Now().UTC(),
	}
	require.NoError(t, f.linkRM.Insert(f.ctx(), dto))
	t.Cleanup(func() { f.rawDB.Exec("DELETE FROM enterprisearchitecture.enterprise_capability_links WHERE id = $1", linkID) })
	require.NoError(t, f.linkRM.Insert(f.ctx(), dto))

	var count int
	f.queryScalar(&count, "SELECT COUNT(*) FROM enterprisearchitecture.enterprise_capability_links WHERE id = $1", linkID)
	assert.Equal(t, 1, count)
}

func TestEnterpriseCapabilityReadModel_InsertReplay(t *testing.T) {
	type replayCase struct {
		modify func(*EnterpriseCapabilityDTO)
		check  func(t *testing.T, f *linkTestFixture, id string)
	}

	cases := map[string]replayCase{
		"idempotent": {
			check: func(t *testing.T, f *linkTestFixture, id string) {
				var count int
				f.queryScalar(&count, "SELECT COUNT(*) FROM enterprisearchitecture.enterprise_capabilities WHERE id = $1", id)
				assert.Equal(t, 1, count)
			},
		},
		"convergence": {
			modify: func(dto *EnterpriseCapabilityDTO) { dto.Name = "Updated Name" },
			check: func(t *testing.T, f *linkTestFixture, id string) {
				var name string
				f.queryScalar(&name, "SELECT name FROM enterprisearchitecture.enterprise_capabilities WHERE id = $1", id)
				assert.Equal(t, "Updated Name", name)
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := newLinkFixture(t)
			id := fmt.Sprintf("ec-%d", time.Now().UnixNano())
			dto := EnterpriseCapabilityDTO{
				ID: id, Name: "Original Name", Description: "Created twice",
				Category: "Business", Active: true, CreatedAt: time.Now().UTC(),
			}
			require.NoError(t, f.ecRM.Insert(f.ctx(), dto))
			t.Cleanup(func() { f.rawDB.Exec("DELETE FROM enterprisearchitecture.enterprise_capabilities WHERE id = $1", id) })
			if tc.modify != nil {
				tc.modify(&dto)
			}
			require.NoError(t, f.ecRM.Insert(f.ctx(), dto))
			tc.check(t, f, id)
		})
	}
}

func TestEnterpriseStrategicImportanceReadModel_Insert_IdempotentReplay(t *testing.T) {
	f := newLinkFixture(t)
	s := time.Now().UnixNano()
	ecID := fmt.Sprintf("ec-si-%d", s)
	f.exec("INSERT INTO enterprisearchitecture.enterprise_capabilities (id, tenant_id, name, created_at) VALUES ($1, 'default', $2, $3)", ecID, "SI EC", time.Now().UTC())
	t.Cleanup(func() { f.rawDB.Exec("DELETE FROM enterprisearchitecture.enterprise_capabilities WHERE id = $1", ecID) })

	id := fmt.Sprintf("esi-replay-%d", s)
	dto := EnterpriseStrategicImportanceDTO{
		ID: id, EnterpriseCapabilityID: ecID, PillarID: "pillar-1",
		PillarName: "Growth", Importance: 5, Rationale: "Critical for growth",
		SetAt: time.Now().UTC(),
	}
	require.NoError(t, f.siRM.Insert(f.ctx(), dto))
	t.Cleanup(func() { f.rawDB.Exec("DELETE FROM enterprisearchitecture.enterprise_strategic_importance WHERE id = $1", id) })
	require.NoError(t, f.siRM.Insert(f.ctx(), dto))

	var count int
	f.queryScalar(&count, "SELECT COUNT(*) FROM enterprisearchitecture.enterprise_strategic_importance WHERE id = $1", id)
	assert.Equal(t, 1, count)
}
