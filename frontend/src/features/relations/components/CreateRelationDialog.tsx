import { Alert, Button, Group, Modal, Select, Stack, Textarea, TextInput } from '@mantine/core';
import type React from 'react';
import { type Control, Controller, type FieldErrors, type UseFormRegister } from 'react-hook-form';
import type { ConnectComponentsFormData, ConnectionKind } from '../../../lib/schemas';
import { CONTAINMENT_KIND_DESCRIPTIONS } from '../../components/utils/containment';
import { useConnectionDialog } from '../hooks/useConnectionDialog';
import { connectionKindOptions, isContainmentKind } from '../utils/connectionKinds';

interface CreateRelationDialogProps {
  isOpen: boolean;
  onClose: () => void;
  sourceComponentId?: string;
  targetComponentId?: string;
}

function getTargetError(fieldError?: string, hasSameSourceAndTarget?: boolean): string | undefined {
  return fieldError || (hasSameSourceAndTarget ? 'Source and target components must be different' : undefined);
}

interface ComponentSelectFieldsProps {
  control: Control<ConnectComponentsFormData>;
  errors: FieldErrors<ConnectComponentsFormData>;
  componentOptions: { value: string; label: string }[];
  isPending: boolean;
  initialSource?: string;
  initialTarget?: string;
  hasSameSourceAndTarget: boolean;
  isContainment: boolean;
}

function ComponentSelectFields({
  control,
  errors,
  componentOptions,
  isPending,
  initialSource,
  initialTarget,
  hasSameSourceAndTarget,
  isContainment,
}: ComponentSelectFieldsProps) {
  return (
    <>
      <Controller
        name="sourceComponentId"
        control={control}
        render={({ field }) => (
          <Select
            label={isContainment ? 'Part' : 'Source Component'}
            placeholder="Select source component"
            data={componentOptions}
            required
            withAsterisk
            disabled={isPending || !!initialSource}
            error={errors.sourceComponentId?.message}
            data-testid="relation-source-select"
            searchable
            {...field}
          />
        )}
      />

      <Controller
        name="targetComponentId"
        control={control}
        render={({ field }) => (
          <Select
            label={isContainment ? 'Parent' : 'Target Component'}
            placeholder="Select target component"
            data={componentOptions}
            required
            withAsterisk
            disabled={isPending || !!initialTarget}
            error={getTargetError(errors.targetComponentId?.message, hasSameSourceAndTarget)}
            data-testid="relation-target-select"
            searchable
            {...field}
          />
        )}
      />
    </>
  );
}

interface RelationDetailFieldsProps {
  register: UseFormRegister<ConnectComponentsFormData>;
  errors: FieldErrors<ConnectComponentsFormData>;
  isPending: boolean;
}

function RelationDetailFields({ register, errors, isPending }: RelationDetailFieldsProps) {
  return (
    <>
      <TextInput
        label="Name"
        placeholder="Enter relation name (optional)"
        {...register('name')}
        disabled={isPending}
        error={errors.name?.message}
        data-testid="relation-name-input"
      />

      <Textarea
        label="Description"
        placeholder="Enter relation description (optional)"
        {...register('description')}
        rows={3}
        disabled={isPending}
        error={errors.description?.message}
        data-testid="relation-description-input"
      />
    </>
  );
}

interface ConnectionKindFieldProps {
  control: Control<ConnectComponentsFormData>;
  isPending: boolean;
  connectionKind: ConnectionKind;
  containmentAllowed: boolean;
}

function ConnectionKindField({ control, isPending, connectionKind, containmentAllowed }: ConnectionKindFieldProps) {
  const description = isContainmentKind(connectionKind) ? CONTAINMENT_KIND_DESCRIPTIONS[connectionKind] : undefined;

  return (
    <Controller
      name="connectionKind"
      control={control}
      render={({ field }) => (
        <Select
          label="Relation Type"
          description={description}
          data={connectionKindOptions(containmentAllowed)}
          required
          withAsterisk
          allowDeselect={false}
          disabled={isPending}
          data-testid="relation-type-select"
          {...field}
        />
      )}
    />
  );
}

function FormActions({
  isPending,
  isValid,
  hasSameSourceAndTarget,
  isContainment,
  onCancel,
}: {
  isPending: boolean;
  isValid: boolean;
  hasSameSourceAndTarget: boolean;
  isContainment: boolean;
  onCancel: () => void;
}) {
  return (
    <Group justify="flex-end" gap="sm">
      <Button variant="default" onClick={onCancel} disabled={isPending} data-testid="create-relation-cancel">
        Cancel
      </Button>
      <Button
        type="submit"
        loading={isPending}
        disabled={!isValid || hasSameSourceAndTarget}
        data-testid="create-relation-submit"
      >
        {isContainment ? 'Attach' : 'Create Relation'}
      </Button>
    </Group>
  );
}

export const CreateRelationDialog: React.FC<CreateRelationDialogProps> = ({
  isOpen,
  onClose,
  sourceComponentId: initialSource,
  targetComponentId: initialTarget,
}) => {
  const dialog = useConnectionDialog({ isOpen, onClose, initialSource, initialTarget });
  const {
    register,
    handleSubmit,
    control,
    formState: { errors, isValid },
  } = dialog.form;

  return (
    <Modal opened={isOpen} onClose={onClose} title="Create Relation" centered data-testid="create-relation-dialog">
      <form onSubmit={handleSubmit(dialog.onSubmit)}>
        <Stack gap="md">
          <ComponentSelectFields
            control={control}
            errors={errors}
            componentOptions={dialog.componentOptions}
            isPending={dialog.isPending}
            initialSource={initialSource}
            initialTarget={initialTarget}
            hasSameSourceAndTarget={dialog.hasSameSourceAndTarget}
            isContainment={dialog.isContainment}
          />

          <ConnectionKindField
            control={control}
            isPending={dialog.isPending}
            connectionKind={dialog.connectionKind}
            containmentAllowed={dialog.containmentAllowed}
          />

          {!dialog.isContainment && <RelationDetailFields register={register} errors={errors} isPending={dialog.isPending} />}

          {dialog.backendError && (
            <Alert color="red" data-testid="create-relation-error">
              {dialog.backendError}
            </Alert>
          )}

          <FormActions
            isPending={dialog.isPending}
            isValid={isValid}
            hasSameSourceAndTarget={dialog.hasSameSourceAndTarget}
            isContainment={dialog.isContainment}
            onCancel={onClose}
          />
        </Stack>
      </form>
    </Modal>
  );
};
