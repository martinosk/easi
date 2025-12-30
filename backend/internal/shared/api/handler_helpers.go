package api

import (
	"net/http"
)

func HandleCommandResult(w http.ResponseWriter, err error, successHandler func()) {
	if err != nil {
		HandleError(w, err)
		return
	}
	successHandler()
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
