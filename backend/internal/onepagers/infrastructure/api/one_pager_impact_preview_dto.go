package api

import (
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/types"
)

type ImpactPreviewDTO struct {
	SubjectType          string      `json:"subjectType"`
	FieldID              string      `json:"fieldId,omitempty"`
	AffectedSubjectCount int         `json:"affectedSubjectCount"`
	Links                types.Links `json:"_links,omitempty"`
}

func BuildImpactPreviewDTO(preview *queries.ImpactPreview, links *OnePagerLinks, fieldKind string) ImpactPreviewDTO {
	return ImpactPreviewDTO{
		SubjectType:          preview.SubjectType,
		FieldID:              preview.FieldID,
		AffectedSubjectCount: preview.AffectedSubjectCount,
		Links:                sharedAPI.Links{"self": links.Get(impactPreviewSelfPath(preview.SubjectType, preview.FieldID, fieldKind))},
	}
}

func impactPreviewSelfPath(subjectType, fieldID, fieldKind string) string {
	path := configurationPath(subjectType) + "/impact-preview"
	if fieldID == "" {
		return path
	}
	path += "?fieldId=" + fieldID
	if fieldKind == string(valueobjects.FieldRefKindBuiltIn) {
		path += "&fieldKind=" + fieldKind
	}
	return path
}
