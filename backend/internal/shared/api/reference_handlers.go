package api

import (
	"net/http"

	"easi/backend/internal/shared/types"

	"github.com/go-chi/chi/v5"
)

type ReferenceDoc struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type XRelatedReferenceDoc struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Example     []types.RelatedLink `json:"example"`
}

var xRelatedReferenceDoc = XRelatedReferenceDoc{
	Title: "x-related Links Contract",
	Description: "Every canvas-renderable entity exposes an `x-related` array " +
		"under `_links`, enumerating creation affordances toward related " +
		"entities. Each entry carries href, methods, title, targetType, and " +
		"relationType. Entries advertising POST in their methods are " +
		"picker-eligible (clicking a canvas handle on the source entity " +
		"opens the create dialog for targetType, then dispatches to the " +
		"existing relation endpoint identified by relationType). Entries " +
		"that advertise only GET are reserved for future list-existing-" +
		"related UIs and are ignored by the picker.",
	Example: []types.RelatedLink{
		{
			Href:         "/api/v1/components",
			Methods:      []string{"POST"},
			Title:        "Component (triggers)",
			TargetType:   "component",
			RelationType: "component-triggers",
		},
	},
}

var referenceDocuments = map[string]ReferenceDoc{
	"components": {
		Title:       "Application Component",
		Description: "An Application Component represents a modular, deployable, and replaceable part of a software system that encapsulates its contents and exposes its functionality through a set of interfaces.",
	},
	"relations/triggering": {
		Title:       "Triggering Relationship",
		Description: "A Triggering relationship represents a temporal or causal dependency between two elements. The source element initiates or triggers the behavior of the target element.",
	},
	"relations/serving": {
		Title:       "Serving Relationship",
		Description: "A Serving relationship represents that an element provides its functionality to another element. The source element serves or provides services to the target element.",
	},
	"relations/generic": {
		Title:       "Relationship",
		Description: "A relationship represents a connection or dependency between two architectural elements. Relationships define how elements interact with or depend on each other.",
	},
}

// @Summary Get component reference documentation
// @Description Returns reference documentation for Application Components
// @Tags Reference
// @Produce json
// @Success 200 {object} ReferenceDoc
// @Router /reference/components [get]
func HandleGetComponentReference(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, referenceDocuments["components"])
}

// @Summary Get triggering relationship reference documentation
// @Description Returns reference documentation for Triggering relationships
// @Tags Reference
// @Produce json
// @Success 200 {object} ReferenceDoc
// @Router /reference/relations/triggering [get]
func HandleGetTriggeringRelationReference(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, referenceDocuments["relations/triggering"])
}

// @Summary Get serving relationship reference documentation
// @Description Returns reference documentation for Serving relationships
// @Tags Reference
// @Produce json
// @Success 200 {object} ReferenceDoc
// @Router /reference/relations/serving [get]
func HandleGetServingRelationReference(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, referenceDocuments["relations/serving"])
}

// @Summary Get generic relationship reference documentation
// @Description Returns reference documentation for generic relationships
// @Tags Reference
// @Produce json
// @Success 200 {object} ReferenceDoc
// @Router /reference/relations/generic [get]
func HandleGetGenericRelationReference(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, referenceDocuments["relations/generic"])
}

// HandleGetXRelatedReference godoc
// @Summary Get x-related HATEOAS contract reference
// @Description Returns the contract for the `_links["x-related"]` array exposed on every canvas-renderable entity (Component, Capability, AcquiredEntity, Vendor, InternalTeam). Each entry advertises a creation affordance from the source entity to a related entity; entries with `POST` in `methods` are picker-eligible.
// @Tags Reference
// @Produce json
// @Success 200 {object} XRelatedReferenceDoc
// @Failure 401 {object} ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} ErrorResponse "Forbidden - insufficient permissions"
// @Router /reference/x-related-links [get]
func HandleGetXRelatedReference(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, xRelatedReferenceDoc)
}

func SetupReferenceRoutes(r chi.Router) {
	r.Route("/reference", func(r chi.Router) {
		r.Get("/components", HandleGetComponentReference)
		r.Get("/relations/triggering", HandleGetTriggeringRelationReference)
		r.Get("/relations/serving", HandleGetServingRelationReference)
		r.Get("/relations/generic", HandleGetGenericRelationReference)
		r.Get("/x-related-links", HandleGetXRelatedReference)
	})
}
