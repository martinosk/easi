package handlers

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"easi/backend/internal/importing/application/commands"
	"easi/backend/internal/importing/application/parsers"
	"easi/backend/internal/importing/application/ports"
	"easi/backend/internal/importing/application/saga"
	"easi/backend/internal/importing/infrastructure/repositories"
)

func TestConfirmImportHandler_WrapsLoadSessionErrorWithContext(t *testing.T) {
	eventStore := newInMemoryEventStore()
	repo := repositories.NewImportSessionRepository(eventStore)
	handler := newConfirmImportHandlerForTests(repo, stubComponentGateway{}, context.Background(), 2*time.Second)

	sessionID := "7f89f0c3-a0f5-4ce0-9f78-f4ccfe668d95"
	_, err := handler.Handle(context.Background(), &commands.ConfirmImport{ID: sessionID})
	if err == nil {
		t.Fatal("expected error for missing import session")
	}
	if !strings.Contains(err.Error(), "load import session "+sessionID) {
		t.Fatalf("expected wrapped context with session id, got %v", err)
	}
	if !errors.Is(err, repositories.ErrImportSessionNotFound) {
		t.Fatalf("expected wrapped error to preserve cause %v, got %v", repositories.ErrImportSessionNotFound, err)
	}
}

type stubComponentGateway struct {
	createComponent func(ctx context.Context, name, description string) (string, error)
}

func (s stubComponentGateway) CreateComponent(ctx context.Context, name, description string) (string, error) {
	if s.createComponent != nil {
		return s.createComponent(ctx, name, description)
	}
	return "component-1", nil
}

func (s stubComponentGateway) CreateRelation(_ context.Context, _, _, _, _, _ string) (string, error) {
	return "relation-1", nil
}

type stubCapabilityGateway struct{}

func (s stubCapabilityGateway) CreateCapability(_ context.Context, _, _, _, _ string) (string, error) {
	return "capability-1", nil
}

func (s stubCapabilityGateway) UpdateMetadata(_ context.Context, _, _, _ string) error {
	return nil
}

func (s stubCapabilityGateway) LinkSystem(_ context.Context, _, _, _, _ string) (string, error) {
	return "link-1", nil
}

func (s stubCapabilityGateway) AssignToDomain(_ context.Context, _, _ string) error {
	return nil
}

type stubValueStreamGateway struct{}

func (s stubValueStreamGateway) CreateValueStream(_ context.Context, _, _ string) (string, error) {
	return "valuestream-1", nil
}

func (s stubValueStreamGateway) AddStage(_ context.Context, _, _, _ string) (string, error) {
	return "stage-1", nil
}

func (s stubValueStreamGateway) MapCapabilityToStage(_ context.Context, _, _, _ string) error {
	return nil
}

func TestConfirmImportHandler_FailureScenarios(t *testing.T) {
	scenarios := []struct {
		name                 string
		executionTimeout     time.Duration
		createComponentFunc  func(context.Context, string, string) (string, error)
		expectedReasonSubstr string
	}{
		{
			name:             "PanicMarksSessionFailed",
			executionTimeout: 2 * time.Second,
			createComponentFunc: func(_ context.Context, _, _ string) (string, error) {
				panic("boom")
			},
			expectedReasonSubstr: "panic",
		},
		{
			name:             "TimeoutMarksSessionFailed",
			executionTimeout: 20 * time.Millisecond,
			createComponentFunc: func(ctx context.Context, _, _ string) (string, error) {
				<-ctx.Done()
				time.Sleep(50 * time.Millisecond)
				return "", ctx.Err()
			},
			expectedReasonSubstr: "timed out",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			eventStore := newInMemoryEventStore()
			repo := repositories.NewImportSessionRepository(eventStore)
			sessionID := createImportSessionForConfirmTests(t, repo)

			componentGateway := stubComponentGateway{
				createComponent: scenario.createComponentFunc,
			}
			handler := newConfirmImportHandlerForTests(repo, componentGateway, context.Background(), scenario.executionTimeout)

			_, err := handler.Handle(context.Background(), &commands.ConfirmImport{ID: sessionID})
			if err != nil {
				t.Fatalf("expected no synchronous error, got %v", err)
			}

			waitForImportFailed(t, repo, sessionID)

			reason := findImportFailedReason(t, eventStore, sessionID)
			if !strings.Contains(reason, scenario.expectedReasonSubstr) {
				t.Fatalf("expected failure reason to mention %q, got %q", scenario.expectedReasonSubstr, reason)
			}
		})
	}
}

func TestConfirmImportHandler_CancelledParentContextMarksSessionFailed(t *testing.T) {
	eventStore := newInMemoryEventStore()
	repo := repositories.NewImportSessionRepository(eventStore)
	sessionID := createImportSessionForConfirmTests(t, repo)

	started := make(chan struct{}, 1)
	cancelObserved := make(chan struct{}, 1)

	componentGateway := stubComponentGateway{
		createComponent: func(ctx context.Context, _, _ string) (string, error) {
			started <- struct{}{}
			<-ctx.Done()
			cancelObserved <- struct{}{}
			time.Sleep(50 * time.Millisecond)
			return "", ctx.Err()
		},
	}
	parentCtx, cancelParent := context.WithCancel(context.Background())
	handler := newConfirmImportHandlerForTests(repo, componentGateway, parentCtx, 2*time.Second)

	_, err := handler.Handle(context.Background(), &commands.ConfirmImport{ID: sessionID})
	if err != nil {
		t.Fatalf("expected no synchronous error, got %v", err)
	}

	select {
	case <-started:
	case <-time.After(1 * time.Second):
		t.Fatal("import execution did not start in time")
	}

	cancelParent()

	select {
	case <-cancelObserved:
	case <-time.After(1 * time.Second):
		t.Fatal("import execution did not observe parent context cancellation")
	}

	waitForImportFailed(t, repo, sessionID)

	reason := findImportFailedReason(t, eventStore, sessionID)
	if !strings.Contains(reason, "cancelled") {
		t.Fatalf("expected failure reason to mention cancellation, got %q", reason)
	}
}

func newConfirmImportHandlerForTests(repo *repositories.ImportSessionRepository, componentGateway ports.ComponentGateway, parentCtx context.Context, executionTimeout time.Duration) *ConfirmImportHandler {
	importSaga := saga.New(componentGateway, stubCapabilityGateway{}, stubValueStreamGateway{})
	return NewConfirmImportHandlerWithExecutionContext(repo, importSaga, parentCtx, executionTimeout)
}

func createImportSessionForConfirmTests(t *testing.T, repo *repositories.ImportSessionRepository) string {
	t.Helper()

	createHandler := NewCreateImportSessionHandler(repo)
	result, err := createHandler.Handle(context.Background(), &commands.CreateImportSession{
		SourceFormat: "archimate-openexchange",
		ParseResult: &parsers.ParseResult{
			Components: []parsers.ParsedElement{{SourceID: "comp-1", Name: "Component 1"}},
		},
	})
	if err != nil {
		t.Fatalf("failed to create import session: %v", err)
	}

	return result.CreatedID
}

func waitForImportFailed(t *testing.T, repo *repositories.ImportSessionRepository, sessionID string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		session, err := repo.GetByID(context.Background(), sessionID)
		if err == nil && session.Status().IsFailed() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("import session %s did not reach failed status in time", sessionID)
}

func findImportFailedReason(t *testing.T, eventStore *inMemoryEventStore, sessionID string) string {
	t.Helper()

	events, err := eventStore.GetEvents(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("failed to load events: %v", err)
	}

	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		if event.EventType() != "ImportFailed" {
			continue
		}
		reason, ok := event.EventData()["reason"].(string)
		if !ok {
			t.Fatalf("failed to read import failure reason from event data")
		}
		return reason
	}

	t.Fatalf("no ImportFailed event found for session %s", sessionID)
	return ""
}
