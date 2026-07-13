package services

import "fmt"

type ActiveJourneyError struct {
	ExistingJourneyID string
}

func (e *ActiveJourneyError) Error() string {
	return fmt.Sprintf("an active journey already exists on this capability (id: %s)", e.ExistingJourneyID)
}
