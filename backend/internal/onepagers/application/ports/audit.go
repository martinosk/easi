package ports

import (
	"context"
	"time"
)

type SubjectAudit struct {
	ActorID    string
	ActorEmail string
	CreatedAt  time.Time
	Found      bool
}

type SubjectAuditReader interface {
	Created(ctx context.Context, aggregateID string) (SubjectAudit, error)
}
