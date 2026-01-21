import React, { useEffect, useState } from 'react';
import { Modal, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useUpdateInternalTeam } from '../hooks/useInternalTeams';
import { editInternalTeamSchema, type EditInternalTeamFormData } from '../../../lib/schemas';
import type { InternalTeam, InternalTeamId } from '../../../api/types';

interface EditInternalTeamDialogProps {
  isOpen: boolean;
  onClose: () => void;
  team: InternalTeam | null;
}

export const EditInternalTeamDialog: React.FC<EditInternalTeamDialogProps> = ({
  isOpen,
  onClose,
  team,
}) => {
  const [backendError, setBackendError] = useState<string | null>(null);
  const updateMutation = useUpdateInternalTeam();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isValid },
  } = useForm<EditInternalTeamFormData>({
    resolver: zodResolver(editInternalTeamSchema),
    mode: 'onChange',
  });

  useEffect(() => {
    if (isOpen && team) {
      reset({
        name: team.name,
        department: team.department || '',
        contactPerson: team.contactPerson || '',
        notes: team.notes || '',
      });
      setBackendError(null);
    }
  }, [isOpen, team, reset]);

  const handleClose = () => {
    onClose();
  };

  const onSubmit = async (data: EditInternalTeamFormData) => {
    if (!team) return;

    setBackendError(null);
    try {
      await updateMutation.mutateAsync({
        id: team.id as InternalTeamId,
        request: {
          name: data.name,
          department: data.department || undefined,
          contactPerson: data.contactPerson || undefined,
          notes: data.notes || undefined,
        },
      });
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to update internal team');
    }
  };

  if (!team) return null;

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Edit Internal Team"
      centered
      data-testid="edit-internal-team-dialog"
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter team name"
            {...register('name')}
            required
            withAsterisk
            autoFocus
            disabled={updateMutation.isPending}
            error={errors.name?.message}
            data-testid="edit-internal-team-name-input"
          />

          <TextInput
            label="Department"
            placeholder="Enter department (optional)"
            {...register('department')}
            disabled={updateMutation.isPending}
            error={errors.department?.message}
            data-testid="edit-internal-team-department-input"
          />

          <TextInput
            label="Contact Person"
            placeholder="Enter contact person (optional)"
            {...register('contactPerson')}
            disabled={updateMutation.isPending}
            error={errors.contactPerson?.message}
            data-testid="edit-internal-team-contact-input"
          />

          <Textarea
            label="Notes"
            placeholder="Enter notes (optional)"
            {...register('notes')}
            rows={3}
            disabled={updateMutation.isPending}
            error={errors.notes?.message}
            data-testid="edit-internal-team-notes-input"
          />

          {backendError && (
            <Alert color="red" data-testid="edit-internal-team-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleClose}
              disabled={updateMutation.isPending}
              data-testid="edit-internal-team-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={updateMutation.isPending}
              disabled={!isValid}
              data-testid="edit-internal-team-submit"
            >
              Save Changes
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
