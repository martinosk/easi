package aggregates

import (
	"errors"
	"time"

	"easi/backend/internal/capabilitymapping/domain/entities"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrL1CannotHaveParent            = errors.New("L1 capabilities cannot have a parent")
	ErrNonL1MustHaveParent           = errors.New("L2-L4 capabilities must have a parent")
	ErrParentMustBeOneLevelAbove     = errors.New("parent must be exactly one level above")
	ErrCapabilityCannotBeOwnParent   = errors.New("capability cannot be its own parent")
	ErrWouldCreateCircularReference  = errors.New("operation would create circular reference")
	ErrWouldExceedMaximumDepth       = errors.New("operation would create L5+ hierarchy")
	ErrOnlyL1CanBeAssignedToDomain   = errors.New("only L1 capabilities can be assigned to business domains")
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
	experts        []*entities.Expert
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

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadCapabilityFromHistory(events []domain.DomainEvent) (*Capability, error) {
	aggregate := &Capability{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (c *Capability) Update(name valueobjects.CapabilityName, description valueobjects.Description) error {
	event := events.NewCapabilityUpdated(
		c.ID(),
		name.Value(),
		description.Value(),
	)

	c.apply(event)
	c.RaiseEvent(event)

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

	c.apply(event)
	c.RaiseEvent(event)

	return nil
}

func (c *Capability) AddExpert(expert *entities.Expert) error {
	event := events.NewCapabilityExpertAdded(
		c.ID(),
		expert.Name(),
		expert.Role(),
		expert.Contact(),
	)

	c.apply(event)
	c.RaiseEvent(event)

	return nil
}

func (c *Capability) AddTag(tag valueobjects.Tag) error {
	for _, existingTag := range c.tags {
		if existingTag.Value() == tag.Value() {
			return nil
		}
	}

	event := events.NewCapabilityTagAdded(c.ID(), tag.Value())

	c.apply(event)
	c.RaiseEvent(event)

	return nil
}

func (c *Capability) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.CapabilityCreated:
		c.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		c.name, _ = valueobjects.NewCapabilityName(e.Name)
		c.description = valueobjects.MustNewDescription(e.Description)
		if e.ParentID != "" {
			c.parentID, _ = valueobjects.NewCapabilityIDFromString(e.ParentID)
		}
		c.level, _ = valueobjects.NewCapabilityLevel(e.Level)
		c.createdAt = e.CreatedAt
		c.status = valueobjects.StatusActive
		c.maturityLevel = valueobjects.MaturityGenesis
	case events.CapabilityUpdated:
		c.name, _ = valueobjects.NewCapabilityName(e.Name)
		c.description = valueobjects.MustNewDescription(e.Description)
	case events.CapabilityMetadataUpdated:
		c.maturityLevel, _ = valueobjects.NewMaturityLevelFromValue(e.MaturityValue)
		c.ownershipModel, _ = valueobjects.NewOwnershipModel(e.OwnershipModel)
		c.primaryOwner = valueobjects.NewOwner(e.PrimaryOwner)
		c.eaOwner = valueobjects.NewOwner(e.EAOwner)
		c.status, _ = valueobjects.NewCapabilityStatus(e.Status)
	case events.CapabilityExpertAdded:
		expert, _ := entities.NewExpert(e.ExpertName, e.ExpertRole, e.ContactInfo)
		c.experts = append(c.experts, expert)
	case events.CapabilityTagAdded:
		tag, _ := valueobjects.NewTag(e.Tag)
		c.tags = append(c.tags, tag)
	case events.CapabilityParentChanged:
		if e.NewParentID != "" {
			c.parentID, _ = valueobjects.NewCapabilityIDFromString(e.NewParentID)
		} else {
			c.parentID = valueobjects.CapabilityID{}
		}
		c.level, _ = valueobjects.NewCapabilityLevel(e.NewLevel)
	case events.CapabilityDeleted:
	}
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

func (c *Capability) Experts() []*entities.Expert {
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

	c.apply(event)
	c.RaiseEvent(event)

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

	c.apply(event)
	c.RaiseEvent(event)

	return nil
}

func validateHierarchy(parentID valueobjects.CapabilityID, level valueobjects.CapabilityLevel) error {
	hasParent := parentID.Value() != ""

	if level == valueobjects.LevelL1 && hasParent {
		return ErrL1CannotHaveParent
	}

	return nil
}
