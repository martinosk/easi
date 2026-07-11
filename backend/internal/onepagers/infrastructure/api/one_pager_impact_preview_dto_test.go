package api

import (
	"testing"

	"easi/backend/internal/onepagers/application/queries"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildImpactPreviewDTO_MapsFields(t *testing.T) {
	preview := &queries.ImpactPreview{SubjectType: "application", FieldID: "contract-link", AffectedSubjectCount: 37}

	dto := BuildImpactPreviewDTO(preview, testLinks())

	assert.Equal(t, "application", dto.SubjectType)
	assert.Equal(t, "contract-link", dto.FieldID)
	assert.Equal(t, 37, dto.AffectedSubjectCount)
}

func TestBuildImpactPreviewDTO_SelfLinkIncludesFieldID(t *testing.T) {
	preview := &queries.ImpactPreview{SubjectType: "application", FieldID: "contract-link", AffectedSubjectCount: 37}

	dto := BuildImpactPreviewDTO(preview, testLinks())

	self, ok := dto.Links["self"]
	require.True(t, ok)
	assert.Equal(t, "/api/v1/one-pagers/configurations/application/impact-preview?fieldId=contract-link", self.Href)
	assert.Equal(t, "GET", self.Method)
}

func TestBuildImpactPreviewDTO_SelfLinkOmitsFieldIDWhenNewField(t *testing.T) {
	preview := &queries.ImpactPreview{SubjectType: "vendor", FieldID: "", AffectedSubjectCount: 120}

	dto := BuildImpactPreviewDTO(preview, testLinks())

	self, ok := dto.Links["self"]
	require.True(t, ok)
	assert.Equal(t, "/api/v1/one-pagers/configurations/vendor/impact-preview", self.Href)
}
