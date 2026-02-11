package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/valueobjects"
	"easi/backend/internal/valuestreams/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUpdateValueStreamRepository struct {
	valueStream *aggregates.ValueStream
	savedCount  int
	getByIDErr  error
	saveErr     error
}

func (m *mockUpdateValueStreamRepository) GetByID(ctx context.Context, id string) (*aggregates.ValueStream, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.valueStream, nil
}

func (m *mockUpdateValueStreamRepository) Save(ctx context.Context, vs *aggregates.ValueStream) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCount++
	return nil
}

type mockUpdateValueStreamReadModel struct {
	nameExists bool
	checkErr   error
}

func (m *mockUpdateValueStreamReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.nameExists, nil
}

func createTestValueStream(t *testing.T, name, description string) *aggregates.ValueStream {
	t.Helper()

	vsName, err := valueobjects.NewValueStreamName(name)
	require.NoError(t, err)

	desc := valueobjects.MustNewDescription(description)

	vs, err := aggregates.NewValueStream(vsName, desc)
	require.NoError(t, err)
	vs.MarkChangesAsCommitted()

	return vs
}

func TestUpdateValueStreamHandler_UpdatesValueStream(t *testing.T) {
	vs := createTestValueStream(t, "Original Name", "Original Description")
	vsID := vs.ID()

	mockRepo := &mockUpdateValueStreamRepository{valueStream: vs}
	mockReadModel := &mockUpdateValueStreamReadModel{nameExists: false}

	handler := NewUpdateValueStreamHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateValueStream{
		ID:          vsID,
		Name:        "Updated Name",
		Description: "Updated Description",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, mockRepo.savedCount, "Handler should save value stream once")
	assert.Equal(t, "Updated Name", vs.Name().Value())
	assert.Equal(t, "Updated Description", vs.Description().Value())
}

func TestUpdateValueStreamHandler_NameExistsForOtherStream_ReturnsError(t *testing.T) {
	vs := createTestValueStream(t, "Original Name", "Description")
	vsID := vs.ID()

	mockRepo := &mockUpdateValueStreamRepository{valueStream: vs}
	mockReadModel := &mockUpdateValueStreamReadModel{nameExists: true}

	handler := NewUpdateValueStreamHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateValueStream{
		ID:          vsID,
		Name:        "Duplicate Name",
		Description: "Description",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrValueStreamNameExists)
	assert.Equal(t, 0, mockRepo.savedCount, "Should not save when name exists")
}

func TestUpdateValueStreamHandler_NotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockUpdateValueStreamRepository{
		getByIDErr: repositories.ErrValueStreamNotFound,
	}
	mockReadModel := &mockUpdateValueStreamReadModel{nameExists: false}

	handler := NewUpdateValueStreamHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateValueStream{
		ID:          "non-existent",
		Name:        "Name",
		Description: "Description",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrValueStreamNotFound)
}

func TestUpdateValueStreamHandler_InvalidName_ReturnsError(t *testing.T) {
	vs := createTestValueStream(t, "Original Name", "Description")
	vsID := vs.ID()

	mockRepo := &mockUpdateValueStreamRepository{valueStream: vs}
	mockReadModel := &mockUpdateValueStreamReadModel{nameExists: false}

	handler := NewUpdateValueStreamHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateValueStream{
		ID:          vsID,
		Name:        "",
		Description: "Description",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, 0, mockRepo.savedCount, "Should not save with invalid name")
}

func TestUpdateValueStreamHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockUpdateValueStreamRepository{}
	mockReadModel := &mockUpdateValueStreamReadModel{}

	handler := NewUpdateValueStreamHandler(mockRepo, mockReadModel)

	invalidCmd := &commands.DeleteValueStream{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestUpdateValueStreamHandler_ReadModelError_ReturnsError(t *testing.T) {
	vs := createTestValueStream(t, "Original Name", "Description")
	vsID := vs.ID()

	mockRepo := &mockUpdateValueStreamRepository{valueStream: vs}
	mockReadModel := &mockUpdateValueStreamReadModel{checkErr: errors.New("database error")}

	handler := NewUpdateValueStreamHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateValueStream{
		ID:          vsID,
		Name:        "New Name",
		Description: "Description",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, 0, mockRepo.savedCount)
}
