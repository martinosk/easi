package api

import (
	"encoding/json"
	"errors"
	"net/http"

	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/types"
)

type Link = types.Link

type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
	Links   map[string]Link   `json:"_links,omitempty"`
}

type CollectionResponse struct {
	Data  any             `json:"data"`
	Links Links           `json:"_links,omitempty"`
	Meta  *CollectionMeta `json:"meta,omitempty"`
}

type CollectionMeta struct {
	Total *int `json:"total,omitempty"`
}

func RespondJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func RespondError(w http.ResponseWriter, statusCode int, err error, message string) {
	if err != nil {
		statusCode = MapErrorToStatusCode(err, statusCode)
	}

	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}
	if err != nil && message == "" {
		response.Message = err.Error()
	}

	var valErr domain.ValidationError
	if errors.As(err, &valErr) && valErr.Field != "" {
		response.Details = map[string]string{
			valErr.Field: valErr.Message,
		}
	}

	RespondJSON(w, statusCode, response)
}

type ErrorWithLinksParams struct {
	StatusCode int
	Err        error
	Message    string
	Details    map[string]string
	Links      map[string]Link
}

func RespondErrorWithLinks(w http.ResponseWriter, params ErrorWithLinksParams) {
	statusCode := params.StatusCode
	if params.Err != nil {
		statusCode = MapErrorToStatusCode(params.Err, statusCode)
	}

	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: params.Message,
		Details: params.Details,
		Links:   params.Links,
	}
	if params.Err != nil && params.Message == "" {
		response.Message = params.Err.Error()
	}

	RespondJSON(w, statusCode, response)
}

func MapErrorToStatusCode(err error, defaultCode int) int {
	if statusCode, _, found := globalRegistry.Lookup(err); found {
		return statusCode
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrDuplicateResource):
		return http.StatusConflict
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, domain.ErrConcurrencyConflict):
		return http.StatusPreconditionFailed
	case errors.Is(err, domain.ErrInvalidOperation):
		return http.StatusConflict
	case errors.As(err, &domain.ValidationError{}):
		return http.StatusBadRequest
	default:
		return defaultCode
	}
}

func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func RespondCreated(w http.ResponseWriter, location string, data any) {
	w.Header().Set("Location", location)
	RespondJSON(w, http.StatusCreated, data)
}

func RespondCreatedNoBody(w http.ResponseWriter, location string) {
	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusCreated)
}

func RespondDeleted(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func RespondCollection(w http.ResponseWriter, statusCode int, data any, links Links) {
	response := CollectionResponse{
		Data:  data,
		Links: links,
	}
	RespondJSON(w, statusCode, response)
}

type CollectionWithTotalParams struct {
	Data       any
	Total      int
	Links      Links
	StatusCode int
}

func RespondCollectionWithTotal(w http.ResponseWriter, params CollectionWithTotalParams) {
	response := CollectionResponse{
		Data:  params.Data,
		Links: params.Links,
		Meta: &CollectionMeta{
			Total: &params.Total,
		},
	}
	RespondJSON(w, params.StatusCode, response)
}
