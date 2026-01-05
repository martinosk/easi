package architecturemodeling

import (
	"context"
)

type ComponentDTO struct {
	ID   string
	Name string
}

type ComponentGateway interface {
	GetByID(ctx context.Context, id string) (*ComponentDTO, error)
}
