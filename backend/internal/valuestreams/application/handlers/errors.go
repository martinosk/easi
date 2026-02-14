package handlers

import (
	"errors"

	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/valueobjects"
	"easi/backend/internal/valuestreams/infrastructure/repositories"
)

var (
	ErrValueStreamNotFound   = errors.New("value stream not found")
	ErrValueStreamNameExists = errors.New("value stream with this name already exists")
	ErrStageNotFound         = errors.New("stage not found")
	ErrStageNameExists       = errors.New("stage with this name already exists in this value stream")
	ErrCapabilityNotFound    = errors.New("capability not found")
)

func mapRepositoryError(err error) error {
	if errors.Is(err, repositories.ErrValueStreamNotFound) {
		return ErrValueStreamNotFound
	}
	return err
}

func mapStageError(err error) error {
	if errors.Is(err, aggregates.ErrStageNotFound) {
		return ErrStageNotFound
	}
	if errors.Is(err, aggregates.ErrStageNameExists) {
		return ErrStageNameExists
	}
	return err
}

func newOptionalPosition(position *int) (*valueobjects.StagePosition, error) {
	if position == nil {
		return nil, nil
	}
	pos, err := valueobjects.NewStagePosition(*position)
	if err != nil {
		return nil, err
	}
	return &pos, nil
}
