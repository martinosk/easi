package api

import (
	"context"
	"net/http"

	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/types"
)

type ConfigurationStatus interface {
	IsConfigured(ctx context.Context) (bool, error)
}

type AssistantStatusResponse struct {
	Configured bool        `json:"configured"`
	Links      types.Links `json:"_links,omitempty"`
}

type AssistantStatusHandlers struct {
	status  ConfigurationStatus
	hateoas *sharedAPI.HATEOASLinks
}

func NewAssistantStatusHandlers(status ConfigurationStatus, hateoas *sharedAPI.HATEOASLinks) *AssistantStatusHandlers {
	return &AssistantStatusHandlers{status: status, hateoas: hateoas}
}

// GetStatus godoc
// @Summary Get assistant availability
// @Description Reports whether the tenant's AI assistant is configured and ready to use. The session advertises the assistant entry point on permission alone; this resource says whether the assistant can actually be used.
// @Tags assistant
// @Produce json
// @Success 200 {object} AssistantStatusResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /assistant/status [get]
func (h *AssistantStatusHandlers) GetStatus(w http.ResponseWriter, r *http.Request) {
	configured, err := h.status.IsConfigured(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to determine assistant status")
		return
	}
	links := types.Links{"self": h.hateoas.Get("/assistant/status")}
	if configured {
		links["x-conversations"] = h.hateoas.Get("/assistant/conversations")
	}
	sharedAPI.RespondJSON(w, http.StatusOK, AssistantStatusResponse{Configured: configured, Links: links})
}
