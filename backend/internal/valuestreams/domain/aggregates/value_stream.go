package aggregates

import (
	"errors"
	"time"

	"easi/backend/internal/valuestreams/domain/entities"
	"easi/backend/internal/valuestreams/domain/events"
	"easi/backend/internal/valuestreams/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrStageNameExists        = errors.New("stage with this name already exists in this value stream")
	ErrStageNotFound          = errors.New("stage not found")
	ErrCapabilityAlreadyMapped = errors.New("capability is already mapped to this stage")
	ErrCapabilityNotMapped    = errors.New("capability is not mapped to this stage")
	ErrInvalidStagePositions  = errors.New("invalid stage positions: must include all stages with contiguous positions starting at 1")
)

type StagePositionUpdate struct {
	StageID  string
	Position int
}

type ValueStream struct {
	domain.AggregateRoot
	name        valueobjects.ValueStreamName
	description valueobjects.Description
	createdAt   time.Time
	stages      []entities.Stage
}

func NewValueStream(
	name valueobjects.ValueStreamName,
	description valueobjects.Description,
) (*ValueStream, error) {
	id := valueobjects.NewValueStreamID()
	aggregate := &ValueStream{
		AggregateRoot: domain.NewAggregateRootWithID(id.Value()),
	}

	event := events.NewValueStreamCreated(
		aggregate.ID(),
		name.Value(),
		description.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadValueStreamFromHistory(events []domain.DomainEvent) (*ValueStream, error) {
	aggregate := &ValueStream{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (v *ValueStream) Update(name valueobjects.ValueStreamName, description valueobjects.Description) error {
	event := events.NewValueStreamUpdated(
		v.ID(),
		name.Value(),
		description.Value(),
	)

	v.apply(event)
	v.RaiseEvent(event)

	return nil
}

func (v *ValueStream) Delete() error {
	event := events.NewValueStreamDeleted(v.ID())

	v.apply(event)
	v.RaiseEvent(event)

	return nil
}

func (v *ValueStream) AddStage(name valueobjects.StageName, description valueobjects.Description, position *valueobjects.StagePosition) (valueobjects.StageID, error) {
	if v.stageNameExists(name, "") {
		return valueobjects.StageID{}, ErrStageNameExists
	}

	stageID := valueobjects.NewStageID()

	var pos int
	if position != nil {
		pos = position.Value()
	} else {
		pos = len(v.stages) + 1
	}

	event := events.NewValueStreamStageAdded(
		v.ID(), stageID.Value(), name.Value(), description.Value(), pos,
	)
	v.apply(event)
	v.RaiseEvent(event)

	return stageID, nil
}

func (v *ValueStream) UpdateStage(stageID valueobjects.StageID, name valueobjects.StageName, description valueobjects.Description) error {
	if _, idx := v.findStage(stageID); idx < 0 {
		return ErrStageNotFound
	}
	if v.stageNameExists(name, stageID.Value()) {
		return ErrStageNameExists
	}

	event := events.NewValueStreamStageUpdated(v.ID(), stageID.Value(), name.Value(), description.Value())
	v.apply(event)
	v.RaiseEvent(event)
	return nil
}

func (v *ValueStream) RemoveStage(stageID valueobjects.StageID) error {
	if _, idx := v.findStage(stageID); idx < 0 {
		return ErrStageNotFound
	}

	event := events.NewValueStreamStageRemoved(v.ID(), stageID.Value())
	v.apply(event)
	v.RaiseEvent(event)
	return nil
}

func (v *ValueStream) ReorderStages(positions []StagePositionUpdate) error {
	if err := v.validateReorderPositions(positions); err != nil {
		return err
	}

	entries := make([]events.StagePositionEntry, len(positions))
	for i, p := range positions {
		entries[i] = events.StagePositionEntry{StageID: p.StageID, Position: p.Position}
	}

	event := events.NewValueStreamStagesReordered(v.ID(), entries)
	v.apply(event)
	v.RaiseEvent(event)
	return nil
}

func (v *ValueStream) validateReorderPositions(positions []StagePositionUpdate) error {
	if len(positions) != len(v.stages) {
		return ErrInvalidStagePositions
	}

	positionSet := make(map[int]bool, len(positions))
	stageSet := make(map[string]bool, len(positions))
	for _, p := range positions {
		if p.Position <= 0 || p.Position > len(v.stages) {
			return ErrInvalidStagePositions
		}
		if positionSet[p.Position] || stageSet[p.StageID] {
			return ErrInvalidStagePositions
		}
		positionSet[p.Position] = true
		stageSet[p.StageID] = true

		if _, idx := v.findStageByStringID(p.StageID); idx < 0 {
			return ErrInvalidStagePositions
		}
	}
	return nil
}

func (v *ValueStream) AddCapabilityToStage(stageID valueobjects.StageID, capRef valueobjects.CapabilityRef, capabilityName string) error {
	stage, idx := v.findStage(stageID)
	if idx < 0 {
		return ErrStageNotFound
	}
	if stage.HasCapability(capRef) {
		return ErrCapabilityAlreadyMapped
	}

	event := events.NewValueStreamStageCapabilityAdded(v.ID(), stageID.Value(), capRef.Value(), capabilityName)
	v.apply(event)
	v.RaiseEvent(event)
	return nil
}

func (v *ValueStream) RemoveCapabilityFromStage(stageID valueobjects.StageID, capRef valueobjects.CapabilityRef) error {
	stage, idx := v.findStage(stageID)
	if idx < 0 {
		return ErrStageNotFound
	}
	if !stage.HasCapability(capRef) {
		return ErrCapabilityNotMapped
	}

	event := events.NewValueStreamStageCapabilityRemoved(v.ID(), stageID.Value(), capRef.Value())
	v.apply(event)
	v.RaiseEvent(event)
	return nil
}

func (v *ValueStream) Stages() []entities.Stage {
	result := make([]entities.Stage, len(v.stages))
	copy(result, v.stages)
	return result
}

func (v *ValueStream) StageCount() int {
	return len(v.stages)
}

func (v *ValueStream) findStage(stageID valueobjects.StageID) (entities.Stage, int) {
	for i, s := range v.stages {
		if s.ID().Equals(stageID) {
			return s, i
		}
	}
	return entities.Stage{}, -1
}

func (v *ValueStream) stageNameExists(name valueobjects.StageName, excludeStageID string) bool {
	for _, s := range v.stages {
		if s.Name().Value() == name.Value() && s.ID().Value() != excludeStageID {
			return true
		}
	}
	return false
}

func (v *ValueStream) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.ValueStreamCreated:
		v.applyCreated(e)
	case events.ValueStreamUpdated:
		v.name, _ = valueobjects.NewValueStreamName(e.Name)
		v.description = valueobjects.MustNewDescription(e.Description)
	default:
		v.applyStageEvent(event)
	}
}

func (v *ValueStream) applyStageEvent(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.ValueStreamStageAdded:
		v.applyStageAdded(e)
	case events.ValueStreamStageUpdated:
		v.applyStageUpdated(e)
	case events.ValueStreamStageRemoved:
		v.applyStageRemoved(e)
	case events.ValueStreamStagesReordered:
		v.applyStagesReordered(e)
	case events.ValueStreamStageCapabilityAdded:
		v.updateStageByID(e.StageID, func(s entities.Stage) entities.Stage {
			ref, _ := valueobjects.NewCapabilityRef(e.CapabilityID)
			return s.WithAddedCapability(ref)
		})
	case events.ValueStreamStageCapabilityRemoved:
		v.updateStageByID(e.StageID, func(s entities.Stage) entities.Stage {
			ref, _ := valueobjects.NewCapabilityRef(e.CapabilityID)
			return s.WithRemovedCapability(ref)
		})
	}
}

func (v *ValueStream) applyCreated(e events.ValueStreamCreated) {
	v.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	v.name, _ = valueobjects.NewValueStreamName(e.Name)
	v.description = valueobjects.MustNewDescription(e.Description)
	v.createdAt = e.CreatedAt
	v.stages = []entities.Stage{}
}

func (v *ValueStream) applyStageAdded(e events.ValueStreamStageAdded) {
	stageID, _ := valueobjects.NewStageIDFromString(e.StageID)
	stageName, _ := valueobjects.NewStageName(e.Name)
	desc := valueobjects.MustNewDescription(e.Description)
	pos, _ := valueobjects.NewStagePosition(e.Position)
	stage := entities.NewStage(stageID, stageName, desc, pos)

	v.shiftStagePositions(e.Position, 1)
	v.stages = append(v.stages, stage)
}

func (v *ValueStream) applyStageUpdated(e events.ValueStreamStageUpdated) {
	v.updateStageByID(e.StageID, func(s entities.Stage) entities.Stage {
		name, _ := valueobjects.NewStageName(e.Name)
		desc := valueobjects.MustNewDescription(e.Description)
		return s.WithName(name).WithDescription(desc)
	})
}

func (v *ValueStream) applyStageRemoved(e events.ValueStreamStageRemoved) {
	removedPos := 0
	newStages := make([]entities.Stage, 0, len(v.stages)-1)
	for _, s := range v.stages {
		if s.ID().Value() == e.StageID {
			removedPos = s.Position().Value()
			continue
		}
		newStages = append(newStages, s)
	}
	v.stages = newStages
	v.shiftStagePositions(removedPos+1, -1)
}

func (v *ValueStream) applyStagesReordered(e events.ValueStreamStagesReordered) {
	for _, p := range e.Positions {
		v.updateStageByID(p.StageID, func(s entities.Stage) entities.Stage {
			newPos, _ := valueobjects.NewStagePosition(p.Position)
			return s.WithPosition(newPos)
		})
	}
}

func (v *ValueStream) updateStageByID(stageID string, fn func(entities.Stage) entities.Stage) {
	if _, idx := v.findStageByStringID(stageID); idx >= 0 {
		v.stages[idx] = fn(v.stages[idx])
	}
}

func (v *ValueStream) findStageByStringID(stageID string) (entities.Stage, int) {
	for i, s := range v.stages {
		if s.ID().Value() == stageID {
			return s, i
		}
	}
	return entities.Stage{}, -1
}

func (v *ValueStream) shiftStagePositions(fromPosition int, delta int) {
	for i := range v.stages {
		if v.stages[i].Position().Value() >= fromPosition {
			newPos, _ := valueobjects.NewStagePosition(v.stages[i].Position().Value() + delta)
			v.stages[i] = v.stages[i].WithPosition(newPos)
		}
	}
}

func (v *ValueStream) Name() valueobjects.ValueStreamName {
	return v.name
}

func (v *ValueStream) Description() valueobjects.Description {
	return v.description
}

func (v *ValueStream) CreatedAt() time.Time {
	return v.createdAt
}
