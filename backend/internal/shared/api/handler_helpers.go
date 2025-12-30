package api

import (
	"net/http"

	"easi/backend/internal/shared/cqrs"
)

func HandleCommandResult(w http.ResponseWriter, result cqrs.CommandResult, err error, successHandler func(createdID string)) {
	if err != nil {
		HandleError(w, err)
		return
	}
	successHandler(result.CreatedID)
}

func HandleError(w http.ResponseWriter, err error) {
	statusCode, message, found := globalRegistry.Lookup(err)
	if found {
		RespondError(w, statusCode, err, message)
		return
	}

	statusCode = MapErrorToStatusCode(err, http.StatusInternalServerError)
	RespondError(w, statusCode, nil, "An unexpected error occurred")
}

func HandleErrorWithDefault(w http.ResponseWriter, err error, defaultMessage string) {
	statusCode, message, found := globalRegistry.Lookup(err)
	if found {
		RespondError(w, statusCode, err, message)
		return
	}

	statusCode = MapErrorToStatusCode(err, http.StatusInternalServerError)
	RespondError(w, statusCode, nil, defaultMessage)
}
