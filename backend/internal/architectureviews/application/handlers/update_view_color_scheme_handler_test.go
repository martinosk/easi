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

type mockLayoutRepository struct {
	updateColorSchemeErr    error
	updateColorSchemeCalled bool
	viewID                  string
	colorScheme             string
}

func (m *mockLayoutRepository) UpdateColorScheme(ctx context.Context, viewID, colorScheme string) error {
	m.updateColorSchemeCalled = true
	m.viewID = viewID
	m.colorScheme = colorScheme
	return m.updateColorSchemeErr
}

type layoutRepositoryForColorScheme interface {
	UpdateColorScheme(ctx context.Context, viewID, colorScheme string) error
}

type testableUpdateViewColorSchemeHandler struct {
	layoutRepository layoutRepositoryForColorScheme
}

func newTestableUpdateViewColorSchemeHandler(layoutRepository layoutRepositoryForColorScheme) *testableUpdateViewColorSchemeHandler {
	return &testableUpdateViewColorSchemeHandler{
		layoutRepository: layoutRepository,
	}
}

func (h *testableUpdateViewColorSchemeHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateViewColorScheme)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	_, err := valueobjects.NewColorScheme(command.ColorScheme)
	if err != nil {
		return err
	}

	return h.layoutRepository.UpdateColorScheme(ctx, command.ViewID, command.ColorScheme)
}

func TestUpdateViewColorSchemeHandler_ValidMaturity(t *testing.T) {
	mockRepo := &mockLayoutRepository{}
	handler := newTestableUpdateViewColorSchemeHandler(mockRepo)

	cmd := &commands.UpdateViewColorScheme{
		ViewID:      "view-123",
		ColorScheme: "maturity",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateColorSchemeCalled)
	assert.Equal(t, "view-123", mockRepo.viewID)
	assert.Equal(t, "maturity", mockRepo.colorScheme)
}

func TestUpdateViewColorSchemeHandler_ValidClassic(t *testing.T) {
	mockRepo := &mockLayoutRepository{}
	handler := newTestableUpdateViewColorSchemeHandler(mockRepo)

	cmd := &commands.UpdateViewColorScheme{
		ViewID:      "view-456",
		ColorScheme: "classic",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateColorSchemeCalled)
	assert.Equal(t, "view-456", mockRepo.viewID)
	assert.Equal(t, "classic", mockRepo.colorScheme)
}

func TestUpdateViewColorSchemeHandler_ValidCustom(t *testing.T) {
	mockRepo := &mockLayoutRepository{}
	handler := newTestableUpdateViewColorSchemeHandler(mockRepo)

	cmd := &commands.UpdateViewColorScheme{
		ViewID:      "view-abc",
		ColorScheme: "custom",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateColorSchemeCalled)
	assert.Equal(t, "view-abc", mockRepo.viewID)
	assert.Equal(t, "custom", mockRepo.colorScheme)
}

func TestUpdateViewColorSchemeHandler_InvalidColorScheme(t *testing.T) {
	mockRepo := &mockLayoutRepository{}
	handler := newTestableUpdateViewColorSchemeHandler(mockRepo)

	cmd := &commands.UpdateViewColorScheme{
		ViewID:      "view-123",
		ColorScheme: "invalid-scheme",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, valueobjects.ErrInvalidColorScheme, err)
	assert.False(t, mockRepo.updateColorSchemeCalled)
}

func TestUpdateViewColorSchemeHandler_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("database connection failed")
	mockRepo := &mockLayoutRepository{
		updateColorSchemeErr: repositoryErr,
	}
	handler := newTestableUpdateViewColorSchemeHandler(mockRepo)

	cmd := &commands.UpdateViewColorScheme{
		ViewID:      "view-123",
		ColorScheme: "maturity",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, repositoryErr, err)
	assert.True(t, mockRepo.updateColorSchemeCalled)
}

func TestUpdateViewColorSchemeHandler_InvalidCommand(t *testing.T) {
	mockRepo := &mockLayoutRepository{}
	handler := newTestableUpdateViewColorSchemeHandler(mockRepo)

	invalidCmd := &commands.CreateView{
		Name: "Test",
	}

	err := handler.Handle(context.Background(), invalidCmd)

	assert.Error(t, err)
	assert.Equal(t, cqrs.ErrInvalidCommand, err)
	assert.False(t, mockRepo.updateColorSchemeCalled)
}
