package saga_test

import (
	"errors"
	"testing"

	"easi/backend/internal/importing/domain/aggregates"
)

func TestImportSaga_CreatesComponents(t *testing.T) {
	f := newFixture()
	data := aggregates.ParsedData{
		Components: []aggregates.ParsedElement{
			{SourceID: "src-1", Name: "Frontend", Description: "UI app"},
			{SourceID: "src-2", Name: "Backend", Description: "API server"},
		},
	}

	result := f.execute(t, data, "", "")

	expectCount(t, "components", result.ComponentsCreated, 2)
	assertNoErrors(t, result)
}

func TestImportSaga_CreatesCapabilitiesWithHierarchy(t *testing.T) {
	f := newFixture()
	data := aggregates.ParsedData{
		Capabilities: []aggregates.ParsedElement{
			{SourceID: "cap-root", Name: "RootCap"},
			{SourceID: "cap-child", Name: "ChildCap"},
		},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "rel-1", Type: "Aggregation", SourceRef: "cap-root", TargetRef: "cap-child"},
		},
	}

	result := f.execute(t, data, "", "")

	expectCount(t, "capabilities", result.CapabilitiesCreated, 2)
	if len(f.capGw.createCalls) != 2 {
		t.Fatalf("expected 2 CreateCapability calls, got %d", len(f.capGw.createCalls))
	}

	rootCall := f.capGw.createCalls[0]
	if rootCall.ParentID != "" {
		t.Errorf("root should have empty parentID, got %q", rootCall.ParentID)
	}
	if rootCall.Level != "L1" {
		t.Errorf("root should be L1, got %q", rootCall.Level)
	}

	childCall := f.capGw.createCalls[1]
	if childCall.ParentID != "cap-RootCap" {
		t.Errorf("child parentID: expected %q, got %q", "cap-RootCap", childCall.ParentID)
	}
	if childCall.Level != "L2" {
		t.Errorf("child should be L2, got %q", childCall.Level)
	}
}

func TestImportSaga_CompositionRelationshipBuildsHierarchy(t *testing.T) {
	f := newFixture()
	data := aggregates.ParsedData{
		Capabilities: []aggregates.ParsedElement{
			{SourceID: "cap-parent", Name: "Parent"},
			{SourceID: "cap-child", Name: "Child"},
		},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "rel-1", Type: "Composition", SourceRef: "cap-parent", TargetRef: "cap-child"},
		},
	}

	result := f.execute(t, data, "", "")

	expectCount(t, "capabilities", result.CapabilitiesCreated, 2)
	if len(f.capGw.createCalls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(f.capGw.createCalls))
	}
	if f.capGw.createCalls[1].ParentID != "cap-Parent" {
		t.Errorf("child should have parentID from Composition, got %q", f.capGw.createCalls[1].ParentID)
	}
}

func TestImportSaga_CreatesRealizationsUsingIDMappings(t *testing.T) {
	f := newFixture()
	data := aggregates.ParsedData{
		Components:   []aggregates.ParsedElement{{SourceID: "comp-1", Name: "MyComp"}},
		Capabilities: []aggregates.ParsedElement{{SourceID: "cap-1", Name: "MyCap"}},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "real-1", Type: "Realization", SourceRef: "comp-1", TargetRef: "cap-1", Name: "realizes"},
		},
	}

	result := f.execute(t, data, "", "")

	expectCount(t, "components", result.ComponentsCreated, 1)
	expectCount(t, "capabilities", result.CapabilitiesCreated, 1)
	expectCount(t, "realizations", result.RealizationsCreated, 1)
}

func TestImportSaga_SkipsRealizationsForUnmappedRefs(t *testing.T) {
	f := newFixture()
	data := aggregates.ParsedData{
		Components: []aggregates.ParsedElement{{SourceID: "comp-1", Name: "MyComp"}},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "real-1", Type: "Realization", SourceRef: "comp-1", TargetRef: "unmapped-cap"},
		},
	}

	result := f.execute(t, data, "", "")

	expectCount(t, "realizations", result.RealizationsCreated, 0)
	expectCount(t, "LinkSystem calls", len(f.capGw.linkSystemCalls), 0)
}

func TestImportSaga_PassesNotesToLinkSystem(t *testing.T) {
	f := newFixture()
	data := aggregates.ParsedData{
		Components:   []aggregates.ParsedElement{{SourceID: "comp-1", Name: "MyComp"}},
		Capabilities: []aggregates.ParsedElement{{SourceID: "cap-1", Name: "MyCap"}},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "real-1", Type: "Realization", SourceRef: "comp-1", TargetRef: "cap-1", Name: "implements", Documentation: "full integration"},
		},
	}

	f.execute(t, data, "", "")

	if len(f.capGw.linkSystemCalls) != 1 {
		t.Fatalf("expected 1 LinkSystem call, got %d", len(f.capGw.linkSystemCalls))
	}
	if f.capGw.linkSystemCalls[0].Notes != "implements - full integration" {
		t.Errorf("expected notes %q, got %q", "implements - full integration", f.capGw.linkSystemCalls[0].Notes)
	}
}

func TestImportSaga_AccumulatesItemErrors(t *testing.T) {
	f := newFixture()
	f.compGw.err = errors.New("create failed")

	data := aggregates.ParsedData{
		Components: []aggregates.ParsedElement{{SourceID: "src-1", Name: "Comp1"}},
	}

	result := f.execute(t, data, "", "")

	expectCount(t, "components", result.ComponentsCreated, 0)
	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0].Action() != "skipped" {
		t.Errorf("expected action 'skipped', got %q", result.Errors[0].Action())
	}
}

func TestImportSaga_AssignsMetadataToCreatedCapabilities(t *testing.T) {
	f := newFixture()
	data := aggregates.ParsedData{
		Capabilities: []aggregates.ParsedElement{{SourceID: "cap-1", Name: "Cap1"}},
	}

	result := f.execute(t, data, "", "john@example.com")

	expectCount(t, "capabilities", result.CapabilitiesCreated, 1)
	if len(f.capGw.metadataCalls) != 1 {
		t.Fatalf("expected 1 UpdateMetadata call, got %d", len(f.capGw.metadataCalls))
	}
	if f.capGw.metadataCalls[0].EAOwner != "john@example.com" {
		t.Errorf("expected eaOwner %q, got %q", "john@example.com", f.capGw.metadataCalls[0].EAOwner)
	}
	if f.capGw.metadataCalls[0].Status != "Active" {
		t.Errorf("expected status %q, got %q", "Active", f.capGw.metadataCalls[0].Status)
	}
}

func TestImportSaga_SkipsMetadataWhenNoEAOwner(t *testing.T) {
	f := newFixture()
	data := aggregates.ParsedData{
		Capabilities: []aggregates.ParsedElement{{SourceID: "cap-1", Name: "Cap1"}},
	}

	f.execute(t, data, "", "")

	expectCount(t, "UpdateMetadata calls", len(f.capGw.metadataCalls), 0)
}

func TestImportSaga_CreatesValueStreamsWithDefaultStage(t *testing.T) {
	f := newFixture()
	data := aggregates.ParsedData{
		ValueStreams: []aggregates.ParsedElement{
			{SourceID: "vs-1", Name: "CustomerJourney", Description: "End-to-end"},
		},
	}

	result := f.execute(t, data, "", "")

	expectCount(t, "value streams", result.ValueStreamsCreated, 1)
}

func TestImportSaga_CreatesDomainAssignments(t *testing.T) {
	f := newFixture()
	data := aggregates.ParsedData{
		Capabilities: []aggregates.ParsedElement{{SourceID: "cap-1", Name: "RootCap"}},
	}

	result := f.execute(t, data, "domain-123", "")

	expectCount(t, "capabilities", result.CapabilitiesCreated, 1)
	expectCount(t, "domain assignments", result.DomainAssignments, 1)
}

func TestImportSaga_SkipsDomainAssignmentWhenNoDomainID(t *testing.T) {
	f := newFixture()
	data := aggregates.ParsedData{
		Capabilities: []aggregates.ParsedElement{{SourceID: "cap-1", Name: "Cap1"}},
	}

	result := f.execute(t, data, "", "")

	expectCount(t, "domain assignments", result.DomainAssignments, 0)
}

func TestImportSaga_MapsServingRelationType(t *testing.T) {
	f := newFixture()
	data := aggregates.ParsedData{
		Components: []aggregates.ParsedElement{
			{SourceID: "comp-a", Name: "ServiceA"},
			{SourceID: "comp-b", Name: "ServiceB"},
			{SourceID: "comp-c", Name: "ServiceC"},
		},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "rel-1", Type: "Serving", SourceRef: "comp-a", TargetRef: "comp-b", Name: "serves"},
			{SourceID: "rel-2", Type: "Triggering", SourceRef: "comp-b", TargetRef: "comp-c", Name: "triggers"},
		},
	}

	result := f.execute(t, data, "", "")

	expectCount(t, "component relations", result.ComponentRelationsCreated, 2)
	if len(f.compGw.relationCalls) != 2 {
		t.Fatalf("expected 2 CreateRelation calls, got %d", len(f.compGw.relationCalls))
	}
	if f.compGw.relationCalls[0].RelationType != "Serves" {
		t.Errorf("Serving should map to 'Serves', got %q", f.compGw.relationCalls[0].RelationType)
	}
	if f.compGw.relationCalls[1].RelationType != "Triggers" {
		t.Errorf("Triggering should map to 'Triggers', got %q", f.compGw.relationCalls[1].RelationType)
	}
}

func TestImportSaga_FailureIsBestEffortAndKeepsSuccessfulProgress(t *testing.T) {
	f := newFixture()
	f.compGw.createErrByName["Broken"] = errors.New("component create failed")

	data := aggregates.ParsedData{
		Components: []aggregates.ParsedElement{
			{SourceID: "comp-1", Name: "Broken"},
			{SourceID: "comp-2", Name: "Healthy"},
		},
		Capabilities: []aggregates.ParsedElement{{SourceID: "cap-1", Name: "Cap1"}},
	}

	result := f.execute(t, data, "", "")

	expectCount(t, "components", result.ComponentsCreated, 1)
	expectCount(t, "capabilities", result.CapabilitiesCreated, 1)
	if len(result.Errors) != 1 {
		t.Fatalf("expected exactly one error, got %d", len(result.Errors))
	}
	if result.Errors[0].Action() != "skipped" {
		t.Fatalf("expected skipped action for failed item, got %q", result.Errors[0].Action())
	}
}

func TestImportSaga_CapabilityFailureDoesNotPreventValueStreamCreation(t *testing.T) {
	f := newFixture()
	f.capGw.createErrByName["BrokenCap"] = errors.New("capability create failed")

	data := aggregates.ParsedData{
		Capabilities: []aggregates.ParsedElement{
			{SourceID: "cap-1", Name: "BrokenCap"},
			{SourceID: "cap-2", Name: "HealthyCap"},
		},
		ValueStreams: []aggregates.ParsedElement{
			{SourceID: "vs-1", Name: "MyStream"},
		},
	}

	result := f.execute(t, data, "", "")

	expectCount(t, "capabilities", result.CapabilitiesCreated, 1)
	expectCount(t, "value streams", result.ValueStreamsCreated, 1)
	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestImportSaga_ValueStreamFailureDoesNotPreventRealizationCreation(t *testing.T) {
	f := newFixture()
	f.vsGw.createErrByName["BrokenStream"] = errors.New("value stream create failed")

	data := aggregates.ParsedData{
		Components:   []aggregates.ParsedElement{{SourceID: "comp-1", Name: "MyComp"}},
		Capabilities: []aggregates.ParsedElement{{SourceID: "cap-1", Name: "MyCap"}},
		ValueStreams: []aggregates.ParsedElement{{SourceID: "vs-1", Name: "BrokenStream"}},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "rel-1", Type: "Realization", SourceRef: "comp-1", TargetRef: "cap-1"},
		},
	}

	result := f.execute(t, data, "", "")

	expectCount(t, "components", result.ComponentsCreated, 1)
	expectCount(t, "capabilities", result.CapabilitiesCreated, 1)
	expectCount(t, "value streams", result.ValueStreamsCreated, 0)
	expectCount(t, "realizations", result.RealizationsCreated, 1)
	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestImportSaga_RealizationFailureDoesNotPreventDomainAssignment(t *testing.T) {
	f := newFixture()
	f.capGw.linkErrByKey["comp-MyComp-cap-MyCap"] = errors.New("link failed")

	data := aggregates.ParsedData{
		Components:   []aggregates.ParsedElement{{SourceID: "comp-1", Name: "MyComp"}},
		Capabilities: []aggregates.ParsedElement{{SourceID: "cap-1", Name: "MyCap"}},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "rel-1", Type: "Realization", SourceRef: "comp-1", TargetRef: "cap-1"},
		},
	}

	result := f.execute(t, data, "domain-1", "")

	expectCount(t, "components", result.ComponentsCreated, 1)
	expectCount(t, "capabilities", result.CapabilitiesCreated, 1)
	expectCount(t, "realizations", result.RealizationsCreated, 0)
	expectCount(t, "domain assignments", result.DomainAssignments, 1)
	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestImportSaga_FullImportWithAllPhases(t *testing.T) {
	f := newFixture()
	data := aggregates.ParsedData{
		Components:   []aggregates.ParsedElement{{SourceID: "comp-1", Name: "Frontend"}},
		Capabilities: []aggregates.ParsedElement{{SourceID: "cap-1", Name: "UserMgmt"}},
		ValueStreams: []aggregates.ParsedElement{{SourceID: "vs-1", Name: "Onboarding"}},
		Relationships: []aggregates.ParsedRelationship{
			{SourceID: "rel-1", Type: "Realization", SourceRef: "comp-1", TargetRef: "cap-1"},
			{SourceID: "rel-2", Type: "Association", SourceRef: "cap-1", TargetRef: "vs-1"},
		},
	}

	result := f.execute(t, data, "domain-1", "owner@co.com")

	expectCount(t, "components", result.ComponentsCreated, 1)
	expectCount(t, "capabilities", result.CapabilitiesCreated, 1)
	expectCount(t, "value streams", result.ValueStreamsCreated, 1)
	expectCount(t, "realizations", result.RealizationsCreated, 1)
	expectCount(t, "domain assignments", result.DomainAssignments, 1)
	expectCount(t, "capability mappings", result.CapabilityMappings, 1)
	assertNoErrors(t, result)
}
