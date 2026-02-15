package saga_test

import (
	"context"
	"testing"

	"easi/backend/internal/importing/application/saga"
	"easi/backend/internal/importing/domain/aggregates"
)

type capabilityCreateCall struct {
	Name, Description, ParentID, Level string
}

type metadataUpdateCall struct {
	ID, EAOwner, Status string
}

type linkSystemCall struct {
	CapabilityID, ComponentID, RealizationLevel, Notes string
}

type relationCreateCall struct {
	SourceID, TargetID, RelationType, Name, Description string
}

type fakeComponentGateway struct {
	createdIDs    map[string]string
	relationCalls []relationCreateCall
	err           error
}

func newFakeComponentGateway() *fakeComponentGateway {
	return &fakeComponentGateway{createdIDs: make(map[string]string)}
}

func (f *fakeComponentGateway) CreateComponent(_ context.Context, name, _ string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	id := "comp-" + name
	f.createdIDs[name] = id
	return id, nil
}

func (f *fakeComponentGateway) CreateRelation(_ context.Context, sourceID, targetID, relationType, name, description string) (string, error) {
	f.relationCalls = append(f.relationCalls, relationCreateCall{sourceID, targetID, relationType, name, description})
	if f.err != nil {
		return "", f.err
	}
	return "rel-" + sourceID + "-" + targetID + "-" + relationType, nil
}

type fakeCapabilityGateway struct {
	createdIDs      map[string]string
	createCalls     []capabilityCreateCall
	metadataCalls   []metadataUpdateCall
	linkSystemCalls []linkSystemCall
	err             error
}

func newFakeCapabilityGateway() *fakeCapabilityGateway {
	return &fakeCapabilityGateway{createdIDs: make(map[string]string)}
}

func (f *fakeCapabilityGateway) CreateCapability(_ context.Context, name, description, parentID, level string) (string, error) {
	f.createCalls = append(f.createCalls, capabilityCreateCall{name, description, parentID, level})
	if f.err != nil {
		return "", f.err
	}
	id := "cap-" + name
	f.createdIDs[name] = id
	return id, nil
}

func (f *fakeCapabilityGateway) UpdateMetadata(_ context.Context, id, eaOwner, status string) error {
	f.metadataCalls = append(f.metadataCalls, metadataUpdateCall{id, eaOwner, status})
	return f.err
}

func (f *fakeCapabilityGateway) LinkSystem(_ context.Context, capabilityID, componentID, realizationLevel, notes string) (string, error) {
	f.linkSystemCalls = append(f.linkSystemCalls, linkSystemCall{capabilityID, componentID, realizationLevel, notes})
	if f.err != nil {
		return "", f.err
	}
	return "real-" + componentID + "-" + capabilityID, nil
}

func (f *fakeCapabilityGateway) AssignToDomain(_ context.Context, _, _ string) error {
	return f.err
}

type fakeValueStreamGateway struct {
	createdIDs map[string]string
	stageIDs   map[string]string
	err        error
}

func newFakeValueStreamGateway() *fakeValueStreamGateway {
	return &fakeValueStreamGateway{
		createdIDs: make(map[string]string),
		stageIDs:   make(map[string]string),
	}
}

func (f *fakeValueStreamGateway) CreateValueStream(_ context.Context, name, _ string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	id := "vs-" + name
	f.createdIDs[name] = id
	return id, nil
}

func (f *fakeValueStreamGateway) AddStage(_ context.Context, vsID, _, _ string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	id := "stage-" + vsID
	f.stageIDs[vsID] = id
	return id, nil
}

func (f *fakeValueStreamGateway) MapCapabilityToStage(_ context.Context, _, _, _ string) error {
	return f.err
}

type fixture struct {
	compGw *fakeComponentGateway
	capGw  *fakeCapabilityGateway
	vsGw   *fakeValueStreamGateway
	saga   *saga.ImportSaga
}

func newFixture() fixture {
	compGw := newFakeComponentGateway()
	capGw := newFakeCapabilityGateway()
	vsGw := newFakeValueStreamGateway()
	return fixture{
		compGw: compGw,
		capGw:  capGw,
		vsGw:   vsGw,
		saga:   saga.New(compGw, capGw, vsGw),
	}
}

func (f fixture) execute(t *testing.T, data aggregates.ParsedData, domainID, eaOwner string) aggregates.ImportResult {
	t.Helper()
	return f.saga.Execute(context.Background(), data, domainID, eaOwner)
}

func expectCount(t *testing.T, label string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s: expected %d, got %d", label, want, got)
	}
}

func assertNoErrors(t *testing.T, result aggregates.ImportResult) {
	t.Helper()
	for _, e := range result.Errors {
		t.Errorf("unexpected import error: %s", e.Error())
	}
}
