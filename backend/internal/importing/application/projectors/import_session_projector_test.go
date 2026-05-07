package projectors

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"easi/backend/internal/importing/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockImportSessionReadModel struct {
	insertedSessions  []readmodels.ImportSessionDTO
	statusUpdates     []statusUpdate
	progressUpdates   []progressUpdate
	completedCalls    []completedCall
	failedCalls       []failedCall
	cancelledIDs      []string
	insertErr         error
	updateStatusErr   error
	updateProgressErr error
	markCompletedErr  error
	markFailedErr     error
	markCancelledErr  error
}

type statusUpdate struct {
	ID     string
	Status string
}

type progressUpdate struct {
	ID       string
	Progress readmodels.ProgressDTO
}

type completedCall struct {
	ID          string
	Result      readmodels.ResultDTO
	CompletedAt time.Time
}

type failedCall struct {
	ID       string
	FailedAt time.Time
}

func (m *mockImportSessionReadModel) Insert(ctx context.Context, dto readmodels.ImportSessionDTO) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.insertedSessions = append(m.insertedSessions, dto)
	return nil
}

func (m *mockImportSessionReadModel) UpdateStatus(ctx context.Context, id, status string) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	m.statusUpdates = append(m.statusUpdates, statusUpdate{ID: id, Status: status})
	return nil
}

func (m *mockImportSessionReadModel) UpdateProgress(ctx context.Context, id string, progress readmodels.ProgressDTO) error {
	if m.updateProgressErr != nil {
		return m.updateProgressErr
	}
	m.progressUpdates = append(m.progressUpdates, progressUpdate{ID: id, Progress: progress})
	return nil
}

func (m *mockImportSessionReadModel) MarkCompleted(ctx context.Context, id string, result readmodels.ResultDTO, completedAt time.Time) error {
	if m.markCompletedErr != nil {
		return m.markCompletedErr
	}
	m.completedCalls = append(m.completedCalls, completedCall{ID: id, Result: result, CompletedAt: completedAt})
	return nil
}

func (m *mockImportSessionReadModel) MarkFailed(ctx context.Context, id string, failedAt time.Time) error {
	if m.markFailedErr != nil {
		return m.markFailedErr
	}
	m.failedCalls = append(m.failedCalls, failedCall{ID: id, FailedAt: failedAt})
	return nil
}

func (m *mockImportSessionReadModel) MarkCancelled(ctx context.Context, id string) error {
	if m.markCancelledErr != nil {
		return m.markCancelledErr
	}
	m.cancelledIDs = append(m.cancelledIDs, id)
	return nil
}

func TestImportSessionProjector_HandleImportCompleted_WithNoErrors_ReturnsEmptySlice(t *testing.T) {
	mockRM := &mockImportSessionReadModel{}
	projector := NewImportSessionProjector(mockRM)

	completedAt := time.Now()
	eventData, err := json.Marshal(map[string]interface{}{
		"id":                        "import-123",
		"capabilitiesCreated":       10,
		"componentsCreated":         5,
		"valueStreamsCreated":       2,
		"realizationsCreated":       3,
		"componentRelationsCreated": 2,
		"capabilityMappings":        4,
		"domainAssignments":         1,
		"errors":                    nil,
		"completedAt":               completedAt,
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ImportCompleted", eventData)
	require.NoError(t, err)

	require.Len(t, mockRM.completedCalls, 1)
	result := mockRM.completedCalls[0].Result

	assert.Equal(t, 10, result.CapabilitiesCreated)
	assert.Equal(t, 5, result.ComponentsCreated)
	assert.Equal(t, 2, result.ValueStreamsCreated)
	assert.Equal(t, 3, result.RealizationsCreated)
	assert.Equal(t, 2, result.ComponentRelationsCreated)
	assert.Equal(t, 4, result.CapabilityMappings)
	assert.Equal(t, 1, result.DomainAssignments)

	assert.NotNil(t, result.Errors, "Errors should never be nil")
	assert.Empty(t, result.Errors, "Errors should be empty slice when no errors occurred")

	resultJSON, err := json.Marshal(result)
	require.NoError(t, err)
	assert.Contains(t, string(resultJSON), `"errors":[]`, "JSON should contain empty array, not null")
}

func TestImportSessionProjector_HandleImportCompleted_WithErrors_ReturnsPopulatedSlice(t *testing.T) {
	mockRM := &mockImportSessionReadModel{}
	projector := NewImportSessionProjector(mockRM)

	completedAt := time.Now()
	eventData, err := json.Marshal(map[string]interface{}{
		"id":                        "import-123",
		"capabilitiesCreated":       8,
		"componentsCreated":         4,
		"realizationsCreated":       2,
		"componentRelationsCreated": 1,
		"domainAssignments":         0,
		"errors": []map[string]interface{}{
			{
				"sourceElement": "Application",
				"sourceName":    "Legacy System",
				"error":         "Duplicate name",
				"action":        "skipped",
			},
			{
				"sourceElement": "Capability",
				"sourceName":    "Payment Processing",
				"error":         "Invalid parent reference",
				"action":        "created_at_root",
			},
		},
		"completedAt": completedAt,
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ImportCompleted", eventData)
	require.NoError(t, err)

	require.Len(t, mockRM.completedCalls, 1)
	result := mockRM.completedCalls[0].Result

	require.Len(t, result.Errors, 2)
	assert.Equal(t, "Application", result.Errors[0].SourceElement)
	assert.Equal(t, "Legacy System", result.Errors[0].SourceName)
	assert.Equal(t, "Duplicate name", result.Errors[0].Error)
	assert.Equal(t, "skipped", result.Errors[0].Action)
	assert.Equal(t, "Capability", result.Errors[1].SourceElement)
	assert.Equal(t, "Payment Processing", result.Errors[1].SourceName)
}

func TestImportSessionProjector_HandleImportSessionCreated(t *testing.T) {
	mockRM := &mockImportSessionReadModel{}
	projector := NewImportSessionProjector(mockRM)

	createdAt := time.Now()
	eventData, err := json.Marshal(map[string]interface{}{
		"id":               "import-456",
		"sourceFormat":     "archimate-openexchange",
		"businessDomainId": "domain-1",
		"preview": map[string]interface{}{
			"supported": map[string]interface{}{
				"capabilities":                    float64(10),
				"components":                      float64(5),
				"valueStreams":                    float64(1),
				"parentChildRelationships":        float64(8),
				"realizations":                    float64(3),
				"componentRelationships":          float64(2),
				"capabilityToValueStreamMappings": float64(2),
			},
			"unsupported": map[string]interface{}{
				"elements":      map[string]interface{}{"Location": float64(2)},
				"relationships": map[string]interface{}{"Influence": float64(1)},
			},
		},
		"createdAt": createdAt,
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ImportSessionCreated", eventData)
	require.NoError(t, err)

	require.Len(t, mockRM.insertedSessions, 1)
	session := mockRM.insertedSessions[0]

	assert.Equal(t, "import-456", session.ID)
	assert.Equal(t, "archimate-openexchange", session.SourceFormat)
	assert.Equal(t, "domain-1", session.BusinessDomainID)
	assert.Equal(t, "pending", session.Status)
	require.NotNil(t, session.Preview)
	assert.Equal(t, 10, session.Preview.Supported.Capabilities)
	assert.Equal(t, 5, session.Preview.Supported.Components)
	assert.Equal(t, 1, session.Preview.Supported.ValueStreams)
	assert.Equal(t, 2, session.Preview.Supported.CapabilityToValueStreamMappings)
}

func TestImportSessionProjector_HandleImportStarted(t *testing.T) {
	mockRM := &mockImportSessionReadModel{}
	projector := NewImportSessionProjector(mockRM)

	eventData, err := json.Marshal(map[string]interface{}{
		"id":         "import-789",
		"totalItems": 15,
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ImportStarted", eventData)
	require.NoError(t, err)

	require.Len(t, mockRM.statusUpdates, 1)
	assert.Equal(t, "import-789", mockRM.statusUpdates[0].ID)
	assert.Equal(t, "importing", mockRM.statusUpdates[0].Status)

	require.Len(t, mockRM.progressUpdates, 1)
	assert.Equal(t, "import-789", mockRM.progressUpdates[0].ID)
	assert.Equal(t, "creating_components", mockRM.progressUpdates[0].Progress.Phase)
	assert.Equal(t, 15, mockRM.progressUpdates[0].Progress.TotalItems)
	assert.Equal(t, 0, mockRM.progressUpdates[0].Progress.CompletedItems)
}

func TestImportSessionProjector_HandleImportProgressUpdated(t *testing.T) {
	mockRM := &mockImportSessionReadModel{}
	projector := NewImportSessionProjector(mockRM)

	eventData, err := json.Marshal(map[string]interface{}{
		"id":             "import-789",
		"phase":          "creating_realizations",
		"totalItems":     15,
		"completedItems": 10,
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ImportProgressUpdated", eventData)
	require.NoError(t, err)

	require.Len(t, mockRM.progressUpdates, 1)
	assert.Equal(t, "import-789", mockRM.progressUpdates[0].ID)
	assert.Equal(t, "creating_realizations", mockRM.progressUpdates[0].Progress.Phase)
	assert.Equal(t, 15, mockRM.progressUpdates[0].Progress.TotalItems)
	assert.Equal(t, 10, mockRM.progressUpdates[0].Progress.CompletedItems)
}

func TestImportSessionProjector_HandleTerminalEvents(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		eventData map[string]interface{}
		assertFn  func(t *testing.T, mockRM *mockImportSessionReadModel)
	}{
		{
			name:      "ImportFailed marks session as failed",
			eventType: "ImportFailed",
			eventData: map[string]interface{}{
				"id":       "import-failed",
				"failedAt": time.Now(),
			},
			assertFn: func(t *testing.T, mockRM *mockImportSessionReadModel) {
				require.Len(t, mockRM.failedCalls, 1)
				assert.Equal(t, "import-failed", mockRM.failedCalls[0].ID)
			},
		},
		{
			name:      "ImportSessionCancelled marks session as cancelled",
			eventType: "ImportSessionCancelled",
			eventData: map[string]interface{}{
				"id": "import-cancelled",
			},
			assertFn: func(t *testing.T, mockRM *mockImportSessionReadModel) {
				require.Len(t, mockRM.cancelledIDs, 1)
				assert.Equal(t, "import-cancelled", mockRM.cancelledIDs[0])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRM := &mockImportSessionReadModel{}
			projector := NewImportSessionProjector(mockRM)

			eventData, err := json.Marshal(tt.eventData)
			require.NoError(t, err)

			err = projector.ProjectEvent(context.Background(), tt.eventType, eventData)
			require.NoError(t, err)

			tt.assertFn(t, mockRM)
		})
	}
}

func TestImportSessionProjector_UnknownEventType_NoOp(t *testing.T) {
	mockRM := &mockImportSessionReadModel{}
	projector := NewImportSessionProjector(mockRM)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte(`{}`))
	require.NoError(t, err)

	assert.Empty(t, mockRM.insertedSessions)
	assert.Empty(t, mockRM.statusUpdates)
	assert.Empty(t, mockRM.progressUpdates)
	assert.Empty(t, mockRM.completedCalls)
	assert.Empty(t, mockRM.failedCalls)
	assert.Empty(t, mockRM.cancelledIDs)
}
