package api

import (
	"testing"

	sharedctx "easi/backend/internal/shared/context"

	"github.com/stretchr/testify/assert"
)

func actorWithPermissions(perms ...string) sharedctx.Actor {
	m := make(map[string]bool, len(perms))
	for _, p := range perms {
		m[p] = true
	}
	return sharedctx.Actor{ID: "test-user", Email: "test@example.com", Permissions: m}
}

func TestStrategyImportanceCollectionLinksForActor_WithWritePermission(t *testing.T) {
	h := NewHATEOASLinks("/api/v1")
	actor := actorWithPermissions("domains:write")

	links := h.StrategyImportanceCollectionLinksForActor("dom-1", "cap-1", actor)

	assert.Contains(t, links, "self")
	assert.Equal(t, "/api/v1/business-domains/dom-1/capabilities/cap-1/importance", links["self"].Href)
	assert.Equal(t, "GET", links["self"].Method)

	assert.Contains(t, links, "create")
	assert.Equal(t, "/api/v1/business-domains/dom-1/capabilities/cap-1/importance", links["create"].Href)
	assert.Equal(t, "POST", links["create"].Method)
}

func TestStrategyImportanceCollectionLinksForActor_WithoutWritePermission(t *testing.T) {
	h := NewHATEOASLinks("/api/v1")
	actor := actorWithPermissions()

	links := h.StrategyImportanceCollectionLinksForActor("dom-1", "cap-1", actor)

	assert.Contains(t, links, "self")
	assert.NotContains(t, links, "create")
}

func TestFitScoresCollectionLinksForActor_WithWritePermission(t *testing.T) {
	h := NewHATEOASLinks("/api/v1")
	actor := actorWithPermissions("components:write")

	links := h.FitScoresCollectionLinksForActor("comp-1", actor)

	assert.Contains(t, links, "self")
	assert.Contains(t, links, "create")
}

func TestFitScoresCollectionLinksForActor_WithoutWritePermission(t *testing.T) {
	h := NewHATEOASLinks("/api/v1")
	actor := actorWithPermissions()

	links := h.FitScoresCollectionLinksForActor("comp-1", actor)

	assert.Contains(t, links, "self")
	assert.NotContains(t, links, "create")
}
