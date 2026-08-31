package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/accessdelegation/application/commands"
	"easi/backend/internal/accessdelegation/application/readmodels"
	authPL "easi/backend/internal/auth/publishedlanguage"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type recordingEditGrantReader struct {
	hasActiveGrantEmail string
	grantByID           *readmodels.EditGrantDTO
}

func (r *recordingEditGrantReader) HasActiveGrant(_ context.Context, granteeEmail, _, _ string) (bool, error) {
	r.hasActiveGrantEmail = granteeEmail
	return false, nil
}

func (r *recordingEditGrantReader) GetByID(_ context.Context, _ string) (*readmodels.EditGrantDTO, error) {
	return r.grantByID, nil
}

func (r *recordingEditGrantReader) GetByGranteeEmail(_ context.Context, _ string) ([]readmodels.EditGrantDTO, error) {
	return nil, nil
}

func (r *recordingEditGrantReader) GetActiveForArtifact(_ context.Context, _, _ string) ([]readmodels.EditGrantDTO, error) {
	return nil, nil
}

type recordingCommandHandler struct {
	dispatched []cqrs.Command
	result     cqrs.CommandResult
	err        error
}

func (h *recordingCommandHandler) Handle(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	h.dispatched = append(h.dispatched, cmd)
	return h.result, h.err
}

type grantTestHarness struct {
	handlers    *EditGrantHandlers
	grants      *recordingCommandHandler
	invitations *recordingCommandHandler
	reader      *recordingEditGrantReader
}

func newGrantTestHarness(invitationResult cqrs.CommandResult, invitationErr error) *grantTestHarness {
	harness := &grantTestHarness{
		grants:      &recordingCommandHandler{result: cqrs.NewResult("grant-1")},
		invitations: &recordingCommandHandler{result: invitationResult, err: invitationErr},
		reader: &recordingEditGrantReader{
			grantByID: &readmodels.EditGrantDTO{
				ID:           "grant-1",
				GrantorID:    "grantor-1",
				GranteeEmail: "newcomer@dfds.com",
				ArtifactType: "capability",
				ArtifactID:   "cap-1",
				Status:       "active",
			},
		},
	}

	commandBus := cqrs.NewInMemoryCommandBus()
	commandBus.Register("CreateEditGrant", harness.grants)
	commandBus.Register(authPL.EnsureInvitation{}.CommandName(), harness.invitations)

	harness.handlers = NewEditGrantHandlers(EditGrantHandlerDeps{
		CommandBus: commandBus,
		ReadModel:  harness.reader,
		Hateoas:    NewEditGrantLinks(sharedAPI.NewHATEOASLinks("/api/v1")),
	})
	return harness
}

func (h *grantTestHarness) createGrant(t *testing.T, granteeEmail string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(map[string]string{
		"granteeEmail": granteeEmail,
		"artifactType": "capability",
		"artifactId":   "cap-1",
		"reason":       "vacation cover",
	})
	require.NoError(t, err)

	actor := sharedctx.NewActor("grantor-1", "grantor@dfds.com", sharedctx.RoleAdmin)
	req := httptest.NewRequest(http.MethodPost, "/edit-grants", bytes.NewReader(body))
	req = req.WithContext(sharedctx.WithActor(req.Context(), actor))
	rec := httptest.NewRecorder()

	h.handlers.CreateEditGrant(rec, req)
	return rec
}

func decodeGrant(t *testing.T, rec *httptest.ResponseRecorder) readmodels.EditGrantDTO {
	t.Helper()
	var grant readmodels.EditGrantDTO
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &grant))
	return grant
}

func TestCreateEditGrant_MixedCaseGranteeEmail_NormalizedEverywhere(t *testing.T) {
	harness := newGrantTestHarness(cqrs.NewResult("invitation-1"), nil)

	rec := harness.createGrant(t, "UdiCr@DFDS.com")

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "udicr@dfds.com", harness.reader.hasActiveGrantEmail)

	require.Len(t, harness.grants.dispatched, 1)
	grantCmd, ok := harness.grants.dispatched[0].(*commands.CreateEditGrant)
	require.True(t, ok)
	assert.Equal(t, "udicr@dfds.com", grantCmd.GranteeEmail)

	require.Len(t, harness.invitations.dispatched, 1)
	inviteCmd, ok := harness.invitations.dispatched[0].(*authPL.EnsureInvitation)
	require.True(t, ok)
	assert.Equal(t, "udicr@dfds.com", inviteCmd.Email)
}

func TestCreateEditGrant_InvalidGranteeEmail_ReturnsBadRequest(t *testing.T) {
	harness := newGrantTestHarness(cqrs.EmptyResult(), nil)

	rec := harness.createGrant(t, "not-an-email")

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Empty(t, harness.invitations.dispatched)
}

func TestCreateEditGrant_GranteeWithoutAccount_EnsuresStakeholderInvitation(t *testing.T) {
	harness := newGrantTestHarness(cqrs.NewResult("invitation-1"), nil)

	rec := harness.createGrant(t, "Newcomer@DFDS.com")

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Len(t, harness.invitations.dispatched, 1)
	assert.Equal(t, &authPL.EnsureInvitation{
		Email:        "newcomer@dfds.com",
		Role:         "stakeholder",
		InviterID:    "grantor-1",
		InviterEmail: "grantor@dfds.com",
	}, harness.invitations.dispatched[0])
	assert.True(t, decodeGrant(t, rec).InvitationCreated)
}

func TestCreateEditGrant_GranteeAlreadyKnown_ReportsNoInvitation(t *testing.T) {
	harness := newGrantTestHarness(cqrs.EmptyResult(), nil)

	rec := harness.createGrant(t, "existing@dfds.com")

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Len(t, harness.invitations.dispatched, 1)
	assert.False(t, decodeGrant(t, rec).InvitationCreated)
}

func TestCreateEditGrant_InvitationRejected_StillCreatesGrant(t *testing.T) {
	harness := newGrantTestHarness(cqrs.EmptyResult(), errors.New("email domain is not allowed"))

	rec := harness.createGrant(t, "outsider@example.com")

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Len(t, harness.grants.dispatched, 1)
	assert.False(t, decodeGrant(t, rec).InvitationCreated)
}
