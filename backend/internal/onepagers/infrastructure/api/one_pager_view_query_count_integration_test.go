//go:build integration
// +build integration

package api

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type queryCounter struct {
	count int64
}

func (c *queryCounter) reset()       { atomic.StoreInt64(&c.count, 0) }
func (c *queryCounter) value() int64 { return atomic.LoadInt64(&c.count) }
func (c *queryCounter) increment()   { atomic.AddInt64(&c.count, 1) }

var testQueryCounter = &queryCounter{}

var registerCountingDriverOnce sync.Once

func registerCountingDriver() {
	registerCountingDriverOnce.Do(func() {
		sql.Register("postgres-counting", countingDriver{})
	})
}

type countingDriver struct{}

func (countingDriver) Open(name string) (driver.Conn, error) {
	conn, err := (&pq.Driver{}).Open(name)
	if err != nil {
		return nil, err
	}
	return countingConn{Conn: conn}, nil
}

type countingConn struct {
	driver.Conn
}

func (c countingConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.Conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return countingStmt{Stmt: stmt}, nil
}

func (c countingConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	queryer, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	testQueryCounter.increment()
	return queryer.QueryContext(ctx, query, args)
}

func (c countingConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	execer, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	testQueryCounter.increment()
	return execer.ExecContext(ctx, query, args)
}

func (c countingConn) CheckNamedValue(nv *driver.NamedValue) error {
	checker, ok := c.Conn.(driver.NamedValueChecker)
	if !ok {
		return driver.ErrSkip
	}
	return checker.CheckNamedValue(nv)
}

func (c countingConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	beginner, ok := c.Conn.(driver.ConnBeginTx)
	if !ok {
		return c.Conn.Begin()
	}
	return beginner.BeginTx(ctx, opts)
}

type countingStmt struct {
	driver.Stmt
}

func (s countingStmt) Exec(args []driver.Value) (driver.Result, error) {
	testQueryCounter.increment()
	return s.Stmt.Exec(args)
}

func (s countingStmt) Query(args []driver.Value) (driver.Rows, error) {
	testQueryCounter.increment()
	return s.Stmt.Query(args)
}

type queryCountContext struct {
	db       *sql.DB
	tenantDB *database.TenantAwareDB
	tenantID string
	router   chi.Router
}

func queryCountTenantMiddleware(tenantID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tid, _ := sharedvo.NewTenantID(tenantID)
			ctx := sharedctx.WithTenant(r.Context(), tid)
			ctx = sharedctx.WithActor(ctx, sharedctx.NewActor(uuid.New().String(), "actor@example.com", sharedctx.RoleAdmin))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func setupQueryCountRouter(t *testing.T, tenantID string) *queryCountContext {
	t.Helper()
	registerCountingDriver()

	dbHost := getEnv("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getEnv("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getEnv("INTEGRATION_TEST_DB_USER", "easi_app")
	dbPassword := getEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev")
	dbName := getEnv("INTEGRATION_TEST_DB_NAME", "easi")
	dbSSLMode := getEnv("INTEGRATION_TEST_DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	db, err := sql.Open("postgres-counting", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	router := chi.NewRouter()
	router.Use(queryCountTenantMiddleware(tenantID))

	require.NoError(t, SetupOnePagersRoutes(OnePagersRoutesDeps{
		Router:          router,
		CommandBus:      cqrs.NewInMemoryCommandBus(),
		EventStore:      eventStore,
		EventBus:        eventBus,
		DB:              tenantDB,
		Hateoas:         sharedAPI.NewHATEOASLinks("/api/v1"),
		AuthMiddleware:  allowAllMiddleware{},
		SessionProvider: &fakeSessionProvider{email: "actor@example.com"},
	}))

	t.Cleanup(func() {
		for _, table := range []string{
			"onepagers.one_pager_facts",
			"onepagers.one_pager_configurations",
			"onepagers.subject_relation_cache",
			"onepagers.one_pager_subject_index",
			"onepagers.maturity_scale_cache",
		} {
			_, _ = db.Exec("DELETE FROM "+table+" WHERE tenant_id = $1", tenantID)
		}
		_ = db.Close()
	})

	return &queryCountContext{db: db, tenantDB: tenantDB, tenantID: tenantID, router: router}
}

func (qc *queryCountContext) get(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	qc.router.ServeHTTP(rec, req)
	return rec
}

func (qc *queryCountContext) tenantContext(t *testing.T) context.Context {
	t.Helper()
	tid, err := sharedvo.NewTenantID(qc.tenantID)
	require.NoError(t, err)
	return sharedctx.WithTenant(context.Background(), tid)
}

func (qc *queryCountContext) seedSubject(t *testing.T, subject readmodels.SubjectKey, name string, values map[string]any) {
	t.Helper()
	attributes := readmodels.SubjectAttributes{}
	for key, value := range values {
		require.NoError(t, attributes.Set(key, value))
	}
	require.NoError(t, readmodels.NewOnePagerSubjectIndexReadModel(qc.tenantDB).Upsert(qc.tenantContext(t), readmodels.SubjectIndexRecord{
		SubjectType: subject.SubjectType, SubjectID: subject.SubjectID, Name: name,
		CreatedAt: time.Now().UTC(), LastUpdatedAt: time.Now().UTC(), Attributes: attributes,
	}))
}

func (qc *queryCountContext) seedConfiguration(t *testing.T, document readmodels.ConfigurationDocument) {
	t.Helper()
	now := time.Now().UTC()
	require.NoError(t, readmodels.NewOnePagerConfigurationReadModel(qc.tenantDB).Insert(qc.tenantContext(t), readmodels.ConfigurationRecord{
		ID:          uuid.New().String(),
		TenantID:    qc.tenantID,
		SubjectType: "capability",
		Document:    document,
		Version:     1,
		CreatedAt:   now,
		ModifiedAt:  now,
		ModifiedBy:  "test@example.com",
	}))
}

func seedOnePagerQueryCountScenario(t *testing.T, qc *queryCountContext, fieldCount int) string {
	t.Helper()
	ctx := qc.tenantContext(t)

	capabilityID := uuid.New().String()
	qc.seedSubject(t, readmodels.SubjectKey{SubjectType: "capability", SubjectID: capabilityID}, "Query Count Capability", map[string]any{
		"description":   "seeded for the constant-query-count contract test",
		"maturityValue": 42,
	})

	customFields := make([]readmodels.CustomFieldRecord, fieldCount)
	displayOrder := []readmodels.FieldRefRecord{{Kind: "builtIn", ID: "maturity"}}
	for i := 0; i < fieldCount; i++ {
		fieldID := uuid.New().String()
		customFields[i] = readmodels.CustomFieldRecord{ID: fieldID, Name: fmt.Sprintf("Field %d", i), Type: "text", Active: true}
		displayOrder = append(displayOrder, readmodels.FieldRefRecord{Kind: "custom", ID: fieldID})
	}
	qc.seedConfiguration(t, readmodels.ConfigurationDocument{CustomFields: customFields, DisplayOrder: displayOrder})

	factsModel := readmodels.NewOnePagerFactsReadModel(qc.tenantDB)
	now := time.Now().UTC()
	for _, field := range customFields {
		value, err := valueobjects.NewTextValue("value for " + field.Name)
		require.NoError(t, err)
		envelope, err := valueobjects.NewValueEnvelope(value)
		require.NoError(t, err)
		require.NoError(t, factsModel.Upsert(ctx, readmodels.FactRecord{
			FactsID:     uuid.New().String(),
			TenantID:    qc.tenantID,
			SubjectType: "capability",
			SubjectID:   capabilityID,
			FieldID:     field.ID,
			Value:       &envelope,
			ValueType:   "text",
			DisplayText: "value for " + field.Name,
			ModifiedAt:  now,
			ModifiedBy:  "test@example.com",
		}))
	}

	return capabilityID
}

func runOnePagerQueryCountScenario(t *testing.T, fieldCount int) int64 {
	t.Helper()
	qc := setupQueryCountRouter(t, "test-qc-"+uuid.New().String())
	capabilityID := seedOnePagerQueryCountScenario(t, qc, fieldCount)

	testQueryCounter.reset()
	rec := qc.get("/one-pagers/capability/" + capabilityID)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	return testQueryCounter.value()
}

func TestOnePagerView_ConstantQueryCount_Integration(t *testing.T) {
	minimalCount := runOnePagerQueryCountScenario(t, 2)
	manyFieldsCount := runOnePagerQueryCountScenario(t, 8)

	assert.Equal(t, minimalCount, manyFieldsCount, "query count must not grow with the number of configured fields")
	assert.LessOrEqual(t, minimalCount, int64(15), "query count must stay within the documented sanity ceiling")
}

var relationBuiltInEntryIDs = []string{"maturity", "realizing-applications", "child-capabilities", "depends-on"}

func seedOnePagerRelationScenario(t *testing.T, qc *queryCountContext, relatedCount int) string {
	t.Helper()
	capabilityID := uuid.New().String()
	qc.seedSubject(t, readmodels.SubjectKey{SubjectType: "capability", SubjectID: capabilityID}, "Relation Capability", nil)
	seedCapabilityRelations(t, qc, capabilityID, relatedCount)

	displayOrder := make([]readmodels.FieldRefRecord, len(relationBuiltInEntryIDs))
	for i, entryID := range relationBuiltInEntryIDs {
		displayOrder[i] = readmodels.FieldRefRecord{Kind: "builtIn", ID: entryID}
	}
	qc.seedConfiguration(t, readmodels.ConfigurationDocument{DisplayOrder: displayOrder})
	return capabilityID
}

func seedCapabilityRelations(t *testing.T, qc *queryCountContext, capabilityID string, relatedCount int) {
	t.Helper()
	ctx := qc.tenantContext(t)
	relations := readmodels.NewSubjectRelationCacheReadModel(qc.tenantDB)
	subject := readmodels.SubjectKey{SubjectType: "capability", SubjectID: capabilityID}

	for i := 0; i < relatedCount; i++ {
		childID := uuid.New().String()
		qc.seedSubject(t, readmodels.SubjectKey{SubjectType: "capability", SubjectID: childID}, fmt.Sprintf("Child %d", i), nil)
		require.NoError(t, relations.Save(ctx, subject, readmodels.RelationEntry{
			EntryID: "child-capabilities", RelatedType: "capability", RelatedID: childID,
		}))

		componentID := uuid.New().String()
		qc.seedSubject(t, readmodels.SubjectKey{SubjectType: "application", SubjectID: componentID}, fmt.Sprintf("App %d", i), nil)
		require.NoError(t, relations.Save(ctx, subject, readmodels.RelationEntry{
			EntryID: "realizing-applications", RelatedType: "application", RelatedID: componentID, EdgeID: uuid.New().String(),
		}))

		targetID := uuid.New().String()
		qc.seedSubject(t, readmodels.SubjectKey{SubjectType: "capability", SubjectID: targetID}, fmt.Sprintf("Target %d", i), nil)
		require.NoError(t, relations.Save(ctx, subject, readmodels.RelationEntry{
			EntryID: "depends-on", RelatedType: "capability", RelatedID: targetID, EdgeID: uuid.New().String(),
		}))
	}
}

func runOnePagerRelationQueryCountScenario(t *testing.T, relatedCount int) int64 {
	t.Helper()
	qc := setupQueryCountRouter(t, "test-qc-"+uuid.New().String())
	capabilityID := seedOnePagerRelationScenario(t, qc, relatedCount)

	testQueryCounter.reset()
	rec := qc.get("/one-pagers/capability/" + capabilityID)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	return testQueryCounter.value()
}

func TestOnePagerView_ConstantQueryCount_WithManyRelations_Integration(t *testing.T) {
	fewRelatedCount := runOnePagerRelationQueryCountScenario(t, 1)
	manyRelatedCount := runOnePagerRelationQueryCountScenario(t, 6)

	assert.Equal(t, fewRelatedCount, manyRelatedCount, "query count must not grow with the number of related entities per relation")
	assert.LessOrEqual(t, fewRelatedCount, int64(12), "every relation of a subject resolves in a single cache query")
}
