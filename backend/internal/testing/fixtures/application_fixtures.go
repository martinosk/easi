//go:build integration

package fixtures

import (
	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/handlers"
	"easi/backend/internal/architecturemodeling/application/projectors"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"

	"github.com/stretchr/testify/require"
)

type ApplicationFixtures struct {
	tc                 *TestContext
	componentReadModel *readmodels.ApplicationComponentReadModel
	relationReadModel  *readmodels.ComponentRelationReadModel
}

func NewApplicationFixtures(tc *TestContext) *ApplicationFixtures {
	componentReadModel := readmodels.NewApplicationComponentReadModel(tc.TenantDB)
	relationReadModel := readmodels.NewComponentRelationReadModel(tc.TenantDB)

	componentRepo := repositories.NewApplicationComponentRepository(tc.EventStore)
	relationRepo := repositories.NewComponentRelationRepository(tc.EventStore)

	componentProjector := projectors.NewApplicationComponentProjector(componentReadModel)
	tc.EventBus.Subscribe("ApplicationComponentCreated", componentProjector)
	tc.EventBus.Subscribe("ApplicationComponentUpdated", componentProjector)
	tc.EventBus.Subscribe("ApplicationComponentDeleted", componentProjector)

	relationProjector := projectors.NewComponentRelationProjector(relationReadModel)
	tc.EventBus.Subscribe("ComponentRelationCreated", relationProjector)
	tc.EventBus.Subscribe("ComponentRelationUpdated", relationProjector)
	tc.EventBus.Subscribe("ComponentRelationDeleted", relationProjector)

	createComponentHandler := handlers.NewCreateApplicationComponentHandler(componentRepo)
	updateComponentHandler := handlers.NewUpdateApplicationComponentHandler(componentRepo)
	deleteComponentHandler := handlers.NewDeleteApplicationComponentHandler(componentRepo, relationReadModel, tc.CommandBus)
	createRelationHandler := handlers.NewCreateComponentRelationHandler(relationRepo)
	updateRelationHandler := handlers.NewUpdateComponentRelationHandler(relationRepo)
	deleteRelationHandler := handlers.NewDeleteComponentRelationHandler(relationRepo)

	tc.CommandBus.Register("CreateApplicationComponent", createComponentHandler)
	tc.CommandBus.Register("UpdateApplicationComponent", updateComponentHandler)
	tc.CommandBus.Register("DeleteApplicationComponent", deleteComponentHandler)
	tc.CommandBus.Register("CreateComponentRelation", createRelationHandler)
	tc.CommandBus.Register("UpdateComponentRelation", updateRelationHandler)
	tc.CommandBus.Register("DeleteComponentRelation", deleteRelationHandler)

	return &ApplicationFixtures{
		tc:                 tc,
		componentReadModel: componentReadModel,
		relationReadModel:  relationReadModel,
	}
}

type CreateComponentInput struct {
	Name        string
	Description string
}

func (f *ApplicationFixtures) CreateComponent(input CreateComponentInput) string {
	cmd := &commands.CreateApplicationComponent{
		Name:        input.Name,
		Description: input.Description,
	}

	result := f.tc.MustDispatch(cmd)
	f.tc.TrackID(result.CreatedID)
	return result.CreatedID
}

func (f *ApplicationFixtures) CreateApplication(name string) string {
	return f.CreateComponent(CreateComponentInput{Name: name})
}

func (f *ApplicationFixtures) GetComponent(id string) *readmodels.ApplicationComponentDTO {
	component, err := f.componentReadModel.GetByID(f.tc.Ctx, id)
	require.NoError(f.tc.T, err)
	return component
}

func (f *ApplicationFixtures) ReadModel() *readmodels.ApplicationComponentReadModel {
	return f.componentReadModel
}

func (f *ApplicationFixtures) RelationReadModel() *readmodels.ComponentRelationReadModel {
	return f.relationReadModel
}
