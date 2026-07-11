package api

import (
	"easi/backend/internal/onepagers/application/queries"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/types"
)

type ImpactPreviewDTO struct {
	SubjectType          string      `json:"subjectType"`
	FieldID              string      `json:"fieldId,omitempty"`
	AffectedSubjectCount int         `json:"affectedSubjectCount"`
	Links                types.Links `json:"_links,omitempty"`
}

func BuildImpactPreviewDTO(preview *queries.ImpactPreview, links *OnePagerLinks) ImpactPreviewDTO {
	return ImpactPreviewDTO{
		SubjectType:          preview.SubjectType,
		FieldID:              preview.FieldID,
		AffectedSubjectCount: preview.AffectedSubjectCount,
		Links:                sharedAPI.Links{"self": links.Get(impactPreviewSelfPath(preview.SubjectType, preview.FieldID))},
	}
}

func impactPreviewSelfPath(subjectType, fieldID string) string {
	path := configurationPath(subjectType) + "/impact-preview"
	if fieldID != "" {
		path += "?fieldId=" + fieldID
	}
	return path
}
