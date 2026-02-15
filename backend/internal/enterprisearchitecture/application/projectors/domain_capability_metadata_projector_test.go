package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMetadataMaturityWriter struct {
	updatedMaturityValues []maturityUpdate
	updateErr             error
}

type maturityUpdate struct {
	CapabilityID  string
	MaturityValue int
}

func (m *mockMetadataMaturityWriter) UpdateMaturityValue(ctx context.Context, capabilityID string, maturityValue int) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedMaturityValues = append(m.updatedMaturityValues, maturityUpdate{CapabilityID: capabilityID, MaturityValue: maturityValue})
	return nil
}

type metadataMaturityWriter interface {
	UpdateMaturityValue(ctx context.Context, capabilityID string, maturityValue int) error
}

type testableMetadataMaturityProjector struct {
	readModel metadataMaturityWriter
}

func newTestableMetadataMaturityProjector(rm metadataMaturityWriter) *testableMetadataMaturityProjector {
	return &testableMetadataMaturityProjector{readModel: rm}
}

func (p *testableMetadataMaturityProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"CapabilityMetadataUpdated": p.handleCapabilityMetadataUpdated,
	}
	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *testableMetadataMaturityProjector) handleCapabilityMetadataUpdated(ctx context.Context, eventData []byte) error {
	var event capabilityMetadataUpdatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}
	return p.readModel.UpdateMaturityValue(ctx, event.ID, event.MaturityValue)
}

func TestMetadataProjector_MetadataUpdated_UpdatesMaturityValue(t *testing.T) {
	tests := []struct {
		name          string
		maturityValue int
	}{
		{"positive value", 3},
		{"zero value", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockMetadataMaturityWriter{}
			projector := newTestableMetadataMaturityProjector(mock)

			capabilityID := uuid.New().String()
			eventData, err := json.Marshal(capabilityMetadataUpdatedEvent{
				ID:            capabilityID,
				MaturityValue: tt.maturityValue,
			})
			require.NoError(t, err)

			err = projector.ProjectEvent(context.Background(), "CapabilityMetadataUpdated", eventData)
			require.NoError(t, err)

			require.Len(t, mock.updatedMaturityValues, 1)
			assert.Equal(t, capabilityID, mock.updatedMaturityValues[0].CapabilityID)
			assert.Equal(t, tt.maturityValue, mock.updatedMaturityValues[0].MaturityValue)
		})
	}
}

func TestMetadataProjector_MetadataUpdated_UnknownEvent_Ignored(t *testing.T) {
	mock := &mockMetadataMaturityWriter{}
	projector := newTestableMetadataMaturityProjector(mock)

	err := projector.ProjectEvent(context.Background(), "SomeOtherEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mock.updatedMaturityValues)
}

func TestMetadataProjector_MetadataUpdated_InvalidJSON_ReturnsError(t *testing.T) {
	mock := &mockMetadataMaturityWriter{}
	projector := newTestableMetadataMaturityProjector(mock)

	err := projector.ProjectEvent(context.Background(), "CapabilityMetadataUpdated", []byte("invalid"))
	assert.Error(t, err)
}

func TestMetadataProjector_MetadataUpdated_ReadModelError_ReturnsError(t *testing.T) {
	mock := &mockMetadataMaturityWriter{updateErr: errors.New("db error")}
	projector := newTestableMetadataMaturityProjector(mock)

	eventData, _ := json.Marshal(capabilityMetadataUpdatedEvent{
		ID:            uuid.New().String(),
		MaturityValue: 5,
	})

	err := projector.ProjectEvent(context.Background(), "CapabilityMetadataUpdated", eventData)
	assert.Error(t, err)
}
