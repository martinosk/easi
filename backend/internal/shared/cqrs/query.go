package cqrs

import "context"

// Query represents a request for data
type Query interface {
	// QueryName returns the name of the query
	QueryName() string
}

// QueryHandler handles a specific query type and returns a result
type QueryHandler interface {
	// Handle processes the query and returns the result
	Handle(ctx context.Context, query Query) (interface{}, error)
}

// QueryBus dispatches queries to their handlers
type QueryBus interface {
	// Dispatch sends a query to its handler and returns the result
	Dispatch(ctx context.Context, query Query) (interface{}, error)

	// Register registers a handler for a query type
	Register(queryName string, handler QueryHandler)
}
