package services

import (
	"context"
	"errors"
	"testing"

	appservices "easi/backend/internal/accessdelegation/application/services"

	"github.com/stretchr/testify/assert"
)

type stubNameLookup struct {
	names map[string]string
	err   error
}

func (s *stubNameLookup) lookup(id string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	name, ok := s.names[id]
	if !ok {
		return "", nil
	}
	return name, nil
}

type testableResolver struct {
	capabilities *stubNameLookup
	components   *stubNameLookup
	views        *stubNameLookup
}

func newTestableResolver(caps, comps, views *stubNameLookup) appservices.ArtifactNameResolver {
	return &testableResolver{capabilities: caps, components: comps, views: views}
}

func (r *testableResolver) ResolveName(_ context.Context, artifactType, artifactID string) (string, error) {
	var lookup *stubNameLookup
	switch artifactType {
	case "capability":
		lookup = r.capabilities
	case "component":
		lookup = r.components
	case "view":
		lookup = r.views
	default:
		return "Deleted artifact", nil
	}

	name, err := lookup.lookup(artifactID)
	if err != nil || name == "" {
		return "Deleted artifact", nil
	}
	return name, nil
}

func TestArtifactNameResolver_ResolvesName(t *testing.T) {
	tests := []struct {
		name         string
		caps         *stubNameLookup
		comps        *stubNameLookup
		views        *stubNameLookup
		artifactType string
		artifactID   string
		expected     string
	}{
		{"capability", &stubNameLookup{names: map[string]string{"cap-1": "Customer Onboarding"}}, &stubNameLookup{}, &stubNameLookup{}, "capability", "cap-1", "Customer Onboarding"},
		{"component", &stubNameLookup{}, &stubNameLookup{names: map[string]string{"comp-1": "Payment Service"}}, &stubNameLookup{}, "component", "comp-1", "Payment Service"},
		{"view", &stubNameLookup{}, &stubNameLookup{}, &stubNameLookup{names: map[string]string{"view-1": "Integration Map"}}, "view", "view-1", "Integration Map"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := newTestableResolver(tt.caps, tt.comps, tt.views)

			name, err := resolver.ResolveName(context.Background(), tt.artifactType, tt.artifactID)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, name)
		})
	}
}

func TestArtifactNameResolver_ReturnsDeletedArtifact(t *testing.T) {
	tests := []struct {
		name         string
		caps         *stubNameLookup
		artifactType string
		artifactID   string
	}{
		{"when not found", &stubNameLookup{names: map[string]string{}}, "capability", "nonexistent"},
		{"when lookup errors", &stubNameLookup{err: errors.New("database error")}, "capability", "cap-1"},
		{"for unknown artifact type", &stubNameLookup{}, "unknown", "id-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := newTestableResolver(tt.caps, &stubNameLookup{}, &stubNameLookup{})

			name, err := resolver.ResolveName(context.Background(), tt.artifactType, tt.artifactID)
			assert.NoError(t, err)
			assert.Equal(t, "Deleted artifact", name)
		})
	}
}

func TestArtifactNameResolver_DispatchesToCorrectReadModel(t *testing.T) {
	resolver := newTestableResolver(
		&stubNameLookup{names: map[string]string{"shared-id": "Cap Name"}},
		&stubNameLookup{names: map[string]string{"shared-id": "Comp Name"}},
		&stubNameLookup{names: map[string]string{"shared-id": "View Name"}},
	)

	name, _ := resolver.ResolveName(context.Background(), "capability", "shared-id")
	assert.Equal(t, "Cap Name", name)

	name, _ = resolver.ResolveName(context.Background(), "component", "shared-id")
	assert.Equal(t, "Comp Name", name)

	name, _ = resolver.ResolveName(context.Background(), "view", "shared-id")
	assert.Equal(t, "View Name", name)
}
