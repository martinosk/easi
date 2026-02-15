package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubNameLookup struct {
	names map[string]string
	err   error
}

func (s *stubNameLookup) GetArtifactName(_ context.Context, artifactID string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.names[artifactID], nil
}

func defaultDeps() ArtifactNameResolverDeps {
	empty := &stubNameLookup{names: map[string]string{}}
	return ArtifactNameResolverDeps{
		Capabilities:     empty,
		Components:       empty,
		Views:            empty,
		Domains:          empty,
		Vendors:          empty,
		AcquiredEntities: empty,
		InternalTeams:    empty,
	}
}

func TestArtifactNameResolver_ResolvesName(t *testing.T) {
	tests := []struct {
		name         string
		override     func(*ArtifactNameResolverDeps)
		artifactType string
		artifactID   string
		expected     string
	}{
		{"capability", func(d *ArtifactNameResolverDeps) {
			d.Capabilities = &stubNameLookup{names: map[string]string{"cap-1": "Customer Onboarding"}}
		}, "capability", "cap-1", "Customer Onboarding"},
		{"component", func(d *ArtifactNameResolverDeps) {
			d.Components = &stubNameLookup{names: map[string]string{"comp-1": "Payment Service"}}
		}, "component", "comp-1", "Payment Service"},
		{"view", func(d *ArtifactNameResolverDeps) {
			d.Views = &stubNameLookup{names: map[string]string{"view-1": "Integration Map"}}
		}, "view", "view-1", "Integration Map"},
		{"domain", func(d *ArtifactNameResolverDeps) {
			d.Domains = &stubNameLookup{names: map[string]string{"dom-1": "Sales"}}
		}, "domain", "dom-1", "Sales"},
		{"vendor", func(d *ArtifactNameResolverDeps) {
			d.Vendors = &stubNameLookup{names: map[string]string{"ven-1": "Acme"}}
		}, "vendor", "ven-1", "Acme"},
		{"acquired_entity", func(d *ArtifactNameResolverDeps) {
			d.AcquiredEntities = &stubNameLookup{names: map[string]string{"ae-1": "Widget Co"}}
		}, "acquired_entity", "ae-1", "Widget Co"},
		{"internal_team", func(d *ArtifactNameResolverDeps) {
			d.InternalTeams = &stubNameLookup{names: map[string]string{"team-1": "Platform"}}
		}, "internal_team", "team-1", "Platform"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := defaultDeps()
			tt.override(&deps)
			resolver := NewArtifactNameResolver(deps)

			name, err := resolver.ResolveName(context.Background(), tt.artifactType, tt.artifactID)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, name)
		})
	}
}

func TestArtifactNameResolver_ReturnsDeletedArtifact(t *testing.T) {
	tests := []struct {
		name         string
		override     func(*ArtifactNameResolverDeps)
		artifactType string
		artifactID   string
	}{
		{"when not found", nil, "capability", "nonexistent"},
		{"when lookup errors", func(d *ArtifactNameResolverDeps) {
			d.Capabilities = &stubNameLookup{err: errors.New("database error")}
		}, "capability", "cap-1"},
		{"for unknown artifact type", nil, "unknown_type", "id-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := defaultDeps()
			if tt.override != nil {
				tt.override(&deps)
			}
			resolver := NewArtifactNameResolver(deps)

			name, err := resolver.ResolveName(context.Background(), tt.artifactType, tt.artifactID)
			assert.NoError(t, err)
			assert.Equal(t, "Deleted artifact", name)
		})
	}
}

func TestArtifactNameResolver_DispatchesToCorrectLookup(t *testing.T) {
	deps := defaultDeps()
	deps.Capabilities = &stubNameLookup{names: map[string]string{"shared-id": "Cap Name"}}
	deps.Components = &stubNameLookup{names: map[string]string{"shared-id": "Comp Name"}}
	deps.Views = &stubNameLookup{names: map[string]string{"shared-id": "View Name"}}
	resolver := NewArtifactNameResolver(deps)

	name, _ := resolver.ResolveName(context.Background(), "capability", "shared-id")
	assert.Equal(t, "Cap Name", name)

	name, _ = resolver.ResolveName(context.Background(), "component", "shared-id")
	assert.Equal(t, "Comp Name", name)

	name, _ = resolver.ResolveName(context.Background(), "view", "shared-id")
	assert.Equal(t, "View Name", name)
}
