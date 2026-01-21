import React, { useEffect, useState } from 'react';
import { Modal, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useCreateInternalTeam } from '../hooks/useInternalTeams';
import { createInternalTeamSchema, type CreateInternalTeamFormData } from '../../../lib/schemas';

interface CreateInternalTeamDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

const DEFAULT_VALUES: CreateInternalTeamFormData = {
  name: '',
  department: '',
  contactPerson: '',
  notes: '',
};

export const CreateInternalTeamDialog: React.FC<CreateInternalTeamDialogProps> = ({
  isOpen,
  onClose,
}) => {
  const [backendError, setBackendError] = useState<string | null>(null);
  const createMutation = useCreateInternalTeam();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isValid },
  } = useForm<CreateInternalTeamFormData>({
    resolver: zodResolver(createInternalTeamSchema),
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

  const onSubmit = async (data: CreateInternalTeamFormData) => {
    setBackendError(null);
    try {
      await createMutation.mutateAsync({
        name: data.name,
        department: data.department || undefined,
        contactPerson: data.contactPerson || undefined,
        notes: data.notes || undefined,
      });
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to create internal team');
    }
  };

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Create Internal Team"
      centered
      data-testid="create-internal-team-dialog"
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter team name (e.g., Platform Engineering)"
            {...register('name')}
            required
            withAsterisk
            autoFocus
            disabled={createMutation.isPending}
            error={errors.name?.message}
            data-testid="internal-team-name-input"
          />

          <TextInput
            label="Department"
            placeholder="Enter department (optional)"
            {...register('department')}
            disabled={createMutation.isPending}
            error={errors.department?.message}
            data-testid="internal-team-department-input"
          />

          <TextInput
            label="Contact Person"
            placeholder="Enter contact person (optional)"
            {...register('contactPerson')}
            disabled={createMutation.isPending}
            error={errors.contactPerson?.message}
            data-testid="internal-team-contact-input"
          />

          <Textarea
            label="Notes"
            placeholder="Enter notes (optional)"
            {...register('notes')}
            rows={3}
            disabled={createMutation.isPending}
            error={errors.notes?.message}
            data-testid="internal-team-notes-input"
          />

          {backendError && (
            <Alert color="red" data-testid="create-internal-team-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={createMutation.isPending}
              data-testid="create-internal-team-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={createMutation.isPending}
              disabled={!isValid}
              data-testid="create-internal-team-submit"
            >
              Create
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
