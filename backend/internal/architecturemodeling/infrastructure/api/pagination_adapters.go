package api

import (
	"time"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
)

type PageableComponent struct {
	Component readmodels.ApplicationComponentDTO
}

func (p PageableComponent) GetID() string {
	return p.Component.ID
}

func (p PageableComponent) GetTimestamp() time.Time {
	return p.Component.CreatedAt
}

func ConvertToPageable(components []readmodels.ApplicationComponentDTO) []sharedAPI.Pageable {
	pageables := make([]sharedAPI.Pageable, len(components))
	for i, c := range components {
		pageables[i] = PageableComponent{Component: c}
	}
	return pageables
}
