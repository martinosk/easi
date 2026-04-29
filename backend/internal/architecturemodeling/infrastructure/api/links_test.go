package api

import (
	"testing"

	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLinks() *ArchitectureModelingLinks {
	return NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
}

func writerActor() sharedctx.Actor {
	return sharedctx.NewActor("user-1", "writer@example.com", sharedctx.RoleArchitect)
}

func readerActor() sharedctx.Actor {
	return sharedctx.NewActor("user-2", "reader@example.com", sharedctx.RoleStakeholder)
}

func TestComponentLinksForActor_WritePermission(t *testing.T) {
	links := newTestLinks().ComponentLinksForActor("comp-123", writerActor())

	alwaysPresent := []string{"x-relations-from", "x-relations-to", "x-origins"}
	for _, key := range alwaysPresent {
		require.Contains(t, links, key, "missing always-present link %s", key)
	}

	writeGated := []string{"x-add-relation", "x-set-origin-acquired-via", "x-set-origin-purchased-from", "x-set-origin-built-by"}
	for _, key := range writeGated {
		require.Contains(t, links, key, "missing write-gated link %s", key)
	}
}

func TestComponentLinksForActor_ReadOnly(t *testing.T) {
	links := newTestLinks().ComponentLinksForActor("comp-123", readerActor())

	alwaysPresent := []string{"x-relations-from", "x-relations-to", "x-origins"}
	for _, key := range alwaysPresent {
		require.Contains(t, links, key, "missing always-present link %s", key)
	}

	writeGated := []string{"x-add-relation", "x-set-origin-acquired-via", "x-set-origin-purchased-from", "x-set-origin-built-by"}
	for _, key := range writeGated {
		assert.NotContains(t, links, key, "read-only user should not see write-gated link %s", key)
	}
}

func TestComponentLinksForActor_URLsContainComponentID(t *testing.T) {
	id := "abc-999"
	links := newTestLinks().ComponentLinksForActor(id, writerActor())

	assert.Equal(t, "/api/v1/relations/from/"+id, links["x-relations-from"].Href)
	assert.Equal(t, "GET", links["x-relations-from"].Method)

	assert.Equal(t, "/api/v1/relations/to/"+id, links["x-relations-to"].Href)
	assert.Equal(t, "GET", links["x-relations-to"].Method)

	assert.Equal(t, "/api/v1/components/"+id+"/origins", links["x-origins"].Href)
	assert.Equal(t, "GET", links["x-origins"].Method)

	assert.Equal(t, "/api/v1/relations", links["x-add-relation"].Href)
	assert.Equal(t, "POST", links["x-add-relation"].Method)

	assert.Equal(t, "/api/v1/components/"+id+"/origin/acquired-via", links["x-set-origin-acquired-via"].Href)
	assert.Equal(t, "PUT", links["x-set-origin-acquired-via"].Method)

	assert.Equal(t, "/api/v1/components/"+id+"/origin/purchased-from", links["x-set-origin-purchased-from"].Href)
	assert.Equal(t, "PUT", links["x-set-origin-purchased-from"].Method)

	assert.Equal(t, "/api/v1/components/"+id+"/origin/built-by", links["x-set-origin-built-by"].Href)
	assert.Equal(t, "PUT", links["x-set-origin-built-by"].Method)
}
