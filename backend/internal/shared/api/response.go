package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"easi/backend/internal/shared/domain"
)

// Link represents a HATEOAS link with optional HTTP method hint
type Link struct {
	Href   string `json:"href"`
	Method string `json:"method,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
	Links   map[string]Link   `json:"_links,omitempty"`
}

// CollectionResponse represents a collection of resources
type CollectionResponse struct {
	Data  interface{}       `json:"data"`
	Links map[string]string `json:"_links,omitempty"`
	Meta  *CollectionMeta   `json:"meta,omitempty"`
}

// CollectionMeta contains metadata about a collection
type CollectionMeta struct {
	Total *int `json:"total,omitempty"` // Total count if available
}

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// RespondError sends an error response with proper status code based on error type
func RespondError(w http.ResponseWriter, statusCode int, err error, message string) {
	// Override status code based on error type if applicable
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

	// Add field-specific details for validation errors
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
		Links:   params.Links,
	}
	if params.Err != nil && params.Message == "" {
		response.Message = params.Err.Error()
	}

	RespondJSON(w, statusCode, response)
}

// MapErrorToStatusCode maps domain errors to appropriate HTTP status codes
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
	case errors.Is(err, domain.ErrInvalidOperation):
		return http.StatusConflict
	case errors.As(err, &domain.ValidationError{}):
		return http.StatusBadRequest
	default:
		return defaultCode
	}
}

// RespondNoContent sends a 204 No Content response
func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// RespondCreated sends a 201 Created response with Location header and body
func RespondCreated(w http.ResponseWriter, location string, data interface{}) {
	w.Header().Set("Location", location)
	RespondJSON(w, http.StatusCreated, data)
}

// RespondCreatedNoBody sends a 201 Created response with Location header only
func RespondCreatedNoBody(w http.ResponseWriter, location string) {
	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusCreated)
}

// RespondDeleted sends a 204 No Content response for successful deletions
func RespondDeleted(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// RespondCollection sends a collection response with consistent wrapping
func RespondCollection(w http.ResponseWriter, statusCode int, data interface{}, links map[string]string) {
	response := CollectionResponse{
		Data:  data,
		Links: links,
	}
	RespondJSON(w, statusCode, response)
}

type CollectionWithTotalParams struct {
	Data       interface{}
	Total      int
	Links      map[string]string
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
