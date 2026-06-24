import { zodResolver } from '@hookform/resolvers/zod';
import { useLayoutEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import type { InternalTeam, InternalTeamId } from '../../../api/types';
import { type EditInternalTeamFormData, editInternalTeamSchema } from '../../../lib/schemas';
import { useUpdateInternalTeam } from '../hooks/useInternalTeams';

const toFormValues = (team: InternalTeam): EditInternalTeamFormData => ({
  name: team.name,
  department: team.department || '',
  contactPerson: team.contactPerson || '',
  notes: team.notes || '',
});

const toRequest = (data: EditInternalTeamFormData) => ({
  name: data.name,
  department: data.department || undefined,
  contactPerson: data.contactPerson || undefined,
  notes: data.notes || undefined,
});

export function useEditInternalTeamForm(isOpen: boolean, team: InternalTeam | null, onClose: () => void) {
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

  useLayoutEffect(() => {
    if (!isOpen || !team) return;
    reset(toFormValues(team));
    if (backendError !== null) queueMicrotask(() => setBackendError(null));
  }, [isOpen, team, reset, backendError]);

  const onSubmit = async (data: EditInternalTeamFormData) => {
    if (!team) return;
    setBackendError(null);
    try {
      await updateMutation.mutateAsync({ id: team.id as InternalTeamId, request: toRequest(data) });
      onClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to update internal team');
    }
  };

  return {
    register,
    errors,
    isValid,
    backendError,
    isPending: updateMutation.isPending,
    submit: handleSubmit(onSubmit),
  };
}
