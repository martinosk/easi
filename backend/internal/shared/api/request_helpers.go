package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func DecodeRequest[T any](r *http.Request) (T, error) {
	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, err
	}
	return req, nil
}

func DecodeRequestInto(r *http.Request, req interface{}) error {
	return json.NewDecoder(r.Body).Decode(req)
}

func DecodeAndValidate[T any](r *http.Request, validator func(T) error) (T, error) {
	req, err := DecodeRequest[T](r)
	if err != nil {
		return req, err
	}
	if validator != nil {
		if err := validator(req); err != nil {
			return req, err
		}
	}
	return req, nil
}

func DecodeRequestOrFail[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	req, err := DecodeRequest[T](r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return req, false
	}
	return req, true
}

func DecodeAndValidateOrFail[T any](w http.ResponseWriter, r *http.Request, validator func(T) error) (T, bool) {
	req, err := DecodeAndValidate(r, validator)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err, "")
		return req, false
	}
	return req, true
}

func GetPathParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

func GetPathParamAsUUID(r *http.Request, name string) (string, error) {
	param := chi.URLParam(r, name)
	if param == "" {
		return "", nil
	}
	if _, err := uuid.Parse(param); err != nil {
		return "", err
	}
	return param, nil
}

func GetPathParamAsUUIDOrFail(w http.ResponseWriter, r *http.Request, name string) (string, bool) {
	id, err := GetPathParamAsUUID(r, name)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err, "Invalid "+name+" format")
		return "", false
	}
	return id, true
}
