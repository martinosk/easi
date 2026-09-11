import { Alert, Divider, Loader, Stack } from '@mantine/core';
import { useMemo, useState } from 'react';
import type { SubjectAttribute, SubjectAttributeSchema } from '../../../api/types';
import { useSubjectAttributeSchema } from '../../../hooks/useSubjectAttributeSchema';
import { hasLink } from '../../../utils/hateoas';
import { useAttributeSchemaActions } from '../hooks/useAttributeSchemaActions';
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

interface AddCustomFieldSectionProps {
  schema: SubjectAttributeSchema | undefined;
  flow: ReturnType<typeof useDefineCustomFieldFlow>;
}

function AddCustomFieldSection({ schema, flow }: AddCustomFieldSectionProps) {
  if (!hasLink(schema, 'x-define')) return null;
  return (
    <>
      <Divider label="Add a custom field" labelPosition="left" />
      <AddCustomFieldForm key={flow.formKey} isSaving={flow.isSaving} onSubmit={flow.handleSubmit} />
    </>
  );
}

function loadFailureMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Failed to load the one-pager configuration';
}

export function OnePagerConfigurationPanel({ subjectType }: OnePagerConfigurationPanelProps) {
  const { data: configuration, isLoading, error } = useOnePagerConfiguration(subjectType);
  const { data: schema } = useSubjectAttributeSchema(subjectType, hasLink(configuration, 'x-attribute-schema'));
  const [renamingAttribute, setRenamingAttribute] = useState<SubjectAttribute | null>(null);
  const [requireConfirmationField, setRequireConfirmationField] = useState<CustomField | null>(null);
  const [requireConfirmationBuiltIn, setRequireConfirmationBuiltIn] = useState<BuiltInField | null>(null);
  const {
    fieldActions,
    includeField,
    confirmRequireField,
    isConfirmingRequired,
    confirmRequireBuiltIn,
    isConfirmingBuiltInRequired,
  } = useOnePagerFieldActions(subjectType, configuration, setRequireConfirmationField, setRequireConfirmationBuiltIn);
  const schemaActions = useAttributeSchemaActions(subjectType, schema, setRenamingAttribute);
  const defineFieldFlow = useDefineCustomFieldFlow(subjectType, configuration, schema);
  const attributesById = useMemo(
    () => new Map((schema?.attributes ?? []).map((attribute) => [attribute.id, attribute])),
    [schema],
  );

  if (isLoading) return <Loader data-testid="one-pager-loading" />;
  if (error || !configuration) {
    return (
      <Alert color="red" data-testid="one-pager-error">
        {loadFailureMessage(error)}
      </Alert>
    );
  }

  return (
    <Stack gap="lg" data-testid={`one-pager-panel-${subjectType}`}>
      <FieldList
        configuration={configuration}
        attributesById={attributesById}
        actions={{ presentation: fieldActions, schema: schemaActions.actions }}
      />

      <BuiltInFieldsCatalog fields={configuration.builtInFields} onInclude={includeField} />
      <RetiredFieldsList attributes={schema?.attributes ?? []} onReactivate={schemaActions.reactivateAttribute} />

      <AddCustomFieldSection schema={schema} flow={defineFieldFlow} />

      {renamingAttribute && (
        <RenameFieldDialog
          key={renamingAttribute.id}
          attribute={renamingAttribute}
          isSaving={schemaActions.isRenaming}
          onSave={(attribute, data) => schemaActions.saveRename(attribute, data, () => setRenamingAttribute(null))}
          onClose={() => setRenamingAttribute(null)}
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
