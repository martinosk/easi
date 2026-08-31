package services

import "context"

type ArtifactNameResolver interface {
	ResolveName(ctx context.Context, artifactType, artifactID string) (string, error)
	ResolveNames(ctx context.Context, artifactType string, artifactIDs []string) (map[string]string, error)
}
