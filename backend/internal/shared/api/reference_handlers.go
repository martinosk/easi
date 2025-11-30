package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ReferenceDoc struct {
	Title       string `json:"title"`
	Description string `json:"description"`
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

func SetupReferenceRoutes(r chi.Router) {
	r.Route("/reference", func(r chi.Router) {
		r.Get("/components", HandleGetComponentReference)
		r.Get("/relations/triggering", HandleGetTriggeringRelationReference)
		r.Get("/relations/serving", HandleGetServingRelationReference)
		r.Get("/relations/generic", HandleGetGenericRelationReference)
	})
}
