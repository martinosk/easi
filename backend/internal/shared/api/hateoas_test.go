package api

import (
	"testing"

	sharedctx "easi/backend/internal/shared/context"

	"github.com/stretchr/testify/assert"
)

func TestNewHATEOASLinks_DefaultsToAPIVersionPrefix(t *testing.T) {
	h := NewHATEOASLinks("")
	assert.Equal(t, APIVersionPrefix, h.Base())
}

func TestNewHATEOASLinks_UsesProvidedBaseURL(t *testing.T) {
	h := NewHATEOASLinks("/api/v1")
	assert.Equal(t, "/api/v1", h.Base())
}

func TestGet_ReturnsCorrectLink(t *testing.T) {
	h := NewHATEOASLinks("/api/v1")
	link := h.Get("/components/123")
	assert.Equal(t, "/api/v1/components/123", link.Href)
	assert.Equal(t, "GET", link.Method)
}

func TestAddStewardshipsLink_ForAnyReaderOfDomains(t *testing.T) {
	h := NewHATEOASLinks("/api/v1")
	for _, role := range []sharedctx.Role{sharedctx.RoleAdmin, sharedctx.RoleArchitect, sharedctx.RoleStakeholder} {
		links := Links{}
		h.AddStewardshipsLink(links, sharedctx.NewActor("u1", "u@example.com", role), "domain 1")
		link, ok := links["x-stewardships"]
		assert.True(t, ok, role)
		assert.Equal(t, "/api/v1/stewardships?domainId=domain+1", link.Href)
		assert.Equal(t, "GET", link.Method)
	}
}

func TestAddStewardshipsLink_AbsentWithoutDomainsRead(t *testing.T) {
	h := NewHATEOASLinks("/api/v1")
	links := Links{}
	h.AddStewardshipsLink(links, sharedctx.Actor{ID: "u1", Permissions: map[string]bool{"components:read": true}}, "domain-1")
	_, ok := links["x-stewardships"]
	assert.False(t, ok)
}
