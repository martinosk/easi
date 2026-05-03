import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Stack, Textarea, TextInput } from '@mantine/core';
import React, { useLayoutEffect, useState } from 'react';
import { type FieldErrors, type UseFormRegister, useForm } from 'react-hook-form';
import { type CreateInternalTeamFormData, createInternalTeamSchema } from '../../../lib/schemas';
import { useCreateInternalTeam } from '../hooks/useInternalTeams';

interface CreatedInternalTeam {
  id: string;
  name: string;
}

interface CreateInternalTeamDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onCreated?: (team: CreatedInternalTeam) => void | Promise<void>;
}

const DEFAULT_VALUES: CreateInternalTeamFormData = {
  name: '',
  department: '',
  contactPerson: '',
  notes: '',
};

interface FormFieldsProps {
  register: UseFormRegister<CreateInternalTeamFormData>;
  errors: FieldErrors<CreateInternalTeamFormData>;
  isPending: boolean;
}

function InternalTeamFields({ register, errors, isPending }: FormFieldsProps) {
  return (
    <>
      <TextInput
        label="Name"
        placeholder="Enter team name (e.g., Platform Engineering)"
        {...register('name')}
        required
        withAsterisk
        autoFocus
        disabled={isPending}
        error={errors.name?.message}
        data-testid="internal-team-name-input"
      />

      <TextInput
        label="Department"
        placeholder="Enter department (optional)"
        {...register('department')}
        disabled={isPending}
        error={errors.department?.message}
        data-testid="internal-team-department-input"
      />

      <TextInput
        label="Contact Person"
        placeholder="Enter contact person (optional)"
        {...register('contactPerson')}
        disabled={isPending}
        error={errors.contactPerson?.message}
        data-testid="internal-team-contact-input"
      />

      <Textarea
        label="Notes"
        placeholder="Enter notes (optional)"
        {...register('notes')}
        rows={3}
        disabled={isPending}
        error={errors.notes?.message}
        data-testid="internal-team-notes-input"
      />
    </>
  );
}

interface FormActionsProps {
  isPending: boolean;
  isValid: boolean;
  onCancel: () => void;
}

function FormActions({ isPending, isValid, onCancel }: FormActionsProps) {
  return (
    <Group justify="flex-end" gap="sm">
      <Button variant="default" onClick={onCancel} disabled={isPending} data-testid="create-internal-team-cancel">
        Cancel
      </Button>
      <Button type="submit" loading={isPending} disabled={!isValid} data-testid="create-internal-team-submit">
        Create
      </Button>
    </Group>
  );
}

function useInternalTeamFormState(
  isOpen: boolean,
  onClose: () => void,
  onCreated?: (team: CreatedInternalTeam) => void | Promise<void>,
) {
  const [backendError, setBackendError] = useState<string | null>(null);
  const [isHandoffPending, setHandoffPending] = useState(false);
  const createMutation = useCreateInternalTeam();
  const form = useForm<CreateInternalTeamFormData>({
    resolver: zodResolver(createInternalTeamSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useLayoutEffect(() => {
    if (!isOpen) return;
    form.reset(DEFAULT_VALUES);
    setBackendError(null);
    setHandoffPending(false);
  }, [isOpen, form]);

  const onSubmit = async (data: CreateInternalTeamFormData) => {
    setBackendError(null);
    try {
      const created = await createMutation.mutateAsync({
        name: data.name,
        department: data.department || undefined,
        contactPerson: data.contactPerson || undefined,
        notes: data.notes || undefined,
      });
      if (onCreated) {
        setHandoffPending(true);
        try {
          await onCreated(created);
        } finally {
          setHandoffPending(false);
        }
      }
      onClose();
    } catch (err) {
      setHandoffPending(false);
      setBackendError(err instanceof Error ? err.message : 'Failed to create internal team');
    }
  };

  return { form, onSubmit, backendError, isPending: createMutation.isPending || isHandoffPending };
}

export const CreateInternalTeamDialog: React.FC<CreateInternalTeamDialogProps> = ({
  isOpen,
  onClose,
  onCreated,
}) => {
  const { form, onSubmit, backendError, isPending } = useInternalTeamFormState(isOpen, onClose, onCreated);
  const {
    register,
    handleSubmit,
    formState: { errors, isValid },
  } = form;

  return (
    <Modal
      opened={isOpen}
      onClose={onClose}
      title="Create Internal Team"
      centered
      data-testid="create-internal-team-dialog"
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <InternalTeamFields register={register} errors={errors} isPending={isPending} />

          {backendError && (
            <Alert color="red" data-testid="create-internal-team-error">
              {backendError}
            </Alert>
          )}

          <FormActions isPending={isPending} isValid={isValid} onCancel={onClose} />
        </Stack>
      </form>
    </Modal>
  );
};
