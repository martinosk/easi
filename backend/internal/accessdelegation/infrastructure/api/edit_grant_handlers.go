package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"easi/backend/internal/accessdelegation/application/commands"
	"easi/backend/internal/accessdelegation/application/readmodels"
	"easi/backend/internal/accessdelegation/application/services"
	"easi/backend/internal/accessdelegation/domain/valueobjects"
	authPL "easi/backend/internal/auth/publishedlanguage"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
)

const granteeInviteeRole = "stakeholder"

type EditGrantReader interface {
	HasActiveGrant(ctx context.Context, granteeEmail, artifactType, artifactID string) (bool, error)
	GetByID(ctx context.Context, id string) (*readmodels.EditGrantDTO, error)
	GetByGranteeEmail(ctx context.Context, email string) ([]readmodels.EditGrantDTO, error)
	GetActiveForArtifact(ctx context.Context, artifactType, artifactID string) ([]readmodels.EditGrantDTO, error)
}

type EditGrantHandlerDeps struct {
	CommandBus   cqrs.CommandBus
	ReadModel    EditGrantReader
	Hateoas      *EditGrantLinks
	NameResolver services.ArtifactNameResolver
}

type EditGrantHandlers struct {
	deps EditGrantHandlerDeps
}

func NewEditGrantHandlers(deps EditGrantHandlerDeps) *EditGrantHandlers {
	return &EditGrantHandlers{deps: deps}
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
	req, ok := h.decodeCreateEditGrantRequest(w, r)
	if !ok {
		return
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

	result, err := h.deps.CommandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	log.Printf("[AUDIT] edit-grant-created grantor=%s grantee=%s artifact-type=%s artifact-id=%s reason=%s", actor.ID, req.GranteeEmail, req.ArtifactType, req.ArtifactID, req.Reason)

	grant := h.fetchAndEnrichCreatedGrant(r.Context(), result.CreatedID, actor)
	if grant != nil {
		grant.InvitationCreated = h.autoInviteIfNeeded(r.Context(), req.GranteeEmail, actor)
	}
	h.respondCreated(w, result.CreatedID, grant)
}

func (h *EditGrantHandlers) decodeCreateEditGrantRequest(w http.ResponseWriter, r *http.Request) (CreateEditGrantRequest, bool) {
	var req CreateEditGrantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return req, false
	}

	granteeEmail, err := valueobjects.NewGranteeEmail(req.GranteeEmail)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid grantee email")
		return req, false
	}
	req.GranteeEmail = granteeEmail.Value()

	if req.Scope == "" {
		req.Scope = "write"
	}
	return req, true
}

func (h *EditGrantHandlers) canGrantEditAccess(actor sharedctx.Actor, artifactType string) bool {
	return actor.CanWrite(sharedctx.PluralResourceName(artifactType)) || actor.HasPermission("edit-grants:manage")
}

func (h *EditGrantHandlers) ensureNoActiveGrant(w http.ResponseWriter, r *http.Request, req CreateEditGrantRequest) error {
	exists, err := h.deps.ReadModel.HasActiveGrant(r.Context(), req.GranteeEmail, req.ArtifactType, req.ArtifactID)
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

func (h *EditGrantHandlers) fetchAndEnrichCreatedGrant(ctx context.Context, id string, actor sharedctx.Actor) *readmodels.EditGrantDTO {
	grant, err := h.deps.ReadModel.GetByID(ctx, id)
	if err != nil || grant == nil {
		return nil
	}
	h.enrichSingleGrant(ctx, grant, actor)
	return grant
}

func (h *EditGrantHandlers) respondCreated(w http.ResponseWriter, id string, grant *readmodels.EditGrantDTO) {
	if grant != nil {
		w.Header().Set("Location", grant.Links["self"].Href)
		sharedAPI.RespondJSON(w, http.StatusCreated, grant)
		return
	}
	sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{"id": id})
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

	grants, err := h.deps.ReadModel.GetByGranteeEmail(r.Context(), actor.Email)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve edit grants")
		return
	}

	h.enrichGrantsWithArtifactNames(r.Context(), grants)
	h.enrichGrantsWithLinks(grants, actor)
	sharedAPI.RespondCollection(w, http.StatusOK, grants, h.deps.Hateoas.EditGrantCollectionLinksForActor(actor))
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

	actor, ok := sharedctx.GetActor(r.Context())
	if !ok {
		sharedAPI.RespondError(w, http.StatusUnauthorized, nil, "Unauthorized")
		return
	}

	grant, err := h.deps.ReadModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve edit grant")
		return
	}
	if grant == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Edit grant not found")
		return
	}

	if !canViewGrant(grant, actor) {
		sharedAPI.RespondError(w, http.StatusForbidden, nil, "You do not have permission to view this grant")
		return
	}

	h.enrichSingleGrant(r.Context(), grant, actor)
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

	grant := h.getGrantOrFail(w, r, id)
	if grant == nil {
		return
	}

	if !canRevokeEditGrant(*grant, actor) {
		sharedAPI.RespondError(w, http.StatusForbidden, nil, "Only the grantor or an admin can revoke this grant")
		return
	}

	cmd := &commands.RevokeEditGrant{ID: id, RevokedBy: actor.ID}
	if _, err := h.deps.CommandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	log.Printf("[AUDIT] edit-grant-revoked actor=%s grant-id=%s", actor.ID, id)
	w.WriteHeader(http.StatusNoContent)
}

func canViewGrant(grant *readmodels.EditGrantDTO, actor sharedctx.Actor) bool {
	return grant.GrantorID == actor.ID || grant.GranteeEmail == actor.Email || actor.HasPermission("edit-grants:manage")
}

func (h *EditGrantHandlers) getGrantOrFail(w http.ResponseWriter, r *http.Request, id string) *readmodels.EditGrantDTO {
	grant, err := h.deps.ReadModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve edit grant")
		return nil
	}
	if grant == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Edit grant not found")
		return nil
	}
	return grant
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

	if !actor.CanWrite(sharedctx.PluralResourceName(artifactType)) && !actor.HasPermission("edit-grants:manage") {
		sharedAPI.RespondError(w, http.StatusForbidden, nil, "You do not have permission to view grants for this artifact")
		return
	}

	grants, err := h.deps.ReadModel.GetActiveForArtifact(r.Context(), artifactType, artifactID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve edit grants")
		return
	}

	h.enrichGrantsWithArtifactNames(r.Context(), grants)
	h.enrichGrantsWithLinks(grants, actor)
	sharedAPI.RespondCollection(w, http.StatusOK, grants, h.deps.Hateoas.EditGrantArtifactCollectionLinks(artifactType, artifactID))
}

func (h *EditGrantHandlers) enrichGrantsWithLinks(grants []readmodels.EditGrantDTO, actor sharedctx.Actor) {
	for i := range grants {
		grants[i].Links = h.deps.Hateoas.EditGrantLinksForActor(grants[i], actor)
		h.deps.Hateoas.AddArtifactLink(grants[i].Links, grants[i])
	}
}

func (h *EditGrantHandlers) enrichGrantsWithArtifactNames(ctx context.Context, grants []readmodels.EditGrantDTO) {
	if h.deps.NameResolver == nil {
		return
	}
	namesByType := make(map[string]map[string]string, len(grants))
	for artifactType, artifactIDs := range groupArtifactIDsByType(grants) {
		names, _ := h.deps.NameResolver.ResolveNames(ctx, artifactType, artifactIDs)
		namesByType[artifactType] = names
	}
	for i := range grants {
		grants[i].ArtifactName = namesByType[grants[i].ArtifactType][grants[i].ArtifactID]
	}
}

func groupArtifactIDsByType(grants []readmodels.EditGrantDTO) map[string][]string {
	idsByType := make(map[string][]string)
	seen := make(map[string]bool, len(grants))
	for i := range grants {
		key := grants[i].ArtifactType + "\x00" + grants[i].ArtifactID
		if seen[key] {
			continue
		}
		seen[key] = true
		idsByType[grants[i].ArtifactType] = append(idsByType[grants[i].ArtifactType], grants[i].ArtifactID)
	}
	return idsByType
}

func (h *EditGrantHandlers) enrichSingleGrant(ctx context.Context, grant *readmodels.EditGrantDTO, actor sharedctx.Actor) {
	if h.deps.NameResolver != nil {
		name, _ := h.deps.NameResolver.ResolveName(ctx, grant.ArtifactType, grant.ArtifactID)
		grant.ArtifactName = name
	}
	grant.Links = h.deps.Hateoas.EditGrantLinksForActor(*grant, actor)
	h.deps.Hateoas.AddArtifactLink(grant.Links, *grant)
}

func (h *EditGrantHandlers) autoInviteIfNeeded(ctx context.Context, granteeEmail string, actor sharedctx.Actor) bool {
	cmd := &authPL.EnsureInvitation{
		Email:        granteeEmail,
		Role:         granteeInviteeRole,
		InviterID:    actor.ID,
		InviterEmail: actor.Email,
	}

	result, err := h.deps.CommandBus.Dispatch(ctx, cmd)
	if err != nil {
		log.Printf("[WARN] edit-grant auto-invite: ensure invitation for %s: %v", granteeEmail, err)
		return false
	}
	if result.CreatedID == "" {
		return false
	}

	log.Printf("[AUDIT] invitation-auto-created email=%s inviter=%s reason=edit-grant", granteeEmail, actor.Email)
	return true
}
