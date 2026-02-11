package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCreateValueStreamRepository struct {
	savedStreams []*aggregates.ValueStream
	saveErr      error
}

func (m *mockCreateValueStreamRepository) Save(ctx context.Context, vs *aggregates.ValueStream) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedStreams = append(m.savedStreams, vs)
	return nil
}

type mockCreateValueStreamReadModel struct {
	nameExists bool
	checkErr   error
}

func (m *mockCreateValueStreamReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.nameExists, nil
}

func TestCreateValueStreamHandler_CreatesValueStream(t *testing.T) {
	mockRepo := &mockCreateValueStreamRepository{}
	mockReadModel := &mockCreateValueStreamReadModel{nameExists: false}

	handler := NewCreateValueStreamHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateValueStream{
		Name:        "Customer Onboarding",
		Description: "End-to-end customer onboarding process",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedStreams, 1, "Handler should create exactly 1 value stream")

	vs := mockRepo.savedStreams[0]
	assert.Equal(t, "Customer Onboarding", vs.Name().Value())
	assert.Equal(t, "End-to-end customer onboarding process", vs.Description().Value())
}

func TestCreateValueStreamHandler_ReturnsCreatedID(t *testing.T) {
	mockRepo := &mockCreateValueStreamRepository{}
	mockReadModel := &mockCreateValueStreamReadModel{nameExists: false}

	handler := NewCreateValueStreamHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateValueStream{
		Name:        "Order Fulfillment",
		Description: "",
	}

	result, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, result.CreatedID, "Result CreatedID should be set after handling")
	assert.Equal(t, mockRepo.savedStreams[0].ID(), result.CreatedID)
}

func TestCreateValueStreamHandler_NameExists_ReturnsError(t *testing.T) {
	mockRepo := &mockCreateValueStreamRepository{}
	mockReadModel := &mockCreateValueStreamReadModel{nameExists: true}

	handler := NewCreateValueStreamHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateValueStream{
		Name:        "Duplicate Name",
		Description: "Should fail",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrValueStreamNameExists)
	assert.Empty(t, mockRepo.savedStreams, "Should not save value stream when name exists")
}

func TestCreateValueStreamHandler_InvalidName_ReturnsError(t *testing.T) {
	mockRepo := &mockCreateValueStreamRepository{}
	mockReadModel := &mockCreateValueStreamReadModel{nameExists: false}

	handler := NewCreateValueStreamHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateValueStream{
		Name:        "",
		Description: "Invalid name",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedStreams, "Should not save value stream with invalid name")
}

func TestCreateValueStreamHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockCreateValueStreamRepository{}
	mockReadModel := &mockCreateValueStreamReadModel{}

	handler := NewCreateValueStreamHandler(mockRepo, mockReadModel)

	invalidCmd := &commands.DeleteValueStream{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestCreateValueStreamHandler_ReadModelError_ReturnsError(t *testing.T) {
	mockRepo := &mockCreateValueStreamRepository{}
	mockReadModel := &mockCreateValueStreamReadModel{checkErr: errors.New("database error")}

	handler := NewCreateValueStreamHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateValueStream{
		Name:        "Test Stream",
		Description: "Test",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedStreams)
}

func TestCreateValueStreamHandler_RepositorySaveError_ReturnsError(t *testing.T) {
	mockRepo := &mockCreateValueStreamRepository{saveErr: errors.New("save failed")}
	mockReadModel := &mockCreateValueStreamReadModel{nameExists: false}

	handler := NewCreateValueStreamHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateValueStream{
		Name:        "Test Stream",
		Description: "Test",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
