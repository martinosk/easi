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
	updatedSourceCapabilityNames   []updateSourceCapabilityNameCall
	updatedComponentNames          []updateComponentNameCall
	insertErr                      error
	insertInheritedErr             error
	updateErr                      error
	deleteErr                      error
	deleteBySourceErr              error
	deleteInheritedBySourceCapsErr error
	deleteByComponentErr           error
	updateSourceCapNameErr         error
	updateCompNameErr              error
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

type updateSourceCapabilityNameCall struct {
	CapabilityID   string
	CapabilityName string
}

type updateComponentNameCall struct {
	ComponentID   string
	ComponentName string
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

func (m *mockRealizationReadModel) UpdateSourceCapabilityName(ctx context.Context, capabilityID, capabilityName string) error {
	if m.updateSourceCapNameErr != nil {
		return m.updateSourceCapNameErr
	}
	m.updatedSourceCapabilityNames = append(m.updatedSourceCapabilityNames, updateSourceCapabilityNameCall{
		CapabilityID:   capabilityID,
		CapabilityName: capabilityName,
	})
	return nil
}

func (m *mockRealizationReadModel) UpdateComponentName(ctx context.Context, componentID, componentName string) error {
	if m.updateCompNameErr != nil {
		return m.updateCompNameErr
	}
	m.updatedComponentNames = append(m.updatedComponentNames, updateComponentNameCall{
		ComponentID:   componentID,
		ComponentName: componentName,
	})
	return nil
}

func newProjector(mockRM *mockRealizationReadModel) *RealizationProjector {
	return NewRealizationProjector(mockRM, nil)
}

func TestRealizationProjector_HandleSystemLinked_InsertsDirectOnly(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	projector := newProjector(mockRealRM)

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

func TestRealizationProjector_HandleRealizationUpdated_UpdatesReadModel(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	projector := newProjector(mockRealRM)

	event := events.SystemRealizationUpdated{
		ID:               "real-1",
		RealizationLevel: "Full",
		Notes:            "updated notes",
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "SystemRealizationUpdated", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.updatedRealizations, 1)
	assert.Equal(t, "real-1", mockRealRM.updatedRealizations[0].ID)
	assert.Equal(t, "Full", mockRealRM.updatedRealizations[0].RealizationLevel)
	assert.Equal(t, "updated notes", mockRealRM.updatedRealizations[0].Notes)
}

func TestRealizationProjector_HandleCapabilityRealizationsInherited_AppliesAll(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	projector := newProjector(mockRealRM)

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
	projector := newProjector(mockRealRM)

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
	projector := newProjector(mockRealRM)

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
	projector := newProjector(mockRealRM)

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

func TestRealizationProjector_HandleCapabilityUpdated_UpdatesSourceCapabilityName(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	projector := newProjector(mockRealRM)

	event := events.CapabilityUpdated{ID: "cap-1", Name: "Renamed Capability"}
	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "CapabilityUpdated", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.updatedSourceCapabilityNames, 1)
	assert.Equal(t, "cap-1", mockRealRM.updatedSourceCapabilityNames[0].CapabilityID)
	assert.Equal(t, "Renamed Capability", mockRealRM.updatedSourceCapabilityNames[0].CapabilityName)
}

func TestRealizationProjector_HandleApplicationComponentUpdated_UpdatesComponentName(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	projector := newProjector(mockRealRM)

	eventData, err := json.Marshal(map[string]string{"id": "comp-1", "name": "New Component Name"})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ApplicationComponentUpdated", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.updatedComponentNames, 1)
	assert.Equal(t, "comp-1", mockRealRM.updatedComponentNames[0].ComponentID)
	assert.Equal(t, "New Component Name", mockRealRM.updatedComponentNames[0].ComponentName)
}

func TestRealizationProjector_HandleApplicationComponentDeleted_DeletesByComponentID(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	projector := newProjector(mockRealRM)

	eventData, err := json.Marshal(map[string]string{"id": "comp-1"})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ApplicationComponentDeleted", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.deletedByComponentIDs, 1)
	assert.Equal(t, "comp-1", mockRealRM.deletedByComponentIDs[0])
}

func TestRealizationProjector_UnknownEventType_NoOp(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	projector := newProjector(mockRealRM)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte(`{}`))
	require.NoError(t, err)
	assert.Empty(t, mockRealRM.insertedRealizations)
	assert.Empty(t, mockRealRM.insertedInheritedRealizations)
	assert.Empty(t, mockRealRM.deletedInheritedBySourceCaps)
}
