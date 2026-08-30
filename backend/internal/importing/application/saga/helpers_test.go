package saga_test

import (
	"context"
	"testing"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/importing/application/saga"
	"easi/backend/internal/importing/domain/aggregates"
	vsPL "easi/backend/internal/valuestreams/publishedlanguage"
)

type fakeEntityStore struct {
	prefix          string
	createdIDs      map[string]string
	createErrByName map[string]error
	err             error
}

func newFakeEntityStore(prefix string) fakeEntityStore {
	return fakeEntityStore{
		prefix:          prefix,
		createdIDs:      make(map[string]string),
		createErrByName: make(map[string]error),
	}
}

func (s *fakeEntityStore) create(name string) (string, error) {
	if err, ok := s.createErrByName[name]; ok {
		return "", err
	}
	if s.err != nil {
		return "", s.err
	}
	id := s.prefix + name
	s.createdIDs[name] = id
	return id, nil
}

type fakeComponentGateway struct {
	fakeEntityStore
	relationCalls []amPL.CreateComponentRelation
}

func newFakeComponentGateway() *fakeComponentGateway {
	return &fakeComponentGateway{fakeEntityStore: newFakeEntityStore("comp-")}
}

func (f *fakeComponentGateway) CreateComponent(_ context.Context, cmd amPL.CreateApplicationComponent) (string, error) {
	return f.create(cmd.Name)
}

func (f *fakeComponentGateway) CreateRelation(_ context.Context, cmd amPL.CreateComponentRelation) (string, error) {
	f.relationCalls = append(f.relationCalls, cmd)
	if f.err != nil {
		return "", f.err
	}
	return "rel-" + cmd.SourceComponentID + "-" + cmd.TargetComponentID + "-" + cmd.RelationType, nil
}

type fakeCapabilityGateway struct {
	fakeEntityStore
	createCalls     []cmPL.CreateCapability
	metadataCalls   []cmPL.UpdateCapabilityMetadata
	linkSystemCalls []cmPL.LinkSystemToCapability
	linkErrByKey    map[string]error
}

func newFakeCapabilityGateway() *fakeCapabilityGateway {
	return &fakeCapabilityGateway{
		fakeEntityStore: newFakeEntityStore("cap-"),
		linkErrByKey:    make(map[string]error),
	}
}

func (f *fakeCapabilityGateway) CreateCapability(_ context.Context, cmd cmPL.CreateCapability) (string, error) {
	f.createCalls = append(f.createCalls, cmd)
	return f.create(cmd.Name)
}

func (f *fakeCapabilityGateway) UpdateMetadata(_ context.Context, cmd cmPL.UpdateCapabilityMetadata) error {
	f.metadataCalls = append(f.metadataCalls, cmd)
	return f.err
}

func (f *fakeCapabilityGateway) LinkSystem(_ context.Context, cmd cmPL.LinkSystemToCapability) (string, error) {
	f.linkSystemCalls = append(f.linkSystemCalls, cmd)
	key := cmd.ComponentID + "-" + cmd.CapabilityID
	if err, ok := f.linkErrByKey[key]; ok {
		return "", err
	}
	if f.err != nil {
		return "", f.err
	}
	return "real-" + key, nil
}

func (f *fakeCapabilityGateway) AssignToDomain(_ context.Context, _ cmPL.AssignCapabilityToDomain) error {
	return f.err
}

type fakeValueStreamGateway struct {
	fakeEntityStore
	stageIDs map[string]string
}

func newFakeValueStreamGateway() *fakeValueStreamGateway {
	return &fakeValueStreamGateway{
		fakeEntityStore: newFakeEntityStore("vs-"),
		stageIDs:        make(map[string]string),
	}
}

func (f *fakeValueStreamGateway) CreateValueStream(_ context.Context, cmd vsPL.CreateValueStream) (string, error) {
	return f.create(cmd.Name)
}

func (f *fakeValueStreamGateway) AddStage(_ context.Context, cmd vsPL.AddStage) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	id := "stage-" + cmd.ValueStreamID
	f.stageIDs[cmd.ValueStreamID] = id
	return id, nil
}

func (f *fakeValueStreamGateway) MapCapabilityToStage(_ context.Context, _ vsPL.AddStageCapability) error {
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

func assertImportCounts(t *testing.T, result aggregates.ImportResult, expected map[string]int) {
	t.Helper()
	actual := map[string]int{
		"components":         result.ComponentsCreated,
		"capabilities":       result.CapabilitiesCreated,
		"value streams":      result.ValueStreamsCreated,
		"realizations":       result.RealizationsCreated,
		"domain assignments": result.DomainAssignments,
	}
	for label, want := range expected {
		expectCount(t, label, actual[label], want)
	}
}
