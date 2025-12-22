package api

import (
	"net/http"
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
	if statusCode, message, found := globalRegistry.Lookup(err); found {
		RespondError(w, statusCode, err, message)
		return
	}

	statusCode := MapErrorToStatusCode(err, http.StatusInternalServerError)
	message := ""
	if statusCode == http.StatusInternalServerError && context != "" {
		message = context
	}
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
