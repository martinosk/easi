package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	authPL "easi/backend/internal/auth/publishedlanguage"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
	"easi/backend/internal/stewardship/application/home"
)

const homePath sharedAPI.ResourcePath = "/home"

var errNoActor = errors.New("no authenticated actor")

type HomeComposer interface {
	Compose(ctx context.Context, caller home.Caller) (*home.Home, error)
}

type HomeHandlers struct {
	composer HomeComposer
	hateoas  *sharedAPI.HATEOASLinks
}

func NewHomeHandlers(composer HomeComposer, hateoas *sharedAPI.HATEOASLinks) *HomeHandlers {
	return &HomeHandlers{composer: composer, hateoas: hateoas}
}

type HomeResponse struct {
	*home.Home
	Links types.Links `json:"_links"`
}

// GetHome godoc
// @Summary Get the caller's home
// @Description Composes the home of the signed-in user: their scope (anchors: stewardships, architected domains, EA-owned capabilities, owned or nominated applications, active edit grants), the portfolio tiles over that scope and My Work. Scope kind is personal with any anchor, otherwise tenant for callers holding domains:write and empty for everyone else. Sections the caller may not read are omitted; the response is never 403. Takes no parameters: every anchor is resolved from the session.
// @Tags home
// @Produce json
// @Security CookieAuth
// @Success 200 {object} HomeResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /home [get]
func (h *HomeHandlers) GetHome(w http.ResponseWriter, r *http.Request) {
	actor, ok := sharedctx.GetActor(r.Context())
	if !ok {
		sharedAPI.RespondError(w, http.StatusUnauthorized, errNoActor, "Authentication required")
		return
	}
	composed, err := h.composer.Compose(r.Context(), home.Caller{
		UserID:            actor.ID,
		Email:             actor.Email,
		MayAssignStewards: actor.HasPermission(authPL.PermDomainsWrite.String()),
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "compose home", "error", err, "actorID", actor.ID)
		sharedAPI.HandleError(w, err)
		return
	}
	composed.RedactFor(readableBy(actor))
	h.linkWorkItems(composed)
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Vary", "Cookie")
	sharedAPI.RespondJSON(w, http.StatusOK, HomeResponse{Home: composed, Links: types.Links{"self": h.hateoas.Get(string(homePath))}})
}

func readableBy(actor sharedctx.Actor) home.Readable {
	return home.Readable{
		Capabilities: actor.HasPermission(authPL.PermCapabilitiesRead.String()),
		Applications: actor.HasPermission(authPL.PermComponentsRead.String()),
		Direction:    actor.HasPermission(authPL.PermArchitectureDirectionRead.String()),
		Domains:      actor.HasPermission(authPL.PermDomainsRead.String()),
	}
}

func (h *HomeHandlers) linkWorkItems(composed *home.Home) {
	if composed.MyWork == nil {
		return
	}
	for i := range composed.MyWork.Items {
		item := &composed.MyWork.Items[i]
		item.Links = types.Links{"x-one-pager": h.hateoas.Get("/one-pagers/" + string(item.SubjectType) + "/" + url.PathEscape(item.ID))}
	}
}
