package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRealizationReadModel struct {
	insertedRealizations           []readmodels.RealizationDTO
	insertedInheritedRealizations  []readmodels.RealizationDTO
	updatedRealizations            []updateCall
	deletedIDs                     []string
	deletedBySourceIDs             []string
	deletedInheritedBySourceCaps   []deleteInheritedBySourceCapsCall
	deletedByComponentIDs          []string
	insertErr                      error
	insertInheritedErr             error
	updateErr                      error
	deleteErr                      error
	deleteBySourceErr              error
	deleteInheritedBySourceCapsErr error
	deleteByComponentErr           error
}

type updateCall struct {
	ID               string
	RealizationLevel string
	Notes            string
}

type deleteInheritedBySourceCapsCall struct {
	SourceRealizationID string
	CapabilityIDs       []string
}

func (m *mockRealizationReadModel) Insert(ctx context.Context, dto readmodels.RealizationDTO) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.insertedRealizations = append(m.insertedRealizations, dto)
	return nil
}

func (m *mockRealizationReadModel) InsertInherited(ctx context.Context, dto readmodels.RealizationDTO) error {
	if m.insertInheritedErr != nil {
		return m.insertInheritedErr
	}
	m.insertedInheritedRealizations = append(m.insertedInheritedRealizations, dto)
	return nil
}

func (m *mockRealizationReadModel) Update(ctx context.Context, id, realizationLevel, notes string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedRealizations = append(m.updatedRealizations, updateCall{ID: id, RealizationLevel: realizationLevel, Notes: notes})
	return nil
}

func (m *mockRealizationReadModel) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedIDs = append(m.deletedIDs, id)
	return nil
}

func (m *mockRealizationReadModel) DeleteBySourceRealizationID(ctx context.Context, sourceRealizationID string) error {
	if m.deleteBySourceErr != nil {
		return m.deleteBySourceErr
	}
	m.deletedBySourceIDs = append(m.deletedBySourceIDs, sourceRealizationID)
	return nil
}

func (m *mockRealizationReadModel) DeleteByComponentID(ctx context.Context, componentID string) error {
	if m.deleteByComponentErr != nil {
		return m.deleteByComponentErr
	}
	m.deletedByComponentIDs = append(m.deletedByComponentIDs, componentID)
	return nil
}

func (m *mockRealizationReadModel) DeleteInheritedBySourceRealizationIDAndCapabilities(ctx context.Context, sourceRealizationID string, capabilityIDs []string) error {
	if m.deleteInheritedBySourceCapsErr != nil {
		return m.deleteInheritedBySourceCapsErr
	}
	m.deletedInheritedBySourceCaps = append(m.deletedInheritedBySourceCaps, deleteInheritedBySourceCapsCall{
		SourceRealizationID: sourceRealizationID,
		CapabilityIDs:       capabilityIDs,
	})
	return nil
}

type realizationReadModelInterface interface {
	Insert(ctx context.Context, dto readmodels.RealizationDTO) error
	InsertInherited(ctx context.Context, dto readmodels.RealizationDTO) error
	Update(ctx context.Context, id, realizationLevel, notes string) error
	Delete(ctx context.Context, id string) error
	DeleteBySourceRealizationID(ctx context.Context, sourceRealizationID string) error
	DeleteInheritedBySourceRealizationIDAndCapabilities(ctx context.Context, sourceRealizationID string, capabilityIDs []string) error
	DeleteByComponentID(ctx context.Context, componentID string) error
}

type testableRealizationProjector struct {
	readModel realizationReadModelInterface
}

func newTestableRealizationProjector(readModel realizationReadModelInterface) *testableRealizationProjector {
	return &testableRealizationProjector{readModel: readModel}
}

func (p *testableRealizationProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "SystemLinkedToCapability":
		return p.handleSystemLinked(ctx, eventData)
	case "SystemRealizationUpdated":
		return p.handleRealizationUpdated(ctx, eventData)
	case "SystemRealizationDeleted":
		return p.handleRealizationDeleted(ctx, eventData)
	case "CapabilityRealizationsInherited":
		return p.handleCapabilityRealizationsInherited(ctx, eventData)
	case "CapabilityRealizationsUninherited":
		return p.handleCapabilityRealizationsUninherited(ctx, eventData)
	case "ApplicationComponentDeleted":
		return p.handleApplicationComponentDeleted(ctx, eventData)
	}
	return nil
}

func (p *testableRealizationProjector) handleApplicationComponentDeleted(ctx context.Context, eventData []byte) error {
	var event applicationComponentDeletedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.DeleteByComponentID(ctx, event.ID)
}

func (p *testableRealizationProjector) handleSystemLinked(ctx context.Context, eventData []byte) error {
	var event events.SystemLinkedToCapability
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	dto := readmodels.RealizationDTO{
		ID:               event.ID,
		CapabilityID:     event.CapabilityID,
		ComponentID:      event.ComponentID,
		ComponentName:    event.ComponentName,
		RealizationLevel: event.RealizationLevel,
		Notes:            event.Notes,
		Origin:           "Direct",
		LinkedAt:         event.LinkedAt,
	}

	return p.readModel.Insert(ctx, dto)
}

func (p *testableRealizationProjector) handleRealizationUpdated(ctx context.Context, eventData []byte) error {
	var event events.SystemRealizationUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.Update(ctx, event.ID, event.RealizationLevel, event.Notes)
}

func (p *testableRealizationProjector) handleRealizationDeleted(ctx context.Context, eventData []byte) error {
	var event events.SystemRealizationDeleted
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	if err := p.readModel.DeleteBySourceRealizationID(ctx, event.ID); err != nil {
		return err
	}
	return p.readModel.Delete(ctx, event.ID)
}

func (p *testableRealizationProjector) handleCapabilityRealizationsInherited(ctx context.Context, eventData []byte) error {
	var event events.CapabilityRealizationsInherited
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	for _, realization := range event.InheritedRealizations {
		dto := readmodels.RealizationDTO{
			CapabilityID:         realization.CapabilityID,
			ComponentID:          realization.ComponentID,
			ComponentName:        realization.ComponentName,
			RealizationLevel:     realization.RealizationLevel,
			Notes:                realization.Notes,
			Origin:               realization.Origin,
			SourceRealizationID:  realization.SourceRealizationID,
			SourceCapabilityID:   realization.SourceCapabilityID,
			SourceCapabilityName: realization.SourceCapabilityName,
			LinkedAt:             realization.LinkedAt,
		}
		if err := p.readModel.InsertInherited(ctx, dto); err != nil {
			return err
		}
	}

	return nil
}

func (p *testableRealizationProjector) handleCapabilityRealizationsUninherited(ctx context.Context, eventData []byte) error {
	var event events.CapabilityRealizationsUninherited
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	for _, removal := range event.Removals {
		if err := p.readModel.DeleteInheritedBySourceRealizationIDAndCapabilities(ctx, removal.SourceRealizationID, removal.CapabilityIDs); err != nil {
			return err
		}
	}

	return nil
}

func TestRealizationProjector_HandleSystemLinked_InsertsDirectOnly(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	projector := newTestableRealizationProjector(mockRealRM)

	event := events.SystemLinkedToCapability{
		ID:               "real-1",
		CapabilityID:     "cap-l2",
		ComponentID:      "comp-1",
		ComponentName:    "Component A",
		RealizationLevel: "Partial",
		Notes:            "note",
		LinkedAt:         time.Now(),
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "SystemLinkedToCapability", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.insertedRealizations, 1)
	assert.Equal(t, "real-1", mockRealRM.insertedRealizations[0].ID)
	assert.Equal(t, "Direct", mockRealRM.insertedRealizations[0].Origin)
	assert.Empty(t, mockRealRM.insertedInheritedRealizations)
}

func TestRealizationProjector_HandleCapabilityRealizationsInherited_AppliesAll(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	projector := newTestableRealizationProjector(mockRealRM)

	event := events.CapabilityRealizationsInherited{
		CapabilityID: "cap-a",
		InheritedRealizations: []events.InheritedRealization{
			{
				CapabilityID:         "cap-parent",
				ComponentID:          "comp-1",
				ComponentName:        "Component A",
				RealizationLevel:     "Full",
				Origin:               "Inherited",
				SourceRealizationID:  "real-1",
				SourceCapabilityID:   "cap-a",
				SourceCapabilityName: "A",
				LinkedAt:             time.Now(),
			},
			{
				CapabilityID:         "cap-root",
				ComponentID:          "comp-1",
				ComponentName:        "Component A",
				RealizationLevel:     "Full",
				Origin:               "Inherited",
				SourceRealizationID:  "real-1",
				SourceCapabilityID:   "cap-a",
				SourceCapabilityName: "A",
				LinkedAt:             time.Now(),
			},
		},
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "CapabilityRealizationsInherited", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.insertedInheritedRealizations, 2)
	assert.Equal(t, "cap-parent", mockRealRM.insertedInheritedRealizations[0].CapabilityID)
	assert.Equal(t, "cap-root", mockRealRM.insertedInheritedRealizations[1].CapabilityID)
	assert.Equal(t, "real-1", mockRealRM.insertedInheritedRealizations[0].SourceRealizationID)
}

func TestRealizationProjector_HandleCapabilityRealizationsUninherited_AppliesRemovals(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	projector := newTestableRealizationProjector(mockRealRM)

	event := events.CapabilityRealizationsUninherited{
		CapabilityID: "cap-a",
		Removals: []events.RealizationInheritanceRemoval{
			{SourceRealizationID: "real-1", CapabilityIDs: []string{"cap-old-parent", "cap-old-root"}},
			{SourceRealizationID: "real-2", CapabilityIDs: []string{"cap-old-parent"}},
		},
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "CapabilityRealizationsUninherited", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.deletedInheritedBySourceCaps, 2)
	assert.Equal(t, "real-1", mockRealRM.deletedInheritedBySourceCaps[0].SourceRealizationID)
	assert.ElementsMatch(t, []string{"cap-old-parent", "cap-old-root"}, mockRealRM.deletedInheritedBySourceCaps[0].CapabilityIDs)
	assert.Equal(t, "real-2", mockRealRM.deletedInheritedBySourceCaps[1].SourceRealizationID)
}

func TestRealizationProjector_HandleCapabilityRealizationsInherited_StopsOnError(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{insertInheritedErr: errors.New("insert failed")}
	projector := newTestableRealizationProjector(mockRealRM)

	event := events.CapabilityRealizationsInherited{
		CapabilityID: "cap-a",
		InheritedRealizations: []events.InheritedRealization{
			{CapabilityID: "cap-parent", ComponentID: "comp-1", RealizationLevel: "Full", Origin: "Inherited", SourceRealizationID: "real-1", LinkedAt: time.Now()},
		},
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "CapabilityRealizationsInherited", eventData)
	require.Error(t, err)
}

func TestRealizationProjector_HandleRealizationDeleted_CascadesInheritedDeletion(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	projector := newTestableRealizationProjector(mockRealRM)

	event := events.SystemRealizationDeleted{ID: "real-1", DeletedAt: time.Now()}
	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "SystemRealizationDeleted", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.deletedBySourceIDs, 1)
	require.Len(t, mockRealRM.deletedIDs, 1)
	assert.Equal(t, "real-1", mockRealRM.deletedBySourceIDs[0])
	assert.Equal(t, "real-1", mockRealRM.deletedIDs[0])
}

func TestRealizationProjector_UnknownEventType_NoOp(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	projector := newTestableRealizationProjector(mockRealRM)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte(`{}`))
	require.NoError(t, err)
	assert.Empty(t, mockRealRM.insertedRealizations)
	assert.Empty(t, mockRealRM.insertedInheritedRealizations)
	assert.Empty(t, mockRealRM.deletedInheritedBySourceCaps)
}
