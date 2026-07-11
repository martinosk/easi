//go:build integration
// +build integration

package readmodels_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/valueobjects"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	subjectTypeApplication = "application"
	subjectTypeVendor      = "vendor"
)

type completenessTestFixture struct {
	db        *sql.DB
	readModel *readmodels.OnePagerFactsReadModel
	ctx       context.Context
	t         *testing.T
}

func applicationKey(subjectID string) readmodels.SubjectKey {
	return readmodels.SubjectKey{SubjectType: subjectTypeApplication, SubjectID: subjectID}
}

type filledCountsQuery struct {
	subjectIDs []string
	fieldIDs   []string
}

func newCompletenessTestFixture(t *testing.T) *completenessTestFixture {
	t.Helper()
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		"localhost", "5432", "easi_app", "localdev", "easi", "disable")
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() { _ = db.Close() })

	tenantDB := database.NewTenantAwareDB(db)
	ctx := sharedctx.WithTenant(context.Background(), sharedvo.DefaultTenantID())

	return &completenessTestFixture{
		db:        db,
		readModel: readmodels.NewOnePagerFactsReadModel(tenantDB),
		ctx:       ctx,
		t:         t,
	}
}

func (f *completenessTestFixture) putFact(subject readmodels.SubjectKey, fieldID string) {
	f.t.Helper()
	value, err := valueobjects.NewTextValue("value for " + fieldID)
	require.NoError(f.t, err)
	envelope, err := valueobjects.NewValueEnvelope(value)
	require.NoError(f.t, err)

	require.NoError(f.t, f.readModel.Upsert(f.ctx, readmodels.FactRecord{
		FactsID:     uuid.New().String(),
		TenantID:    sharedvo.DefaultTenantID().Value(),
		SubjectType: subject.SubjectType,
		SubjectID:   subject.SubjectID,
		FieldID:     fieldID,
		Value:       &envelope,
		ValueType:   "text",
		DisplayText: "value for " + fieldID,
		ModifiedAt:  time.Now().UTC(),
		ModifiedBy:  "test@example.com",
	}))
	f.t.Cleanup(func() {
		_, _ = f.db.Exec(
			"DELETE FROM onepagers.one_pager_facts WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = $3 AND field_id = $4",
			sharedvo.DefaultTenantID().Value(), subject.SubjectType, subject.SubjectID, fieldID,
		)
	})
}

func (f *completenessTestFixture) clearFact(subject readmodels.SubjectKey, fieldID string) {
	f.t.Helper()
	require.NoError(f.t, f.readModel.Clear(f.ctx, readmodels.ClearFactParams{
		SubjectType: subject.SubjectType,
		SubjectID:   subject.SubjectID,
		FieldID:     fieldID,
		ModifiedAt:  time.Now().UTC(),
		ModifiedBy:  "test@example.com",
	}))
}

func (f *completenessTestFixture) assertFilledFieldCounts(query filledCountsQuery, want map[string]int) {
	f.t.Helper()
	counts, err := f.readModel.FilledFieldCounts(f.ctx, subjectTypeApplication, query.subjectIDs, query.fieldIDs)
	require.NoError(f.t, err)
	assert.Equal(f.t, want, counts)
}

func (f *completenessTestFixture) assertSubjectsWithValue(fieldID string, want int) {
	f.t.Helper()
	count, err := f.readModel.CountSubjectsWithValue(f.ctx, subjectTypeApplication, fieldID)
	require.NoError(f.t, err)
	assert.Equal(f.t, want, count)
}

func TestOnePagerFactsReadModel_FilledFieldCounts_ScopedCounting(t *testing.T) {
	cases := []struct {
		name  string
		setup func(f *completenessTestFixture) (filledCountsQuery, map[string]int)
	}{
		{
			name: "counts only non-null values",
			setup: func(f *completenessTestFixture) (filledCountsQuery, map[string]int) {
				subjectA := uuid.New().String()
				fieldOne := uuid.New().String()
				fieldTwo := uuid.New().String()

				f.putFact(applicationKey(subjectA), fieldOne)
				f.putFact(applicationKey(subjectA), fieldTwo)
				f.clearFact(applicationKey(subjectA), fieldTwo)

				return filledCountsQuery{subjectIDs: []string{subjectA}, fieldIDs: []string{fieldOne, fieldTwo}}, map[string]int{subjectA: 1}
			},
		},
		{
			name: "scopes by field list",
			setup: func(f *completenessTestFixture) (filledCountsQuery, map[string]int) {
				subjectA := uuid.New().String()
				trackedField := uuid.New().String()
				untrackedField := uuid.New().String()

				f.putFact(applicationKey(subjectA), trackedField)
				f.putFact(applicationKey(subjectA), untrackedField)

				return filledCountsQuery{subjectIDs: []string{subjectA}, fieldIDs: []string{trackedField}}, map[string]int{subjectA: 1}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newCompletenessTestFixture(t)
			query, want := tc.setup(f)
			f.assertFilledFieldCounts(query, want)
		})
	}
}

func TestOnePagerFactsReadModel_FilledFieldCounts_ScopesBySubjectTypeAndSubjectIDs(t *testing.T) {
	f := newCompletenessTestFixture(t)
	subjectA := uuid.New().String()
	subjectB := uuid.New().String()
	subjectNoFacts := uuid.New().String()
	vendorSubject := uuid.New().String()
	fieldOne := uuid.New().String()
	fieldTwo := uuid.New().String()

	f.putFact(applicationKey(subjectA), fieldOne)
	f.putFact(applicationKey(subjectA), fieldTwo)
	f.putFact(applicationKey(subjectB), fieldOne)
	f.putFact(readmodels.SubjectKey{SubjectType: subjectTypeVendor, SubjectID: vendorSubject}, fieldOne)

	f.assertFilledFieldCounts(
		filledCountsQuery{
			subjectIDs: []string{subjectA, subjectB, subjectNoFacts, vendorSubject},
			fieldIDs:   []string{fieldOne, fieldTwo},
		},
		map[string]int{subjectA: 2, subjectB: 1})
}

func TestOnePagerFactsReadModel_FilledFieldCounts_EmptyInputsReturnEmptyMapWithoutQuerying(t *testing.T) {
	readModel := readmodels.NewOnePagerFactsReadModel(nil)

	counts, err := readModel.FilledFieldCounts(context.Background(), subjectTypeApplication, nil, []string{"field-1"})
	require.NoError(t, err)
	assert.Empty(t, counts)

	counts, err = readModel.FilledFieldCounts(context.Background(), subjectTypeApplication, []string{"subject-1"}, nil)
	require.NoError(t, err)
	assert.Empty(t, counts)
}

func TestOnePagerFactsReadModel_CountSubjectsWithValue(t *testing.T) {
	cases := []struct {
		name  string
		setup func(f *completenessTestFixture, fieldID string)
		want  int
	}{
		{
			name: "counts distinct subjects with non-null value",
			setup: func(f *completenessTestFixture, fieldID string) {
				f.putFact(applicationKey(uuid.New().String()), fieldID)
				f.putFact(applicationKey(uuid.New().String()), fieldID)
			},
			want: 2,
		},
		{
			name: "ignores other fields and subject types",
			setup: func(f *completenessTestFixture, fieldID string) {
				subjectA := uuid.New().String()
				f.putFact(applicationKey(subjectA), fieldID)
				f.putFact(applicationKey(subjectA), uuid.New().String())
				f.putFact(readmodels.SubjectKey{SubjectType: subjectTypeVendor, SubjectID: uuid.New().String()}, fieldID)
			},
			want: 1,
		},
		{
			name: "excludes cleared values",
			setup: func(f *completenessTestFixture, fieldID string) {
				subjectA := uuid.New().String()
				subjectB := uuid.New().String()
				f.putFact(applicationKey(subjectA), fieldID)
				f.putFact(applicationKey(subjectB), fieldID)
				f.clearFact(applicationKey(subjectB), fieldID)
			},
			want: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newCompletenessTestFixture(t)
			fieldID := uuid.New().String()
			tc.setup(f, fieldID)
			f.assertSubjectsWithValue(fieldID, tc.want)
		})
	}
}
