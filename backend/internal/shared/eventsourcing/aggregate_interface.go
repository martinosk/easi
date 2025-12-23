package domain

type EventSourcedAggregate interface {
	ID() string
	Version() int
	GetUncommittedChanges() []DomainEvent
	MarkChangesAsCommitted()
}
