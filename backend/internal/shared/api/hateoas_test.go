package api

import (
	"testing"

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
