package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"easi/backend/internal/accessdelegation/application/commands"
	"easi/backend/internal/accessdelegation/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
)

type EditGrantHandlers struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.EditGrantReadModel
	hateoas    *sharedAPI.HATEOASLinks
}

func NewEditGrantHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.EditGrantReadModel,
	hateoas *sharedAPI.HATEOASLinks,
) *EditGrantHandlers {
	return &EditGrantHandlers{
		commandBus: commandBus,
		readModel:  readModel,
		hateoas:    hateoas,
	}
}

type CreateEditGrantRequest struct {
	GranteeEmail string `json:"granteeEmail"`
	ArtifactType string `json:"artifactType"`
	ArtifactID   string `json:"artifactId"`
	Scope        string `json:"scope"`
	Reason       string `json:"reason"`
}

// CreateEditGrant godoc
// @Summary Create a new edit grant
// @Description Grants temporary edit access to a specific artifact for another user
// @Tags edit-grants
// @Accept json
// @Produce json
// @Param grant body CreateEditGrantRequest true "Edit grant data"
// @Success 201 {object} easi_backend_internal_accessdelegation_application_readmodels.EditGrantDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse "Active grant already exists"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /edit-grants [post]
func (h *EditGrantHandlers) CreateEditGrant(w http.ResponseWriter, r *http.Request) {
	var req CreateEditGrantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	if req.Scope == "" {
		req.Scope = "write"
	}

	actor, ok := sharedctx.GetActor(r.Context())
	if !ok {
		sharedAPI.RespondError(w, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	if !h.canGrantEditAccess(actor, req.ArtifactType) {
		sharedAPI.RespondError(w, http.StatusForbidden, nil, "You do not have permission to grant edit access for this artifact type")
		return
	}

	if err := h.ensureNoActiveGrant(w, r, req); err != nil {
		return
	}

	cmd := &commands.CreateEditGrant{
		GrantorID:    actor.ID,
		GrantorEmail: actor.Email,
		GranteeEmail: req.GranteeEmail,
		ArtifactType: req.ArtifactType,
		ArtifactID:   req.ArtifactID,
		Scope:        req.Scope,
		Reason:       req.Reason,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	h.respondCreated(w, r, result.CreatedID, actor)
}

func (h *EditGrantHandlers) canGrantEditAccess(actor sharedctx.Actor, artifactType string) bool {
	return actor.CanWrite(artifactType+"s") || actor.HasPermission("edit-grants:manage")
}

func (h *EditGrantHandlers) ensureNoActiveGrant(w http.ResponseWriter, r *http.Request, req CreateEditGrantRequest) error {
	exists, err := h.readModel.ExistsActiveGrant(r.Context(), req.GranteeEmail, req.ArtifactType, req.ArtifactID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to check existing grants")
		return err
	}
	if exists {
		sharedAPI.RespondError(w, http.StatusConflict, nil, "An active edit grant already exists for this user and artifact")
		return fmt.Errorf("active grant exists")
	}
	return nil
}

func (h *EditGrantHandlers) respondCreated(w http.ResponseWriter, r *http.Request, id string, actor sharedctx.Actor) {
	w.Header().Set("Location", h.hateoas.EditGrantLinksForActor(id, "active", actor.ID, actor)["self"].Href)

	grant, err := h.readModel.GetByID(r.Context(), id)
	if err != nil || grant == nil {
		sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{"id": id})
		return
	}

	grant.Links = h.hateoas.EditGrantLinksForActor(grant.ID, grant.Status, grant.GrantorID, actor)
	sharedAPI.RespondJSON(w, http.StatusCreated, grant)
}

// GetMyEditGrants godoc
// @Summary Get my active edit grants
// @Description Retrieves all active edit grants where the current user is the grantee
// @Tags edit-grants
// @Produce json
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_accessdelegation_application_readmodels.EditGrantDTO}
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /edit-grants [get]
func (h *EditGrantHandlers) GetMyEditGrants(w http.ResponseWriter, r *http.Request) {
	actor, ok := sharedctx.GetActor(r.Context())
	if !ok {
		sharedAPI.RespondError(w, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	grants, err := h.readModel.GetByGranteeEmail(r.Context(), actor.Email)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve edit grants")
		return
	}

	h.enrichGrantsWithLinks(grants, actor)
	sharedAPI.RespondCollection(w, http.StatusOK, grants, h.hateoas.EditGrantCollectionLinksForActor(actor))
}

// GetEditGrantByID godoc
// @Summary Get an edit grant by ID
// @Description Retrieves a specific edit grant by its ID
// @Tags edit-grants
// @Produce json
// @Param id path string true "Edit Grant ID"
// @Success 200 {object} easi_backend_internal_accessdelegation_application_readmodels.EditGrantDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /edit-grants/{id} [get]
func (h *EditGrantHandlers) GetEditGrantByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	grant, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve edit grant")
		return
	}
	if grant == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Edit grant not found")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	grant.Links = h.hateoas.EditGrantLinksForActor(grant.ID, grant.Status, grant.GrantorID, actor)
	sharedAPI.RespondJSON(w, http.StatusOK, grant)
}

// RevokeEditGrant godoc
// @Summary Revoke an edit grant
// @Description Revokes an active edit grant. Only the grantor or an admin can revoke a grant.
// @Tags edit-grants
// @Produce json
// @Param id path string true "Edit Grant ID"
// @Success 204 "No Content"
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse "Grant already revoked or expired"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /edit-grants/{id} [delete]
func (h *EditGrantHandlers) RevokeEditGrant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	actor, ok := sharedctx.GetActor(r.Context())
	if !ok {
		sharedAPI.RespondError(w, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	grant, err := h.getGrantOrFail(w, r, id)
	if grant == nil {
		return
	}

	if !canRevokeGrant(grant, actor) {
		sharedAPI.RespondError(w, http.StatusForbidden, nil, "Only the grantor or an admin can revoke this grant")
		return
	}

	cmd := &commands.RevokeEditGrant{ID: id, RevokedBy: actor.ID}
	if _, err = h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func canRevokeGrant(grant *readmodels.EditGrantDTO, actor sharedctx.Actor) bool {
	return grant.GrantorID == actor.ID || actor.Role == "admin"
}

func (h *EditGrantHandlers) getGrantOrFail(w http.ResponseWriter, r *http.Request, id string) (*readmodels.EditGrantDTO, error) {
	grant, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve edit grant")
		return nil, err
	}
	if grant == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Edit grant not found")
		return nil, fmt.Errorf("not found")
	}
	return grant, nil
}

// GetEditGrantsForArtifact godoc
// @Summary Get edit grants for an artifact
// @Description Retrieves all active edit grants for a specific artifact
// @Tags edit-grants
// @Produce json
// @Param artifactType path string true "Artifact type (e.g., capability, component, view)"
// @Param artifactId path string true "Artifact ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_accessdelegation_application_readmodels.EditGrantDTO}
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /edit-grants/artifact/{artifactType}/{artifactId} [get]
func (h *EditGrantHandlers) GetEditGrantsForArtifact(w http.ResponseWriter, r *http.Request) {
	artifactType := chi.URLParam(r, "artifactType")
	artifactID := chi.URLParam(r, "artifactId")

	actor, ok := sharedctx.GetActor(r.Context())
	if !ok {
		sharedAPI.RespondError(w, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	grants, err := h.readModel.GetActiveForArtifact(r.Context(), artifactType, artifactID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve edit grants")
		return
	}

	h.enrichGrantsWithLinks(grants, actor)
	sharedAPI.RespondCollection(w, http.StatusOK, grants, h.hateoas.EditGrantArtifactCollectionLinks(artifactType, artifactID))
}

func (h *EditGrantHandlers) enrichGrantsWithLinks(grants []readmodels.EditGrantDTO, actor sharedctx.Actor) {
	for i := range grants {
		grants[i].Links = h.hateoas.EditGrantLinksForActor(grants[i].ID, grants[i].Status, grants[i].GrantorID, actor)
	}
}
