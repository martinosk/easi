package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"easi/backend/internal/onepagers/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

const qualityBasePath = "/one-pager-quality"

type qualityPageSource interface {
	Page(ctx context.Context, query readmodels.SubjectIndexQuery) ([]readmodels.SubjectIndexRecord, bool, error)
}

type OnePagerQualityHandlers struct {
	query qualityPageSource
	links *OnePagerLinks
}

func NewOnePagerQualityHandlers(query qualityPageSource, links *OnePagerLinks) *OnePagerQualityHandlers {
	return &OnePagerQualityHandlers{query: query, links: links}
}

// GetQualityList godoc
// @Summary Get the One-Pager Quality master list
// @Description Returns a cursor-paginated list of every one-pager subject the caller may read, across all five subject types, each row carrying the subject name, type, completeness signal, creator, and created / last-updated dates. Sortable by completeness (default; incomplete subjects first), creator, name, created, or updated. Filtered to the subject types the caller may read (capabilities:read, components:read); a caller holding neither receives 403.
// @Tags one-pagers
// @Produce json
// @Param sort query string false "Sort dimension" Enums(completeness, creator, name, created, updated) default(completeness)
// @Param order query string false "Sort order" Enums(asc, desc) default(asc)
// @Param limit query int false "Number of items per page (max 100)" default(50)
// @Param after query string false "Cursor for pagination (opaque token)"
// @Success 200 {object} easi_backend_internal_shared_api.PaginatedResponse{data=[]api.QualityRowDTO}
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pager-quality [get]
func (h *OnePagerQualityHandlers) GetQualityList(w http.ResponseWriter, r *http.Request) {
	actor, _ := sharedctx.GetActor(r.Context())
	subjectTypes := readableSubjectTypes(actor)
	if len(subjectTypes) == 0 {
		sharedAPI.RespondError(w, http.StatusForbidden, nil, "You do not have permission to read any one-pager subjects")
		return
	}

	params, ok := h.parseQualityParams(w, r)
	if !ok {
		return
	}

	records, hasMore, err := h.query.Page(r.Context(), readmodels.SubjectIndexQuery{
		SubjectTypes: subjectTypes,
		Sort:         params.sort,
		Order:        params.order,
		Limit:        params.limit,
		After:        params.after,
	})
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve the one-pager quality list")
		return
	}

	h.respond(w, qualityResponseInput{params: params, records: records, hasMore: hasMore, actor: actor})
}

type qualityParams struct {
	sort  string
	order string
	limit int
	after *readmodels.SubjectIndexRecord
	raw   sharedAPI.PaginationParams
}

func (h *OnePagerQualityHandlers) parseQualityParams(w http.ResponseWriter, r *http.Request) (qualityParams, bool) {
	sort, ok := parseQualitySort(r.URL.Query().Get("sort"))
	if !ok {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Invalid sort dimension")
		return qualityParams{}, false
	}
	order, ok := parseQualityOrder(r.URL.Query().Get("order"))
	if !ok {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Invalid sort order")
		return qualityParams{}, false
	}
	pagination := sharedAPI.ParsePaginationParams(r)
	after, err := decodeQualityCursor(pagination.After)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pagination cursor")
		return qualityParams{}, false
	}
	return qualityParams{sort: sort, order: order, limit: pagination.Limit, after: after, raw: pagination}, true
}

type qualityResponseInput struct {
	params  qualityParams
	records []readmodels.SubjectIndexRecord
	hasMore bool
	actor   sharedctx.Actor
}

func (h *OnePagerQualityHandlers) respond(w http.ResponseWriter, input qualityResponseInput) {
	rows := make([]QualityRowDTO, len(input.records))
	for i, record := range input.records {
		rows[i] = toQualityRow(record, input.actor, h.links)
	}

	nextCursor := ""
	if input.hasMore && len(input.records) > 0 {
		nextCursor = encodeQualityCursor(input.records[len(input.records)-1])
	}

	sharedAPI.RespondJSON(w, http.StatusOK, sharedAPI.PaginatedResponse{
		Data: rows,
		Pagination: sharedAPI.PaginationInfo{
			HasMore: input.hasMore,
			Limit:   input.params.limit,
			Cursor:  nextCursor,
		},
		Links: h.buildLinks(input.params, nextCursor),
	})
}

func (h *OnePagerQualityHandlers) buildLinks(params qualityParams, nextCursor string) types.Links {
	links := types.Links{
		"self": h.links.Get(qualityBasePath + "?" + params.query("")),
	}
	if nextCursor != "" {
		links["next"] = h.links.Get(qualityBasePath + "?" + params.query(nextCursor))
	}
	return links
}

func (p qualityParams) query(after string) string {
	values := url.Values{}
	values.Set("sort", p.sort)
	values.Set("order", p.order)
	values.Set("limit", strconv.Itoa(p.limit))
	if after == "" {
		after = p.raw.After
	}
	if after != "" {
		values.Set("after", after)
	}
	return values.Encode()
}
