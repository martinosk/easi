package api

import (
	"context"
	"net/http"

	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/types"
)

type EnterpriseCapabilityReadModels struct {
	Capability       *readmodels.EnterpriseCapabilityReadModel
	Link             *readmodels.EnterpriseCapabilityLinkReadModel
	Importance       *readmodels.EnterpriseStrategicImportanceReadModel
	MaturityAnalysis *readmodels.MaturityAnalysisReadModel
}

type EnterpriseCapabilityHandlers struct {
	commandBus      cqrs.CommandBus
	readModels      *EnterpriseCapabilityReadModels
	sessionProvider authPL.SessionProvider
	hateoas         *EnterpriseArchLinks
}

func NewEnterpriseCapabilityHandlers(
	commandBus cqrs.CommandBus,
	readModels *EnterpriseCapabilityReadModels,
	sessionProvider authPL.SessionProvider,
) *EnterpriseCapabilityHandlers {
	return &EnterpriseCapabilityHandlers{
		commandBus:      commandBus,
		readModels:      readModels,
		sessionProvider: sessionProvider,
		hateoas:         NewEnterpriseArchLinks(sharedAPI.NewHATEOASLinks("")),
	}
}

type EnterpriseCapabilityWriteRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

type CreateEnterpriseCapabilityRequest = EnterpriseCapabilityWriteRequest
type UpdateEnterpriseCapabilityRequest = EnterpriseCapabilityWriteRequest

// CreateEnterpriseCapability godoc
// @Summary Create a new enterprise capability
// @Description Creates a new enterprise capability for grouping domain capabilities
// @Tags enterprise-capabilities
// @Accept json
// @Produce json
// @Param capability body CreateEnterpriseCapabilityRequest true "Enterprise capability data"
// @Success 201 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseCapabilityDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities [post]
func (h *EnterpriseCapabilityHandlers) CreateEnterpriseCapability(w http.ResponseWriter, r *http.Request) {
	h.handleCapabilityWrite(w, r, "", func(req EnterpriseCapabilityWriteRequest, _ string) cqrs.Command {
		return &commands.CreateEnterpriseCapability{Name: req.Name, Description: req.Description, Category: req.Category}
	})
}

// GetAllEnterpriseCapabilities godoc
// @Summary Get all enterprise capabilities
// @Description Retrieves all active enterprise capabilities with optional pagination
// @Tags enterprise-capabilities
// @Produce json
// @Param limit query int false "Maximum number of results (default 20, max 100)"
// @Param cursor query string false "Pagination cursor for next page"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseCapabilityDTO}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities [get]
func (h *EnterpriseCapabilityHandlers) GetAllEnterpriseCapabilities(w http.ResponseWriter, r *http.Request) {
	capabilities, err := h.readModels.Capability.GetAll(r.Context())
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	for i := range capabilities {
		capabilities[i].Links = h.hateoas.EnterpriseCapabilityLinksForActor(capabilities[i].ID, actor)
	}

	sharedAPI.RespondCollection(w, http.StatusOK, capabilities, h.hateoas.EnterpriseCapabilityCollectionLinks())
}

// GetEnterpriseCapabilityByID godoc
// @Summary Get an enterprise capability by ID
// @Description Retrieves a specific enterprise capability by its ID
// @Tags enterprise-capabilities
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseCapabilityDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id} [get]
func (h *EnterpriseCapabilityHandlers) GetEnterpriseCapabilityByID(w http.ResponseWriter, r *http.Request) {
	capability := h.getCapabilityOrNotFound(w, r, sharedAPI.GetPathParam(r, "id"))
	if capability == nil {
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	capability.Links = h.hateoas.EnterpriseCapabilityLinksForActor(capability.ID, actor)
	sharedAPI.RespondJSON(w, http.StatusOK, capability)
}

// UpdateEnterpriseCapability godoc
// @Summary Update an enterprise capability
// @Description Updates the name, description, and category of an enterprise capability
// @Tags enterprise-capabilities
// @Accept json
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Param capability body UpdateEnterpriseCapabilityRequest true "Updated capability data"
// @Success 200 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseCapabilityDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id} [put]
func (h *EnterpriseCapabilityHandlers) UpdateEnterpriseCapability(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")
	h.handleCapabilityWrite(w, r, id, func(req EnterpriseCapabilityWriteRequest, capID string) cqrs.Command {
		return &commands.UpdateEnterpriseCapability{ID: capID, Name: req.Name, Description: req.Description, Category: req.Category}
	})
}

// DeleteEnterpriseCapability godoc
// @Summary Delete an enterprise capability
// @Description Soft deletes an enterprise capability (marks as inactive)
// @Tags enterprise-capabilities
// @Param id path string true "Enterprise capability ID"
// @Success 204 "No Content"
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id} [delete]
func (h *EnterpriseCapabilityHandlers) DeleteEnterpriseCapability(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")
	if h.getCapabilityOrNotFound(w, r, id) == nil {
		return
	}
	h.dispatchDelete(w, r, &commands.DeleteEnterpriseCapability{ID: id})
}

type scopedCollection[T any] struct {
	fetch           func(ctx context.Context, ecID string) ([]T, error)
	decorate        func(r *http.Request, ecID string, items []T)
	collectionLinks func(ecID string) types.Links
}

func respondScopedCollection[T any](
	w http.ResponseWriter,
	r *http.Request,
	h *EnterpriseCapabilityHandlers,
	endpoint scopedCollection[T],
) {
	ecID, ok := h.requireCapability(w, r)
	if !ok {
		return
	}
	items, err := endpoint.fetch(r.Context(), ecID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	endpoint.decorate(r, ecID, items)
	sharedAPI.RespondCollection(w, http.StatusOK, items, endpoint.collectionLinks(ecID))
}

func (h *EnterpriseCapabilityHandlers) handleCapabilityWrite(
	w http.ResponseWriter,
	r *http.Request,
	capabilityID string,
	buildCmd func(req EnterpriseCapabilityWriteRequest, capabilityID string) cqrs.Command,
) {
	req, ok := sharedAPI.DecodeRequestOrFail[EnterpriseCapabilityWriteRequest](w, r)
	if !ok {
		return
	}
	h.dispatchAndRespondWithCapability(w, r, buildCmd(req, capabilityID), capabilityID)
}

func getOrNotFound[T any](w http.ResponseWriter, fetchFn func() (*T, error), resourceName string) *T {
	result, err := fetchFn()
	if err != nil {
		sharedAPI.HandleError(w, err)
		return nil
	}
	if result == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, resourceName+" not found")
		return nil
	}
	return result
}

func (h *EnterpriseCapabilityHandlers) requireCapability(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := sharedAPI.GetPathParam(r, "id")
	if h.getCapabilityOrNotFound(w, r, id) == nil {
		return "", false
	}
	return id, true
}

func (h *EnterpriseCapabilityHandlers) getCapabilityOrNotFound(w http.ResponseWriter, r *http.Request, id string) *readmodels.EnterpriseCapabilityDTO {
	return getOrNotFound(w, func() (*readmodels.EnterpriseCapabilityDTO, error) {
		return h.readModels.Capability.GetByID(r.Context(), id)
	}, "Enterprise capability")
}

func (h *EnterpriseCapabilityHandlers) respondWithCapability(w http.ResponseWriter, r *http.Request, capabilityID string, statusCode int) {
	capability, err := h.readModels.Capability.GetByID(r.Context(), capabilityID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/enterprise-capabilities"), sharedAPI.ResourceID(capabilityID))

	if capability == nil {
		if statusCode == http.StatusCreated {
			sharedAPI.RespondCreatedNoBody(w, location)
			return
		}
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Enterprise capability not found")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	capability.Links = h.hateoas.EnterpriseCapabilityLinksForActor(capability.ID, actor)
	if statusCode == http.StatusCreated {
		sharedAPI.RespondCreated(w, location, capability)
	} else {
		sharedAPI.RespondJSON(w, statusCode, capability)
	}
}

func (h *EnterpriseCapabilityHandlers) dispatchAndRespondWithCapability(w http.ResponseWriter, r *http.Request, cmd cqrs.Command, capabilityID string) {
	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(resultID string) {
		id := capabilityID
		statusCode := http.StatusOK
		if id == "" {
			id = resultID
			statusCode = http.StatusCreated
		}
		h.respondWithCapability(w, r, id, statusCode)
	})
}

func (h *EnterpriseCapabilityHandlers) dispatchDelete(w http.ResponseWriter, r *http.Request, cmd cqrs.Command) {
	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		sharedAPI.RespondDeleted(w)
	})
}
