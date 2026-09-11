package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ConfigurationFinder interface {
	GetBySubjectType(ctx context.Context, subjectType string) (*readmodels.ConfigurationRecord, error)
}

type CustomFieldInclusionReactor struct {
	configs  ConfigurationFinder
	commands CommandDispatcher
}

func NewCustomFieldInclusionReactor(configs ConfigurationFinder, commandDispatcher CommandDispatcher) *CustomFieldInclusionReactor {
	return &CustomFieldInclusionReactor{configs: configs, commands: commandDispatcher}
}

func CustomFieldInclusionEventTypes() []string {
	return []string{mmPL.SubjectAttributeDefined, mmPL.SubjectAttributeRetired, mmPL.SubjectAttributeReactivated}
}

func (r *CustomFieldInclusionReactor) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
	}
	return r.ProjectEvent(ctx, event.EventType(), eventData)
}

func (r *CustomFieldInclusionReactor) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case mmPL.SubjectAttributeDefined, mmPL.SubjectAttributeReactivated:
		return r.react(ctx, eventType, eventData, r.include)
	case mmPL.SubjectAttributeRetired:
		return r.react(ctx, eventType, eventData, r.exclude)
	}
	return nil
}

func (r *CustomFieldInclusionReactor) react(ctx context.Context, eventType string, eventData []byte, act func(context.Context, mmPL.SubjectAttributeEventBase) error) error {
	var base mmPL.SubjectAttributeEventBase
	if err := json.Unmarshal(eventData, &base); err != nil {
		return fmt.Errorf("unmarshal %s event base: %w", eventType, err)
	}
	return act(ctx, base)
}

func (r *CustomFieldInclusionReactor) include(ctx context.Context, attribute mmPL.SubjectAttributeEventBase) error {
	configID, err := r.ensureConfiguration(ctx, attribute)
	if err != nil {
		return err
	}
	command := &commands.IncludeCustomField{ConfigID: configID, FieldID: attribute.AttributeID, ModifiedBy: attribute.ModifiedBy}
	return r.dispatchInclusion(ctx, attribute, command, aggregates.ErrCustomFieldAlreadyIncluded)
}

func (r *CustomFieldInclusionReactor) exclude(ctx context.Context, attribute mmPL.SubjectAttributeEventBase) error {
	config, err := r.findConfiguration(ctx, attribute)
	if err != nil || config == nil {
		return err
	}
	command := &commands.ExcludeCustomField{ConfigID: config.ID, FieldID: attribute.AttributeID, ModifiedBy: attribute.ModifiedBy}
	return r.dispatchInclusion(ctx, attribute, command, aggregates.ErrCustomFieldNotIncluded)
}

func (r *CustomFieldInclusionReactor) dispatchInclusion(ctx context.Context, attribute mmPL.SubjectAttributeEventBase, command cqrs.Command, alreadyInState error) error {
	if _, err := r.commands.Dispatch(ctx, command); err != nil && !errors.Is(err, alreadyInState) {
		return fmt.Errorf("%s for custom field %s on %s one-pager: %w", command.CommandName(), attribute.AttributeID, attribute.SubjectType, err)
	}
	return nil
}

func (r *CustomFieldInclusionReactor) findConfiguration(ctx context.Context, attribute mmPL.SubjectAttributeEventBase) (*readmodels.ConfigurationRecord, error) {
	config, err := r.configs.GetBySubjectType(ctx, attribute.SubjectType)
	if err != nil {
		return nil, fmt.Errorf("look up %s one-pager configuration: %w", attribute.SubjectType, err)
	}
	return config, nil
}

func (r *CustomFieldInclusionReactor) ensureConfiguration(ctx context.Context, attribute mmPL.SubjectAttributeEventBase) (string, error) {
	config, err := r.findConfiguration(ctx, attribute)
	if err != nil {
		return "", err
	}
	if config != nil {
		return config.ID, nil
	}
	result, err := r.commands.Dispatch(ctx, &commands.CreateOnePagerConfiguration{
		TenantID:    attribute.TenantID,
		SubjectType: attribute.SubjectType,
		CreatedBy:   attribute.ModifiedBy,
	})
	if err != nil {
		return "", fmt.Errorf("create %s one-pager configuration: %w", attribute.SubjectType, err)
	}
	return result.CreatedID, nil
}
