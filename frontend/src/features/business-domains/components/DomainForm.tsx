import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, NativeSelect, Stack, Textarea, TextInput } from '@mantine/core';
import { useEffect, useState } from 'react';
import type { UseFormRegister } from 'react-hook-form';
import { useForm } from 'react-hook-form';
import type { BusinessDomain } from '../../../api/types';
import { type BusinessDomainFormData, businessDomainSchema } from '../../../lib/schemas';
import { useEAOwnerCandidates } from '../../users/hooks/useUsers';
import type { User } from '../../users/types';

type SubmitFn = (name: string, description: string, domainArchitectId?: string) => Promise<void>;

interface DomainFormProps {
  domain?: BusinessDomain;
  mode: 'create' | 'edit';
  onSubmit: SubmitFn;
  onCancel: () => void;
}

function buildDefaults(domain?: BusinessDomain): BusinessDomainFormData {
  return {
    name: domain?.name ?? '',
    description: domain?.description ?? '',
    domainArchitectId: domain?.domainArchitectId ?? '',
  };
}

function userOptionLabel(user: User): string {
  const display = user.name || user.email;
  return `${display} (${user.role})`;
}

function getSubmitButtonText(isSubmitting: boolean, mode: 'create' | 'edit'): string {
  if (isSubmitting) return 'Saving...';
  return mode === 'create' ? 'Create' : 'Save';
}

function useDomainForm(domain: BusinessDomain | undefined, onSubmit: SubmitFn) {
  const [backendError, setBackendError] = useState<string | null>(null);
  const form = useForm<BusinessDomainFormData>({
    resolver: zodResolver(businessDomainSchema),
    defaultValues: buildDefaults(domain),
    mode: 'onChange',
  });
  const { reset, handleSubmit, watch } = form;

  useEffect(() => {
    reset(buildDefaults(domain));
    setBackendError(null);
  }, [domain, reset]);

  const submit = handleSubmit(async (data) => {
    setBackendError(null);
    try {
      await onSubmit(data.name, data.description, data.domainArchitectId || undefined);
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'An error occurred');
    }
  });

  return { form, backendError, submit, nameValue: watch('name') };
}

type Register = UseFormRegister<BusinessDomainFormData>;

interface DomainArchitectFieldProps {
  register: Register;
  isSubmitting: boolean;
}

function DomainArchitectField({ register, isSubmitting }: DomainArchitectFieldProps) {
  const { data: eligibleUsers = [], isLoading: isLoadingUsers } = useEAOwnerCandidates();
  const options = [
    { value: '', label: '-- Select Domain Architect (optional) --' },
    ...eligibleUsers.map((user) => ({ value: user.id, label: userOptionLabel(user) })),
  ];

  return (
    <NativeSelect
      label="Domain Architect"
      data={options}
      {...register('domainArchitectId')}
      disabled={isSubmitting || isLoadingUsers}
      data-testid="domain-architect-select"
      description={isLoadingUsers ? 'Loading eligible users...' : undefined}
    />
  );
}

interface FormActionsProps {
  mode: 'create' | 'edit';
  isSubmitting: boolean;
  canSubmit: boolean;
  onCancel: () => void;
}

function FormActions({ mode, isSubmitting, canSubmit, onCancel }: FormActionsProps) {
  return (
    <Group justify="flex-end" gap="sm">
      <Button variant="default" onClick={onCancel} disabled={isSubmitting} data-testid="domain-form-cancel">
        Cancel
      </Button>
      <Button type="submit" loading={isSubmitting} disabled={!canSubmit} data-testid="domain-form-submit">
        {getSubmitButtonText(isSubmitting, mode)}
      </Button>
    </Group>
  );
}

export function DomainForm({ domain, mode, onSubmit, onCancel }: DomainFormProps) {
  const { form, backendError, submit, nameValue } = useDomainForm(domain, onSubmit);
  const {
    register,
    formState: { errors, isSubmitting },
  } = form;
  const canSubmit = !isSubmitting && nameValue.trim().length > 0;

  return (
    <form onSubmit={submit} data-testid="domain-form">
      <Stack gap="md">
        <TextInput
          label="Name"
          placeholder="Enter domain name"
          {...register('name')}
          withAsterisk
          autoFocus
          disabled={isSubmitting}
          error={errors.name?.message}
          data-testid="domain-name-input"
          errorProps={{ 'data-testid': 'domain-name-error' }}
        />
        <Textarea
          label="Description"
          placeholder="Enter domain description (optional)"
          {...register('description')}
          rows={4}
          disabled={isSubmitting}
          error={errors.description?.message}
          data-testid="domain-description-input"
          errorProps={{ 'data-testid': 'domain-description-error' }}
        />
        <DomainArchitectField register={register} isSubmitting={isSubmitting} />
        {backendError && (
          <Alert color="red" data-testid="domain-form-error">
            {backendError}
          </Alert>
        )}
        <FormActions mode={mode} isSubmitting={isSubmitting} canSubmit={canSubmit} onCancel={onCancel} />
      </Stack>
    </form>
  );
}
