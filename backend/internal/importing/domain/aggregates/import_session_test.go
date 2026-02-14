package aggregates

import (
	"testing"

	"easi/backend/internal/importing/domain/valueobjects"
)

func TestNewImportSession(t *testing.T) {
	sourceFormat, _ := valueobjects.NewSourceFormat("archimate-openexchange")
	preview := valueobjects.NewImportPreview(
		valueobjects.SupportedCounts{Capabilities: 10, Components: 5},
		valueobjects.UnsupportedCounts{},
	)
	parsedData := ParsedData{
		Capabilities: []ParsedElement{{SourceID: "cap-1", Name: "Cap 1"}},
	}

	session, err := NewImportSession(ImportSessionConfig{
		SourceFormat:      sourceFormat,
		BusinessDomainID:  "",
		CapabilityEAOwner: "",
		Preview:           preview,
		ParsedData:        parsedData,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if session.ID() == "" {
		t.Error("expected non-empty ID")
	}
	if !session.Status().IsPending() {
		t.Error("expected status to be pending")
	}
	if !session.SourceFormat().IsArchiMateOpenExchange() {
		t.Error("expected source format to be archimate-openexchange")
	}
	if session.Preview().Supported().Capabilities != 10 {
		t.Errorf("expected 10 capabilities, got %d", session.Preview().Supported().Capabilities)
	}

	uncommitted := session.GetUncommittedChanges()
	if len(uncommitted) != 1 {
		t.Errorf("expected 1 uncommitted event, got %d", len(uncommitted))
	}
	if uncommitted[0].EventType() != "ImportSessionCreated" {
		t.Errorf("expected ImportSessionCreated event, got %s", uncommitted[0].EventType())
	}
}

func TestImportSession_StartImport(t *testing.T) {
	session := createTestSession(t)

	err := session.StartImport()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !session.Status().IsImporting() {
		t.Error("expected status to be importing")
	}

	uncommitted := session.GetUncommittedChanges()
	found := false
	for _, e := range uncommitted {
		if e.EventType() == "ImportStarted" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ImportStarted event to be raised")
	}
}

func TestImportSession_StartImport_AlreadyStarted(t *testing.T) {
	session := createTestSession(t)
	_ = session.StartImport()

	err := session.StartImport()
	if err == nil {
		t.Error("expected error when starting already started import")
	}
	if err != ErrImportAlreadyStarted {
		t.Errorf("expected ErrImportAlreadyStarted, got %v", err)
	}
}

func TestImportSession_UpdateProgress(t *testing.T) {
	session := createTestSession(t)
	_ = session.StartImport()

	progress, _ := valueobjects.NewImportProgress("creating_components", 100, 50)
	err := session.UpdateProgress(progress)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if session.Progress().Phase() != "creating_components" {
		t.Errorf("expected phase 'creating_components', got %q", session.Progress().Phase())
	}
}

func TestImportSession_UpdateProgress_NotImporting(t *testing.T) {
	session := createTestSession(t)

	progress, _ := valueobjects.NewImportProgress("creating_components", 100, 50)
	err := session.UpdateProgress(progress)
	if err == nil {
		t.Error("expected error when updating progress of non-importing session")
	}
	if err != ErrImportNotStarted {
		t.Errorf("expected ErrImportNotStarted, got %v", err)
	}
}

func TestImportSession_Complete(t *testing.T) {
	session := createTestSession(t)
	_ = session.StartImport()

	result := ImportResult{
		CapabilitiesCreated: 10,
		ComponentsCreated:   5,
		RealizationsCreated: 3,
		DomainAssignments:   2,
	}
	err := session.Complete(result)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !session.Status().IsCompleted() {
		t.Error("expected status to be completed")
	}
	if session.Result().CapabilitiesCreated != 10 {
		t.Errorf("expected 10 capabilities created, got %d", session.Result().CapabilitiesCreated)
	}
}

func TestImportSession_Fail(t *testing.T) {
	session := createTestSession(t)
	_ = session.StartImport()

	err := session.Fail("Something went wrong")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !session.Status().IsFailed() {
		t.Error("expected status to be failed")
	}
}

func TestImportSession_Cancel(t *testing.T) {
	session := createTestSession(t)

	err := session.Cancel()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !session.IsCancelled() {
		t.Error("expected session to be cancelled")
	}
}

func TestImportSession_Cancel_AlreadyStarted(t *testing.T) {
	session := createTestSession(t)
	_ = session.StartImport()

	err := session.Cancel()
	if err == nil {
		t.Error("expected error when cancelling started import")
	}
	if err != ErrCannotCancelStartedImport {
		t.Errorf("expected ErrCannotCancelStartedImport, got %v", err)
	}
}

func TestImportSession_BusinessDomainID(t *testing.T) {
	sourceFormat, _ := valueobjects.NewSourceFormat("archimate-openexchange")
	preview := valueobjects.NewImportPreview(
		valueobjects.SupportedCounts{},
		valueobjects.UnsupportedCounts{},
	)
	parsedData := ParsedData{}

	session, _ := NewImportSession(ImportSessionConfig{
		SourceFormat:      sourceFormat,
		BusinessDomainID:  "domain-123",
		CapabilityEAOwner: "",
		Preview:           preview,
		ParsedData:        parsedData,
	})

	if session.BusinessDomainID() != "domain-123" {
		t.Errorf("expected 'domain-123', got %q", session.BusinessDomainID())
	}
}

func createTestSession(t *testing.T) *ImportSession {
	t.Helper()
	sourceFormat, _ := valueobjects.NewSourceFormat("archimate-openexchange")
	preview := valueobjects.NewImportPreview(
		valueobjects.SupportedCounts{Capabilities: 10, Components: 5},
		valueobjects.UnsupportedCounts{},
	)
	parsedData := ParsedData{
		Capabilities: []ParsedElement{{SourceID: "cap-1", Name: "Cap 1"}},
	}

	session, err := NewImportSession(ImportSessionConfig{
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
