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

	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	metaReadModels "easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/onepagers/application/ports"
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

type capabilityBuiltInSource struct {
	readModel *capReadModels.CapabilityReadModel
}

func (s capabilityBuiltInSource) FetchSubject(ctx context.Context, subjectID string) (*ports.SubjectSnapshot, error) {
	dto, err := s.readModel.GetByID(ctx, subjectID)
	if err != nil {
		return nil, err
	}
	if dto == nil {
		return nil, nil
	}
	fields := map[string]ports.BuiltInFieldValue{
		"name":     ports.TextValue{Text: dto.Name},
		"maturity": ports.MaturityValue{Value: dto.MaturityValue},
	}
	if dto.Description != "" {
		fields["description"] = ports.TextValue{Text: dto.Description}
	}
	return &ports.SubjectSnapshot{Name: dto.Name, Fields: fields}, nil
}

func (s capabilityBuiltInSource) CountSubjects(ctx context.Context) (int, error) {
	return s.readModel.Count(ctx)
}

type queryCountMaturityScaleAdapter struct {
	configurations *metaReadModels.MetaModelConfigurationReadModel
}

func (a queryCountMaturityScaleAdapter) Sections(ctx context.Context) ([]ports.MaturitySection, error) {
	config, err := a.configurations.GetByTenantID(ctx)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}
	sections := make([]ports.MaturitySection, len(config.Sections))
	for i, section := range config.Sections {
		sections[i] = ports.MaturitySection{Name: section.Name, MinValue: section.MinValue, MaxValue: section.MaxValue}
	}
	return sections, nil
}

type stubSubjectExistenceChecker struct{}

func (stubSubjectExistenceChecker) SubjectExists(_ context.Context, _, _ string) (bool, error) {
	return true, nil
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
	commandBus := cqrs.NewInMemoryCommandBus()

	router := chi.NewRouter()
	router.Use(queryCountTenantMiddleware(tenantID))

	require.NoError(t, SetupOnePagersRoutes(OnePagersRoutesDeps{
		Router:          router,
		CommandBus:      commandBus,
		EventStore:      eventStore,
		EventBus:        eventBus,
		DB:              tenantDB,
		Hateoas:         sharedAPI.NewHATEOASLinks("/api/v1"),
		AuthMiddleware:  allowAllMiddleware{},
		SessionProvider: &fakeSessionProvider{email: "actor@example.com"},
		Subjects:        stubSubjectExistenceChecker{},
		BuiltInFields: map[string]ports.BuiltInFieldSource{
			"capability": capabilityBuiltInSource{readModel: capReadModels.NewCapabilityReadModel(tenantDB)},
		},
		MaturityScale: queryCountMaturityScaleAdapter{configurations: metaReadModels.NewMetaModelConfigurationReadModel(tenantDB)},
	}))

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM onepagers.one_pager_facts WHERE tenant_id = $1", tenantID)
		_, _ = db.Exec("DELETE FROM onepagers.one_pager_configurations WHERE tenant_id = $1", tenantID)
		_, _ = db.Exec("DELETE FROM capabilitymapping.capabilities WHERE tenant_id = $1", tenantID)
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

func seedOnePagerQueryCountScenario(t *testing.T, qc *queryCountContext, fieldCount int) string {
	t.Helper()
	tid, err := sharedvo.NewTenantID(qc.tenantID)
	require.NoError(t, err)
	ctx := sharedctx.WithTenant(context.Background(), tid)

	capabilityID := uuid.New().String()
	capModel := capReadModels.NewCapabilityReadModel(qc.tenantDB)
	require.NoError(t, capModel.Insert(ctx, capReadModels.CapabilityDTO{
		ID:          capabilityID,
		Name:        "Query Count Capability",
		Description: "seeded for the constant-query-count contract test",
		Level:       "L1",
		CreatedAt:   time.Now().UTC(),
	}))

	customFields := make([]readmodels.CustomFieldRecord, fieldCount)
	displayOrder := make([]readmodels.FieldRefRecord, 0, fieldCount+1)
	displayOrder = append(displayOrder, readmodels.FieldRefRecord{Kind: "builtIn", ID: "maturity"})
	for i := 0; i < fieldCount; i++ {
		fieldID := uuid.New().String()
		customFields[i] = readmodels.CustomFieldRecord{ID: fieldID, Name: fmt.Sprintf("Field %d", i), Type: "text", Active: true}
		displayOrder = append(displayOrder, readmodels.FieldRefRecord{Kind: "custom", ID: fieldID})
	}

	configModel := readmodels.NewOnePagerConfigurationReadModel(qc.tenantDB)
	now := time.Now().UTC()
	require.NoError(t, configModel.Insert(ctx, readmodels.ConfigurationRecord{
		ID:          uuid.New().String(),
		TenantID:    qc.tenantID,
		SubjectType: "capability",
		Document: readmodels.ConfigurationDocument{
			CustomFields: customFields,
			DisplayOrder: displayOrder,
		},
		Version:    1,
		CreatedAt:  now,
		ModifiedAt: now,
		ModifiedBy: "test@example.com",
	}))

	factsModel := readmodels.NewOnePagerFactsReadModel(qc.tenantDB)
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
	tenantID := "test-qc-" + uuid.New().String()
	qc := setupQueryCountRouter(t, tenantID)
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
