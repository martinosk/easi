import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Stack, Text, TextInput } from '@mantine/core';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { type InviteToEditFormData, inviteToEditSchema } from '../../../lib/schemas';
import type { ArtifactType, CreateEditGrantRequest } from '../types';

const artifactTypeLabels: Record<ArtifactType, string> = {
  capability: 'capability',
  component: 'application component',
  view: 'view',
  domain: 'business domain',
  vendor: 'vendor',
  internal_team: 'internal team',
  acquired_entity: 'acquired entity',
};

interface InviteToEditDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (request: CreateEditGrantRequest) => Promise<void>;
  artifactType: ArtifactType;
  artifactId: string;
}

const DEFAULT_VALUES: InviteToEditFormData = {
  granteeEmail: '',
  reason: '',
};

export function InviteToEditDialog({
  isOpen,
  onClose,
  onSubmit,
  artifactType,
  artifactId,
}: InviteToEditDialogProps) {
  const [error, setError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<InviteToEditFormData>({
    resolver: zodResolver(inviteToEditSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onSubmit',
  });

  useEffect(() => {
    if (!isOpen) return;
    reset(DEFAULT_VALUES);
    setError(null);
  }, [isOpen, reset]);

  const submit = handleSubmit(async (data) => {
    setError(null);
    try {
      await onSubmit({
        granteeEmail: data.granteeEmail,
        artifactType,
        artifactId,
        reason: data.reason || undefined,
      });
      reset(DEFAULT_VALUES);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to grant edit access');
    }
  });

  const handleCancel = () => {
    reset(DEFAULT_VALUES);
    setError(null);
    onClose();
  };

  return (
    <Modal opened={isOpen} onClose={handleCancel} title="Invite to Edit..." centered data-testid="invite-to-edit-dialog">
      <form onSubmit={submit}>
        <Stack gap="md">
          <Text c="dimmed" size="sm">
            Grant temporary edit access for this {artifactTypeLabels[artifactType]} to a stakeholder.
          </Text>

          <TextInput
            label="User Email"
            type="email"
            placeholder="stakeholder@company.com"
            withAsterisk
            disabled={isSubmitting}
            error={errors.granteeEmail?.message}
            data-testid="grantee-email-input"
            {...register('granteeEmail')}
          />

          <TextInput
            label="Reason"
            placeholder="Optional reason for granting access"
            disabled={isSubmitting}
            error={errors.reason?.message}
            data-testid="grant-reason-input"
            {...register('reason')}
          />

          {error && (
            <Alert color="red" data-testid="grant-error-message">
              {error}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={handleCancel}
              disabled={isSubmitting}
              data-testid="grant-cancel-btn"
            >
              Cancel
            </Button>
            <Button type="submit" loading={isSubmitting} data-testid="grant-submit-btn">
              Grant Edit Access
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
}
