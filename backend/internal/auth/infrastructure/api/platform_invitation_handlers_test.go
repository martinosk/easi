package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/auth/application/commands"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubTenantCatalog struct {
	known bool
	err   error
}

func (c stubTenantCatalog) ExistsByID(_ context.Context, _ string) (bool, error) {
	return c.known, c.err
}

type capturingInvitationHandler struct {
	command  *commands.CreateInvitation
	tenantID string
	err      error
}

func (h *capturingInvitationHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	if h.err != nil {
		return cqrs.EmptyResult(), h.err
	}
	h.command = cmd.(*commands.CreateInvitation)
	tenant, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	h.tenantID = tenant.Value()
	return cqrs.NewResult("invitation-42"), nil
}

func postPlatformInvitation(catalog stubTenantCatalog, handler *capturingInvitationHandler, body string) *httptest.ResponseRecorder {
	commandBus := cqrs.NewInMemoryCommandBus()
	commandBus.Register("CreateInvitation", handler)
	handlers := NewPlatformInvitationHandlers(commandBus, catalog)

	req := httptest.NewRequest("POST", "/api/v1/auth/invitations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handlers.CreateInvitation(w, req)
	return w
}

func TestPlatformInvitationHandlers_CreatesInvitationInTheNamedTenant(t *testing.T) {
	handler := &capturingInvitationHandler{}

	w := postPlatformInvitation(stubTenantCatalog{known: true}, handler,
		`{"tenantId":"acme","email":"admin@acme.com","role":"architect"}`)

	assert.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, handler.command)
	assert.Equal(t, "admin@acme.com", handler.command.Email)
	assert.Equal(t, "architect", handler.command.Role)
	assert.Equal(t, "acme", handler.tenantID)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "invitation-42", body["id"])
	assert.Equal(t, "acme", body["tenantId"])
	assert.Empty(t, w.Header().Get("Location"))
}

func TestPlatformInvitationHandlers_DefaultsRoleToAdmin(t *testing.T) {
	handler := &capturingInvitationHandler{}

	w := postPlatformInvitation(stubTenantCatalog{known: true}, handler,
		`{"tenantId":"acme","email":"admin@acme.com"}`)

	assert.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, handler.command)
	assert.Equal(t, "admin", handler.command.Role)
}

func TestPlatformInvitationHandlers_ReportsUnknownTenant(t *testing.T) {
	w := postPlatformInvitation(stubTenantCatalog{known: false}, &capturingInvitationHandler{},
		`{"tenantId":"ghost","email":"admin@ghost.com"}`)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPlatformInvitationHandlers_RejectsIncompleteRequests(t *testing.T) {
	cases := map[string]string{
		"invalid json":   `{`,
		"missing email":  `{"tenantId":"acme"}`,
		"missing tenant": `{"email":"admin@acme.com"}`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			w := postPlatformInvitation(stubTenantCatalog{known: true}, &capturingInvitationHandler{}, body)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestPlatformInvitationHandlers_MapsCommandFailureToAStatusCode(t *testing.T) {
	handler := &capturingInvitationHandler{err: errors.New("boom")}

	w := postPlatformInvitation(stubTenantCatalog{known: true}, handler,
		`{"tenantId":"acme","email":"admin@acme.com"}`)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
