package api

import (
	"testing"

	"easi/backend/internal/onepagers/application/ports"

	"github.com/stretchr/testify/assert"
)

func TestMapReferences_PreservesOrderAndSubjectType(t *testing.T) {
	type edge struct {
		id   string
		name string
	}
	edges := []edge{{id: "a", name: "Alpha"}, {id: "b", name: "Beta"}}

	value := mapReferences(edges, func(e edge) ports.Reference {
		return ports.Reference{ID: e.id, Label: e.name, SubjectType: "application"}
	})

	assert.Equal(t, ports.ReferenceListValue{References: []ports.Reference{
		{ID: "a", Label: "Alpha", SubjectType: "application"},
		{ID: "b", Label: "Beta", SubjectType: "application"},
	}}, value)
}

func TestMapReferences_EmptyEdgesYieldEmptyList(t *testing.T) {
	value := mapReferences(nil, func(e string) ports.Reference { return ports.Reference{ID: e} })

	assert.Empty(t, value.References)
	assert.False(t, ports.ValueFilled(value))
}

func TestNamedReferences_ResolvesLabelsFromNameMapPreservingIDOrder(t *testing.T) {
	value := namedReferences([]string{"c-2", "c-1"}, "capability", map[string]string{"c-1": "Billing", "c-2": "Orders"})

	assert.Equal(t, ports.ReferenceListValue{References: []ports.Reference{
		{ID: "c-2", Label: "Orders", SubjectType: "capability"},
		{ID: "c-1", Label: "Billing", SubjectType: "capability"},
	}}, value)
}
