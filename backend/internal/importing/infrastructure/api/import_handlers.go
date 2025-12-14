package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"easi/backend/internal/importing/application/commands"
	"easi/backend/internal/importing/application/parsers"
	"easi/backend/internal/importing/application/readmodels"
	"easi/backend/internal/importing/domain/aggregates"
	"easi/backend/internal/importing/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
)

const (
	maxFileSize = 50 << 20
)

type ImportHandlers struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.ImportSessionReadModel
	parser     *parsers.ArchiMateParser
}

func NewImportHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ImportSessionReadModel,
) *ImportHandlers {
	return &ImportHandlers{
		commandBus: commandBus,
		readModel:  readModel,
		parser:     parsers.NewArchiMateParser(),
	}
}

// CreateImportSession godoc
// @Summary Create an import session
// @Description Uploads an ArchiMate Open Exchange XML file and creates a new import session for preview
// @Tags imports
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "ArchiMate XML file"
// @Param sourceFormat formData string true "Source format (e.g., 'archimate')"
// @Param businessDomainId formData string false "Target business domain ID"
// @Success 201 {object} readmodels.ImportSessionDTO "Import session created"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid request or missing required fields"
// @Failure 413 {object} sharedAPI.ErrorResponse "File exceeds maximum size"
// @Failure 415 {object} sharedAPI.ErrorResponse "Unsupported media type"
// @Failure 422 {object} sharedAPI.ErrorResponse "Invalid ArchiMate format"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /api/v1/imports [post]
func (h *ImportHandlers) CreateImportSession(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			sharedAPI.RespondError(w, http.StatusRequestEntityTooLarge, err, "File exceeds maximum size of 50MB")
			return
		}
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid multipart form data")
		return
	}

	sourceFormat := r.FormValue("sourceFormat")
	if sourceFormat == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "sourceFormat is required")
		return
	}

	if _, err := valueobjects.NewSourceFormat(sourceFormat); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, err.Error())
		return
	}

	businessDomainID := r.FormValue("businessDomainId")

	file, header, err := r.FormFile("file")
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "file is required")
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xml") {
		sharedAPI.RespondError(w, http.StatusUnsupportedMediaType, nil, "File must be an XML file")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "xml") && !strings.Contains(contentType, "application/octet-stream") {
		sharedAPI.RespondError(w, http.StatusUnsupportedMediaType, nil, "File must be an XML file")
		return
	}

	parseResult, err := h.parser.Parse(file)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnprocessableEntity, err, "Invalid ArchiMate Open Exchange format")
		return
	}

	cmd := &commands.CreateImportSession{
		SourceFormat:     sourceFormat,
		BusinessDomainID: businessDomainID,
		ParseResult:      parseResult,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create import session")
		return
	}

	session, err := h.readModel.GetByID(r.Context(), cmd.ID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve import session")
		return
	}

	session.Links = h.getLinksForStatus(cmd.ID, session.Status)

	location := fmt.Sprintf("/api/v1/imports/%s", cmd.ID)
	w.Header().Set("Location", location)
	sharedAPI.RespondJSON(w, http.StatusCreated, session)
}

// GetImportSession godoc
// @Summary Get an import session
// @Description Retrieves the details of an import session by ID
// @Tags imports
// @Produce json
// @Param id path string true "Import session ID"
// @Success 200 {object} readmodels.ImportSessionDTO "Import session details"
// @Failure 400 {object} sharedAPI.ErrorResponse "Missing import session ID"
// @Failure 404 {object} sharedAPI.ErrorResponse "Import session not found"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /api/v1/imports/{id} [get]
func (h *ImportHandlers) GetImportSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Import session ID is required")
		return
	}

	session, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve import session")
		return
	}

	if session == nil {
		sharedAPI.RespondErrorWithLinks(w, http.StatusNotFound, nil, "Import session not found", map[string]sharedAPI.Link{
			"create": {Href: "/api/v1/imports", Method: "POST"},
		})
		return
	}

	session.Links = h.getLinksForStatus(id, session.Status)
	sharedAPI.RespondJSON(w, http.StatusOK, session)
}

// ConfirmImport godoc
// @Summary Confirm an import session
// @Description Confirms and starts processing an import session
// @Tags imports
// @Produce json
// @Param id path string true "Import session ID"
// @Success 202 {object} readmodels.ImportSessionDTO "Import confirmed and processing started"
// @Failure 400 {object} sharedAPI.ErrorResponse "Missing import session ID"
// @Failure 404 {object} sharedAPI.ErrorResponse "Import session not found"
// @Failure 409 {object} sharedAPI.ErrorResponse "Import already started or completed"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /api/v1/imports/{id}/confirm [post]
func (h *ImportHandlers) ConfirmImport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Import session ID is required")
		return
	}

	session, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve import session")
		return
	}

	if session == nil {
		sharedAPI.RespondErrorWithLinks(w, http.StatusNotFound, nil, "Import session not found", map[string]sharedAPI.Link{
			"create": {Href: "/api/v1/imports", Method: "POST"},
		})
		return
	}

	baseURL := fmt.Sprintf("/api/v1/imports/%s", id)
	if session.Status != "pending" {
		sharedAPI.RespondErrorWithLinks(w, http.StatusConflict, aggregates.ErrImportAlreadyStarted, "Import session has already been started or completed", map[string]sharedAPI.Link{
			"self": {Href: baseURL},
		})
		return
	}

	cmd := &commands.ConfirmImport{ID: id}
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if errors.Is(err, aggregates.ErrImportAlreadyStarted) {
			sharedAPI.RespondError(w, http.StatusConflict, err, "Import session has already been started")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to confirm import")
		return
	}

	session, err = h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve import session")
		return
	}

	session.Links = h.getLinksForStatus(id, session.Status)

	w.Header().Set("Retry-After", "2")
	sharedAPI.RespondJSON(w, http.StatusAccepted, session)
}

// DeleteImportSession godoc
// @Summary Cancel an import session
// @Description Cancels a pending import session. Cannot cancel imports that have already started or completed.
// @Tags imports
// @Param id path string true "Import session ID"
// @Success 204 "Import session cancelled"
// @Failure 400 {object} sharedAPI.ErrorResponse "Missing import session ID"
// @Failure 404 {object} sharedAPI.ErrorResponse "Import session not found"
// @Failure 409 {object} sharedAPI.ErrorResponse "Cannot cancel import that has already started or completed"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /api/v1/imports/{id} [delete]
func (h *ImportHandlers) DeleteImportSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Import session ID is required")
		return
	}

	session, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve import session")
		return
	}

	if session == nil {
		sharedAPI.RespondErrorWithLinks(w, http.StatusNotFound, nil, "Import session not found", map[string]sharedAPI.Link{
			"create": {Href: "/api/v1/imports", Method: "POST"},
		})
		return
	}

	baseURL := fmt.Sprintf("/api/v1/imports/%s", id)
	if session.Status != "pending" {
		sharedAPI.RespondErrorWithLinks(w, http.StatusConflict, nil, "Cannot cancel import that has already started or completed", map[string]sharedAPI.Link{
			"self": {Href: baseURL},
		})
		return
	}

	cmd := &commands.CancelImport{ID: id}
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if errors.Is(err, aggregates.ErrCannotCancelStartedImport) {
			sharedAPI.RespondError(w, http.StatusConflict, err, "Cannot cancel import that has already started")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to cancel import")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ImportHandlers) getLinksForStatus(id, status string) map[string]sharedAPI.Link {
	baseURL := fmt.Sprintf("/api/v1/imports/%s", id)
	links := map[string]sharedAPI.Link{
		"self": {Href: baseURL},
	}

	if status == "pending" {
		links["confirm"] = sharedAPI.Link{Href: baseURL + "/confirm", Method: "POST"}
		links["delete"] = sharedAPI.Link{Href: baseURL, Method: "DELETE"}
	}

	return links
}
