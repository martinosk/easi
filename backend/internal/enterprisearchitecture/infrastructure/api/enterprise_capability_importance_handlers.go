package api

import (
	"net/http"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

type SetStrategicImportanceRequest struct {
	PillarID   string `json:"pillarId"`
	PillarName string `json:"pillarName"`
	Importance int    `json:"importance"`
	Rationale  string `json:"rationale,omitempty"`
}

type UpdateStrategicImportanceRequest struct {
	Importance int    `json:"importance"`
	Rationale  string `json:"rationale,omitempty"`
}

// GetStrategicImportance godoc
// @Summary Get strategic importance ratings
// @Description Retrieves all strategic importance ratings for an enterprise capability
// @Tags enterprise-capabilities
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseStrategicImportanceDTO}
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/strategic-importance [get]
func (h *EnterpriseCapabilityHandlers) GetStrategicImportance(w http.ResponseWriter, r *http.Request) {
	respondScopedCollection(w, r, h, scopedCollection[readmodels.EnterpriseStrategicImportanceDTO]{
		fetch: h.readModels.Importance.GetByEnterpriseCapabilityID,
		decorate: func(req *http.Request, ecID string, items []readmodels.EnterpriseStrategicImportanceDTO) {
			actor, _ := sharedctx.GetActor(req.Context())
			for i := range items {
				items[i].Links = h.hateoas.EnterpriseStrategicImportanceLinksForActor(ecID, items[i].ID, actor)
			}
		},
		collectionLinks: h.hateoas.EnterpriseStrategicImportanceCollectionLinks,
	})
}

// SetStrategicImportance godoc
// @Summary Set strategic importance
// @Description Sets the strategic importance of an enterprise capability for a specific strategy pillar
// @Tags enterprise-capabilities
// @Accept json
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Param importance body SetStrategicImportanceRequest true "Strategic importance data"
// @Success 201 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseStrategicImportanceDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/strategic-importance [post]
func (h *EnterpriseCapabilityHandlers) SetStrategicImportance(w http.ResponseWriter, r *http.Request) {
	enterpriseCapabilityID := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[SetStrategicImportanceRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: enterpriseCapabilityID,
		PillarID:               req.PillarID,
		PillarName:             req.PillarName,
		Importance:             req.Importance,
		Rationale:              req.Rationale,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(createdID string) {
		location := sharedAPI.BuildSubResourceLink(sharedAPI.ResourcePath("/enterprise-capabilities"), sharedAPI.ResourceID(enterpriseCapabilityID), sharedAPI.ResourcePath("/strategic-importance/"+createdID))
		rating, err := h.readModels.Importance.GetByID(r.Context(), createdID)
		if err != nil || rating == nil {
			sharedAPI.RespondCreatedNoBody(w, location)
			return
		}
		h.decorateImportance(r, enterpriseCapabilityID, rating)
		sharedAPI.RespondCreated(w, location, rating)
	})
}

// UpdateStrategicImportance godoc
// @Summary Update strategic importance
// @Description Updates the strategic importance rating for a specific pillar
// @Tags enterprise-capabilities
// @Accept json
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Param importanceId path string true "Strategic importance ID"
// @Param importance body UpdateStrategicImportanceRequest true "Updated importance data"
// @Success 200 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseStrategicImportanceDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/strategic-importance/{importanceId} [put]
func (h *EnterpriseCapabilityHandlers) UpdateStrategicImportance(w http.ResponseWriter, r *http.Request) {
	enterpriseCapabilityID := sharedAPI.GetPathParam(r, "id")
	importanceID := sharedAPI.GetPathParam(r, "importanceId")

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateStrategicImportanceRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.UpdateEnterpriseStrategicImportance{
		ID:         importanceID,
		Importance: req.Importance,
		Rationale:  req.Rationale,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		rating, err := h.readModels.Importance.GetByID(r.Context(), importanceID)
		if err != nil {
			sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Update succeeded but failed to retrieve updated resource")
			return
		}
		if rating == nil {
			sharedAPI.RespondError(w, http.StatusInternalServerError, nil, "Update succeeded but resource not found")
			return
		}
		h.decorateImportance(r, enterpriseCapabilityID, rating)
		sharedAPI.RespondJSON(w, http.StatusOK, rating)
	})
}

// RemoveStrategicImportance godoc
// @Summary Remove strategic importance
// @Description Removes the strategic importance rating for a specific pillar
// @Tags enterprise-capabilities
// @Param id path string true "Enterprise capability ID"
// @Param importanceId path string true "Strategic importance ID"
// @Success 204 "No Content"
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/strategic-importance/{importanceId} [delete]
func (h *EnterpriseCapabilityHandlers) RemoveStrategicImportance(w http.ResponseWriter, r *http.Request) {
	importanceID := sharedAPI.GetPathParam(r, "importanceId")
	if h.getImportanceOrNotFound(w, r, importanceID) == nil {
		return
	}
	h.dispatchDelete(w, r, &commands.RemoveEnterpriseStrategicImportance{ID: importanceID})
}

func (h *EnterpriseCapabilityHandlers) decorateImportance(r *http.Request, enterpriseCapabilityID string, rating *readmodels.EnterpriseStrategicImportanceDTO) {
	actor, _ := sharedctx.GetActor(r.Context())
	rating.Links = h.hateoas.EnterpriseStrategicImportanceLinksForActor(enterpriseCapabilityID, rating.ID, actor)
}

func (h *EnterpriseCapabilityHandlers) getImportanceOrNotFound(w http.ResponseWriter, r *http.Request, id string) *readmodels.EnterpriseStrategicImportanceDTO {
	return getOrNotFound(w, func() (*readmodels.EnterpriseStrategicImportanceDTO, error) {
		return h.readModels.Importance.GetByID(r.Context(), id)
	}, "Importance rating")
}
