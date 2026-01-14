package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/entities"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
)

type ComponentExpertHandlers struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.ApplicationComponentReadModel
}

func NewComponentExpertHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ApplicationComponentReadModel,
) *ComponentExpertHandlers {
	return &ComponentExpertHandlers{
		commandBus: commandBus,
		readModel:  readModel,
	}
}

type AddComponentExpertRequest struct {
	ExpertName  string `json:"expertName"`
	ExpertRole  string `json:"expertRole"`
	ContactInfo string `json:"contactInfo"`
}

func isExpertValidationError(err error) bool {
	return errors.Is(err, entities.ErrExpertNameEmpty) ||
		errors.Is(err, entities.ErrExpertRoleEmpty) ||
		errors.Is(err, entities.ErrExpertContactEmpty)
}

func isComponentNotFoundError(err error) bool {
	return errors.Is(err, repositories.ErrComponentNotFound)
}

// AddComponentExpert godoc
// @Summary Add an expert to an application component
// @Description Associates a subject matter expert with an application component
// @Tags components
// @Accept json
// @Param id path string true "Component ID"
// @Param expert body AddComponentExpertRequest true "Expert data"
// @Success 204 "No Content"
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{id}/experts [post]
func (h *ComponentExpertHandlers) AddComponentExpert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req AddComponentExpertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	cmd := &commands.AddApplicationComponentExpert{
		ComponentID: id,
		ExpertName:  req.ExpertName,
		ExpertRole:  req.ExpertRole,
		ContactInfo: req.ContactInfo,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if isComponentNotFoundError(err) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Component not found")
			return
		}
		if isExpertValidationError(err) {
			sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to add expert")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveComponentExpert godoc
// @Summary Remove an expert from an application component
// @Description Removes a subject matter expert from an application component
// @Tags components
// @Param id path string true "Component ID"
// @Param name path string true "Expert name"
// @Success 204 "No Content"
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{id}/experts/{name} [delete]
func (h *ComponentExpertHandlers) RemoveComponentExpert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	expertName := chi.URLParam(r, "name")

	cmd := &commands.RemoveApplicationComponentExpert{
		ComponentID: id,
		ExpertName:  expertName,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if isComponentNotFoundError(err) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Component not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to remove expert")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetExpertRoles godoc
// @Summary Get distinct expert roles for autocomplete
// @Description Retrieves distinct expert roles used across all application components for autocomplete support
// @Tags components
// @Produce json
// @Success 200 {object} map[string][]string "Roles list"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/expert-roles [get]
func (h *ComponentExpertHandlers) GetExpertRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.readModel.GetDistinctExpertRoles(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve expert roles")
		return
	}

	if roles == nil {
		roles = []string{}
	}

	sharedAPI.RespondJSON(w, http.StatusOK, map[string][]string{
		"roles": roles,
	})
}
