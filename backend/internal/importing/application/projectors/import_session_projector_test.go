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
	insertedSessions []readmodels.ImportSessionDTO
	statusUpdates    []statusUpdate
	progressUpdates  []progressUpdate
	completedCalls   []completedCall
	failedCalls      []failedCall
	cancelledIDs     []string
	insertErr        error
	updateStatusErr  error
	updateProgressErr error
	markCompletedErr error
	markFailedErr    error
	markCancelledErr error
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

type importSessionReadModelInterface interface {
	Insert(ctx context.Context, dto readmodels.ImportSessionDTO) error
	UpdateStatus(ctx context.Context, id, status string) error
	UpdateProgress(ctx context.Context, id string, progress readmodels.ProgressDTO) error
	MarkCompleted(ctx context.Context, id string, result readmodels.ResultDTO, completedAt time.Time) error
	MarkFailed(ctx context.Context, id string, failedAt time.Time) error
	MarkCancelled(ctx context.Context, id string) error
}

type testableImportSessionProjector struct {
	readModel importSessionReadModelInterface
}

func newTestableImportSessionProjector(readModel importSessionReadModelInterface) *testableImportSessionProjector {
	return &testableImportSessionProjector{readModel: readModel}
}

func (p *testableImportSessionProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case "ImportSessionCreated":
		return p.handleImportSessionCreated(ctx, eventData)
	case "ImportStarted":
		return p.handleImportStarted(ctx, eventData)
	case "ImportProgressUpdated":
		return p.handleImportProgressUpdated(ctx, eventData)
	case "ImportCompleted":
		return p.handleImportCompleted(ctx, eventData)
	case "ImportFailed":
		return p.handleImportFailed(ctx, eventData)
	case "ImportSessionCancelled":
		return p.handleImportSessionCancelled(ctx, eventData)
	}
	return nil
}

func (p *testableImportSessionProjector) handleImportSessionCreated(ctx context.Context, eventData []byte) error {
	var data struct {
		ID               string                 `json:"id"`
		SourceFormat     string                 `json:"sourceFormat"`
		BusinessDomainID string                 `json:"businessDomainId"`
		Preview          map[string]interface{} `json:"preview"`
		CreatedAt        time.Time              `json:"createdAt"`
	}
	if err := json.Unmarshal(eventData, &data); err != nil {
		return err
	}

	preview := readmodels.PreviewDTO{}
	if supported, ok := data.Preview["supported"].(map[string]interface{}); ok {
		preview.Supported = readmodels.SupportedCountsDTO{
			Capabilities:             getIntFromMap(supported, "capabilities"),
			Components:               getIntFromMap(supported, "components"),
			ParentChildRelationships: getIntFromMap(supported, "parentChildRelationships"),
			Realizations:             getIntFromMap(supported, "realizations"),
			ComponentRelationships:   getIntFromMap(supported, "componentRelationships"),
		}
	}
	if unsupported, ok := data.Preview["unsupported"].(map[string]interface{}); ok {
		preview.Unsupported = readmodels.UnsupportedCountsDTO{
			Elements:      getStringIntMap(unsupported, "elements"),
			Relationships: getStringIntMap(unsupported, "relationships"),
		}
	}

	dto := readmodels.ImportSessionDTO{
		ID:               data.ID,
		SourceFormat:     data.SourceFormat,
		BusinessDomainID: data.BusinessDomainID,
		Status:           "pending",
		Preview:          &preview,
		CreatedAt:        data.CreatedAt,
	}

	return p.readModel.Insert(ctx, dto)
}

func (p *testableImportSessionProjector) handleImportStarted(ctx context.Context, eventData []byte) error {
	var data struct {
		ID         string `json:"id"`
		TotalItems int    `json:"totalItems"`
	}
	if err := json.Unmarshal(eventData, &data); err != nil {
		return err
	}

	if err := p.readModel.UpdateStatus(ctx, data.ID, "importing"); err != nil {
		return err
	}

	progress := readmodels.ProgressDTO{
		Phase:          "creating_components",
		TotalItems:     data.TotalItems,
		CompletedItems: 0,
	}

	return p.readModel.UpdateProgress(ctx, data.ID, progress)
}

func (p *testableImportSessionProjector) handleImportProgressUpdated(ctx context.Context, eventData []byte) error {
	var data struct {
		ID             string `json:"id"`
		Phase          string `json:"phase"`
		TotalItems     int    `json:"totalItems"`
		CompletedItems int    `json:"completedItems"`
	}
	if err := json.Unmarshal(eventData, &data); err != nil {
		return err
	}

	progress := readmodels.ProgressDTO{
		Phase:          data.Phase,
		TotalItems:     data.TotalItems,
		CompletedItems: data.CompletedItems,
	}

	return p.readModel.UpdateProgress(ctx, data.ID, progress)
}

func (p *testableImportSessionProjector) handleImportCompleted(ctx context.Context, eventData []byte) error {
	var data struct {
		ID                        string                   `json:"id"`
		CapabilitiesCreated       int                      `json:"capabilitiesCreated"`
		ComponentsCreated         int                      `json:"componentsCreated"`
		RealizationsCreated       int                      `json:"realizationsCreated"`
		ComponentRelationsCreated int                      `json:"componentRelationsCreated"`
		DomainAssignments         int                      `json:"domainAssignments"`
		Errors                    []map[string]interface{} `json:"errors"`
		CompletedAt               time.Time                `json:"completedAt"`
	}
	if err := json.Unmarshal(eventData, &data); err != nil {
		return err
	}

	errors := make([]readmodels.ImportErrorDTO, 0, len(data.Errors))
	for _, e := range data.Errors {
		errors = append(errors, readmodels.ImportErrorDTO{
			SourceElement: getString(e, "sourceElement"),
			SourceName:    getString(e, "sourceName"),
			Error:         getString(e, "error"),
			Action:        getString(e, "action"),
		})
	}

	result := readmodels.ResultDTO{
		CapabilitiesCreated:       data.CapabilitiesCreated,
		ComponentsCreated:         data.ComponentsCreated,
		RealizationsCreated:       data.RealizationsCreated,
		ComponentRelationsCreated: data.ComponentRelationsCreated,
		DomainAssignments:         data.DomainAssignments,
		Errors:                    errors,
	}

	return p.readModel.MarkCompleted(ctx, data.ID, result, data.CompletedAt)
}

func (p *testableImportSessionProjector) handleImportFailed(ctx context.Context, eventData []byte) error {
	var data struct {
		ID       string    `json:"id"`
		FailedAt time.Time `json:"failedAt"`
	}
	if err := json.Unmarshal(eventData, &data); err != nil {
		return err
	}

	return p.readModel.MarkFailed(ctx, data.ID, data.FailedAt)
}

func (p *testableImportSessionProjector) handleImportSessionCancelled(ctx context.Context, eventData []byte) error {
	var data struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(eventData, &data); err != nil {
		return err
	}

	return p.readModel.MarkCancelled(ctx, data.ID)
}

func TestImportSessionProjector_HandleImportCompleted_WithNoErrors_ReturnsEmptySlice(t *testing.T) {
	mockRM := &mockImportSessionReadModel{}
	projector := newTestableImportSessionProjector(mockRM)

	completedAt := time.Now()
	eventData, err := json.Marshal(map[string]interface{}{
		"id":                        "import-123",
		"capabilitiesCreated":       10,
		"componentsCreated":         5,
		"realizationsCreated":       3,
		"componentRelationsCreated": 2,
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
	assert.Equal(t, 3, result.RealizationsCreated)
	assert.Equal(t, 2, result.ComponentRelationsCreated)
	assert.Equal(t, 1, result.DomainAssignments)

	assert.NotNil(t, result.Errors, "Errors should never be nil")
	assert.Empty(t, result.Errors, "Errors should be empty slice when no errors occurred")

	resultJSON, err := json.Marshal(result)
	require.NoError(t, err)
	assert.Contains(t, string(resultJSON), `"errors":[]`, "JSON should contain empty array, not null")
}

func TestImportSessionProjector_HandleImportCompleted_WithErrors_ReturnsPopulatedSlice(t *testing.T) {
	mockRM := &mockImportSessionReadModel{}
	projector := newTestableImportSessionProjector(mockRM)

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
	projector := newTestableImportSessionProjector(mockRM)

	createdAt := time.Now()
	eventData, err := json.Marshal(map[string]interface{}{
		"id":               "import-456",
		"sourceFormat":     "archimate-openexchange",
		"businessDomainId": "domain-1",
		"preview": map[string]interface{}{
			"supported": map[string]interface{}{
				"capabilities":             float64(10),
				"components":               float64(5),
				"parentChildRelationships": float64(8),
				"realizations":             float64(3),
				"componentRelationships":   float64(2),
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
}

func TestImportSessionProjector_HandleImportStarted(t *testing.T) {
	mockRM := &mockImportSessionReadModel{}
	projector := newTestableImportSessionProjector(mockRM)

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
	projector := newTestableImportSessionProjector(mockRM)

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

func TestImportSessionProjector_HandleImportFailed(t *testing.T) {
	mockRM := &mockImportSessionReadModel{}
	projector := newTestableImportSessionProjector(mockRM)

	failedAt := time.Now()
	eventData, err := json.Marshal(map[string]interface{}{
		"id":       "import-failed",
		"failedAt": failedAt,
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ImportFailed", eventData)
	require.NoError(t, err)

	require.Len(t, mockRM.failedCalls, 1)
	assert.Equal(t, "import-failed", mockRM.failedCalls[0].ID)
}

func TestImportSessionProjector_HandleImportSessionCancelled(t *testing.T) {
	mockRM := &mockImportSessionReadModel{}
	projector := newTestableImportSessionProjector(mockRM)

	eventData, err := json.Marshal(map[string]interface{}{
		"id": "import-cancelled",
	})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ImportSessionCancelled", eventData)
	require.NoError(t, err)

	require.Len(t, mockRM.cancelledIDs, 1)
	assert.Equal(t, "import-cancelled", mockRM.cancelledIDs[0])
}

func TestImportSessionProjector_UnknownEventType_NoOp(t *testing.T) {
	mockRM := &mockImportSessionReadModel{}
	projector := newTestableImportSessionProjector(mockRM)

	err := projector.ProjectEvent(context.Background(), "UnknownEvent", []byte(`{}`))
	require.NoError(t, err)

	assert.Empty(t, mockRM.insertedSessions)
	assert.Empty(t, mockRM.statusUpdates)
	assert.Empty(t, mockRM.progressUpdates)
	assert.Empty(t, mockRM.completedCalls)
	assert.Empty(t, mockRM.failedCalls)
	assert.Empty(t, mockRM.cancelledIDs)
}
