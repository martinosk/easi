package orchestrator

import (
	"context"
	"fmt"
	"testing"

	capabilityCommands "easi/backend/internal/capabilitymapping/application/commands"
	valueStreamCommands "easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/importing/domain/aggregates"
	"easi/backend/internal/importing/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type mockCommandBus struct {
	dispatchedCommands []cqrs.Command
	dispatchError      error
	idCounter          int
}

func (m *mockCommandBus) Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	if m.dispatchError != nil {
		return cqrs.EmptyResult(), m.dispatchError
	}
	m.dispatchedCommands = append(m.dispatchedCommands, cmd)
	m.idCounter++

	createdID := m.generateIDForCommand(cmd)
	return cqrs.NewResult(createdID), nil
}

func (m *mockCommandBus) generateIDForCommand(cmd cqrs.Command) string {
	prefixes := map[string]string{
		"CreateApplicationComponent": "comp-",
		"CreateComponentRelation":    "rel-",
		"CreateCapability":           "cap-",
		"LinkSystemToCapability":     "link-",
		"AssignCapabilityToDomain":   "assign-",
		"CreateValueStream":          "vs-",
		"AddStage":                   "stage-",
		"AddStageCapability":         "mapping-",
	}
	if prefix, ok := prefixes[cmd.CommandName()]; ok {
		return fmt.Sprintf("%s%d", prefix, m.idCounter)
	}
	return ""
}

func (m *mockCommandBus) Register(name string, handler cqrs.CommandHandler) {}

type mockRepository struct {
	session   *aggregates.ImportSession
	saveError error
	saved     []*aggregates.ImportSession
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*aggregates.ImportSession, error) {
	return m.session, nil
}

func (m *mockRepository) Save(ctx context.Context, session *aggregates.ImportSession) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.saved = append(m.saved, session)
	return nil
}

func TestImportOrchestrator_Execute_CreatesComponents(t *testing.T) {
	commandBus := &mockCommandBus{}

	session := createTestSession(t, aggregates.ParsedData{
		Components: []aggregates.ParsedElement{
			{SourceID: "comp-1", Name: "Component 1", Description: "Desc 1"},
			{SourceID: "comp-2", Name: "Component 2", Description: "Desc 2"},
		},
	})
	session.StartImport()

	repo := &mockRepository{session: session}

	orchestrator := NewImportOrchestrator(commandBus, repo)
	result, err := orchestrator.Execute(context.Background(), session)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ComponentsCreated != 2 {
		t.Errorf("expected 2 components created, got %d", result.ComponentsCreated)
	}

	componentCommands := 0
	for _, cmd := range commandBus.dispatchedCommands {
		if cmd.CommandName() == "CreateApplicationComponent" {
			componentCommands++
		}
	}
	if componentCommands != 2 {
		t.Errorf("expected 2 CreateApplicationComponent commands, got %d", componentCommands)
	}
}

func TestImportOrchestrator_Execute_CreatesCapabilitiesInHierarchyOrder(t *testing.T) {
	parsedData := aggregates.ParsedData{
		Capabilities: []aggregates.ParsedElement{
			{SourceID: "cap-1", Name: "Parent Capability"},
			{SourceID: "cap-2", Name: "Child Capability"},
		},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "rel-1", Type: "Aggregation", SourceRef: "cap-1", TargetRef: "cap-2"},
		},
	}
	result := executeOrchestrator(t, parsedData)
	if result.CapabilitiesCreated != 2 {
		t.Errorf("expected 2 capabilities created, got %d", result.CapabilitiesCreated)
	}
}

func TestImportOrchestrator_Execute_CreatesRealizations(t *testing.T) {
	parsedData := aggregates.ParsedData{
		Components: []aggregates.ParsedElement{
			{SourceID: "comp-1", Name: "Component 1"},
		},
		Capabilities: []aggregates.ParsedElement{
			{SourceID: "cap-1", Name: "Capability 1"},
		},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "rel-1", Type: "Realization", SourceRef: "comp-1", TargetRef: "cap-1", Name: "Supports", Documentation: "Notes here"},
		},
	}
	result := executeOrchestrator(t, parsedData)
	if result.RealizationsCreated != 1 {
		t.Errorf("expected 1 realization created, got %d", result.RealizationsCreated)
	}
}

func TestImportOrchestrator_Execute_CapsDeepLevelsToL4(t *testing.T) {
	commandBus := &mockCommandBus{}

	session := createTestSession(t, aggregates.ParsedData{
		Capabilities: []aggregates.ParsedElement{
			{SourceID: "cap-l1", Name: "L1 Capability"},
			{SourceID: "cap-l2", Name: "L2 Capability"},
			{SourceID: "cap-l3", Name: "L3 Capability"},
			{SourceID: "cap-l4", Name: "L4 Capability"},
			{SourceID: "cap-l5", Name: "L5 Capability (should be capped to L4)"},
			{SourceID: "cap-l6", Name: "L6 Capability (should be capped to L4)"},
		},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "rel-1", Type: "Composition", SourceRef: "cap-l1", TargetRef: "cap-l2"},
			{SourceID: "rel-2", Type: "Composition", SourceRef: "cap-l2", TargetRef: "cap-l3"},
			{SourceID: "rel-3", Type: "Composition", SourceRef: "cap-l3", TargetRef: "cap-l4"},
			{SourceID: "rel-4", Type: "Composition", SourceRef: "cap-l4", TargetRef: "cap-l5"},
			{SourceID: "rel-5", Type: "Composition", SourceRef: "cap-l5", TargetRef: "cap-l6"},
		},
	})
	session.StartImport()

	repo := &mockRepository{session: session}

	orchestrator := NewImportOrchestrator(commandBus, repo)
	result, err := orchestrator.Execute(context.Background(), session)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.CapabilitiesCreated != 6 {
		t.Errorf("expected 6 capabilities created, got %d", result.CapabilitiesCreated)
	}

	expectedLevels := map[string]string{
		"L1 Capability":                          "L1",
		"L2 Capability":                          "L2",
		"L3 Capability":                          "L3",
		"L4 Capability":                          "L4",
		"L5 Capability (should be capped to L4)": "L4",
		"L6 Capability (should be capped to L4)": "L4",
	}

	for _, cmd := range commandBus.dispatchedCommands {
		if createCmd, ok := cmd.(*capabilityCommands.CreateCapability); ok {
			expectedLevel := expectedLevels[createCmd.Name]
			if createCmd.Level != expectedLevel {
				t.Errorf("capability %q: expected level %s, got %s", createCmd.Name, expectedLevel, createCmd.Level)
			}
		}
	}
}

func createTestSession(t *testing.T, parsedData aggregates.ParsedData) *aggregates.ImportSession {
	t.Helper()
	sourceFormat, _ := valueobjects.NewSourceFormat("archimate-openexchange")
	preview := valueobjects.NewImportPreview(
		valueobjects.SupportedCounts{},
		valueobjects.UnsupportedCounts{},
	)

	session, err := aggregates.NewImportSession(aggregates.ImportSessionConfig{
		SourceFormat:      sourceFormat,
		BusinessDomainID:  "",
		CapabilityEAOwner: "",
		Preview:           preview,
		ParsedData:        parsedData,
	})
	if err != nil {
		t.Fatalf("failed to create test session: %v", err)
	}
	return session
}

func TestImportOrchestrator_Execute_CreatesValueStreams(t *testing.T) {
	commandBus := &mockCommandBus{}

	session := createTestSession(t, aggregates.ParsedData{
		ValueStreams: []aggregates.ParsedElement{
			{SourceID: "vs-1", Name: "Order to Cash", Description: "End-to-end order fulfillment"},
			{SourceID: "vs-2", Name: "Procure to Pay", Description: "Procurement process"},
		},
	})
	session.StartImport()

	repo := &mockRepository{session: session}
	orchestrator := NewImportOrchestrator(commandBus, repo)
	result, err := orchestrator.Execute(context.Background(), session)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ValueStreamsCreated != 2 {
		t.Errorf("expected 2 value streams created, got %d", result.ValueStreamsCreated)
	}

	createVSCommands := countCommandsByName(commandBus.dispatchedCommands, "CreateValueStream")
	addStageCommands := countCommandsByName(commandBus.dispatchedCommands, "AddStage")

	if createVSCommands != 2 {
		t.Errorf("expected 2 CreateValueStream commands, got %d", createVSCommands)
	}
	if addStageCommands != 2 {
		t.Errorf("expected 2 AddStage commands (one per ValueStream), got %d", addStageCommands)
	}

	verifyStageNames(t, commandBus.dispatchedCommands, "Main Flow")
}

func countCommandsByName(commands []cqrs.Command, name string) int {
	count := 0
	for _, cmd := range commands {
		if cmd.CommandName() == name {
			count++
		}
	}
	return count
}

func verifyStageNames(t *testing.T, commands []cqrs.Command, expectedName string) {
	for _, cmd := range commands {
		if stageCmd, ok := cmd.(*valueStreamCommands.AddStage); ok {
			if stageCmd.Name != expectedName {
				t.Errorf("expected stage name %q, got %q", expectedName, stageCmd.Name)
			}
		}
	}
}

func executeOrchestrator(t *testing.T, parsedData aggregates.ParsedData) aggregates.ImportResult {
	commandBus := &mockCommandBus{}
	session := createTestSession(t, parsedData)
	session.StartImport()

	repo := &mockRepository{session: session}
	orchestrator := NewImportOrchestrator(commandBus, repo)
	result, err := orchestrator.Execute(context.Background(), session)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	return result
}

func TestImportOrchestrator_Execute_MapsCapabilitiesToValueStreamStages(t *testing.T) {
	commandBus := &mockCommandBus{}

	session := createTestSession(t, aggregates.ParsedData{
		Capabilities: []aggregates.ParsedElement{
			{SourceID: "cap-1", Name: "Order Management"},
		},
		ValueStreams: []aggregates.ParsedElement{
			{SourceID: "vs-1", Name: "Order to Cash"},
		},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "rel-1", Type: "Serving", SourceRef: "cap-1", TargetRef: "vs-1"},
		},
	})
	session.StartImport()

	repo := &mockRepository{session: session}
	orchestrator := NewImportOrchestrator(commandBus, repo)
	result, err := orchestrator.Execute(context.Background(), session)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.CapabilityMappings != 1 {
		t.Errorf("expected 1 capability mapping, got %d", result.CapabilityMappings)
	}

	mappingCommands := 0
	for _, cmd := range commandBus.dispatchedCommands {
		if cmd.CommandName() == "AddStageCapability" {
			mappingCommands++
		}
	}

	if mappingCommands != 1 {
		t.Errorf("expected 1 AddStageCapability command, got %d", mappingCommands)
	}
}

func TestImportOrchestrator_Execute_CapabilityToValueStreamRelationshipsDoNotAffectComponentRelations(t *testing.T) {
	commandBus := &mockCommandBus{}

	session := createTestSession(t, aggregates.ParsedData{
		Components: []aggregates.ParsedElement{
			{SourceID: "comp-1", Name: "Component 1"},
			{SourceID: "comp-2", Name: "Component 2"},
		},
		Capabilities: []aggregates.ParsedElement{
			{SourceID: "cap-1", Name: "Capability 1"},
		},
		ValueStreams: []aggregates.ParsedElement{
			{SourceID: "vs-1", Name: "ValueStream 1"},
		},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "rel-1", Type: "Serving", SourceRef: "comp-1", TargetRef: "comp-2"},
			{SourceID: "rel-2", Type: "Serving", SourceRef: "cap-1", TargetRef: "vs-1"},
		},
	})
	session.StartImport()

	repo := &mockRepository{session: session}
	orchestrator := NewImportOrchestrator(commandBus, repo)
	result, err := orchestrator.Execute(context.Background(), session)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ComponentRelationsCreated != 1 {
		t.Errorf("expected 1 component relation created, got %d", result.ComponentRelationsCreated)
	}
	if result.CapabilityMappings != 1 {
		t.Errorf("expected 1 capability mapping created, got %d", result.CapabilityMappings)
	}
	if len(result.Errors) > 0 {
		t.Errorf("expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	createRelCommands := countCommandsByName(commandBus.dispatchedCommands, "CreateComponentRelation")
	addStageCapsCommands := countCommandsByName(commandBus.dispatchedCommands, "AddStageCapability")

	if createRelCommands != 1 {
		t.Errorf("expected 1 CreateComponentRelation command, got %d", createRelCommands)
	}
	if addStageCapsCommands != 1 {
		t.Errorf("expected 1 AddStageCapability command, got %d", addStageCapsCommands)
	}
}

func TestImportOrchestrator_Execute_CapabilityToValueStreamRealizationDoesNotAffectComponentRealizations(t *testing.T) {
	commandBus := &mockCommandBus{}

	session := createTestSession(t, aggregates.ParsedData{
		Components: []aggregates.ParsedElement{
			{SourceID: "comp-1", Name: "Component 1"},
		},
		Capabilities: []aggregates.ParsedElement{
			{SourceID: "cap-1", Name: "Capability 1"},
			{SourceID: "cap-2", Name: "Capability 2"},
		},
		ValueStreams: []aggregates.ParsedElement{
			{SourceID: "vs-1", Name: "ValueStream 1"},
		},
		Relationships: []aggregates.ParsedRelationship{
			// Component→Capability Realization (should be processed)
			{SourceID: "rel-1", Type: "Realization", SourceRef: "comp-1", TargetRef: "cap-1"},
			// Capability→ValueStream Realization (should NOT be processed by realization phase)
			{SourceID: "rel-2", Type: "Realization", SourceRef: "cap-2", TargetRef: "vs-1"},
		},
	})
	session.StartImport()

	repo := &mockRepository{session: session}
	orchestrator := NewImportOrchestrator(commandBus, repo)
	result, err := orchestrator.Execute(context.Background(), session)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should have 1 realization and 1 capability mapping
	if result.RealizationsCreated != 1 {
		t.Errorf("expected 1 realization created, got %d", result.RealizationsCreated)
	}
	if result.CapabilityMappings != 1 {
		t.Errorf("expected 1 capability mapping created, got %d", result.CapabilityMappings)
	}

	// Verify no errors occurred
	if len(result.Errors) > 0 {
		t.Errorf("expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}
}
