package adapters

import "context"

type NoOpComponentDeleter struct{}

func NewNoOpComponentDeleter() *NoOpComponentDeleter {
	return &NoOpComponentDeleter{}
}

func (d *NoOpComponentDeleter) DeleteComponent(_ context.Context, _ string) error {
	return nil
}
