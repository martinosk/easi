package api

import (
	"context"
	"net/http"

	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

type LayoutRepository interface {
	UpsertElementPosition(ctx context.Context, ref repositories.ElementRef, pos repositories.Position) error
	DeleteElementPosition(ctx context.Context, ref repositories.ElementRef) error
}

type ViewElementHandlers struct {
	layoutRepo   LayoutRepository
	readModel    *readmodels.ArchitectureViewReadModel
	errorHandler *sharedAPI.ErrorHandler
}

func NewViewElementHandlers(
	layoutRepo LayoutRepository,
	readModel *readmodels.ArchitectureViewReadModel,
) *ViewElementHandlers {
	return &ViewElementHandlers{
		layoutRepo:   layoutRepo,
		readModel:    readModel,
		errorHandler: sharedAPI.NewErrorHandler(),
	}
}

func (h *ViewElementHandlers) checkViewEditPermission(w http.ResponseWriter, r *http.Request, viewID string) bool {
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

type elementOp struct {
	pathParam string
	errorMsg  string
}

type ElementPositionRequest struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func withViewElement[T any](h *ViewElementHandlers, w http.ResponseWriter, r *http.Request, op elementOp, decode func() (T, bool), execute func(ctx context.Context, viewID, elementID string, payload T) error) {
	viewID := sharedAPI.GetPathParam(r, "id")
	elementID := sharedAPI.GetPathParam(r, op.pathParam)

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	payload, ok := decode()
	if !ok {
		return
	}

	if err := execute(r.Context(), viewID, elementID, payload); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, op.errorMsg)
		return
	}

	sharedAPI.RespondNoContent(w)
}

func (h *ViewElementHandlers) updateElementPosition(w http.ResponseWriter, r *http.Request, op elementOp, elementType repositories.ElementType) {
	withViewElement(h, w, r, op,
		func() (ElementPositionRequest, bool) {
			return sharedAPI.DecodeRequestOrFail[ElementPositionRequest](w, r)
		},
		func(ctx context.Context, viewID, elementID string, req ElementPositionRequest) error {
			ref := repositories.ElementRef{ViewID: viewID, ElementID: elementID, ElementType: elementType}
			return h.layoutRepo.UpsertElementPosition(ctx, ref, repositories.Position{X: req.X, Y: req.Y})
		})
}

func (h *ViewElementHandlers) removeElement(w http.ResponseWriter, r *http.Request, op elementOp, elementType repositories.ElementType) {
	withViewElement(h, w, r, op,
		func() (struct{}, bool) { return struct{}{}, true },
		func(ctx context.Context, viewID, elementID string, _ struct{}) error {
			return h.layoutRepo.DeleteElementPosition(ctx, repositories.ElementRef{ViewID: viewID, ElementID: elementID, ElementType: elementType})
		})
}

type addElementConfig struct {
	fieldName   string
	subPath     string
	errorMsg    string
	elementType repositories.ElementType
}

type addElementRequest interface {
	entityID() string
	position() repositories.Position
}

type AddCapabilityRequest struct {
	CapabilityID string  `json:"capabilityId"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
}

func (r AddCapabilityRequest) entityID() string                  { return r.CapabilityID }
func (r AddCapabilityRequest) position() repositories.Position { return repositories.Position{X: r.X, Y: r.Y} }

type AddOriginEntityRequest struct {
	OriginEntityID string  `json:"originEntityId"`
	X              float64 `json:"x"`
	Y              float64 `json:"y"`
}

func (r AddOriginEntityRequest) entityID() string                  { return r.OriginEntityID }
func (r AddOriginEntityRequest) position() repositories.Position { return repositories.Position{X: r.X, Y: r.Y} }

func handleAddElement[T addElementRequest](h *ViewElementHandlers, w http.ResponseWriter, r *http.Request, cfg addElementConfig) {
	viewID := sharedAPI.GetPathParam(r, "id")
	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[T](w, r)
	if !ok {
		return
	}

	if req.entityID() == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, cfg.fieldName+" is required")
		return
	}

	ref := repositories.ElementRef{ViewID: viewID, ElementID: req.entityID(), ElementType: cfg.elementType}
	if err := h.layoutRepo.UpsertElementPosition(r.Context(), ref, req.position()); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, cfg.errorMsg)
		return
	}

	location := sharedAPI.BuildSubResourceLink(sharedAPI.ResourcePath("/views"), sharedAPI.ResourceID(viewID), sharedAPI.ResourcePath(cfg.subPath))
	sharedAPI.RespondCreatedNoBody(w, location)
}

// AddCapabilityToView godoc
// @Summary Add a capability to a view
// @Description Adds a capability node to an architecture view at the specified position
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param capability body AddCapabilityRequest true "Capability to add with position"
// @Success 201
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/capabilities [post]
func (h *ViewElementHandlers) AddCapabilityToView(w http.ResponseWriter, r *http.Request) {
	handleAddElement[AddCapabilityRequest](h, w, r, addElementConfig{
		fieldName:   "capabilityId",
		subPath:     "/capabilities",
		errorMsg:    "Failed to add capability to view",
		elementType: repositories.ElementTypeCapability,
	})
}

// UpdateCapabilityPosition godoc
// @Summary Update capability position in a view
// @Description Updates the position of a capability node in an architecture view
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param capabilityId path string true "Capability ID"
// @Param position body ElementPositionRequest true "New position"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/capabilities/{capabilityId}/position [patch]
func (h *ViewElementHandlers) UpdateCapabilityPosition(w http.ResponseWriter, r *http.Request) {
	h.updateElementPosition(w, r, elementOp{pathParam: "capabilityId", errorMsg: "Failed to update capability position"}, repositories.ElementTypeCapability)
}

// RemoveCapabilityFromView godoc
// @Summary Remove a capability from a view
// @Description Removes a capability node from an architecture view
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Param capabilityId path string true "Capability ID"
// @Success 204
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/capabilities/{capabilityId} [delete]
func (h *ViewElementHandlers) RemoveCapabilityFromView(w http.ResponseWriter, r *http.Request) {
	h.removeElement(w, r, elementOp{pathParam: "capabilityId", errorMsg: "Failed to remove capability from view"}, repositories.ElementTypeCapability)
}

// AddOriginEntityToView godoc
// @Summary Add an origin entity to a view
// @Description Adds an origin entity node to an architecture view at the specified position
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param originEntity body AddOriginEntityRequest true "Origin entity to add with position"
// @Success 201
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/origin-entities [post]
func (h *ViewElementHandlers) AddOriginEntityToView(w http.ResponseWriter, r *http.Request) {
	handleAddElement[AddOriginEntityRequest](h, w, r, addElementConfig{
		fieldName:   "originEntityId",
		subPath:     "/origin-entities",
		errorMsg:    "Failed to add origin entity to view",
		elementType: repositories.ElementTypeOriginEntity,
	})
}

// UpdateOriginEntityPosition godoc
// @Summary Update origin entity position in a view
// @Description Updates the position of an origin entity node in an architecture view
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param originEntityId path string true "Origin Entity ID"
// @Param position body ElementPositionRequest true "New position"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/origin-entities/{originEntityId}/position [patch]
func (h *ViewElementHandlers) UpdateOriginEntityPosition(w http.ResponseWriter, r *http.Request) {
	h.updateElementPosition(w, r, elementOp{pathParam: "originEntityId", errorMsg: "Failed to update origin entity position"}, repositories.ElementTypeOriginEntity)
}

// RemoveOriginEntityFromView godoc
// @Summary Remove an origin entity from a view
// @Description Removes an origin entity node from an architecture view
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Param originEntityId path string true "Origin Entity ID"
// @Success 204
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/origin-entities/{originEntityId} [delete]
func (h *ViewElementHandlers) RemoveOriginEntityFromView(w http.ResponseWriter, r *http.Request) {
	h.removeElement(w, r, elementOp{pathParam: "originEntityId", errorMsg: "Failed to remove origin entity from view"}, repositories.ElementTypeOriginEntity)
}
