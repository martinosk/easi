package readmodels

import (
	"encoding/json"
	"testing"

	"easi/backend/internal/shared/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationComponentDTO_MarshalJSON_SplicesXRelatedIntoLinks(t *testing.T) {
	dto := ApplicationComponentDTO{
		ID:    "c1",
		Name:  "Comp",
		Links: types.Links{"self": types.Link{Href: "/api/v1/components/c1", Method: "GET"}},
		XRelated: []types.RelatedLink{{
			Href:         "/api/v1/components",
			Methods:      []string{"POST"},
			Title:        "Component (related)",
			TargetType:   "component",
			RelationType: "component-relation",
		}},
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))

	links, ok := out["_links"].(map[string]any)
	require.True(t, ok, "_links should be an object")

	related, ok := links["x-related"].([]any)
	require.True(t, ok, "x-related should be an array, got %T", links["x-related"])
	require.Len(t, related, 1)

	entry := related[0].(map[string]any)
	assert.Equal(t, "/api/v1/components", entry["href"])
	assert.Equal(t, []any{"POST"}, entry["methods"])
	assert.Equal(t, "Component (related)", entry["title"])
	assert.Equal(t, "component", entry["targetType"])
	assert.Equal(t, "component-relation", entry["relationType"])
}

func TestApplicationComponentDTO_MarshalJSON_OmitsXRelatedWhenEmpty(t *testing.T) {
	dto := ApplicationComponentDTO{
		ID:    "c1",
		Name:  "Comp",
		Links: types.Links{"self": types.Link{Href: "/api/v1/components/c1", Method: "GET"}},
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))

	links, ok := out["_links"].(map[string]any)
	require.True(t, ok)
	_, hasXRelated := links["x-related"]
	assert.False(t, hasXRelated, "x-related should be omitted when XRelated is empty")
}

func TestApplicationComponentDTO_MarshalJSON_PreservesScalarFields(t *testing.T) {
	dto := ApplicationComponentDTO{
		ID:          "c1",
		Name:        "Comp",
		Description: "desc",
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))

	assert.Equal(t, "c1", out["id"])
	assert.Equal(t, "Comp", out["name"])
	assert.Equal(t, "desc", out["description"])
}

func sampleRelated() []types.RelatedLink {
	return []types.RelatedLink{{
		Href: "/api/v1/components", Methods: []string{"POST"},
		Title: "Component (origin)", TargetType: "component", RelationType: "origin-acquired-via",
	}}
}

func TestOriginDTO_MarshalJSON_SplicesXRelated(t *testing.T) {
	cases := []struct {
		name string
		dto  any
	}{
		{
			name: "AcquiredEntity",
			dto: AcquiredEntityDTO{
				ID: "ae1", Name: "Acme",
				Links:    types.Links{"self": types.Link{Href: "/api/v1/acquired-entities/ae1", Method: "GET"}},
				XRelated: sampleRelated(),
			},
		},
		{
			name: "Vendor",
			dto: VendorDTO{
				ID: "v1", Name: "Vendor",
				Links:    types.Links{"self": types.Link{Href: "/api/v1/vendors/v1", Method: "GET"}},
				XRelated: sampleRelated(),
			},
		},
		{
			name: "InternalTeam",
			dto: InternalTeamDTO{
				ID: "t1", Name: "Team",
				Links:    types.Links{"self": types.Link{Href: "/api/v1/internal-teams/t1", Method: "GET"}},
				XRelated: sampleRelated(),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.dto)
			require.NoError(t, err)
			var out map[string]any
			require.NoError(t, json.Unmarshal(data, &out))
			xr := out["_links"].(map[string]any)["x-related"].([]any)
			require.Len(t, xr, 1)
		})
	}
}
