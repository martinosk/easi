package api

import (
	"net/http"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type ValueStreamHandlers struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.ValueStreamReadModel
	hateoas    *ValueStreamsLinks
}

func NewValueStreamHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ValueStreamReadModel,
	hateoas *ValueStreamsLinks,
) *ValueStreamHandlers {
	return &ValueStreamHandlers{
		commandBus: commandBus,
		readModel:  readModel,
		hateoas:    hateoas,
	}
}

type CreateValueStreamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type UpdateValueStreamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CreateValueStream godoc
// @Summary Create a new value stream
// @Description Creates a new value stream
// @Tags value-streams
// @Accept json
// @Produce json
// @Param valueStream body CreateValueStreamRequest true "Value stream data"
// @Success 201 {object} easi_backend_internal_valuestreams_application_readmodels.ValueStreamDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /value-streams [post]
func (h *ValueStreamHandlers) CreateValueStream(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[CreateValueStreamRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.CreateValueStream{
		Name:        req.Name,
		Description: req.Description,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "")
		return
	}

	h.respondWithValueStream(w, r, result.CreatedID, http.StatusCreated)
}

// GetAllValueStreams godoc
// @Summary List all value streams
// @Description Returns all value streams for the tenant
// @Tags value-streams
// @Produce json
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_valuestreams_application_readmodels.ValueStreamDTO}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /value-streams [get]
func (h *ValueStreamHandlers) GetAllValueStreams(w http.ResponseWriter, r *http.Request) {
	streams, err := h.readModel.GetAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve value streams")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	for i := range streams {
		streams[i].Links = h.hateoas.ValueStreamLinksForActor(streams[i].ID, actor)
	}

	sharedAPI.RespondCollection(w, http.StatusOK, streams, h.hateoas.ValueStreamCollectionLinksForActor(actor))
}

// GetValueStreamByID godoc
// @Summary Get a value stream by ID
// @Description Returns a single value stream with its details
// @Tags value-streams
// @Produce json
// @Param id path string true "Value Stream ID"
// @Success 200 {object} easi_backend_internal_valuestreams_application_readmodels.ValueStreamDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /value-streams/{id} [get]
func (h *ValueStreamHandlers) GetValueStreamByID(w http.ResponseWriter, r *http.Request) {
	vs := h.getValueStreamOrNotFound(w, r, sharedAPI.GetPathParam(r, "id"))
	if vs == nil {
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	vs.Links = h.hateoas.ValueStreamLinksForActor(vs.ID, actor)
	sharedAPI.RespondJSON(w, http.StatusOK, vs)
}

// UpdateValueStream godoc
// @Summary Update a value stream
// @Description Updates an existing value stream's name and description
// @Tags value-streams
// @Accept json
// @Produce json
// @Param id path string true "Value Stream ID"
// @Param valueStream body UpdateValueStreamRequest true "Updated value stream data"
// @Success 200 {object} easi_backend_internal_valuestreams_application_readmodels.ValueStreamDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /value-streams/{id} [put]
func (h *ValueStreamHandlers) UpdateValueStream(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateValueStreamRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.UpdateValueStream{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		h.respondWithValueStream(w, r, id, http.StatusOK)
	})
}

// DeleteValueStream godoc
// @Summary Delete a value stream
// @Description Deletes a value stream and all its stages
// @Tags value-streams
// @Param id path string true "Value Stream ID"
// @Success 204 "No Content"
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /value-streams/{id} [delete]
func (h *ValueStreamHandlers) DeleteValueStream(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")
	if h.getValueStreamOrNotFound(w, r, id) == nil {
		return
	}

	result, err := h.commandBus.Dispatch(r.Context(), &commands.DeleteValueStream{ID: id})
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		sharedAPI.RespondDeleted(w)
	})
}

func (h *ValueStreamHandlers) getValueStreamOrNotFound(w http.ResponseWriter, r *http.Request, id string) *readmodels.ValueStreamDTO {
	vs, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve value stream")
		return nil
	}
	if vs == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Value stream not found")
		return nil
	}
	return vs
}

func (h *ValueStreamHandlers) respondWithValueStream(w http.ResponseWriter, r *http.Request, vsID string, statusCode int) {
	vs, err := h.readModel.GetByID(r.Context(), vsID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve value stream")
		return
	}

	location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/value-streams"), sharedAPI.ResourceID(vsID))

	if vs == nil {
		if statusCode == http.StatusCreated {
			sharedAPI.RespondCreated(w, location, map[string]string{
				"id":      vsID,
				"message": "Value stream created, processing",
			})
			return
		}
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Value stream not found")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	vs.Links = h.hateoas.ValueStreamLinksForActor(vs.ID, actor)
	if statusCode == http.StatusCreated {
		sharedAPI.RespondCreated(w, location, vs)
	} else {
		sharedAPI.RespondJSON(w, statusCode, vs)
	}
}
