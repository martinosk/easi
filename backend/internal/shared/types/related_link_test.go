package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelatedLink_MarshalsToSpecJSONShape(t *testing.T) {
	rl := RelatedLink{
		Href:         "/api/v1/components",
		Methods:      []string{"POST"},
		Title:        "Component (related)",
		TargetType:   "component",
		RelationType: "component-relation",
	}

	data, err := json.Marshal(rl)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, "/api/v1/components", got["href"])
	assert.Equal(t, []any{"POST"}, got["methods"])
	assert.Equal(t, "Component (related)", got["title"])
	assert.Equal(t, "component", got["targetType"])
	assert.Equal(t, "component-relation", got["relationType"])
}

func TestRelatedLink_MultipleMethodsArePreserved(t *testing.T) {
	rl := RelatedLink{
		Href:         "/api/v1/capabilities/abc/requires",
		Methods:      []string{"GET", "POST"},
		Title:        "Required Capabilities",
		TargetType:   "capability",
		RelationType: "capability-requires",
	}

	data, err := json.Marshal(rl)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, []any{"GET", "POST"}, got["methods"])
}

func TestSpliceXRelated_AddsXRelatedToExistingLinks(t *testing.T) {
	in := []byte(`{"id":"c1","_links":{"self":{"href":"/api/v1/components/c1","method":"GET"}}}`)
	related := []RelatedLink{{
		Href: "/api/v1/components", Methods: []string{"POST"},
		Title: "Component (related)", TargetType: "component", RelationType: "component-relation",
	}}

	out, err := SpliceXRelated(in, related)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(out, &got))
	links := got["_links"].(map[string]any)
	assert.NotNil(t, links["self"], "existing self link should be preserved")
	xr := links["x-related"].([]any)
	require.Len(t, xr, 1)
	entry := xr[0].(map[string]any)
	assert.Equal(t, "component-relation", entry["relationType"])
}

func TestSpliceXRelated_CreatesLinksWhenAbsent(t *testing.T) {
	in := []byte(`{"id":"c1"}`)
	related := []RelatedLink{{
		Href: "/api/v1/components", Methods: []string{"POST"},
		Title: "Component (related)", TargetType: "component", RelationType: "component-relation",
	}}

	out, err := SpliceXRelated(in, related)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(out, &got))
	links, ok := got["_links"].(map[string]any)
	require.True(t, ok, "_links should be created")
	xr := links["x-related"].([]any)
	require.Len(t, xr, 1)
}

func TestSpliceXRelated_ReturnsInputUnchangedWhenRelatedEmpty(t *testing.T) {
	in := []byte(`{"id":"c1","_links":{"self":{"href":"/api/v1/components/c1","method":"GET"}}}`)
	out, err := SpliceXRelated(in, nil)
	require.NoError(t, err)
	assert.Equal(t, string(in), string(out))
}
