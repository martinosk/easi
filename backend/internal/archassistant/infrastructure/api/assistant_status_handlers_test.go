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
