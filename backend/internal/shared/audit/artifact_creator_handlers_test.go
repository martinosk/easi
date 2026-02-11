package audit

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockArtifactCreatorReader struct {
	creators []ArtifactCreator
	err      error
}

func (m *mockArtifactCreatorReader) GetArtifactCreators(ctx context.Context) ([]ArtifactCreator, error) {
	return m.creators, m.err
}

func TestGetArtifactCreators_ReturnsCreatorsAsJSON(t *testing.T) {
	reader := &mockArtifactCreatorReader{
		creators: []ArtifactCreator{
			{AggregateID: "agg-1", CreatorID: "user-1"},
			{AggregateID: "agg-2", CreatorID: "user-2"},
			{AggregateID: "agg-3", CreatorID: "user-1"},
		},
	}
	handlers := NewArtifactCreatorHandlers(reader)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/artifact-creators", nil)
	w := httptest.NewRecorder()

	handlers.GetArtifactCreators(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response ArtifactCreatorsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 3, len(response.Data))
	assert.Equal(t, "agg-1", response.Data[0].AggregateID)
	assert.Equal(t, "user-1", response.Data[0].CreatorID)
	assert.Equal(t, "agg-2", response.Data[1].AggregateID)
	assert.Equal(t, "user-2", response.Data[1].CreatorID)
	assert.Equal(t, "agg-3", response.Data[2].AggregateID)
	assert.Equal(t, "user-1", response.Data[2].CreatorID)
}

func TestGetArtifactCreators_EmptyList_ReturnsEmptyArray(t *testing.T) {
	reader := &mockArtifactCreatorReader{
		creators: []ArtifactCreator{},
	}
	handlers := NewArtifactCreatorHandlers(reader)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/artifact-creators", nil)
	w := httptest.NewRecorder()

	handlers.GetArtifactCreators(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ArtifactCreatorsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Data)
	assert.Equal(t, 0, len(response.Data))
}

func TestGetArtifactCreators_ReadModelError_Returns500(t *testing.T) {
	reader := &mockArtifactCreatorReader{
		err: errors.New("database connection failed"),
	}
	handlers := NewArtifactCreatorHandlers(reader)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/artifact-creators", nil)
	w := httptest.NewRecorder()

	handlers.GetArtifactCreators(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
