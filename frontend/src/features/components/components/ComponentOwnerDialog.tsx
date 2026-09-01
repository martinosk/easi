import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Select, Stack } from '@mantine/core';
import React, { useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import type { Component } from '../../../api/types';
import { type OwnerReferenceFormData, ownerReferenceSchema } from '../../../lib/schemas';
import { useInternalTeamsQuery } from '../../origin-entities/hooks/useInternalTeams';
import { useActiveUsers } from '../../users/hooks/useUsers';
import { useAssignComponentOwner, useNominateComponentOwner } from '../hooks/useComponentOwnership';

export type OwnerDialogMode = 'nominate' | 'assign';

interface ComponentOwnerDialogProps {
  mode: OwnerDialogMode;
  component: Component;
  isOpen: boolean;
  onClose: () => void;
}

const KIND_OPTIONS = [
  { value: 'user', label: 'User' },
  { value: 'team', label: 'Internal Team' },
];

const DEFAULT_VALUES: OwnerReferenceFormData = { ownerKind: 'user', ownerId: '' };

function useOwnerOptions(kind: OwnerReferenceFormData['ownerKind']) {
  const { data: users = [] } = useActiveUsers();
  const { data: teams = [] } = useInternalTeamsQuery();

  if (kind === 'team') {
    return teams.map((team) => ({ value: team.id, label: team.name }));
  }
  return users.map((user) => ({ value: user.id, label: user.name || user.email }));
}

export const ComponentOwnerDialog: React.FC<ComponentOwnerDialogProps> = ({ mode, component, isOpen, onClose }) => {
  const [backendError, setBackendError] = useState<string | null>(null);
  const nominateMutation = useNominateComponentOwner();
  const assignMutation = useAssignComponentOwner();
  const mutation = mode === 'nominate' ? nominateMutation : assignMutation;

  const {
    control,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { isValid },
  } = useForm<OwnerReferenceFormData>({
    resolver: zodResolver(ownerReferenceSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useEffect(() => {
    if (isOpen) {
      reset(DEFAULT_VALUES);
      setBackendError(null);
    }
  }, [isOpen, reset]);

  const ownerKind = watch('ownerKind');
  const ownerOptions = useOwnerOptions(ownerKind);
  const title = mode === 'nominate' ? 'Nominate Owner' : 'Assign Owner';

  const onSubmit = async (data: OwnerReferenceFormData) => {
    setBackendError(null);
    try {
      await mutation.mutateAsync({ component, request: { ownerKind: data.ownerKind, ownerId: data.ownerId } });
      onClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : `Failed to ${mode} owner`);
    }
  };

  return (
    <Modal opened={isOpen} onClose={onClose} title={title} centered data-testid="component-owner-dialog">
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <Controller
            name="ownerKind"
            control={control}
            render={({ field }) => (
              <Select
                label="Owner Type"
                data={KIND_OPTIONS}
                required
                withAsterisk
                allowDeselect={false}
                disabled={mutation.isPending}
                data-testid="owner-kind-select"
                {...field}
                onChange={(value) => {
                  field.onChange(value);
                  setValue('ownerId', '', { shouldValidate: true });
                }}
              />
            )}
          />

          <Controller
            name="ownerId"
            control={control}
            render={({ field }) => (
              <Select
                label={ownerKind === 'team' ? 'Internal Team' : 'User'}
                placeholder={ownerKind === 'team' ? 'Select an internal team' : 'Select a user'}
                data={ownerOptions}
                required
                withAsterisk
                searchable
                disabled={mutation.isPending}
                data-testid="owner-select"
                {...field}
              />
            )}
          />

          {backendError && (
            <Alert color="red" data-testid="owner-dialog-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button variant="default" onClick={onClose} disabled={mutation.isPending} data-testid="owner-dialog-cancel">
              Cancel
            </Button>
            <Button type="submit" loading={mutation.isPending} disabled={!isValid} data-testid="owner-dialog-submit">
              {title}
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
