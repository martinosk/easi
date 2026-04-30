package api

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleGetXRelatedReference_DescribesContract(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/reference/x-related-links", nil)

	HandleGetXRelatedReference(w, r)

	assert.Equal(t, 200, w.Code)
	var body XRelatedReferenceDoc
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.NotEmpty(t, body.Title)
	assert.Contains(t, body.Description, "x-related")
	require.NotEmpty(t, body.Example, "expected at least one example RelatedLink")

	example := body.Example[0]
	assert.NotEmpty(t, example.Href, "example.href must be non-empty")
	assert.Contains(t, example.Methods, "POST", "example.methods must include POST so consumers see a picker-eligible entry")
	assert.NotEmpty(t, example.Title, "example.title (picker label) must be non-empty")
	assert.NotEmpty(t, example.TargetType, "example.targetType (which create dialog to open) must be non-empty")
	assert.NotEmpty(t, example.RelationType, "example.relationType (relation endpoint identifier) must be non-empty")
}
