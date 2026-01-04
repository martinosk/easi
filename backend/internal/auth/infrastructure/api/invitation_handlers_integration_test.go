//go:build integration
// +build integration

package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/application/handlers"
	"easi/backend/internal/auth/application/projectors"
	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/application/services"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type invitationTestContext struct {
	db         *sql.DB
	testID     string
	createdIDs []string
}

func (ctx *invitationTestContext) setTenantContext(t *testing.T) {
	_, err := ctx.db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", testTenantID()))
	require.NoError(t, err)
}

func setupInvitationTestDB(t *testing.T) (*invitationTestContext, func()) {
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

	ctx := &invitationTestContext{
		db:         db,
		testID:     testID,
		createdIDs: make([]string, 0),
	}

	cleanup := func() {
		for _, id := range ctx.createdIDs {
			db.Exec("DELETE FROM invitations WHERE id = $1", id)
			db.Exec("DELETE FROM events WHERE aggregate_id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

func (ctx *invitationTestContext) trackID(id string) {
	ctx.createdIDs = append(ctx.createdIDs, id)
}

func (ctx *invitationTestContext) makeRequest(t *testing.T, method, url string, body []byte, urlParams map[string]string) (*httptest.ResponseRecorder, *http.Request) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, url, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req = withTestTenant(req)

	if len(urlParams) > 0 {
		rctx := chi.NewRouteContext()
		for key, value := range urlParams {
			rctx.URLParams.Add(key, value)
		}
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}

	return httptest.NewRecorder(), req
}

func setupInvitationHandlers(db *sql.DB) (*InvitationHandlers, *readmodels.InvitationReadModel, cqrs.CommandBus) {
	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	readModel := readmodels.NewInvitationReadModel(tenantDB)
	invitationRepo := repositories.NewInvitationRepository(eventStore)

	projector := projectors.NewInvitationProjector(readModel)
	eventBus.Subscribe("InvitationCreated", projector)
	eventBus.Subscribe("InvitationAccepted", projector)
	eventBus.Subscribe("InvitationRevoked", projector)
	eventBus.Subscribe("InvitationExpired", projector)

	createHandler := handlers.NewCreateInvitationHandler(invitationRepo)
	revokeHandler := handlers.NewRevokeInvitationHandler(invitationRepo)
	acceptHandler := handlers.NewAcceptInvitationHandler(invitationRepo, readModel)
	expireHandler := handlers.NewMarkInvitationExpiredHandler(invitationRepo)

	commandBus.Register("CreateInvitation", createHandler)
	commandBus.Register("RevokeInvitation", revokeHandler)
	commandBus.Register("AcceptInvitation", acceptHandler)
	commandBus.Register("MarkInvitationExpired", expireHandler)

	domainChecker := readmodels.NewTenantDomainChecker(tenantDB)
	invitationHandlers := NewInvitationHandlers(commandBus, readModel, domainChecker)

	return invitationHandlers, readModel, commandBus
}

func TestCreateInvitation_Integration(t *testing.T) {
	testCtx, cleanup := setupInvitationTestDB(t)
	defer cleanup()

	invitationHandlers, readModel, _ := setupInvitationHandlers(testCtx.db)

	reqBody := CreateInvitationRequest{
		Email: fmt.Sprintf("test-%s@acme.com", testCtx.testID),
		Role:  "architect",
	}
	body, _ := json.Marshal(reqBody)

	w, req := testCtx.makeRequest(t, http.MethodPost, "/api/v1/invitations", body, nil)
	invitationHandlers.CreateInvitation(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var response readmodels.InvitationDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.ID)
	assert.Equal(t, reqBody.Email, response.Email)
	assert.Equal(t, reqBody.Role, response.Role)
	assert.Equal(t, "pending", response.Status)
	testCtx.trackID(response.ID)

	invitation, err := readModel.GetByID(tenantContext(), response.ID)
	require.NoError(t, err)
	require.NotNil(t, invitation)
	assert.Equal(t, "pending", invitation.Status)
}

func TestCreateInvitation_DuplicateEmail_Integration(t *testing.T) {
	testCtx, cleanup := setupInvitationTestDB(t)
	defer cleanup()

	invitationHandlers, _, _ := setupInvitationHandlers(testCtx.db)

	email := fmt.Sprintf("duplicate-%s@acme.com", testCtx.testID)
	reqBody := CreateInvitationRequest{
		Email: email,
		Role:  "architect",
	}
	body, _ := json.Marshal(reqBody)

	w1, req1 := testCtx.makeRequest(t, http.MethodPost, "/api/v1/invitations", body, nil)
	invitationHandlers.CreateInvitation(w1, req1)
	require.Equal(t, http.StatusCreated, w1.Code)

	var response readmodels.InvitationDTO
	json.Unmarshal(w1.Body.Bytes(), &response)
	testCtx.trackID(response.ID)

	w2, req2 := testCtx.makeRequest(t, http.MethodPost, "/api/v1/invitations", body, nil)
	invitationHandlers.CreateInvitation(w2, req2)

	assert.Equal(t, http.StatusConflict, w2.Code)
}

func TestGetInvitationByID_Integration(t *testing.T) {
	testCtx, cleanup := setupInvitationTestDB(t)
	defer cleanup()

	invitationHandlers, _, _ := setupInvitationHandlers(testCtx.db)

	reqBody := CreateInvitationRequest{
		Email: fmt.Sprintf("getbyid-%s@acme.com", testCtx.testID),
		Role:  "stakeholder",
	}
	body, _ := json.Marshal(reqBody)

	wCreate, reqCreate := testCtx.makeRequest(t, http.MethodPost, "/api/v1/invitations", body, nil)
	invitationHandlers.CreateInvitation(wCreate, reqCreate)
	require.Equal(t, http.StatusCreated, wCreate.Code)

	var created readmodels.InvitationDTO
	json.Unmarshal(wCreate.Body.Bytes(), &created)
	testCtx.trackID(created.ID)

	wGet, reqGet := testCtx.makeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/invitations/%s", created.ID), nil, map[string]string{"id": created.ID})
	invitationHandlers.GetInvitationByID(wGet, reqGet)

	require.Equal(t, http.StatusOK, wGet.Code)

	var fetched readmodels.InvitationDTO
	err := json.Unmarshal(wGet.Body.Bytes(), &fetched)
	require.NoError(t, err)

	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.Email, fetched.Email)
	assert.Equal(t, "stakeholder", fetched.Role)
}

func TestRevokeInvitation_Integration(t *testing.T) {
	testCtx, cleanup := setupInvitationTestDB(t)
	defer cleanup()

	invitationHandlers, readModel, _ := setupInvitationHandlers(testCtx.db)

	reqBody := CreateInvitationRequest{
		Email: fmt.Sprintf("revoke-%s@acme.com", testCtx.testID),
		Role:  "architect",
	}
	body, _ := json.Marshal(reqBody)

	wCreate, reqCreate := testCtx.makeRequest(t, http.MethodPost, "/api/v1/invitations", body, nil)
	invitationHandlers.CreateInvitation(wCreate, reqCreate)
	require.Equal(t, http.StatusCreated, wCreate.Code)

	var created readmodels.InvitationDTO
	json.Unmarshal(wCreate.Body.Bytes(), &created)
	testCtx.trackID(created.ID)

	revokeBody, _ := json.Marshal(UpdateInvitationRequest{Status: "revoked"})
	wRevoke, reqRevoke := testCtx.makeRequest(t, http.MethodPatch, fmt.Sprintf("/api/v1/invitations/%s", created.ID), revokeBody, map[string]string{"id": created.ID})
	invitationHandlers.UpdateInvitation(wRevoke, reqRevoke)

	require.Equal(t, http.StatusOK, wRevoke.Code)

	invitation, err := readModel.GetByID(tenantContext(), created.ID)
	require.NoError(t, err)
	require.NotNil(t, invitation)
	assert.Equal(t, "revoked", invitation.Status)
}

func TestGetAllInvitations_Integration(t *testing.T) {
	testCtx, cleanup := setupInvitationTestDB(t)
	defer cleanup()

	invitationHandlers, _, _ := setupInvitationHandlers(testCtx.db)

	for i := 0; i < 3; i++ {
		reqBody := CreateInvitationRequest{
			Email: fmt.Sprintf("list-%d-%s@acme.com", i, testCtx.testID),
			Role:  "architect",
		}
		body, _ := json.Marshal(reqBody)

		w, req := testCtx.makeRequest(t, http.MethodPost, "/api/v1/invitations", body, nil)
		invitationHandlers.CreateInvitation(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		var created readmodels.InvitationDTO
		json.Unmarshal(w.Body.Bytes(), &created)
		testCtx.trackID(created.ID)
	}

	wList, reqList := testCtx.makeRequest(t, http.MethodGet, "/api/v1/invitations?limit=10", nil, nil)
	invitationHandlers.GetAllInvitations(wList, reqList)

	require.Equal(t, http.StatusOK, wList.Code)

	var response map[string]interface{}
	err := json.Unmarshal(wList.Body.Bytes(), &response)
	require.NoError(t, err)

	data, ok := response["data"].([]interface{})
	require.True(t, ok, "response should have data array")
	assert.GreaterOrEqual(t, len(data), 3)
}

func TestCreateInvitation_InvalidEmail_Integration(t *testing.T) {
	testCtx, cleanup := setupInvitationTestDB(t)
	defer cleanup()

	invitationHandlers, _, _ := setupInvitationHandlers(testCtx.db)

	reqBody := CreateInvitationRequest{
		Email: "invalid-email",
		Role:  "architect",
	}
	body, _ := json.Marshal(reqBody)

	w, req := testCtx.makeRequest(t, http.MethodPost, "/api/v1/invitations", body, nil)
	invitationHandlers.CreateInvitation(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateInvitation_InvalidRole_Integration(t *testing.T) {
	testCtx, cleanup := setupInvitationTestDB(t)
	defer cleanup()

	invitationHandlers, _, _ := setupInvitationHandlers(testCtx.db)

	reqBody := CreateInvitationRequest{
		Email: fmt.Sprintf("test-%s@acme.com", testCtx.testID),
		Role:  "superadmin",
	}
	body, _ := json.Marshal(reqBody)

	w, req := testCtx.makeRequest(t, http.MethodPost, "/api/v1/invitations", body, nil)
	invitationHandlers.CreateInvitation(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateInvitation_UnregisteredDomain_Integration(t *testing.T) {
	testCtx, cleanup := setupInvitationTestDB(t)
	defer cleanup()

	invitationHandlers, _, _ := setupInvitationHandlers(testCtx.db)

	reqBody := CreateInvitationRequest{
		Email: fmt.Sprintf("user-%s@notallowed.com", testCtx.testID),
		Role:  "architect",
	}
	body, _ := json.Marshal(reqBody)

	w, req := testCtx.makeRequest(t, http.MethodPost, "/api/v1/invitations", body, nil)
	invitationHandlers.CreateInvitation(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "Email domain is not registered")
}

func TestInvitationExpiration_QueryDoesNotReturnExpired_Integration(t *testing.T) {
	testCtx, cleanup := setupInvitationTestDB(t)
	defer cleanup()

	invitationHandlers, readModel, _ := setupInvitationHandlers(testCtx.db)

	email := fmt.Sprintf("expire-%s@acme.com", testCtx.testID)
	reqBody := CreateInvitationRequest{
		Email: email,
		Role:  "architect",
	}
	body, _ := json.Marshal(reqBody)

	wCreate, reqCreate := testCtx.makeRequest(t, http.MethodPost, "/api/v1/invitations", body, nil)
	invitationHandlers.CreateInvitation(wCreate, reqCreate)
	require.Equal(t, http.StatusCreated, wCreate.Code)

	var created readmodels.InvitationDTO
	json.Unmarshal(wCreate.Body.Bytes(), &created)
	testCtx.trackID(created.ID)

	ctx := tenantContext()
	_, err := testCtx.db.ExecContext(ctx,
		"UPDATE invitations SET expires_at = $1 WHERE id = $2",
		time.Now().UTC().Add(-1*time.Hour),
		created.ID,
	)
	require.NoError(t, err)

	invitation, err := readModel.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, invitation)
	assert.True(t, time.Now().UTC().After(invitation.ExpiresAt))

	pendingInvitation, err := readModel.GetPendingByEmail(ctx, email)
	require.NoError(t, err)
	assert.Nil(t, pendingInvitation, "GetPendingByEmail should not return expired invitations")
}

func TestAcceptInvitation_WhenExpired_Integration(t *testing.T) {
	testCtx, cleanup := setupInvitationTestDB(t)
	defer cleanup()

	invitationHandlers, readModel, commandBus := setupInvitationHandlers(testCtx.db)

	email := fmt.Sprintf("expired-accept-%s@acme.com", testCtx.testID)
	reqBody := CreateInvitationRequest{
		Email: email,
		Role:  "architect",
	}
	body, _ := json.Marshal(reqBody)

	wCreate, reqCreate := testCtx.makeRequest(t, http.MethodPost, "/api/v1/invitations", body, nil)
	invitationHandlers.CreateInvitation(wCreate, reqCreate)
	require.Equal(t, http.StatusCreated, wCreate.Code)

	var created readmodels.InvitationDTO
	json.Unmarshal(wCreate.Body.Bytes(), &created)
	testCtx.trackID(created.ID)

	ctx := tenantContext()
	_, err := testCtx.db.ExecContext(ctx,
		"UPDATE invitations SET expires_at = $1 WHERE id = $2",
		time.Now().UTC().Add(-1*time.Hour),
		created.ID,
	)
	require.NoError(t, err)

	acceptCmd := &commands.AcceptInvitation{
		Email: email,
	}
	_, err = commandBus.Dispatch(ctx, acceptCmd)
	assert.Error(t, err)

	invitation, err := readModel.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "pending", invitation.Status)
}

func TestLoginService_UninvitedUserBlocked_Integration(t *testing.T) {
	testCtx, cleanup := setupInvitationTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(testCtx.db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	invitationReadModel := readmodels.NewInvitationReadModel(tenantDB)
	userReadModel := readmodels.NewUserReadModel(tenantDB)
	invitationRepo := repositories.NewInvitationRepository(eventStore)
	userRepo := repositories.NewUserAggregateRepository(eventStore)

	invitationProjector := projectors.NewInvitationProjector(invitationReadModel)
	eventBus.Subscribe("InvitationCreated", invitationProjector)
	eventBus.Subscribe("InvitationAccepted", invitationProjector)
	eventBus.Subscribe("InvitationRevoked", invitationProjector)
	eventBus.Subscribe("InvitationExpired", invitationProjector)

	userProjector := projectors.NewUserProjector(userReadModel)
	eventBus.Subscribe("UserCreated", userProjector)

	acceptHandler := handlers.NewAcceptInvitationHandler(invitationRepo, invitationReadModel)
	commandBus.Register("AcceptInvitation", acceptHandler)

	loginService := services.NewLoginService(userReadModel, invitationReadModel, commandBus, userRepo)

	ctx := tenantContext()
	uninvitedEmail := fmt.Sprintf("uninvited-%s@acme.com", testCtx.testID)

	result, err := loginService.ProcessLogin(ctx, uninvitedEmail, "Test User")
	assert.ErrorIs(t, err, services.ErrNoValidInvitation)
	assert.Nil(t, result)

	user, err := userReadModel.GetByEmail(ctx, uninvitedEmail)
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestLoginService_ValidInvitationCreatesUser_Integration(t *testing.T) {
	testCtx, cleanup := setupInvitationTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(testCtx.db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	invitationReadModel := readmodels.NewInvitationReadModel(tenantDB)
	userReadModel := readmodels.NewUserReadModel(tenantDB)
	invitationRepo := repositories.NewInvitationRepository(eventStore)
	userRepo := repositories.NewUserAggregateRepository(eventStore)

	invitationProjector := projectors.NewInvitationProjector(invitationReadModel)
	eventBus.Subscribe("InvitationCreated", invitationProjector)
	eventBus.Subscribe("InvitationAccepted", invitationProjector)
	eventBus.Subscribe("InvitationRevoked", invitationProjector)
	eventBus.Subscribe("InvitationExpired", invitationProjector)

	userProjector := projectors.NewUserProjector(userReadModel)
	eventBus.Subscribe("UserCreated", userProjector)

	createHandler := handlers.NewCreateInvitationHandler(invitationRepo)
	acceptHandler := handlers.NewAcceptInvitationHandler(invitationRepo, invitationReadModel)
	commandBus.Register("CreateInvitation", createHandler)
	commandBus.Register("AcceptInvitation", acceptHandler)

	loginService := services.NewLoginService(userReadModel, invitationReadModel, commandBus, userRepo)

	ctx := tenantContext()
	email := fmt.Sprintf("invited-%s@acme.com", testCtx.testID)

	createCmd := &commands.CreateInvitation{
		Email: email,
		Role:  "stakeholder",
	}
	_, err := commandBus.Dispatch(ctx, createCmd)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	invitation, err := invitationReadModel.GetPendingByEmail(ctx, email)
	require.NoError(t, err)
	require.NotNil(t, invitation)
	testCtx.trackID(invitation.ID)

	result, err := loginService.ProcessLogin(ctx, email, "Test User")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, email, result.Email)
	assert.Equal(t, "stakeholder", result.Role)
	assert.True(t, result.IsNew)

	user, err := userReadModel.GetByEmail(ctx, email)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, "stakeholder", user.Role)
	assert.Equal(t, "active", user.Status)

	acceptedInvitation, err := invitationReadModel.GetByID(ctx, invitation.ID)
	require.NoError(t, err)
	assert.Equal(t, "accepted", acceptedInvitation.Status)

	defer testCtx.db.Exec("DELETE FROM users WHERE email = $1", email)
}

func TestLoginService_ExpiredInvitationMarkedAsExpired_Integration(t *testing.T) {
	testCtx, cleanup := setupInvitationTestDB(t)
	defer cleanup()

	tenantDB := database.NewTenantAwareDB(testCtx.db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	invitationReadModel := readmodels.NewInvitationReadModel(tenantDB)
	userReadModel := readmodels.NewUserReadModel(tenantDB)
	invitationRepo := repositories.NewInvitationRepository(eventStore)
	userRepo := repositories.NewUserAggregateRepository(eventStore)

	invitationProjector := projectors.NewInvitationProjector(invitationReadModel)
	eventBus.Subscribe("InvitationCreated", invitationProjector)
	eventBus.Subscribe("InvitationAccepted", invitationProjector)
	eventBus.Subscribe("InvitationRevoked", invitationProjector)
	eventBus.Subscribe("InvitationExpired", invitationProjector)

	userProjector := projectors.NewUserProjector(userReadModel)
	eventBus.Subscribe("UserCreated", userProjector)

	createHandler := handlers.NewCreateInvitationHandler(invitationRepo)
	acceptHandler := handlers.NewAcceptInvitationHandler(invitationRepo, invitationReadModel)
	expireHandler := handlers.NewMarkInvitationExpiredHandler(invitationRepo)
	commandBus.Register("CreateInvitation", createHandler)
	commandBus.Register("AcceptInvitation", acceptHandler)
	commandBus.Register("MarkInvitationExpired", expireHandler)

	loginService := services.NewLoginService(userReadModel, invitationReadModel, commandBus, userRepo)

	ctx := tenantContext()
	email := fmt.Sprintf("lazy-expire-%d@acme.com", time.Now().UnixNano())

	createCmd := &commands.CreateInvitation{
		Email: email,
		Role:  "architect",
	}
	_, err := commandBus.Dispatch(ctx, createCmd)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	invitation, err := invitationReadModel.GetPendingByEmail(ctx, email)
	require.NoError(t, err)
	require.NotNil(t, invitation)
	testCtx.trackID(invitation.ID)

	_, err = testCtx.db.ExecContext(ctx,
		"UPDATE invitations SET expires_at = $1 WHERE id = $2",
		time.Now().UTC().Add(-1*time.Hour),
		invitation.ID,
	)
	require.NoError(t, err)

	result, err := loginService.ProcessLogin(ctx, email, "Test User")
	assert.ErrorIs(t, err, services.ErrNoValidInvitation)
	assert.Nil(t, result)

	expiredInvitation, err := invitationReadModel.GetByID(ctx, invitation.ID)
	require.NoError(t, err)
	require.NotNil(t, expiredInvitation)
	assert.Equal(t, "expired", expiredInvitation.Status, "Expired invitation should be marked as expired via lazy evaluation")

	user, err := userReadModel.GetByEmail(ctx, email)
	require.NoError(t, err)
	assert.Nil(t, user, "No user should be created for expired invitation")
}
