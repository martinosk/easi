package api

import (
	"net/http"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type ViewHandlers struct {
	commandBus   cqrs.CommandBus
	readModel    *readmodels.ArchitectureViewReadModel
	hateoas      *ViewLinks
	errorHandler *sharedAPI.ErrorHandler
}

func NewViewHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ArchitectureViewReadModel,
	hateoas *ViewLinks,
) *ViewHandlers {
	return &ViewHandlers{
		commandBus:   commandBus,
		readModel:    readModel,
		hateoas:      hateoas,
		errorHandler: sharedAPI.NewErrorHandler(),
	}
}

func (h *ViewHandlers) dispatchCommand(w http.ResponseWriter, r *http.Request, cmd cqrs.Command) {
	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		w.WriteHeader(http.StatusNoContent)
	})
}

type viewSettingUpdate[T any] struct {
	validate  func(value string) (T, error)
	createCmd func(viewID string, validated T) cqrs.Command
}

func handleViewSetting[T any](h *ViewHandlers, w http.ResponseWriter, r *http.Request, fieldName string, update viewSettingUpdate[T]) {
	viewID := sharedAPI.GetPathParam(r, "id")
	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	var body map[string]string
	if err := sharedAPI.DecodeRequestInto(r, &body); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	validated, err := update.validate(body[fieldName])
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	h.dispatchCommand(w, r, update.createCmd(viewID, validated))
}

type CreateViewRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type RenameViewRequest struct {
	Name string `json:"name"`
}

type ChangeVisibilityRequest struct {
	IsPrivate bool `json:"isPrivate"`
}

// CreateView godoc
// @Summary Create a new architecture view
// @Description Creates a new architecture view for organizing components
// @Tags views
// @Accept json
// @Produce json
// @Param view body CreateViewRequest true "View data"
// @Success 201 {object} readmodels.ArchitectureViewDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views [post]
func (h *ViewHandlers) CreateView(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[CreateViewRequest](w, r)
	if !ok {
		return
	}

	if _, err := valueobjects.NewViewName(req.Name); err != nil {
		h.errorHandler.HandleValidationError(w, err)
		return
	}

	cmd := &commands.CreateView{
		Name:        req.Name,
		Description: req.Description,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create view")
		return
	}

	location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/views"), sharedAPI.ResourceID(result.CreatedID))
	view, err := h.readModel.GetByID(r.Context(), result.CreatedID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created view")
		return
	}

	if view == nil {
		sharedAPI.RespondCreated(w, location, map[string]string{
			"id":      result.CreatedID,
			"message": "View created, processing",
		})
		return
	}

	view.Links = h.buildViewLinks(r, view)
	sharedAPI.RespondCreated(w, location, view)
}

// GetAllViews godoc
// @Summary Get all architecture views
// @Description Retrieves all architecture views
// @Tags views
// @Produce json
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_architectureviews_application_readmodels.ArchitectureViewDTO}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views [get]
func (h *ViewHandlers) GetAllViews(w http.ResponseWriter, r *http.Request) {
	views, err := h.readModel.GetAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve views")
		return
	}

	for i := range views {
		views[i].Links = h.buildViewLinks(r, &views[i])
		AddElementLinks(&views[i], h.canEditView(r, &views[i]))
	}

	links := sharedAPI.Links{
		"self": sharedAPI.NewLink("/api/v1/views", "GET"),
	}

	sharedAPI.RespondCollection(w, http.StatusOK, views, links)
}

// GetViewByID godoc
// @Summary Get an architecture view by ID
// @Description Retrieves a specific architecture view by its ID with all component positions
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Success 200 {object} readmodels.ArchitectureViewDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id} [get]
func (h *ViewHandlers) GetViewByID(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	view, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		h.errorHandler.HandleError(w, err, "Failed to retrieve view")
		return
	}

	if view == nil {
		h.errorHandler.HandleNotFound(w, "View")
		return
	}

	view.Links = h.buildViewLinks(r, view)
	AddElementLinks(view, h.canEditView(r, view))

	sharedAPI.RespondJSON(w, http.StatusOK, view)
}

// RenameView godoc
// @Summary Rename an architecture view
// @Description Renames an architecture view
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param view body RenameViewRequest true "New view name"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/name [patch]
func (h *ViewHandlers) RenameView(w http.ResponseWriter, r *http.Request) {
	handleViewSetting(h, w, r, "name", viewSettingUpdate[valueobjects.ViewName]{
		validate: valueobjects.NewViewName,
		createCmd: func(viewID string, name valueobjects.ViewName) cqrs.Command {
			return &commands.RenameView{ViewID: viewID, NewName: name.String()}
		},
	})
}

// DeleteView godoc
// @Summary Delete an architecture view
// @Description Deletes an architecture view (cannot delete default view)
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse "Cannot delete default view"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id} [delete]
func (h *ViewHandlers) DeleteView(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	h.dispatchCommand(w, r, &commands.DeleteView{
		ViewID: viewID,
	})
}

// ChangeVisibility godoc
// @Summary Change view visibility
// @Description Toggle view between private and public
// @Tags views
// @Accept json
// @Param id path string true "View ID"
// @Param request body ChangeVisibilityRequest true "Visibility request"
// @Success 204
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/visibility [patch]
func (h *ViewHandlers) ChangeVisibility(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[ChangeVisibilityRequest](w, r)
	if !ok {
		return
	}

	h.dispatchCommand(w, r, &commands.ChangeViewVisibility{
		ViewID:    viewID,
		IsPrivate: req.IsPrivate,
	})
}

// SetDefaultView godoc
// @Summary Set a view as the default view
// @Description Sets an architecture view as the default view
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/default [put]
func (h *ViewHandlers) SetDefaultView(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	h.dispatchCommand(w, r, &commands.SetDefaultView{
		ViewID: viewID,
	})
}

// UpdateEdgeType godoc
// @Summary Update edge type for a view
// @Description Updates the edge rendering style for an architecture view
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param edgeType body object{edgeType=string} true "Edge type update"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/edge-type [patch]
func (h *ViewHandlers) UpdateEdgeType(w http.ResponseWriter, r *http.Request) {
	handleViewSetting(h, w, r, "edgeType", viewSettingUpdate[valueobjects.EdgeType]{
		validate: valueobjects.NewEdgeType,
		createCmd: func(viewID string, edgeType valueobjects.EdgeType) cqrs.Command {
			return &commands.UpdateViewEdgeType{ViewID: viewID, EdgeType: edgeType.String()}
		},
	})
}

// UpdateLayoutDirection godoc
// @Summary Update layout direction for a view
// @Description Updates the auto-layout direction for an architecture view
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param layoutDirection body object{layoutDirection=string} true "Layout direction update"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/layout-direction [patch]
func (h *ViewHandlers) UpdateLayoutDirection(w http.ResponseWriter, r *http.Request) {
	handleViewSetting(h, w, r, "layoutDirection", viewSettingUpdate[valueobjects.LayoutDirection]{
		validate: valueobjects.NewLayoutDirection,
		createCmd: func(viewID string, layoutDir valueobjects.LayoutDirection) cqrs.Command {
			return &commands.UpdateViewLayoutDirection{ViewID: viewID, LayoutDirection: layoutDir.String()}
		},
	})
}

func (h *ViewHandlers) buildViewLinks(r *http.Request, view *readmodels.ArchitectureViewDTO) sharedAPI.Links {
	actor, _ := sharedctx.GetActor(r.Context())
	viewInfo := ViewInfo{
		ID:          view.ID,
		IsPrivate:   view.IsPrivate,
		IsDefault:   view.IsDefault,
		OwnerUserID: view.OwnerUserID,
	}
	return h.hateoas.ViewLinksForActor(viewInfo, actor)
}

func isOwnerOfView(ownerUserID *string, actorID string) bool {
	return ownerUserID != nil && *ownerUserID == actorID
}

func (h *ViewHandlers) canEditView(r *http.Request, view *readmodels.ArchitectureViewDTO) bool {
	actor, _ := sharedctx.GetActor(r.Context())
	isOwner := view.OwnerUserID != nil && *view.OwnerUserID == actor.ID
	canEditThisView := !view.IsPrivate || isOwner
	return canEditThisView && actor.CanWrite("views")
}

func (h *ViewHandlers) checkViewEditPermission(w http.ResponseWriter, r *http.Request, viewID string) bool {
	actor, ok := sharedctx.GetActor(r.Context())
	if !ok {
		sharedAPI.RespondError(w, http.StatusUnauthorized, nil, "Authentication required")
		return false
	}

	authInfo, err := h.readModel.GetAuthInfo(r.Context(), viewID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to check permissions")
		return false
	}
	if authInfo == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "View not found")
		return false
	}

	if !authInfo.IsPrivate {
		return true
	}

	if !isOwnerOfView(authInfo.OwnerUserID, actor.ID) {
		sharedAPI.RespondError(w, http.StatusForbidden, nil, "Access denied")
		return false
	}

	return true
}
