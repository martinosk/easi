package aggregates

import (
	"errors"
	"fmt"
	"time"

	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrL1CannotHaveParent           = errors.New("L1 capabilities cannot have a parent")
	ErrNonL1MustHaveParent          = errors.New("L2-L4 capabilities must have a parent")
	ErrParentMustBeOneLevelAbove    = errors.New("parent must be exactly one level above")
	ErrCapabilityCannotBeOwnParent  = errors.New("capability cannot be its own parent")
	ErrWouldCreateCircularReference = errors.New("operation would create circular reference")
	ErrWouldExceedMaximumDepth      = errors.New("operation would create L5+ hierarchy")
	ErrOnlyL1CanBeAssignedToDomain  = errors.New("only L1 capabilities can be assigned to business domains")
)

type Capability struct {
	domain.AggregateRoot
	name           valueobjects.CapabilityName
	description    valueobjects.Description
	parentID       valueobjects.CapabilityID
	level          valueobjects.CapabilityLevel
	createdAt      time.Time
	maturityLevel  valueobjects.MaturityLevel
	ownershipModel valueobjects.OwnershipModel
	primaryOwner   valueobjects.Owner
	eaOwner        valueobjects.Owner
	status         valueobjects.CapabilityStatus
	experts        []valueobjects.Expert
	tags           []valueobjects.Tag
}

func NewCapability(
	name valueobjects.CapabilityName,
	description valueobjects.Description,
	parentID valueobjects.CapabilityID,
	level valueobjects.CapabilityLevel,
) (*Capability, error) {
	if err := validateHierarchy(parentID, level); err != nil {
		return nil, err
	}

	aggregate := &Capability{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewCapabilityCreated(
		aggregate.ID(),
		name.Value(),
		description.Value(),
		parentID.Value(),
		level.Value(),
	)

	aggregate.raise(event)

	return aggregate, nil
}

func LoadCapabilityFromHistory(events []domain.DomainEvent) (*Capability, error) {
	aggregate := &Capability{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	var applyErr error
	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		if applyErr != nil {
			return
		}
		applyErr = aggregate.apply(event)
	})
	if applyErr != nil {
		return nil, applyErr
	}

	return aggregate, nil
}

func (c *Capability) Update(name valueobjects.CapabilityName, description valueobjects.Description) error {
	event := events.NewCapabilityUpdated(
		c.ID(),
		name.Value(),
		description.Value(),
	)

	c.raise(event)

	return nil
}

func (c *Capability) UpdateMetadata(metadata valueobjects.CapabilityMetadata) error {
	event := events.NewCapabilityMetadataUpdated(
		c.ID(),
		"",
		0,
		metadata.MaturityLevel().Value(),
		metadata.OwnershipModel().Value(),
		metadata.PrimaryOwner().Value(),
		metadata.EAOwner().Value(),
		metadata.Status().Value(),
	)

	c.raise(event)

	return nil
}

func (c *Capability) AddExpert(expert valueobjects.Expert) error {
	event := events.NewCapabilityExpertAdded(
		c.ID(),
		expert.Name(),
		expert.Role(),
		expert.Contact(),
	)

	c.raise(event)

	return nil
}

func (c *Capability) RemoveExpert(expert valueobjects.Expert) error {
	event := events.NewCapabilityExpertRemoved(
		c.ID(),
		expert.Name(),
		expert.Role(),
		expert.Contact(),
	)

	c.raise(event)

	return nil
}

func (c *Capability) AddTag(tag valueobjects.Tag) error {
	for _, existingTag := range c.tags {
		if existingTag.Value() == tag.Value() {
			return nil
		}
	}

	event := events.NewCapabilityTagAdded(c.ID(), tag.Value())

	c.raise(event)

	return nil
}

func (c *Capability) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.CapabilityCreated:
		return c.applyCreated(e)
	case events.CapabilityUpdated:
		return c.applyUpdated(e)
	case events.CapabilityMetadataUpdated:
		return c.applyMetadataUpdated(e)
	case events.CapabilityExpertAdded:
		return c.applyExpertAdded(e)
	case events.CapabilityExpertRemoved:
		c.applyExpertRemoved(e)
	case events.CapabilityTagAdded:
		return c.applyTagAdded(e)
	case events.CapabilityParentChanged:
		return c.applyParentChanged(e)
	case events.CapabilityLevelChanged:
		return c.applyLevelChanged(e)
	}
	return nil
}

func (c *Capability) applyCreated(e events.CapabilityCreated) error {
	c.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	name, err := valueobjects.NewCapabilityName(e.Name)
	if err != nil {
		return fmt.Errorf("%w: capability name %q: %v", domain.ErrCorruptedEvent, e.Name, err)
	}
	c.name = name
	c.description = valueobjects.MustNewDescription(e.Description)
	if e.ParentID != "" {
		parentID, err := valueobjects.NewCapabilityIDFromString(e.ParentID)
		if err != nil {
			return fmt.Errorf("%w: parent ID %q: %v", domain.ErrCorruptedEvent, e.ParentID, err)
		}
		c.parentID = parentID
	}
	level, err := valueobjects.NewCapabilityLevel(e.Level)
	if err != nil {
		return fmt.Errorf("%w: capability level %q: %v", domain.ErrCorruptedEvent, e.Level, err)
	}
	c.level = level
	c.createdAt = e.CreatedAt
	c.status = valueobjects.StatusActive
	c.maturityLevel = valueobjects.MaturityGenesis
	return nil
}

func (c *Capability) applyUpdated(e events.CapabilityUpdated) error {
	name, err := valueobjects.NewCapabilityName(e.Name)
	if err != nil {
		return fmt.Errorf("%w: capability name %q: %v", domain.ErrCorruptedEvent, e.Name, err)
	}
	c.name = name
	c.description = valueobjects.MustNewDescription(e.Description)
	return nil
}

func (c *Capability) applyMetadataUpdated(e events.CapabilityMetadataUpdated) error {
	maturityLevel, err := valueobjects.NewMaturityLevelFromValue(e.MaturityValue)
	if err != nil {
		return fmt.Errorf("%w: maturity level %d: %v", domain.ErrCorruptedEvent, e.MaturityValue, err)
	}
	c.maturityLevel = maturityLevel
	ownershipModel, err := valueobjects.NewOwnershipModel(e.OwnershipModel)
	if err != nil {
		return fmt.Errorf("%w: ownership model %q: %v", domain.ErrCorruptedEvent, e.OwnershipModel, err)
	}
	c.ownershipModel = ownershipModel
	c.primaryOwner = valueobjects.NewOwner(e.PrimaryOwner)
	c.eaOwner = valueobjects.NewOwner(e.EAOwner)
	status, err := valueobjects.NewCapabilityStatus(e.Status)
	if err != nil {
		return fmt.Errorf("%w: capability status %q: %v", domain.ErrCorruptedEvent, e.Status, err)
	}
	c.status = status
	return nil
}

func (c *Capability) applyExpertAdded(e events.CapabilityExpertAdded) error {
	expert := valueobjects.MustNewExpert(e.ExpertName, e.ExpertRole, e.ContactInfo, e.AddedAt)
	c.experts = append(c.experts, expert)
	return nil
}

func (c *Capability) applyExpertRemoved(e events.CapabilityExpertRemoved) {
	c.experts = removeExpert(c.experts, e.ExpertName, e.ExpertRole, e.ContactInfo)
}

func removeExpert(experts []valueobjects.Expert, name, role, contact string) []valueobjects.Expert {
	result := make([]valueobjects.Expert, 0, len(experts))
	for _, expert := range experts {
		if !expert.MatchesValues(name, role, contact) {
			result = append(result, expert)
		}
	}
	return result
}

func (c *Capability) applyTagAdded(e events.CapabilityTagAdded) error {
	tag, err := valueobjects.NewTag(e.Tag)
	if err != nil {
		return fmt.Errorf("%w: tag %q: %v", domain.ErrCorruptedEvent, e.Tag, err)
	}
	c.tags = append(c.tags, tag)
	return nil
}

func (c *Capability) applyParentChanged(e events.CapabilityParentChanged) error {
	if e.NewParentID != "" {
		parentID, err := valueobjects.NewCapabilityIDFromString(e.NewParentID)
		if err != nil {
			return fmt.Errorf("%w: new parent ID %q: %v", domain.ErrCorruptedEvent, e.NewParentID, err)
		}
		c.parentID = parentID
	} else {
		c.parentID = valueobjects.CapabilityID{}
	}
	level, err := valueobjects.NewCapabilityLevel(e.NewLevel)
	if err != nil {
		return fmt.Errorf("%w: new level %q: %v", domain.ErrCorruptedEvent, e.NewLevel, err)
	}
	c.level = level
	return nil
}

func (c *Capability) applyLevelChanged(e events.CapabilityLevelChanged) error {
	level, err := valueobjects.NewCapabilityLevel(e.NewLevel)
	if err != nil {
		return fmt.Errorf("%w: new level %q: %v", domain.ErrCorruptedEvent, e.NewLevel, err)
	}
	c.level = level
	return nil
}

func (c *Capability) Name() valueobjects.CapabilityName {
	return c.name
}

func (c *Capability) Description() valueobjects.Description {
	return c.description
}

func (c *Capability) ParentID() valueobjects.CapabilityID {
	return c.parentID
}

func (c *Capability) Level() valueobjects.CapabilityLevel {
	return c.level
}

func (c *Capability) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Capability) MaturityLevel() valueobjects.MaturityLevel {
	return c.maturityLevel
}

func (c *Capability) OwnershipModel() valueobjects.OwnershipModel {
	return c.ownershipModel
}

func (c *Capability) PrimaryOwner() valueobjects.Owner {
	return c.primaryOwner
}

func (c *Capability) EAOwner() valueobjects.Owner {
	return c.eaOwner
}

func (c *Capability) Status() valueobjects.CapabilityStatus {
	return c.status
}

func (c *Capability) Experts() []valueobjects.Expert {
	return c.experts
}

func (c *Capability) Tags() []valueobjects.Tag {
	return c.tags
}

func (c *Capability) ChangeParent(newParentID valueobjects.CapabilityID, newLevel valueobjects.CapabilityLevel) error {
	if newParentID.Value() == c.ID() {
		return ErrCapabilityCannotBeOwnParent
	}

	if newLevel.NumericValue() > 4 {
		return ErrWouldExceedMaximumDepth
	}

	event := events.NewCapabilityParentChanged(
		c.ID(),
		c.parentID.Value(),
		newParentID.Value(),
		c.level.Value(),
		newLevel.Value(),
	)

	c.raise(event)

	return nil
}

func (c *Capability) ChangeLevel(newLevel valueobjects.CapabilityLevel) error {
	if newLevel.NumericValue() > 4 {
		return ErrWouldExceedMaximumDepth
	}

	event := events.NewCapabilityLevelChanged(
		c.ID(),
		c.level.Value(),
		newLevel.Value(),
	)

	c.raise(event)

	return nil
}

func (c *Capability) CanBeAssignedToDomain() error {
	if c.level != valueobjects.LevelL1 {
		return ErrOnlyL1CanBeAssignedToDomain
	}
	return nil
}

func (c *Capability) Delete() error {
	event := events.NewCapabilityDeleted(c.ID())

	c.raise(event)

	return nil
}

func (c *Capability) raise(event domain.DomainEvent) {
	if err := c.apply(event); err != nil {
		panic(fmt.Sprintf("capabilitymapping: in-process apply failed: %v", err))
	}
	c.RaiseEvent(event)
}

func validateHierarchy(parentID valueobjects.CapabilityID, level valueobjects.CapabilityLevel) error {
	hasParent := parentID.Value() != ""

	if level == valueobjects.LevelL1 && hasParent {
		return ErrL1CannotHaveParent
	}

	return nil
}
