//go:build integration

package fixtures

import (
	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/projectors"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/infrastructure/adapters"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"

	"github.com/stretchr/testify/require"
)

type BusinessDomainFixtures struct {
	tc                  *TestContext
	domainReadModel     *readmodels.BusinessDomainReadModel
	assignmentReadModel *readmodels.DomainCapabilityAssignmentReadModel
	capabilityReadModel *readmodels.CapabilityReadModel
}

func NewBusinessDomainFixtures(tc *TestContext) *BusinessDomainFixtures {
	domainReadModel := readmodels.NewBusinessDomainReadModel(tc.TenantDB)
	assignmentReadModel := readmodels.NewDomainCapabilityAssignmentReadModel(tc.TenantDB)
	capabilityReadModel := readmodels.NewCapabilityReadModel(tc.TenantDB)

	domainRepo := repositories.NewBusinessDomainRepository(tc.EventStore)
	assignmentRepo := repositories.NewBusinessDomainAssignmentRepository(tc.EventStore)
	capabilityRepo := repositories.NewCapabilityRepository(tc.EventStore)

	domainProjector := projectors.NewBusinessDomainProjector(domainReadModel)
	tc.EventBus.Subscribe("BusinessDomainCreated", domainProjector)
	tc.EventBus.Subscribe("BusinessDomainUpdated", domainProjector)
	tc.EventBus.Subscribe("BusinessDomainDeleted", domainProjector)
	tc.EventBus.Subscribe("CapabilityAssignedToDomain", domainProjector)
	tc.EventBus.Subscribe("CapabilityUnassignedFromDomain", domainProjector)

	assignmentProjector := projectors.NewBusinessDomainAssignmentProjector(assignmentReadModel, domainReadModel, capabilityReadModel)
	tc.EventBus.Subscribe("CapabilityAssignedToDomain", assignmentProjector)
	tc.EventBus.Subscribe("CapabilityUnassignedFromDomain", assignmentProjector)

	assignmentChecker := adapters.NewBusinessDomainAssignmentCheckerAdapter(assignmentReadModel)
	deletionService := services.NewBusinessDomainDeletionService(assignmentChecker)

	tc.CommandBus.Register("CreateBusinessDomain", handlers.NewCreateBusinessDomainHandler(domainRepo, domainReadModel))
	tc.CommandBus.Register("UpdateBusinessDomain", handlers.NewUpdateBusinessDomainHandler(domainRepo, domainReadModel))
	tc.CommandBus.Register("DeleteBusinessDomain", handlers.NewDeleteBusinessDomainHandler(domainRepo, deletionService))
	tc.CommandBus.Register("AssignCapabilityToDomain", handlers.NewAssignCapabilityToDomainHandler(assignmentRepo, capabilityRepo, domainReadModel, assignmentReadModel))
	tc.CommandBus.Register("UnassignCapabilityFromDomain", handlers.NewUnassignCapabilityFromDomainHandler(assignmentRepo))

	return &BusinessDomainFixtures{
		tc:                  tc,
		domainReadModel:     domainReadModel,
		assignmentReadModel: assignmentReadModel,
		capabilityReadModel: capabilityReadModel,
	}
}

type CreateBusinessDomainInput struct {
	Name        string
	Description string
}

func (f *BusinessDomainFixtures) CreateBusinessDomain(input CreateBusinessDomainInput) string {
	cmd := &commands.CreateBusinessDomain{
		Name:        input.Name,
		Description: input.Description,
	}

	result := f.tc.MustDispatch(cmd)
	f.tc.TrackID(result.CreatedID)
	return result.CreatedID
}

func (f *BusinessDomainFixtures) CreateDomain(name string) string {
	return f.CreateBusinessDomain(CreateBusinessDomainInput{Name: name})
}

func (f *BusinessDomainFixtures) AssignCapabilityToDomain(capabilityID, domainID string) string {
	cmd := &commands.AssignCapabilityToDomain{
		CapabilityID:     capabilityID,
		BusinessDomainID: domainID,
	}

	result := f.tc.MustDispatch(cmd)
	f.tc.TrackID(result.CreatedID)
	return result.CreatedID
}

func (f *BusinessDomainFixtures) GetDomain(id string) *readmodels.BusinessDomainDTO {
	domain, err := f.domainReadModel.GetByID(f.tc.Ctx, id)
	require.NoError(f.tc.T, err)
	return domain
}

func (f *BusinessDomainFixtures) ReadModel() *readmodels.BusinessDomainReadModel {
	return f.domainReadModel
}

func (f *BusinessDomainFixtures) AssignmentReadModel() *readmodels.DomainCapabilityAssignmentReadModel {
	return f.assignmentReadModel
}
