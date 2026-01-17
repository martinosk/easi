package commands

import "time"

type CreateAcquiredEntity struct {
	Name              string
	AcquisitionDate   *time.Time
	IntegrationStatus string
	Notes             string
}

func (c CreateAcquiredEntity) CommandName() string {
	return "CreateAcquiredEntity"
}
