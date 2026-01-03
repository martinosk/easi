package audit

import (
	"net/http"
	"strconv"

	sharedAPI "easi/backend/internal/shared/api"

	"github.com/go-chi/chi/v5"
)

type AuditHandlers struct {
	readModel *AuditHistoryReadModel
	hateoas   *sharedAPI.HATEOASLinks
}

func NewAuditHandlers(readModel *AuditHistoryReadModel, hateoas *sharedAPI.HATEOASLinks) *AuditHandlers {
	return &AuditHandlers{
		readModel: readModel,
		hateoas:   hateoas,
	}
}

// GetAuditHistory godoc
// @Summary Get audit history for an aggregate
// @Description Retrieves paginated audit history entries for a specific aggregate by ID. Returns all events that have occurred on the aggregate, including event type, data, timestamp, version, and actor information. Requires audit:read permission (Admin, Architect, or Stakeholder role).
// @Tags audit
// @Produce json
// @Param aggregateId path string true "Aggregate ID (UUID)"
// @Param limit query int false "Number of items per page (default: 50, max: 100)" default(50)
// @Param cursor query string false "Opaque cursor token for pagination"
// @Success 200 {object} AuditHistoryResponse
// @Failure 400 {object} sharedAPI.ErrorResponse "Missing or invalid aggregateId"
// @Failure 401 {object} sharedAPI.ErrorResponse "Authentication required"
// @Failure 403 {object} sharedAPI.ErrorResponse "Insufficient permissions - requires audit:read"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security ApiKeyAuth
// @Router /audit/{aggregateId} [get]
func (h *AuditHandlers) GetAuditHistory(w http.ResponseWriter, r *http.Request) {
	aggregateID := chi.URLParam(r, "aggregateId")
	if aggregateID == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "aggregateId is required")
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	cursor := r.URL.Query().Get("cursor")

	entries, hasMore, nextCursor, err := h.readModel.GetHistoryByAggregateID(r.Context(), aggregateID, limit, cursor)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve audit history")
		return
	}

	links := map[string]string{
		"self": h.hateoas.AuditHistory(aggregateID),
	}

	response := AuditHistoryResponse{
		Entries: entries,
		Links:   links,
	}

	if hasMore {
		response.Pagination = &PaginationInfo{
			HasMore:    hasMore,
			NextCursor: nextCursor,
		}
	}

	sharedAPI.RespondJSON(w, http.StatusOK, response)
}
