package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/session"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
)

type UserHandlers struct {
	commandBus       cqrs.CommandBus
	userReadModel    *readmodels.UserReadModel
	sessionManager   *session.SessionManager
	paginationHelper *sharedAPI.PaginationHelper
}

func NewUserHandlers(
	commandBus cqrs.CommandBus,
	userReadModel *readmodels.UserReadModel,
	sessionManager *session.SessionManager,
) *UserHandlers {
	return &UserHandlers{
		commandBus:       commandBus,
		userReadModel:    userReadModel,
		sessionManager:   sessionManager,
		paginationHelper: sharedAPI.NewPaginationHelper("/api/v1/users"),
	}
}

type UserResponse struct {
	ID          string             `json:"id"`
	Email       string             `json:"email"`
	Name        *string            `json:"name,omitempty"`
	Role        string             `json:"role"`
	Status      string             `json:"status"`
	CreatedAt   time.Time          `json:"createdAt"`
	LastLoginAt *time.Time         `json:"lastLoginAt,omitempty"`
	InvitedBy   *InvitedByInfo     `json:"invitedBy,omitempty"`
	Links       map[string]string  `json:"_links"`
}

type InvitedByInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type UpdateUserRequest struct {
	Role   *string `json:"role,omitempty"`
	Status *string `json:"status,omitempty"`
}

// GetAllUsers godoc
// @Summary List all users
// @Description Returns a paginated list of all users in the current tenant with optional filtering
// @Tags users
// @Accept json
// @Produce json
// @Param after query string false "Pagination cursor for next page"
// @Param limit query int false "Number of items per page (default 50, max 100)"
// @Param status query string false "Filter by status (active, disabled)"
// @Param role query string false "Filter by role (admin, member)"
// @Success 200 {object} sharedAPI.PaginatedResponse{data=[]UserResponse} "Paginated list of users with HATEOAS links"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid pagination cursor"
// @Failure 401 {object} sharedAPI.ErrorResponse "Not authenticated"
// @Failure 403 {object} sharedAPI.ErrorResponse "Insufficient permissions"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /users [get]
func (h *UserHandlers) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	params := sharedAPI.ParsePaginationParams(r)
	statusFilter := r.URL.Query().Get("status")
	roleFilter := r.URL.Query().Get("role")

	afterCursor, afterTimestamp, err := h.paginationHelper.ProcessCursor(params.After)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pagination cursor")
		return
	}

	currentUserID, err := h.getCurrentUserID(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Failed to get current user")
		return
	}

	users, hasMore, err := h.userReadModel.GetAllPaginated(r.Context(), readmodels.UserPaginationFilter{
		Limit:          params.Limit,
		AfterCursor:    afterCursor,
		AfterTimestamp: afterTimestamp,
		StatusFilter:   statusFilter,
		RoleFilter:     roleFilter,
	})
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve users")
		return
	}

	responses := make([]UserResponse, len(users))
	for i, user := range users {
		isCurrentUser := user.ID.String() == currentUserID
		isLastAdmin, _ := h.userReadModel.IsLastActiveAdmin(r.Context(), user.ID.String())
		responses[i] = h.toUserResponse(user, isCurrentUser, isLastAdmin)
	}

	pageables := h.convertToPageable(users)
	nextCursor := h.paginationHelper.GenerateNextCursor(pageables, hasMore)
	selfLink := h.paginationHelper.BuildSelfLink(params)

	sharedAPI.RespondPaginated(w, sharedAPI.PaginatedResponseParams{
		StatusCode: http.StatusOK,
		Data:       responses,
		HasMore:    hasMore,
		NextCursor: nextCursor,
		Limit:      params.Limit,
		SelfLink:   selfLink,
		BaseLink:   "/api/v1/users",
	})
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Returns a single user by their unique identifier
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID format)"
// @Success 200 {object} UserResponse "User details with HATEOAS links"
// @Failure 401 {object} sharedAPI.ErrorResponse "Not authenticated"
// @Failure 403 {object} sharedAPI.ErrorResponse "Insufficient permissions"
// @Failure 404 {object} sharedAPI.ErrorResponse "User not found"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /users/{id} [get]
func (h *UserHandlers) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.userReadModel.GetByIDString(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve user")
		return
	}
	if user == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "User not found")
		return
	}

	currentUserID, err := h.getCurrentUserID(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Failed to get current user")
		return
	}

	isCurrentUser := user.ID.String() == currentUserID
	isLastAdmin, _ := h.userReadModel.IsLastActiveAdmin(r.Context(), user.ID.String())

	response := h.toUserResponse(*user, isCurrentUser, isLastAdmin)
	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

// UpdateUser godoc
// @Summary Update user
// @Description Updates user properties (role and/or status). Cannot demote the last admin or disable yourself.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID format)"
// @Param request body UpdateUserRequest true "Fields to update"
// @Success 200 {object} UserResponse "Updated user with HATEOAS links"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid request body, role, or status"
// @Failure 401 {object} sharedAPI.ErrorResponse "Not authenticated"
// @Failure 403 {object} sharedAPI.ErrorResponse "Insufficient permissions"
// @Failure 404 {object} sharedAPI.ErrorResponse "User not found"
// @Failure 409 {object} sharedAPI.ErrorResponse "Business rule violation (last admin, already disabled, etc.)"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /users/{id} [patch]
func (h *UserHandlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	if req.Role == nil && req.Status == nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "At least one field (role or status) must be provided")
		return
	}

	currentUserID, err := h.getCurrentUserID(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Failed to get current user")
		return
	}

	if req.Role != nil {
		if _, err := valueobjects.RoleFromString(*req.Role); err != nil {
			sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid role")
			return
		}

		cmd := &commands.ChangeUserRole{
			UserID:      id,
			NewRole:     *req.Role,
			ChangedByID: currentUserID,
		}
		if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
			h.handleCommandError(w, err, "Failed to change user role")
			return
		}
	}

	if req.Status != nil {
		switch *req.Status {
		case "disabled":
			cmd := &commands.DisableUser{
				UserID:       id,
				DisabledByID: currentUserID,
			}
			if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
				h.handleCommandError(w, err, "Failed to disable user")
				return
			}
		case "active":
			cmd := &commands.EnableUser{
				UserID:      id,
				EnabledByID: currentUserID,
			}
			if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
				h.handleCommandError(w, err, "Failed to enable user")
				return
			}
		default:
			sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Invalid status. Must be 'active' or 'disabled'")
			return
		}
	}

	h.respondWithUpdatedUser(w, r, id)
}

func (h *UserHandlers) getCurrentUserID(r *http.Request) (string, error) {
	authSession, err := h.sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil {
		return "", err
	}
	return authSession.UserID().String(), nil
}

func (h *UserHandlers) handleCommandError(w http.ResponseWriter, err error, defaultMessage string) {
	switch {
	case errors.Is(err, repositories.ErrUserAggregateNotFound):
		sharedAPI.RespondError(w, http.StatusNotFound, err, "User not found")
	case errors.Is(err, aggregates.ErrCannotDisableSelf):
		sharedAPI.RespondError(w, http.StatusConflict, err, "Cannot disable your own account")
	case errors.Is(err, aggregates.ErrCannotDemoteLastAdmin):
		sharedAPI.RespondError(w, http.StatusConflict, err, "Cannot demote the last admin in tenant")
	case errors.Is(err, aggregates.ErrCannotDisableLastAdmin):
		sharedAPI.RespondError(w, http.StatusConflict, err, "Cannot disable the last admin in tenant")
	case errors.Is(err, aggregates.ErrUserAlreadyActive):
		sharedAPI.RespondError(w, http.StatusConflict, err, "User is already active")
	case errors.Is(err, aggregates.ErrUserAlreadyDisabled):
		sharedAPI.RespondError(w, http.StatusConflict, err, "User is already disabled")
	case errors.Is(err, aggregates.ErrSameRole):
		sharedAPI.RespondError(w, http.StatusConflict, err, "User already has this role")
	case errors.Is(err, valueobjects.ErrInvalidRole):
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid role")
	default:
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, defaultMessage)
	}
}

func (h *UserHandlers) respondWithUpdatedUser(w http.ResponseWriter, r *http.Request, id string) {
	user, err := h.userReadModel.GetByIDString(r.Context(), id)
	if err != nil || user == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve updated user")
		return
	}

	currentUserID, _ := h.getCurrentUserID(r)
	isCurrentUser := user.ID.String() == currentUserID
	isLastAdmin, _ := h.userReadModel.IsLastActiveAdmin(r.Context(), user.ID.String())

	response := h.toUserResponse(*user, isCurrentUser, isLastAdmin)
	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

func (h *UserHandlers) toUserResponse(user readmodels.UserDTO, isCurrentUser, isLastAdmin bool) UserResponse {
	return UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		Name:        user.Name,
		Role:        user.Role,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt,
		LastLoginAt: user.LastLoginAt,
		Links:       h.userLinks(user.ID.String(), user.Status, user.Role, isCurrentUser, isLastAdmin),
	}
}

func (h *UserHandlers) userLinks(userID, status, role string, isCurrentUser, isLastAdmin bool) map[string]string {
	links := map[string]string{
		"self": fmt.Sprintf("/api/v1/users/%s", userID),
	}

	if isCurrentUser {
		return links
	}

	isAdmin := role == "admin"
	canModify := !isLastAdmin || !isAdmin

	if canModify {
		links["update"] = fmt.Sprintf("/api/v1/users/%s", userID)
	}

	return links
}

func (h *UserHandlers) convertToPageable(users []readmodels.UserDTO) []sharedAPI.Pageable {
	pageables := make([]sharedAPI.Pageable, len(users))
	for i := range users {
		pageables[i] = &userPageable{dto: &users[i]}
	}
	return pageables
}

type userPageable struct {
	dto *readmodels.UserDTO
}

func (p *userPageable) GetID() string {
	return p.dto.ID.String()
}

func (p *userPageable) GetTimestamp() time.Time {
	return p.dto.CreatedAt
}
