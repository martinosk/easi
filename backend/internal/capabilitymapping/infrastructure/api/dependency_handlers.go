package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
)

type DependencyHandlers struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.DependencyReadModel
	hateoas    *CapabilityMappingLinks
}

func NewDependencyHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.DependencyReadModel,
	hateoas *CapabilityMappingLinks,
) *DependencyHandlers {
	return &DependencyHandlers{
		commandBus: commandBus,
		readModel:  readModel,
		hateoas:    hateoas,
	}
}

type CreateDependencyRequest struct {
	SourceCapabilityID string `json:"sourceCapabilityId"`
	TargetCapabilityID string `json:"targetCapabilityId"`
	DependencyType     string `json:"dependencyType"`
	Description        string `json:"description,omitempty"`
}

// CreateDependency godoc
// @Summary Create a capability dependency
// @Description Creates a dependency relationship between two capabilities
// @Tags capability-dependencies
// @Accept json
// @Produce json
// @Param dependency body CreateDependencyRequest true "Dependency data"
// @Success 201 {object} easi_backend_internal_capabilitymapping_application_readmodels.DependencyDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capability-dependencies [post]
func (h *DependencyHandlers) CreateDependency(w http.ResponseWriter, r *http.Request) {
	var req CreateDependencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	_, err := valueobjects.NewCapabilityIDFromString(req.SourceCapabilityID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid source capability ID")
		return
	}

	_, err = valueobjects.NewCapabilityIDFromString(req.TargetCapabilityID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid target capability ID")
		return
	}

	_, err = valueobjects.NewDependencyType(req.DependencyType)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	cmd := &commands.CreateCapabilityDependency{
		SourceCapabilityID: req.SourceCapabilityID,
		TargetCapabilityID: req.TargetCapabilityID,
		DependencyType:     req.DependencyType,
		Description:        req.Description,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, handlers.ErrSourceCapabilityNotFound) || errors.Is(err, handlers.ErrTargetCapabilityNotFound) {
			sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create dependency")
		return
	}

	dependency, err := h.readModel.GetByID(r.Context(), result.CreatedID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created dependency")
		return
	}

	if dependency == nil {
		location := fmt.Sprintf("/api/capability-dependencies/%s", result.CreatedID)
		w.Header().Set("Location", location)
		sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{
			"id":      result.CreatedID,
			"message": "Dependency created, processing",
		})
		return
	}

	dependency.Links = h.hateoas.DependencyLinks(dependency.ID, dependency.SourceCapabilityID, dependency.TargetCapabilityID)

	location := fmt.Sprintf("/api/capability-dependencies/%s", dependency.ID)
	w.Header().Set("Location", location)
	sharedAPI.RespondJSON(w, http.StatusCreated, dependency)
}

// GetAllDependencies godoc
// @Summary Get all capability dependencies
// @Description Retrieves all capability dependencies in the system
// @Tags capability-dependencies
// @Produce json
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_capabilitymapping_application_readmodels.DependencyDTO}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capability-dependencies [get]
func (h *DependencyHandlers) GetAllDependencies(w http.ResponseWriter, r *http.Request) {
	dependencies, err := h.readModel.GetAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve dependencies")
		return
	}

	for i := range dependencies {
		dependencies[i].Links = h.hateoas.DependencyLinks(dependencies[i].ID, dependencies[i].SourceCapabilityID, dependencies[i].TargetCapabilityID)
	}

	links := sharedAPI.Links{
		"self": sharedAPI.NewLink("/api/v1/capability-dependencies", "GET"),
	}

	sharedAPI.RespondCollection(w, http.StatusOK, dependencies, links)
}

// GetOutgoingDependencies godoc
// @Summary Get outgoing dependencies for a capability
// @Description Retrieves all dependencies where the specified capability is the source
// @Tags capabilities
// @Produce json
// @Param id path string true "Capability ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_capabilitymapping_application_readmodels.DependencyDTO}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/dependencies/outgoing [get]
func (h *DependencyHandlers) GetOutgoingDependencies(w http.ResponseWriter, r *http.Request) {
	capabilityID := chi.URLParam(r, "id")

	dependencies, err := h.readModel.GetOutgoing(r.Context(), capabilityID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve outgoing dependencies")
		return
	}

	for i := range dependencies {
		dependencies[i].Links = h.hateoas.DependencyLinks(dependencies[i].ID, dependencies[i].SourceCapabilityID, dependencies[i].TargetCapabilityID)
	}

	links := sharedAPI.Links{
		"self": sharedAPI.NewLink("/api/v1/capabilities/"+capabilityID+"/dependencies/outgoing", "GET"),
		"up":   sharedAPI.NewLink("/api/v1/capabilities/"+capabilityID, "GET"),
	}

	sharedAPI.RespondCollection(w, http.StatusOK, dependencies, links)
}

// GetIncomingDependencies godoc
// @Summary Get incoming dependencies for a capability
// @Description Retrieves all dependencies where the specified capability is the target
// @Tags capabilities
// @Produce json
// @Param id path string true "Capability ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_capabilitymapping_application_readmodels.DependencyDTO}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/dependencies/incoming [get]
func (h *DependencyHandlers) GetIncomingDependencies(w http.ResponseWriter, r *http.Request) {
	capabilityID := chi.URLParam(r, "id")

	dependencies, err := h.readModel.GetIncoming(r.Context(), capabilityID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve incoming dependencies")
		return
	}

	for i := range dependencies {
		dependencies[i].Links = h.hateoas.DependencyLinks(dependencies[i].ID, dependencies[i].SourceCapabilityID, dependencies[i].TargetCapabilityID)
	}

	links := sharedAPI.Links{
		"self": sharedAPI.NewLink("/api/v1/capabilities/"+capabilityID+"/dependencies/incoming", "GET"),
		"up":   sharedAPI.NewLink("/api/v1/capabilities/"+capabilityID, "GET"),
	}

	sharedAPI.RespondCollection(w, http.StatusOK, dependencies, links)
}

// DeleteDependency godoc
// @Summary Delete a capability dependency
// @Description Deletes a capability dependency relationship
// @Tags capability-dependencies
// @Param id path string true "Dependency ID"
// @Success 204 "No Content"
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capability-dependencies/{id} [delete]
func (h *DependencyHandlers) DeleteDependency(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cmd := &commands.DeleteCapabilityDependency{
		ID: id,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if errors.Is(err, repositories.ErrDependencyNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Dependency not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to delete dependency")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
