package projectors

import (
	"context"
	"encoding/json"
	"testing"

	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/events"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDirectionStore struct {
	inserts             []readmodels.InsertDirectionParams
	statusUpdates       map[string]string
	narrativeUpdates    map[string]string
	horizonUpdates      map[string]string
	placementUpdates    map[string][]readmodels.DirectionPlacementDTO
	sourceReplaceCalls  map[string][]string
}

func newMockDirectionStore() *mockDirectionStore {
	return &mockDirectionStore{
		statusUpdates:      map[string]string{},
		narrativeUpdates:   map[string]string{},
		horizonUpdates:     map[string]string{},
		placementUpdates:   map[string][]readmodels.DirectionPlacementDTO{},
		sourceReplaceCalls: map[string][]string{},
	}
}

func (m *mockDirectionStore) Insert(_ context.Context, p readmodels.InsertDirectionParams) error {
	m.inserts = append(m.inserts, p)
	return nil
}
func (m *mockDirectionStore) UpdateStatus(_ context.Context, id, status string) error {
	m.statusUpdates[id] = status
	return nil
}
func (m *mockDirectionStore) UpdateNarrative(_ context.Context, id, narrative string) error {
	m.narrativeUpdates[id] = narrative
	return nil
}
func (m *mockDirectionStore) UpdateHorizon(_ context.Context, id, horizon string) error {
	m.horizonUpdates[id] = horizon
	return nil
}
func (m *mockDirectionStore) UpdatePlacements(_ context.Context, id string, placements []readmodels.DirectionPlacementDTO) error {
	m.placementUpdates[id] = placements
	return nil
}
func (m *mockDirectionStore) ReplaceSourceCapabilities(_ context.Context, id string, sources []string) error {
	m.sourceReplaceCalls[id] = sources
	return nil
}

func projectViaJSON(t *testing.T, projector *DirectionProjector, eventType string, payload map[string]interface{}) error {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	return projector.ProjectEvent(context.Background(), eventType, data)
}

func TestDirectionProjector_Drafted_InsertsRow(t *testing.T) {
	store := newMockDirectionStore()
	projector := NewDirectionProjector(store)

	id := uuid.New().String()
	ec := uuid.New().String()
	src1, src2 := uuid.New().String(), uuid.New().String()
	dom := uuid.New().String()
	evt := events.NewDirectionDraftedFromFields(events.DirectionDraftedFields{
		ID:                     id,
		EnterpriseCapabilityID: ec,
		Type:                   "consolidate",
		SourceCapabilityIDs:    []string{src1, src2},
		Placements:             []events.PlacementData{{TargetBusinessDomainID: dom, ResultingName: "Unified"}},
		Horizon:                "next",
		Narrative:              "Some narrative.",
	})

	require.NoError(t, projectViaJSON(t, projector, evt.EventType(), evt.EventData()))

	require.Len(t, store.inserts, 1)
	got := store.inserts[0]
	assert.Equal(t, id, got.ID)
	assert.Equal(t, ec, got.EnterpriseCapabilityID)
	assert.Equal(t, "consolidate", got.Type)
	assert.Equal(t, "draft", got.Status)
	assert.Equal(t, "next", got.Horizon)
	assert.Equal(t, []string{src1, src2}, got.SourceCapabilityIDs)
	require.Len(t, got.Placements, 1)
	assert.Equal(t, dom, got.Placements[0].TargetBusinessDomainID)
}

func TestDirectionProjector_StatusEvents_UpdateStatus(t *testing.T) {
	store := newMockDirectionStore()
	projector := NewDirectionProjector(store)
	id := uuid.New().String()

	cases := []struct {
		event events.DirectionProposed
		expectedStatus string
		eventType      string
	}{}
	_ = cases // illustrative; we'll do manually

	require.NoError(t, projectViaJSON(t, projector, "DirectionProposed", events.NewDirectionProposed(id).EventData()))
	assert.Equal(t, "proposed", store.statusUpdates[id])

	require.NoError(t, projectViaJSON(t, projector, "DirectionAgreed", events.NewDirectionAgreed(id).EventData()))
	assert.Equal(t, "agreed", store.statusUpdates[id])

	require.NoError(t, projectViaJSON(t, projector, "DirectionRejected", events.NewDirectionRejected(id).EventData()))
	assert.Equal(t, "rejected", store.statusUpdates[id])
}

func TestDirectionProjector_NarrativeUpdated(t *testing.T) {
	store := newMockDirectionStore()
	projector := NewDirectionProjector(store)
	id := uuid.New().String()

	evt := events.NewDirectionNarrativeUpdated(id, "Refined.")
	require.NoError(t, projectViaJSON(t, projector, evt.EventType(), evt.EventData()))

	assert.Equal(t, "Refined.", store.narrativeUpdates[id])
}

func TestDirectionProjector_HorizonChanged(t *testing.T) {
	store := newMockDirectionStore()
	projector := NewDirectionProjector(store)
	id := uuid.New().String()

	evt := events.NewDirectionHorizonChanged(id, "later")
	require.NoError(t, projectViaJSON(t, projector, evt.EventType(), evt.EventData()))

	assert.Equal(t, "later", store.horizonUpdates[id])
}

func TestDirectionProjector_PlacementsChanged(t *testing.T) {
	store := newMockDirectionStore()
	projector := NewDirectionProjector(store)
	id := uuid.New().String()

	dom := uuid.New().String()
	evt := events.NewDirectionPlacementsChanged(id, []events.PlacementData{{TargetBusinessDomainID: dom}})
	require.NoError(t, projectViaJSON(t, projector, evt.EventType(), evt.EventData()))

	require.Len(t, store.placementUpdates[id], 1)
	assert.Equal(t, dom, store.placementUpdates[id][0].TargetBusinessDomainID)
}

func TestDirectionProjector_SourceCapabilitiesChanged(t *testing.T) {
	store := newMockDirectionStore()
	projector := NewDirectionProjector(store)
	id := uuid.New().String()

	src1, src2, src3 := uuid.New().String(), uuid.New().String(), uuid.New().String()
	evt := events.NewDirectionSourceCapabilitiesChanged(id, []string{src1, src2, src3})
	require.NoError(t, projectViaJSON(t, projector, evt.EventType(), evt.EventData()))

	assert.Equal(t, []string{src1, src2, src3}, store.sourceReplaceCalls[id])
}

func TestDirectionProjector_UnknownEvent_Ignored(t *testing.T) {
	store := newMockDirectionStore()
	projector := NewDirectionProjector(store)
	require.NoError(t, projector.ProjectEvent(context.Background(), "Unknown", []byte("{}")))
	assert.Empty(t, store.inserts)
}
