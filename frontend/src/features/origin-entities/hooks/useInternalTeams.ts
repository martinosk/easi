import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { originEntitiesApi } from '../api/originEntitiesApi';
import { internalTeamsQueryKeys } from '../queryKeys';
import { invalidateFor } from '../../../lib/invalidateFor';
import { internalTeamsMutationEffects } from '../mutationEffects';
import type {
  InternalTeam,
  InternalTeamId,
  CreateInternalTeamRequest,
  UpdateInternalTeamRequest,
  ComponentId,
} from '../../../api/types';
import toast from 'react-hot-toast';

export interface UseInternalTeamsResult {
  internalTeams: InternalTeam[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  createTeam: (request: CreateInternalTeamRequest) => Promise<InternalTeam>;
  updateTeam: (id: InternalTeamId, request: UpdateInternalTeamRequest) => Promise<InternalTeam>;
  deleteTeam: (id: InternalTeamId, name: string) => Promise<void>;
}

export function useInternalTeams(): UseInternalTeamsResult {
  const query = useInternalTeamsQuery();
  const createMutation = useCreateInternalTeam();
  const updateMutation = useUpdateInternalTeam();
  const deleteMutation = useDeleteInternalTeam();

  const createTeam = useCallback(
    async (request: CreateInternalTeamRequest): Promise<InternalTeam> => {
      return createMutation.mutateAsync(request);
    },
    [createMutation]
  );

  const updateTeam = useCallback(
    async (id: InternalTeamId, request: UpdateInternalTeamRequest): Promise<InternalTeam> => {
      return updateMutation.mutateAsync({ id, request });
    },
    [updateMutation]
  );

  const deleteTeam = useCallback(
    async (id: InternalTeamId, name: string): Promise<void> => {
      await deleteMutation.mutateAsync({ id, name });
    },
    [deleteMutation]
  );

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    internalTeams: query.data ?? [],
    isLoading: query.isLoading,
    error: query.error,
    refetch,
    createTeam,
    updateTeam,
    deleteTeam,
  };
}

export function useInternalTeamsQuery() {
  return useQuery({
    queryKey: internalTeamsQueryKeys.lists(),
    queryFn: () => originEntitiesApi.internalTeams.getAll(),
  });
}

export function useInternalTeam(id: InternalTeamId | undefined) {
  return useQuery({
    queryKey: internalTeamsQueryKeys.detail(id!),
    queryFn: () => originEntitiesApi.internalTeams.getById(id!),
    enabled: !!id,
  });
}

function useTeamMutation<TArgs, TResult>(
  mutationFn: (args: TArgs) => Promise<TResult>,
  onMutationSuccess: (queryClient: ReturnType<typeof useQueryClient>, result: TResult, args: TArgs) => void,
  errorMessage: string
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: (result, args) => onMutationSuccess(queryClient, result, args),
    onError: () => toast.error(errorMessage),
  });
}

export function useCreateInternalTeam() {
  return useTeamMutation(
    (request: CreateInternalTeamRequest) => originEntitiesApi.internalTeams.create(request),
    (qc, newTeam) => {
      invalidateFor(qc, internalTeamsMutationEffects.create());
      toast.success(`Internal team "${newTeam.name}" created successfully`);
    },
    'Failed to create internal team'
  );
}

export function useUpdateInternalTeam() {
  return useTeamMutation(
    ({ id, request }: { id: InternalTeamId; request: UpdateInternalTeamRequest }) =>
      originEntitiesApi.internalTeams.update(id, request),
    (qc, updatedTeam, { id }) => {
      invalidateFor(qc, internalTeamsMutationEffects.update(id));
      toast.success(`Internal team "${updatedTeam.name}" updated`);
    },
    'Failed to update internal team'
  );
}

export function useDeleteInternalTeam() {
  return useTeamMutation(
    ({ id }: { id: InternalTeamId; name: string }) =>
      originEntitiesApi.internalTeams.delete(id),
    (qc, _, { id, name }) => {
      invalidateFor(qc, internalTeamsMutationEffects.delete(id));
      toast.success(`Internal team "${name}" deleted`);
    },
    'Failed to delete internal team'
  );
}

export function useLinkComponentToInternalTeam() {
  return useTeamMutation(
    ({ componentId, teamId, notes }: { componentId: ComponentId; teamId: InternalTeamId; notes?: string }) =>
      originEntitiesApi.internalTeams.linkComponent(componentId, teamId, notes),
    (qc, _, { teamId, componentId }) => {
      invalidateFor(qc, internalTeamsMutationEffects.linkComponent(teamId, componentId));
      toast.success('Component linked to internal team');
    },
    'Failed to link component to internal team'
  );
}

export function useUnlinkComponentFromInternalTeam() {
  return useTeamMutation(
    ({ componentId }: { teamId: InternalTeamId; componentId: ComponentId }) =>
      originEntitiesApi.internalTeams.unlinkComponent(componentId),
    (qc, _, { teamId, componentId }) => {
      invalidateFor(qc, internalTeamsMutationEffects.unlinkComponent(teamId, componentId));
      toast.success('Component unlinked');
    },
    'Failed to unlink component'
  );
}
