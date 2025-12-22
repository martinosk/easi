package api

import (
	"errors"
	"net/http"
)

type ErrorMapping struct {
	Error      error
	StatusCode int
	Message    string
}

type DomainErrorMapper struct {
	mappings []ErrorMapping
	handler  *ErrorHandler
}

func NewDomainErrorMapper() *DomainErrorMapper {
	return &DomainErrorMapper{
		mappings: []ErrorMapping{},
		handler:  NewErrorHandler(),
	}
}

func (m *DomainErrorMapper) AddMapping(err error, statusCode int, message string) *DomainErrorMapper {
	m.mappings = append(m.mappings, ErrorMapping{
		Error:      err,
		StatusCode: statusCode,
		Message:    message,
	})
	return m
}

func (m *DomainErrorMapper) HandleError(w http.ResponseWriter, err error, defaultContext string) {
	for _, mapping := range m.mappings {
		if errors.Is(err, mapping.Error) {
			RespondError(w, mapping.StatusCode, err, mapping.Message)
			return
		}
	}

	m.handler.HandleError(w, err, defaultContext)
}

func (m *DomainErrorMapper) MapToStatus(err error) (int, string) {
	for _, mapping := range m.mappings {
		if errors.Is(err, mapping.Error) {
			return mapping.StatusCode, mapping.Message
		}
	}

	if statusCode, message, found := globalRegistry.Lookup(err); found {
		return statusCode, message
	}

	return MapErrorToStatusCode(err, http.StatusInternalServerError), ""
}
