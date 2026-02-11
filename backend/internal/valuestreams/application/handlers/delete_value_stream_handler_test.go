package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDeleteValueStreamRepository struct {
	valueStream *aggregates.ValueStream
	savedCount  int
	getByIDErr  error
	saveErr     error
}

func (m *mockDeleteValueStreamRepository) GetByID(ctx context.Context, id string) (*aggregates.ValueStream, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.valueStream, nil
}

func (m *mockDeleteValueStreamRepository) Save(ctx context.Context, vs *aggregates.ValueStream) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCount++
	return nil
}

func TestDeleteValueStreamHandler_DeletesValueStream(t *testing.T) {
	vs := createTestValueStream(t, "Test Stream", "Description")
	vsID := vs.ID()

	mockRepo := &mockDeleteValueStreamRepository{valueStream: vs}

	handler := NewDeleteValueStreamHandler(mockRepo)

	cmd := &commands.DeleteValueStream{
		ID: vsID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, mockRepo.savedCount, "Handler should save value stream once")
}

func TestDeleteValueStreamHandler_NotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockDeleteValueStreamRepository{
		getByIDErr: repositories.ErrValueStreamNotFound,
	}

	handler := NewDeleteValueStreamHandler(mockRepo)

	cmd := &commands.DeleteValueStream{
		ID: "non-existent",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrValueStreamNotFound)
}

func TestDeleteValueStreamHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockDeleteValueStreamRepository{}

	handler := NewDeleteValueStreamHandler(mockRepo)

	invalidCmd := &commands.CreateValueStream{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestDeleteValueStreamHandler_RaisesDeletedEvent(t *testing.T) {
	vs := createTestValueStream(t, "Test Stream", "Description")
	vsID := vs.ID()

	mockRepo := &mockDeleteValueStreamRepository{valueStream: vs}

	handler := NewDeleteValueStreamHandler(mockRepo)

	cmd := &commands.DeleteValueStream{
		ID: vsID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	uncommittedEvents := vs.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ValueStreamDeleted", uncommittedEvents[0].EventType())
}
