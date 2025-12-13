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

func (h *PaginationHelper) ProcessCursor(after string) (string, int64, error) {
	if after == "" {
		return "", 0, nil
	}

	cursor, err := DecodeCursor(after)
	if err != nil {
		return "", 0, err
	}

	if cursor == nil {
		return "", 0, nil
	}

	return cursor.ID, cursor.Timestamp, nil
}

type Pageable interface {
	GetID() string
	GetTimestamp() time.Time
}

func (h *PaginationHelper) GenerateNextCursor(items []Pageable, hasMore bool) string {
	if !hasMore || len(items) == 0 {
		return ""
	}

	lastItem := items[len(items)-1]
	return EncodeCursor(lastItem.GetID(), lastItem.GetTimestamp())
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
