package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/accessdelegation/application/commands"
	"easi/backend/internal/accessdelegation/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
)

type recordingEditGrantReader struct {
	hasActiveGrantEmail string
	grantByID           *readmodels.EditGrantDTO
}

func (r *recordingEditGrantReader) HasActiveGrant(ctx context.Context, granteeEmail, artifactType, artifactID string) (bool, error) {
	r.hasActiveGrantEmail = granteeEmail
	return false, nil
}

func (r *recordingEditGrantReader) GetByID(ctx context.Context, id string) (*readmodels.EditGrantDTO, error) {
	return r.grantByID, nil
}

func (r *recordingEditGrantReader) GetByGranteeEmail(ctx context.Context, email string) ([]readmodels.EditGrantDTO, error) {
	return nil, nil
}

func (r *recordingEditGrantReader) GetActiveForArtifact(ctx context.Context, artifactType, artifactID string) ([]readmodels.EditGrantDTO, error) {
	return nil, nil
}

type recordingGrantCommandBus struct {
	dispatched []cqrs.Command
}

func (b *recordingGrantCommandBus) Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	b.dispatched = append(b.dispatched, cmd)
	return cqrs.NewResult("grant-1"), nil
}

func (b *recordingGrantCommandBus) Register(commandName string, handler cqrs.CommandHandler) {}

type recordingUserLookup struct {
	queriedEmail string
}

func (l *recordingUserLookup) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	l.queriedEmail = email
	return false, nil
}

type recordingInvitationChecker struct {
	queriedEmail string
}

func (c *recordingInvitationChecker) HasPendingByEmail(ctx context.Context, email string) (bool, error) {
	c.queriedEmail = email
	return false, nil
}

func TestCreateEditGrant_MixedCaseGranteeEmail_NormalizedEverywhere(t *testing.T) {
	reader := &recordingEditGrantReader{
		grantByID: &readmodels.EditGrantDTO{
			ID:           "grant-1",
			GrantorID:    "grantor-1",
			GranteeEmail: "udicr@dfds.com",
			ArtifactType: "capability",
			ArtifactID:   "cap-1",
			Status:       "active",
		},
	}
	commandBus := &recordingGrantCommandBus{}
	userLookup := &recordingUserLookup{}
	invChecker := &recordingInvitationChecker{}

	handlers := NewEditGrantHandlers(EditGrantHandlerDeps{
		CommandBus: commandBus,
		ReadModel:  reader,
		Hateoas:    NewEditGrantLinks(sharedAPI.NewHATEOASLinks("/api/v1")),
		UserLookup: userLookup,
		InvChecker: invChecker,
		EventBus:   events.NewInMemoryEventBus(),
	})

	body, err := json.Marshal(map[string]string{
		"granteeEmail": "UdiCr@DFDS.com",
		"artifactType": "capability",
		"artifactId":   "cap-1",
		"reason":       "vacation cover",
	})
	require.NoError(t, err)

	actor := sharedctx.NewActor("grantor-1", "grantor@dfds.com", sharedctx.RoleAdmin)
	req := httptest.NewRequest(http.MethodPost, "/edit-grants", bytes.NewReader(body))
	req = req.WithContext(sharedctx.WithActor(req.Context(), actor))
	rec := httptest.NewRecorder()

	handlers.CreateEditGrant(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "udicr@dfds.com", reader.hasActiveGrantEmail)
	assert.Equal(t, "udicr@dfds.com", userLookup.queriedEmail)
	assert.Equal(t, "udicr@dfds.com", invChecker.queriedEmail)

	require.Len(t, commandBus.dispatched, 1)
	cmd, ok := commandBus.dispatched[0].(*commands.CreateEditGrant)
	require.True(t, ok)
	assert.Equal(t, "udicr@dfds.com", cmd.GranteeEmail)
}

func TestCreateEditGrant_InvalidGranteeEmail_ReturnsBadRequest(t *testing.T) {
	handlers := NewEditGrantHandlers(EditGrantHandlerDeps{
		CommandBus: &recordingGrantCommandBus{},
		ReadModel:  &recordingEditGrantReader{},
	})

	body, err := json.Marshal(map[string]string{
		"granteeEmail": "not-an-email",
		"artifactType": "capability",
		"artifactId":   "cap-1",
	})
	require.NoError(t, err)

	actor := sharedctx.NewActor("grantor-1", "grantor@dfds.com", sharedctx.RoleAdmin)
	req := httptest.NewRequest(http.MethodPost, "/edit-grants", bytes.NewReader(body))
	req = req.WithContext(sharedctx.WithActor(req.Context(), actor))
	rec := httptest.NewRecorder()

	handlers.CreateEditGrant(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
