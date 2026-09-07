package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
)

var subjectTypeByDeletionEvent = map[string]string{
	capPL.CapabilityDeleted:          "capability",
	amPL.ApplicationComponentDeleted: "application",
	amPL.AcquiredEntityDeleted:       "acquired-entity",
	amPL.VendorDeleted:               "vendor",
	amPL.InternalTeamDeleted:         "internal-team",
}

func SubjectDeletionEventTypes() []string {
	eventTypes := make([]string, 0, len(subjectTypeByDeletionEvent))
	for eventType := range subjectTypeByDeletionEvent {
		eventTypes = append(eventTypes, eventType)
	}
	return eventTypes
}

type FactsFinder interface {
	FactsIDForSubject(ctx context.Context, subject readmodels.SubjectKey) (string, error)
}

type CommandDispatcher interface {
	Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error)
}

type SubjectDeletedReactor struct {
	facts    FactsFinder
	commands CommandDispatcher
}

func NewSubjectDeletedReactor(facts FactsFinder, commandDispatcher CommandDispatcher) *SubjectDeletedReactor {
	return &SubjectDeletedReactor{facts: facts, commands: commandDispatcher}
}

func (r *SubjectDeletedReactor) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
	}
	return r.ProjectEvent(ctx, event.EventType(), eventData)
}

type subjectDeletedEvent struct {
	ID string `json:"id"`
}

func (r *SubjectDeletedReactor) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	subjectType, found := subjectTypeByDeletionEvent[eventType]
	if !found {
		return nil
	}

	var event subjectDeletedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal %s event: %w", eventType, err)
	}

	return r.archiveFacts(ctx, subjectType, event.ID)
}

func (r *SubjectDeletedReactor) archiveFacts(ctx context.Context, subjectType, subjectID string) error {
	factsID, err := r.facts.FactsIDForSubject(ctx, readmodels.SubjectKey{SubjectType: subjectType, SubjectID: subjectID})
	if err != nil {
		return fmt.Errorf("look up one-pager facts for deleted %s %s: %w", subjectType, subjectID, err)
	}
	if factsID == "" {
		return nil
	}

	if _, err := r.commands.Dispatch(ctx, &commands.ArchiveOnePagerFacts{
		FactsID: factsID,
		Reason:  aggregates.ArchiveReasonSubjectDeleted,
	}); err != nil {
		return fmt.Errorf("archive one-pager facts %s for deleted %s %s: %w", factsID, subjectType, subjectID, err)
	}
	return nil
}
