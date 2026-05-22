import { Alert, Button, Group, Modal, NativeSelect, Stack, Text } from '@mantine/core';
import { type FormEvent, useEffect, useState } from 'react';
import type { UserRole } from '../../auth/types';
import type { User } from '../types';

interface ChangeRoleModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (newRole: UserRole) => Promise<void>;
  user: User;
}

const ROLE_OPTIONS = [
  { value: 'stakeholder', label: 'Stakeholder' },
  { value: 'architect', label: 'Architect' },
  { value: 'admin', label: 'Admin' },
];

export function ChangeRoleModal({ isOpen, onClose, onSubmit, user }: ChangeRoleModalProps) {
  const [role, setRole] = useState<UserRole>(user.role);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setRole(user.role);
  }, [user]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      await onSubmit(role);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to change role');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    setRole(user.role);
    setError(null);
    onClose();
  };

  return (
    <Modal opened={isOpen} onClose={handleCancel} title="Change User Role" centered data-testid="change-role-modal">
      <form onSubmit={handleSubmit}>
        <Stack gap="md">
          <Text c="dimmed" size="sm">
            Change the role for {user.email}
          </Text>

          <NativeSelect
            label="Role"
            data={ROLE_OPTIONS}
            value={role}
            onChange={(event) => setRole(event.currentTarget.value as UserRole)}
            withAsterisk
            disabled={isSubmitting}
            data-testid="change-role-select"
          />

          {error && (
            <Alert color="red" data-testid="change-role-error">
              {error}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleCancel}
              disabled={isSubmitting}
              data-testid="change-role-cancel-btn"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={isSubmitting}
              disabled={role === user.role}
              data-testid="change-role-submit-btn"
            >
              {isSubmitting ? 'Changing...' : 'Change Role'}
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
}
