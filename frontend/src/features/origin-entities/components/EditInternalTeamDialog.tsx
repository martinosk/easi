import { Alert, Button, Group, Modal, Stack, Textarea, TextInput } from '@mantine/core';
import React from 'react';
import type { InternalTeam } from '../../../api/types';
import { useEditInternalTeamForm } from './useEditInternalTeamForm';

interface EditInternalTeamDialogProps {
  isOpen: boolean;
  onClose: () => void;
  team: InternalTeam | null;
}

export const EditInternalTeamDialog: React.FC<EditInternalTeamDialogProps> = ({ isOpen, onClose, team }) => {
  const { register, errors, isValid, backendError, isPending, submit } = useEditInternalTeamForm(isOpen, team, onClose);

  if (!team) return null;

  return (
    <Modal opened={isOpen} onClose={onClose} title="Edit Internal Team" centered data-testid="edit-internal-team-dialog">
      <form onSubmit={submit}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter team name"
            {...register('name')}
            required
            withAsterisk
            autoFocus
            disabled={isPending}
            error={errors.name?.message}
            data-testid="edit-internal-team-name-input"
          />

          <TextInput
            label="Department"
            placeholder="Enter department (optional)"
            {...register('department')}
            disabled={isPending}
            error={errors.department?.message}
            data-testid="edit-internal-team-department-input"
          />

          <TextInput
            label="Contact Person"
            placeholder="Enter contact person (optional)"
            {...register('contactPerson')}
            disabled={isPending}
            error={errors.contactPerson?.message}
            data-testid="edit-internal-team-contact-input"
          />

          <Textarea
            label="Notes"
            placeholder="Enter notes (optional)"
            {...register('notes')}
            rows={3}
            disabled={isPending}
            error={errors.notes?.message}
            data-testid="edit-internal-team-notes-input"
          />

          {backendError && (
            <Alert color="red" data-testid="edit-internal-team-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button variant="default" onClick={onClose} disabled={isPending} data-testid="edit-internal-team-cancel">
              Cancel
            </Button>
            <Button type="submit" loading={isPending} disabled={!isValid} data-testid="edit-internal-team-submit">
              Save Changes
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
