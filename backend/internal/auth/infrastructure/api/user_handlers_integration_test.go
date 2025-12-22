//go:build integration
// +build integration

package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/application/handlers"
	"easi/backend/internal/auth/application/projectors"
	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userTestContext struct {
	db         *sql.DB
	testID     string
	createdIDs []string
}

func setupUserTestDB(t *testing.T) (*userTestContext, func()) {
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

	err = db.Ping()
	require.NoError(t, err)

	testID := fmt.Sprintf("%d", time.Now().UnixNano())

	ctx := &userTestContext{
		db:         db,
		testID:     testID,
		createdIDs: make([]string, 0),
	}

	cleanup := func() {
		for _, id := range ctx.createdIDs {
			db.Exec("DELETE FROM users WHERE id = $1", id)
			db.Exec("DELETE FROM events WHERE aggregate_id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

func (ctx *userTestContext) trackID(id string) {
	ctx.createdIDs = append(ctx.createdIDs, id)
}

type userTestFixture struct {
	handlers       *UserHandlers
	readModel      *readmodels.UserReadModel
	commandBus     cqrs.CommandBus
	sessionManager *session.SessionManager
	scsManager     *scs.SessionManager
	router         chi.Router
}

func setupUserHandlers(db *sql.DB) *userTestFixture {
	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	readModel := readmodels.NewUserReadModel(tenantDB)
	userRepo := repositories.NewUserAggregateRepository(eventStore)

	projector := projectors.NewUserProjector(readModel)
	eventBus.Subscribe("UserCreated", projector)
	eventBus.Subscribe("UserRoleChanged", projector)
	eventBus.Subscribe("UserDisabled", projector)
	eventBus.Subscribe("UserEnabled", projector)

	changeRoleHandler := handlers.NewChangeUserRoleHandler(userRepo, readModel)
	disableHandler := handlers.NewDisableUserHandler(userRepo, readModel)
	enableHandler := handlers.NewEnableUserHandler(userRepo)

	commandBus.Register("ChangeUserRole", changeRoleHandler)
	commandBus.Register("DisableUser", disableHandler)
	commandBus.Register("EnableUser", enableHandler)

	scsManager := scs.New()
	scsManager.Store = memstore.New()
	scsManager.Lifetime = time.Hour
	sessionManager := session.NewSessionManager(scsManager)

	userHandlers := NewUserHandlers(commandBus, readModel, sessionManager)

	return &userTestFixture{
		handlers:       userHandlers,
		readModel:      readModel,
		commandBus:     commandBus,
		sessionManager: sessionManager,
		scsManager:     scsManager,
	}
}

func (f *userTestFixture) createRouter(userID string) chi.Router {
	setupHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authSession := createTestUserSession(userID)
		f.sessionManager.StoreAuthenticatedSession(r.Context(), authSession)
		w.WriteHeader(http.StatusOK)
	})

	r := chi.NewRouter()
	r.Use(f.scsManager.LoadAndSave)
	r.Post("/setup", setupHandler)
	r.Get("/api/v1/users", withTestTenantMiddleware(f.handlers.GetAllUsers))
	r.Get("/api/v1/users/{id}", withTestTenantMiddleware(f.handlers.GetUserByID))
	r.Patch("/api/v1/users/{id}", withTestTenantMiddleware(f.handlers.UpdateUser))
	return r
}

func withTestTenantMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = withTestTenant(r)
		next(w, r)
	}
}

func (f *userTestFixture) setupSessionAndGetCookies(t *testing.T, router chi.Router) []*http.Cookie {
	req := httptest.NewRequest(http.MethodPost, "/setup", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	cookies := rec.Result().Cookies()
	require.NotEmpty(t, cookies, "Should have session cookie")
	return cookies
}

func createTestUserSession(userID string) session.AuthSession {
	parsedID, _ := uuid.Parse(userID)
	sessionData := fmt.Sprintf(`{
		"tenantId": "acme",
		"userId": "%s",
		"userEmail": "test@acme.com",
		"accessToken": "test-token",
		"refreshToken": "test-refresh",
		"tokenExpiry": "%s",
		"authenticated": true
	}`, userID, time.Now().Add(1*time.Hour).Format(time.RFC3339))

	authSession, _ := session.UnmarshalAuthSession([]byte(sessionData))
	_ = parsedID
	return authSession
}

func createTestUser(t *testing.T, db *sql.DB, email, role string, testCtx *userTestContext) string {
	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	userRepo := repositories.NewUserAggregateRepository(eventStore)
	readModel := readmodels.NewUserReadModel(tenantDB)

	projector := projectors.NewUserProjector(readModel)
	eventBus.Subscribe("UserCreated", projector)
	eventBus.Subscribe("UserRoleChanged", projector)
	eventBus.Subscribe("UserDisabled", projector)
	eventBus.Subscribe("UserEnabled", projector)

	roleVO, err := valueobjects.RoleFromString(role)
	require.NoError(t, err)

	emailVO, err := valueobjects.NewEmail(email)
	require.NoError(t, err)

	user, err := aggregates.NewUser(emailVO, "Test User", roleVO, "", "")
	require.NoError(t, err)

	ctx := tenantContext()
	err = userRepo.Save(ctx, user)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	testCtx.trackID(user.ID())
	return user.ID()
}

func TestGetAllUsers_Integration(t *testing.T) {
	testCtx, cleanup := setupUserTestDB(t)
	defer cleanup()

	fixture := setupUserHandlers(testCtx.db)

	adminID := createTestUser(t, testCtx.db, fmt.Sprintf("admin-list-%s@acme.com", testCtx.testID), "admin", testCtx)

	for i := 0; i < 3; i++ {
		email := fmt.Sprintf("list-user-%d-%s@acme.com", i, testCtx.testID)
		createTestUser(t, testCtx.db, email, "architect", testCtx)
	}

	router := fixture.createRouter(adminID)
	cookies := fixture.setupSessionAndGetCookies(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?limit=10", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	data, ok := response["data"].([]interface{})
	require.True(t, ok, "response should have data array")
	assert.GreaterOrEqual(t, len(data), 4)
}

func TestGetUserByID_Integration(t *testing.T) {
	testCtx, cleanup := setupUserTestDB(t)
	defer cleanup()

	fixture := setupUserHandlers(testCtx.db)

	adminID := createTestUser(t, testCtx.db, fmt.Sprintf("admin-getbyid-%s@acme.com", testCtx.testID), "admin", testCtx)

	email := fmt.Sprintf("get-user-%s@acme.com", testCtx.testID)
	userID := createTestUser(t, testCtx.db, email, "stakeholder", testCtx)

	router := fixture.createRouter(adminID)
	cookies := fixture.setupSessionAndGetCookies(t, router)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s", userID), nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response UserResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, userID, response.ID)
	assert.Equal(t, email, response.Email)
	assert.Equal(t, "stakeholder", response.Role)
	assert.Equal(t, "active", response.Status)
}

func TestGetUserByID_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupUserTestDB(t)
	defer cleanup()

	fixture := setupUserHandlers(testCtx.db)

	adminID := createTestUser(t, testCtx.db, fmt.Sprintf("admin-notfound-%s@acme.com", testCtx.testID), "admin", testCtx)

	router := fixture.createRouter(adminID)
	cookies := fixture.setupSessionAndGetCookies(t, router)

	nonExistentID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s", nonExistentID), nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestChangeUserRole_Integration(t *testing.T) {
	testCtx, cleanup := setupUserTestDB(t)
	defer cleanup()

	fixture := setupUserHandlers(testCtx.db)

	adminID := createTestUser(t, testCtx.db, fmt.Sprintf("admin-changerole-%s@acme.com", testCtx.testID), "admin", testCtx)

	email := fmt.Sprintf("change-role-%s@acme.com", testCtx.testID)
	userID := createTestUser(t, testCtx.db, email, "architect", testCtx)

	router := fixture.createRouter(adminID)
	cookies := fixture.setupSessionAndGetCookies(t, router)

	body, _ := json.Marshal(UpdateUserRequest{Role: stringPtr("stakeholder")})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/users/%s", userID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response UserResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "stakeholder", response.Role)

	ctx := tenantContext()
	user, err := fixture.readModel.GetByIDString(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, "stakeholder", user.Role)
}

func TestChangeUserRole_InvalidRole_Integration(t *testing.T) {
	testCtx, cleanup := setupUserTestDB(t)
	defer cleanup()

	fixture := setupUserHandlers(testCtx.db)

	adminID := createTestUser(t, testCtx.db, fmt.Sprintf("admin-invalidrole-%s@acme.com", testCtx.testID), "admin", testCtx)

	email := fmt.Sprintf("invalid-role-%s@acme.com", testCtx.testID)
	userID := createTestUser(t, testCtx.db, email, "architect", testCtx)

	router := fixture.createRouter(adminID)
	cookies := fixture.setupSessionAndGetCookies(t, router)

	body, _ := json.Marshal(UpdateUserRequest{Role: stringPtr("superadmin")})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/users/%s", userID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDisableUser_Integration(t *testing.T) {
	testCtx, cleanup := setupUserTestDB(t)
	defer cleanup()

	fixture := setupUserHandlers(testCtx.db)

	adminID := createTestUser(t, testCtx.db, fmt.Sprintf("admin-disable-%s@acme.com", testCtx.testID), "admin", testCtx)

	email := fmt.Sprintf("disable-user-%s@acme.com", testCtx.testID)
	userID := createTestUser(t, testCtx.db, email, "architect", testCtx)

	router := fixture.createRouter(adminID)
	cookies := fixture.setupSessionAndGetCookies(t, router)

	body, _ := json.Marshal(UpdateUserRequest{Status: stringPtr("disabled")})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/users/%s", userID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response UserResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "disabled", response.Status)

	ctx := tenantContext()
	user, err := fixture.readModel.GetByIDString(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, "disabled", user.Status)
}

func TestEnableUser_Integration(t *testing.T) {
	testCtx, cleanup := setupUserTestDB(t)
	defer cleanup()

	fixture := setupUserHandlers(testCtx.db)

	adminID := createTestUser(t, testCtx.db, fmt.Sprintf("admin-enable-%s@acme.com", testCtx.testID), "admin", testCtx)

	email := fmt.Sprintf("enable-user-%s@acme.com", testCtx.testID)
	userID := createTestUser(t, testCtx.db, email, "architect", testCtx)

	ctx := tenantContext()
	disableCmd := &commands.DisableUser{
		UserID:       userID,
		DisabledByID: adminID,
	}
	err := fixture.commandBus.Dispatch(ctx, disableCmd)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	user, err := fixture.readModel.GetByIDString(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, "disabled", user.Status)

	router := fixture.createRouter(adminID)
	cookies := fixture.setupSessionAndGetCookies(t, router)

	body, _ := json.Marshal(UpdateUserRequest{Status: stringPtr("active")})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/users/%s", userID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response UserResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "active", response.Status)
}

func TestCannotDemoteLastAdmin_Integration(t *testing.T) {
	testCtx, cleanup := setupUserTestDB(t)
	defer cleanup()

	_, err := testCtx.db.Exec("SET app.current_tenant = 'acme'")
	require.NoError(t, err)
	_, err = testCtx.db.Exec("UPDATE users SET role = 'architect' WHERE role = 'admin'")
	require.NoError(t, err)

	fixture := setupUserHandlers(testCtx.db)

	adminEmail := fmt.Sprintf("sole-admin-demote-%s@acme.com", testCtx.testID)
	adminID := createTestUser(t, testCtx.db, adminEmail, "admin", testCtx)

	otherUserID := createTestUser(t, testCtx.db, fmt.Sprintf("other-user-demote-%s@acme.com", testCtx.testID), "architect", testCtx)

	router := fixture.createRouter(otherUserID)
	cookies := fixture.setupSessionAndGetCookies(t, router)

	body, _ := json.Marshal(UpdateUserRequest{Role: stringPtr("architect")})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/users/%s", adminID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "last admin")

	ctx := tenantContext()
	admin, err := fixture.readModel.GetByIDString(ctx, adminID)
	require.NoError(t, err)
	assert.Equal(t, "admin", admin.Role)
}

func TestCannotDisableSelf_Integration(t *testing.T) {
	testCtx, cleanup := setupUserTestDB(t)
	defer cleanup()

	fixture := setupUserHandlers(testCtx.db)

	createTestUser(t, testCtx.db, fmt.Sprintf("other-admin-self-%s@acme.com", testCtx.testID), "admin", testCtx)

	email := fmt.Sprintf("self-disable-%s@acme.com", testCtx.testID)
	userID := createTestUser(t, testCtx.db, email, "admin", testCtx)

	router := fixture.createRouter(userID)
	cookies := fixture.setupSessionAndGetCookies(t, router)

	body, _ := json.Marshal(UpdateUserRequest{Status: stringPtr("disabled")})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/users/%s", userID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "own account")

	ctx := tenantContext()
	user, err := fixture.readModel.GetByIDString(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, "active", user.Status)
}

func TestCannotDisableLastAdmin_Integration(t *testing.T) {
	testCtx, cleanup := setupUserTestDB(t)
	defer cleanup()

	_, err := testCtx.db.Exec("SET app.current_tenant = 'acme'")
	require.NoError(t, err)
	_, err = testCtx.db.Exec("UPDATE users SET role = 'architect' WHERE role = 'admin'")
	require.NoError(t, err)

	fixture := setupUserHandlers(testCtx.db)

	adminEmail := fmt.Sprintf("sole-admin-disable-%s@acme.com", testCtx.testID)
	adminID := createTestUser(t, testCtx.db, adminEmail, "admin", testCtx)

	otherUserID := createTestUser(t, testCtx.db, fmt.Sprintf("other-user-disable-%s@acme.com", testCtx.testID), "architect", testCtx)

	router := fixture.createRouter(otherUserID)
	cookies := fixture.setupSessionAndGetCookies(t, router)

	body, _ := json.Marshal(UpdateUserRequest{Status: stringPtr("disabled")})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/users/%s", adminID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "last admin")

	ctx := tenantContext()
	admin, err := fixture.readModel.GetByIDString(ctx, adminID)
	require.NoError(t, err)
	assert.Equal(t, "active", admin.Status)
}

func TestDemoteAdmin_WithMultipleAdmins_Integration(t *testing.T) {
	testCtx, cleanup := setupUserTestDB(t)
	defer cleanup()

	fixture := setupUserHandlers(testCtx.db)

	admin1Email := fmt.Sprintf("admin1-%s@acme.com", testCtx.testID)
	admin1ID := createTestUser(t, testCtx.db, admin1Email, "admin", testCtx)

	admin2Email := fmt.Sprintf("admin2-%s@acme.com", testCtx.testID)
	admin2ID := createTestUser(t, testCtx.db, admin2Email, "admin", testCtx)

	router := fixture.createRouter(admin2ID)
	cookies := fixture.setupSessionAndGetCookies(t, router)

	body, _ := json.Marshal(UpdateUserRequest{Role: stringPtr("architect")})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/users/%s", admin1ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	ctx := tenantContext()
	admin1, err := fixture.readModel.GetByIDString(ctx, admin1ID)
	require.NoError(t, err)
	assert.Equal(t, "architect", admin1.Role)

	admin2, err := fixture.readModel.GetByIDString(ctx, admin2ID)
	require.NoError(t, err)
	assert.Equal(t, "admin", admin2.Role)
}

func TestFilterUsersByStatus_Integration(t *testing.T) {
	testCtx, cleanup := setupUserTestDB(t)
	defer cleanup()

	fixture := setupUserHandlers(testCtx.db)

	adminID := createTestUser(t, testCtx.db, fmt.Sprintf("admin-filter-status-%s@acme.com", testCtx.testID), "admin", testCtx)

	activeEmail := fmt.Sprintf("filter-active-%s@acme.com", testCtx.testID)
	createTestUser(t, testCtx.db, activeEmail, "architect", testCtx)

	disabledEmail := fmt.Sprintf("filter-disabled-%s@acme.com", testCtx.testID)
	disabledUserID := createTestUser(t, testCtx.db, disabledEmail, "stakeholder", testCtx)

	ctx := tenantContext()
	disableCmd := &commands.DisableUser{
		UserID:       disabledUserID,
		DisabledByID: adminID,
	}
	err := fixture.commandBus.Dispatch(ctx, disableCmd)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	router := fixture.createRouter(adminID)
	cookies := fixture.setupSessionAndGetCookies(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?status=disabled&limit=50", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].([]interface{})
	for _, item := range data {
		user := item.(map[string]interface{})
		assert.Equal(t, "disabled", user["status"])
	}
}

func TestFilterUsersByRole_Integration(t *testing.T) {
	testCtx, cleanup := setupUserTestDB(t)
	defer cleanup()

	fixture := setupUserHandlers(testCtx.db)

	adminEmail := fmt.Sprintf("role-admin-%s@acme.com", testCtx.testID)
	adminID := createTestUser(t, testCtx.db, adminEmail, "admin", testCtx)

	architectEmail := fmt.Sprintf("role-architect-%s@acme.com", testCtx.testID)
	createTestUser(t, testCtx.db, architectEmail, "architect", testCtx)

	router := fixture.createRouter(adminID)
	cookies := fixture.setupSessionAndGetCookies(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?role=admin&limit=50", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].([]interface{})
	for _, item := range data {
		user := item.(map[string]interface{})
		assert.Equal(t, "admin", user["role"])
	}
}

func stringPtr(s string) *string {
	return &s
}
