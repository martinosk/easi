package projectors

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRealizationReadModel struct {
	insertedRealizations          []readmodels.RealizationDTO
	insertedInheritedRealizations []readmodels.RealizationDTO
	updatedRealizations           []updateCall
	deletedIDs                    []string
	deletedBySourceIDs            []string
	realizationsByCapability      map[string][]readmodels.RealizationDTO
	insertErr                     error
	insertInheritedErr            error
	updateErr                     error
	deleteErr                     error
	deleteBySourceErr             error
	getByCapabilityErr            error
}

type updateCall struct {
	ID               string
	RealizationLevel string
	Notes            string
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

func (m *mockRealizationReadModel) GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.RealizationDTO, error) {
	if m.getByCapabilityErr != nil {
		return nil, m.getByCapabilityErr
	}
	if m.realizationsByCapability == nil {
		return nil, nil
	}
	return m.realizationsByCapability[capabilityID], nil
}

type mockCapabilityReadModelForProjector struct {
	capabilities map[string]*readmodels.CapabilityDTO
	getErr       error
}

func (m *mockCapabilityReadModelForProjector) GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	cap, ok := m.capabilities[id]
	if !ok {
		return nil, nil
	}
	return cap, nil
}

type realizationReadModelInterface interface {
	Insert(ctx context.Context, dto readmodels.RealizationDTO) error
	InsertInherited(ctx context.Context, dto readmodels.RealizationDTO) error
	Update(ctx context.Context, id, realizationLevel, notes string) error
	Delete(ctx context.Context, id string) error
	DeleteBySourceRealizationID(ctx context.Context, sourceRealizationID string) error
	GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.RealizationDTO, error)
}

type capabilityReadModelInterface interface {
	GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error)
}

type testableRealizationProjector struct {
	readModel           realizationReadModelInterface
	capabilityReadModel capabilityReadModelInterface
}

func newTestableRealizationProjector(
	readModel realizationReadModelInterface,
	capabilityReadModel capabilityReadModelInterface,
) *testableRealizationProjector {
	return &testableRealizationProjector{
		readModel:           readModel,
		capabilityReadModel: capabilityReadModel,
	}
}

func (p *testableRealizationProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "SystemLinkedToCapability":
		return p.handleSystemLinked(ctx, eventData)
	case "SystemRealizationUpdated":
		return p.handleRealizationUpdated(ctx, eventData)
	case "SystemRealizationDeleted":
		return p.handleRealizationDeleted(ctx, eventData)
	case "CapabilityParentChanged":
		return p.handleCapabilityParentChanged(ctx, eventData)
	}
	return nil
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
		RealizationLevel: event.RealizationLevel,
		Notes:            event.Notes,
		Origin:           "Direct",
		LinkedAt:         event.LinkedAt,
	}

	if err := p.readModel.Insert(ctx, dto); err != nil {
		return err
	}

	return p.createInheritedRealizationsForAncestors(ctx, dto)
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

func (p *testableRealizationProjector) createInheritedRealizationsForAncestors(ctx context.Context, source readmodels.RealizationDTO) error {
	capability, err := p.capabilityReadModel.GetByID(ctx, source.CapabilityID)
	if err != nil {
		return err
	}
	if capability == nil || capability.ParentID == "" {
		return nil
	}

	source.SourceCapabilityID = source.CapabilityID
	source.SourceCapabilityName = capability.Name
	nextSource := source
	nextSource.CapabilityID = capability.ParentID
	return p.propagateInheritedRealizations(ctx, nextSource)
}

func (p *testableRealizationProjector) handleCapabilityParentChanged(ctx context.Context, eventData []byte) error {
	var event events.CapabilityParentChanged
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	if event.NewParentID == "" {
		return nil
	}

	realizations, err := p.readModel.GetByCapabilityID(ctx, event.CapabilityID)
	if err != nil {
		return err
	}

	capability, err := p.capabilityReadModel.GetByID(ctx, event.CapabilityID)
	if err != nil {
		return err
	}

	for _, realization := range realizations {
		sourceID := realization.ID
		sourceCapabilityID := event.CapabilityID
		sourceCapabilityName := ""
		if capability != nil {
			sourceCapabilityName = capability.Name
		}

		if realization.Origin == "Inherited" && realization.SourceRealizationID != "" {
			sourceID = realization.SourceRealizationID
			sourceCapabilityID = realization.SourceCapabilityID
			sourceCapabilityName = realization.SourceCapabilityName
		}

		source := readmodels.RealizationDTO{
			ID:                   sourceID,
			CapabilityID:         event.NewParentID,
			ComponentID:          realization.ComponentID,
			ComponentName:        realization.ComponentName,
			SourceCapabilityID:   sourceCapabilityID,
			SourceCapabilityName: sourceCapabilityName,
			LinkedAt:             realization.LinkedAt,
		}

		if err := p.propagateInheritedRealizations(ctx, source); err != nil {
			return err
		}
	}

	return nil
}

func (p *testableRealizationProjector) propagateInheritedRealizations(ctx context.Context, source readmodels.RealizationDTO) error {
	capability, err := p.capabilityReadModel.GetByID(ctx, source.CapabilityID)
	if err != nil {
		return err
	}
	if capability == nil {
		return nil
	}

	inheritedDTO := readmodels.RealizationDTO{
		CapabilityID:         source.CapabilityID,
		ComponentID:          source.ComponentID,
		ComponentName:        source.ComponentName,
		RealizationLevel:     "Full",
		Origin:               "Inherited",
		SourceRealizationID:  source.ID,
		SourceCapabilityID:   source.SourceCapabilityID,
		SourceCapabilityName: source.SourceCapabilityName,
		LinkedAt:             source.LinkedAt,
	}

	if err := p.readModel.InsertInherited(ctx, inheritedDTO); err != nil {
		return err
	}

	if capability.ParentID == "" {
		return nil
	}

	nextSource := source
	nextSource.CapabilityID = capability.ParentID
	return p.propagateInheritedRealizations(ctx, nextSource)
}

func TestRealizationProjector_HandleSystemLinked_CreatesDirectRealization(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	mockCapRM := &mockCapabilityReadModelForProjector{
		capabilities: map[string]*readmodels.CapabilityDTO{
			"cap-l1": {ID: "cap-l1", Name: "L1 Capability", Level: "L1", ParentID: ""},
		},
	}

	projector := newTestableRealizationProjector(mockRealRM, mockCapRM)

	event := events.SystemLinkedToCapability{
		ID:               "real-1",
		CapabilityID:     "cap-l1",
		ComponentID:      "comp-1",
		RealizationLevel: "Full",
		Notes:            "Test realization",
		LinkedAt:         time.Now(),
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "SystemLinkedToCapability", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.insertedRealizations, 1)
	assert.Equal(t, "real-1", mockRealRM.insertedRealizations[0].ID)
	assert.Equal(t, "cap-l1", mockRealRM.insertedRealizations[0].CapabilityID)
	assert.Equal(t, "comp-1", mockRealRM.insertedRealizations[0].ComponentID)
	assert.Equal(t, "Direct", mockRealRM.insertedRealizations[0].Origin)
	assert.Empty(t, mockRealRM.insertedInheritedRealizations, "No inherited realizations should be created for L1 capability without parent")
}

func TestRealizationProjector_HandleSystemLinked_CreatesInheritedRealizationsForAncestors(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	mockCapRM := &mockCapabilityReadModelForProjector{
		capabilities: map[string]*readmodels.CapabilityDTO{
			"cap-l1": {ID: "cap-l1", Name: "L1 Capability", Level: "L1", ParentID: ""},
			"cap-l2": {ID: "cap-l2", Name: "L2 Capability", Level: "L2", ParentID: "cap-l1"},
			"cap-l3": {ID: "cap-l3", Name: "L3 Capability", Level: "L3", ParentID: "cap-l2"},
		},
	}

	projector := newTestableRealizationProjector(mockRealRM, mockCapRM)

	event := events.SystemLinkedToCapability{
		ID:               "real-1",
		CapabilityID:     "cap-l3",
		ComponentID:      "comp-1",
		RealizationLevel: "Partial",
		Notes:            "L3 realization",
		LinkedAt:         time.Now(),
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "SystemLinkedToCapability", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.insertedRealizations, 1, "Should have 1 Direct realization")
	assert.Equal(t, "real-1", mockRealRM.insertedRealizations[0].ID)
	assert.Equal(t, "cap-l3", mockRealRM.insertedRealizations[0].CapabilityID)
	assert.Equal(t, "Direct", mockRealRM.insertedRealizations[0].Origin)

	require.Len(t, mockRealRM.insertedInheritedRealizations, 2, "Should have 2 Inherited realizations (L2 and L1)")

	l2Inherited := mockRealRM.insertedInheritedRealizations[0]
	assert.Equal(t, "cap-l2", l2Inherited.CapabilityID)
	assert.Equal(t, "comp-1", l2Inherited.ComponentID)
	assert.Equal(t, "Full", l2Inherited.RealizationLevel, "Inherited realizations should always be Full")
	assert.Equal(t, "Inherited", l2Inherited.Origin)
	assert.Equal(t, "real-1", l2Inherited.SourceRealizationID)
	assert.Equal(t, "cap-l3", l2Inherited.SourceCapabilityID, "Should track source capability ID")
	assert.Equal(t, "L3 Capability", l2Inherited.SourceCapabilityName, "Should track source capability name")

	l1Inherited := mockRealRM.insertedInheritedRealizations[1]
	assert.Equal(t, "cap-l1", l1Inherited.CapabilityID)
	assert.Equal(t, "comp-1", l1Inherited.ComponentID)
	assert.Equal(t, "Full", l1Inherited.RealizationLevel)
	assert.Equal(t, "Inherited", l1Inherited.Origin)
	assert.Equal(t, "real-1", l1Inherited.SourceRealizationID)
	assert.Equal(t, "cap-l3", l1Inherited.SourceCapabilityID, "All inherited should track same source capability")
	assert.Equal(t, "L3 Capability", l1Inherited.SourceCapabilityName)
}

func TestRealizationProjector_HandleSystemLinked_FourLevelHierarchy(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	mockCapRM := &mockCapabilityReadModelForProjector{
		capabilities: map[string]*readmodels.CapabilityDTO{
			"cap-l1": {ID: "cap-l1", Name: "L1", Level: "L1", ParentID: ""},
			"cap-l2": {ID: "cap-l2", Name: "L2", Level: "L2", ParentID: "cap-l1"},
			"cap-l3": {ID: "cap-l3", Name: "L3", Level: "L3", ParentID: "cap-l2"},
			"cap-l4": {ID: "cap-l4", Name: "L4", Level: "L4", ParentID: "cap-l3"},
		},
	}

	projector := newTestableRealizationProjector(mockRealRM, mockCapRM)

	event := events.SystemLinkedToCapability{
		ID:               "real-1",
		CapabilityID:     "cap-l4",
		ComponentID:      "comp-1",
		RealizationLevel: "Planned",
		LinkedAt:         time.Now(),
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "SystemLinkedToCapability", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.insertedRealizations, 1, "Should have 1 Direct realization")
	require.Len(t, mockRealRM.insertedInheritedRealizations, 3, "Should have 3 Inherited realizations (L3, L2, L1)")

	expectedParents := []string{"cap-l3", "cap-l2", "cap-l1"}
	for i, expected := range expectedParents {
		assert.Equal(t, expected, mockRealRM.insertedInheritedRealizations[i].CapabilityID,
			"Inherited realization %d should be for capability %s", i, expected)
		assert.Equal(t, "real-1", mockRealRM.insertedInheritedRealizations[i].SourceRealizationID,
			"All inherited realizations should reference the source realization")
	}
}

func TestRealizationProjector_HandleRealizationUpdated(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	mockCapRM := &mockCapabilityReadModelForProjector{}

	projector := newTestableRealizationProjector(mockRealRM, mockCapRM)

	event := events.SystemRealizationUpdated{
		ID:               "real-1",
		RealizationLevel: "Partial",
		Notes:            "Updated notes",
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "SystemRealizationUpdated", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.updatedRealizations, 1)
	assert.Equal(t, "real-1", mockRealRM.updatedRealizations[0].ID)
	assert.Equal(t, "Partial", mockRealRM.updatedRealizations[0].RealizationLevel)
	assert.Equal(t, "Updated notes", mockRealRM.updatedRealizations[0].Notes)
}

func TestRealizationProjector_HandleRealizationDeleted_CascadesInheritedDeletion(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	mockCapRM := &mockCapabilityReadModelForProjector{}

	projector := newTestableRealizationProjector(mockRealRM, mockCapRM)

	event := events.SystemRealizationDeleted{
		ID:        "real-1",
		DeletedAt: time.Now(),
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "SystemRealizationDeleted", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.deletedBySourceIDs, 1, "Should cascade delete by source realization ID")
	assert.Equal(t, "real-1", mockRealRM.deletedBySourceIDs[0])

	require.Len(t, mockRealRM.deletedIDs, 1, "Should delete the direct realization")
	assert.Equal(t, "real-1", mockRealRM.deletedIDs[0])
}

func TestRealizationProjector_HandleRealizationDeleted_DeletesInheritedFirst(t *testing.T) {
	deletionOrder := []string{}

	mockCapRM := &mockCapabilityReadModelForProjector{}

	projector := &testableRealizationProjector{
		readModel: &orderTrackingReadModel{
			deletionOrder: &deletionOrder,
		},
		capabilityReadModel: mockCapRM,
	}

	event := events.SystemRealizationDeleted{
		ID:        "real-1",
		DeletedAt: time.Now(),
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "SystemRealizationDeleted", eventData)
	require.NoError(t, err)

	require.Len(t, deletionOrder, 2, "Should have 2 deletion operations")
	assert.Equal(t, "inherited:real-1", deletionOrder[0], "Should delete inherited realizations first")
	assert.Equal(t, "direct:real-1", deletionOrder[1], "Should delete direct realization second")
}

type orderTrackingReadModel struct {
	deletionOrder *[]string
}

func (o *orderTrackingReadModel) Insert(ctx context.Context, dto readmodels.RealizationDTO) error {
	return nil
}

func (o *orderTrackingReadModel) InsertInherited(ctx context.Context, dto readmodels.RealizationDTO) error {
	return nil
}

func (o *orderTrackingReadModel) Update(ctx context.Context, id, realizationLevel, notes string) error {
	return nil
}

func (o *orderTrackingReadModel) Delete(ctx context.Context, id string) error {
	*o.deletionOrder = append(*o.deletionOrder, "direct:"+id)
	return nil
}

func (o *orderTrackingReadModel) DeleteBySourceRealizationID(ctx context.Context, sourceRealizationID string) error {
	*o.deletionOrder = append(*o.deletionOrder, "inherited:"+sourceRealizationID)
	return nil
}

func (o *orderTrackingReadModel) GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.RealizationDTO, error) {
	return nil, nil
}

func TestRealizationProjector_UnknownEventType_NoOp(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	mockCapRM := &mockCapabilityReadModelForProjector{}

	projector := newTestableRealizationProjector(mockRealRM, mockCapRM)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte(`{}`))
	require.NoError(t, err)

	assert.Empty(t, mockRealRM.insertedRealizations)
	assert.Empty(t, mockRealRM.insertedInheritedRealizations)
	assert.Empty(t, mockRealRM.updatedRealizations)
	assert.Empty(t, mockRealRM.deletedIDs)
	assert.Empty(t, mockRealRM.deletedBySourceIDs)
}

func TestRealizationProjector_HandleSystemLinked_CapabilityNotFound_StopsRecursion(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	mockCapRM := &mockCapabilityReadModelForProjector{
		capabilities: map[string]*readmodels.CapabilityDTO{},
	}

	projector := newTestableRealizationProjector(mockRealRM, mockCapRM)

	event := events.SystemLinkedToCapability{
		ID:               "real-1",
		CapabilityID:     "non-existent-cap",
		ComponentID:      "comp-1",
		RealizationLevel: "Full",
		LinkedAt:         time.Now(),
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "SystemLinkedToCapability", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.insertedRealizations, 1, "Should still insert the direct realization")
	assert.Empty(t, mockRealRM.insertedInheritedRealizations, "Should not create inherited realizations when capability not found")
}

func TestRealizationProjector_InheritedRealizationsAlwaysUseFull(t *testing.T) {
	testCases := []struct {
		name             string
		realizationLevel string
	}{
		{"Planned becomes Full", "Planned"},
		{"Partial becomes Full", "Partial"},
		{"Full stays Full", "Full"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRealRM := &mockRealizationReadModel{}
			mockCapRM := &mockCapabilityReadModelForProjector{
				capabilities: map[string]*readmodels.CapabilityDTO{
					"cap-l1": {ID: "cap-l1", Name: "L1", Level: "L1", ParentID: ""},
					"cap-l2": {ID: "cap-l2", Name: "L2", Level: "L2", ParentID: "cap-l1"},
				},
			}

			projector := newTestableRealizationProjector(mockRealRM, mockCapRM)

			event := events.SystemLinkedToCapability{
				ID:               "real-1",
				CapabilityID:     "cap-l2",
				ComponentID:      "comp-1",
				RealizationLevel: tc.realizationLevel,
				LinkedAt:         time.Now(),
			}

			eventData, err := json.Marshal(event)
			require.NoError(t, err)

			err = projector.ProjectEvent(context.Background(), "SystemLinkedToCapability", eventData)
			require.NoError(t, err)

			assert.Equal(t, tc.realizationLevel, mockRealRM.insertedRealizations[0].RealizationLevel,
				"Direct realization should keep original level")

			require.Len(t, mockRealRM.insertedInheritedRealizations, 1)
			assert.Equal(t, "Full", mockRealRM.insertedInheritedRealizations[0].RealizationLevel,
				"Inherited realization should always be Full regardless of source level")
		})
	}
}

func TestRealizationProjector_HandleCapabilityParentChanged_PropagatesInheritedRealizations(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{
		realizationsByCapability: map[string][]readmodels.RealizationDTO{
			"cap-A": {
				{
					ID:                  "inherited-to-A",
					CapabilityID:        "cap-A",
					ComponentID:         "comp-X",
					RealizationLevel:    "Full",
					Origin:              "Inherited",
					SourceRealizationID: "real-direct-B",
					LinkedAt:            time.Now(),
				},
			},
		},
	}
	mockCapRM := &mockCapabilityReadModelForProjector{
		capabilities: map[string]*readmodels.CapabilityDTO{
			"cap-C": {ID: "cap-C", Name: "C", Level: "L1", ParentID: ""},
			"cap-A": {ID: "cap-A", Name: "A", Level: "L2", ParentID: "cap-C"},
			"cap-B": {ID: "cap-B", Name: "B", Level: "L3", ParentID: "cap-A"},
		},
	}

	projector := newTestableRealizationProjector(mockRealRM, mockCapRM)

	event := events.CapabilityParentChanged{
		CapabilityID: "cap-A",
		OldParentID:  "",
		NewParentID:  "cap-C",
		OldLevel:     "L1",
		NewLevel:     "L2",
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "CapabilityParentChanged", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.insertedInheritedRealizations, 1, "Should create inherited realization to new parent C")
	assert.Equal(t, "cap-C", mockRealRM.insertedInheritedRealizations[0].CapabilityID)
	assert.Equal(t, "comp-X", mockRealRM.insertedInheritedRealizations[0].ComponentID)
	assert.Equal(t, "Inherited", mockRealRM.insertedInheritedRealizations[0].Origin)
	assert.Equal(t, "real-direct-B", mockRealRM.insertedInheritedRealizations[0].SourceRealizationID,
		"Should reference the original direct realization")
}

func TestRealizationProjector_HandleCapabilityParentChanged_PropagatesUpEntireNewAncestry(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{
		realizationsByCapability: map[string][]readmodels.RealizationDTO{
			"cap-A": {
				{
					ID:                  "inherited-to-A",
					CapabilityID:        "cap-A",
					ComponentID:         "comp-X",
					RealizationLevel:    "Full",
					Origin:              "Inherited",
					SourceRealizationID: "real-direct-B",
					LinkedAt:            time.Now(),
				},
			},
		},
	}
	mockCapRM := &mockCapabilityReadModelForProjector{
		capabilities: map[string]*readmodels.CapabilityDTO{
			"cap-D": {ID: "cap-D", Name: "D", Level: "L1", ParentID: ""},
			"cap-C": {ID: "cap-C", Name: "C", Level: "L2", ParentID: "cap-D"},
			"cap-A": {ID: "cap-A", Name: "A", Level: "L3", ParentID: "cap-C"},
			"cap-B": {ID: "cap-B", Name: "B", Level: "L4", ParentID: "cap-A"},
		},
	}

	projector := newTestableRealizationProjector(mockRealRM, mockCapRM)

	event := events.CapabilityParentChanged{
		CapabilityID: "cap-A",
		OldParentID:  "",
		NewParentID:  "cap-C",
		OldLevel:     "L1",
		NewLevel:     "L3",
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "CapabilityParentChanged", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.insertedInheritedRealizations, 2, "Should create inherited realizations to C and D")

	assert.Equal(t, "cap-C", mockRealRM.insertedInheritedRealizations[0].CapabilityID)
	assert.Equal(t, "cap-D", mockRealRM.insertedInheritedRealizations[1].CapabilityID)

	for _, inherited := range mockRealRM.insertedInheritedRealizations {
		assert.Equal(t, "comp-X", inherited.ComponentID)
		assert.Equal(t, "real-direct-B", inherited.SourceRealizationID)
	}
}

func TestRealizationProjector_HandleCapabilityParentChanged_NoOpWhenNoNewParent(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{}
	mockCapRM := &mockCapabilityReadModelForProjector{}

	projector := newTestableRealizationProjector(mockRealRM, mockCapRM)

	event := events.CapabilityParentChanged{
		CapabilityID: "cap-A",
		OldParentID:  "cap-C",
		NewParentID:  "",
		OldLevel:     "L2",
		NewLevel:     "L1",
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "CapabilityParentChanged", eventData)
	require.NoError(t, err)

	assert.Empty(t, mockRealRM.insertedInheritedRealizations, "Should not create any realizations when removing parent")
}

func TestRealizationProjector_HandleCapabilityParentChanged_HandlesDirectRealizations(t *testing.T) {
	mockRealRM := &mockRealizationReadModel{
		realizationsByCapability: map[string][]readmodels.RealizationDTO{
			"cap-A": {
				{
					ID:               "real-direct-A",
					CapabilityID:     "cap-A",
					ComponentID:      "comp-Y",
					RealizationLevel: "Partial",
					Origin:           "Direct",
					LinkedAt:         time.Now(),
				},
			},
		},
	}
	mockCapRM := &mockCapabilityReadModelForProjector{
		capabilities: map[string]*readmodels.CapabilityDTO{
			"cap-C": {ID: "cap-C", Name: "C", Level: "L1", ParentID: ""},
			"cap-A": {ID: "cap-A", Name: "A", Level: "L2", ParentID: "cap-C"},
		},
	}

	projector := newTestableRealizationProjector(mockRealRM, mockCapRM)

	event := events.CapabilityParentChanged{
		CapabilityID: "cap-A",
		OldParentID:  "",
		NewParentID:  "cap-C",
		OldLevel:     "L1",
		NewLevel:     "L2",
	}

	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "CapabilityParentChanged", eventData)
	require.NoError(t, err)

	require.Len(t, mockRealRM.insertedInheritedRealizations, 1)
	assert.Equal(t, "cap-C", mockRealRM.insertedInheritedRealizations[0].CapabilityID)
	assert.Equal(t, "comp-Y", mockRealRM.insertedInheritedRealizations[0].ComponentID)
	assert.Equal(t, "real-direct-A", mockRealRM.insertedInheritedRealizations[0].SourceRealizationID,
		"Should reference the direct realization's own ID")
	assert.Equal(t, "cap-A", mockRealRM.insertedInheritedRealizations[0].SourceCapabilityID,
		"Should set SourceCapabilityID to the capability where the direct realization exists")
	assert.Equal(t, "A", mockRealRM.insertedInheritedRealizations[0].SourceCapabilityName,
		"Should set SourceCapabilityName")
}
