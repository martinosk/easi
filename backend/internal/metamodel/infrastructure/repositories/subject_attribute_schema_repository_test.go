package repositories

import (
	"testing"

	"easi/backend/internal/metamodel/domain/events"

	"github.com/stretchr/testify/assert"
)

func TestSubjectAttributeSchemaDeserializers_CoverEverySchemaEventType(t *testing.T) {
	for _, eventType := range events.SubjectAttributeSchemaEventTypes() {
		assert.Truef(t, subjectAttributeSchemaEventDeserializers.HasDeserializerFor(eventType),
			"No deserializer registered for %q", eventType)
	}
}
