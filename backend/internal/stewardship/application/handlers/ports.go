package handlers

import (
	"context"
	"errors"
	"fmt"

	"easi/backend/internal/stewardship/domain/aggregates"
	"easi/backend/internal/stewardship/domain/valueobjects"
)

var ErrDomainNotFound = errors.New("business domain not found")

type StewardshipRepository interface {
	Save(ctx context.Context, s *aggregates.Stewardship) error
	GetByID(ctx context.Context, id string) (*aggregates.Stewardship, error)
}

type LiveStewardshipLookup interface {
	FindLiveStewardshipID(ctx context.Context, domainID, concern string) (string, bool, error)
}

type DomainDirectory interface {
	DomainExists(ctx context.Context, domainID string) (bool, error)
}

type UserDirectory interface {
	UserStanding(ctx context.Context, userID string) (valueobjects.UserStanding, error)
}

func concernOfExistingDomain(ctx context.Context, domains DomainDirectory, domainID, concernValue string) (valueobjects.Concern, error) {
	concern, err := valueobjects.NewConcern(concernValue)
	if err != nil {
		return valueobjects.Concern{}, err
	}
	exists, err := domains.DomainExists(ctx, domainID)
	if err != nil {
		return valueobjects.Concern{}, fmt.Errorf("check business domain %s exists: %w", domainID, err)
	}
	if !exists {
		return valueobjects.Concern{}, fmt.Errorf("%w: %s", ErrDomainNotFound, domainID)
	}
	return concern, nil
}

type liveStewardships struct {
	lookup     LiveStewardshipLookup
	repository StewardshipRepository
}

func (l liveStewardships) find(ctx context.Context, domainID string, concern valueobjects.Concern) (*aggregates.Stewardship, error) {
	id, found, err := l.lookup.FindLiveStewardshipID(ctx, domainID, concern.Value())
	if err != nil {
		return nil, fmt.Errorf("find %s stewardship of domain %s: %w", concern, domainID, err)
	}
	if !found {
		return nil, nil
	}
	return l.repository.GetByID(ctx, id)
}
