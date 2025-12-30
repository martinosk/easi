package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBusinessDomainRepository struct {
	savedDomains []*aggregates.BusinessDomain
	saveErr      error
}

func (m *mockBusinessDomainRepository) Save(ctx context.Context, domain *aggregates.BusinessDomain) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedDomains = append(m.savedDomains, domain)
	return nil
}

type mockBusinessDomainReadModel struct {
	nameExists bool
	checkErr   error
}

func (m *mockBusinessDomainReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.nameExists, nil
}

type businessDomainRepository interface {
	Save(ctx context.Context, domain *aggregates.BusinessDomain) error
}

type businessDomainReadModelForCreate interface {
	NameExists(ctx context.Context, name, excludeID string) (bool, error)
}

type testableCreateBusinessDomainHandler struct {
	repository businessDomainRepository
	readModel  businessDomainReadModelForCreate
}

func newTestableCreateBusinessDomainHandler(
	repository businessDomainRepository,
	readModel businessDomainReadModelForCreate,
) *testableCreateBusinessDomainHandler {
	return &testableCreateBusinessDomainHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *testableCreateBusinessDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateBusinessDomain)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	exists, err := h.readModel.NameExists(ctx, command.Name, "")
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if exists {
		return cqrs.EmptyResult(), ErrBusinessDomainNameExists
	}

	name, err := valueobjects.NewDomainName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	description := valueobjects.MustNewDescription(command.Description)

	domain, err := aggregates.NewBusinessDomain(name, description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, domain); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(domain.ID()), nil
}

func TestCreateBusinessDomainHandler_CreatesBusinessDomain(t *testing.T) {
	mockRepo := &mockBusinessDomainRepository{}
	mockReadModel := &mockBusinessDomainReadModel{nameExists: false}

	handler := newTestableCreateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateBusinessDomain{
		Name:        "Customer Management",
		Description: "Manages customer relationships",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedDomains, 1, "Handler should create exactly 1 domain")

	domain := mockRepo.savedDomains[0]
	assert.Equal(t, "Customer Management", domain.Name().Value())
	assert.Equal(t, "Manages customer relationships", domain.Description().Value())
}

func TestCreateBusinessDomainHandler_ReturnsCreatedID(t *testing.T) {
	mockRepo := &mockBusinessDomainRepository{}
	mockReadModel := &mockBusinessDomainReadModel{nameExists: false}

	handler := newTestableCreateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateBusinessDomain{
		Name:        "Order Processing",
		Description: "",
	}

	result, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, result.CreatedID, "Result CreatedID should be set after handling")
	assert.Equal(t, mockRepo.savedDomains[0].ID(), result.CreatedID)
}

func TestCreateBusinessDomainHandler_NameExists_ReturnsError(t *testing.T) {
	mockRepo := &mockBusinessDomainRepository{}
	mockReadModel := &mockBusinessDomainReadModel{nameExists: true}

	handler := newTestableCreateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateBusinessDomain{
		Name:        "Duplicate Name",
		Description: "Should fail",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrBusinessDomainNameExists)
	assert.Empty(t, mockRepo.savedDomains, "Should not save domain when name exists")
}

func TestCreateBusinessDomainHandler_InvalidName_ReturnsError(t *testing.T) {
	mockRepo := &mockBusinessDomainRepository{}
	mockReadModel := &mockBusinessDomainReadModel{nameExists: false}

	handler := newTestableCreateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateBusinessDomain{
		Name:        "",
		Description: "Invalid name",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedDomains, "Should not save domain with invalid name")
}

func TestCreateBusinessDomainHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockBusinessDomainRepository{}
	mockReadModel := &mockBusinessDomainReadModel{}

	handler := newTestableCreateBusinessDomainHandler(mockRepo, mockReadModel)

	invalidCmd := &commands.DeleteBusinessDomain{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestCreateBusinessDomainHandler_ReadModelError_ReturnsError(t *testing.T) {
	mockRepo := &mockBusinessDomainRepository{}
	mockReadModel := &mockBusinessDomainReadModel{checkErr: errors.New("database error")}

	handler := newTestableCreateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.CreateBusinessDomain{
		Name:        "Test Domain",
		Description: "Test",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedDomains)
}
