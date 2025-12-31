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

type mockElementColorUpdater struct {
	updateElementColorErr    error
	updateElementColorCalled bool
	viewID                   string
	elementID                string
	elementType              repositories.ElementType
	color                    string
}

func (m *mockElementColorUpdater) UpdateElementColor(ctx context.Context, viewID, elementID string, elementType repositories.ElementType, color string) error {
	m.updateElementColorCalled = true
	m.viewID = viewID
	m.elementID = elementID
	m.elementType = elementType
	m.color = color
	return m.updateElementColorErr
}

func TestUpdateElementColorHandler_ValidComponent(t *testing.T) {
	mockRepo := &mockElementColorUpdater{}
	handler := NewUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "#FF5733",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateElementColorCalled)
	assert.Equal(t, "view-123", mockRepo.viewID)
	assert.Equal(t, "comp-456", mockRepo.elementID)
	assert.Equal(t, repositories.ElementTypeComponent, mockRepo.elementType)
	assert.Equal(t, "#FF5733", mockRepo.color)
}

func TestUpdateElementColorHandler_ValidCapability(t *testing.T) {
	mockRepo := &mockElementColorUpdater{}
	handler := NewUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-789",
		ElementID:   "cap-101",
		ElementType: "capability",
		Color:       "#00FF00",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateElementColorCalled)
	assert.Equal(t, "view-789", mockRepo.viewID)
	assert.Equal(t, "cap-101", mockRepo.elementID)
	assert.Equal(t, repositories.ElementTypeCapability, mockRepo.elementType)
	assert.Equal(t, "#00FF00", mockRepo.color)
}

func TestUpdateElementColorHandler_InvalidHexColor(t *testing.T) {
	mockRepo := &mockElementColorUpdater{}
	handler := NewUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "invalid",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, valueobjects.ErrInvalidHexColor, err)
	assert.False(t, mockRepo.updateElementColorCalled)
}

func TestUpdateElementColorHandler_InvalidHexColorMissingHash(t *testing.T) {
	mockRepo := &mockElementColorUpdater{}
	handler := NewUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "FF5733",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, valueobjects.ErrInvalidHexColor, err)
	assert.False(t, mockRepo.updateElementColorCalled)
}

func TestUpdateElementColorHandler_InvalidHexColorTooShort(t *testing.T) {
	mockRepo := &mockElementColorUpdater{}
	handler := NewUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "#FFF",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, valueobjects.ErrInvalidHexColor, err)
	assert.False(t, mockRepo.updateElementColorCalled)
}

func TestUpdateElementColorHandler_InvalidElementType(t *testing.T) {
	mockRepo := &mockElementColorUpdater{}
	handler := NewUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "elem-456",
		ElementType: "invalid-type",
		Color:       "#FF5733",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid element type")
	assert.False(t, mockRepo.updateElementColorCalled)
}

func TestUpdateElementColorHandler_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("database connection failed")
	mockRepo := &mockElementColorUpdater{
		updateElementColorErr: repositoryErr,
	}
	handler := NewUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "#FF5733",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, repositoryErr, err)
	assert.True(t, mockRepo.updateElementColorCalled)
}

func TestUpdateElementColorHandler_InvalidCommand(t *testing.T) {
	mockRepo := &mockElementColorUpdater{}
	handler := NewUpdateElementColorHandler(mockRepo)

	invalidCmd := &commands.CreateView{
		Name: "Test",
	}

	_, err := handler.Handle(context.Background(), invalidCmd)

	assert.Error(t, err)
	assert.Equal(t, cqrs.ErrInvalidCommand, err)
	assert.False(t, mockRepo.updateElementColorCalled)
}

func TestUpdateElementColorHandler_ValidLowercaseColor(t *testing.T) {
	mockRepo := &mockElementColorUpdater{}
	handler := NewUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "#ffffff",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateElementColorCalled)
	assert.Equal(t, "#ffffff", mockRepo.color)
}

func TestUpdateElementColorHandler_ValidMixedCaseColor(t *testing.T) {
	mockRepo := &mockElementColorUpdater{}
	handler := NewUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "#FfA5b3",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateElementColorCalled)
	assert.Equal(t, "#FfA5b3", mockRepo.color)
}
