package api

import (
	"net/http"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type StageHandlers struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.ValueStreamReadModel
	hateoas    *ValueStreamsLinks
}

func NewStageHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ValueStreamReadModel,
	hateoas *ValueStreamsLinks,
) *StageHandlers {
	return &StageHandlers{
		commandBus: commandBus,
		readModel:  readModel,
		hateoas:    hateoas,
	}
}

type CreateStageRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Position    *int   `json:"position,omitempty"`
}

type UpdateStageRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type ReorderStagesRequest struct {
	Positions []StagePositionRequest `json:"positions"`
}

type StagePositionRequest struct {
	StageID  string `json:"stageId"`
	Position int    `json:"position"`
}

type AddStageCapabilityRequest struct {
	CapabilityID string `json:"capabilityId"`
}

// CreateStage godoc
// @Summary Add a stage to a value stream
// @Description Creates a new stage in the value stream
// @Tags value-streams
// @Accept json
// @Produce json
// @Param id path string true "Value Stream ID"
// @Param stage body CreateStageRequest true "Stage data"
// @Success 201 {object} readmodels.ValueStreamDetailDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /value-streams/{id}/stages [post]
func (h *StageHandlers) CreateStage(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[CreateStageRequest](w, r)
	if !ok {
		return
	}

	h.dispatchStageCommand(w, r, http.StatusCreated, &commands.AddStage{
		ValueStreamID: sharedAPI.GetPathParam(r, "id"),
		Name:          req.Name,
		Description:   req.Description,
		Position:      req.Position,
	})
}

// UpdateStage godoc
// @Summary Update a stage
// @Description Updates a stage's name and description
// @Tags value-streams
// @Accept json
// @Produce json
// @Param id path string true "Value Stream ID"
// @Param stageId path string true "Stage ID"
// @Param stage body UpdateStageRequest true "Updated stage data"
// @Success 200 {object} readmodels.ValueStreamDetailDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /value-streams/{id}/stages/{stageId} [put]
func (h *StageHandlers) UpdateStage(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[UpdateStageRequest](w, r)
	if !ok {
		return
	}

	h.dispatchStageCommand(w, r, http.StatusOK, &commands.UpdateStage{
		ValueStreamID: sharedAPI.GetPathParam(r, "id"),
		StageID:       sharedAPI.GetPathParam(r, "stageId"),
		Name:          req.Name,
		Description:   req.Description,
	})
}

// DeleteStage godoc
// @Summary Remove a stage
// @Description Removes a stage from the value stream
// @Tags value-streams
// @Param id path string true "Value Stream ID"
// @Param stageId path string true "Stage ID"
// @Success 200 {object} readmodels.ValueStreamDetailDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /value-streams/{id}/stages/{stageId} [delete]
func (h *StageHandlers) DeleteStage(w http.ResponseWriter, r *http.Request) {
	h.dispatchStageCommand(w, r, http.StatusOK, &commands.RemoveStage{
		ValueStreamID: sharedAPI.GetPathParam(r, "id"),
		StageID:       sharedAPI.GetPathParam(r, "stageId"),
	})
}

// ReorderStages godoc
// @Summary Reorder stages
// @Description Updates the positions of all stages in a value stream
// @Tags value-streams
// @Accept json
// @Produce json
// @Param id path string true "Value Stream ID"
// @Param positions body ReorderStagesRequest true "New positions"
// @Success 200 {object} readmodels.ValueStreamDetailDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /value-streams/{id}/stages/positions [put]
func (h *StageHandlers) ReorderStages(w http.ResponseWriter, r *http.Request) {
	vsID := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[ReorderStagesRequest](w, r)
	if !ok {
		return
	}

	positions := make([]commands.StagePositionEntry, len(req.Positions))
	for i, p := range req.Positions {
		positions[i] = commands.StagePositionEntry{
			StageID:  p.StageID,
			Position: p.Position,
		}
	}

	h.dispatchStageCommand(w, r, http.StatusOK, &commands.ReorderStages{
		ValueStreamID: vsID,
		Positions:     positions,
	})
}

// AddStageCapability godoc
// @Summary Map a capability to a stage
// @Description Adds a capability mapping to a stage
// @Tags value-streams
// @Accept json
// @Produce json
// @Param id path string true "Value Stream ID"
// @Param stageId path string true "Stage ID"
// @Param capability body AddStageCapabilityRequest true "Capability data"
// @Success 200 {object} readmodels.ValueStreamDetailDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /value-streams/{id}/stages/{stageId}/capabilities [post]
func (h *StageHandlers) AddStageCapability(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[AddStageCapabilityRequest](w, r)
	if !ok {
		return
	}

	h.dispatchStageCommand(w, r, http.StatusOK, &commands.AddStageCapability{
		ValueStreamID: sharedAPI.GetPathParam(r, "id"),
		StageID:       sharedAPI.GetPathParam(r, "stageId"),
		CapabilityID:  req.CapabilityID,
	})
}

// RemoveStageCapability godoc
// @Summary Remove a capability from a stage
// @Description Removes a capability mapping from a stage
// @Tags value-streams
// @Param id path string true "Value Stream ID"
// @Param stageId path string true "Stage ID"
// @Param capabilityId path string true "Capability ID"
// @Success 200 {object} readmodels.ValueStreamDetailDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /value-streams/{id}/stages/{stageId}/capabilities/{capabilityId} [delete]
func (h *StageHandlers) RemoveStageCapability(w http.ResponseWriter, r *http.Request) {
	h.dispatchStageCommand(w, r, http.StatusOK, &commands.RemoveStageCapability{
		ValueStreamID: sharedAPI.GetPathParam(r, "id"),
		StageID:       sharedAPI.GetPathParam(r, "stageId"),
		CapabilityID:  sharedAPI.GetPathParam(r, "capabilityId"),
	})
}

// GetValueStreamCapabilities godoc
// @Summary Get all capabilities mapped to a value stream
// @Description Returns all capability mappings across all stages
// @Tags value-streams
// @Produce json
// @Param id path string true "Value Stream ID"
// @Success 200 {array} readmodels.StageCapabilityMappingDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /value-streams/{id}/capabilities [get]
func (h *StageHandlers) GetValueStreamCapabilities(w http.ResponseWriter, r *http.Request) {
	vsID := sharedAPI.GetPathParam(r, "id")

	caps, err := h.readModel.GetCapabilitiesByValueStreamID(r.Context(), vsID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve capabilities")
		return
	}

	if caps == nil {
		caps = []readmodels.StageCapabilityMappingDTO{}
	}

	sharedAPI.RespondJSON(w, http.StatusOK, caps)
}

func (h *StageHandlers) dispatchStageCommand(w http.ResponseWriter, r *http.Request, statusCode int, cmd cqrs.Command) {
	vsID := sharedAPI.GetPathParam(r, "id")
	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		h.respondWithDetail(w, r, vsID, statusCode)
	})
}

func (h *StageHandlers) respondWithDetail(w http.ResponseWriter, r *http.Request, vsID string, statusCode int) {
	detail, err := h.readModel.GetValueStreamDetail(r.Context(), vsID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve value stream detail")
		return
	}
	if detail == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Value stream not found")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	detail.Links = h.hateoas.ValueStreamLinksForActor(detail.ID, actor)

	for i := range detail.Stages {
		detail.Stages[i].Links = h.hateoas.StageLinksForActor(detail.ID, detail.Stages[i].ID, actor)
	}

	for i := range detail.StageCapabilities {
		detail.StageCapabilities[i].Links = h.hateoas.StageCapabilityLinksForActor(
			detail.ID, detail.StageCapabilities[i].StageID, detail.StageCapabilities[i].CapabilityID, actor,
		)
	}

	if statusCode == http.StatusCreated {
		location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/value-streams"), sharedAPI.ResourceID(vsID))
		sharedAPI.RespondCreated(w, location, detail)
	} else {
		sharedAPI.RespondJSON(w, statusCode, detail)
	}
}
