package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/domain"

	"github.com/go-chi/chi/v5"
)

func isParentChangeError(err error) bool {
	return errors.Is(err, aggregates.ErrCapabilityCannotBeOwnParent) ||
		errors.Is(err, aggregates.ErrWouldCreateCircularReference) ||
		errors.Is(err, aggregates.ErrWouldExceedMaximumDepth) ||
		errors.Is(err, domain.ErrInvalidValue) ||
		errors.Is(err, handlers.ErrParentCapabilityNotFound)
}

type ChangeCapabilityParentRequest struct {
	ParentID string `json:"parentId"`
}

// ChangeCapabilityParent godoc
// @Summary Change capability parent
// @Description Changes the parent of a capability and recalculates levels for the entire subtree
// @Tags capabilities
// @Accept json
// @Produce json
// @Param id path string true "Capability ID"
// @Param parent body ChangeCapabilityParentRequest true "New parent data"
// @Success 204 "No Content"
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /capabilities/{id}/parent [patch]
func (h *CapabilityHandlers) ChangeCapabilityParent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req ChangeCapabilityParentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	cmd := &commands.ChangeCapabilityParent{
		CapabilityID: id,
		NewParentID:  req.ParentID,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if errors.Is(err, repositories.ErrCapabilityNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Capability not found")
			return
		}
		if errors.Is(err, handlers.ErrParentCapabilityNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Parent capability not found")
			return
		}
		if isParentChangeError(err) {
			sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to change capability parent")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
