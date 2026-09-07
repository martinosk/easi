package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	authPL "easi/backend/internal/auth/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type UserNameCacheWriter interface {
	Upsert(ctx context.Context, id, name, email string) error
}

type UserNameCacheProjector struct {
	cache UserNameCacheWriter
}

func NewUserNameCacheProjector(cache UserNameCacheWriter) *UserNameCacheProjector {
	return &UserNameCacheProjector{cache: cache}
}

func (p *UserNameCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *UserNameCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	if eventType != authPL.UserCreated {
		return nil
	}
	var event struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal UserCreated event data: %w", err)
	}
	if event.ID == "" {
		return nil
	}
	if err := p.cache.Upsert(ctx, event.ID, event.Name, event.Email); err != nil {
		return fmt.Errorf("project UserCreated cache upsert for user %s: %w", event.ID, err)
	}
	return nil
}
