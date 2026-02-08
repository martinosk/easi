package api

import (
	"net/http"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type InternalTeamHandlers struct {
	commandBus       cqrs.CommandBus
	readModel        *readmodels.InternalTeamReadModel
	paginationHelper *sharedAPI.PaginationHelper
	hateoas          *ArchitectureModelingLinks
}

func NewInternalTeamHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.InternalTeamReadModel,
	hateoas *ArchitectureModelingLinks,
) *InternalTeamHandlers {
	return &InternalTeamHandlers{
		commandBus:       commandBus,
		readModel:        readModel,
		paginationHelper: sharedAPI.NewPaginationHelper("/api/v1/internal-teams"),
		hateoas:          hateoas,
	}
}

type CreateInternalTeamRequest struct {
	Name          string `json:"name"`
	Department    string `json:"department,omitempty"`
	ContactPerson string `json:"contactPerson,omitempty"`
	Notes         string `json:"notes,omitempty"`
}

type UpdateInternalTeamRequest struct {
	Name          string `json:"name"`
	Department    string `json:"department,omitempty"`
	ContactPerson string `json:"contactPerson,omitempty"`
	Notes         string `json:"notes,omitempty"`
}

// CreateInternalTeam godoc
// @Summary Create a new internal team
// @Description Creates a new internal team (in-house development team)
// @Tags internal-teams
// @Accept json
// @Produce json
// @Param team body CreateInternalTeamRequest true "Internal team data"
// @Success 201 {object} readmodels.InternalTeamDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /internal-teams [post]
func (h *InternalTeamHandlers) CreateInternalTeam(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[CreateInternalTeamRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.CreateInternalTeam{
		Name:          req.Name,
		Department:    req.Department,
		ContactPerson: req.ContactPerson,
		Notes:         req.Notes,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/internal-teams"), sharedAPI.ResourceID(result.CreatedID))
	team, err := h.readModel.GetByID(r.Context(), result.CreatedID)
	if err != nil {
		sharedAPI.HandleErrorWithDefault(w, err, "Failed to retrieve created team")
		return
	}

	if team == nil {
		sharedAPI.RespondCreated(w, location, map[string]string{
			"id":      result.CreatedID,
			"message": "Team created, processing",
		})
		return
	}

	h.enrichWithLinks(r, team)
	sharedAPI.RespondCreated(w, location, team)
}

// GetAllInternalTeams godoc
// @Summary Get all internal teams
// @Description Retrieves all internal teams with cursor-based pagination
// @Tags internal-teams
// @Produce json
// @Param limit query int false "Number of items per page (max 100)" default(50)
// @Param after query string false "Cursor for pagination"
// @Success 200 {object} sharedAPI.PaginatedResponse{data=[]readmodels.InternalTeamDTO}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /internal-teams [get]
func (h *InternalTeamHandlers) GetAllInternalTeams(w http.ResponseWriter, r *http.Request) {
	params := sharedAPI.ParsePaginationParams(r)

	afterID, afterName, err := h.paginationHelper.ProcessNameCursor(params.After)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pagination cursor")
		return
	}

	teams, hasMore, err := h.readModel.GetAllPaginated(r.Context(), params.Limit, afterID, afterName)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve internal teams")
		return
	}

	for i := range teams {
		h.enrichWithLinks(r, &teams[i])
	}

	pageables := ConvertInternalTeamsToNamePageable(teams)
	nextCursor := h.paginationHelper.GenerateNextNameCursor(pageables, hasMore)
	selfLink := h.paginationHelper.BuildSelfLink(params)

	sharedAPI.RespondPaginated(w, sharedAPI.PaginatedResponseParams{
		StatusCode: http.StatusOK,
		Data:       teams,
		HasMore:    hasMore,
		NextCursor: nextCursor,
		Limit:      params.Limit,
		SelfLink:   selfLink,
		BaseLink:   "/api/v1/internal-teams",
	})
}

// GetInternalTeamByID godoc
// @Summary Get an internal team by ID
// @Description Retrieves a specific internal team by its ID
// @Tags internal-teams
// @Produce json
// @Param id path string true "Internal team ID"
// @Success 200 {object} readmodels.InternalTeamDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /internal-teams/{id} [get]
func (h *InternalTeamHandlers) GetInternalTeamByID(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	team, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve internal team")
		return
	}

	if team == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Internal team not found")
		return
	}

	h.enrichWithLinks(r, team)
	sharedAPI.RespondJSON(w, http.StatusOK, team)
}

// UpdateInternalTeam godoc
// @Summary Update an internal team
// @Description Updates an existing internal team
// @Tags internal-teams
// @Accept json
// @Produce json
// @Param id path string true "Internal team ID"
// @Param team body UpdateInternalTeamRequest true "Updated team data"
// @Success 200 {object} readmodels.InternalTeamDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /internal-teams/{id} [put]
func (h *InternalTeamHandlers) UpdateInternalTeam(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateInternalTeamRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.UpdateInternalTeam{
		ID:            id,
		Name:          req.Name,
		Department:    req.Department,
		ContactPerson: req.ContactPerson,
		Notes:         req.Notes,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	team, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.HandleErrorWithDefault(w, err, "Failed to retrieve updated team")
		return
	}

	if team == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Internal team not found")
		return
	}

	h.enrichWithLinks(r, team)
	sharedAPI.RespondJSON(w, http.StatusOK, team)
}

// DeleteInternalTeam godoc
// @Summary Delete an internal team
// @Description Permanently deletes an internal team
// @Tags internal-teams
// @Produce json
// @Param id path string true "Internal team ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /internal-teams/{id} [delete]
func (h *InternalTeamHandlers) DeleteInternalTeam(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	cmd := &commands.DeleteInternalTeam{
		ID: id,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		sharedAPI.RespondDeleted(w)
	})
}

func (h *InternalTeamHandlers) enrichWithLinks(r *http.Request, team *readmodels.InternalTeamDTO) {
	actor, _ := sharedctx.GetActor(r.Context())
	team.Links = h.hateoas.InternalTeamLinksForActor(team.ID, actor)
}
