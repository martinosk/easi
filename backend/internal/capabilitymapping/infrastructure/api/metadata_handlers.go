package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/go-chi/chi/v5"
)

func isValidationError(err error) bool {
	return errors.Is(err, valueobjects.ErrInvalidMaturityLevel) ||
		errors.Is(err, valueobjects.ErrMaturityValueOutOfRange) ||
		errors.Is(err, valueobjects.ErrInvalidOwnershipModel) ||
		errors.Is(err, valueobjects.ErrInvalidCapabilityStatus) ||
		errors.Is(err, valueobjects.ErrTagEmpty) ||
		errors.Is(err, valueobjects.ErrExpertNameEmpty) ||
		errors.Is(err, valueobjects.ErrExpertRoleEmpty) ||
		errors.Is(err, valueobjects.ErrExpertContactEmpty)
}

func isNotFoundError(err error) bool {
	return errors.Is(err, repositories.ErrCapabilityNotFound)
}

type UpdateCapabilityMetadataRequest struct {
	MaturityValue  *int   `json:"maturityValue,omitempty"`
	MaturityLevel  string `json:"maturityLevel,omitempty"`
	OwnershipModel string `json:"ownershipModel,omitempty"`
	PrimaryOwner   string `json:"primaryOwner,omitempty"`
	EAOwner        string `json:"eaOwner,omitempty"`
	Status         string `json:"status"`
}

type AddCapabilityExpertRequest struct {
	ExpertName  string `json:"expertName"`
	ExpertRole  string `json:"expertRole"`
	ContactInfo string `json:"contactInfo"`
}

type AddCapabilityTagRequest struct {
	Tag string `json:"tag"`
}

// UpdateCapabilityMetadata godoc
// @Summary Update capability metadata
// @Description Updates metadata fields like maturity level, ownership, and strategy alignment
// @Tags capabilities
// @Accept json
// @Produce json
// @Param id path string true "Capability ID"
// @Param metadata body UpdateCapabilityMetadataRequest true "Metadata"
// @Success 200 {object} easi_backend_internal_capabilitymapping_application_readmodels.CapabilityDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/metadata [put]
func (h *CapabilityHandlers) UpdateCapabilityMetadata(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateCapabilityMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	var maturityValue int
	if req.MaturityValue != nil {
		maturityValue = *req.MaturityValue
	}

	cmd := &commands.UpdateCapabilityMetadata{
		ID:             id,
		MaturityValue:  maturityValue,
		MaturityLevel:  req.MaturityLevel,
		OwnershipModel: req.OwnershipModel,
		PrimaryOwner:   req.PrimaryOwner,
		EAOwner:        req.EAOwner,
		Status:         req.Status,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if isNotFoundError(err) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Capability not found")
			return
		}
		if isValidationError(err) {
			sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to update capability metadata")
		return
	}

	capability, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve updated capability")
		return
	}

	if capability == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Capability not found")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	h.addLinksToCapability(capability, actor)

	sharedAPI.RespondJSON(w, http.StatusOK, capability)
}

// AddCapabilityExpert godoc
// @Summary Add an expert to a capability
// @Description Associates a subject matter expert with a capability
// @Tags capabilities
// @Accept json
// @Param id path string true "Capability ID"
// @Param expert body AddCapabilityExpertRequest true "Expert data"
// @Success 204 "No Content"
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/experts [post]
func (h *CapabilityHandlers) AddCapabilityExpert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req AddCapabilityExpertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	cmd := &commands.AddCapabilityExpert{
		CapabilityID: id,
		ExpertName:   req.ExpertName,
		ExpertRole:   req.ExpertRole,
		ContactInfo:  req.ContactInfo,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if isNotFoundError(err) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Capability not found")
			return
		}
		if isValidationError(err) {
			sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to add expert")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveCapabilityExpert godoc
// @Summary Remove an expert from a capability
// @Description Removes a subject matter expert from a capability
// @Tags capabilities
// @Param id path string true "Capability ID"
// @Param name query string true "Expert name"
// @Param role query string true "Expert role"
// @Param contact query string true "Contact info"
// @Success 204 "No Content"
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/experts [delete]
func (h *CapabilityHandlers) RemoveCapabilityExpert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	expertName := r.URL.Query().Get("name")
	expertRole := r.URL.Query().Get("role")
	contactInfo := r.URL.Query().Get("contact")

	if !hasAllRequiredParams(expertName, expertRole, contactInfo) {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Missing required query parameters: name, role, contact")
		return
	}

	cmd := &commands.RemoveCapabilityExpert{
		CapabilityID: id,
		ExpertName:   expertName,
		ExpertRole:   expertRole,
		ContactInfo:  contactInfo,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if isNotFoundError(err) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Capability not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to remove expert")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func hasAllRequiredParams(params ...string) bool {
	for _, p := range params {
		if p == "" {
			return false
		}
	}
	return true
}

// GetExpertRoles godoc
// @Summary Get distinct expert roles for autocomplete
// @Description Retrieves distinct expert roles used across all capabilities for autocomplete support
// @Tags capabilities
// @Produce json
// @Success 200 {object} map[string][]string "Roles list"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/expert-roles [get]
func (h *CapabilityHandlers) GetExpertRoles(w http.ResponseWriter, r *http.Request) {
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

// AddCapabilityTag godoc
// @Summary Add a tag to a capability
// @Description Associates a tag with a capability for categorization
// @Tags capabilities
// @Accept json
// @Param id path string true "Capability ID"
// @Param tag body AddCapabilityTagRequest true "Tag data"
// @Success 204 "No Content"
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/tags [post]
func (h *CapabilityHandlers) AddCapabilityTag(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req AddCapabilityTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	cmd := &commands.AddCapabilityTag{
		CapabilityID: id,
		Tag:          req.Tag,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if isNotFoundError(err) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Capability not found")
			return
		}
		if isValidationError(err) {
			sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to add tag")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
