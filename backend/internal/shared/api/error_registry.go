package api

import (
	"errors"
	"net/http"
	"sync"
)

type ErrorCategory int

const (
	CategoryNotFound ErrorCategory = iota
	CategoryValidation
	CategoryConflict
	CategoryForbidden
	CategoryUnauthorized
	CategoryInternal
)

func (c ErrorCategory) StatusCode() int {
	switch c {
	case CategoryNotFound:
		return http.StatusNotFound
	case CategoryValidation:
		return http.StatusBadRequest
	case CategoryConflict:
		return http.StatusConflict
	case CategoryForbidden:
		return http.StatusForbidden
	case CategoryUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

type ErrorRegistration struct {
	Error    error
	Category ErrorCategory
	Message  string
}

type ErrorRegistry struct {
	mu            sync.RWMutex
	registrations []ErrorRegistration
}

var globalRegistry = &ErrorRegistry{
	registrations: []ErrorRegistration{},
}

func GetErrorRegistry() *ErrorRegistry {
	return globalRegistry
}

func (r *ErrorRegistry) Register(reg ErrorRegistration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registrations = append(r.registrations, reg)
}

func (r *ErrorRegistry) RegisterNotFound(err error, msg string) {
	r.Register(ErrorRegistration{Error: err, Category: CategoryNotFound, Message: msg})
}

func (r *ErrorRegistry) RegisterValidation(err error, msg string) {
	r.Register(ErrorRegistration{Error: err, Category: CategoryValidation, Message: msg})
}

func (r *ErrorRegistry) RegisterConflict(err error, msg string) {
	r.Register(ErrorRegistration{Error: err, Category: CategoryConflict, Message: msg})
}

func (r *ErrorRegistry) Lookup(err error) (int, string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, reg := range r.registrations {
		if errors.Is(err, reg.Error) {
			return reg.Category.StatusCode(), reg.Message, true
		}
	}
	return 0, "", false
}
