package api

import (
	"fmt"
	"time"
)

type PaginationHelper struct {
	basePath string
}

func NewPaginationHelper(basePath string) *PaginationHelper {
	return &PaginationHelper{
		basePath: basePath,
	}
}

func processCursor[C any, V any](after string, decode func(string) (*C, error), project func(*C) (string, V)) (string, V, error) {
	var zeroVal V
	if after == "" {
		return "", zeroVal, nil
	}

	cursor, err := decode(after)
	if err != nil {
		return "", zeroVal, err
	}

	if cursor == nil {
		return "", zeroVal, nil
	}

	id, val := project(cursor)
	return id, val, nil
}

func (h *PaginationHelper) ProcessCursor(after string) (string, int64, error) {
	return processCursor(after, DecodeCursor, func(c *Cursor) (string, int64) {
		return c.ID, c.Timestamp
	})
}

type Pageable interface {
	GetID() string
	GetTimestamp() time.Time
}

type NamePageable interface {
	GetID() string
	GetName() string
}

func (h *PaginationHelper) GenerateNextCursor(items []Pageable, hasMore bool) string {
	if !hasMore || len(items) == 0 {
		return ""
	}

	lastItem := items[len(items)-1]
	return EncodeCursor(lastItem.GetID(), lastItem.GetTimestamp())
}

func (h *PaginationHelper) GenerateNextNameCursor(items []NamePageable, hasMore bool) string {
	if !hasMore || len(items) == 0 {
		return ""
	}

	lastItem := items[len(items)-1]
	return EncodeNameCursor(lastItem.GetID(), lastItem.GetName())
}

func (h *PaginationHelper) ProcessNameCursor(after string) (string, string, error) {
	return processCursor(after, DecodeNameCursor, func(c *NameCursor) (string, string) {
		return c.ID, c.Name
	})
}

func (h *PaginationHelper) BuildSelfLink(params PaginationParams) string {
	if params.After == "" {
		return h.basePath
	}
	return fmt.Sprintf("%s?after=%s&limit=%d", h.basePath, params.After, params.Limit)
}

func (h *PaginationHelper) BuildLinks(params PaginationParams, hasMore bool, nextCursor string) map[string]string {
	links := map[string]string{
		"self": h.BuildSelfLink(params),
	}

	if hasMore && nextCursor != "" {
		links["next"] = fmt.Sprintf("%s?after=%s&limit=%d", h.basePath, nextCursor, params.Limit)
	}

	return links
}
