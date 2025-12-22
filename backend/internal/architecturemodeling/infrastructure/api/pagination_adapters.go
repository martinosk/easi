package api

import (
	"easi/backend/internal/architecturemodeling/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
)

type NamePageableComponent struct {
	Component readmodels.ApplicationComponentDTO
}

func (p NamePageableComponent) GetID() string {
	return p.Component.ID
}

func (p NamePageableComponent) GetName() string {
	return p.Component.Name
}

func ConvertToNamePageable(components []readmodels.ApplicationComponentDTO) []sharedAPI.NamePageable {
	pageables := make([]sharedAPI.NamePageable, len(components))
	for i, c := range components {
		pageables[i] = NamePageableComponent{Component: c}
	}
	return pageables
}
