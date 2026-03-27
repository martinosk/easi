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
	ref                     repositories.ElementRef
}

func (m *mockElementColorClearer) ClearElementColor(ctx context.Context, ref repositories.ElementRef) error {
	m.clearElementColorCalled = true
	m.ref = ref
	return m.clearElementColorErr
}

func TestClearElementColorHandler_ValidElementTypes(t *testing.T) {
	tests := []struct {
		name        string
		viewID      string
		elementID   string
		elementType string
		expectedRef repositories.ElementRef
	}{
		{"component", "view-123", "comp-456", "component", repositories.ElementRef{ViewID: "view-123", ElementID: "comp-456", ElementType: repositories.ElementTypeComponent}},
		{"capability", "view-789", "cap-101", "capability", repositories.ElementRef{ViewID: "view-789", ElementID: "cap-101", ElementType: repositories.ElementTypeCapability}},
		{"empty view ID", "", "comp-456", "component", repositories.ElementRef{ViewID: "", ElementID: "comp-456", ElementType: repositories.ElementTypeComponent}},
		{"empty element ID", "view-123", "", "component", repositories.ElementRef{ViewID: "view-123", ElementID: "", ElementType: repositories.ElementTypeComponent}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockElementColorClearer{}
			handler := NewClearElementColorHandler(mockRepo)

			cmd := &commands.ClearElementColor{
				ViewID:      tt.viewID,
				ElementID:   tt.elementID,
				ElementType: tt.elementType,
			}

			_, err := handler.Handle(context.Background(), cmd)

			assert.NoError(t, err)
			assert.True(t, mockRepo.clearElementColorCalled)
			assert.Equal(t, tt.expectedRef, mockRepo.ref)
		})
	}
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
