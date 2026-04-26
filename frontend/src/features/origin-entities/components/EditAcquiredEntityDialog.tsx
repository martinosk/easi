import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Select, Stack, Textarea, TextInput } from '@mantine/core';
import React, { useLayoutEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import type { AcquiredEntity, AcquiredEntityId, IntegrationStatus } from '../../../api/types';
import { type EditAcquiredEntityFormData, editAcquiredEntitySchema } from '../../../lib/schemas';
import { useUpdateAcquiredEntity } from '../hooks/useAcquiredEntities';

interface EditAcquiredEntityDialogProps {
  isOpen: boolean;
  onClose: () => void;
  entity: AcquiredEntity | null;
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

const integrationStatusFromApi: Record<IntegrationStatus, string> = {
  NOT_STARTED: 'NotStarted',
  IN_PROGRESS: 'InProgress',
  COMPLETED: 'Completed',
};

function entityToFormValues(entity: AcquiredEntity): EditAcquiredEntityFormData {
  const formStatus = entity.integrationStatus ? integrationStatusFromApi[entity.integrationStatus] : undefined;
  return {
    name: entity.name,
    acquisitionDate: entity.acquisitionDate?.split('T')[0] || '',
    integrationStatus: formStatus as EditAcquiredEntityFormData['integrationStatus'],
    notes: entity.notes || '',
  };
}

function formValuesToApiRequest(data: EditAcquiredEntityFormData) {
  const apiStatus = data.integrationStatus ? integrationStatusToApi[data.integrationStatus] : undefined;
  return {
    name: data.name,
    acquisitionDate: data.acquisitionDate || undefined,
    integrationStatus: apiStatus,
    notes: data.notes || undefined,
  };
}

function useEditAcquiredEntityForm(entity: AcquiredEntity | null, isOpen: boolean, onClose: () => void) {
  const [backendError, setBackendError] = useState<string | null>(null);
  const updateMutation = useUpdateAcquiredEntity();

  const form = useForm<EditAcquiredEntityFormData>({
    resolver: zodResolver(editAcquiredEntitySchema),
    mode: 'onChange',
  });

  useLayoutEffect(() => {
    if (isOpen && entity) {
      form.reset(entityToFormValues(entity));
      if (backendError !== null) queueMicrotask(() => setBackendError(null));
    }
  }, [isOpen, entity, form, backendError]);

  const submit = async (data: EditAcquiredEntityFormData) => {
    if (!entity) return;
    setBackendError(null);
    try {
      await updateMutation.mutateAsync({
        id: entity.id as AcquiredEntityId,
        request: formValuesToApiRequest(data),
      });
      onClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to update acquired entity');
    }
  };

  return { form, submit, backendError, isPending: updateMutation.isPending };
}

export const EditAcquiredEntityDialog: React.FC<EditAcquiredEntityDialogProps> = ({ isOpen, onClose, entity }) => {
  const { form, submit, backendError, isPending } = useEditAcquiredEntityForm(entity, isOpen, onClose);
  const {
    register,
    handleSubmit,
    control,
    formState: { errors, isValid },
  } = form;

  if (!entity) return null;

  return (
    <Modal
      opened={isOpen}
      onClose={onClose}
      title="Edit Acquired Entity"
      centered
      data-testid="edit-acquired-entity-dialog"
    >
      <form onSubmit={handleSubmit(submit)}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter entity name"
            {...register('name')}
            required
            withAsterisk
            autoFocus
            disabled={isPending}
            error={errors.name?.message}
            data-testid="edit-acquired-entity-name-input"
          />

          <TextInput
            label="Acquisition Date"
            type="date"
            {...register('acquisitionDate')}
            disabled={isPending}
            error={errors.acquisitionDate?.message}
            data-testid="edit-acquired-entity-date-input"
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
                data-testid="edit-acquired-entity-status-select"
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
            data-testid="edit-acquired-entity-notes-input"
          />

          {backendError && (
            <Alert color="red" data-testid="edit-acquired-entity-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={onClose}
              disabled={isPending}
              data-testid="edit-acquired-entity-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={isPending}
              disabled={!isValid}
              data-testid="edit-acquired-entity-submit"
            >
              Save Changes
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
