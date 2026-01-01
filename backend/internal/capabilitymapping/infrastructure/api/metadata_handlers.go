package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/entities"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"

	"github.com/go-chi/chi/v5"
)

func isValidationError(err error) bool {
	return errors.Is(err, valueobjects.ErrInvalidPillarWeight) ||
		errors.Is(err, valueobjects.ErrInvalidMaturityLevel) ||
		errors.Is(err, valueobjects.ErrMaturityValueOutOfRange) ||
		errors.Is(err, valueobjects.ErrInvalidOwnershipModel) ||
		errors.Is(err, valueobjects.ErrInvalidCapabilityStatus) ||
		errors.Is(err, valueobjects.ErrInvalidStrategyPillar) ||
		errors.Is(err, valueobjects.ErrTagEmpty) ||
		errors.Is(err, entities.ErrExpertNameEmpty) ||
		errors.Is(err, entities.ErrExpertRoleEmpty) ||
		errors.Is(err, entities.ErrExpertContactEmpty)
}

func isNotFoundError(err error) bool {
	return errors.Is(err, repositories.ErrCapabilityNotFound)
}

type UpdateCapabilityMetadataRequest struct {
	StrategyPillar string `json:"strategyPillar,omitempty"`
	PillarWeight   int    `json:"pillarWeight,omitempty"`
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
		StrategyPillar: req.StrategyPillar,
		PillarWeight:   req.PillarWeight,
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

	capability.Links = h.hateoas.CapabilityLinks(capability.ID, capability.ParentID)

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
