package projectors

import (
	"context"
	"fmt"

	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/onepagers/application/readmodels"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

const transferActor = "system@easi.io"

type PendingDefinitionSource interface {
	PendingTransfers(ctx context.Context) ([]readmodels.SubjectDefinition, error)
	MarkTransferred(ctx context.Context, fieldID string) error
}

type CustomFieldSchemaTransfer struct {
	tenants  TenantDirectory
	pending  PendingDefinitionSource
	commands CommandDispatcher
}

func NewCustomFieldSchemaTransfer(tenants TenantDirectory, pending PendingDefinitionSource, commandDispatcher CommandDispatcher) *CustomFieldSchemaTransfer {
	return &CustomFieldSchemaTransfer{tenants: tenants, pending: pending, commands: commandDispatcher}
}

func (t *CustomFieldSchemaTransfer) Run(ctx context.Context) error {
	tenantIDs, err := t.tenants.TenantIDs(ctx)
	if err != nil {
		return fmt.Errorf("list tenants for custom field schema transfer: %w", err)
	}
	for _, tenantID := range tenantIDs {
		if err := t.transferTenant(ctx, tenantID); err != nil {
			return err
		}
	}
	return nil
}

func (t *CustomFieldSchemaTransfer) transferTenant(ctx context.Context, tenantID string) error {
	tenant, err := sharedvo.NewTenantID(tenantID)
	if err != nil {
		return fmt.Errorf("parse tenant %s for custom field schema transfer: %w", tenantID, err)
	}
	tenantCtx := sharedctx.WithTenant(ctx, tenant)

	pending, err := t.pending.PendingTransfers(tenantCtx)
	if err != nil {
		return fmt.Errorf("list pending custom field definitions for tenant %s: %w", tenantID, err)
	}
	for _, definition := range pending {
		if err := t.transfer(tenantCtx, tenantID, definition); err != nil {
			return err
		}
	}
	return nil
}

func (t *CustomFieldSchemaTransfer) transfer(ctx context.Context, tenantID string, definition readmodels.SubjectDefinition) error {
	if _, err := t.commands.Dispatch(ctx, importCommand(tenantID, definition)); err != nil {
		return fmt.Errorf("transfer custom field %s of %s to MetaModel: %w", definition.Field.ID, definition.SubjectType, err)
	}
	if err := t.pending.MarkTransferred(ctx, definition.Field.ID); err != nil {
		return err
	}
	return nil
}

func importCommand(tenantID string, definition readmodels.SubjectDefinition) *mmPL.ImportSubjectAttribute {
	field := definition.Field
	var options []mmPL.ImportedOption
	for _, option := range field.Options {
		options = append(options, mmPL.ImportedOption{ID: option.ID, Label: option.Label, Active: option.Active})
	}
	return &mmPL.ImportSubjectAttribute{
		TenantID:    tenantID,
		SubjectType: definition.SubjectType,
		ImportedBy:  transferActor,
		Attribute: mmPL.ImportedAttribute{
			ID:       field.ID,
			Name:     field.Name,
			Type:     field.Type,
			HelpText: field.HelpText,
			Options:  options,
			Min:      field.Min,
			Max:      field.Max,
			Active:   field.Active,
		},
	}
}
