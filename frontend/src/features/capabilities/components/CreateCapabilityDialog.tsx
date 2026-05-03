import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Select, Stack, Textarea, TextInput } from '@mantine/core';
import React, { useLayoutEffect, useMemo, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { MaturitySlider } from '../../../components/shared/MaturitySlider';
import { useMaturityScale } from '../../../hooks/useMaturityScale';
import { useStatuses } from '../../../hooks/useMetadata';
import { type CreateCapabilityFormData, createCapabilitySchema } from '../../../lib/schemas';
import { getDefaultSections, getMaturityBounds } from '../../../utils/maturity';
import { useCreateCapability, useUpdateCapabilityMetadata } from '../hooks/useCapabilities';

export type CapabilityLevel = 'L1' | 'L2' | 'L3' | 'L4';

interface CreatedCapability {
  id: string;
  name: string;
  level: string;
}

interface CreateCapabilityDialogProps {
  isOpen: boolean;
  onClose: () => void;
  prefill?: { level?: CapabilityLevel };
  onCreated?: (capability: CreatedCapability) => void | Promise<void>;
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

function useStatusOptions() {
  const { data: statusesData, isLoading: isLoadingStatuses } = useStatuses();
  const statuses = statusesData ?? DEFAULT_STATUSES;
  const statusOptions = [...statuses]
    .sort((a, b) => a.sortOrder - b.sortOrder)
    .map((s) => ({ value: s.value, label: s.displayName }));
  return { statusOptions, isLoadingStatuses };
}

function useMaturitySchema() {
  const { data: maturityScale } = useMaturityScale();
  const sections = maturityScale?.sections ?? getDefaultSections();
  const maturityBounds = useMemo(() => getMaturityBounds(sections), [sections]);
  const schema = useMemo(() => createCapabilitySchema(maturityBounds), [maturityBounds]);
  return schema;
}

interface CapabilitySubmitDeps {
  prefill?: { level?: CapabilityLevel };
  onCreated?: (capability: CreatedCapability) => void | Promise<void>;
  onClose: () => void;
  setBackendError: (msg: string | null) => void;
}

function buildCapabilitySubmitter(
  createCapability: ReturnType<typeof useCreateCapability>,
  updateMetadata: ReturnType<typeof useUpdateCapabilityMetadata>,
  deps: CapabilitySubmitDeps,
) {
  return async (data: CreateCapabilityFormData) => {
    deps.setBackendError(null);
    try {
      const capability = await createCapability.mutateAsync({
        name: data.name,
        description: data.description || undefined,
        level: deps.prefill?.level ?? 'L1',
      });
      await updateMetadata.mutateAsync({
        id: capability.id,
        request: { status: data.status, maturityValue: data.maturityValue },
      });
      if (deps.onCreated) await deps.onCreated(capability);
      deps.onClose();
    } catch (err) {
      deps.setBackendError(err instanceof Error ? err.message : 'Failed to create capability');
    }
  };
}

function useCreateCapabilityForm(
  isOpen: boolean,
  onClose: () => void,
  prefill?: { level?: CapabilityLevel },
  onCreated?: (capability: CreatedCapability) => void | Promise<void>,
) {
  const [backendError, setBackendError] = useState<string | null>(null);
  const createCapability = useCreateCapability();
  const updateMetadata = useUpdateCapabilityMetadata();
  const { statusOptions, isLoadingStatuses } = useStatusOptions();
  const schema = useMaturitySchema();

  const isCreating = createCapability.isPending || updateMetadata.isPending;

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

  useLayoutEffect(() => {
    if (!isOpen) return;
    reset(DEFAULT_VALUES);
    if (backendError !== null) queueMicrotask(() => setBackendError(null));
  }, [isOpen, reset, backendError]);

  const onSubmit = buildCapabilitySubmitter(createCapability, updateMetadata, {
    prefill,
    onCreated,
    onClose,
    setBackendError,
  });

  return {
    register,
    handleSubmit,
    control,
    errors,
    isValid,
    backendError,
    isCreating,
    isLoadingStatuses,
    statusOptions,
    onSubmit,
  };
}

export const CreateCapabilityDialog: React.FC<CreateCapabilityDialogProps> = ({
  isOpen,
  onClose,
  prefill,
  onCreated,
}) => {
  const {
    register,
    handleSubmit,
    control,
    errors,
    isValid,
    backendError,
    isCreating,
    isLoadingStatuses,
    statusOptions,
    onSubmit,
  } = useCreateCapabilityForm(isOpen, onClose, prefill, onCreated);

  return (
    <Modal opened={isOpen} onClose={onClose} title="Create Capability" centered data-testid="create-capability-dialog">
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
              <MaturitySlider value={field.value} onChange={field.onChange} disabled={isCreating} />
            )}
          />

          {backendError && (
            <Alert color="red" data-testid="create-capability-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button variant="default" onClick={onClose} disabled={isCreating} data-testid="create-capability-cancel">
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
