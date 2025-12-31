package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
)

type mockColorSchemeUpdater struct {
	updateColorSchemeErr    error
	updateColorSchemeCalled bool
	viewID                  string
	colorScheme             string
}

func (m *mockColorSchemeUpdater) UpdateColorScheme(ctx context.Context, viewID, colorScheme string) error {
	m.updateColorSchemeCalled = true
	m.viewID = viewID
	m.colorScheme = colorScheme
	return m.updateColorSchemeErr
}

func TestUpdateViewColorSchemeHandler_ValidMaturity(t *testing.T) {
	mockRepo := &mockColorSchemeUpdater{}
	handler := NewUpdateViewColorSchemeHandler(mockRepo)

	cmd := &commands.UpdateViewColorScheme{
		ViewID:      "view-123",
		ColorScheme: "maturity",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateColorSchemeCalled)
	assert.Equal(t, "view-123", mockRepo.viewID)
	assert.Equal(t, "maturity", mockRepo.colorScheme)
}

func TestUpdateViewColorSchemeHandler_ValidClassic(t *testing.T) {
	mockRepo := &mockColorSchemeUpdater{}
	handler := NewUpdateViewColorSchemeHandler(mockRepo)

	cmd := &commands.UpdateViewColorScheme{
		ViewID:      "view-456",
		ColorScheme: "classic",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateColorSchemeCalled)
	assert.Equal(t, "view-456", mockRepo.viewID)
	assert.Equal(t, "classic", mockRepo.colorScheme)
}

func TestUpdateViewColorSchemeHandler_ValidCustom(t *testing.T) {
	mockRepo := &mockColorSchemeUpdater{}
	handler := NewUpdateViewColorSchemeHandler(mockRepo)

	cmd := &commands.UpdateViewColorScheme{
		ViewID:      "view-abc",
		ColorScheme: "custom",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateColorSchemeCalled)
	assert.Equal(t, "view-abc", mockRepo.viewID)
	assert.Equal(t, "custom", mockRepo.colorScheme)
}

func TestUpdateViewColorSchemeHandler_InvalidColorScheme(t *testing.T) {
	mockRepo := &mockColorSchemeUpdater{}
	handler := NewUpdateViewColorSchemeHandler(mockRepo)

	cmd := &commands.UpdateViewColorScheme{
		ViewID:      "view-123",
		ColorScheme: "invalid-scheme",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, valueobjects.ErrInvalidColorScheme, err)
	assert.False(t, mockRepo.updateColorSchemeCalled)
}

func TestUpdateViewColorSchemeHandler_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("database connection failed")
	mockRepo := &mockColorSchemeUpdater{
		updateColorSchemeErr: repositoryErr,
	}
	handler := NewUpdateViewColorSchemeHandler(mockRepo)

	cmd := &commands.UpdateViewColorScheme{
		ViewID:      "view-123",
		ColorScheme: "maturity",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, repositoryErr, err)
	assert.True(t, mockRepo.updateColorSchemeCalled)
}

func TestUpdateViewColorSchemeHandler_InvalidCommand(t *testing.T) {
	mockRepo := &mockColorSchemeUpdater{}
	handler := NewUpdateViewColorSchemeHandler(mockRepo)

	invalidCmd := &commands.CreateView{
		Name: "Test",
	}

	_, err := handler.Handle(context.Background(), invalidCmd)

	assert.Error(t, err)
	assert.Equal(t, cqrs.ErrInvalidCommand, err)
	assert.False(t, mockRepo.updateColorSchemeCalled)
}
