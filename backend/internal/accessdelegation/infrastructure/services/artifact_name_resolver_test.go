package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubArtifactNameCache struct {
	names         map[string]map[string]string
	err           error
	queriedType   string
	queriedIDs    []string
	queryAttempts int
}

func (s *stubArtifactNameCache) NamesByIDs(_ context.Context, artifactType string, artifactIDs []string) (map[string]string, error) {
	s.queriedType = artifactType
	s.queriedIDs = artifactIDs
	s.queryAttempts++
	if s.err != nil {
		return nil, s.err
	}
	return s.names[artifactType], nil
}

func cacheWith(artifactType, artifactID, name string) *stubArtifactNameCache {
	return &stubArtifactNameCache{names: map[string]map[string]string{artifactType: {artifactID: name}}}
}

func TestArtifactNameResolver_ResolvesCachedName(t *testing.T) {
	tests := []struct {
		artifactType string
		artifactID   string
		expected     string
	}{
		{"capability", "cap-1", "Customer Onboarding"},
		{"component", "comp-1", "Payment Service"},
		{"view", "view-1", "Integration Map"},
		{"domain", "dom-1", "Sales"},
		{"vendor", "ven-1", "Acme"},
		{"acquired_entity", "ae-1", "Widget Co"},
		{"internal_team", "team-1", "Platform"},
	}

	for _, tt := range tests {
		t.Run(tt.artifactType, func(t *testing.T) {
			cache := cacheWith(tt.artifactType, tt.artifactID, tt.expected)
			resolver := NewArtifactNameResolver(cache)

			name, err := resolver.ResolveName(context.Background(), tt.artifactType, tt.artifactID)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, name)
			assert.Equal(t, tt.artifactType, cache.queriedType)
			assert.Equal(t, []string{tt.artifactID}, cache.queriedIDs)
		})
	}
}

func TestArtifactNameResolver_ReturnsDeletedArtifact(t *testing.T) {
	tests := []struct {
		name         string
		cache        *stubArtifactNameCache
		artifactType string
		artifactID   string
	}{
		{"when the artifact is not cached", cacheWith("capability", "cap-1", "Onboarding"), "capability", "nonexistent"},
		{"when the cache errors", &stubArtifactNameCache{err: errors.New("database error")}, "capability", "cap-1"},
		{"for an unknown artifact type", cacheWith("capability", "cap-1", "Onboarding"), "unknown_type", "cap-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewArtifactNameResolver(tt.cache)

			name, err := resolver.ResolveName(context.Background(), tt.artifactType, tt.artifactID)

			require.NoError(t, err)
			assert.Equal(t, "Deleted artifact", name)
		})
	}
}

func TestArtifactNameResolver_KeepsArtifactTypesApart(t *testing.T) {
	cache := &stubArtifactNameCache{names: map[string]map[string]string{
		"capability": {"shared-id": "Cap Name"},
		"component":  {"shared-id": "Comp Name"},
		"view":       {"shared-id": "View Name"},
	}}
	resolver := NewArtifactNameResolver(cache)

	for artifactType, expected := range map[string]string{"capability": "Cap Name", "component": "Comp Name", "view": "View Name"} {
		name, err := resolver.ResolveName(context.Background(), artifactType, "shared-id")
		require.NoError(t, err)
		assert.Equal(t, expected, name)
	}
	assert.Equal(t, 3, cache.queryAttempts)
}
