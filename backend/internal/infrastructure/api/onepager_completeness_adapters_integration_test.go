//go:build integration
// +build integration

package api

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	archAPI "easi/backend/internal/architecturemodeling/infrastructure/api"
	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	opReadModels "easi/backend/internal/onepagers/application/readmodels"
	opValueObjects "easi/backend/internal/onepagers/domain/valueobjects"
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

type completenessQueryCounter struct {
	count int64
}

func (c *completenessQueryCounter) reset()       { atomic.StoreInt64(&c.count, 0) }
func (c *completenessQueryCounter) value() int64 { return atomic.LoadInt64(&c.count) }
func (c *completenessQueryCounter) increment()   { atomic.AddInt64(&c.count, 1) }

var completenessCounter = &completenessQueryCounter{}

var registerCompletenessDriverOnce sync.Once

func registerCompletenessCountingDriver() {
	registerCompletenessDriverOnce.Do(func() {
		sql.Register("postgres-counting-completeness", completenessCountingDriver{})
	})
}

type completenessCountingDriver struct{}

func (completenessCountingDriver) Open(name string) (driver.Conn, error) {
	conn, err := (&pq.Driver{}).Open(name)
	if err != nil {
		return nil, err
	}
	return completenessCountingConn{Conn: conn}, nil
}

type completenessCountingConn struct {
	driver.Conn
}

func (c completenessCountingConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.Conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return completenessCountingStmt{Stmt: stmt}, nil
}

func (c completenessCountingConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	queryer, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	completenessCounter.increment()
	return queryer.QueryContext(ctx, query, args)
}

func (c completenessCountingConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	execer, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	completenessCounter.increment()
	return execer.ExecContext(ctx, query, args)
}

func (c completenessCountingConn) CheckNamedValue(nv *driver.NamedValue) error {
	checker, ok := c.Conn.(driver.NamedValueChecker)
	if !ok {
		return driver.ErrSkip
	}
	return checker.CheckNamedValue(nv)
}

func (c completenessCountingConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	beginner, ok := c.Conn.(driver.ConnBeginTx)
	if !ok {
		return c.Conn.Begin()
	}
	return beginner.BeginTx(ctx, opts)
}

type completenessCountingStmt struct {
	driver.Stmt
}

func (s completenessCountingStmt) Exec(args []driver.Value) (driver.Result, error) {
	completenessCounter.increment()
	return s.Stmt.Exec(args)
}

func (s completenessCountingStmt) Query(args []driver.Value) (driver.Rows, error) {
	completenessCounter.increment()
	return s.Stmt.Query(args)
}

type completenessAllowAllMiddleware struct{}

func (completenessAllowAllMiddleware) RequirePermission(_ authPL.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return next }
}

func completenessTenantMiddleware(tenantID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tid, _ := sharedvo.NewTenantID(tenantID)
			ctx := sharedctx.WithTenant(r.Context(), tid)
			ctx = sharedctx.WithActor(ctx, sharedctx.NewActor("test-actor", "actor@example.com", sharedctx.RoleAdmin))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type completenessHarness struct {
	db       *sql.DB
	tenantDB *database.TenantAwareDB
	tenantID string
	ctx      context.Context
	wired    chi.Router
	plain    chi.Router
}

func setupCompletenessHarness(t *testing.T) *completenessHarness {
	t.Helper()
	registerCompletenessCountingDriver()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("INTEGRATION_TEST_DB_HOST", "localhost"),
		getEnv("INTEGRATION_TEST_DB_PORT", "5432"),
		getEnv("INTEGRATION_TEST_DB_USER", "easi_app"),
		getEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev"),
		getEnv("INTEGRATION_TEST_DB_NAME", "easi"),
		getEnv("INTEGRATION_TEST_DB_SSLMODE", "disable"),
	)
	db, err := sql.Open("postgres-counting-completeness", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	tenantID := "test-opc-" + uuid.New().String()
	tid, err := sharedvo.NewTenantID(tenantID)
	require.NoError(t, err)

	tenantDB := database.NewTenantAwareDB(db)
	h := &completenessHarness{
		db:       db,
		tenantDB: tenantDB,
		tenantID: tenantID,
		ctx:      sharedctx.WithTenant(context.Background(), tid),
		wired:    buildComponentsRouter(t, tenantDB, tenantID, true),
		plain:    buildComponentsRouter(t, tenantDB, tenantID, false),
	}

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM onepagers.one_pager_facts WHERE tenant_id = $1", tenantID)
		_, _ = db.Exec("DELETE FROM onepagers.one_pager_configurations WHERE tenant_id = $1", tenantID)
		_, _ = db.Exec("DELETE FROM architecturemodeling.application_components WHERE tenant_id = $1", tenantID)
		_ = db.Close()
	})

	return h
}

func buildComponentsRouter(t *testing.T, tenantDB *database.TenantAwareDB, tenantID string, withCompleteness bool) chi.Router {
	t.Helper()
	router := chi.NewRouter()
	router.Use(completenessTenantMiddleware(tenantID))

	cfg := archAPI.RouteConfig{
		Router:         router,
		CommandBus:     cqrs.NewInMemoryCommandBus(),
		EventStore:     eventstore.NewPostgresEventStore(tenantDB),
		EventBus:       events.NewInMemoryEventBus(),
		DB:             tenantDB,
		HATEOAS:        sharedAPI.NewHATEOASLinks("/api/v1"),
		AuthMiddleware: completenessAllowAllMiddleware{},
	}
	if withCompleteness {
		indicators := newOnePagerCompletenessIndicators(tenantDB)
		cfg.OnePagerCompleteness = archAPI.OnePagerCompletenessSources{
			Components:       onePagerCompletenessFor(indicators, "application"),
			AcquiredEntities: onePagerCompletenessFor(indicators, "acquired-entity"),
			Vendors:          onePagerCompletenessFor(indicators, "vendor"),
			InternalTeams:    onePagerCompletenessFor(indicators, "internal-team"),
		}
	}
	require.NoError(t, archAPI.SetupArchitectureModelingRoutes(cfg))
	return router
}

func (h *completenessHarness) seedComponents(t *testing.T, count int) []string {
	t.Helper()
	readModel := archReadModels.NewApplicationComponentReadModel(h.tenantDB)
	ids := make([]string, count)
	for i := 0; i < count; i++ {
		ids[i] = uuid.New().String()
		require.NoError(t, readModel.Insert(h.ctx, archReadModels.ApplicationComponentDTO{
			ID:        ids[i],
			Name:      fmt.Sprintf("OPC Component %c", 'A'+i),
			CreatedAt: time.Now().UTC(),
		}))
	}
	return ids
}

func (h *completenessHarness) seedConfiguration(t *testing.T, required bool) string {
	t.Helper()
	fieldID := uuid.New().String()
	now := time.Now().UTC()
	require.NoError(t, opReadModels.NewOnePagerConfigurationReadModel(h.tenantDB).Insert(h.ctx, opReadModels.ConfigurationRecord{
		ID:          uuid.New().String(),
		TenantID:    h.tenantID,
		SubjectType: "application",
		Document: opReadModels.ConfigurationDocument{
			CustomFields: []opReadModels.CustomFieldRecord{
				{ID: fieldID, Name: "Contract link", Type: "text", Required: required, Active: true},
			},
			DisplayOrder: []opReadModels.FieldRefRecord{{Kind: "custom", ID: fieldID}},
		},
		Version:    1,
		CreatedAt:  now,
		ModifiedAt: now,
		ModifiedBy: "test@example.com",
	}))
	return fieldID
}

func (h *completenessHarness) seedFacts(t *testing.T, fieldID string, subjectIDs ...string) {
	t.Helper()
	factsModel := opReadModels.NewOnePagerFactsReadModel(h.tenantDB)
	now := time.Now().UTC()
	for _, subjectID := range subjectIDs {
		value, err := opValueObjects.NewTextValue("filled value")
		require.NoError(t, err)
		envelope, err := opValueObjects.NewValueEnvelope(value)
		require.NoError(t, err)
		require.NoError(t, factsModel.Upsert(h.ctx, opReadModels.FactRecord{
			FactsID:     uuid.New().String(),
			TenantID:    h.tenantID,
			SubjectType: "application",
			SubjectID:   subjectID,
			FieldID:     fieldID,
			Value:       &envelope,
			ValueType:   "text",
			DisplayText: "filled value",
			ModifiedAt:  now,
			ModifiedBy:  "test@example.com",
		}))
	}
}

type componentsPage struct {
	rows       []map[string]any
	pagination map[string]any
	raw        string
}

func fetchComponentsPage(t *testing.T, router chi.Router, path string) componentsPage {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var response struct {
		Data       []map[string]any `json:"data"`
		Pagination map[string]any   `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	return componentsPage{rows: response.Data, pagination: response.Pagination, raw: rec.Body.String()}
}

func pageIDs(page componentsPage) []string {
	ids := make([]string, len(page.rows))
	for i, row := range page.rows {
		ids[i] = row["id"].(string)
	}
	return ids
}

func pageIndicators(page componentsPage) []any {
	indicators := make([]any, len(page.rows))
	for i, row := range page.rows {
		indicators[i] = row["onePagerComplete"]
	}
	return indicators
}

func nextPagePath(t *testing.T, page componentsPage) string {
	t.Helper()
	cursor, ok := page.pagination["cursor"].(string)
	require.True(t, ok, "expected a next-page cursor")
	require.NotEmpty(t, cursor)
	return "/components?limit=2&after=" + url.QueryEscape(cursor)
}

func measuredQueryCount(t *testing.T, router chi.Router, path string) (int64, componentsPage) {
	t.Helper()
	completenessCounter.reset()
	page := fetchComponentsPage(t, router, path)
	return completenessCounter.value(), page
}

func TestComponentsList_CompletenessIndicatorsAcrossPages_Integration(t *testing.T) {
	h := setupCompletenessHarness(t)
	ids := h.seedComponents(t, 5)
	fieldID := h.seedConfiguration(t, true)
	h.seedFacts(t, fieldID, ids[0], ids[2], ids[4])

	wiredPage1 := fetchComponentsPage(t, h.wired, "/components?limit=2")
	require.Equal(t, []string{ids[0], ids[1]}, pageIDs(wiredPage1))
	assert.Equal(t, []any{true, false}, pageIndicators(wiredPage1))
	assert.Equal(t, true, wiredPage1.pagination["hasMore"])

	wiredPage2 := fetchComponentsPage(t, h.wired, nextPagePath(t, wiredPage1))
	require.Equal(t, []string{ids[2], ids[3]}, pageIDs(wiredPage2))
	assert.Equal(t, []any{true, false}, pageIndicators(wiredPage2))

	plainPage1 := fetchComponentsPage(t, h.plain, "/components?limit=2")
	plainPage2 := fetchComponentsPage(t, h.plain, nextPagePath(t, plainPage1))

	assert.Equal(t, pageIDs(plainPage1), pageIDs(wiredPage1), "row ordering must be unchanged by the indicator")
	assert.Equal(t, pageIDs(plainPage2), pageIDs(wiredPage2), "row ordering must be unchanged by the indicator")
	assert.Equal(t, plainPage1.pagination, wiredPage1.pagination, "pagination envelope must be unchanged by the indicator")
	assert.Equal(t, plainPage2.pagination, wiredPage2.pagination, "pagination envelope must be unchanged by the indicator")
	assert.NotContains(t, plainPage1.raw, "onePagerComplete")
	assert.NotContains(t, plainPage2.raw, "onePagerComplete")
}

func TestComponentsList_ConstantQueryCountPerPage_Integration(t *testing.T) {
	h := setupCompletenessHarness(t)
	ids := h.seedComponents(t, 5)
	fieldID := h.seedConfiguration(t, true)
	h.seedFacts(t, fieldID, ids[0], ids[2], ids[4])

	countTwoRows, pageTwo := measuredQueryCount(t, h.wired, "/components?limit=2")
	countFourRows, pageFour := measuredQueryCount(t, h.wired, "/components?limit=4")

	require.Len(t, pageTwo.rows, 2)
	require.Len(t, pageFour.rows, 4)
	assert.Equal(t, countTwoRows, countFourRows, "query count per page must not grow with the number of rows")
}

func TestComponentsList_NoActiveRequiredField_ByteIdenticalAndFactsQuerySkipped_Integration(t *testing.T) {
	requiredHarness := setupCompletenessHarness(t)
	requiredIDs := requiredHarness.seedComponents(t, 3)
	requiredField := requiredHarness.seedConfiguration(t, true)
	requiredHarness.seedFacts(t, requiredField, requiredIDs[0])

	optionalHarness := setupCompletenessHarness(t)
	optionalHarness.seedComponents(t, 3)
	optionalHarness.seedConfiguration(t, false)

	requiredWiredCount, _ := measuredQueryCount(t, requiredHarness.wired, "/components?limit=10")
	requiredPlainCount, _ := measuredQueryCount(t, requiredHarness.plain, "/components?limit=10")
	optionalWiredCount, optionalWiredPage := measuredQueryCount(t, optionalHarness.wired, "/components?limit=10")
	optionalPlainCount, optionalPlainPage := measuredQueryCount(t, optionalHarness.plain, "/components?limit=10")

	assert.Equal(t, optionalPlainPage.raw, optionalWiredPage.raw, "response must be byte-identical when no active required field exists")
	assert.NotContains(t, optionalWiredPage.raw, "onePagerComplete")
	assert.Less(t, optionalWiredCount-optionalPlainCount, requiredWiredCount-requiredPlainCount,
		"the facts query must be skipped when the configuration has no active required field")
}
