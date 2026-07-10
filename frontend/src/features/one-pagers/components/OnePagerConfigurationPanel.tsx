import { Alert, Divider, Loader, Stack } from '@mantine/core';
import { useState } from 'react';
import type { DefineCustomFieldFormData } from '../../../lib/schemas/onePagerConfiguration';
import { hasLink } from '../../../utils/hateoas';
import { useOnePagerConfiguration } from '../hooks/useOnePagerConfiguration';
import { useOnePagerFieldActions } from '../hooks/useOnePagerFieldActions';
import { useDefineCustomField } from '../hooks/useOnePagerMutations';
import type { CustomField, DefineCustomFieldRequest, OnePagerSubjectType } from '../types';
import { AddCustomFieldForm } from './AddCustomFieldForm';
import { BuiltInFieldsCatalog } from './BuiltInFieldsCatalog';
import { FieldList } from './FieldList';
import { RenameFieldDialog } from './RenameFieldDialog';
import { RetiredFieldsList } from './RetiredFieldsList';

interface OnePagerConfigurationPanelProps {
  subjectType: OnePagerSubjectType;
}

export function OnePagerConfigurationPanel({ subjectType }: OnePagerConfigurationPanelProps) {
  const { data: configuration, isLoading, error } = useOnePagerConfiguration(subjectType);
  const [renamingField, setRenamingField] = useState<CustomField | null>(null);
  const { fieldActions, includeField, reactivateField, saveRename, isRenaming } = useOnePagerFieldActions(
    subjectType,
    configuration,
    setRenamingField,
  );
  const defineField = useDefineCustomField(subjectType);

  if (isLoading) return <Loader data-testid="one-pager-loading" />;
  if (error || !configuration) {
    return (
      <Alert color="red" data-testid="one-pager-error">
        {error instanceof Error ? error.message : 'Failed to load the one-pager configuration'}
      </Alert>
    );
  }

  const handleDefineField = (data: DefineCustomFieldFormData) => {
    const request: DefineCustomFieldRequest = {
      name: data.name,
      fieldType: data.fieldType,
      required: data.required,
      helpText: data.helpText,
      options: data.fieldType === 'selection' ? data.options : undefined,
      version: configuration.version,
    };
    defineField.mutate({ configuration, request });
  };

  return (
    <Stack gap="lg" data-testid={`one-pager-panel-${subjectType}`}>
      <FieldList configuration={configuration} actions={fieldActions} />

      <BuiltInFieldsCatalog fields={configuration.builtInFields} onInclude={includeField} />
      <RetiredFieldsList fields={configuration.customFields} onReactivate={reactivateField} />

      {hasLink(configuration, 'x-define-custom-field') ? (
        <>
          <Divider label="Add a custom field" labelPosition="left" />
          <AddCustomFieldForm isSaving={defineField.isPending} onSubmit={handleDefineField} />
        </>
      ) : null}

      {renamingField && (
        <RenameFieldDialog
          key={renamingField.id}
          field={renamingField}
          isSaving={isRenaming}
          onSave={(field, data) => saveRename(field, data, () => setRenamingField(null))}
          onClose={() => setRenamingField(null)}
        />
      )}
    </Stack>
  );
}
