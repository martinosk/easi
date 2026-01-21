import React, { useEffect, useState } from 'react';
import { Modal, TextInput, Textarea, Button, Group, Stack, Alert, Select } from '@mantine/core';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useUpdateAcquiredEntity } from '../hooks/useAcquiredEntities';
import { editAcquiredEntitySchema, type EditAcquiredEntityFormData } from '../../../lib/schemas';
import type { AcquiredEntity, AcquiredEntityId, IntegrationStatus } from '../../../api/types';

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

export const EditAcquiredEntityDialog: React.FC<EditAcquiredEntityDialogProps> = ({
  isOpen,
  onClose,
  entity,
}) => {
  const [backendError, setBackendError] = useState<string | null>(null);
  const updateMutation = useUpdateAcquiredEntity();

  const {
    register,
    handleSubmit,
    reset,
    control,
    formState: { errors, isValid },
  } = useForm<EditAcquiredEntityFormData>({
    resolver: zodResolver(editAcquiredEntitySchema),
    mode: 'onChange',
  });

  useEffect(() => {
    if (isOpen && entity) {
      const formStatus = entity.integrationStatus ? integrationStatusFromApi[entity.integrationStatus] : undefined;
      reset({
        name: entity.name,
        acquisitionDate: entity.acquisitionDate?.split('T')[0] || '',
        integrationStatus: formStatus as EditAcquiredEntityFormData['integrationStatus'],
        notes: entity.notes || '',
      });
      setBackendError(null);
    }
  }, [isOpen, entity, reset]);

  const handleClose = () => {
    onClose();
  };

  const onSubmit = async (data: EditAcquiredEntityFormData) => {
    if (!entity) return;

    setBackendError(null);
    try {
      const apiStatus = data.integrationStatus ? integrationStatusToApi[data.integrationStatus] : undefined;
      await updateMutation.mutateAsync({
        id: entity.id as AcquiredEntityId,
        request: {
          name: data.name,
          acquisitionDate: data.acquisitionDate || undefined,
          integrationStatus: apiStatus,
          notes: data.notes || undefined,
        },
      });
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to update acquired entity');
    }
  };

  if (!entity) return null;

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Edit Acquired Entity"
      centered
      data-testid="edit-acquired-entity-dialog"
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter entity name"
            {...register('name')}
            required
            withAsterisk
            autoFocus
            disabled={updateMutation.isPending}
            error={errors.name?.message}
            data-testid="edit-acquired-entity-name-input"
          />

          <TextInput
            label="Acquisition Date"
            type="date"
            {...register('acquisitionDate')}
            disabled={updateMutation.isPending}
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
                disabled={updateMutation.isPending}
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
            disabled={updateMutation.isPending}
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
              onClick={handleClose}
              disabled={updateMutation.isPending}
              data-testid="edit-acquired-entity-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={updateMutation.isPending}
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
