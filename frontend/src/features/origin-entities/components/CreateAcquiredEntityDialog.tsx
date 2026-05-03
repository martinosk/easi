import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Select, Stack, Textarea, TextInput } from '@mantine/core';
import React, { useLayoutEffect, useState } from 'react';
import {
  Controller,
  type Control,
  type FieldErrors,
  type UseFormRegister,
  useForm,
} from 'react-hook-form';
import type { IntegrationStatus } from '../../../api/types';
import { type CreateAcquiredEntityFormData, createAcquiredEntitySchema } from '../../../lib/schemas';
import { useCreateAcquiredEntity } from '../hooks/useAcquiredEntities';

interface CreatedAcquiredEntity {
  id: string;
  name: string;
}

interface CreateAcquiredEntityDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onCreated?: (entity: CreatedAcquiredEntity) => void | Promise<void>;
}

const INTEGRATION_STATUS_OPTIONS = [
  { value: 'NotStarted', label: 'Not Started' },
  { value: 'InProgress', label: 'In Progress' },
  { value: 'Completed', label: 'Completed' },
];

const integrationStatusToApi: Record<string, IntegrationStatus> = {
  NotStarted: 'NOT_STARTED',
  InProgress: 'IN_PROGRESS',
  Completed: 'COMPLETED',
};

const DEFAULT_VALUES: CreateAcquiredEntityFormData = {
  name: '',
  acquisitionDate: '',
  integrationStatus: 'NotStarted',
  notes: '',
};

interface FormFieldsProps {
  register: UseFormRegister<CreateAcquiredEntityFormData>;
  control: Control<CreateAcquiredEntityFormData>;
  errors: FieldErrors<CreateAcquiredEntityFormData>;
  isPending: boolean;
}

function AcquiredEntityFields({ register, control, errors, isPending }: FormFieldsProps) {
  return (
    <>
      <TextInput
        label="Name"
        placeholder="Enter entity name (e.g., TechCorp)"
        {...register('name')}
        required
        withAsterisk
        autoFocus
        disabled={isPending}
        error={errors.name?.message}
        data-testid="acquired-entity-name-input"
      />

      <TextInput
        label="Acquisition Date"
        type="date"
        {...register('acquisitionDate')}
        disabled={isPending}
        error={errors.acquisitionDate?.message}
        data-testid="acquired-entity-date-input"
      />

      <Controller
        name="integrationStatus"
        control={control}
        render={({ field }) => (
          <Select
            label="Integration Status"
            data={INTEGRATION_STATUS_OPTIONS}
            {...field}
            disabled={isPending}
            error={errors.integrationStatus?.message}
            data-testid="acquired-entity-status-select"
          />
        )}
      />

      <Textarea
        label="Notes"
        placeholder="Enter notes (optional)"
        {...register('notes')}
        rows={3}
        disabled={isPending}
        error={errors.notes?.message}
        data-testid="acquired-entity-notes-input"
      />
    </>
  );
}

interface FormActionsProps {
  isPending: boolean;
  isValid: boolean;
  onCancel: () => void;
}

function FormActions({ isPending, isValid, onCancel }: FormActionsProps) {
  return (
    <Group justify="flex-end" gap="sm">
      <Button variant="default" onClick={onCancel} disabled={isPending} data-testid="create-acquired-entity-cancel">
        Cancel
      </Button>
      <Button
        type="submit"
        loading={isPending}
        disabled={!isValid}
        data-testid="create-acquired-entity-submit"
      >
        Create
      </Button>
    </Group>
  );
}

function useAcquiredEntityFormState(
  isOpen: boolean,
  onClose: () => void,
  onCreated?: (entity: CreatedAcquiredEntity) => void | Promise<void>,
) {
  const [backendError, setBackendError] = useState<string | null>(null);
  const createMutation = useCreateAcquiredEntity();
  const form = useForm<CreateAcquiredEntityFormData>({
    resolver: zodResolver(createAcquiredEntitySchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useLayoutEffect(() => {
    if (!isOpen) return;
    form.reset(DEFAULT_VALUES);
    if (backendError !== null) queueMicrotask(() => setBackendError(null));
  }, [isOpen, form, backendError]);

  const onSubmit = async (data: CreateAcquiredEntityFormData) => {
    setBackendError(null);
    try {
      const apiStatus = data.integrationStatus ? integrationStatusToApi[data.integrationStatus] : undefined;
      const created = await createMutation.mutateAsync({
        name: data.name,
        acquisitionDate: data.acquisitionDate || undefined,
        integrationStatus: apiStatus,
        notes: data.notes || undefined,
      });
      if (onCreated) await onCreated(created);
      onClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to create acquired entity');
    }
  };

  return { form, onSubmit, backendError, isPending: createMutation.isPending };
}

export const CreateAcquiredEntityDialog: React.FC<CreateAcquiredEntityDialogProps> = ({
  isOpen,
  onClose,
  onCreated,
}) => {
  const { form, onSubmit, backendError, isPending } = useAcquiredEntityFormState(isOpen, onClose, onCreated);
  const {
    register,
    handleSubmit,
    control,
    formState: { errors, isValid },
  } = form;

  return (
    <Modal
      opened={isOpen}
      onClose={onClose}
      title="Create Acquired Entity"
      centered
      data-testid="create-acquired-entity-dialog"
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <AcquiredEntityFields register={register} control={control} errors={errors} isPending={isPending} />

          {backendError && (
            <Alert color="red" data-testid="create-acquired-entity-error">
              {backendError}
            </Alert>
          )}

          <FormActions isPending={isPending} isValid={isValid} onCancel={onClose} />
        </Stack>
      </form>
    </Modal>
  );
};
