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

func TestUpdateElementColorHandler_ValidInputs(t *testing.T) {
	tests := []struct {
		name             string
		viewID           string
		elementID        string
		elementType      string
		color            string
		expectedRepoType repositories.ElementType
	}{
		{"component", "view-123", "comp-456", "component", "#FF5733", repositories.ElementTypeComponent},
		{"capability", "view-789", "cap-101", "capability", "#00FF00", repositories.ElementTypeCapability},
		{"lowercase color", "view-123", "comp-456", "component", "#ffffff", repositories.ElementTypeComponent},
		{"mixed case color", "view-123", "comp-456", "component", "#FfA5b3", repositories.ElementTypeComponent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockElementColorUpdater{}
			handler := NewUpdateElementColorHandler(mockRepo)

			cmd := &commands.UpdateElementColor{
				ViewID:      tt.viewID,
				ElementID:   tt.elementID,
				ElementType: tt.elementType,
				Color:       tt.color,
			}

			_, err := handler.Handle(context.Background(), cmd)

			assert.NoError(t, err)
			assert.True(t, mockRepo.updateElementColorCalled)
			assert.Equal(t, tt.viewID, mockRepo.viewID)
			assert.Equal(t, tt.elementID, mockRepo.elementID)
			assert.Equal(t, tt.expectedRepoType, mockRepo.elementType)
			assert.Equal(t, tt.color, mockRepo.color)
		})
	}
}

func TestUpdateElementColorHandler_InvalidColors(t *testing.T) {
	tests := []struct {
		name  string
		color string
	}{
		{"not hex format", "invalid"},
		{"missing hash prefix", "FF5733"},
		{"too short", "#FFF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockElementColorUpdater{}
			handler := NewUpdateElementColorHandler(mockRepo)

			cmd := &commands.UpdateElementColor{
				ViewID:      "view-123",
				ElementID:   "comp-456",
				ElementType: "component",
				Color:       tt.color,
			}

			_, err := handler.Handle(context.Background(), cmd)

			assert.Error(t, err)
			assert.Equal(t, valueobjects.ErrInvalidHexColor, err)
			assert.False(t, mockRepo.updateElementColorCalled)
		})
	}
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
