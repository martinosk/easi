package context

import (
	"context"
)

const ActorContextKey contextKey = "actor"

type Actor struct {
	ID    string
	Email string
}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, ActorContextKey, actor)
}

func GetActor(ctx context.Context) (Actor, bool) {
	actor, ok := ctx.Value(ActorContextKey).(Actor)
	return actor, ok
}
