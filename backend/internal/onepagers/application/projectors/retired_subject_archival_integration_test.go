//go:build integration

package projectors_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/application/projectors"
	"easi/backend/internal/onepagers/application/readmodels"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubTenantDirectory struct {
	ids []string
}

func (s stubTenantDirectory) TenantIDs(_ context.Context) ([]string, error) {
	return s.ids, nil
}

type capturingDispatcher struct {
	factsIDs []string
	tenants  []string
}

func (d *capturingDispatcher) Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	archive, ok := cmd.(*commands.ArchiveOnePagerFacts)
	if ok {
		tenant, err := sharedctx.GetTenant(ctx)
		if err != nil {
			return cqrs.EmptyResult(), err
		}
		d.factsIDs = append(d.factsIDs, archive.FactsID)
		d.tenants = append(d.tenants, tenant.Value())
	}
	return cqrs.EmptyResult(), nil
}

func tenantContext(t *testing.T, tenant string) context.Context {
	t.Helper()
	tenantID, err := sharedvo.NewTenantID(tenant)
	require.NoError(t, err)
	return sharedctx.WithTenant(context.Background(), tenantID)
}

type retiredSubjectStore struct {
	t     *testing.T
	index *readmodels.OnePagerSubjectIndexReadModel
	facts *readmodels.OnePagerFactsReadModel
}

func (s retiredSubjectStore) seed(tenant, subjectID, factsID string) {
	s.t.Helper()
	ctx := tenantContext(s.t, tenant)
	require.NoError(s.t, s.index.Upsert(ctx, readmodels.SubjectIndexRecord{
		SubjectType: "enterprise-capability", SubjectID: subjectID, Name: "EC " + subjectID,
		CreatedAt: time.Now(), LastUpdatedAt: time.Now(),
	}))
	if factsID == "" {
		return
	}
	require.NoError(s.t, s.facts.Upsert(ctx, readmodels.FactRecord{
		FactsID: factsID, TenantID: tenant, SubjectType: "enterprise-capability", SubjectID: subjectID,
		FieldID: "f-1", ValueType: "text", DisplayText: "v", ModifiedAt: time.Now(), ModifiedBy: "test@dfds.com",
	}))
}

func TestRetiredSubjectArchival_WalksRLSGuardedReadModelsPerTenant(t *testing.T) {
	connStr := "host=localhost port=5432 user=easi_app password=localdev dbname=easi sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() { _ = db.Close() })

	tenantA := "test-rsa-" + uuid.NewString()[:8]
	tenantB := "test-rsa-" + uuid.NewString()[:8]
	t.Cleanup(func() {
		for _, tenant := range []string{tenantA, tenantB} {
			_, _ = db.Exec("SET app.current_tenant = '" + tenant + "'")
			_, _ = db.Exec("DELETE FROM onepagers.one_pager_subject_index WHERE tenant_id = $1", tenant)
			_, _ = db.Exec("DELETE FROM onepagers.one_pager_facts WHERE tenant_id = $1", tenant)
		}
	})

	tenantDB := database.NewTenantAwareDB(db)
	store := retiredSubjectStore{
		t:     t,
		index: readmodels.NewOnePagerSubjectIndexReadModel(tenantDB),
		facts: readmodels.NewOnePagerFactsReadModel(tenantDB),
	}

	factsA := uuid.NewString()
	factsB := uuid.NewString()
	store.seed(tenantA, "ec-with-facts-a", factsA)
	store.seed(tenantA, "ec-without-facts", "")
	store.seed(tenantB, "ec-with-facts-b", factsB)

	dispatcher := &capturingDispatcher{}
	archival := projectors.NewRetiredSubjectArchival(stubTenantDirectory{ids: []string{tenantA, tenantB}}, store.index, store.facts, dispatcher)

	require.NoError(t, archival.Run(context.Background()))

	assert.ElementsMatch(t, []string{factsA, factsB}, dispatcher.factsIDs, "one archive per subject that holds facts")
	assert.ElementsMatch(t, []string{tenantA, tenantB}, dispatcher.tenants, "each archive runs in its own tenant's context")
}
