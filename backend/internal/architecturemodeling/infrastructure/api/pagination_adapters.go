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

type NamePageableAcquiredEntity struct {
	Entity readmodels.AcquiredEntityDTO
}

func (p NamePageableAcquiredEntity) GetID() string {
	return p.Entity.ID
}

func (p NamePageableAcquiredEntity) GetName() string {
	return p.Entity.Name
}

func ConvertAcquiredEntitiesToNamePageable(entities []readmodels.AcquiredEntityDTO) []sharedAPI.NamePageable {
	pageables := make([]sharedAPI.NamePageable, len(entities))
	for i, e := range entities {
		pageables[i] = NamePageableAcquiredEntity{Entity: e}
	}
	return pageables
}

type NamePageableVendor struct {
	Vendor readmodels.VendorDTO
}

func (p NamePageableVendor) GetID() string {
	return p.Vendor.ID
}

func (p NamePageableVendor) GetName() string {
	return p.Vendor.Name
}

func ConvertVendorsToNamePageable(vendors []readmodels.VendorDTO) []sharedAPI.NamePageable {
	pageables := make([]sharedAPI.NamePageable, len(vendors))
	for i, v := range vendors {
		pageables[i] = NamePageableVendor{Vendor: v}
	}
	return pageables
}

type NamePageableInternalTeam struct {
	Team readmodels.InternalTeamDTO
}

func (p NamePageableInternalTeam) GetID() string {
	return p.Team.ID
}

func (p NamePageableInternalTeam) GetName() string {
	return p.Team.Name
}

func ConvertInternalTeamsToNamePageable(teams []readmodels.InternalTeamDTO) []sharedAPI.NamePageable {
	pageables := make([]sharedAPI.NamePageable, len(teams))
	for i, t := range teams {
		pageables[i] = NamePageableInternalTeam{Team: t}
	}
	return pageables
}
