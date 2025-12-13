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

type mockLayoutRepositoryForClearColor struct {
	clearElementColorErr    error
	clearElementColorCalled bool
	viewID                  string
	elementID               string
	elementType             repositories.ElementType
}

func (m *mockLayoutRepositoryForClearColor) ClearElementColor(ctx context.Context, viewID, elementID string, elementType repositories.ElementType) error {
	m.clearElementColorCalled = true
	m.viewID = viewID
	m.elementID = elementID
	m.elementType = elementType
	return m.clearElementColorErr
}

type layoutRepositoryForClearElementColor interface {
	ClearElementColor(ctx context.Context, viewID, elementID string, elementType repositories.ElementType) error
}

type testableClearElementColorHandler struct {
	layoutRepository layoutRepositoryForClearElementColor
}

func newTestableClearElementColorHandler(layoutRepository layoutRepositoryForClearElementColor) *testableClearElementColorHandler {
	return &testableClearElementColorHandler{
		layoutRepository: layoutRepository,
	}
}

func (h *testableClearElementColorHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.ClearElementColor)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	var elementType repositories.ElementType
	switch command.ElementType {
	case "component":
		elementType = repositories.ElementTypeComponent
	case "capability":
		elementType = repositories.ElementTypeCapability
	default:
		return errors.New("invalid element type: must be 'component' or 'capability'")
	}

	return h.layoutRepository.ClearElementColor(ctx, command.ViewID, command.ElementID, elementType)
}

func TestClearElementColorHandler_ValidComponent(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForClearColor{}
	handler := newTestableClearElementColorHandler(mockRepo)

	cmd := &commands.ClearElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.clearElementColorCalled)
	assert.Equal(t, "view-123", mockRepo.viewID)
	assert.Equal(t, "comp-456", mockRepo.elementID)
	assert.Equal(t, repositories.ElementTypeComponent, mockRepo.elementType)
}

func TestClearElementColorHandler_ValidCapability(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForClearColor{}
	handler := newTestableClearElementColorHandler(mockRepo)

	cmd := &commands.ClearElementColor{
		ViewID:      "view-789",
		ElementID:   "cap-101",
		ElementType: "capability",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.clearElementColorCalled)
	assert.Equal(t, "view-789", mockRepo.viewID)
	assert.Equal(t, "cap-101", mockRepo.elementID)
	assert.Equal(t, repositories.ElementTypeCapability, mockRepo.elementType)
}

func TestClearElementColorHandler_InvalidElementType(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForClearColor{}
	handler := newTestableClearElementColorHandler(mockRepo)

	cmd := &commands.ClearElementColor{
		ViewID:      "view-123",
		ElementID:   "elem-456",
		ElementType: "invalid-type",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid element type")
	assert.False(t, mockRepo.clearElementColorCalled)
}

func TestClearElementColorHandler_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("database connection failed")
	mockRepo := &mockLayoutRepositoryForClearColor{
		clearElementColorErr: repositoryErr,
	}
	handler := newTestableClearElementColorHandler(mockRepo)

	cmd := &commands.ClearElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, repositoryErr, err)
	assert.True(t, mockRepo.clearElementColorCalled)
}

func TestClearElementColorHandler_InvalidCommand(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForClearColor{}
	handler := newTestableClearElementColorHandler(mockRepo)

	invalidCmd := &commands.CreateView{
		Name: "Test",
	}

	err := handler.Handle(context.Background(), invalidCmd)

	assert.Error(t, err)
	assert.Equal(t, cqrs.ErrInvalidCommand, err)
	assert.False(t, mockRepo.clearElementColorCalled)
}

func TestClearElementColorHandler_EmptyViewID(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForClearColor{}
	handler := newTestableClearElementColorHandler(mockRepo)

	cmd := &commands.ClearElementColor{
		ViewID:      "",
		ElementID:   "comp-456",
		ElementType: "component",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.clearElementColorCalled)
	assert.Equal(t, "", mockRepo.viewID)
}

func TestClearElementColorHandler_EmptyElementID(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForClearColor{}
	handler := newTestableClearElementColorHandler(mockRepo)

	cmd := &commands.ClearElementColor{
		ViewID:      "view-123",
		ElementID:   "",
		ElementType: "component",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.clearElementColorCalled)
	assert.Equal(t, "", mockRepo.elementID)
}
