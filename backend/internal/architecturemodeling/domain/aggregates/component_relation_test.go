package aggregates

import (
	"testing"

	"easi/backend/internal/architecturemodeling/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewComponentRelation_ValidInputs(t *testing.T) {
	sourceID, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)

	targetID, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)

	relationType, err := valueobjects.NewRelationType("Triggers")
	require.NoError(t, err)

	name := valueobjects.MustNewDescription("User triggers order")
	description := valueobjects.MustNewDescription("When user submits order, it triggers order processing")

	properties := valueobjects.NewRelationProperties(valueobjects.RelationPropertiesParams{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		Name:         name,
		Description:  description,
	})
	relation, err := NewComponentRelation(properties)

	require.NoError(t, err)
	assert.NotNil(t, relation)
	assert.NotEmpty(t, relation.ID())
	assert.Equal(t, sourceID, relation.SourceComponentID())
	assert.Equal(t, targetID, relation.TargetComponentID())
	assert.Equal(t, relationType, relation.RelationType())
	assert.Equal(t, name, relation.Name())
	assert.Equal(t, description, relation.Description())
	assert.NotZero(t, relation.CreatedAt())
}

func TestNewComponentRelation_SelfReference(t *testing.T) {
	componentID, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)

	relationType, err := valueobjects.NewRelationType("Serves")
	require.NoError(t, err)

	name := valueobjects.MustNewDescription("Self relation")
	description := valueobjects.MustNewDescription("Should not be allowed")

	properties := valueobjects.NewRelationProperties(valueobjects.RelationPropertiesParams{
		SourceID:     componentID,
		TargetID:     componentID,
		RelationType: relationType,
		Name:         name,
		Description:  description,
	})
	relation, err := NewComponentRelation(properties)

	assert.Error(t, err)
	assert.Equal(t, ErrSelfReference, err)
	assert.Nil(t, relation)
}

func TestNewComponentRelation_RaisesCreatedEvent(t *testing.T) {
	sourceID, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)

	targetID, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)

	relationType, err := valueobjects.NewRelationType("Triggers")
	require.NoError(t, err)

	name := valueobjects.MustNewDescription("Test relation")
	description := valueobjects.MustNewDescription("Test description")

	properties := valueobjects.NewRelationProperties(valueobjects.RelationPropertiesParams{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		Name:         name,
		Description:  description,
	})
	relation, err := NewComponentRelation(properties)
	require.NoError(t, err)

	uncommittedEvents := relation.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ComponentRelationCreated", uncommittedEvents[0].EventType())
}

func TestNewComponentRelation_WithServesType(t *testing.T) {
	sourceID, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)

	targetID, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)

	relationType, err := valueobjects.NewRelationType("Serves")
	require.NoError(t, err)

	name := valueobjects.MustNewDescription("API serves UI")
	description := valueobjects.MustNewDescription("API provides services to UI")

	properties := valueobjects.NewRelationProperties(valueobjects.RelationPropertiesParams{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		Name:         name,
		Description:  description,
	})
	relation, err := NewComponentRelation(properties)

	require.NoError(t, err)
	assert.NotNil(t, relation)
	assert.Equal(t, relationType, relation.RelationType())
	assert.Equal(t, "Serves", relation.RelationType().Value())
}

func TestComponentRelation_Update(t *testing.T) {
	sourceID, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)

	targetID, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)

	relationType, err := valueobjects.NewRelationType("Triggers")
	require.NoError(t, err)

	name := valueobjects.MustNewDescription("Original name")
	description := valueobjects.MustNewDescription("Original description")

	properties := valueobjects.NewRelationProperties(valueobjects.RelationPropertiesParams{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		Name:         name,
		Description:  description,
	})
	relation, err := NewComponentRelation(properties)
	require.NoError(t, err)

	// Clear uncommitted events to test update event separately
	relation.MarkChangesAsCommitted()

	// Update the relation
	newName := valueobjects.MustNewDescription("Updated name")
	newDescription := valueobjects.MustNewDescription("Updated description")

	err = relation.Update(newName, newDescription)

	require.NoError(t, err)
	assert.Equal(t, newName, relation.Name())
	assert.Equal(t, newDescription, relation.Description())

	// Verify update event was raised
	uncommittedEvents := relation.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ComponentRelationUpdated", uncommittedEvents[0].EventType())
}

func TestLoadComponentRelationFromHistory(t *testing.T) {
	// First, create a relation and capture its events
	sourceID, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)

	targetID, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)

	relationType, err := valueobjects.NewRelationType("Triggers")
	require.NoError(t, err)

	name := valueobjects.MustNewDescription("Test relation")
	description := valueobjects.MustNewDescription("Test description")

	properties := valueobjects.NewRelationProperties(valueobjects.RelationPropertiesParams{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		Name:         name,
		Description:  description,
	})
	originalRelation, err := NewComponentRelation(properties)
	require.NoError(t, err)

	events := originalRelation.GetUncommittedChanges()

	// Now reconstruct from history
	reconstructedRelation, err := LoadComponentRelationFromHistory(events)

	require.NoError(t, err)
	assert.NotNil(t, reconstructedRelation)
	assert.Equal(t, originalRelation.ID(), reconstructedRelation.ID())
	assert.Equal(t, originalRelation.SourceComponentID(), reconstructedRelation.SourceComponentID())
	assert.Equal(t, originalRelation.TargetComponentID(), reconstructedRelation.TargetComponentID())
	assert.Equal(t, originalRelation.RelationType(), reconstructedRelation.RelationType())
	assert.Equal(t, originalRelation.Name(), reconstructedRelation.Name())
	assert.Equal(t, originalRelation.Description(), reconstructedRelation.Description())
}

func TestComponentRelation_WithEmptyNameAndDescription(t *testing.T) {
	sourceID, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)

	targetID, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)

	relationType, err := valueobjects.NewRelationType("Serves")
	require.NoError(t, err)

	name := valueobjects.MustNewDescription("")
	description := valueobjects.MustNewDescription("")

	properties := valueobjects.NewRelationProperties(valueobjects.RelationPropertiesParams{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		Name:         name,
		Description:  description,
	})
	relation, err := NewComponentRelation(properties)

	require.NoError(t, err)
	assert.NotNil(t, relation)
	assert.Equal(t, "", relation.Name().Value())
	assert.Equal(t, "", relation.Description().Value())
}
