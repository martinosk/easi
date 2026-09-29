package projectors

import (
	"context"

	authPL "easi/backend/internal/auth/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/stewardship/application/readmodels"
)

type UserCacheWriter interface {
	Upsert(ctx context.Context, user readmodels.CachedUser) error
	SetActive(ctx context.Context, userID string, active bool) error
}

type UserCacheProjector struct {
	cache UserCacheWriter
}

func NewUserCacheProjector(cache UserCacheWriter) *UserCacheProjector {
	return &UserCacheProjector{cache: cache}
}

type userCreatedPayload struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

func (p *UserCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	switch event.EventType() {
	case authPL.UserCreated:
		return p.cacheCreatedUser(ctx, event)
	case authPL.UserDisabled:
		return p.setActive(ctx, event, false)
	case authPL.UserEnabled:
		return p.setActive(ctx, event, true)
	}
	return nil
}

func (p *UserCacheProjector) cacheCreatedUser(ctx context.Context, event domain.DomainEvent) error {
	payload, err := decodePayload[userCreatedPayload](event)
	if err != nil {
		return err
	}
	return p.cache.Upsert(ctx, readmodels.CachedUser{
		ID:     subjectID(event, payload.ID),
		Name:   payload.Name,
		Email:  payload.Email,
		Active: payload.Status != "disabled",
	})
}

func (p *UserCacheProjector) setActive(ctx context.Context, event domain.DomainEvent, active bool) error {
	payload, err := decodePayload[idPayload](event)
	if err != nil {
		return err
	}
	return p.cache.SetActive(ctx, subjectID(event, payload.ID), active)
}
