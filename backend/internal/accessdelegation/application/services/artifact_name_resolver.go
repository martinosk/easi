package services

import "context"

type ArtifactNameResolver interface {
	ResolveName(ctx context.Context, artifactType, artifactID string) (string, error)
}
