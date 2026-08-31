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
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"

	"github.com/stretchr/testify/require"
)

type CapabilityFixtures struct {
	tc                  *TestContext
	capabilityReadModel *readmodels.CapabilityReadModel
	assignmentReadModel *readmodels.DomainCapabilityAssignmentReadModel
}

func NewCapabilityFixtures(tc *TestContext) *CapabilityFixtures {
	capabilityReadModel := readmodels.NewCapabilityReadModel(tc.TenantDB)
	assignmentReadModel := readmodels.NewDomainCapabilityAssignmentReadModel(tc.TenantDB)

	capabilityRepo := repositories.NewCapabilityRepository(tc.EventStore)
	childrenChecker := adapters.NewCapabilityChildrenCheckerAdapter(capabilityReadModel)
	deletionService := services.NewCapabilityDeletionService(childrenChecker)

	projector := projectors.NewCapabilityProjector(capabilityReadModel, assignmentReadModel)
	tc.EventBus.Subscribe(cmPL.CapabilityCreated, projector)
	tc.EventBus.Subscribe(cmPL.CapabilityUpdated, projector)
	tc.EventBus.Subscribe(cmPL.CapabilityMetadataUpdated, projector)
	tc.EventBus.Subscribe(cmPL.CapabilityExpertAdded, projector)
	tc.EventBus.Subscribe(cmPL.CapabilityTagAdded, projector)
	tc.EventBus.Subscribe(cmPL.CapabilityDeleted, projector)

	tc.CommandBus.Register("CreateCapability", handlers.NewCreateCapabilityHandler(capabilityRepo))
	tc.CommandBus.Register("UpdateCapability", handlers.NewUpdateCapabilityHandler(capabilityRepo))
	tc.CommandBus.Register("UpdateCapabilityMetadata", handlers.NewUpdateCapabilityMetadataHandler(capabilityRepo, readmodels.NewUserNameCacheReadModel(tc.TenantDB)))
	tc.CommandBus.Register("AddCapabilityExpert", handlers.NewAddCapabilityExpertHandler(capabilityRepo))
	tc.CommandBus.Register("AddCapabilityTag", handlers.NewAddCapabilityTagHandler(capabilityRepo))
	realizationReadModel := readmodels.NewRealizationReadModel(tc.TenantDB)
	tc.CommandBus.Register("DeleteCapability", handlers.NewDeleteCapabilityHandler(capabilityRepo, deletionService, realizationReadModel, capabilityReadModel))

	return &CapabilityFixtures{
		tc:                  tc,
		capabilityReadModel: capabilityReadModel,
		assignmentReadModel: assignmentReadModel,
	}
}

type CreateCapabilityInput struct {
	Name        string
	Description string
	ParentID    string
	Level       string
}

func (cf *CapabilityFixtures) CreateCapability(input CreateCapabilityInput) string {
	if input.Level == "" {
		input.Level = "L1"
	}

	cmd := &commands.CreateCapability{
		Name:        input.Name,
		Description: input.Description,
		ParentID:    input.ParentID,
		Level:       input.Level,
	}

	result := cf.tc.MustDispatch(cmd)
	cf.tc.TrackID(result.CreatedID)
	return result.CreatedID
}

func (cf *CapabilityFixtures) CreateL1Capability(name string) string {
	return cf.CreateCapability(CreateCapabilityInput{
		Name:  name,
		Level: "L1",
	})
}

func (cf *CapabilityFixtures) CreateChildCapability(name, parentID, level string) string {
	return cf.CreateCapability(CreateCapabilityInput{
		Name:     name,
		ParentID: parentID,
		Level:    level,
	})
}

func (cf *CapabilityFixtures) GetCapability(id string) *readmodels.CapabilityDTO {
	capability, err := cf.capabilityReadModel.GetByID(cf.tc.Ctx, id)
	require.NoError(cf.tc.T, err)
	return capability
}

func (cf *CapabilityFixtures) ReadModel() *readmodels.CapabilityReadModel {
	return cf.capabilityReadModel
}
