package handlers

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"easi/backend/internal/importing/application/commands"
	"easi/backend/internal/importing/application/parsers"
	"easi/backend/internal/importing/domain/aggregates"
	"easi/backend/internal/importing/infrastructure/repositories"
	domain "easi/backend/internal/shared/eventsourcing"
)

type inMemoryEventStore struct {
	mu     sync.RWMutex
	events map[string][]domain.DomainEvent
}

func newInMemoryEventStore() *inMemoryEventStore {
	return &inMemoryEventStore{
		events: make(map[string][]domain.DomainEvent),
	}
}

func (s *inMemoryEventStore) SaveEvents(_ context.Context, aggregateID string, events []domain.DomainEvent, expectedVersion int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.events[aggregateID]
	if len(existing) != expectedVersion {
		return domain.ErrConcurrencyConflict
	}

	for _, evt := range events {
		jsonData, _ := json.Marshal(evt.EventData())
		stored := domain.NewGenericDomainEvent(aggregateID, evt.EventType(), jsonData, evt.OccurredAt())
		s.events[aggregateID] = append(s.events[aggregateID], stored)
	}
	return nil
}

func (s *inMemoryEventStore) GetEvents(_ context.Context, aggregateID string) ([]domain.DomainEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.events[aggregateID]
	if len(events) == 0 {
		return nil, nil
	}

	result := make([]domain.DomainEvent, len(events))
	copy(result, events)
	return result, nil
}

func TestCreateImportSessionHandler_IncludesValueStreamsInParsedData(t *testing.T) {
	es := newInMemoryEventStore()
	repo := repositories.NewImportSessionRepository(es)
	handler := NewCreateImportSessionHandler(repo)

	cmd := &commands.CreateImportSession{
		SourceFormat: "archimate-openexchange",
		ParseResult: &parsers.ParseResult{
			Capabilities: []parsers.ParsedElement{
				{SourceID: "cap-1", Name: "Order Management"},
			},
			Components: []parsers.ParsedElement{
				{SourceID: "comp-1", Name: "CRM System"},
			},
			ValueStreams: []parsers.ParsedElement{
				{SourceID: "vs-1", Name: "Order to Cash", Description: "End-to-end order fulfillment"},
				{SourceID: "vs-2", Name: "Procure to Pay", Description: "Procurement process"},
			},
			Relationships: []parsers.ParsedRelationship{
				{SourceID: "rel-1", Type: "Serving", SourceRef: "cap-1", TargetRef: "vs-1"},
			},
		},
	}

	result, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	sessionID := result.CreatedID
	session, err := repo.GetByID(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("failed to load session: %v", err)
	}

	valueStreams := session.ParsedData().ValueStreams

	if len(valueStreams) != 2 {
		t.Fatalf("expected 2 value streams in parsed data, got %d", len(valueStreams))
	}

	assertParsedElement(t, valueStreams[0], aggregates.ParsedElement{
		SourceID: "vs-1", Name: "Order to Cash", Description: "End-to-end order fulfillment",
	})
	assertParsedElement(t, valueStreams[1], aggregates.ParsedElement{
		SourceID: "vs-2", Name: "Procure to Pay", Description: "Procurement process",
	})
}

func assertParsedElement(t *testing.T, actual, expected aggregates.ParsedElement) {
	t.Helper()
	if actual.SourceID != expected.SourceID {
		t.Errorf("expected source ID %q, got %q", expected.SourceID, actual.SourceID)
	}
	if actual.Name != expected.Name {
		t.Errorf("expected name %q, got %q", expected.Name, actual.Name)
	}
	if actual.Description != expected.Description {
		t.Errorf("expected description %q, got %q", expected.Description, actual.Description)
	}
}
