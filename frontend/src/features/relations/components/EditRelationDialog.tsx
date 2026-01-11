import React, { useEffect, useState } from 'react';
import { Modal, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useUpdateRelation } from '../hooks/useRelations';
import { editRelationSchema, type EditRelationFormData } from '../../../lib/schemas';
import type { Relation } from '../../../api/types';

interface EditRelationDialogProps {
  isOpen: boolean;
  onClose: () => void;
  relation: Relation | null;
}

const DEFAULT_VALUES: EditRelationFormData = { name: '', description: '' };

export const EditRelationDialog: React.FC<EditRelationDialogProps> = ({
  isOpen,
  onClose,
  relation,
}) => {
  const [backendError, setBackendError] = useState<string | null>(null);
  const updateRelationMutation = useUpdateRelation();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<EditRelationFormData>({
    resolver: zodResolver(editRelationSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useEffect(() => {
    if (isOpen && relation) {
      reset({
        name: relation.name || '',
        description: relation.description || '',
      });
      setBackendError(null);
    }
  }, [isOpen, relation, reset]);

  const handleClose = () => {
    onClose();
  };

  const onSubmit = async (data: EditRelationFormData) => {
    if (!relation) return;
    setBackendError(null);
    try {
      await updateRelationMutation.mutateAsync({
        relation,
        request: {
          name: data.name || undefined,
          description: data.description || undefined,
        },
      });
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to update relation');
    }
  };

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Edit Relation"
      centered
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter relation name (optional)"
            {...register('name')}
            disabled={updateRelationMutation.isPending}
            autoFocus
            error={errors.name?.message}
          />

          <Textarea
            label="Description"
            placeholder="Enter relation description (optional)"
            {...register('description')}
            rows={3}
            disabled={updateRelationMutation.isPending}
            error={errors.description?.message}
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
              disabled={updateRelationMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={updateRelationMutation.isPending}
            >
              Save Changes
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
