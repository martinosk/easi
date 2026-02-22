package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/valuestreams/domain/aggregates"

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

func TestCreateValueStreamHandler_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		repo      *mockCreateValueStreamRepository
		readModel *mockCreateValueStreamReadModel
		cmd       cqrs.Command
		wantErr   error
		notSaved  bool
	}{
		{
			name:      "name exists",
			repo:      &mockCreateValueStreamRepository{},
			readModel: &mockCreateValueStreamReadModel{nameExists: true},
			cmd:       &commands.CreateValueStream{Name: "Duplicate Name", Description: "Should fail"},
			wantErr:   ErrValueStreamNameExists,
			notSaved:  true,
		},
		{
			name:      "invalid name",
			repo:      &mockCreateValueStreamRepository{},
			readModel: &mockCreateValueStreamReadModel{nameExists: false},
			cmd:       &commands.CreateValueStream{Name: "", Description: "Invalid name"},
			notSaved:  true,
		},
		{
			name:      "invalid command",
			repo:      &mockCreateValueStreamRepository{},
			readModel: &mockCreateValueStreamReadModel{},
			cmd:       &commands.DeleteValueStream{},
			wantErr:   cqrs.ErrInvalidCommand,
		},
		{
			name:      "read model error",
			repo:      &mockCreateValueStreamRepository{},
			readModel: &mockCreateValueStreamReadModel{checkErr: errors.New("database error")},
			cmd:       &commands.CreateValueStream{Name: "Test Stream", Description: "Test"},
			notSaved:  true,
		},
		{
			name:      "repository save error",
			repo:      &mockCreateValueStreamRepository{saveErr: errors.New("save failed")},
			readModel: &mockCreateValueStreamReadModel{nameExists: false},
			cmd:       &commands.CreateValueStream{Name: "Test Stream", Description: "Test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewCreateValueStreamHandler(tt.repo, tt.readModel)
			_, err := handler.Handle(context.Background(), tt.cmd)
			assert.Error(t, err)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}
			if tt.notSaved {
				assert.Empty(t, tt.repo.savedStreams)
			}
		})
	}
}
