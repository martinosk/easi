import React, { useEffect, useState } from 'react';
import { Modal, Select, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useForm, Controller, type Control, type FieldErrors, type UseFormRegister } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useComponents } from '../../components/hooks/useComponents';
import { useCreateRelation } from '../hooks/useRelations';
import { createRelationSchema, type CreateRelationFormData } from '../../../lib/schemas';
import { toComponentId } from '../../../api/types';

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

function getDefaultValues(initialSource?: string, initialTarget?: string): CreateRelationFormData {
  return {
    sourceComponentId: initialSource || '',
    targetComponentId: initialTarget || '',
    relationType: 'Triggers',
    name: '',
    description: '',
  };
}

function getTargetError(fieldError?: string, hasSameSourceAndTarget?: boolean): string | undefined {
  return fieldError || (hasSameSourceAndTarget ? 'Source and target components must be different' : undefined);
}

interface ComponentSelectFieldsProps {
  control: Control<CreateRelationFormData>;
  errors: FieldErrors<CreateRelationFormData>;
  componentOptions: { value: string; label: string }[];
  isPending: boolean;
  initialSource?: string;
  initialTarget?: string;
  hasSameSourceAndTarget: boolean;
}

function ComponentSelectFields({
  control,
  errors,
  componentOptions,
  isPending,
  initialSource,
  initialTarget,
  hasSameSourceAndTarget,
}: ComponentSelectFieldsProps) {
  return (
    <>
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
            label="Target Component"
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
  control: Control<CreateRelationFormData>;
  register: UseFormRegister<CreateRelationFormData>;
  errors: FieldErrors<CreateRelationFormData>;
  isPending: boolean;
}

function RelationDetailFields({ control, register, errors, isPending }: RelationDetailFieldsProps) {
  return (
    <>
      <Controller
        name="relationType"
        control={control}
        render={({ field }) => (
          <Select
            label="Relation Type"
            data={RELATION_TYPE_OPTIONS}
            required
            withAsterisk
            disabled={isPending}
            data-testid="relation-type-select"
            {...field}
          />
        )}
      />

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

function FormActions({ isPending, isValid, hasSameSourceAndTarget, onCancel }: {
  isPending: boolean;
  isValid: boolean;
  hasSameSourceAndTarget: boolean;
  onCancel: () => void;
}) {
  return (
    <Group justify="flex-end" gap="sm">
      <Button
        variant="default"
        onClick={onCancel}
        disabled={isPending}
        data-testid="create-relation-cancel"
      >
        Cancel
      </Button>
      <Button
        type="submit"
        loading={isPending}
        disabled={!isValid || hasSameSourceAndTarget}
        data-testid="create-relation-submit"
      >
        Create Relation
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
  const [backendError, setBackendError] = useState<string | null>(null);

  const { data: components = [] } = useComponents();
  const createRelationMutation = useCreateRelation();

  const {
    register,
    handleSubmit,
    control,
    reset,
    watch,
    formState: { errors, isValid },
  } = useForm<CreateRelationFormData>({
    resolver: zodResolver(createRelationSchema),
    defaultValues: getDefaultValues(initialSource, initialTarget),
    mode: 'onChange',
  });

  useEffect(() => {
    if (isOpen) {
      reset(getDefaultValues(initialSource, initialTarget));
      setBackendError(null);
    }
  }, [isOpen, initialSource, initialTarget, reset]);

  const onSubmit = async (data: CreateRelationFormData) => {
    setBackendError(null);
    try {
      await createRelationMutation.mutateAsync({
        sourceComponentId: toComponentId(data.sourceComponentId),
        targetComponentId: toComponentId(data.targetComponentId),
        relationType: data.relationType,
        name: data.name || undefined,
        description: data.description || undefined,
      });
      onClose();
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
      onClose={onClose}
      title="Create Relation"
      centered
      data-testid="create-relation-dialog"
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <ComponentSelectFields
            control={control}
            errors={errors}
            componentOptions={componentOptions}
            isPending={createRelationMutation.isPending}
            initialSource={initialSource}
            initialTarget={initialTarget}
            hasSameSourceAndTarget={hasSameSourceAndTarget}
          />

          <RelationDetailFields
            control={control}
            register={register}
            errors={errors}
            isPending={createRelationMutation.isPending}
          />

          {backendError && (
            <Alert color="red" data-testid="create-relation-error">
              {backendError}
            </Alert>
          )}

          <FormActions
            isPending={createRelationMutation.isPending}
            isValid={isValid}
            hasSameSourceAndTarget={hasSameSourceAndTarget}
            onCancel={onClose}
          />
        </Stack>
      </form>
    </Modal>
  );
};
