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

type mockLayoutRepositoryForColor struct {
	updateElementColorErr    error
	updateElementColorCalled bool
	viewID                   string
	elementID                string
	elementType              repositories.ElementType
	color                    string
}

func (m *mockLayoutRepositoryForColor) UpdateElementColor(ctx context.Context, viewID, elementID string, elementType repositories.ElementType, color string) error {
	m.updateElementColorCalled = true
	m.viewID = viewID
	m.elementID = elementID
	m.elementType = elementType
	m.color = color
	return m.updateElementColorErr
}

type layoutRepositoryForElementColor interface {
	UpdateElementColor(ctx context.Context, viewID, elementID string, elementType repositories.ElementType, color string) error
}

type testableUpdateElementColorHandler struct {
	layoutRepository layoutRepositoryForElementColor
}

func newTestableUpdateElementColorHandler(layoutRepository layoutRepositoryForElementColor) *testableUpdateElementColorHandler {
	return &testableUpdateElementColorHandler{
		layoutRepository: layoutRepository,
	}
}

func (h *testableUpdateElementColorHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateElementColor)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	_, err := valueobjects.NewHexColor(command.Color)
	if err != nil {
		return err
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

	return h.layoutRepository.UpdateElementColor(ctx, command.ViewID, command.ElementID, elementType, command.Color)
}

func TestUpdateElementColorHandler_ValidComponent(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForColor{}
	handler := newTestableUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "#FF5733",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateElementColorCalled)
	assert.Equal(t, "view-123", mockRepo.viewID)
	assert.Equal(t, "comp-456", mockRepo.elementID)
	assert.Equal(t, repositories.ElementTypeComponent, mockRepo.elementType)
	assert.Equal(t, "#FF5733", mockRepo.color)
}

func TestUpdateElementColorHandler_ValidCapability(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForColor{}
	handler := newTestableUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-789",
		ElementID:   "cap-101",
		ElementType: "capability",
		Color:       "#00FF00",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateElementColorCalled)
	assert.Equal(t, "view-789", mockRepo.viewID)
	assert.Equal(t, "cap-101", mockRepo.elementID)
	assert.Equal(t, repositories.ElementTypeCapability, mockRepo.elementType)
	assert.Equal(t, "#00FF00", mockRepo.color)
}

func TestUpdateElementColorHandler_InvalidHexColor(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForColor{}
	handler := newTestableUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "invalid",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, valueobjects.ErrInvalidHexColor, err)
	assert.False(t, mockRepo.updateElementColorCalled)
}

func TestUpdateElementColorHandler_InvalidHexColorMissingHash(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForColor{}
	handler := newTestableUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "FF5733",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, valueobjects.ErrInvalidHexColor, err)
	assert.False(t, mockRepo.updateElementColorCalled)
}

func TestUpdateElementColorHandler_InvalidHexColorTooShort(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForColor{}
	handler := newTestableUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "#FFF",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, valueobjects.ErrInvalidHexColor, err)
	assert.False(t, mockRepo.updateElementColorCalled)
}

func TestUpdateElementColorHandler_InvalidElementType(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForColor{}
	handler := newTestableUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "elem-456",
		ElementType: "invalid-type",
		Color:       "#FF5733",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid element type")
	assert.False(t, mockRepo.updateElementColorCalled)
}

func TestUpdateElementColorHandler_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("database connection failed")
	mockRepo := &mockLayoutRepositoryForColor{
		updateElementColorErr: repositoryErr,
	}
	handler := newTestableUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "#FF5733",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
	assert.Equal(t, repositoryErr, err)
	assert.True(t, mockRepo.updateElementColorCalled)
}

func TestUpdateElementColorHandler_InvalidCommand(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForColor{}
	handler := newTestableUpdateElementColorHandler(mockRepo)

	invalidCmd := &commands.CreateView{
		Name: "Test",
	}

	err := handler.Handle(context.Background(), invalidCmd)

	assert.Error(t, err)
	assert.Equal(t, cqrs.ErrInvalidCommand, err)
	assert.False(t, mockRepo.updateElementColorCalled)
}

func TestUpdateElementColorHandler_ValidLowercaseColor(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForColor{}
	handler := newTestableUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "#ffffff",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateElementColorCalled)
	assert.Equal(t, "#ffffff", mockRepo.color)
}

func TestUpdateElementColorHandler_ValidMixedCaseColor(t *testing.T) {
	mockRepo := &mockLayoutRepositoryForColor{}
	handler := newTestableUpdateElementColorHandler(mockRepo)

	cmd := &commands.UpdateElementColor{
		ViewID:      "view-123",
		ElementID:   "comp-456",
		ElementType: "component",
		Color:       "#FfA5b3",
	}

	err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.True(t, mockRepo.updateElementColorCalled)
	assert.Equal(t, "#FfA5b3", mockRepo.color)
}
