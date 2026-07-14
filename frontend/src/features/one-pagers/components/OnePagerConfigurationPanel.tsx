import { Alert, Divider, Loader, Stack } from '@mantine/core';
import { useState } from 'react';
import { hasLink } from '../../../utils/hateoas';
import { useDefineCustomFieldFlow } from '../hooks/useDefineCustomFieldFlow';
import { useOnePagerConfiguration } from '../hooks/useOnePagerConfiguration';
import { useOnePagerFieldActions } from '../hooks/useOnePagerFieldActions';
import type { BuiltInField, CustomField, OnePagerConfiguration, OnePagerSubjectType } from '../types';
import { AddCustomFieldForm } from './AddCustomFieldForm';
import { BuiltInFieldsCatalog } from './BuiltInFieldsCatalog';
import { FieldList } from './FieldList';
import { ImpactPreviewDialog } from './ImpactPreviewDialog';
import { RenameFieldDialog } from './RenameFieldDialog';
import { RetiredFieldsList } from './RetiredFieldsList';

interface OnePagerConfigurationPanelProps {
  subjectType: OnePagerSubjectType;
}

interface RequireFieldDialogsProps {
  configuration: OnePagerConfiguration;
  requireConfirmationField: CustomField | null;
  isConfirmingRequired: boolean;
  onConfirmField: (field: CustomField) => void;
  onCancelField: () => void;
  requireConfirmationBuiltIn: BuiltInField | null;
  isConfirmingBuiltInRequired: boolean;
  onConfirmBuiltIn: (field: BuiltInField) => void;
  onCancelBuiltIn: () => void;
  pendingNewFieldName: string | undefined;
  isSavingNewField: boolean;
  onConfirmNewField: () => void;
  onCancelNewField: () => void;
}

function RequireFieldDialogs({
  configuration,
  requireConfirmationField,
  isConfirmingRequired,
  onConfirmField,
  onCancelField,
  requireConfirmationBuiltIn,
  isConfirmingBuiltInRequired,
  onConfirmBuiltIn,
  onCancelBuiltIn,
  pendingNewFieldName,
  isSavingNewField,
  onConfirmNewField,
  onCancelNewField,
}: RequireFieldDialogsProps) {
  return (
    <>
      {requireConfirmationField && (
        <ImpactPreviewDialog
          key={requireConfirmationField.id}
          configuration={configuration}
          fieldName={requireConfirmationField.name}
          fieldId={requireConfirmationField.id}
          isConfirming={isConfirmingRequired}
          onConfirm={() => onConfirmField(requireConfirmationField)}
          onCancel={onCancelField}
        />
      )}

      {requireConfirmationBuiltIn && (
        <ImpactPreviewDialog
          key={requireConfirmationBuiltIn.id}
          configuration={configuration}
          fieldName={requireConfirmationBuiltIn.label}
          fieldId={requireConfirmationBuiltIn.id}
          fieldKind="builtIn"
          isConfirming={isConfirmingBuiltInRequired}
          onConfirm={() => onConfirmBuiltIn(requireConfirmationBuiltIn)}
          onCancel={onCancelBuiltIn}
        />
      )}

      {pendingNewFieldName !== undefined && (
        <ImpactPreviewDialog
          configuration={configuration}
          fieldName={pendingNewFieldName}
          isConfirming={isSavingNewField}
          onConfirm={onConfirmNewField}
          onCancel={onCancelNewField}
        />
      )}
    </>
  );
}

export function OnePagerConfigurationPanel({ subjectType }: OnePagerConfigurationPanelProps) {
  const { data: configuration, isLoading, error } = useOnePagerConfiguration(subjectType);
  const [renamingField, setRenamingField] = useState<CustomField | null>(null);
  const [requireConfirmationField, setRequireConfirmationField] = useState<CustomField | null>(null);
  const [requireConfirmationBuiltIn, setRequireConfirmationBuiltIn] = useState<BuiltInField | null>(null);
  const {
    fieldActions,
    includeField,
    reactivateField,
    saveRename,
    isRenaming,
    confirmRequireField,
    isConfirmingRequired,
    confirmRequireBuiltIn,
    isConfirmingBuiltInRequired,
  } = useOnePagerFieldActions(
    subjectType,
    configuration,
    setRenamingField,
    setRequireConfirmationField,
    setRequireConfirmationBuiltIn,
  );
  const defineFieldFlow = useDefineCustomFieldFlow(subjectType, configuration);

  if (isLoading) return <Loader data-testid="one-pager-loading" />;
  if (error || !configuration) {
    return (
      <Alert color="red" data-testid="one-pager-error">
        {error instanceof Error ? error.message : 'Failed to load the one-pager configuration'}
      </Alert>
    );
  }

  return (
    <Stack gap="lg" data-testid={`one-pager-panel-${subjectType}`}>
      <FieldList configuration={configuration} actions={fieldActions} />

      <BuiltInFieldsCatalog fields={configuration.builtInFields} onInclude={includeField} />
      <RetiredFieldsList fields={configuration.customFields} onReactivate={reactivateField} />

      {hasLink(configuration, 'x-define-custom-field') && (
        <>
          <Divider label="Add a custom field" labelPosition="left" />
          <AddCustomFieldForm
            key={defineFieldFlow.formKey}
            isSaving={defineFieldFlow.isSaving}
            onSubmit={defineFieldFlow.handleSubmit}
          />
        </>
      )}

      {renamingField && (
        <RenameFieldDialog
          key={renamingField.id}
          field={renamingField}
          isSaving={isRenaming}
          onSave={(field, data) => saveRename(field, data, () => setRenamingField(null))}
          onClose={() => setRenamingField(null)}
        />
      )}

      <RequireFieldDialogs
        configuration={configuration}
        requireConfirmationField={requireConfirmationField}
        isConfirmingRequired={isConfirmingRequired}
        onConfirmField={(field) => confirmRequireField(field, () => setRequireConfirmationField(null))}
        onCancelField={() => setRequireConfirmationField(null)}
        requireConfirmationBuiltIn={requireConfirmationBuiltIn}
        isConfirmingBuiltInRequired={isConfirmingBuiltInRequired}
        onConfirmBuiltIn={(field) => confirmRequireBuiltIn(field, () => setRequireConfirmationBuiltIn(null))}
        onCancelBuiltIn={() => setRequireConfirmationBuiltIn(null)}
        pendingNewFieldName={defineFieldFlow.pendingNewField?.name}
        isSavingNewField={defineFieldFlow.isSaving}
        onConfirmNewField={defineFieldFlow.confirmPendingField}
        onCancelNewField={defineFieldFlow.cancelPendingField}
      />
    </Stack>
  );
}
