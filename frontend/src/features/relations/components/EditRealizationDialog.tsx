import React, { useEffect, useState } from 'react';
import { Modal, Select, Textarea, Button, Group, Stack, Alert, Text, Box } from '@mantine/core';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useCapabilities, useUpdateRealization } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { editRealizationSchema, type EditRealizationFormData } from '../../../lib/schemas';
import type { CapabilityRealization, RealizationLevel } from '../../../api/types';

interface EditRealizationDialogProps {
  isOpen: boolean;
  onClose: () => void;
  realization: CapabilityRealization | null;
}

const REALIZATION_LEVEL_OPTIONS = [
  { value: 'Full', label: 'Full (100%)' },
  { value: 'Partial', label: 'Partial' },
  { value: 'Planned', label: 'Planned' },
];

const DEFAULT_VALUES: EditRealizationFormData = {
  realizationLevel: 'Full',
  notes: '',
};

export const EditRealizationDialog: React.FC<EditRealizationDialogProps> = ({
  isOpen,
  onClose,
  realization,
}) => {
  const [backendError, setBackendError] = useState<string | null>(null);

  const updateRealizationMutation = useUpdateRealization();
  const { data: capabilities = [] } = useCapabilities();
  const { data: components = [] } = useComponents();

  const capability = realization
    ? capabilities.find((c) => c.id === realization.capabilityId)
    : null;
  const component = realization
    ? components.find((c) => c.id === realization.componentId)
    : null;

  const {
    register,
    handleSubmit,
    control,
    reset,
    formState: { errors },
  } = useForm<EditRealizationFormData>({
    resolver: zodResolver(editRealizationSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useEffect(() => {
    if (isOpen && realization) {
      reset({
        realizationLevel: realization.realizationLevel,
        notes: realization.notes || '',
      });
      setBackendError(null);
    }
  }, [isOpen, realization, reset]);

  const handleClose = () => {
    onClose();
  };

  const onSubmit = async (data: EditRealizationFormData) => {
    if (!realization) return;
    setBackendError(null);
    try {
      await updateRealizationMutation.mutateAsync({
        id: realization.id,
        capabilityId: realization.capabilityId,
        componentId: realization.componentId,
        request: {
          realizationLevel: data.realizationLevel as RealizationLevel,
          notes: data.notes || undefined,
        },
      });
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to update realization');
    }
  };

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Edit Realization"
      centered
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <Box>
            <Text size="sm" fw={500} mb={4}>
              Capability
            </Text>
            <Text size="sm" c="dimmed">
              {capability?.name || 'Unknown'}
            </Text>
          </Box>

          <Box>
            <Text size="sm" fw={500} mb={4}>
              Application
            </Text>
            <Text size="sm" c="dimmed">
              {component?.name || 'Unknown'}
            </Text>
          </Box>

          <Controller
            name="realizationLevel"
            control={control}
            render={({ field }) => (
              <Select
                label="Realization Level"
                data={REALIZATION_LEVEL_OPTIONS}
                disabled={updateRealizationMutation.isPending}
                error={errors.realizationLevel?.message}
                {...field}
              />
            )}
          />

          <Textarea
            label="Notes"
            placeholder="Enter notes (optional)"
            {...register('notes')}
            rows={4}
            disabled={updateRealizationMutation.isPending}
            error={errors.notes?.message}
          />

          {backendError && (
            <Alert color="red">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={updateRealizationMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={updateRealizationMutation.isPending}
            >
              Update Realization
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
