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
	case strings.Contains(errMsg, "not found"):
		return http.StatusNotFound, ""

	case strings.Contains(errMsg, "already exists") ||
	     strings.Contains(errMsg, "already in") ||
	     strings.Contains(errMsg, "duplicate"):
		return http.StatusConflict, ""

	case strings.Contains(errMsg, "invalid") ||
	     strings.Contains(errMsg, "cannot be empty") ||
	     strings.Contains(errMsg, "too long") ||
	     strings.Contains(errMsg, "too short") ||
	     strings.Contains(errMsg, "must be") ||
	     strings.Contains(errMsg, "validation"):
		return http.StatusBadRequest, ""

	case strings.Contains(errMsg, "unauthorized"):
		return http.StatusUnauthorized, ""

	case strings.Contains(errMsg, "forbidden") ||
	     strings.Contains(errMsg, "permission"):
		return http.StatusForbidden, ""

	case strings.Contains(errMsg, "cannot delete") ||
	     strings.Contains(errMsg, "has children") ||
	     strings.Contains(errMsg, "in use"):
		return http.StatusConflict, ""

	default:
		if context != "" {
			return http.StatusInternalServerError, context
		}
		return http.StatusInternalServerError, h.defaultMessage
	}
}

func (h *ErrorHandler) IsNotFoundError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "not found")
}

func (h *ErrorHandler) IsConflictError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "already exists") ||
	       strings.Contains(errMsg, "duplicate") ||
	       strings.Contains(errMsg, "conflict")
}

func (h *ErrorHandler) IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "invalid") ||
	       strings.Contains(errMsg, "validation") ||
	       strings.Contains(errMsg, "must be") ||
	       strings.Contains(errMsg, "cannot be")
}