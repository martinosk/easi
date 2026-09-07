package aggregates

import (
	"fmt"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

func (a *ApplicationComponent) ClassifyHosting(classification valueobjects.HostingClassification) error {
	event := events.NewApplicationHostingClassified(a.ID(), classification.String())

	if err := a.apply(event); err != nil {
		return err
	}
	a.RaiseEvent(event)

	return nil
}

func (a *ApplicationComponent) Hosting() valueobjects.HostingClassification {
	return a.hosting
}

func (a *ApplicationComponent) applyHostingClassified(e events.ApplicationHostingClassified) error {
	classification, err := valueobjects.NewHostingClassification(e.Hosting)
	if err != nil {
		return fmt.Errorf("%w: hosting classification: %v", domain.ErrCorruptedEvent, err)
	}
	a.hosting = classification
	return nil
}
