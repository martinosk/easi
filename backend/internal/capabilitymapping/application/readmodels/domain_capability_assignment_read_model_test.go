//go:build integration

package readmodels

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"easi/backend/internal/infrastructure/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type assignmentTestFixture struct {
	t         *testing.T
	rawDB     *sql.DB
	readModel *DomainCapabilityAssignmentReadModel
}

func newAssignmentFixture(t *testing.T) *assignmentTestFixture {
	db, cleanup := setupTestDB(t)
	t.Cleanup(cleanup)
	tenantDB := database.NewTenantAwareDB(db)
	return &assignmentTestFixture{
		t:         t,
		rawDB:     db,
		readModel: NewDomainCapabilityAssignmentReadModel(tenantDB),
	}
}

func (f *assignmentTestFixture) makeDTO(prefix string) AssignmentDTO {
	suffix := time.Now().UnixNano()
	return AssignmentDTO{
		AssignmentID:       fmt.Sprintf("%s-assign-%d", prefix, suffix),
		BusinessDomainID:   fmt.Sprintf("%s-bd-%d", prefix, suffix),
		BusinessDomainName: "Finance",
		CapabilityID:       fmt.Sprintf("%s-cap-%d", prefix, suffix),
		CapabilityName:     "Financial Reporting",
		CapabilityLevel:    "L1",
		AssignedAt:         time.Now().UTC(),
	}
}

func (f *assignmentTestFixture) insertViaSQL(dto AssignmentDTO) {
	setTenantContext(f.t, f.rawDB)
	_, err := f.rawDB.Exec(
		"INSERT INTO domain_capability_assignments (assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_name, capability_level, assigned_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		dto.AssignmentID, "default", dto.BusinessDomainID, dto.BusinessDomainName,
		dto.CapabilityID, dto.CapabilityName, dto.CapabilityLevel, dto.AssignedAt,
	)
	require.NoError(f.t, err)
	f.t.Cleanup(func() {
		f.rawDB.Exec("DELETE FROM domain_capability_assignments WHERE assignment_id = $1", dto.AssignmentID)
	})
}

func (f *assignmentTestFixture) insertViaReadModel(dto AssignmentDTO) {
	require.NoError(f.t, f.readModel.Insert(tenantContext(), dto))
	f.t.Cleanup(func() {
		f.rawDB.Exec("DELETE FROM domain_capability_assignments WHERE assignment_id = $1", dto.AssignmentID)
	})
}

func (f *assignmentTestFixture) queryScalar(dest any, query string, args ...any) {
	setTenantContext(f.t, f.rawDB)
	require.NoError(f.t, f.rawDB.QueryRow(query, args...).Scan(dest))
}

func TestDomainCapabilityAssignmentReadModel_Insert(t *testing.T) {
	f := newAssignmentFixture(t)
	dto := f.makeDTO("test")
	f.insertViaReadModel(dto)

	var domainID, capID string
	f.queryScalar(&domainID, "SELECT business_domain_id FROM domain_capability_assignments WHERE assignment_id = $1", dto.AssignmentID)
	f.queryScalar(&capID, "SELECT capability_id FROM domain_capability_assignments WHERE assignment_id = $1", dto.AssignmentID)
	assert.Equal(t, dto.BusinessDomainID, domainID)
	assert.Equal(t, dto.CapabilityID, capID)
}

func TestDomainCapabilityAssignmentReadModel_Insert_IdempotentReplay(t *testing.T) {
	f := newAssignmentFixture(t)
	dto := f.makeDTO("replay")
	f.insertViaReadModel(dto)
	f.insertViaReadModel(dto)

	var count int
	f.queryScalar(&count, "SELECT COUNT(*) FROM domain_capability_assignments WHERE assignment_id = $1", dto.AssignmentID)
	assert.Equal(t, 1, count)
}

func TestDomainCapabilityAssignmentReadModel_Insert_ReplayConvergence(t *testing.T) {
	f := newAssignmentFixture(t)
	dto := f.makeDTO("converge")
	f.insertViaReadModel(dto)

	dto.CapabilityName = "Updated Name"
	f.insertViaReadModel(dto)

	var name string
	f.queryScalar(&name, "SELECT capability_name FROM domain_capability_assignments WHERE assignment_id = $1", dto.AssignmentID)
	assert.Equal(t, "Updated Name", name)
}

func TestDomainCapabilityAssignmentReadModel_Delete(t *testing.T) {
	f := newAssignmentFixture(t)
	dto := f.makeDTO("delete")
	f.insertViaSQL(dto)

	require.NoError(t, f.readModel.Delete(tenantContext(), dto.AssignmentID))

	var count int
	f.queryScalar(&count, "SELECT COUNT(*) FROM domain_capability_assignments WHERE assignment_id = $1", dto.AssignmentID)
	assert.Equal(t, 0, count)
}

func TestDomainCapabilityAssignmentReadModel_GetByRelation(t *testing.T) {
	type getByRelationCase struct {
		sharedField string
		makeDTOs    func(f *assignmentTestFixture, suffix int64, sharedID string) (AssignmentDTO, AssignmentDTO)
		query       func(rm *DomainCapabilityAssignmentReadModel, sharedID string) ([]AssignmentDTO, error)
		verifyPair  func(t *testing.T, dto1, dto2 AssignmentDTO, results []AssignmentDTO)
	}

	cases := map[string]getByRelationCase{
		"by domain ID": {
			sharedField: "bd",
			makeDTOs: func(f *assignmentTestFixture, suffix int64, domainID string) (AssignmentDTO, AssignmentDTO) {
				return AssignmentDTO{
						AssignmentID: fmt.Sprintf("a1-%d", suffix), BusinessDomainID: domainID,
						BusinessDomainName: "Finance", CapabilityID: fmt.Sprintf("c1-%d", suffix),
						CapabilityName: "Financial Reporting", CapabilityLevel: "L1", AssignedAt: time.Now().UTC(),
					}, AssignmentDTO{
						AssignmentID: fmt.Sprintf("a2-%d", suffix), BusinessDomainID: domainID,
						BusinessDomainName: "Finance", CapabilityID: fmt.Sprintf("c2-%d", suffix),
						CapabilityName: "Budget Planning", CapabilityLevel: "L1", AssignedAt: time.Now().UTC(),
					}
			},
			query: func(rm *DomainCapabilityAssignmentReadModel, id string) ([]AssignmentDTO, error) {
				return rm.GetByDomainID(tenantContext(), id)
			},
			verifyPair: func(t *testing.T, dto1, dto2 AssignmentDTO, results []AssignmentDTO) {
				namesByID := make(map[string]string)
				for _, a := range results {
					namesByID[a.AssignmentID] = a.CapabilityName
				}
				assert.Equal(t, "Financial Reporting", namesByID[dto1.AssignmentID])
				assert.Equal(t, "Budget Planning", namesByID[dto2.AssignmentID])
			},
		},
		"by capability ID": {
			sharedField: "cap",
			makeDTOs: func(f *assignmentTestFixture, suffix int64, capabilityID string) (AssignmentDTO, AssignmentDTO) {
				return AssignmentDTO{
						AssignmentID: fmt.Sprintf("a1-%d", suffix), BusinessDomainID: fmt.Sprintf("d1-%d", suffix),
						BusinessDomainName: "Finance", CapabilityID: capabilityID,
						CapabilityName: "Financial Reporting", CapabilityLevel: "L1", AssignedAt: time.Now().UTC(),
					}, AssignmentDTO{
						AssignmentID: fmt.Sprintf("a2-%d", suffix), BusinessDomainID: fmt.Sprintf("d2-%d", suffix),
						BusinessDomainName: "Operations", CapabilityID: capabilityID,
						CapabilityName: "Financial Reporting", CapabilityLevel: "L1", AssignedAt: time.Now().UTC(),
					}
			},
			query: func(rm *DomainCapabilityAssignmentReadModel, id string) ([]AssignmentDTO, error) {
				return rm.GetByCapabilityID(tenantContext(), id)
			},
			verifyPair: func(t *testing.T, dto1, dto2 AssignmentDTO, results []AssignmentDTO) {
				namesByDomain := make(map[string]string)
				for _, a := range results {
					namesByDomain[a.BusinessDomainID] = a.BusinessDomainName
				}
				assert.Equal(t, "Finance", namesByDomain[dto1.BusinessDomainID])
				assert.Equal(t, "Operations", namesByDomain[dto2.BusinessDomainID])
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := newAssignmentFixture(t)
			suffix := time.Now().UnixNano()
			sharedID := fmt.Sprintf("%s-shared-%d", tc.sharedField, suffix)

			dto1, dto2 := tc.makeDTOs(f, suffix, sharedID)
			f.insertViaSQL(dto1)
			f.insertViaSQL(dto2)

			results, err := tc.query(f.readModel, sharedID)
			require.NoError(t, err)
			assert.Len(t, results, 2)
			tc.verifyPair(t, dto1, dto2, results)
		})
	}
}

func TestDomainCapabilityAssignmentReadModel_GetByDomainAndCapability(t *testing.T) {
	f := newAssignmentFixture(t)
	dto := f.makeDTO("lookup")
	f.insertViaSQL(dto)

	assignment, err := f.readModel.GetByDomainAndCapability(tenantContext(), dto.BusinessDomainID, dto.CapabilityID)
	require.NoError(t, err)
	require.NotNil(t, assignment)
	assert.Equal(t, dto.AssignmentID, assignment.AssignmentID)
	assert.Equal(t, dto.BusinessDomainID, assignment.BusinessDomainID)
	assert.Equal(t, dto.CapabilityID, assignment.CapabilityID)
}

func TestDomainCapabilityAssignmentReadModel_GetByDomainAndCapability_NotFound(t *testing.T) {
	f := newAssignmentFixture(t)

	assignment, err := f.readModel.GetByDomainAndCapability(tenantContext(), "bd-nonexistent", "cap-nonexistent")
	require.NoError(t, err)
	assert.Nil(t, assignment)
}

func TestDomainCapabilityAssignmentReadModel_AssignmentExists(t *testing.T) {
	f := newAssignmentFixture(t)
	dto := f.makeDTO("exists")
	f.insertViaSQL(dto)

	ctx := tenantContext()
	exists, err := f.readModel.AssignmentExists(ctx, dto.BusinessDomainID, dto.CapabilityID)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = f.readModel.AssignmentExists(ctx, "bd-nonexistent", dto.CapabilityID)
	require.NoError(t, err)
	assert.False(t, exists)
}
