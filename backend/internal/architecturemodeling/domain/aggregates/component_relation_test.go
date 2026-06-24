package aggregates

import (
	"testing"

	"easi/backend/internal/architecturemodeling/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type relationSpec struct {
	relationType string
	name         string
	description  string
	sourceID     *valueobjects.ComponentID
	targetID     *valueobjects.ComponentID
}

type relationFixture struct {
	sourceID     valueobjects.ComponentID
	targetID     valueobjects.ComponentID
	relationType valueobjects.RelationType
	name         valueobjects.Description
	description  valueobjects.Description
	properties   valueobjects.RelationProperties
}

func newComponentID(t *testing.T) valueobjects.ComponentID {
	t.Helper()
	id, err := valueobjects.NewComponentIDFromString(uuid.New().String())
	require.NoError(t, err)
	return id
}

func newRelationFixture(t *testing.T, spec relationSpec) relationFixture {
	t.Helper()

	sourceID := newComponentID(t)
	if spec.sourceID != nil {
		sourceID = *spec.sourceID
	}
	targetID := newComponentID(t)
	if spec.targetID != nil {
		targetID = *spec.targetID
	}

	relationType, err := valueobjects.NewRelationType(spec.relationType)
	require.NoError(t, err)

	nameDesc := valueobjects.MustNewDescription(spec.name)
	descriptionDesc := valueobjects.MustNewDescription(spec.description)

	properties := valueobjects.NewRelationProperties(valueobjects.RelationPropertiesParams{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		Name:         nameDesc,
		Description:  descriptionDesc,
	})

	return relationFixture{
		sourceID:     sourceID,
		targetID:     targetID,
		relationType: relationType,
		name:         nameDesc,
		description:  descriptionDesc,
		properties:   properties,
	}
}

func TestNewComponentRelation_ValidInputs(t *testing.T) {
	f := newRelationFixture(t, relationSpec{
		relationType: "Triggers",
		name:         "User triggers order",
		description:  "When user submits order, it triggers order processing",
	})

	relation, err := NewComponentRelation(f.properties)

	require.NoError(t, err)
	assert.NotNil(t, relation)
	assert.NotEmpty(t, relation.ID())
	assert.Equal(t, f.sourceID, relation.SourceComponentID())
	assert.Equal(t, f.targetID, relation.TargetComponentID())
	assert.Equal(t, f.relationType, relation.RelationType())
	assert.Equal(t, f.name, relation.Name())
	assert.Equal(t, f.description, relation.Description())
	assert.NotZero(t, relation.CreatedAt())
}

func TestNewComponentRelation_SelfReference(t *testing.T) {
	componentID := newComponentID(t)

	f := newRelationFixture(t, relationSpec{
		relationType: "Serves",
		name:         "Self relation",
		description:  "Should not be allowed",
		sourceID:     &componentID,
		targetID:     &componentID,
	})

	relation, err := NewComponentRelation(f.properties)

	assert.Error(t, err)
	assert.Equal(t, ErrSelfReference, err)
	assert.Nil(t, relation)
}

func TestNewComponentRelation_RaisesCreatedEvent(t *testing.T) {
	f := newRelationFixture(t, relationSpec{
		relationType: "Triggers",
		name:         "Test relation",
		description:  "Test description",
	})

	relation, err := NewComponentRelation(f.properties)
	require.NoError(t, err)

	uncommittedEvents := relation.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ComponentRelationCreated", uncommittedEvents[0].EventType())
}

func TestNewComponentRelation_ValidVariants(t *testing.T) {
	cases := []struct {
		name   string
		spec   relationSpec
		verify func(t *testing.T, f relationFixture, relation *ComponentRelation)
	}{
		{
			name: "serves type",
			spec: relationSpec{relationType: "Serves", name: "API serves UI", description: "API provides services to UI"},
			verify: func(t *testing.T, f relationFixture, relation *ComponentRelation) {
				assert.Equal(t, f.relationType, relation.RelationType())
				assert.Equal(t, "Serves", relation.RelationType().Value())
			},
		},
		{
			name: "empty name and description",
			spec: relationSpec{relationType: "Serves", name: "", description: ""},
			verify: func(t *testing.T, _ relationFixture, relation *ComponentRelation) {
				assert.Equal(t, "", relation.Name().Value())
				assert.Equal(t, "", relation.Description().Value())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newRelationFixture(t, tc.spec)

			relation, err := NewComponentRelation(f.properties)

			require.NoError(t, err)
			assert.NotNil(t, relation)
			tc.verify(t, f, relation)
		})
	}
}

func TestComponentRelation_Update(t *testing.T) {
	f := newRelationFixture(t, relationSpec{
		relationType: "Triggers",
		name:         "Original name",
		description:  "Original description",
	})

	relation, err := NewComponentRelation(f.properties)
	require.NoError(t, err)

	relation.MarkChangesAsCommitted()

	newName := valueobjects.MustNewDescription("Updated name")
	newDescription := valueobjects.MustNewDescription("Updated description")

	err = relation.Update(newName, newDescription)

	require.NoError(t, err)
	assert.Equal(t, newName, relation.Name())
	assert.Equal(t, newDescription, relation.Description())

	uncommittedEvents := relation.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ComponentRelationUpdated", uncommittedEvents[0].EventType())
}

func TestLoadComponentRelationFromHistory(t *testing.T) {
	f := newRelationFixture(t, relationSpec{
		relationType: "Triggers",
		name:         "Test relation",
		description:  "Test description",
	})

	originalRelation, err := NewComponentRelation(f.properties)
	require.NoError(t, err)

	events := originalRelation.GetUncommittedChanges()

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

