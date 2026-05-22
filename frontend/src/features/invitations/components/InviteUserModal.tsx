import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, NativeSelect, Stack, TextInput } from '@mantine/core';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { type InviteUserFormData, inviteUserSchema } from '../../../lib/schemas';
import type { CreateInvitationRequest } from '../types';

interface InviteUserModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (request: CreateInvitationRequest) => Promise<void>;
}

const DEFAULT_VALUES: InviteUserFormData = {
  email: '',
  role: 'stakeholder',
};

const ROLE_OPTIONS = [
  { value: 'stakeholder', label: 'Stakeholder' },
  { value: 'architect', label: 'Architect' },
  { value: 'admin', label: 'Admin' },
];

export function InviteUserModal({ isOpen, onClose, onSubmit }: InviteUserModalProps) {
  const [error, setError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting, isValid },
  } = useForm<InviteUserFormData>({
    resolver: zodResolver(inviteUserSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useEffect(() => {
    if (!isOpen) return;
    reset(DEFAULT_VALUES);
    setError(null);
  }, [isOpen, reset]);

  const submit = handleSubmit(async (data) => {
    setError(null);
    try {
      await onSubmit({ email: data.email, role: data.role });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create invitation');
    }
  });

  return (
    <Modal opened={isOpen} onClose={onClose} title="Invite User" centered data-testid="invite-user-modal">
      <form onSubmit={submit}>
        <Stack gap="md">
          <TextInput
            label="Email"
            placeholder="user@company.com"
            type="email"
            withAsterisk
            disabled={isSubmitting}
            error={errors.email?.message}
            data-testid="invite-email-input"
            {...register('email')}
          />

          <NativeSelect
            label="Role"
            data={ROLE_OPTIONS}
            withAsterisk
            disabled={isSubmitting}
            data-testid="invite-role-select"
            {...register('role')}
          />

          {error && (
            <Alert color="red" data-testid="invite-error-message">
              {error}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={onClose}
              disabled={isSubmitting}
              data-testid="invite-cancel-btn"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={isSubmitting}
              disabled={!isValid}
              data-testid="invite-submit-btn"
            >
              Create Invitation
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
}
