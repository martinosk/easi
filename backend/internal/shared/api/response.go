package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/easi/backend/internal/shared/domain"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Data  interface{}       `json:"data,omitempty"`
	Links map[string]string `json:"_links,omitempty"`
}

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
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

// MapErrorToStatusCode maps domain errors to appropriate HTTP status codes
func MapErrorToStatusCode(err error, defaultCode int) int {
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

// RespondSuccess sends a success response with optional HATEOAS links
func RespondSuccess(w http.ResponseWriter, statusCode int, data interface{}, links map[string]string) {
	response := SuccessResponse{
		Data:  data,
		Links: links,
	}
	RespondJSON(w, statusCode, response)
}

// RespondCreated sends a 201 Created response with Location header
func RespondCreated(w http.ResponseWriter, location string, data interface{}, links map[string]string) {
	w.Header().Set("Location", location)
	RespondSuccess(w, http.StatusCreated, data, links)
}

// RespondNoContent sends a 204 No Content response
func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
