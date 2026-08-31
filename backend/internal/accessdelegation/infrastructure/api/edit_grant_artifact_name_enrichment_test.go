package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"easi/backend/internal/accessdelegation/application/readmodels"
)

type countingNameResolver struct {
	callsByType map[string]int
	namesByType map[string]map[string]string
}

func newCountingNameResolver(namesByType map[string]map[string]string) *countingNameResolver {
	return &countingNameResolver{
		callsByType: make(map[string]int),
		namesByType: namesByType,
	}
}

func (r *countingNameResolver) ResolveName(ctx context.Context, artifactType, artifactID string) (string, error) {
	names, err := r.ResolveNames(ctx, artifactType, []string{artifactID})
	if err != nil {
		return "", err
	}
	return names[artifactID], nil
}

func (r *countingNameResolver) ResolveNames(_ context.Context, artifactType string, artifactIDs []string) (map[string]string, error) {
	r.callsByType[artifactType]++
	cached := r.namesByType[artifactType]
	resolved := make(map[string]string, len(artifactIDs))
	for _, id := range artifactIDs {
		if name := cached[id]; name != "" {
			resolved[id] = name
			continue
		}
		resolved[id] = "Deleted artifact"
	}
	return resolved, nil
}

func TestEnrichGrantsWithArtifactNames_BatchesOneQueryPerArtifactType(t *testing.T) {
	resolver := newCountingNameResolver(map[string]map[string]string{
		"capability": {"cap-1": "Customer Onboarding", "cap-2": "Billing"},
		"component":  {"comp-1": "Payment Service"},
	})
	handlers := NewEditGrantHandlers(EditGrantHandlerDeps{NameResolver: resolver})

	grants := []readmodels.EditGrantDTO{
		{ArtifactType: "capability", ArtifactID: "cap-1"},
		{ArtifactType: "capability", ArtifactID: "cap-2"},
		{ArtifactType: "capability", ArtifactID: "cap-gone"},
		{ArtifactType: "component", ArtifactID: "comp-1"},
	}

	handlers.enrichGrantsWithArtifactNames(context.Background(), grants)

	assert.Equal(t, 1, resolver.callsByType["capability"])
	assert.Equal(t, 1, resolver.callsByType["component"])
	assert.Equal(t, "Customer Onboarding", grants[0].ArtifactName)
	assert.Equal(t, "Billing", grants[1].ArtifactName)
	assert.Equal(t, "Deleted artifact", grants[2].ArtifactName)
	assert.Equal(t, "Payment Service", grants[3].ArtifactName)
}

func TestEnrichGrantsWithArtifactNames_DedupesRepeatedArtifactWithinType(t *testing.T) {
	resolver := newCountingNameResolver(map[string]map[string]string{
		"capability": {"cap-1": "Customer Onboarding"},
	})
	handlers := NewEditGrantHandlers(EditGrantHandlerDeps{NameResolver: resolver})

	grants := []readmodels.EditGrantDTO{
		{ArtifactType: "capability", ArtifactID: "cap-1", GranteeEmail: "a@dfds.com"},
		{ArtifactType: "capability", ArtifactID: "cap-1", GranteeEmail: "b@dfds.com"},
	}

	handlers.enrichGrantsWithArtifactNames(context.Background(), grants)

	assert.Equal(t, 1, resolver.callsByType["capability"])
	assert.Equal(t, "Customer Onboarding", grants[0].ArtifactName)
	assert.Equal(t, "Customer Onboarding", grants[1].ArtifactName)
}

func TestEnrichGrantsWithArtifactNames_NoResolver_LeavesNamesUnset(t *testing.T) {
	handlers := NewEditGrantHandlers(EditGrantHandlerDeps{})
	grants := []readmodels.EditGrantDTO{{ArtifactType: "capability", ArtifactID: "cap-1"}}

	handlers.enrichGrantsWithArtifactNames(context.Background(), grants)

	assert.Empty(t, grants[0].ArtifactName)
}
