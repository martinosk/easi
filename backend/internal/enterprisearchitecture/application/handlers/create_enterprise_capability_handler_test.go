package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCreateCapabilityRepository struct {
	savedCapabilities []*aggregates.EnterpriseCapability
	saveErr           error
}

func (m *mockCreateCapabilityRepository) Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCapabilities = append(m.savedCapabilities, capability)
	return nil
}

type mockCreateCapabilityReadModel struct {
	nameExists bool
	checkErr   error
}

func (m *mockCreateCapabilityReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.nameExists, nil
}

func TestCreateEnterpriseCapabilityHandler_CreatesCapability(t *testing.T) {
	mockRepo := &mockCreateCapabilityRepository{}
	mockReadModel := &mockCreateCapabilityReadModel{nameExists: false}

	handler := NewCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateEnterpriseCapability{
		Name:        "Payroll Management",
		Description: "Manages employee payroll and compensation",
		Category:    "HR",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedCapabilities, 1)
	capability := mockRepo.savedCapabilities[0]
	assert.Equal(t, "Payroll Management", capability.Name().Value())
	assert.Equal(t, "Manages employee payroll and compensation", capability.Description().Value())
	assert.Equal(t, "HR", capability.Category().Value())
}

func TestCreateEnterpriseCapabilityHandler_ReturnsCreatedID(t *testing.T) {
	mockRepo := &mockCreateCapabilityRepository{}
	mockReadModel := &mockCreateCapabilityReadModel{nameExists: false}

	handler := NewCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateEnterpriseCapability{
		Name:        "Order Processing",
		Description: "",
		Category:    "",
	}

	result, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, result.CreatedID)
	assert.Equal(t, mockRepo.savedCapabilities[0].ID(), result.CreatedID)
}

func TestCreateEnterpriseCapabilityHandler_NameExists_ReturnsError(t *testing.T) {
	mockRepo := &mockCreateCapabilityRepository{}
	mockReadModel := &mockCreateCapabilityReadModel{nameExists: true}

	handler := NewCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateEnterpriseCapability{
		Name:        "Duplicate Name",
		Description: "Should fail",
		Category:    "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrEnterpriseCapabilityNameExists)
	assert.Empty(t, mockRepo.savedCapabilities)
}

func TestCreateEnterpriseCapabilityHandler_InvalidName_ReturnsError(t *testing.T) {
	mockRepo := &mockCreateCapabilityRepository{}
	mockReadModel := &mockCreateCapabilityReadModel{nameExists: false}

	handler := NewCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateEnterpriseCapability{
		Name:        "",
		Description: "Invalid name",
		Category:    "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedCapabilities)
}

func TestCreateEnterpriseCapabilityHandler_HandlesOptionalDescriptionAndCategory(t *testing.T) {
	mockRepo := &mockCreateCapabilityRepository{}
	mockReadModel := &mockCreateCapabilityReadModel{nameExists: false}

	handler := NewCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateEnterpriseCapability{
		Name:        "Minimal Capability",
		Description: "",
		Category:    "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	capability := mockRepo.savedCapabilities[0]
	assert.Equal(t, "Minimal Capability", capability.Name().Value())
	assert.Empty(t, capability.Description().Value())
	assert.Empty(t, capability.Category().Value())
}

func TestCreateEnterpriseCapabilityHandler_ReadModelError_ReturnsError(t *testing.T) {
	mockRepo := &mockCreateCapabilityRepository{}
	mockReadModel := &mockCreateCapabilityReadModel{checkErr: errors.New("database error")}

	handler := NewCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateEnterpriseCapability{
		Name:        "Test Capability",
		Description: "Test",
		Category:    "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedCapabilities)
}

func TestCreateEnterpriseCapabilityHandler_RepositoryError_ReturnsError(t *testing.T) {
	mockRepo := &mockCreateCapabilityRepository{saveErr: errors.New("save error")}
	mockReadModel := &mockCreateCapabilityReadModel{nameExists: false}

	handler := NewCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateEnterpriseCapability{
		Name:        "Test Capability",
		Description: "Test",
		Category:    "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
