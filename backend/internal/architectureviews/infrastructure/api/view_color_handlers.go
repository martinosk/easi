package api

import (
	"fmt"
	"net/http"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/types"
)

type ViewColorHandlers struct {
	commandBus   cqrs.CommandBus
	readModel    *readmodels.ArchitectureViewReadModel
	hateoas      *ViewLinks
	errorHandler *sharedAPI.ErrorHandler
}

func NewViewColorHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ArchitectureViewReadModel,
	hateoas *ViewLinks,
) *ViewColorHandlers {
	return &ViewColorHandlers{
		commandBus:   commandBus,
		readModel:    readModel,
		hateoas:      hateoas,
		errorHandler: sharedAPI.NewErrorHandler(),
	}
}

func (h *ViewColorHandlers) checkViewEditPermission(w http.ResponseWriter, r *http.Request, viewID string) bool {
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

type elementParams struct {
	viewID      string
	elementID   string
	elementType string
}

type UpdateElementColorRequest struct {
	Color string `json:"color" example:"#FF5733"`
}

func (h *ViewColorHandlers) updateElementColor(w http.ResponseWriter, r *http.Request, params elementParams) {
	if !h.checkViewEditPermission(w, r, params.viewID) {
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateElementColorRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.UpdateElementColor{
		ViewID:      params.viewID,
		ElementID:   params.elementID,
		ElementType: params.elementType,
		Color:       req.Color,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	sharedAPI.RespondNoContent(w)
}

func (h *ViewColorHandlers) clearElementColor(w http.ResponseWriter, r *http.Request, params elementParams) {
	if !h.checkViewEditPermission(w, r, params.viewID) {
		return
	}

	cmd := &commands.ClearElementColor{
		ViewID:      params.viewID,
		ElementID:   params.elementID,
		ElementType: params.elementType,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, fmt.Sprintf("Failed to clear %s color", params.elementType))
		return
	}

	sharedAPI.RespondNoContent(w)
}

type UpdateColorSchemeRequest struct {
	ColorScheme string `json:"colorScheme"`
}

// UpdateColorScheme godoc
// @Summary Update color scheme for a view
// @Description Updates the color scheme for an architecture view. Valid schemes: maturity, classic, custom
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param colorScheme body UpdateColorSchemeRequest true "Color scheme update request"
// @Success 200 {object} object{colorScheme=string,_links=map[string]string}
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/color-scheme [patch]
func (h *ViewColorHandlers) UpdateColorScheme(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateColorSchemeRequest](w, r)
	if !ok {
		return
	}

	if _, err := valueobjects.NewColorScheme(req.ColorScheme); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	if _, err := h.commandBus.Dispatch(r.Context(), &commands.UpdateViewColorScheme{
		ViewID:      viewID,
		ColorScheme: req.ColorScheme,
	}); err != nil {
		h.errorHandler.HandleError(w, err, "Failed to update color scheme")
		return
	}

	links := sharedAPI.NewResourceLinks().
		SelfSubResource(sharedAPI.ResourcePath("/views"), sharedAPI.ResourceID(viewID), sharedAPI.ResourcePath("/color-scheme")).
		Related(sharedAPI.LinkRelation("view"), sharedAPI.ResourcePath("/views"), sharedAPI.ResourceID(viewID)).
		Build()

	response := struct {
		ColorScheme string      `json:"colorScheme"`
		Links       types.Links `json:"_links"`
	}{
		ColorScheme: req.ColorScheme,
		Links:       links,
	}

	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

// UpdateComponentColor godoc
// @Summary Update custom color for a component in a view
// @Description Sets a custom hex color for a component when using the custom color scheme
// @Tags views
// @Accept json
// @Param id path string true "View ID"
// @Param componentId path string true "Component ID"
// @Param color body UpdateElementColorRequest true "Color update request with hex color (e.g., #FF5733)"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid hex color format"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/components/{componentId}/color [patch]
func (h *ViewColorHandlers) UpdateComponentColor(w http.ResponseWriter, r *http.Request) {
	h.updateElementColor(w, r, elementParams{
		viewID:      sharedAPI.GetPathParam(r, "id"),
		elementID:   sharedAPI.GetPathParam(r, "componentId"),
		elementType: "component",
	})
}

// UpdateCapabilityColor godoc
// @Summary Update custom color for a capability in a view
// @Description Sets a custom hex color for a capability when using the custom color scheme
// @Tags views
// @Accept json
// @Param id path string true "View ID"
// @Param capabilityId path string true "Capability ID"
// @Param color body UpdateElementColorRequest true "Color update request with hex color (e.g., #FF5733)"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid hex color format"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/capabilities/{capabilityId}/color [patch]
func (h *ViewColorHandlers) UpdateCapabilityColor(w http.ResponseWriter, r *http.Request) {
	h.updateElementColor(w, r, elementParams{
		viewID:      sharedAPI.GetPathParam(r, "id"),
		elementID:   sharedAPI.GetPathParam(r, "capabilityId"),
		elementType: "capability",
	})
}

// ClearComponentColor godoc
// @Summary Clear custom color for a component in a view
// @Description Removes the custom color from a component, returning it to the default color scheme
// @Tags views
// @Param id path string true "View ID"
// @Param componentId path string true "Component ID"
// @Success 204
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/components/{componentId}/color [delete]
func (h *ViewColorHandlers) ClearComponentColor(w http.ResponseWriter, r *http.Request) {
	h.clearElementColor(w, r, elementParams{
		viewID:      sharedAPI.GetPathParam(r, "id"),
		elementID:   sharedAPI.GetPathParam(r, "componentId"),
		elementType: "component",
	})
}

// ClearCapabilityColor godoc
// @Summary Clear custom color for a capability in a view
// @Description Removes the custom color from a capability, returning it to the default color scheme
// @Tags views
// @Param id path string true "View ID"
// @Param capabilityId path string true "Capability ID"
// @Success 204
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/capabilities/{capabilityId}/color [delete]
func (h *ViewColorHandlers) ClearCapabilityColor(w http.ResponseWriter, r *http.Request) {
	h.clearElementColor(w, r, elementParams{
		viewID:      sharedAPI.GetPathParam(r, "id"),
		elementID:   sharedAPI.GetPathParam(r, "capabilityId"),
		elementType: "capability",
	})
}

func BuildElementLinks(viewID, elementType, elementID string, canEdit bool) sharedAPI.Links {
	links := sharedAPI.Links{}
	if !canEdit {
		return links
	}
	basePath := fmt.Sprintf("/api/v1/views/%s/%s/%s", viewID, elementType, elementID)
	links["x-update-color"] = sharedAPI.NewLink(basePath+"/color", "PATCH")
	links["x-clear-color"] = sharedAPI.NewLink(basePath+"/color", "DELETE")
	links["x-update-position"] = sharedAPI.NewLink(basePath+"/position", "PATCH")
	links["x-remove"] = sharedAPI.NewLink(basePath, "DELETE")
	return links
}

func BuildOriginEntityLinks(viewID, originEntityID string, canEdit bool) sharedAPI.Links {
	links := sharedAPI.Links{}
	if !canEdit {
		return links
	}
	basePath := fmt.Sprintf("/api/v1/views/%s/origin-entities/%s", viewID, originEntityID)
	links["x-update-position"] = sharedAPI.NewLink(basePath+"/position", "PATCH")
	links["x-remove"] = sharedAPI.NewLink(basePath, "DELETE")
	return links
}

func AddElementLinks(view *readmodels.ArchitectureViewDTO, canEdit bool) {
	for i := range view.Components {
		view.Components[i].Links = BuildElementLinks(view.ID, "components", view.Components[i].ComponentID, canEdit)
	}

	for i := range view.Capabilities {
		view.Capabilities[i].Links = BuildElementLinks(view.ID, "capabilities", view.Capabilities[i].CapabilityID, canEdit)
	}

	for i := range view.OriginEntities {
		view.OriginEntities[i].Links = BuildOriginEntityLinks(view.ID, view.OriginEntities[i].OriginEntityID, canEdit)
	}
}
