package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
)

type mockColorSchemeUpdater struct {
	updatePreferenceErr    error
	updatePreferenceCalled bool
	viewID                 string
	key                    repositories.PreferenceKey
	value                  string
}

func (m *mockColorSchemeUpdater) UpdatePreference(ctx context.Context, viewID string, key repositories.PreferenceKey, value string) error {
	m.updatePreferenceCalled = true
	m.viewID = viewID
	m.key = key
	m.value = value
	return m.updatePreferenceErr
}

func TestUpdateViewColorSchemeHandler_ValidSchemes(t *testing.T) {
	tests := []struct {
		name        string
		viewID      string
		colorScheme string
	}{
		{"maturity", "view-123", "maturity"},
		{"classic", "view-456", "classic"},
		{"custom", "view-abc", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockColorSchemeUpdater{}
			handler := NewUpdateViewColorSchemeHandler(mockRepo)

			cmd := &commands.UpdateViewColorScheme{
				ViewID:      tt.viewID,
				ColorScheme: tt.colorScheme,
			}

			_, err := handler.Handle(context.Background(), cmd)

			assert.NoError(t, err)
			assert.True(t, mockRepo.updatePreferenceCalled)
			assert.Equal(t, tt.viewID, mockRepo.viewID)
			assert.Equal(t, repositories.PreferenceKeyColorScheme, mockRepo.key)
			assert.Equal(t, tt.colorScheme, mockRepo.value)
		})
	}
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
	assert.False(t, mockRepo.updatePreferenceCalled)
}

func TestUpdateViewColorSchemeHandler_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("database connection failed")
	mockRepo := &mockColorSchemeUpdater{
		updatePreferenceErr: repositoryErr,
	}
	handler := NewUpdateViewColorSchemeHandler(mockRepo)

	cmd := &commands.UpdateViewColorScheme{
		ViewID:      "view-123",
		ColorScheme: "maturity",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, repositoryErr, err)
	assert.True(t, mockRepo.updatePreferenceCalled)
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
	assert.False(t, mockRepo.updatePreferenceCalled)
}
