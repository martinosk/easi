//go:build integration
// +build integration

package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"easi/backend/internal/auth/infrastructure/session"
	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type integrationContext struct {
	db         *sql.DB
	tenantID   string
	router     chi.Router
	cookies    []*http.Cookie
	createdIDs []string
}

func setupIntegration(t *testing.T) *integrationContext {
	t.Helper()

	dbHost := getEnv("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getEnv("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getEnv("INTEGRATION_TEST_DB_USER", "easi_app")
	dbPassword := getEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev")
	dbName := getEnv("INTEGRATION_TEST_DB_NAME", "easi")
	dbSSLMode := getEnv("INTEGRATION_TEST_DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	ic := &integrationContext{
		db:       db,
		tenantID: fmt.Sprintf("test-%s", uuid.New().String()),
	}
	t.Cleanup(func() {
		for _, id := range ic.createdIDs {
			db.Exec("DELETE FROM onepagers.one_pager_configurations WHERE id = $1", id)
			db.Exec("DELETE FROM infrastructure.events WHERE aggregate_id = $1", id)
		}
		db.Exec("DELETE FROM onepagers.one_pager_configurations WHERE tenant_id = $1", ic.tenantID)
		db.Close()
	})

	scsManager := scs.New()
	scsManager.Store = memstore.New()
	scsManager.Lifetime = time.Hour
	sessionManager := session.NewSessionManager(scsManager)

	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)
	commandBus := cqrs.NewInMemoryCommandBus()

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Use(ic.tenantAndActorMiddleware())

	require.NoError(t, SetupOnePagersRoutes(OnePagersRoutesDeps{
		Router:          router,
		CommandBus:      commandBus,
		EventStore:      eventStore,
		EventBus:        eventBus,
		DB:              tenantDB,
		Hateoas:         sharedAPI.NewHATEOASLinks("/api/v1"),
		AuthMiddleware:  allowAllMiddleware{},
		SessionProvider: sessionManager,
	}))
	ic.router = router
	ic.cookies = sessionCookies(t, scsManager, sessionManager, ic.tenantID)

	return ic
}

type allowAllMiddleware struct{}

func (allowAllMiddleware) RequirePermission(_ authPL.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return next }
}

func (ic *integrationContext) tenantAndActorMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID, _ := sharedvo.NewTenantID(ic.tenantID)
			ctx := sharedctx.WithTenant(r.Context(), tenantID)
			ctx = sharedctx.WithActor(ctx, sharedctx.NewActor(uuid.New().String(), "admin@example.com", sharedctx.RoleAdmin))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func sessionCookies(t *testing.T, scsManager *scs.SessionManager, sessionManager *session.SessionManager, tenantID string) []*http.Cookie {
	t.Helper()
	sessionData := fmt.Sprintf(`{
		"tenantId": "%s",
		"userId": "%s",
		"userEmail": "admin@example.com",
		"accessToken": "test-token",
		"refreshToken": "test-refresh",
		"tokenExpiry": "%s",
		"authenticated": true
	}`, tenantID, uuid.New().String(), time.Now().Add(time.Hour).Format(time.RFC3339))
	authSession, err := session.UnmarshalAuthSession([]byte(sessionData))
	require.NoError(t, err)

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Post("/setup", func(w http.ResponseWriter, r *http.Request) {
		sessionManager.StoreAuthenticatedSession(r.Context(), authSession)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/setup", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	cookies := rec.Result().Cookies()
	require.NotEmpty(t, cookies)
	return cookies
}

func (ic *integrationContext) do(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(payload)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, cookie := range ic.cookies {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	ic.router.ServeHTTP(rec, req)
	return rec
}

func (ic *integrationContext) getConfiguration(t *testing.T) OnePagerConfigurationDTO {
	t.Helper()
	rec := ic.do(t, http.MethodGet, "/one-pagers/configurations/application", nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var dto OnePagerConfigurationDTO
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &dto))
	ic.createdIDs = append(ic.createdIDs, dto.ID)
	return dto
}

func TestGetConfiguration_LazilyCreatesDefault_Integration(t *testing.T) {
	ic := setupIntegration(t)

	dto := ic.getConfiguration(t)

	assert.Equal(t, "application", dto.SubjectType)
	assert.Equal(t, 1, dto.Version)
	included := make([]string, 0)
	for _, field := range dto.BuiltInFields {
		if field.Included {
			included = append(included, field.ID)
		}
	}
	assert.ElementsMatch(t, []string{"name", "description", "experts"}, included)
	assert.Empty(t, dto.CustomFields)
	assert.Contains(t, dto.Links, "self")
	assert.Contains(t, dto.Links, "x-define-custom-field")

	second := ic.getConfiguration(t)
	assert.Equal(t, dto.ID, second.ID)

	var count int
	require.NoError(t, ic.db.QueryRow(
		"SELECT COUNT(*) FROM onepagers.one_pager_configurations WHERE tenant_id = $1", ic.tenantID,
	).Scan(&count))
	assert.Equal(t, 1, count)
}

func TestDefineCustomField_Integration(t *testing.T) {
	ic := setupIntegration(t)
	dto := ic.getConfiguration(t)

	rec := ic.do(t, http.MethodPost, "/one-pagers/configurations/application/custom-fields", map[string]any{
		"name":      "Hosting model",
		"fieldType": "selection",
		"options":   []string{"On-prem", "Cloud"},
		"version":   dto.Version,
	})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var updated OnePagerConfigurationDTO
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &updated))
	require.Len(t, updated.CustomFields, 1)
	assert.Equal(t, "Hosting model", updated.CustomFields[0].Name)
	assert.Equal(t, 2, updated.Version)
	assert.Len(t, updated.DisplayOrder, 4)

	stale := ic.do(t, http.MethodPost, "/one-pagers/configurations/application/custom-fields", map[string]any{
		"name":      "Another",
		"fieldType": "text",
		"version":   dto.Version,
	})
	assert.Equal(t, http.StatusConflict, stale.Code)
}

func TestSetNumberFieldBounds_Integration(t *testing.T) {
	ic := setupIntegration(t)
	dto := ic.getConfiguration(t)

	defineRec := ic.do(t, http.MethodPost, "/one-pagers/configurations/application/custom-fields", map[string]any{
		"name":      "Maturity score",
		"fieldType": "number",
		"version":   dto.Version,
	})
	require.Equal(t, http.StatusCreated, defineRec.Code, defineRec.Body.String())
	var defined OnePagerConfigurationDTO
	require.NoError(t, json.Unmarshal(defineRec.Body.Bytes(), &defined))
	fieldID := defined.CustomFields[0].ID

	boundsRec := ic.do(t, http.MethodPut, fmt.Sprintf("/one-pagers/configurations/application/custom-fields/%s/bounds", fieldID), map[string]any{
		"min":     0,
		"max":     5,
		"version": defined.Version,
	})
	require.Equal(t, http.StatusOK, boundsRec.Code, boundsRec.Body.String())

	var bounded OnePagerConfigurationDTO
	require.NoError(t, json.Unmarshal(boundsRec.Body.Bytes(), &bounded))
	require.NotNil(t, bounded.CustomFields[0].Min)
	require.NotNil(t, bounded.CustomFields[0].Max)
	assert.Equal(t, 0.0, *bounded.CustomFields[0].Min)
	assert.Equal(t, 5.0, *bounded.CustomFields[0].Max)

	stale := ic.do(t, http.MethodPut, fmt.Sprintf("/one-pagers/configurations/application/custom-fields/%s/bounds", fieldID), map[string]any{
		"min":     0,
		"max":     3,
		"version": defined.Version,
	})
	assert.Equal(t, http.StatusConflict, stale.Code)

	invalid := ic.do(t, http.MethodPut, fmt.Sprintf("/one-pagers/configurations/application/custom-fields/%s/bounds", fieldID), map[string]any{
		"min":     10,
		"max":     5,
		"version": bounded.Version,
	})
	assert.Equal(t, http.StatusBadRequest, invalid.Code)
}

func findBuiltInField(dto OnePagerConfigurationDTO, id string) (BuiltInFieldDTO, bool) {
	for _, field := range dto.BuiltInFields {
		if field.ID == id {
			return field, true
		}
	}
	return BuiltInFieldDTO{}, false
}

func TestChangeBuiltInFieldRequirement_Integration(t *testing.T) {
	ic := setupIntegration(t)
	dto := ic.getConfiguration(t)

	experts, ok := findBuiltInField(dto, "experts")
	require.True(t, ok)
	assert.False(t, experts.Required)
	require.Contains(t, experts.Links, "x-set-requirement")

	rec := ic.do(t, http.MethodPut, "/one-pagers/configurations/application/built-in-fields/experts/requirement", map[string]any{
		"required": true,
		"version":  dto.Version,
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var updated OnePagerConfigurationDTO
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &updated))
	requiredExperts, ok := findBuiltInField(updated, "experts")
	require.True(t, ok)
	assert.True(t, requiredExperts.Required)
	assert.Equal(t, dto.Version+1, updated.Version)

	reread := ic.getConfiguration(t)
	rereadExperts, ok := findBuiltInField(reread, "experts")
	require.True(t, ok)
	assert.True(t, rereadExperts.Required, "required flag persists across reads")

	stale := ic.do(t, http.MethodPut, "/one-pagers/configurations/application/built-in-fields/experts/requirement", map[string]any{
		"required": false,
		"version":  dto.Version,
	})
	assert.Equal(t, http.StatusConflict, stale.Code)
}

func TestChangeBuiltInFieldRequirement_RejectsExcludedBuiltIn_Integration(t *testing.T) {
	ic := setupIntegration(t)
	dto := ic.getConfiguration(t)

	excludeRec := ic.do(t, http.MethodPost, "/one-pagers/configurations/application/built-in-fields/experts/exclude", map[string]any{
		"version": dto.Version,
	})
	require.Equal(t, http.StatusOK, excludeRec.Code, excludeRec.Body.String())
	var excluded OnePagerConfigurationDTO
	require.NoError(t, json.Unmarshal(excludeRec.Body.Bytes(), &excluded))

	excludedExperts, ok := findBuiltInField(excluded, "experts")
	require.True(t, ok)
	assert.False(t, excludedExperts.Included)
	assert.NotContains(t, excludedExperts.Links, "x-set-requirement", "excluded built-in offers no set-requirement affordance")

	rejected := ic.do(t, http.MethodPut, "/one-pagers/configurations/application/built-in-fields/experts/requirement", map[string]any{
		"required": true,
		"version":  excluded.Version,
	})
	assert.Equal(t, http.StatusConflict, rejected.Code, rejected.Body.String())
}

func TestUniqueConstraint_Backstop_Integration(t *testing.T) {
	ic := setupIntegration(t)
	dto := ic.getConfiguration(t)

	duplicate := uuid.New().String()
	ic.createdIDs = append(ic.createdIDs, duplicate)
	_, err := ic.db.Exec(
		`INSERT INTO onepagers.one_pager_configurations
		(id, tenant_id, subject_type, configuration, version, created_at, modified_at, modified_by)
		VALUES ($1, $2, $3, '{}', 1, NOW(), NOW(), 'test@example.com')`,
		duplicate, ic.tenantID, dto.SubjectType,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "uq_one_pager_configurations_subject_type")
}
