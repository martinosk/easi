package aggregates

import (
	"fmt"

	"easi/backend/internal/architecturedirection/domain/entities"
	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type plannedTransition struct {
	capabilityID valueobjects.PhysicalCapabilityRef
	kind         valueobjects.JourneyKind
	fromApps     []valueobjects.ApplicationRef
	toApp        valueobjects.ApplicationRef
	note         sharedvo.Description
}

func decodePlannedTransition(evt events.JourneyPlanned) (plannedTransition, error) {
	capabilityID, err := valueobjects.NewPhysicalCapabilityRef(evt.CapabilityID)
	if err != nil {
		return plannedTransition{}, fmt.Errorf("%w: capability ref %q: %v", ErrCorruptedCapabilityJourneyEvent, evt.CapabilityID, err)
	}
	kind, err := valueobjects.NewJourneyKind(evt.Kind)
	if err != nil {
		return plannedTransition{}, fmt.Errorf("%w: kind %q: %v", ErrCorruptedCapabilityJourneyEvent, evt.Kind, err)
	}
	fromApps, err := decodeApplicationRefs(evt.FromComponentIDs)
	if err != nil {
		return plannedTransition{}, err
	}
	toApp, err := valueobjects.NewApplicationRef(evt.ToComponentID)
	if err != nil {
		return plannedTransition{}, fmt.Errorf("%w: to component ref %q: %v", ErrCorruptedCapabilityJourneyEvent, evt.ToComponentID, err)
	}
	note, err := sharedvo.NewDescription(evt.Note)
	if err != nil {
		return plannedTransition{}, fmt.Errorf("%w: note: %v", ErrCorruptedCapabilityJourneyEvent, err)
	}
	return plannedTransition{
		capabilityID: capabilityID,
		kind:         kind,
		fromApps:     fromApps,
		toApp:        toApp,
		note:         note,
	}, nil
}

type plannedDestination struct {
	targetPeriod *valueobjects.TargetPeriod
	targetDomain *valueobjects.BusinessDomainRef
	targetParent *valueobjects.PhysicalCapabilityRef
}

func decodePlannedDestination(evt events.JourneyPlanned) (plannedDestination, error) {
	targetPeriod, err := decodeTargetPeriod(evt.TargetPeriod)
	if err != nil {
		return plannedDestination{}, err
	}
	targetDomain, err := decodeOptionalRef(evt.TargetDomainID, "target domain ref", valueobjects.NewBusinessDomainRef)
	if err != nil {
		return plannedDestination{}, err
	}
	targetParent, err := decodeOptionalRef(evt.TargetParentID, "target parent ref", valueobjects.NewPhysicalCapabilityRef)
	if err != nil {
		return plannedDestination{}, err
	}
	return plannedDestination{
		targetPeriod: targetPeriod,
		targetDomain: targetDomain,
		targetParent: targetParent,
	}, nil
}

func decodeMilestone(snapshot milestoneSnapshot) (entities.Milestone, error) {
	targetPeriod, err := decodeTargetPeriod(snapshot.targetPeriod)
	if err != nil {
		return entities.Milestone{}, err
	}
	milestoneStatus, err := valueobjects.NewMilestoneStatus(snapshot.status)
	if err != nil {
		return entities.Milestone{}, fmt.Errorf("%w: milestone status %q: %v", ErrCorruptedCapabilityJourneyEvent, snapshot.status, err)
	}
	milestone, err := entities.NewMilestone(snapshot.id, snapshot.label, targetPeriod, milestoneStatus)
	if err != nil {
		return entities.Milestone{}, fmt.Errorf("%w: milestone: %v", ErrCorruptedCapabilityJourneyEvent, err)
	}
	return milestone, nil
}

func decodeApplicationRefs(values []string) ([]valueobjects.ApplicationRef, error) {
	out := make([]valueobjects.ApplicationRef, len(values))
	for i, v := range values {
		ref, err := valueobjects.NewApplicationRef(v)
		if err != nil {
			return nil, fmt.Errorf("%w: component ref %q: %v", ErrCorruptedCapabilityJourneyEvent, v, err)
		}
		out[i] = ref
	}
	return out, nil
}

func decodeOptionalRef[T any](value, refName string, construct func(string) (T, error)) (*T, error) {
	if value == "" {
		return nil, nil
	}
	ref, err := construct(value)
	if err != nil {
		return nil, fmt.Errorf("%w: %s %q: %v", ErrCorruptedCapabilityJourneyEvent, refName, value, err)
	}
	return &ref, nil
}

func decodeTargetPeriod(data *events.TargetPeriodData) (*valueobjects.TargetPeriod, error) {
	if data == nil {
		return nil, nil
	}
	tp, err := valueobjects.NewTargetPeriod(data.Year, data.Quarter)
	if err != nil {
		return nil, fmt.Errorf("%w: target period: %v", ErrCorruptedCapabilityJourneyEvent, err)
	}
	return &tp, nil
}

type milestoneSnapshot struct {
	id           string
	label        string
	targetPeriod *events.TargetPeriodData
	status       string
}
