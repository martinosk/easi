package readmodels

import (
	"encoding/json"
	"testing"

	"easi/backend/internal/shared/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilityDTO_MarshalJSON_SplicesXRelatedIntoLinks(t *testing.T) {
	dto := CapabilityDTO{
		ID:    "cap1",
		Name:  "Cap",
		Level: "L2",
		Links: types.Links{"self": types.Link{Href: "/api/v1/capabilities/cap1", Method: "GET"}},
		XRelated: []types.RelatedLink{{
			Href:         "/api/v1/capabilities",
			Methods:      []string{"POST"},
			Title:        "Capability (child of)",
			TargetType:   "capability",
			RelationType: "capability-parent",
		}},
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))

	links := out["_links"].(map[string]any)
	xr := links["x-related"].([]any)
	require.Len(t, xr, 1)
	entry := xr[0].(map[string]any)
	assert.Equal(t, "capability-parent", entry["relationType"])
}

func TestCapabilityDTO_MarshalJSON_OmitsXRelatedWhenEmpty(t *testing.T) {
	dto := CapabilityDTO{
		ID:    "cap1",
		Name:  "Cap",
		Level: "L4",
		Links: types.Links{"self": types.Link{Href: "/api/v1/capabilities/cap1", Method: "GET"}},
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))

	links := out["_links"].(map[string]any)
	_, hasXRelated := links["x-related"]
	assert.False(t, hasXRelated)
}
