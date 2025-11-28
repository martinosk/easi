package api

import (
	"net/http"
	"strings"
)

type ErrorHandler struct {
	defaultMessage string
}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		defaultMessage: "An error occurred processing your request",
	}
}

func (h *ErrorHandler) HandleError(w http.ResponseWriter, err error, context string) {
	statusCode, message := h.mapError(err, context)
	RespondError(w, statusCode, err, message)
}

func (h *ErrorHandler) HandleValidationError(w http.ResponseWriter, err error) {
	RespondError(w, http.StatusBadRequest, err, "")
}

func (h *ErrorHandler) HandleNotFound(w http.ResponseWriter, resource string) {
	RespondError(w, http.StatusNotFound, nil, resource+" not found")
}

func (h *ErrorHandler) HandleConflict(w http.ResponseWriter, err error, message string) {
	if message == "" {
		message = "Resource conflict"
	}
	RespondError(w, http.StatusConflict, err, message)
}

func (h *ErrorHandler) mapError(err error, context string) (int, string) {
	if err == nil {
		return http.StatusInternalServerError, h.defaultMessage
	}

	errMsg := strings.ToLower(err.Error())

	switch {
	case containsAny(errMsg, "not found"):
		return http.StatusNotFound, ""
	case h.isConflictMessage(errMsg):
		return http.StatusConflict, ""
	case h.isValidationMessage(errMsg):
		return http.StatusBadRequest, ""
	case containsAny(errMsg, "unauthorized"):
		return http.StatusUnauthorized, ""
	case containsAny(errMsg, "forbidden", "permission"):
		return http.StatusForbidden, ""
	default:
		return h.defaultErrorResponse(context)
	}
}

func (h *ErrorHandler) isConflictMessage(errMsg string) bool {
	return containsAny(errMsg, "already exists", "already in", "duplicate", "cannot delete", "has children", "in use")
}

func (h *ErrorHandler) isValidationMessage(errMsg string) bool {
	return containsAny(errMsg, "invalid", "cannot be empty", "too long", "too short", "must be", "validation")
}

func (h *ErrorHandler) defaultErrorResponse(context string) (int, string) {
	if context != "" {
		return http.StatusInternalServerError, context
	}
	return http.StatusInternalServerError, h.defaultMessage
}

func containsAny(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

func (h *ErrorHandler) IsNotFoundError(err error) bool {
	return h.matchesErrorPattern(err, "not found")
}

func (h *ErrorHandler) IsConflictError(err error) bool {
	return h.matchesErrorPattern(err, "already exists", "duplicate", "conflict")
}

func (h *ErrorHandler) IsValidationError(err error) bool {
	return h.matchesErrorPattern(err, "invalid", "validation", "must be", "cannot be")
}

func (h *ErrorHandler) matchesErrorPattern(err error, patterns ...string) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return containsAny(errMsg, patterns...)
}