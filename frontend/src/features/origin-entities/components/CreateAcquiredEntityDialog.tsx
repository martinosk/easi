import React, { useEffect, useState } from 'react';
import { Modal, TextInput, Textarea, Button, Group, Stack, Alert, Select } from '@mantine/core';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useCreateAcquiredEntity } from '../hooks/useAcquiredEntities';
import { createAcquiredEntitySchema, type CreateAcquiredEntityFormData } from '../../../lib/schemas';
import type { IntegrationStatus } from '../../../api/types';

interface CreateAcquiredEntityDialogProps {
  isOpen: boolean;
  onClose: () => void;
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

export const CreateAcquiredEntityDialog: React.FC<CreateAcquiredEntityDialogProps> = ({
  isOpen,
  onClose,
}) => {
  const [backendError, setBackendError] = useState<string | null>(null);
  const createMutation = useCreateAcquiredEntity();

  const {
    register,
    handleSubmit,
    reset,
    control,
    formState: { errors, isValid },
  } = useForm<CreateAcquiredEntityFormData>({
    resolver: zodResolver(createAcquiredEntitySchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useEffect(() => {
    if (isOpen) {
      reset(DEFAULT_VALUES);
      setBackendError(null);
    }
  }, [isOpen, reset]);

  const handleClose = () => {
    onClose();
  };

  const onSubmit = async (data: CreateAcquiredEntityFormData) => {
    setBackendError(null);
    try {
      const apiStatus = data.integrationStatus ? integrationStatusToApi[data.integrationStatus] : undefined;
      await createMutation.mutateAsync({
        name: data.name,
        acquisitionDate: data.acquisitionDate || undefined,
        integrationStatus: apiStatus,
        notes: data.notes || undefined,
      });
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to create acquired entity');
    }
  };

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Create Acquired Entity"
      centered
      data-testid="create-acquired-entity-dialog"
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter entity name (e.g., TechCorp)"
            {...register('name')}
            required
            withAsterisk
            autoFocus
            disabled={createMutation.isPending}
            error={errors.name?.message}
            data-testid="acquired-entity-name-input"
          />

          <TextInput
            label="Acquisition Date"
            type="date"
            {...register('acquisitionDate')}
            disabled={createMutation.isPending}
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
                disabled={createMutation.isPending}
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
            disabled={createMutation.isPending}
            error={errors.notes?.message}
            data-testid="acquired-entity-notes-input"
          />

          {backendError && (
            <Alert color="red" data-testid="create-acquired-entity-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={createMutation.isPending}
              data-testid="create-acquired-entity-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={createMutation.isPending}
              disabled={!isValid}
              data-testid="create-acquired-entity-submit"
            >
              Create
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
