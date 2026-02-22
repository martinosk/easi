package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"easi/backend/internal/shared/types"
)

const DefaultPageSize = 50
const MaxPageSize = 100

// PaginationParams represents pagination query parameters
type PaginationParams struct {
	Limit  int
	After  string // Opaque cursor for forward pagination
	Before string // Opaque cursor for backward pagination
}

// Cursor represents the internal structure of the pagination cursor
type Cursor struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"ts"`
}

// NameCursor represents a cursor for name-based alphabetical pagination
type NameCursor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	HasMore bool   `json:"hasMore"`
	Limit   int    `json:"limit"`
	Cursor  string `json:"cursor,omitempty"` // Next cursor if hasMore is true
}

// PaginatedResponse wraps data with pagination info and HATEOAS links
type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
	Links      types.Links    `json:"_links"`
}

// ParsePaginationParams extracts pagination parameters from the request
func ParsePaginationParams(r *http.Request) PaginationParams {
	params := PaginationParams{
		Limit:  DefaultPageSize,
		After:  r.URL.Query().Get("after"),
		Before: r.URL.Query().Get("before"),
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			if limit > 0 && limit <= MaxPageSize {
				params.Limit = limit
			}
		}
	}

	return params
}

// DecodeCursor decodes a base64 encoded pagination cursor
func DecodeCursor(encoded string) (*Cursor, error) {
	if encoded == "" {
		return nil, nil
	}

	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var cursor Cursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, err
	}

	return &cursor, nil
}

// EncodeCursor creates an opaque pagination cursor from an ID and timestamp
func EncodeCursor(id string, timestamp time.Time) string {
	cursor := Cursor{
		ID:        id,
		Timestamp: timestamp.Unix(),
	}
	data, _ := json.Marshal(cursor)
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeNameCursor decodes a base64 encoded name-based pagination cursor
func DecodeNameCursor(encoded string) (*NameCursor, error) {
	if encoded == "" {
		return nil, nil
	}

	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var cursor NameCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, err
	}

	return &cursor, nil
}

// EncodeNameCursor creates an opaque pagination cursor from an ID and name
func EncodeNameCursor(id string, name string) string {
	cursor := NameCursor{
		ID:   id,
		Name: name,
	}
	data, _ := json.Marshal(cursor)
	return base64.URLEncoding.EncodeToString(data)
}

type PaginatedResponseParams struct {
	StatusCode int
	Data       interface{}
	HasMore    bool
	NextCursor string
	Limit      int
	SelfLink   string
	BaseLink   string
}

func RespondPaginated(w http.ResponseWriter, params PaginatedResponseParams) {
	links := types.Links{
		"self": types.Link{Href: params.SelfLink, Method: "GET"},
	}

	if params.NextCursor != "" && params.HasMore {
		links["next"] = types.Link{
			Href:   params.BaseLink + "?after=" + params.NextCursor + "&limit=" + strconv.Itoa(params.Limit),
			Method: "GET",
		}
	}

	response := PaginatedResponse{
		Data: params.Data,
		Pagination: PaginationInfo{
			HasMore: params.HasMore,
			Limit:   params.Limit,
			Cursor:  params.NextCursor,
		},
		Links: links,
	}

	RespondJSON(w, params.StatusCode, response)
}
