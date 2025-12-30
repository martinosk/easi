package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnterpriseCapabilityRepository struct {
	savedCapabilities []*aggregates.EnterpriseCapability
	saveErr           error
}

func (m *mockEnterpriseCapabilityRepository) Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCapabilities = append(m.savedCapabilities, capability)
	return nil
}

type mockEnterpriseCapabilityReadModel struct {
	nameExists bool
	checkErr   error
}

func (m *mockEnterpriseCapabilityReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.nameExists, nil
}

type enterpriseCapabilityRepository interface {
	Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error
}

type enterpriseCapabilityReadModelForCreate interface {
	NameExists(ctx context.Context, name, excludeID string) (bool, error)
}

type testableCreateEnterpriseCapabilityHandler struct {
	repository enterpriseCapabilityRepository
	readModel  enterpriseCapabilityReadModelForCreate
}

func newTestableCreateEnterpriseCapabilityHandler(
	repository enterpriseCapabilityRepository,
	readModel enterpriseCapabilityReadModelForCreate,
) *testableCreateEnterpriseCapabilityHandler {
	return &testableCreateEnterpriseCapabilityHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *testableCreateEnterpriseCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateEnterpriseCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	exists, err := h.readModel.NameExists(ctx, command.Name, "")
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if exists {
		return cqrs.EmptyResult(), ErrEnterpriseCapabilityNameExists
	}

	name, err := valueobjects.NewEnterpriseCapabilityName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	category, err := valueobjects.NewCategory(command.Category)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	capability, err := aggregates.NewEnterpriseCapability(name, description, category)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(capability.ID()), nil
}

func TestCreateEnterpriseCapabilityHandler_CreatesCapability(t *testing.T) {
	mockRepo := &mockEnterpriseCapabilityRepository{}
	mockReadModel := &mockEnterpriseCapabilityReadModel{nameExists: false}

	handler := newTestableCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

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
	mockRepo := &mockEnterpriseCapabilityRepository{}
	mockReadModel := &mockEnterpriseCapabilityReadModel{nameExists: false}

	handler := newTestableCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

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
	mockRepo := &mockEnterpriseCapabilityRepository{}
	mockReadModel := &mockEnterpriseCapabilityReadModel{nameExists: true}

	handler := newTestableCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

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
	mockRepo := &mockEnterpriseCapabilityRepository{}
	mockReadModel := &mockEnterpriseCapabilityReadModel{nameExists: false}

	handler := newTestableCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

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
	mockRepo := &mockEnterpriseCapabilityRepository{}
	mockReadModel := &mockEnterpriseCapabilityReadModel{nameExists: false}

	handler := newTestableCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

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
	mockRepo := &mockEnterpriseCapabilityRepository{}
	mockReadModel := &mockEnterpriseCapabilityReadModel{checkErr: errors.New("database error")}

	handler := newTestableCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

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
	mockRepo := &mockEnterpriseCapabilityRepository{saveErr: errors.New("save error")}
	mockReadModel := &mockEnterpriseCapabilityReadModel{nameExists: false}

	handler := newTestableCreateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateEnterpriseCapability{
		Name:        "Test Capability",
		Description: "Test",
		Category:    "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
