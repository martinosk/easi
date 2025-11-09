package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
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

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	HasMore bool   `json:"hasMore"`
	Limit   int    `json:"limit"`
	Cursor  string `json:"cursor,omitempty"` // Next cursor if hasMore is true
}

// PaginatedResponse wraps data with pagination info and HATEOAS links
type PaginatedResponse struct {
	Data       interface{}       `json:"data"`
	Pagination PaginationInfo    `json:"pagination"`
	Links      map[string]string `json:"_links"`
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

// RespondPaginated sends a paginated response with proper HATEOAS links
func RespondPaginated(w http.ResponseWriter, statusCode int, data interface{}, hasMore bool, nextCursor string, limit int, selfLink string, baseLink string) {
	links := map[string]string{
		"self": selfLink,
	}

	if nextCursor != "" && hasMore {
		links["next"] = baseLink + "?after=" + nextCursor + "&limit=" + strconv.Itoa(limit)
	}

	response := PaginatedResponse{
		Data: data,
		Pagination: PaginationInfo{
			HasMore: hasMore,
			Limit:   limit,
			Cursor:  nextCursor,
		},
		Links: links,
	}

	RespondJSON(w, statusCode, response)
}
