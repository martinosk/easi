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

type fakeEntityStore struct {
	createdIDs      map[string]string
	createErrByName map[string]error
	err             error
}

func newFakeEntityStore() fakeEntityStore {
	return fakeEntityStore{
		createdIDs:      make(map[string]string),
		createErrByName: make(map[string]error),
	}
}

func (s *fakeEntityStore) create(prefix, name string) (string, error) {
	if err, ok := s.createErrByName[name]; ok {
		return "", err
	}
	if s.err != nil {
		return "", s.err
	}
	id := prefix + name
	s.createdIDs[name] = id
	return id, nil
}

type fakeComponentGateway struct {
	fakeEntityStore
	relationCalls []relationCreateCall
}

func newFakeComponentGateway() *fakeComponentGateway {
	return &fakeComponentGateway{fakeEntityStore: newFakeEntityStore()}
}

func (f *fakeComponentGateway) CreateComponent(_ context.Context, name, _ string) (string, error) {
	return f.create("comp-", name)
}

func (f *fakeComponentGateway) CreateRelation(
	_ context.Context, sourceID, targetID, relationType, name, description string,
) (string, error) {
	f.relationCalls = append(f.relationCalls, relationCreateCall{sourceID, targetID, relationType, name, description})
	if f.err != nil {
		return "", f.err
	}
	return "rel-" + sourceID + "-" + targetID + "-" + relationType, nil
}

type fakeCapabilityGateway struct {
	fakeEntityStore
	createCalls     []capabilityCreateCall
	metadataCalls   []metadataUpdateCall
	linkSystemCalls []linkSystemCall
	linkErrByKey    map[string]error
}

func newFakeCapabilityGateway() *fakeCapabilityGateway {
	return &fakeCapabilityGateway{
		fakeEntityStore: newFakeEntityStore(),
		linkErrByKey:    make(map[string]error),
	}
}

func (f *fakeCapabilityGateway) CreateCapability(
	_ context.Context, name, description, parentID, level string,
) (string, error) {
	f.createCalls = append(f.createCalls, capabilityCreateCall{name, description, parentID, level})
	return f.create("cap-", name)
}

func (f *fakeCapabilityGateway) UpdateMetadata(_ context.Context, id, eaOwner, status string) error {
	f.metadataCalls = append(f.metadataCalls, metadataUpdateCall{id, eaOwner, status})
	return f.err
}

func (f *fakeCapabilityGateway) LinkSystem(
	_ context.Context, capabilityID, componentID, realizationLevel, notes string,
) (string, error) {
	f.linkSystemCalls = append(f.linkSystemCalls, linkSystemCall{capabilityID, componentID, realizationLevel, notes})
	key := componentID + "-" + capabilityID
	if err, ok := f.linkErrByKey[key]; ok {
		return "", err
	}
	if f.err != nil {
		return "", f.err
	}
	return "real-" + key, nil
}

func (f *fakeCapabilityGateway) AssignToDomain(_ context.Context, _, _ string) error {
	return f.err
}

type fakeValueStreamGateway struct {
	fakeEntityStore
	stageIDs map[string]string
}

func newFakeValueStreamGateway() *fakeValueStreamGateway {
	return &fakeValueStreamGateway{
		fakeEntityStore: newFakeEntityStore(),
		stageIDs:        make(map[string]string),
	}
}

func (f *fakeValueStreamGateway) CreateValueStream(_ context.Context, name, _ string) (string, error) {
	return f.create("vs-", name)
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
