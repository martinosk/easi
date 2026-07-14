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
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type indexFixture struct {
	db        *sql.DB
	rm        *readmodels.OnePagerSubjectIndexReadModel
	ctx       context.Context
	tenant    string
	t         *testing.T
	baseTime  time.Time
	upsertSeq int
}

func newIndexFixture(t *testing.T) *indexFixture {
	t.Helper()
	connStr := "host=localhost port=5432 user=easi_app password=localdev dbname=easi sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() { _ = db.Close() })

	tenantValue := "test-si-" + uuid.NewString()[:8]
	tenantID, err := sharedvo.NewTenantID(tenantValue)
	require.NoError(t, err)
	ctx := sharedctx.WithTenant(context.Background(), tenantID)

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM onepagers.one_pager_subject_index WHERE tenant_id = $1", tenantValue)
	})

	return &indexFixture{
		db:       db,
		rm:       readmodels.NewOnePagerSubjectIndexReadModel(database.NewTenantAwareDB(db)),
		ctx:      ctx,
		tenant:   tenantValue,
		t:        t,
		baseTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

type seedRow struct {
	subjectType string
	subjectID   string
	name        string
	email       string
	required    int
	filled      int
}

func (f *indexFixture) seed(row seedRow) {
	f.t.Helper()
	f.upsertSeq++
	at := f.baseTime.Add(time.Duration(f.upsertSeq) * time.Hour)
	require.NoError(f.t, f.rm.Upsert(f.ctx, readmodels.SubjectIndexRecord{
		SubjectType:    row.subjectType,
		SubjectID:      row.subjectID,
		Name:           row.name,
		CreatorActorID: "actor-" + row.subjectID,
		CreatorEmail:   row.email,
		CreatedAt:      at,
		LastUpdatedAt:  at,
		RequiredCount:  row.required,
		FilledCount:    row.filled,
	}))
}

func (f *indexFixture) allPages(query readmodels.SubjectIndexQuery) []readmodels.SubjectIndexRecord {
	f.t.Helper()
	var all []readmodels.SubjectIndexRecord
	seen := map[string]bool{}
	for {
		page, hasMore, err := f.rm.Page(f.ctx, query)
		require.NoError(f.t, err)
		for i := range page {
			key := page[i].SubjectType + "/" + page[i].SubjectID
			require.False(f.t, seen[key], "subject %s appeared on two pages", key)
			seen[key] = true
			all = append(all, page[i])
		}
		if !hasMore {
			break
		}
		last := page[len(page)-1]
		query.After = &last
	}
	return all
}

func TestSubjectIndex_CRUDRoundtrip(t *testing.T) {
	f := newIndexFixture(t)
	f.seed(seedRow{subjectType: "application", subjectID: "app-1", name: "Alpha", email: "a@x.com", required: 2, filled: 1})

	require.NoError(t, f.rm.ApplySubjectChange(f.ctx, readmodels.SubjectChange{
		Subject: readmodels.SubjectKey{SubjectType: "application", SubjectID: "app-1"},
		Name:    "Alpha Renamed", Counts: readmodels.CompletenessCounts{Required: 2, Filled: 2},
		OccurredAt: f.baseTime.Add(48 * time.Hour),
	}))
	page, _, err := f.rm.Page(f.ctx, readmodels.SubjectIndexQuery{SubjectTypes: []string{"application"}, Sort: readmodels.SortName, Order: readmodels.OrderAsc, Limit: 10})
	require.NoError(t, err)
	require.Len(t, page, 1)
	assert.Equal(t, "Alpha Renamed", page[0].Name)
	assert.Equal(t, readmodels.SignalComplete, page[0].Signal())

	require.NoError(t, f.rm.ApplyCompleteness(f.ctx,
		readmodels.SubjectKey{SubjectType: "application", SubjectID: "app-1"},
		readmodels.CompletenessCounts{Required: 3, Filled: 1}))
	page, _, _ = f.rm.Page(f.ctx, readmodels.SubjectIndexQuery{SubjectTypes: []string{"application"}, Sort: readmodels.SortName, Limit: 10})
	assert.Equal(t, readmodels.SignalIncomplete, page[0].Signal())

	ids, err := f.rm.SubjectIDs(f.ctx, "application")
	require.NoError(t, err)
	assert.Equal(t, []string{"app-1"}, ids)

	require.NoError(t, f.rm.Delete(f.ctx, readmodels.SubjectKey{SubjectType: "application", SubjectID: "app-1"}))
	page, _, _ = f.rm.Page(f.ctx, readmodels.SubjectIndexQuery{SubjectTypes: []string{"application"}, Sort: readmodels.SortName, Limit: 10})
	assert.Empty(t, page)
}

func TestSubjectIndex_PaginationIsTotalOrderPerSort(t *testing.T) {
	f := newIndexFixture(t)
	names := []string{"Mango", "apple", "Cherry", "banana", "Date", "elderberry", "Fig"}
	for i, name := range names {
		email := fmt.Sprintf("user%d@x.com", (len(names)-i)%3)
		f.seed(seedRow{subjectType: "application", subjectID: fmt.Sprintf("app-%02d", i), name: name, email: email, required: i % 3, filled: i % 2})
	}
	f.seed(seedRow{subjectType: "vendor", subjectID: "ven-1", name: "Zeta", email: "z@x.com", required: 1})

	subjectTypes := []string{"application", "vendor"}
	for _, sort := range []string{readmodels.SortName, readmodels.SortCreator, readmodels.SortCreated, readmodels.SortUpdated, readmodels.SortCompleteness} {
		for _, order := range []string{readmodels.OrderAsc, readmodels.OrderDesc} {
			t.Run(sort+"-"+order, func(t *testing.T) {
				query := readmodels.SubjectIndexQuery{SubjectTypes: subjectTypes, Sort: sort, Order: order, Limit: 2}
				all := f.allPages(query)
				assert.Len(t, all, len(names)+1, "every subject appears exactly once across pages")
			})
		}
	}
}

func TestSubjectIndex_CompletenessOrderingIncompleteFirst(t *testing.T) {
	f := newIndexFixture(t)
	f.seed(seedRow{subjectType: "application", subjectID: "complete-1", name: "C", email: "c@x.com", required: 2, filled: 2})
	f.seed(seedRow{subjectType: "application", subjectID: "na-1", name: "N", email: "n@x.com"})
	f.seed(seedRow{subjectType: "application", subjectID: "incomplete-1", name: "I1", email: "i@x.com", required: 3, filled: 1})
	f.seed(seedRow{subjectType: "application", subjectID: "incomplete-2", name: "I2", email: "j@x.com", required: 4, filled: 1})

	all := f.allPages(readmodels.SubjectIndexQuery{
		SubjectTypes: []string{"application"}, Sort: readmodels.SortCompleteness, Order: readmodels.OrderAsc, Limit: 10,
	})

	require.Len(t, all, 4)
	assert.Equal(t, readmodels.SignalIncomplete, all[0].Signal())
	assert.Equal(t, "incomplete-2", all[0].SubjectID, "more missing fields ranks first")
	assert.Equal(t, readmodels.SignalIncomplete, all[1].Signal())
	assert.Equal(t, readmodels.SignalComplete, all[2].Signal())
	assert.Equal(t, readmodels.SignalNotApplicable, all[3].Signal())
}

func TestSubjectIndex_SubjectTypeFilterScopesResults(t *testing.T) {
	f := newIndexFixture(t)
	f.seed(seedRow{subjectType: "application", subjectID: "app-1", name: "App", email: "a@x.com"})
	f.seed(seedRow{subjectType: "capability", subjectID: "cap-1", name: "Cap", email: "c@x.com"})
	f.seed(seedRow{subjectType: "vendor", subjectID: "ven-1", name: "Ven", email: "v@x.com"})

	page, _, err := f.rm.Page(f.ctx, readmodels.SubjectIndexQuery{
		SubjectTypes: []string{"application", "vendor"}, Sort: readmodels.SortName, Order: readmodels.OrderAsc, Limit: 10,
	})
	require.NoError(t, err)

	ids := map[string]bool{}
	for _, r := range page {
		ids[r.SubjectID] = true
	}
	assert.True(t, ids["app-1"])
	assert.True(t, ids["ven-1"])
	assert.False(t, ids["cap-1"], "capability is filtered out")
}
