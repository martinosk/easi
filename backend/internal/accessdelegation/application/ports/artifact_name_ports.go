package ports

import "context"

type ArtifactNameLookup interface {
	GetArtifactName(ctx context.Context, artifactID string) (string, error)
}

type UserEmailLookup interface {
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type InvitationChecker interface {
	HasPendingByEmail(ctx context.Context, email string) (bool, error)
}

type DomainAllowlistChecker interface {
	IsDomainAllowed(ctx context.Context, email string) (bool, error)
}
