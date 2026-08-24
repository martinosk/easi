package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSendMessageInput_WriteFlagClampedToActorPermissions(t *testing.T) {
	cases := []struct {
		name      string
		role      sharedctx.Role
		requested bool
		expected  bool
	}{
		{"architect keeps requested write flag", sharedctx.RoleArchitect, true, true},
		{"stakeholder write flag is clamped", sharedctx.RoleStakeholder, true, false},
		{"architect without flag stays read-only", sharedctx.RoleArchitect, false, false},
	}

	h := &ConversationHandlers{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			convID := uuid.New().String()
			body, _ := json.Marshal(SendMessageRequest{Content: "hello", AllowWriteOperations: tc.requested})
			req := httptest.NewRequest("POST", "/assistant/conversations/"+convID+"/messages", bytes.NewReader(body))

			ctx := sharedctx.WithActor(req.Context(), sharedctx.NewActor("user-1", "user@example.com", tc.role))
			tenantID, _ := sharedvo.NewTenantID("tenant-1")
			ctx = sharedctx.WithTenant(ctx, tenantID)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", convID)
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

			input, herr := h.parseSendMessageInput(req.WithContext(ctx))
			require.Nil(t, herr)
			assert.Equal(t, tc.expected, input.allowWriteOperations)
		})
	}
}
