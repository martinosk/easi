package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

type stubConfigurationStatus struct {
	configured bool
}

func (s stubConfigurationStatus) IsConfigured(context.Context) (bool, error) {
	return s.configured, nil
}

func TestGetAssistantStatus_ReportsConfiguration(t *testing.T) {
	for _, configured := range []bool{true, false} {
		handlers := NewAssistantStatusHandlers(stubConfigurationStatus{configured: configured}, sharedAPI.NewHATEOASLinks("/api/v1"))
		rec := httptest.NewRecorder()

		handlers.GetStatus(rec, httptest.NewRequest(http.MethodGet, "/assistant/status", nil))

		require.Equal(t, http.StatusOK, rec.Code)
		var body AssistantStatusResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, configured, body.Configured)
		assert.Equal(t, "/api/v1/assistant/status", body.Links["self"].Href)
		if configured {
			assert.Equal(t, "/api/v1/assistant/conversations", body.Links["x-conversations"].Href)
		} else {
			assert.NotContains(t, body.Links, "x-conversations")
		}
	}
}

func TestGetAssistantStatus_ConversationsWriteLink(t *testing.T) {
	cases := []struct {
		name        string
		permissions map[string]bool
		wantWrite   bool
	}{
		{"actor with capabilities write gets write link", map[string]bool{"capabilities:write": true}, true},
		{"actor with components write gets write link", map[string]bool{"components:write": true}, true},
		{"read-only actor has no write link", map[string]bool{"capabilities:read": true}, false},
		{"actor without permissions has no write link", nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handlers := NewAssistantStatusHandlers(stubConfigurationStatus{configured: true}, sharedAPI.NewHATEOASLinks("/api/v1"))
			rec := httptest.NewRecorder()
			ctx := sharedctx.WithActor(context.Background(), sharedctx.Actor{Permissions: tc.permissions})
			req := httptest.NewRequest(http.MethodGet, "/assistant/status", nil).WithContext(ctx)

			handlers.GetStatus(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			var body AssistantStatusResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			link, ok := body.Links["x-conversations-write"]
			assert.Equal(t, tc.wantWrite, ok)
			if tc.wantWrite {
				assert.Equal(t, "/api/v1/assistant/conversations", link.Href)
			}
		})
	}
}

func TestGetAssistantStatus_NotConfigured_NoConversationsWriteLink(t *testing.T) {
	handlers := NewAssistantStatusHandlers(stubConfigurationStatus{configured: false}, sharedAPI.NewHATEOASLinks("/api/v1"))
	rec := httptest.NewRecorder()
	ctx := sharedctx.WithActor(context.Background(), sharedctx.Actor{Permissions: map[string]bool{"capabilities:write": true}})
	req := httptest.NewRequest(http.MethodGet, "/assistant/status", nil).WithContext(ctx)

	handlers.GetStatus(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body AssistantStatusResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotContains(t, body.Links, "x-conversations-write")
}
