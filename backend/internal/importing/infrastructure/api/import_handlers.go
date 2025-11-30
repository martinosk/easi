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
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Import session not found")
		return
	}

	session.Links = h.getLinksForStatus(id, session.Status)
	sharedAPI.RespondJSON(w, http.StatusOK, session)
}

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
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Import session not found")
		return
	}

	if session.Status != "pending" {
		sharedAPI.RespondError(w, http.StatusConflict, aggregates.ErrImportAlreadyStarted, "Import session has already been started or completed")
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
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Import session not found")
		return
	}

	if session.Status != "pending" {
		sharedAPI.RespondError(w, http.StatusConflict, nil, "Cannot cancel import that has already started or completed")
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

func (h *ImportHandlers) getLinksForStatus(id, status string) map[string]string {
	baseURL := fmt.Sprintf("/api/v1/imports/%s", id)
	links := map[string]string{
		"self": baseURL,
	}

	if status == "pending" {
		links["confirm"] = baseURL + "/confirm"
		links["delete"] = baseURL
	}

	return links
}
