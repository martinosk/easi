package commands

import "time"

type UpdateAcquiredEntity struct {
	ID                string
	Name              string
	AcquisitionDate   *time.Time
	IntegrationStatus string
	Notes             string
}

func (c UpdateAcquiredEntity) CommandName() string {
	return "UpdateAcquiredEntity"
}
