package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
)

type mockElementColorClearer struct {
	clearElementColorErr    error
	clearElementColorCalled bool
	viewID                  string
	elementID               string
	elementType             repositories.ElementType
}

func (m *mockElementColorClearer) ClearElementColor(ctx context.Context, viewID, elementID string, elementType repositories.ElementType) error {
	m.clearElementColorCalled = true
	m.viewID = viewID
	m.elementID = elementID
	m.elementType = elementType
	return m.clearElementColorErr
}

func TestClearElementColorHandler_ValidComponent(t *testing.T) {
	mockRepo := &mockElementColorClearer{}
	handler := NewClearElementColorHandler(mockRepo)

	cmd := &commands.ClearElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.clearElementColorCalled)
	assert.Equal(t, "view-123", mockRepo.viewID)
	assert.Equal(t, "comp-456", mockRepo.elementID)
	assert.Equal(t, repositories.ElementTypeComponent, mockRepo.elementType)
}

func TestClearElementColorHandler_ValidCapability(t *testing.T) {
	mockRepo := &mockElementColorClearer{}
	handler := NewClearElementColorHandler(mockRepo)

	cmd := &commands.ClearElementColor{
		ViewID:      "view-789",
		ElementID:   "cap-101",
		ElementType: "capability",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.clearElementColorCalled)
	assert.Equal(t, "view-789", mockRepo.viewID)
	assert.Equal(t, "cap-101", mockRepo.elementID)
	assert.Equal(t, repositories.ElementTypeCapability, mockRepo.elementType)
}

func TestClearElementColorHandler_InvalidElementType(t *testing.T) {
	mockRepo := &mockElementColorClearer{}
	handler := NewClearElementColorHandler(mockRepo)

	cmd := &commands.ClearElementColor{
		ViewID:      "view-123",
		ElementID:   "elem-456",
		ElementType: "invalid-type",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid element type")
	assert.False(t, mockRepo.clearElementColorCalled)
}

func TestClearElementColorHandler_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("database connection failed")
	mockRepo := &mockElementColorClearer{
		clearElementColorErr: repositoryErr,
	}
	handler := NewClearElementColorHandler(mockRepo)

	cmd := &commands.ClearElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, repositoryErr, err)
	assert.True(t, mockRepo.clearElementColorCalled)
}

func TestClearElementColorHandler_InvalidCommand(t *testing.T) {
	mockRepo := &mockElementColorClearer{}
	handler := NewClearElementColorHandler(mockRepo)

	invalidCmd := &commands.CreateView{
		Name: "Test",
	}

	_, err := handler.Handle(context.Background(), invalidCmd)

	assert.Error(t, err)
	assert.Equal(t, cqrs.ErrInvalidCommand, err)
	assert.False(t, mockRepo.clearElementColorCalled)
}

func TestClearElementColorHandler_EmptyViewID(t *testing.T) {
	mockRepo := &mockElementColorClearer{}
	handler := NewClearElementColorHandler(mockRepo)

	cmd := &commands.ClearElementColor{
		ViewID:      "",
		ElementID:   "comp-456",
		ElementType: "component",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.clearElementColorCalled)
	assert.Equal(t, "", mockRepo.viewID)
}

func TestClearElementColorHandler_EmptyElementID(t *testing.T) {
	mockRepo := &mockElementColorClearer{}
	handler := NewClearElementColorHandler(mockRepo)

	cmd := &commands.ClearElementColor{
		ViewID:      "view-123",
		ElementID:   "",
		ElementType: "component",
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.clearElementColorCalled)
	assert.Equal(t, "", mockRepo.elementID)
}
