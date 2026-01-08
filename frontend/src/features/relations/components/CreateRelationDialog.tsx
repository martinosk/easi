import React, { useEffect, useState } from 'react';
import { Modal, Select, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useComponents } from '../../components/hooks/useComponents';
import { useCreateRelation } from '../hooks/useRelations';
import { createRelationSchema, type CreateRelationFormData } from '../../../lib/schemas';
import type { ComponentId } from '../../../api/types';

interface CreateRelationDialogProps {
  isOpen: boolean;
  onClose: () => void;
  sourceComponentId?: string;
  targetComponentId?: string;
}

const RELATION_TYPE_OPTIONS = [
  { value: 'Triggers', label: 'Triggers' },
  { value: 'Serves', label: 'Serves' },
];

export const CreateRelationDialog: React.FC<CreateRelationDialogProps> = ({
  isOpen,
  onClose,
  sourceComponentId: initialSource,
  targetComponentId: initialTarget,
}) => {
  const [backendError, setBackendError] = useState<string | null>(null);

  const { data: components = [] } = useComponents();
  const createRelationMutation = useCreateRelation();

  const createDefaultValues = (): CreateRelationFormData => ({
    sourceComponentId: initialSource || '',
    targetComponentId: initialTarget || '',
    relationType: 'Triggers',
    name: '',
    description: '',
  });

  const {
    register,
    handleSubmit,
    control,
    reset,
    watch,
    formState: { errors, isValid },
  } = useForm<CreateRelationFormData>({
    resolver: zodResolver(createRelationSchema),
    defaultValues: createDefaultValues(),
    mode: 'onChange',
  });

  useEffect(() => {
    if (isOpen) {
      reset(createDefaultValues());
      setBackendError(null);
    }
  }, [isOpen, initialSource, initialTarget, reset]);

  const handleClose = () => {
    onClose();
  };

  const onSubmit = async (data: CreateRelationFormData) => {
    setBackendError(null);
    try {
      await createRelationMutation.mutateAsync({
        sourceComponentId: data.sourceComponentId as ComponentId,
        targetComponentId: data.targetComponentId as ComponentId,
        relationType: data.relationType,
        name: data.name || undefined,
        description: data.description || undefined,
      });
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to create relation');
    }
  };

  const componentOptions = components.map((c) => ({
    value: c.id,
    label: c.name,
  }));

  const sourceComponentId = watch('sourceComponentId');
  const targetComponentId = watch('targetComponentId');
  const hasSameSourceAndTarget = Boolean(
    sourceComponentId && targetComponentId && sourceComponentId === targetComponentId
  );

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Create Relation"
      centered
      data-testid="create-relation-dialog"
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <Controller
            name="sourceComponentId"
            control={control}
            render={({ field }) => (
              <Select
                label="Source Component"
                placeholder="Select source component"
                data={componentOptions}
                required
                withAsterisk
                disabled={createRelationMutation.isPending || !!initialSource}
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
                label="Target Component"
                placeholder="Select target component"
                data={componentOptions}
                required
                withAsterisk
                disabled={createRelationMutation.isPending || !!initialTarget}
                error={
                  errors.targetComponentId?.message ||
                  (hasSameSourceAndTarget ? 'Source and target components must be different' : undefined)
                }
                data-testid="relation-target-select"
                searchable
                {...field}
              />
            )}
          />

          <Controller
            name="relationType"
            control={control}
            render={({ field }) => (
              <Select
                label="Relation Type"
                data={RELATION_TYPE_OPTIONS}
                required
                withAsterisk
                disabled={createRelationMutation.isPending}
                data-testid="relation-type-select"
                {...field}
              />
            )}
          />

          <TextInput
            label="Name"
            placeholder="Enter relation name (optional)"
            {...register('name')}
            disabled={createRelationMutation.isPending}
            error={errors.name?.message}
            data-testid="relation-name-input"
          />

          <Textarea
            label="Description"
            placeholder="Enter relation description (optional)"
            {...register('description')}
            rows={3}
            disabled={createRelationMutation.isPending}
            error={errors.description?.message}
            data-testid="relation-description-input"
          />

          {backendError && (
            <Alert color="red" data-testid="create-relation-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={createRelationMutation.isPending}
              data-testid="create-relation-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={createRelationMutation.isPending}
              disabled={!isValid || hasSameSourceAndTarget}
              data-testid="create-relation-submit"
            >
              Create Relation
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
