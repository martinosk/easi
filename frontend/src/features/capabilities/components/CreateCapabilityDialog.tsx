import React, { useEffect, useState, useMemo } from 'react';
import { Modal, TextInput, Textarea, Select, Button, Group, Stack, Alert } from '@mantine/core';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useCreateCapability, useUpdateCapabilityMetadata } from '../hooks/useCapabilities';
import { useStatuses } from '../../../hooks/useMetadata';
import { useMaturityScale } from '../../../hooks/useMaturityScale';
import { MaturitySlider } from '../../../components/shared/MaturitySlider';
import { createCapabilitySchema, type CreateCapabilityFormData } from '../../../lib/schemas';
import { getMaturityBounds, getDefaultSections } from '../../../utils/maturity';

interface CreateCapabilityDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

const DEFAULT_STATUSES = [
  { value: 'Active', displayName: 'Active', sortOrder: 1 },
  { value: 'Planned', displayName: 'Planned', sortOrder: 2 },
  { value: 'Deprecated', displayName: 'Deprecated', sortOrder: 3 },
];

const DEFAULT_VALUES: CreateCapabilityFormData = {
  name: '',
  description: '',
  status: 'Active',
  maturityValue: 12,
};

function useCreateCapabilityForm(isOpen: boolean, onClose: () => void) {
  const [backendError, setBackendError] = useState<string | null>(null);

  const { data: statusesData, isLoading: isLoadingStatuses } = useStatuses();
  const { data: maturityScale } = useMaturityScale();
  const createCapabilityMutation = useCreateCapability();
  const updateMetadataMutation = useUpdateCapabilityMetadata();

  const sections = maturityScale?.sections ?? getDefaultSections();
  const statuses = statusesData ?? DEFAULT_STATUSES;
  const isCreating = createCapabilityMutation.isPending || updateMetadataMutation.isPending;

  const maturityBounds = useMemo(() => getMaturityBounds(sections), [sections]);
  const schema = useMemo(() => createCapabilitySchema(maturityBounds), [maturityBounds]);

  const {
    register,
    handleSubmit,
    control,
    reset,
    formState: { errors, isValid },
  } = useForm<CreateCapabilityFormData>({
    resolver: zodResolver(schema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useEffect(() => {
    if (isOpen) {
      reset(DEFAULT_VALUES);
      setBackendError(null);
    }
  }, [isOpen, reset]);

  const statusOptions = [...statuses]
    .sort((a, b) => a.sortOrder - b.sortOrder)
    .map((s) => ({ value: s.value, label: s.displayName }));

  const onSubmit = async (data: CreateCapabilityFormData) => {
    setBackendError(null);
    try {
      const capability = await createCapabilityMutation.mutateAsync({
        name: data.name,
        description: data.description || undefined,
        level: 'L1',
      });
      await updateMetadataMutation.mutateAsync({
        id: capability.id,
        request: {
          status: data.status,
          maturityValue: data.maturityValue,
        },
      });
      onClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to create capability');
    }
  };

  return {
    register, handleSubmit, control, errors, isValid,
    backendError, isCreating, isLoadingStatuses,
    statusOptions, onSubmit,
  };
}

export const CreateCapabilityDialog: React.FC<CreateCapabilityDialogProps> = ({
  isOpen,
  onClose,
}) => {
  const {
    register, handleSubmit, control, errors, isValid,
    backendError, isCreating, isLoadingStatuses,
    statusOptions, onSubmit,
  } = useCreateCapabilityForm(isOpen, onClose);

  return (
    <Modal
      opened={isOpen}
      onClose={onClose}
      title="Create Capability"
      centered
      data-testid="create-capability-dialog"
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter capability name"
            {...register('name')}
            required
            withAsterisk
            autoFocus
            disabled={isCreating}
            error={errors.name?.message}
            data-testid="capability-name-input"
          />

          <Textarea
            label="Description"
            placeholder="Enter capability description (optional)"
            {...register('description')}
            rows={3}
            disabled={isCreating}
            error={errors.description?.message}
            data-testid="capability-description-input"
          />

          <Controller
            name="status"
            control={control}
            render={({ field }) => (
              <Select
                label="Status"
                data={isLoadingStatuses ? [] : statusOptions}
                disabled={isCreating || isLoadingStatuses}
                data-testid="capability-status-select"
                {...field}
              />
            )}
          />

          <Controller
            name="maturityValue"
            control={control}
            render={({ field }) => (
              <MaturitySlider
                value={field.value}
                onChange={field.onChange}
                disabled={isCreating}
              />
            )}
          />

          {backendError && (
            <Alert color="red" data-testid="create-capability-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={onClose}
              disabled={isCreating}
              data-testid="create-capability-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={isCreating}
              disabled={isLoadingStatuses || !isValid}
              data-testid="create-capability-submit"
            >
              Create
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
